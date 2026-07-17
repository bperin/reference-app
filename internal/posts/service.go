package posts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.With(slog.String("component", "posts_service")),
	}
}

func (s *Service) CreatePost(ctx context.Context, authorID uuid.UUID, title, content string) (*Post, error) {
	if title == "" || content == "" {
		return nil, ErrInvalidInput
	}

	p := &Post{
		ID:        uuid.New(),
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		Published: false,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create post repository: %w", err)
	}

	s.logger.InfoContext(ctx, "post created successfully", slog.String("post_id", p.ID.String()))
	return p, nil
}

func (s *Service) GetPost(ctx context.Context, id uuid.UUID) (*Post, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	return p, nil
}

func (s *Service) UpdatePost(ctx context.Context, id, authorID uuid.UUID, title, content string) (*Post, error) {
	if title == "" || content == "" {
		return nil, ErrInvalidInput
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post for update: %w", err)
	}

	if p.AuthorID != authorID {
		return nil, ErrForbiddenOwner
	}

	p.Title = title
	p.Content = content

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}

	s.logger.InfoContext(ctx, "post updated successfully", slog.String("post_id", p.ID.String()))
	return p, nil
}

func (s *Service) PublishPost(ctx context.Context, id, authorID uuid.UUID) (*Post, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post for publish: %w", err)
	}

	if p.AuthorID != authorID {
		return nil, ErrForbiddenOwner
	}

	p.Published = true

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("publish post: %w", err)
	}

	s.logger.InfoContext(ctx, "post published successfully", slog.String("post_id", p.ID.String()))
	return p, nil
}

func (s *Service) DeletePost(ctx context.Context, id, authorID uuid.UUID) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get post for delete: %w", err)
	}

	if p.AuthorID != authorID {
		return ErrForbiddenOwner
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	s.logger.InfoContext(ctx, "post deleted successfully", slog.String("post_id", p.ID.String()))
	return nil
}

func (s *Service) ListPosts(ctx context.Context, limit, offset int32, showAll bool) ([]*Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var postsList []*Post
	var err error

	if showAll {
		postsList, err = s.repo.ListAll(ctx, limit, offset)
	} else {
		postsList, err = s.repo.ListPublished(ctx, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("list posts repository: %w", err)
	}

	return postsList, nil
}
