package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/example/reference-app/internal/config"
	"github.com/coreos/go-oidc/v3/oidc"
)

type EventarcVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewEventarcVerifier(ctx context.Context, cfg *config.Config) (*EventarcVerifier, error) {
	provider, err := oidc.NewProvider(ctx, cfg.EventarcIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.EventarcAudience})
	return &EventarcVerifier{verifier: verifier}, nil
}

func (v *EventarcVerifier) Verify(r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("missing or invalid authorization header")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	_, err := v.verifier.Verify(r.Context(), token)
	if err != nil {
		return fmt.Errorf("verify token: %w", err)
	}
	return nil
}
