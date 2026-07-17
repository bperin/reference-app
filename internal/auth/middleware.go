package auth

import (
	"net/http"
	"strings"

	"github.com/example/reference-app/internal/http/response"
	"github.com/example/reference-app/internal/logging"
	"github.com/example/reference-app/internal/security"
)

func RequireBearer(issuer *security.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			scheme, token, ok := strings.Cut(authorization, " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "bearer token required")
				return
			}

			claims, err := issuer.ValidateAccessToken(strings.TrimSpace(token))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "invalid or expired bearer token")
				return
			}

			next.ServeHTTP(w, r.WithContext(logging.WithUserID(r.Context(), claims.UserID)))
		})
	}
}
