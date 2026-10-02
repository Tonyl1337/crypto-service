FROM golang:1.25.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /crypto-service \
    ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux go install \
    -tags 'postgres' \
    github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3


FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /crypto-service /app/crypto-service
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY configs /app/configs
COPY api /app/api
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/crypto-service"]