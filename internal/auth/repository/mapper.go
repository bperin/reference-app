package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/example/reference-app/internal/auth"
	authsqlc "github.com/example/reference-app/internal/auth/repository/sqlc"
)

func toDomain(row authsqlc.RefreshToken) *auth.RefreshSession {
	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		revoked := row.RevokedAt.Time
		revokedAt = &revoked
	}
	return &auth.RefreshSession{
		ID:        fromUUID(row.ID),
		UserID:    fromUUID(row.UserID),
		TokenHash: row.TokenHash,
		FamilyID:  fromUUID(row.FamilyID),
		ExpiresAt: row.ExpiresAt.Time,
		RevokedAt: revokedAt,
		CreatedAt: row.CreatedAt.Time,
	}
}

func toUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

func toTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
