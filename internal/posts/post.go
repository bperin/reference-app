package posts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("post not found")
	ErrForbiddenOwner = errors.New("only the author can perform this action")
	ErrInvalidInput   = errors.New("invalid post input")
)

type Post struct {
	ID        uuid.UUID
	AuthorID  uuid.UUID
	Title     string
	Content   string
	Published bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	Create(ctx context.Context, p *Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
	Update(ctx context.Context, p *Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListPublished(ctx context.Context, limit, offset int32) ([]*Post, error)
	ListAll(ctx context.Context, limit, offset int32) ([]*Post, error)
}
