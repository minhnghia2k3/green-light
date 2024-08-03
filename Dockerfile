# Use the official Golang image as the base image
FROM golang:1.22.5-alpine3.19

WORKDIR /app

# Install golang-migrate
RUN curl -L -o /tmp/migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz \
    && tar -xzvf /tmp/migrate.tar.gz -C /usr/local/bin \
    && rm /tmp/migrate.tar.gz \
    && chmod +x /usr/local/bin/migrate

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

# Compile application
RUN go build -ldflags '-s -w' -o ./bin/api ./cmd/api
RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o=./bin/linux_amd64/api ./cmd/api

CMD ["./bin/api"]