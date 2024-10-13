# Stage 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p logs  # Создаем папку logs для логов ибо так использую

CMD ["./main"]

# запускайте используй подключение к сети бд example: docker run --rm --network song-network --env-file .env song-library