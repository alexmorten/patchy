package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alexmorten/patchy/db"
	"github.com/alexmorten/patchy/internal/email"
	"github.com/jackc/pgx/v5"
)

const messageBeginning = "From mboxrd@z"

func main() {
	filename := "archive.utf8.txt"

	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	conn, err := pgx.Connect(context.Background(), "postgresql://localhost:5432/patchy")
	if err != nil {
		panic(err)
	}
	querier := db.New(conn)

	f, err := os.Open(filename)
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
	messageID := email.ExtractMessageID(text)
	if messageID == "" {
		fmt.Println("Warning: No Message-ID found in message:")
		fmt.Println(text)
		return
	}

	_, err := querier.CreateDocument(context.Background(), db.CreateDocumentParams{
		Text:      text,
		Url:       messageID,
		MessageID: messageID,
	})

	if err != nil {
		fmt.Println(err)
	}
}
