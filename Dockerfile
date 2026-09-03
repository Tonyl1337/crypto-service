FROM golang:1.25.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /crypto-service \
    ./cmd/server


FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /crypto-service /app/crypto-service
COPY configs /app/configs

EXPOSE 8080

CMD ["/app/crypto-service"]