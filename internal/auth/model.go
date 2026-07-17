package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrRefreshTokenRevoked    = errors.New("refresh token revoked")
	ErrRefreshTokenExpired    = errors.New("refresh token expired")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
)

type RefreshSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	FamilyID  uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (s *RefreshSession) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *RefreshSession) IsExpired() bool {
	return !time.Now().Before(s.ExpiresAt)
}

type RefreshRepository interface {
	CreateSession(ctx context.Context, session *RefreshSession) error
	GetSessionByHash(ctx context.Context, tokenHash string) (*RefreshSession, error)
	RotateSession(ctx context.Context, currentID uuid.UUID, replacement *RefreshSession) error
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
}
