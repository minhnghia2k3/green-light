run:
	go run ./cmd/api

healthcheck:
	curl -i http://localhost:4000/v1/healthcheck

migrate.up:
	migrate -path=./migrations -database=postgres://root:secret@localhost:5432/greenlight?sslmode=disable up


migrate.down:
	migrate -path=./migrations -database=postgres://root:secret@localhost:5432/greenlight?sslmode=disable down


PHONY: run