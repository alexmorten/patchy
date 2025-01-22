package search

import (
	"context"
	"fmt"
	"log"

	"github.com/alexmorten/patchy/db" // Replace with the actual import path of your db package

	"github.com/jackc/pgx/v5"
	"github.com/meilisearch/meilisearch-go"
)

func IndexDocuments() {
	// Connect to the database
	conn, err := connectToDatabase()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Create a new Queries instance
	queries := db.New(conn)

	// List all documents from the database
	docs, err := queries.ListDocuments(context.Background())
	if err != nil {
		log.Fatalf("Failed to list documents: %v\n", err)
	}

	// Initialize Meilisearch client
	client := meilisearch.New("http://localhost:7700")

	// Prepare documents for Meilisearch
	// var meiliDocs []map[string]interface{}
	// for _, doc := range docs {
	// 	meiliDocs = append(meiliDocs, map[string]interface{}{
	// 		"id":   doc.ID,
	// 		"text": doc.Text,
	// 		"url":  doc.Url,
	// 	})
	// }

	// Index documents in Meilisearch
	index := client.Index("documents")
	task, err := index.AddDocuments(docs)
	if err != nil {
		log.Fatalf("Failed to index documents: %v\n", err)
	}

	fmt.Printf("Documents indexed with task UID: %d\n", task.TaskUID)
}

func connectToDatabase() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgresql://localhost:5432/patchy")
	if err != nil {
		return nil, err
	}
	return conn, nil
}
