install:
	go install "github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

gen:
	sqlc generate
