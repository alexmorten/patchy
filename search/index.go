package search

import (
	"context"
	"fmt"
	"log"
	"time"

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
	listAllDocs(queries, func(docs []db.Doc) {

		task, err := index.AddDocuments(docs)
		if err != nil {
			log.Fatalf("Failed to index documents: %v\n", err)
		}

		fmt.Printf("Documents indexed with task UID: %d\n", task.TaskUID)
	})
}

func listAllDocs(queries *db.Queries, f func(docs []db.Doc)) {
	var id int32
	hasAlreadySlept := false
	for {
		docs, err := queries.ListDocuments(context.Background(), db.ListDocumentsParams{
			Limit:  1000,
			Offset: id,
		})
		if err != nil {
			log.Fatalf("Failed to list documents: %v\n", err)
		}
		if len(docs) == 0 {
			if hasAlreadySlept {
				return
			}
			fmt.Printf("waiting for more...")
			time.Sleep(5 * time.Second)
			hasAlreadySlept = true
			continue
		}

		hasAlreadySlept = false

		f(docs)
		id = int32(docs[len(docs)-1].ID)
	}
}

func connectToDatabase() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgresql://localhost:5432/patchy")
	if err != nil {
		return nil, err
	}
	return conn, nil
}
