package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/example/reference-app/internal/posts"
	postssqlc "github.com/example/reference-app/internal/posts/repository/sqlc"
)

func toDomain(p postssqlc.Post) *posts.Post {
	return &posts.Post{
		ID:        fromPgUUID(p.ID),
		AuthorID:  fromPgUUID(p.AuthorID),
		Title:     p.Title,
		Content:   p.Content,
		Published: p.Published,
		CreatedAt: fromPgTime(p.CreatedAt),
		UpdatedAt: fromPgTime(p.UpdatedAt),
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
