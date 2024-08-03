# Use a valid golang image as the builder
FROM golang:1.22.5-alpine3.19 as builder

# Install necessary tools
RUN apk add --no-cache curl tar

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar -xz -C /usr/local/bin

WORKDIR /app
COPY . .

RUN go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian10
COPY --from=builder /bin/api /bin/api
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

CMD ["/bin/api"]
