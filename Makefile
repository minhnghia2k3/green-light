include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #


help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run the cmd/api application
run:
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -smtp-sender=${SMTP_SENDER}

## healthcheck: perform GET to api /v1/healthcheck
healthcheck:
	curl -i http://localhost:4000/v1/healthcheck

# for local postgresql
psql:
	psql ${GREENLIGHT_DB_DSN}

## migrate.create: create a new database migration
migrate.create:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext sql -dir ./migrations ${name}

# DSN: postgres://root:secret@localhost:5432/greenlight?sslmode=disable
## migrate.up: apply all up database migrations
migrate.up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} up

## migrate.down: apply down single sequence of migration at a time
migrate.down: confirm
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} down 1

## docker.start: start a specific docker container
docker.start:
	docker start ${CONTAINER_NAME}

## docker.stop: stop a specific docker container
docker.stop:
	docker stop ${CONTAINER_NAME}

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

audit: vendor
	@echo 'Formatting th code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ==================================================================================== #
# DOCUMENTATION
# ==================================================================================== #
swag:
	@echo 'Documenting cmd/api...'
	swag init -g cmd/api/main.go

# ==================================================================================== #
# BUILD
# ==================================================================================== #
build: audit
	@echo 'Building cmd/api...'
	go build -ldflags '-s -w' -o ./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #
exec:
	@echo 'Executing binary cmd/api...'
	./bin/api -db-dsn=${GREENLIGHT_DB_DSN}



PHONY: run healthcheck help confirm psql migrate.create migrate.up migrate.down docker.start docker.stop vendor build