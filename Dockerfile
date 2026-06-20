FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o location-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/location-service .
COPY --from=builder /app/database/migrations ./database/migrations

EXPOSE 8088

CMD ["./location-service"]
