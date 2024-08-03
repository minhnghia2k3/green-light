FROM golang:1.22.5-alpine3.20

ARG DATABASE_URL
# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . .

# Migrate databasee
RUN migrate -path=./migrations -database=${DATABASE_URL} up

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

# Run
CMD ["./bin/linux_amd64/api"]