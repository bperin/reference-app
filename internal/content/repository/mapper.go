package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/example/reference-app/internal/content"
	contentsqlc "github.com/example/reference-app/internal/content/repository/sqlc"
)

func toDomain(c contentsqlc.Content) *content.Content {
	return &content.Content{
		ID:                fromPgUUID(c.ID),
		UserID:            fromPgUUID(c.UserID),
		ParentID:          fromPgUUIDPtr(c.ParentID),
		ProjectID:         fromPgUUIDPtr(c.ProjectID),
		DeclaredMimeType:  c.DeclaredMimeType,
		EffectiveMimeType: c.EffectiveMimeType,
		Bucket:            c.Bucket,
		ObjectName:        c.ObjectName,
		GcsUri:            c.GcsUri,
		PublicUrl:         c.PublicUrl,
		UploadStatus:      c.UploadStatus,
		ObjectGeneration:  c.ObjectGeneration,
		SizeBytes:         c.SizeBytes,
		CompletedAt:       fromPgTimePtr(c.CompletedAt),
		CreatedAt:         fromPgTime(c.CreatedAt),
		UpdatedAt:         fromPgTime(c.UpdatedAt),
	}
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(pg pgtype.UUID) uuid.UUID {
	if !pg.Valid {
		return uuid.Nil
	}
	return uuid.UUID(pg.Bytes)
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromPgUUIDPtr(pg pgtype.UUID) *uuid.UUID {
	if !pg.Valid {
		return nil
	}
	u := uuid.UUID(pg.Bytes)
	return &u
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

func toPgTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromPgTimePtr(pg pgtype.Timestamptz) *time.Time {
	if !pg.Valid {
		return nil
	}
	t := pg.Time
	return &t
}
