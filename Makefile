CORSFLAGS?=
run:
	go run ./cmd/api -cors-trusted-origins=$(CORSFLAGS)

healthcheck:
	curl -i http://localhost:4000/v1/healthcheck

migrate.up:
	migrate -path=./migrations -database=postgres://root:secret@localhost:5432/greenlight?sslmode=disable up


migrate.down:
	migrate -path=./migrations -database=postgres://root:secret@localhost:5432/greenlight?sslmode=disable down 1

docker.start:
	docker start postgres_db

docker.stop:
	docker stop postgres_db

PHONY: run