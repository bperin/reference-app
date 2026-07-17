package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"cloud.google.com/go/storage"
)

type Repository interface {
	Create(ctx context.Context, c *Content) error
	GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*Content, error)
	ListChildren(ctx context.Context, parentID, userID uuid.UUID, limit, offset int32) ([]*Content, error)
	GetByBucketAndName(ctx context.Context, bucket, objectName string) (*Content, error)
	Complete(ctx context.Context, c *Content) error
}

type Storage interface {
	GetObjectKey(userID, contentID string, originalName string) string
	SignUploadURL(ctx context.Context, objectName, contentType string, duration time.Duration) (string, error)
	GetObjectAttributes(ctx context.Context, objectName string) (*storage.ObjectAttrs, error)
}

type Service struct {
	repo    Repository
	storage Storage
	bucket  string
}

func NewService(repo Repository, storage Storage, bucket string) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		bucket:  bucket,
	}
}

func (s *Service) Reserve(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, mimeType, filename string) (*Content, string, error) {
	contentID := uuid.New()
	objectName := s.storage.GetObjectKey(userID.String(), contentID.String(), filename)

	c := &Content{
		ID:               contentID,
		UserID:           userID,
		ParentID:         parentID,
		DeclaredMimeType: mimeType,
		Bucket:           s.bucket,
		ObjectName:       objectName,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, "", fmt.Errorf("create content: %w", err)
	}

	signedURL, err := s.storage.SignUploadURL(ctx, objectName, mimeType, 15*time.Minute)
	if err != nil {
		return nil, "", fmt.Errorf("sign url: %w", err)
	}

	return c, signedURL, nil
}

func (s *Service) Complete(ctx context.Context, bucket, objectName string) (*Content, error) {
	c, err := s.repo.GetByBucketAndName(ctx, bucket, objectName)
	if err != nil {
		return nil, err
	}

	attrs, err := s.storage.GetObjectAttributes(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("get gcs attributes: %w", err)
	}

	c.EffectiveMimeType = &attrs.ContentType
	c.GcsUri = new(string)
	*c.GcsUri = fmt.Sprintf("gs://%s/%s", bucket, objectName)
	c.ObjectGeneration = &attrs.Generation
	c.SizeBytes = &attrs.Size
	c.CompletedAt = &attrs.Updated

	if err := s.repo.Complete(ctx, c); err != nil {
		return nil, fmt.Errorf("complete repository: %w", err)
	}

	return c, nil
}
