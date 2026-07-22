# CryptoCheck API

Go API and background worker for CryptoCheck. It provides authentication, users,
posts, comments, follows, reactions, news feeds, and the token scanner.

## Run locally

```bash
cp .env.example .env
go run cmd/api/main.go
```

The API listens on `http://localhost:8080`.

## Run dependencies with Docker

```bash
cd deployment
docker compose up --build
```

This starts MongoDB, RabbitMQ, the API, and the worker. Configure secrets and
external API keys in the root `.env` file; do not commit it.

## Test

```bash
go test ./...
```
