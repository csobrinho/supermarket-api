# Build stage.
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/supermarket/*

# Runner stage.
FROM alpine:latest AS runner
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
ENTRYPOINT ["./main"]
