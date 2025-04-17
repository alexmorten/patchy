-- name: GetDocumentByID :one
SELECT * FROM docs
WHERE id = $1 LIMIT 1;

-- name: ListDocuments :many
SELECT * FROM docs
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CreateDocument :one
INSERT INTO docs (
  text, url, message_id
) VALUES (
  $1, $2, $3
)
ON CONFLICT (message_id) 
DO UPDATE SET
  text = EXCLUDED.text,
  url = EXCLUDED.url
RETURNING *;
