package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/example/reference-app/internal/users"
	userssqlc "github.com/example/reference-app/internal/users/repository/sqlc"
)

func toDomain(u userssqlc.User) *users.User {
	return &users.User{
		ID:           fromPgUUID(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		CreatedAt:    fromPgTime(u.CreatedAt),
		UpdatedAt:    fromPgTime(u.UpdatedAt),
	}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(pg pgtype.UUID) uuid.UUID {
	if !pg.Valid {
		return uuid.Nil
	}
	return pg.Bytes
}

func toPgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromPgTime(pg pgtype.Timestamptz) time.Time {
	if !pg.Valid {
		return time.Time{}
	}
	return pg.Time
}
