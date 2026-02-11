# Build stage.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /app

# Cache Go modules.
COPY go.mod go.sum ./
RUN go mod download

# Build for target platform.
ARG TARGETOS TARGETARCH
ARG VERSION=dev
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-X main.version=${VERSION}" \
    -o main cmd/supermarket/*

# Runner stage - use scratch for minimal image.
FROM scratch AS runner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
