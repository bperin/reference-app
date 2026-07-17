package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/example/reference-app/internal/logging"
	"github.com/example/reference-app/internal/security"
)

func TestRequireBearerRejectsMissingAuthorization(t *testing.T) {
	issuer := security.NewTokenIssuer("secret", "reference-app", "reference-app", time.Minute)
	handler := RequireBearer(issuer)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called without a bearer token")
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/me", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRequireBearerAddsUserToContext(t *testing.T) {
	issuer := security.NewTokenIssuer("secret", "reference-app", "reference-app", time.Minute)
	userID := uuid.New()
	token, err := issuer.IssueAccessToken(userID, []string{"user"}, []string{"user"})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	handler := RequireBearer(issuer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUserID, ok := logging.GetUserID(r.Context())
		if !ok || actualUserID != userID {
			t.Fatalf("authenticated user = (%v, %v), want (%v, true)", actualUserID, ok, userID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
