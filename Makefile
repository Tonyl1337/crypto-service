APP_NAME=crypto-rates

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

build:
	go build -o bin/$(APP_NAME) ./cmd/server