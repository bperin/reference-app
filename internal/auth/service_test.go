package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/example/reference-app/internal/users"
)

type fakeUserRepository struct {
	created *users.User
}

func (r *fakeUserRepository) Create(_ context.Context, user *users.User) error {
	r.created = user
	return nil
}

func (*fakeUserRepository) GetByID(context.Context, uuid.UUID) (*users.User, error) {
	return nil, users.ErrNotFound
}

func (*fakeUserRepository) GetByEmail(context.Context, string) (*users.User, error) {
	return nil, users.ErrNotFound
}

func (*fakeUserRepository) Update(context.Context, *users.User) error { return nil }

func (*fakeUserRepository) Delete(context.Context, uuid.UUID) error { return nil }

type fakeRefreshRepository struct{}

func (*fakeRefreshRepository) CreateSession(context.Context, *RefreshSession) error { return nil }

func (*fakeRefreshRepository) GetSessionByHash(context.Context, string) (*RefreshSession, error) {
	return nil, ErrRefreshSessionNotFound
}

func (*fakeRefreshRepository) RotateSession(context.Context, uuid.UUID, *RefreshSession) error {
	return nil
}

func (*fakeRefreshRepository) RevokeSession(context.Context, uuid.UUID) error { return nil }

func (*fakeRefreshRepository) RevokeFamily(context.Context, uuid.UUID) error { return nil }

type fakePasswordVerifier struct{}

func (*fakePasswordVerifier) Hash(string) (string, error) { return "hash", nil }

func (*fakePasswordVerifier) Verify(string, string) error { return nil }

type fakeAccessTokenIssuer struct{}

func (*fakeAccessTokenIssuer) IssueAccessToken(uuid.UUID, []string, []string) (string, error) {
	return "access-token", nil
}

type fakeRefreshTokenGenerator struct{}

func (*fakeRefreshTokenGenerator) Generate() (string, error) { return "refresh-token", nil }

func (*fakeRefreshTokenGenerator) Hash(string) string { return "refresh-token-hash" }

func TestRegisterUserSetsPersistenceTimestamps(t *testing.T) {
	usersRepo := &fakeUserRepository{}
	service := NewService(
		usersRepo,
		&fakeRefreshRepository{},
		&fakePasswordVerifier{},
		&fakeAccessTokenIssuer{},
		&fakeRefreshTokenGenerator{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		0,
		0,
	)

	_, err := service.RegisterUser(context.Background(), "user@example.com", "password", "User")
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if usersRepo.created == nil {
		t.Fatal("RegisterUser() did not persist a user")
	}
	if usersRepo.created.CreatedAt.IsZero() || usersRepo.created.UpdatedAt.IsZero() {
		t.Fatalf("RegisterUser() timestamps = (%v, %v), want both populated", usersRepo.created.CreatedAt, usersRepo.created.UpdatedAt)
	}
}
