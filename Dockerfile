# Stage 1 — Build the Go binary
FROM golang:1.24-alpine AS builder

# Enable static binary
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

RUN go mod download

# Copy the entire project
COPY . .

# Build the Go API
RUN go build -ldflags="-s -w" -o server ./cmd/grpcapi

# Stage 2 — Final ultra-slim image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder stage
COPY --from=builder /app/server /server

# Expose your Go API port
EXPOSE 50051

# Run the server
CMD ["/server"]