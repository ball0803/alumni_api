# Step 1: Build the Go application
ARG GO_VERSION=1.24.1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

# Step 2: Create the final image
FROM debian:bookworm

# Copy the Go binary from the builder image
COPY --from=builder /run-app /usr/local/bin/

# Set the default command to run the app
CMD ["/usr/local/bin/run-app"]

