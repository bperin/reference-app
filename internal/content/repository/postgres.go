package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/example/reference-app/internal/content"
	contentsqlc "github.com/example/reference-app/internal/content/repository/sqlc"
)

type PostgresRepository struct {
	store contentsqlc.Querier
}

func NewPostgres(store contentsqlc.Querier) *PostgresRepository {
	return &PostgresRepository{
		store: store,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, c *content.Content) error {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.UploadStatus = "pending"

	params := contentsqlc.CreateContentParams{
		ID:               toPgUUID(c.ID),
		UserID:           toPgUUID(c.UserID),
		ParentID:         toPgUUIDPtr(c.ParentID),
		ProjectID:        toPgUUIDPtr(c.ProjectID),
		DeclaredMimeType: c.DeclaredMimeType,
		Bucket:           c.Bucket,
		ObjectName:       c.ObjectName,
		CreatedAt:        toPgTime(c.CreatedAt),
		UpdatedAt:        toPgTime(c.UpdatedAt),
	}

	record, err := r.store.CreateContent(ctx, params)
	if err != nil {
		return fmt.Errorf("postgres create content: %w", err)
	}

	*c = *toDomain(record)
	return nil
}

func (r *PostgresRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*content.Content, error) {
	params := contentsqlc.GetContentByIDAndUserParams{
		ID:     toPgUUID(id),
		UserID: toPgUUID(userID),
	}
	record, err := r.store.GetContentByIDAndUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, content.ErrNotFound
		}
		return nil, fmt.Errorf("postgres get content by id and user: %w", err)
	}
	return toDomain(record), nil
}

func (r *PostgresRepository) ListChildren(ctx context.Context, parentID uuid.UUID, limit, offset int32) ([]*content.Content, error) {
	records, err := r.store.ListContentChildren(ctx, contentsqlc.ListContentChildrenParams{
		ParentID:     toPgUUIDPtr(&parentID),
		ResultLimit:  limit,
		ResultOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres list content children: %w", err)
	}

	result := make([]*content.Content, len(records))
	for i, rec := range records {
		result[i] = toDomain(rec)
	}
	return result, nil
}

func (r *PostgresRepository) GetByBucketAndName(ctx context.Context, bucket, objectName string) (*content.Content, error) {
	params := contentsqlc.GetContentByBucketAndNameParams{
		Bucket:     bucket,
		ObjectName: objectName,
	}
	record, err := r.store.GetContentByBucketAndName(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, content.ErrNotFound
		}
		return nil, fmt.Errorf("postgres get content by bucket and name: %w", err)
	}
	return toDomain(record), nil
}

func (r *PostgresRepository) Complete(ctx context.Context, c *content.Content) error {
	c.UpdatedAt = time.Now()

	params := contentsqlc.CompleteContentParams{
		Bucket:            c.Bucket,
		ObjectName:        c.ObjectName,
		EffectiveMimeType: c.EffectiveMimeType,
		GcsUri:            c.GcsUri,
		PublicUrl:         c.PublicUrl,
		ObjectGeneration:  c.ObjectGeneration,
		SizeBytes:         c.SizeBytes,
		CompletedAt:       toPgTimePtr(c.CompletedAt),
		UpdatedAt:         toPgTime(c.UpdatedAt),
	}

	record, err := r.store.CompleteContent(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return content.ErrStaleGeneration
		}
		return fmt.Errorf("postgres complete content: %w", err)
	}

	*c = *toDomain(record)
	return nil
}
