//go:build integration

package content_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/reference-app/internal/auth"
	"github.com/example/reference-app/internal/content"
	contentrepo "github.com/example/reference-app/internal/content/repository"
	contentsqlc "github.com/example/reference-app/internal/content/repository/sqlc"
	"github.com/example/reference-app/internal/database"
	"github.com/example/reference-app/internal/security"
	"github.com/example/reference-app/internal/users"
	usersrepo "github.com/example/reference-app/internal/users/repository"
	userssqlc "github.com/example/reference-app/internal/users/repository/sqlc"
)

type mockStorage struct{}

func (m *mockStorage) GetObjectKey(userID, contentID string, originalName string) string {
	return fmt.Sprintf("%s/%s/%s", userID, contentID, originalName)
}

func (m *mockStorage) SignUploadURL(ctx context.Context, objectName, contentType string, duration time.Duration) (string, error) {
	return "https://storage.googleapis.com/test-bucket/" + objectName, nil
}

func (m *mockStorage) GetObjectAttributes(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	return &storage.ObjectAttrs{
		ContentType: "text/html",
		Generation:  42,
		Size:        512,
		Updated:     time.Now(),
	}, nil
}

type mockVerifier struct{}

func (m *mockVerifier) Verify(r *http.Request) error {
	return nil
}

func getTestDatabaseURL() (string, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}

	cmd := exec.Command("gcloud", "secrets", "versions", "access", "latest", "--project=slap-agent-builder", "--secret=reference-app-database-url")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("execute gcloud: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func TestE2EContentStorageFlow(t *testing.T) {
	ctx := context.Background()

	dsn, err := getTestDatabaseURL()
	if err != nil {
		t.Skipf("Skipping integration test: database URL not available: %v", err)
	}

	// 1. Initialize DB Pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations to ensure schema is fresh
	// RunMigrations expects "db/migrations" relative to working directory, so change to project root temporarily
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir("../../"); err != nil {
		t.Fatalf("change working directory to root: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	if err := database.RunMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	// 2. Setup domain services
	usersStore := userssqlc.New(pool)
	usersRepository := usersrepo.NewPostgres(usersStore)

	contentStore := contentsqlc.New(pool)
	contentRepository := contentrepo.NewPostgres(contentStore)

	storageAdapter := &mockStorage{}
	contentService := content.NewService(contentRepository, storageAdapter, "test-bucket")

	tokenIssuer := security.NewTokenIssuer("supersecretjwtkey", "reference-app", "reference-app", 15*time.Minute)
	requireBearer := auth.RequireBearer(tokenIssuer)

	contentHandler := content.NewHandler(contentService)
	eventarcHandler := content.NewEventarcHandler(contentService, &mockVerifier{})

	// 3. Router setup
	r := chi.NewRouter()
	content.RegisterRoutes(r, contentHandler, eventarcHandler, requireBearer)

	server := httptest.NewServer(r)
	defer server.Close()

	// 4. Create and authenticate a test user
	userID := uuid.New()
	testUser := &users.User{
		ID:           userID,
		Email:        fmt.Sprintf("test-%s@example.com", userID.String()[:8]),
		PasswordHash: "hashedpassword",
		DisplayName:  "Integration Test User",
	}

	if err := usersRepository.Create(ctx, testUser); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	token, err := tokenIssuer.IssueAccessToken(userID, []string{"user"}, []string{"user"})
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	// 5. Flow Step 1: Request Reservation (Reserve Endpoint)
	reserveBody, _ := json.Marshal(map[string]interface{}{
		"mime_type": "text/html",
		"filename":  "document.html",
	})

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/contents", bytes.NewReader(reserveBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("reserve request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Fatalf("reserve status = %d, want 201; response = %s", resp.StatusCode, buf.String())
	}

	var reserveResp struct {
		Content   content.Content `json:"content"`
		SignedURL string          `json:"signed_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reserveResp); err != nil {
		t.Fatalf("decode reserve response: %v", err)
	}

	if reserveResp.Content.UploadStatus != "pending" {
		t.Errorf("expected upload status 'pending', got %q", reserveResp.Content.UploadStatus)
	}
	if !strings.Contains(reserveResp.SignedURL, "document.html") {
		t.Errorf("expected signed URL to contain filename, got %q", reserveResp.SignedURL)
	}

	// 6. Flow Step 2: Trigger Eventarc completion callback (simulating GCS finalization)
	eventBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"bucket": "test-bucket",
			"name":   reserveResp.Content.ObjectName,
		},
	})

	reqEvent, _ := http.NewRequest(http.MethodPost, server.URL+"/events/storage", bytes.NewReader(eventBody))
	reqEvent.Header.Set("Content-Type", "application/json")

	respEvent, err := http.DefaultClient.Do(reqEvent)
	if err != nil {
		t.Fatalf("eventarc callback request failed: %v", err)
	}
	defer respEvent.Body.Close()

	if respEvent.StatusCode != http.StatusNoContent {
		buf := new(bytes.Buffer)
		buf.ReadFrom(respEvent.Body)
		t.Fatalf("eventarc status = %d, want 204; response = %s", respEvent.StatusCode, buf.String())
	}

	// 7. Flow Step 3: Verify Database reflects completion
	dbContent, err := contentRepository.GetByBucketAndName(ctx, "test-bucket", reserveResp.Content.ObjectName)
	if err != nil {
		t.Fatalf("get content from db: %v", err)
	}

	if dbContent.UploadStatus != "uploaded" {
		t.Errorf("expected status 'uploaded', got %q", dbContent.UploadStatus)
	}
	if dbContent.EffectiveMimeType == nil || *dbContent.EffectiveMimeType != "text/html" {
		t.Errorf("expected effective MIME 'text/html', got %v", dbContent.EffectiveMimeType)
	}
	if dbContent.SizeBytes == nil || *dbContent.SizeBytes != 512 {
		t.Errorf("expected size 512, got %v", dbContent.SizeBytes)
	}
}
