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

# Set version and commit environment variables for build
RUN export VERSION=$(echo $(git describe --tags) | sed 's/^v//') && \
    export COMMIT=$(git log -1 --format='%H') && \
    make build

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    wget \
    gosu \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1000 sequoia && \
    useradd -r -u 1000 -g sequoia -s /bin/sh sequoia

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sequoia .

# Copy entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Copy required shared libraries from wasmvm
RUN mkdir -p /usr/local/lib
COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm@v1.2.6/internal/api/libwasmvm.*.so /usr/local/lib/
RUN ldconfig

# Create necessary directories with proper permissions
RUN mkdir -p /home/sequoia/.sequoia/data && \
    mkdir -p /home/sequoia/.sequoia/blockstore && \
    mkdir -p /home/sequoia/.sequoia/config && \
    mkdir -p /home/sequoia/.sequoia/logs && \
    chown -R sequoia:sequoia /home/sequoia && \
    chown -R sequoia:sequoia /app && \
    chmod -R 755 /home/sequoia/.sequoia

# Note: User switching is handled by the entrypoint script

# Expose ports
EXPOSE 3333 4005 4001

# Set environment variables
ENV HOME=/home/sequoia
ENV SEQUOIA_HOME=/home/sequoia/.sequoia

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3333/health || exit 1

# Set entrypoint and default command
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["./sequoia", "start"]
