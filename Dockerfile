FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tun ./cmd/main.go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/tun .

# app will expect a config, but ConfigMap will mount it here
VOLUME ["/app/config"]

ENV CONFIG_PATH=/app/config/config.toml

EXPOSE 8080

ENTRYPOINT ["/app/tun"]