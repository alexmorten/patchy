package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func setupTestDB(t *testing.T) *Queries {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres@localhost:5432/patchy")
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}
	t.Cleanup(func() {
		conn.Close(context.Background())
	})
	return New(conn)
}

func TestCreateDocument(t *testing.T) {
	q := setupTestDB(t)
	ctx := context.Background()

	// First insertion
	doc1, err := q.CreateDocument(ctx, CreateDocumentParams{
		Text:      "Original text",
		Url:       "http://example.com",
		MessageID: "test@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Second insertion with same message_id should update
	doc2, err := q.CreateDocument(ctx, CreateDocumentParams{
		Text:      "Updated text",
		Url:       "http://example.com/updated",
		MessageID: "test@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	// Verify it's the same record
	if doc1.ID != doc2.ID {
		t.Errorf("Expected same ID for upserted document, got %d and %d", doc1.ID, doc2.ID)
	}
	if doc2.Text != "Updated text" {
		t.Errorf("Expected updated text, got %s", doc2.Text)
	}
	if doc2.Url != "http://example.com/updated" {
		t.Errorf("Expected updated URL, got %s", doc2.Url)
	}
}
