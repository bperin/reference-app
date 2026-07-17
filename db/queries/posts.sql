-- name: CreatePost :one
INSERT INTO posts (
    id,
    author_id,
    title,
    content,
    published,
    created_at,
    updated_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(author_id),
    sqlc.arg(title),
    sqlc.arg(content),
    sqlc.arg(published),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
)
RETURNING *;

-- name: GetPostByID :one
SELECT *
FROM posts
WHERE id = sqlc.arg(id);

-- name: UpdatePost :one
UPDATE posts
SET
    title = sqlc.arg(title),
    content = sqlc.arg(content),
    published = sqlc.arg(published),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND author_id = sqlc.arg(author_id)
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = sqlc.arg(id) AND author_id = sqlc.arg(author_id);

-- name: ListPublishedPosts :many
SELECT *
FROM posts
WHERE published = TRUE
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit)
OFFSET sqlc.arg(result_offset);

-- name: ListAllPosts :many
SELECT *
FROM posts
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit)
OFFSET sqlc.arg(result_offset);
