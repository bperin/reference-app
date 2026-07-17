package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/example/reference-app/internal/posts"
	postssqlc "github.com/example/reference-app/internal/posts/repository/sqlc"
)

type PostgresRepository struct {
	store postssqlc.Querier
}

func NewPostgres(store postssqlc.Querier) *PostgresRepository {
	return &PostgresRepository{
		store: store,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, p *posts.Post) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	params := postssqlc.CreatePostParams{
		ID:        toPgUUID(p.ID),
		AuthorID:  toPgUUID(p.AuthorID),
		Title:     p.Title,
		Content:   p.Content,
		Published: p.Published,
		CreatedAt: toPgTime(p.CreatedAt),
		UpdatedAt: toPgTime(p.UpdatedAt),
	}

	record, err := r.store.CreatePost(ctx, params)
	if err != nil {
		return fmt.Errorf("postgres create post: %w", err)
	}

	*p = *toDomain(record)
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*posts.Post, error) {
	params := postssqlc.GetPostByIDParams{
		ID: toPgUUID(id),
	}
	record, err := r.store.GetPostByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, posts.ErrNotFound
		}
		return nil, fmt.Errorf("postgres get post: %w", err)
	}
	return toDomain(record), nil
}

func (r *PostgresRepository) Update(ctx context.Context, p *posts.Post) error {
	p.UpdatedAt = time.Now()

	params := postssqlc.UpdatePostParams{
		ID:        toPgUUID(p.ID),
		AuthorID:  toPgUUID(p.AuthorID),
		Title:     p.Title,
		Content:   p.Content,
		Published: p.Published,
		UpdatedAt: toPgTime(p.UpdatedAt),
	}

	record, err := r.store.UpdatePost(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return posts.ErrNotFound
		}
		return fmt.Errorf("postgres update post: %w", err)
	}

	*p = *toDomain(record)
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	params := postssqlc.DeletePostParams{
		ID:       toPgUUID(p.ID),
		AuthorID: toPgUUID(p.AuthorID),
	}

	err = r.store.DeletePost(ctx, params)
	if err != nil {
		return fmt.Errorf("postgres delete post: %w", err)
	}

	return nil
}

func (r *PostgresRepository) ListPublished(ctx context.Context, limit, offset int32) ([]*posts.Post, error) {
	records, err := r.store.ListPublishedPosts(ctx, postssqlc.ListPublishedPostsParams{
		ResultLimit:  limit,
		ResultOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres list published posts: %w", err)
	}

	result := make([]*posts.Post, len(records))
	for i, rec := range records {
		result[i] = toDomain(rec)
	}
	return result, nil
}

func (r *PostgresRepository) ListAll(ctx context.Context, limit, offset int32) ([]*posts.Post, error) {
	records, err := r.store.ListAllPosts(ctx, postssqlc.ListAllPostsParams{
		ResultLimit:  limit,
		ResultOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres list all posts: %w", err)
	}

	result := make([]*posts.Post, len(records))
	for i, rec := range records {
		result[i] = toDomain(rec)
	}
	return result, nil
}
