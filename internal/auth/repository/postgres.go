package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/reference-app/internal/auth"
	authsqlc "github.com/example/reference-app/internal/auth/repository/sqlc"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

var _ auth.RefreshRepository = (*PostgresRepository)(nil)

func (r *PostgresRepository) CreateSession(ctx context.Context, session *auth.RefreshSession) error {
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	row, err := authsqlc.New(r.pool).CreateRefreshToken(ctx, authsqlc.CreateRefreshTokenParams{
		ID:        toUUID(session.ID),
		UserID:    toUUID(session.UserID),
		TokenHash: session.TokenHash,
		FamilyID:  toUUID(session.FamilyID),
		ExpiresAt: toTimestamp(session.ExpiresAt),
		CreatedAt: toTimestamp(session.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}
	*session = *toDomain(row)
	return nil
}

func (r *PostgresRepository) GetSessionByHash(ctx context.Context, tokenHash string) (*auth.RefreshSession, error) {
	row, err := authsqlc.New(r.pool).GetRefreshTokenByHash(ctx, authsqlc.GetRefreshTokenByHashParams{TokenHash: tokenHash})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrRefreshSessionNotFound
		}
		return nil, fmt.Errorf("get refresh session: %w", err)
	}
	return toDomain(row), nil
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	if err := authsqlc.New(r.pool).RevokeRefreshToken(ctx, authsqlc.RevokeRefreshTokenParams{
		ID:        toUUID(id),
		RevokedAt: toTimestamp(time.Now()),
	}); err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	if err := authsqlc.New(r.pool).RevokeTokenFamily(ctx, authsqlc.RevokeTokenFamilyParams{
		FamilyID:  toUUID(familyID),
		RevokedAt: toTimestamp(time.Now()),
	}); err != nil {
		return fmt.Errorf("revoke refresh family: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RotateSession(ctx context.Context, currentID uuid.UUID, replacement *auth.RefreshSession) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := authsqlc.New(tx)
	if _, err := queries.ConsumeRefreshToken(ctx, authsqlc.ConsumeRefreshTokenParams{
		ID:        toUUID(currentID),
		RevokedAt: toTimestamp(time.Now()),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.ErrRefreshSessionNotFound
		}
		return fmt.Errorf("consume refresh session: %w", err)
	}

	row, err := queries.CreateRefreshToken(ctx, authsqlc.CreateRefreshTokenParams{
		ID:        toUUID(replacement.ID),
		UserID:    toUUID(replacement.UserID),
		TokenHash: replacement.TokenHash,
		FamilyID:  toUUID(replacement.FamilyID),
		ExpiresAt: toTimestamp(replacement.ExpiresAt),
		CreatedAt: toTimestamp(replacement.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("create rotated refresh session: %w", err)
	}
	*replacement = *toDomain(row)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit refresh rotation: %w", err)
	}
	return nil
}
