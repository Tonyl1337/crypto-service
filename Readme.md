# Crypto Rates Service

A Go service for tracking cryptocurrency rates with a REST API and Telegram bot.

The service periodically fetches BTC and ETH market data from CoinGecko, stores it in PostgreSQL, exposes the data through a REST API, and can automatically send rate updates to Telegram users.

## Features

- BTC and ETH rate tracking
- Automatic rate updates
- PostgreSQL persistence
- REST API
- Telegram bot
- Automatic Telegram subscriptions
- Configurable notification interval
- Graceful shutdown
- Docker and Docker Compose support
- Database migrations
- Unit tests
- OpenAPI specification
- GitHub Actions CI

## Tech Stack

- Go 1.25
- PostgreSQL 16
- pgx
- CoinGecko API
- Telegram Bot API
- Docker
- Docker Compose
- golang-migrate
- OpenAPI 3.0
- GitHub Actions

## Architecture

The application is split into several layers:

```text
                    ┌─────────────────┐
                    │    CoinGecko    │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ CoinGecko Client│
                    └────────┬────────┘
                             │
                             ▼
┌──────────────┐    ┌─────────────────┐    ┌──────────────┐
│   REST API   │◄───│   Rate Service  │───►│  PostgreSQL  │
└──────────────┘    └─────────────────┘    └──────────────┘
                             ▲
                             │
                    ┌────────┴────────┐
                    │                 │
             ┌──────┴──────┐  ┌──────┴───────────┐
             │ Rate Updater│  │   Telegram Bot   │
             └─────────────┘  └────────┬─────────┘
                                       │
                              ┌────────┴─────────┐
                              │ Subscription     │
                              │ Sender           │
                              └──────────────────┘
```

### Project structure

```text
crypto-service/
├── cmd/
│   └── server/
│       └── main.go
├── configs/
├── docs/
│   └── openapi.yaml
├── internal/
│   ├── app/
│   ├── client/
│   │   └── coingecko/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── repository/
│   │   └── postgres/
│   ├── scheduler/
│   ├── service/
│   └── transport/
│       ├── rest/
│       └── telegram/
├── migrations/
├── .github/
│   └── workflows/
│       └── ci.yml
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## REST API

The REST server runs on:

```text
http://localhost:8080
```

### Get latest rates

```http
GET /rates
```

Example:

```bash
curl http://localhost:8080/rates
```

Example response:

```json
[
  {
    "symbol": "BTC",
    "price": 63470,
    "day_low": 63267,
    "day_high": 64329,
    "change_1h": 0.1
  },
  {
    "symbol": "ETH",
    "price": 1883.48,
    "day_low": 1876.31,
    "day_high": 1918.99,
    "change_1h": -0.3
  }
]
```

### Get rates by cryptocurrency

```http
GET /rates/{symbol}
```

Supported symbols:

```text
BTC
ETH
```

Example:

```bash
curl http://localhost:8080/rates/BTC
```

Symbols are case-insensitive:

```bash
curl http://localhost:8080/rates/eth
```

Unsupported symbols return:

```text
400 Bad Request
```

The full API specification is available in:

```text
docs/openapi.yaml
```

## Telegram Bot

The Telegram bot supports the following commands:

| Command | Description |
|---|---|
| `/start` | Show available commands |
| `/rates` | Show current BTC and ETH rates |
| `/rates BTC` | Show BTC rates |
| `/rates ETH` | Show ETH rates |
| `/start_auto N` | Enable automatic updates every N minutes |
| `/stop_auto` | Disable automatic updates |

Aliases `/start-auto N` and `/stop-auto` are also supported.

Example:

```text
/start_auto 5
```

enables automatic Telegram updates every 5 minutes.

## Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Set your Telegram bot token and database password:

```env
TELEGRAM_TOKEN=your_telegram_bot_token
DATABASE_PASSWORD=your_database_password
```

Do not commit the `.env` file.

## Running with Docker Compose

Build and start the application and PostgreSQL:

```bash
docker compose up --build
```

Or run in the background:

```bash
docker compose up -d --build
```

Check running containers:

```bash
docker compose ps
```

Stop the application:

```bash
docker compose down
```

To also remove the PostgreSQL volume:

```bash
docker compose down -v
```

> `docker compose down -v` deletes the database data stored in the Docker volume.

## Database Migrations

Database migrations are located in:

```text
migrations/
```

Current migrations create:

- rates table
- Telegram subscriptions table
- subscription `last_sent_at` field

If migrations are not applied automatically in your Docker setup, install `golang-migrate` and run:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:${DATABASE_PASSWORD}@localhost:5432/crypto?sslmode=disable" \
  up
```

Rollback the latest migration:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:${DATABASE_PASSWORD}@localhost:5432/crypto?sslmode=disable" \
  down 1
```

## Running Locally

Start PostgreSQL:

```bash
docker compose up -d postgres
```

Configure the application and then run:

```bash
go run ./cmd/server
```

The REST API will be available at:

```text
http://localhost:8080
```

## Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run tests with coverage:

```bash
go test ./... -cover
```

## Code Quality

Format the project:

```bash
gofmt -w .
```

Run static analysis:

```bash
go vet ./...
```

Build the project:

```bash
go build ./...
```

## CI

GitHub Actions runs automatically on pushes and pull requests to `main`.

The CI pipeline checks:

1. Go dependencies
2. Go formatting
3. `go vet`
4. Unit tests
5. Go build
6. Docker image build

Workflow:

```text
.github/workflows/ci.yml
```

## Graceful Shutdown

The application handles `SIGINT` and `SIGTERM`.

During shutdown it stops:

- HTTP server
- rate updater
- Telegram polling
- subscription sender

This allows the application to terminate cleanly both locally and inside Docker.

## Rate Updates

The rate updater fetches cryptocurrency data when the service starts and then periodically according to the configured scheduler interval.

The default interval is:

```text
5m
```

Telegram subscription intervals are configured separately by each user through `/start_auto`.

## License

This project is currently unlicensed.