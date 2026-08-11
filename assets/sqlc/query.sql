
-- name: GetTags :one
SELECT * FROM tag
WHERE id = ? LIMIT 1;

-- name: ListTags :many
SELECT * FROM tag
ORDER BY name;

-- name: CreateTag :one
INSERT INTO tag ( name )
VALUES ( ? ) RETURNING *;