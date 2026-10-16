# Build stage
FROM golang:1.22.12-bookworm AS builder

WORKDIR /src

# Download dependencies first for better build caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Compile the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/word-game .

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --no-create-home appuser

WORKDIR /app

COPY --from=builder /out/word-game /app/word-game

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/app/word-game"]
