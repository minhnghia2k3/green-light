# Use the official Golang image as the base image
FROM golang:1.22.5-alpine3.19

# Install dependencies
RUN apk add --no-cache git

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz -C /usr/local/bin \
    && chmod +x /usr/local/bin/migrate

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -ldflags '-s -w' -o ./bin/api ./cmd/api

# Command to run the migrations and start the app
CMD ["sh", "-c", "migrate -path=./migrations -database=${DATABASE_URL} up && ./bin/api"]

# Expose port (change if your app listens on a different port)
EXPOSE 4000
