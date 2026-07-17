package posts

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakePostRepo struct {
	posts map[uuid.UUID]*Post
}

func (f *fakePostRepo) Create(ctx context.Context, p *Post) error {
	f.posts[p.ID] = p
	return nil
}

func (f *fakePostRepo) GetByID(ctx context.Context, id uuid.UUID) (*Post, error) {
	p, ok := f.posts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (f *fakePostRepo) Update(ctx context.Context, p *Post) error {
	f.posts[p.ID] = p
	return nil
}

func (f *fakePostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.posts, id)
	return nil
}

func (f *fakePostRepo) ListPublished(ctx context.Context, limit, offset int32) ([]*Post, error) {
	var res []*Post
	for _, p := range f.posts {
		if p.Published {
			res = append(res, p)
		}
	}
	return res, nil
}

func (f *fakePostRepo) ListAll(ctx context.Context, limit, offset int32) ([]*Post, error) {
	var res []*Post
	for _, p := range f.posts {
		res = append(res, p)
	}
	return res, nil
}

func TestPostService_CRUD(t *testing.T) {
	repo := &fakePostRepo{posts: make(map[uuid.UUID]*Post)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewService(repo, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	authorID := uuid.New()

	// 1. Create Post
	p, err := service.CreatePost(ctx, authorID, "Hello World", "Content of the post")
	if err != nil {
		t.Fatalf("unexpected error creating post: %v", err)
	}
	if p.Title != "Hello World" {
		t.Errorf("expected Title 'Hello World', got %q", p.Title)
	}
	if p.Published {
		t.Error("expected new post to be unpublished by default")
	}

	// 2. Read Post
	retrieved, err := service.GetPost(ctx, p.ID)
	if err != nil {
		t.Fatalf("unexpected error getting post: %v", err)
	}
	if retrieved.Content != "Content of the post" {
		t.Errorf("expected Content 'Content of the post', got %q", retrieved.Content)
	}

	// 3. Update Post (Owner)
	updated, err := service.UpdatePost(ctx, p.ID, authorID, "Hello Updated", "New content")
	if err != nil {
		t.Fatalf("unexpected error updating post: %v", err)
	}
	if updated.Title != "Hello Updated" {
		t.Errorf("expected updated title 'Hello Updated', got %q", updated.Title)
	}

	// 4. Update Post (Non-Owner should fail)
	wrongAuthor := uuid.New()
	_, err = service.UpdatePost(ctx, p.ID, wrongAuthor, "Hack Title", "Hack Content")
	if err == nil {
		t.Error("expected updating post by non-owner to return error")
	}

	// 5. Publish Post (Owner)
	published, err := service.PublishPost(ctx, p.ID, authorID)
	if err != nil {
		t.Fatalf("unexpected error publishing post: %v", err)
	}
	if !published.Published {
		t.Error("expected post to be published")
	}

	// 6. Delete Post (Owner)
	err = service.DeletePost(ctx, p.ID, authorID)
	if err != nil {
		t.Fatalf("unexpected error deleting post: %v", err)
	}

	_, err = service.GetPost(ctx, p.ID)
	if err == nil {
		t.Error("expected error loading deleted post, got nil")
	}
}
