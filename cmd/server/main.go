package main

import (
	"context"
	"fmt"

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
	config, err := pgxpool.ParseConfig("postgresql://localhost:5432/patchy")
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
	s := server.NewServer(true, querier)
	fmt.Println(s.ListenAndServe())
}
