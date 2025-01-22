package main

import (
	"fmt"
	"os"

	"github.com/meilisearch/meilisearch-go"
)

func main() {
	client := meilisearch.New("http://localhost:7700")

	searchRes, err := client.Index("documents").Search("uring",
		&meilisearch.SearchRequest{
			AttributesToHighlight: []string{"Text"},
			Limit:                 10,
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b, err := searchRes.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
