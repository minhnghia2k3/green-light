run:
	go run ./cmd/api

healthcheck:
	curl -i http://localhost:4000/v1/healthcheck

PHONY: run