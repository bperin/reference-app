package content

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// Mocks

type MockRepository struct {
	CreateFn             func(ctx context.Context, c *Content) error
	GetByIDAndUserFn     func(ctx context.Context, id, userID uuid.UUID) (*Content, error)
	ListChildrenFn       func(ctx context.Context, parentID, userID uuid.UUID, limit, offset int32) ([]*Content, error)
	GetByBucketAndNameFn func(ctx context.Context, bucket, objectName string) (*Content, error)
	CompleteFn           func(ctx context.Context, c *Content) error
}

func (m *MockRepository) Create(ctx context.Context, c *Content) error { return m.CreateFn(ctx, c) }
func (m *MockRepository) GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*Content, error) {
	return m.GetByIDAndUserFn(ctx, id, userID)
}
func (m *MockRepository) ListChildren(ctx context.Context, parentID, userID uuid.UUID, limit, offset int32) ([]*Content, error) {
	return m.ListChildrenFn(ctx, parentID, userID, limit, offset)
}
func (m *MockRepository) GetByBucketAndName(ctx context.Context, bucket, objectName string) (*Content, error) {
	return m.GetByBucketAndNameFn(ctx, bucket, objectName)
}
func (m *MockRepository) Complete(ctx context.Context, c *Content) error { return m.CompleteFn(ctx, c) }

type MockStorage struct {
	GetObjectKeyFn        func(userID, contentID string, originalName string) string
	SignUploadURLFn       func(ctx context.Context, objectName, contentType string, duration time.Duration) (string, error)
	GetObjectAttributesFn func(ctx context.Context, objectName string) (*storage.ObjectAttrs, error)
}

func (m *MockStorage) GetObjectKey(userID, contentID string, originalName string) string {
	return m.GetObjectKeyFn(userID, contentID, originalName)
}
func (m *MockStorage) SignUploadURL(ctx context.Context, objectName, contentType string, duration time.Duration) (string, error) {
	return m.SignUploadURLFn(ctx, objectName, contentType, duration)
}
func (m *MockStorage) GetObjectAttributes(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	return m.GetObjectAttributesFn(ctx, objectName)
}

// Tests

func TestReserve(t *testing.T) {
	repo := &MockRepository{}
	st := &MockStorage{}
	svc := NewService(repo, st, "test-bucket")

	userID := uuid.New()
	parentID := uuid.New()
	mime := "image/png"
	name := "test.png"
	objectName := "users/" + userID.String() + "/test.png"

	st.GetObjectKeyFn = func(uID, cID, n string) string { return objectName }
	repo.CreateFn = func(ctx context.Context, c *Content) error {
		if c.UserID != userID || c.ParentID == nil || *c.ParentID != parentID {
			t.Errorf("expected userID %v and parentID %v, got %v and %v", userID, parentID, c.UserID, c.ParentID)
		}
		return nil
	}
	st.SignUploadURLFn = func(ctx context.Context, oN, cT string, d time.Duration) (string, error) {
		return "http://signed.url", nil
	}

	c, url, err := svc.Reserve(context.Background(), userID, &parentID, mime, name)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, c.UserID)
	}
	if url != "http://signed.url" {
		t.Errorf("expected url http://signed.url, got %v", url)
	}
}

func TestComplete_Success(t *testing.T) {
	repo := &MockRepository{}
	st := &MockStorage{}
	svc := NewService(repo, st, "test-bucket")

	bucket := "test-bucket"
	objectName := "test-object"
	c := &Content{ID: uuid.New()}
	attrs := &storage.ObjectAttrs{
		ContentType: "image/png",
		Generation:  1,
		Size:        1024,
		Updated:     time.Now(),
	}

	repo.GetByBucketAndNameFn = func(ctx context.Context, b, oN string) (*Content, error) {
		return c, nil
	}
	st.GetObjectAttributesFn = func(ctx context.Context, oN string) (*storage.ObjectAttrs, error) {
		return attrs, nil
	}
	repo.CompleteFn = func(ctx context.Context, content *Content) error {
		if content.EffectiveMimeType == nil || *content.EffectiveMimeType != "image/png" {
			t.Errorf("expected mime image/png, got %v", content.EffectiveMimeType)
		}
		return nil
	}

	res, err := svc.Complete(context.Background(), bucket, objectName)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.EffectiveMimeType == nil || *res.EffectiveMimeType != "image/png" {
		t.Errorf("expected mime image/png, got %v", res.EffectiveMimeType)
	}
}
