-- name: CreateContent :one
INSERT INTO content (
    id,
    user_id,
    parent_id,
    project_id,
    declared_mime_type,
    bucket,
    object_name,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(parent_id),
    sqlc.arg(project_id),
    sqlc.arg(declared_mime_type),
    sqlc.arg(bucket),
    sqlc.arg(object_name),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
) RETURNING *;

-- name: GetContentByIDAndUser :one
SELECT * FROM content
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: ListContentChildren :many
SELECT * FROM content
WHERE parent_id = sqlc.arg(parent_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit) OFFSET sqlc.arg(result_offset);

-- name: GetContentByBucketAndName :one
SELECT * FROM content
WHERE bucket = sqlc.arg(bucket) AND object_name = sqlc.arg(object_name);

-- name: CompleteContent :one
UPDATE content
SET
    upload_status = 'uploaded',
    effective_mime_type = sqlc.arg(effective_mime_type),
    gcs_uri = sqlc.arg(gcs_uri),
    public_url = sqlc.arg(public_url),
    object_generation = sqlc.arg(object_generation),
    size_bytes = sqlc.arg(size_bytes),
    completed_at = sqlc.arg(completed_at),
    updated_at = sqlc.arg(updated_at)
WHERE bucket = sqlc.arg(bucket) AND object_name = sqlc.arg(object_name)
  AND (object_generation IS NULL OR object_generation <= sqlc.arg(object_generation)::bigint)
RETURNING *;
