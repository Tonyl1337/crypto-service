APP_NAME=crypto-rates

.PHONY: run test test-cover fmt vet build check docker-up docker-down docker-logs migrate-up migrate-down

run:
	go run ./cmd/server

test:
	go test ./...

test-cover:
	go test ./... -cover

fmt:
	gofmt -w .

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/server

check: fmt vet test build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://postgres:$${DATABASE_PASSWORD}@postgres:5432/crypto?sslmode=disable" \
		down 1