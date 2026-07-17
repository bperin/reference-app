package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/example/reference-app/internal/users"
)

func TestHandlerPasswordGrantReturnsOAuthTokenResponse(t *testing.T) {
	userID := uuid.New()
	usersRepo := &oauthUserRepository{user: &users.User{
		ID:           userID,
		Email:        "user@example.com",
		PasswordHash: "hash",
		DisplayName:  "User",
	}}
	service := NewService(
		usersRepo,
		&fakeRefreshRepository{},
		&fakePasswordVerifier{},
		&fakeAccessTokenIssuer{},
		&fakeRefreshTokenGenerator{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		time.Minute,
		time.Hour,
	)
	handler := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))

	request := httptest.NewRequest(http.MethodPost, "/auth/oauth/token", strings.NewReader(
		`{"grant_type":"password","username":"user@example.com","password":"correct-password"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.Token(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"access_token":"access-token"`) {
		t.Fatalf("response body = %s, want OAuth access token", response.Body.String())
	}
}

type oauthUserRepository struct {
	user *users.User
}

func (r *oauthUserRepository) Create(_ context.Context, user *users.User) error {
	r.user = user
	return nil
}

func (r *oauthUserRepository) GetByID(_ context.Context, id uuid.UUID) (*users.User, error) {
	if r.user != nil && r.user.ID == id {
		return r.user, nil
	}
	return nil, users.ErrNotFound
}

func (r *oauthUserRepository) GetByEmail(_ context.Context, email string) (*users.User, error) {
	if r.user != nil && r.user.Email == email {
		return r.user, nil
	}
	return nil, users.ErrNotFound
}

func (*oauthUserRepository) Update(context.Context, *users.User) error { return nil }

func (*oauthUserRepository) Delete(context.Context, uuid.UUID) error { return nil }
