FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notification-service ./cmd/notification-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/notification-service .
CMD ["./notification-service"]