# Build stage
FROM golang:1.24-bullseye AS builder

# Install build dependencies
# RUN apt-get update && apt-get install -y \
#     sqlite3 \
#     build-essential \
#     && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod and sum files for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w" -o forum .

# Runtime stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

# Create non-root user for security
RUN groupadd -r forum && useradd -r -g forum forum

# Set working directory
WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/forum ./forum

# Copy static files and templates
COPY --from=builder /app/public ./public
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/config/config.json ./config/config.json
COPY --from=builder /app/pkg/mydb.db ./pkg/mydb.db

# Copy database setup files if they exist
COPY --from=builder /app/internal/setup.sql ./internal/setup.sql

# Create internal directory and set permissions
RUN mkdir -p /app/internal && \
    chown -R forum:forum /app && \
    chmod +x ./forum

# Switch to non-root user
USER forum

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ || exit 1

# Run the application
CMD ["./forum"]