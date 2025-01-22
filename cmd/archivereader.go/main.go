package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alexmorten/patchy/db"
	"github.com/jackc/pgx/v5"
)

const messageBeginning = "From mboxrd@z"

func main() {
	conn, err := pgx.Connect(context.Background(), "postgresql://localhost:5432/patchy")
	if err != nil {
		panic(err)
	}
	querier := db.New(conn)

	f, err := os.Open("archive.utf8.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	text := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, messageBeginning) {
			insertMessage(text, querier)
			text = ""
			continue
		}
		text += line + "\n"

	}
}

func insertMessage(text string, querier *db.Queries) {
	_, err := querier.CreateDocument(context.Background(), db.CreateDocumentParams{
		Text: text,
		Url:  "",
	})

	if err != nil {
		fmt.Println(err)
	}
}
