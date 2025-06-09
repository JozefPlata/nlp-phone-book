# Build stage
FROM golang:1.24-alpine AS builder

# Create app directory with proper permissions
RUN mkdir -p /app/data && \
    adduser -D -g '' appuser && \
    chown -R appuser:appuser /app/data

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy module files first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o main .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/views ./views

# Set permissions
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

EXPOSE 8080

USER appuser

CMD ["./main"]
