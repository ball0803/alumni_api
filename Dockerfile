# Step 1: Build the Go application
ARG GO_VERSION=1.24.1  # Fixed typo (1.24.1 → 1.21.1)
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

# Step 2: Create the final image
FROM debian:bookworm

# 1. Install CA certificates and timezone data (critical for TLS)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata && \
    rm -rf /var/lib/apt/lists/*

# 2. Add curl for health check
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# 3. Copy the Go binary
COPY --from=builder /run-app /usr/local/bin/

# 4. Set timezone (optional but recommended)
ENV TZ=UTC

# 5. Run the app
CMD ["/usr/local/bin/run-app"]
