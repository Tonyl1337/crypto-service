APP_NAME=crypto-rates

DB_URL=postgres://postgres:postgres@postgres:5432/crypto?sslmode=disable

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

migrate-up:
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(DB_URL)" up

migrate-down:
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(DB_URL)" down

migrate-version:
	docker compose run --rm migrate \
		-path=/migrations \
		-database "$(DB_URL)" version