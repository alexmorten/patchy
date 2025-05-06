package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/alexmorten/patchy/db"
	"github.com/alexmorten/patchy/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

// func main() {
// 	client := meilisearch.New("http://localhost:7700")

// 	searchRes, err := client.Index("documents").Search("example",
// 		&meilisearch.SearchRequest{
// 			AttributesToHighlight: []string{"Text"},
// 			Limit:                 10,
// 		})
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// 	b, err := searchRes.MarshalJSON()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(string(b))
// }

func main() {
	// Parse command line flags
	domain := flag.String("domain", "", "Domain name for HTTPS with Let's Encrypt (e.g., example.com)")
	flag.Parse()

	// Also check environment variable if flag is not provided
	if *domain == "" {
		*domain = os.Getenv("DOMAIN")
	}

	if *domain != "" {
		fmt.Printf("Starting server with autocert for domain: %s\n", *domain)
	} else {
		fmt.Println("Starting server in HTTP mode (no domain provided)")
	}

	// Get database connection string from environment variable or use default
	connString := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connString == "" {
		connString = "postgresql://localhost:5432/patchy"
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		panic(err)
	}

	// Set reasonable pool size limits
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	querier := db.New(pool)
	s := server.NewServer(*domain, querier)
	fmt.Println(s.ListenAndServe())
}
