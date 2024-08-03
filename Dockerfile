# Use the official Golang image as the base image
FROM golang:1.22.5-alpine3.19

# Set the Current Working Directory inside the container
WORKDIR /app

# Install dependencies
RUN apk add --no-cache curl tar

# Install golang-migrate
RUN curl -L -o /tmp/migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz \
    && tar -xzvf /tmp/migrate.tar.gz -C /usr/local/bin \
    && rm /tmp/migrate.tar.gz \
    && chmod +x /usr/local/bin/migrate

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the source code into the container
COPY . .

# Compile application
RUN go build -ldflags '-s -w' -o ./bin/api ./cmd/api
RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/linux_amd64/api ./cmd/api

# Command to run the application
CMD ["./bin/api"]

# Expose port (change if your app listens on a different port)
EXPOSE 8080
