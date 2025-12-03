FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

RUN go build -mod=vendor -o main ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
