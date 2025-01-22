install:
	go install "github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

gen:
	sqlc generate

create-db:
	psql -U postgres -h localhost -c "CREATE DATABASE patchy"
	psql -U postgres -h localhost -d patchy -f ./db/schema.sql

drop-db:
	psql -U postgres -c "DROP DATABASE patchy"

seed-db:
	psql -U postgres -h localhost -d patchy -f ./db/seeds.sql
