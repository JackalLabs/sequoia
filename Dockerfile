# Build stage
FROM golang:1.24.0-bullseye AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    make \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy source code
COPY go.mod go.sum ./


# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1000 sequoia && \
    useradd -r -u 1000 -g sequoia -s /bin/sh sequoia

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sequoia .

# Copy required shared libraries from wasmvm
RUN mkdir -p /usr/local/lib
COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm@v1.2.6/internal/api/libwasmvm.aarch64.so /usr/local/lib/
RUN ldconfig

# Create necessary directories
RUN mkdir -p /home/sequoia/.sequoia/data && \
    mkdir -p /home/sequoia/.sequoia/blockstore && \
    chown -R sequoia:sequoia /home/sequoia && \
    chown -R sequoia:sequoia /app

# Switch to non-root user
USER sequoia

# Expose ports
EXPOSE 3333 4005 4001

# Set environment variables
ENV HOME=/home/sequoia
ENV SEQUOIA_HOME=/home/sequoia/.sequoia

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3333/health || exit 1

# Default command
CMD ["./sequoia", "start"]
