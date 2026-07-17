-- +goose Up
CREATE TABLE content (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES content(id) ON DELETE SET NULL,
    project_id UUID,
    declared_mime_type VARCHAR(255) NOT NULL,
    effective_mime_type VARCHAR(255),
    bucket VARCHAR(255) NOT NULL,
    object_name TEXT NOT NULL,
    gcs_uri TEXT,
    public_url TEXT,
    upload_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    object_generation BIGINT,
    size_bytes BIGINT,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT content_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT content_upload_status_check CHECK (upload_status IN ('pending', 'uploaded', 'failed')),
    CONSTRAINT content_size_bytes_nonnegative CHECK (size_bytes IS NULL OR size_bytes >= 0),
    CONSTRAINT content_bucket_object_name_unique UNIQUE (bucket, object_name)
);

CREATE INDEX content_user_created_at_idx ON content (user_id, created_at DESC);
CREATE INDEX content_parent_id_idx ON content (parent_id);

-- +goose Down
DROP TABLE content;
