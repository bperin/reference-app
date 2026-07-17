package content

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("content not found")
var ErrStaleGeneration = errors.New("stale object generation")

type Content struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	ParentID          *uuid.UUID
	ProjectID         *uuid.UUID
	DeclaredMimeType  string
	EffectiveMimeType *string
	Bucket            string
	ObjectName        string
	GcsUri            *string
	PublicUrl         *string
	UploadStatus      string
	ObjectGeneration  *int64
	SizeBytes         *int64
	CompletedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
