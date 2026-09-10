# Crypto Rates Service

A Go service for tracking cryptocurrency rates with a REST API and Telegram bot.

The service periodically fetches BTC and ETH market data from CoinGecko, stores it in PostgreSQL, exposes the data through a REST API, and can automatically send rate updates to Telegram users.

## Features

- BTC and ETH rate tracking
- Automatic rate updates on startup and at a configurable interval
- PostgreSQL persistence
- REST API
- Telegram bot
- Automatic Telegram subscriptions
- Configurable notification interval for each Telegram user
- Graceful shutdown
- Docker and Docker Compose support
- Automatic database migrations with Docker Compose
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

### Project Structure

```text
crypto-service/
├── cmd/
│   └── server/
│       └── main.go
├── configs/
│   └── config.example.yaml
├── docs/
│   └── openapi.yaml
├── internal/
│   ├── app/
│   ├── client/
│   │   └── coingecko/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── logger/
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
├── .env.example
├── .gitignore
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## REST API

The REST server runs on:

```text
http://localhost:8080
```

### Get Latest Rates

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

### Get Rates by Cryptocurrency

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

This enables automatic Telegram updates every 5 minutes.

Each Telegram user can configure their own notification interval.

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

Application configuration examples are available in:

```text
configs/config.example.yaml
```

Sensitive values such as the Telegram bot token and database password are supplied through environment variables and must not be committed to the repository.

The `.env` file and local configuration files are ignored by Git.

## Running with Docker Compose

Build and start the complete application stack:

```bash
docker compose up --build
```

Or run it in the background:

```bash
docker compose up -d --build
```

Docker Compose starts:

- PostgreSQL
- database migration service
- Crypto Rates Service

The application waits for PostgreSQL to become healthy and for database migrations to complete before starting.

Check container status:

```bash
docker compose ps -a
```

A successful startup should show:

```text
PostgreSQL        running (healthy)
Migration service exited with code 0
Application       running
```

View application logs:

```bash
docker compose logs -f app
```

Stop the application:

```bash
docker compose down
```

To also remove the PostgreSQL volume:

```bash
docker compose down -v
```

> `docker compose down -v` permanently deletes the database data stored in the Docker volume.

## Database Migrations

Database migrations are located in:

```text
migrations/
```

Current migrations create and manage:

- `rates` table
- `telegram_subscriptions` table
- subscription `last_sent_at` field

When the application is started with Docker Compose, migrations are applied automatically before the application starts.

The `migrate` service waits for PostgreSQL to become healthy, applies all pending migrations, and exits successfully. The application starts only after the migration service completes.

Start the complete stack:

```bash
docker compose up -d --build
```

Check migration status through Docker Compose:

```bash
docker compose ps -a
```

The migration container should finish with exit code `0`.

### Manual Migrations

If needed, migrations can also be applied manually using `golang-migrate`:

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

Create the local environment file:

```bash
cp .env.example .env
```

Configure the required environment variables and then run:

```bash
go run ./cmd/server
```

The REST API will be available at:

```text
http://localhost:8080
```

## Makefile

The project provides several useful Makefile commands.

Run the application:

```bash
make run
```

Run tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-cover
```

Format the code:

```bash
make fmt
```

Run static analysis:

```bash
make vet
```

Build the application:

```bash
make build
```

Run formatting, static analysis, tests, and build together:

```bash
make check
```

Start the Docker Compose stack:

```bash
make docker-up
```

Stop the Docker Compose stack:

```bash
make docker-down
```

View application logs:

```bash
make docker-logs
```

## Tests

Run all tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-cover
```

You can also run tests directly with Go:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

The project includes tests for core application logic, including:

- CoinGecko client
- rate service
- scheduler
- REST handlers
- REST responses
- Telegram command handlers
- Telegram subscription logic

## Code Quality

Format the project:

```bash
make fmt
```

Run static analysis:

```bash
make vet
```

Build the application:

```bash
make build
```

Run the complete local check:

```bash
make check
```

The equivalent Go commands are:

```bash
gofmt -w .
go vet ./...
go test ./...
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

This ensures that changes pushed to the repository compile successfully, pass tests, follow Go formatting rules, and produce a valid Docker image.

## Graceful Shutdown

The application handles `SIGINT` and `SIGTERM`.

During shutdown it stops:

- HTTP server
- rate updater
- Telegram polling
- subscription sender

This allows the application to terminate cleanly both locally and inside Docker.

## Rate Updates

The rate updater fetches BTC and ETH market data when the service starts and then periodically according to the configured scheduler interval.

The default update interval is:

```text
5m
```

Telegram subscription intervals are configured separately by each user through:

```text
/start_auto N
```

For example:

```text
/start_auto 10
```

enables automatic rate notifications every 10 minutes for that Telegram chat.

## OpenAPI

The REST API is documented using OpenAPI 3.0.

Specification:

```text
docs/openapi.yaml
```

The specification documents:

- `GET /rates`
- `GET /rates/{symbol}`
- rate response schema
- error response schema

## Security

Secrets are not stored in the repository.

Local sensitive values are provided through:

```text
.env
```

The following files are excluded from Git:

```text
.env
configs/config.yaml
configs/config.docker.yaml
```

Example configuration files contain placeholders only and can safely be committed.

Never commit a real Telegram bot token or database password.

## License

This project is currently unlicensed.