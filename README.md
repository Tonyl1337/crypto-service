# Crypto Rates Service

A Go service for tracking cryptocurrency rates with a REST API and Telegram bot.

The service fetches cryptocurrency market data from CoinGecko, stores rates in PostgreSQL, exposes them through a REST API, and supports automatic Telegram notifications.

## Features

- BTC and ETH periodic rate tracking
- Lookup of additional cryptocurrencies through CoinGecko
- Automatic rate updates on startup and at a configurable interval
- PostgreSQL persistence
- REST API
- Health check endpoint
- OpenAPI specification
- Telegram bot
- Per-cryptocurrency Telegram subscriptions
- Configurable notification intervals
- Graceful shutdown
- Docker and Docker Compose
- Automatic database migrations
- Unit tests
- PostgreSQL integration tests
- GitHub Actions CI

## Tech Stack

- Go 1.25
- PostgreSQL 16
- pgx
- CoinGecko API
- Telegram Bot API
- Docker / Docker Compose
- golang-migrate
- OpenAPI 3.0
- GitHub Actions

## Architecture

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
                               ┌───────┴──────────┐
                               │ Subscription     │
                               │ Sender           │
                               └──────────────────┘
```

## Project Structure

```text
crypto-service/
├── api/
│   └── openapi.yaml
├── cmd/
│   └── server/
├── configs/
├── internal/
│   ├── app/
│   ├── client/coingecko/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── repository/postgres/
│   ├── scheduler/
│   ├── service/
│   └── transport/
│       ├── rest/
│       └── telegram/
├── migrations/
├── .github/workflows/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## REST API

The API is available by default at:

```text
http://localhost:8080
```

### Health Check

```http
GET /health
```

Example:

```bash
curl http://localhost:8080/health
```

Response:

```json
{"status":"ok"}
```

### Get Latest Rates

```http
GET /rates
```

Returns the latest stored BTC and ETH rates.

Example:

```bash
curl http://localhost:8080/rates
```

Example response:

```json
[
  {
    "symbol": "BTC",
    "price": 82994,
    "day_low": 82581,
    "day_high": 84265,
    "change_1h": -0.1912
  },
  {
    "symbol": "ETH",
    "price": 2660.39,
    "day_low": 2636.99,
    "day_high": 2717.37,
    "change_1h": -0.4981
  }
]
```

### Get Cryptocurrency Rate

```http
GET /rates/{symbol}
```

The endpoint can resolve cryptocurrency symbols or CoinGecko IDs.

Examples:

```bash
curl http://localhost:8080/rates/BTC
curl http://localhost:8080/rates/ETH
curl http://localhost:8080/rates/SOL
curl http://localhost:8080/rates/bitcoin
```

The service resolves the cryptocurrency through CoinGecko and retrieves its current market data.

## OpenAPI

The OpenAPI 3.0 specification is available at:

```text
api/openapi.yaml
```

It is also served by the application:

```text
http://localhost:8080/openapi.yaml
```

## Telegram Bot

Available commands:

| Command | Description |
|---|---|
| `/start` | Show available commands |
| `/rates` | Show current BTC and ETH rates |
| `/rates BTC` | Show a selected cryptocurrency |
| `/rates SOL` | Resolve and show another cryptocurrency |
| `/start_auto SOL 10` | Send SOL rate every 10 minutes |
| `/stop_auto SOL` | Disable automatic SOL updates |
| `/stop_auto` | Disable all automatic updates |

Hyphen aliases such as `/start-auto` and `/stop-auto` are also supported.

Each Telegram chat can maintain its own cryptocurrency subscriptions and notification intervals.

## Configuration

Create the environment file:

```bash
cp .env.example .env
```

Configure:

```env
TELEGRAM_TOKEN=your_telegram_bot_token
DATABASE_PASSWORD=your_database_password
```

Example application configuration is available in:

```text
configs/config.example.yaml
```

Secrets must not be committed to Git.

Local `.env`, `configs/config.yaml`, and `configs/config.docker.yaml` files are ignored.

## Running with Docker Compose

Start the complete stack:

```bash
docker compose up -d --build
```

Docker Compose starts:

- PostgreSQL
- migration service
- Crypto Rates Service

PostgreSQL must become healthy before migrations run. The application starts after migrations complete successfully.

Check status:

```bash
docker compose ps -a
```

View logs:

```bash
docker compose logs -f app
```

Stop:

```bash
docker compose down
```

Remove containers and database volume:

```bash
docker compose down -v
```

> `docker compose down -v` permanently deletes data stored in the PostgreSQL Docker volume.

## Database Migrations

Migrations are stored in:

```text
migrations/
```

They manage the `rates` and `telegram_subscriptions` tables and their subsequent schema changes.

Docker Compose applies pending migrations automatically.

Manual migration:

```bash
migrate \
  -path migrations \
  -database "postgres://postgres:${DATABASE_PASSWORD}@localhost:5432/crypto?sslmode=disable" \
  up
```

Rollback one migration:

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

Create `.env`:

```bash
cp .env.example .env
```

Apply migrations and run:

```bash
go run ./cmd/server
```

## Makefile

Useful commands:

```bash
make run
make test
make test-cover
make fmt
make vet
make build
make check
make docker-up
make docker-down
make docker-logs
```

## Tests

### Unit Tests

Run:

```bash
go test ./...
```

The project includes tests for:

- CoinGecko client and resolver
- rate service
- subscription service
- schedulers
- REST handlers
- REST responses
- Telegram handlers
- Telegram subscription logic

### PostgreSQL Integration Tests

Repository integration tests run against a real PostgreSQL database.

Create a test database and apply migrations, then run:

```bash
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/crypto_test?sslmode=disable" \
go test -tags=integration \
  ./internal/repository/postgres/... \
  -v
```

Integration tests cover the rate and subscription PostgreSQL repositories, including saving, querying, upserts, automatic subscription selection, marking notifications as sent, and deletion.

## Code Quality

Run the complete local check:

```bash
go vet ./...
go test ./...
go build ./...
```

Or:

```bash
make check
```

## CI

GitHub Actions runs on pushes and pull requests to `main`.

The pipeline contains separate jobs for:

- unit tests
- `go vet`
- Go build
- PostgreSQL integration tests

The integration job starts PostgreSQL 16, applies database migrations, and executes tests using the `integration` build tag.

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

This allows the application to shut down cleanly both locally and inside Docker.

## Rate Updates

BTC and ETH are fetched when the service starts and periodically according to the configured scheduler interval.

Default interval:

```text
5m
```

Other cryptocurrencies can be resolved and fetched on demand through the REST API or Telegram bot.

Telegram automatic notifications have independent per-subscription intervals.

Example:

```text
/start_auto SOL 10
```

This subscribes the Telegram chat to SOL updates every 10 minutes.

## Security

Secrets are supplied through environment variables and local configuration.

Never commit:

- Telegram bot tokens
- database passwords
- `.env`
- local production configuration

The repository contains only example configuration with placeholder values.

## License

This project is currently unlicensed.