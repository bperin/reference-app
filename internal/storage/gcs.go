package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSClient interface {
	Bucket(name string) *storage.BucketHandle
	Close() error
}

type GCSAdapter struct {
	client GCSClient
	bucket string
}

func NewGCSAdapter(ctx context.Context, bucket, credentialsPath string) (*GCSAdapter, error) {
	opts := []option.ClientOption{}
	if credentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("new gcs client: %w", err)
	}

	return &GCSAdapter{
		client: client,
		bucket: bucket,
	}, nil
}

func (a *GCSAdapter) Close() error {
	return a.client.Close()
}

func (a *GCSAdapter) GetObjectKey(userID, contentID string, originalName string) string {
	return fmt.Sprintf("%s/%s/%s", userID, contentID, originalName)
}

func (a *GCSAdapter) SignUploadURL(ctx context.Context, objectName, contentType string, duration time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(duration),
		ContentType: contentType,
	}

	return a.client.Bucket(a.bucket).SignedURL(objectName, opts)
}

func (a *GCSAdapter) GetObjectAttributes(ctx context.Context, objectName string) (*storage.ObjectAttrs, error) {
	return a.client.Bucket(a.bucket).Object(objectName).Attrs(ctx)
}
