package users

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeUserRepo struct {
	users map[uuid.UUID]*User
}

func (f *fakeUserRepo) Create(ctx context.Context, u *User) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (f *fakeUserRepo) Update(ctx context.Context, u *User) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(f.users, id)
	return nil
}

func TestUserService_Profile(t *testing.T) {
	repo := &fakeUserRepo{users: make(map[uuid.UUID]*User)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewService(repo, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	userID := uuid.New()
	user := &User{
		ID:          userID,
		Email:       "test@example.com",
		DisplayName: "Original Name",
	}

	_ = repo.Create(ctx, user)

	// Test profile read
	retrieved, err := service.GetProfile(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error getting profile: %v", err)
	}
	if retrieved.DisplayName != "Original Name" {
		t.Errorf("expected DisplayName 'Original Name', got %q", retrieved.DisplayName)
	}

	// Test profile update
	updated, err := service.UpdateProfile(ctx, userID, "Updated Name")
	if err != nil {
		t.Fatalf("unexpected error updating profile: %v", err)
	}
	if updated.DisplayName != "Updated Name" {
		t.Errorf("expected DisplayName 'Updated Name', got %q", updated.DisplayName)
	}

	// Test update validation error
	_, err = service.UpdateProfile(ctx, userID, "")
	if err == nil {
		t.Error("expected error when updating profile with empty display name")
	}

	// Test account deletion
	err = service.DeleteAccount(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error deleting profile: %v", err)
	}

	_, err = service.GetProfile(ctx, userID)
	if err == nil {
		t.Error("expected error getting deleted profile, got nil")
	}
}
