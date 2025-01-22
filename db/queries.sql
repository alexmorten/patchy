-- name: GetDocumentByID :one
SELECT * FROM docs
WHERE id = $1 LIMIT 1;

-- name: ListDocuments :many
SELECT * FROM docs
ORDER BY id;

-- name: CreateDocument :one
INSERT INTO docs (
  text, url
) VALUES (
  $1, $2
)
RETURNING *;
