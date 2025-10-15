# Sequoia Docker Setup

This document explains how to run Sequoia using Docker and Docker Compose.

## Prerequisites

- Docker
- Docker Compose
- At least 1GB of free disk space for data storage

## Quick Start

1. **Clone and navigate to the repository:**
   ```bash
   git clone <repository-url>
   cd sequoia
   ```

2. **Create necessary directories:**
   ```bash
   mkdir -p data blockstore config logs
   ```

3. **Initialize Sequoia configuration:**
   ```bash
   # Build the image first
   docker-compose build
   
   # Initialize configuration (this will create config files)
   docker-compose run --rm sequoia init
   ```

4. **Start Sequoia:**
   ```bash
   docker-compose up -d
   ```

5. **Check logs:**
   ```bash
   docker-compose logs -f sequoia
   ```

## Volume Mappings

The Docker setup uses the following volume mappings:

| Container Path | Host Path | Description |
|----------------|-----------|-------------|
| `/home/sequoia/.sequoia/data` | `./data` | Main data directory - stores all provider data |
| `/home/sequoia/.sequoia/blockstore` | `./blockstore` | IPFS block storage |
| `/home/sequoia/.sequoia/config` | `./config` | Configuration files and wallet |
| `/home/sequoia/.sequoia/logs` | `./logs` | Application logs |

## Configuration

### Environment Variables

You can override configuration using environment variables in the `docker-compose.yml` file:

- `data_directory`: Data storage directory (default: `/home/sequoia/.sequoia/data`)
- `api_config.port`: API port (default: `3333`)
- `api_config.ipfs_port`: IPFS port (default: `4005`)
- `domain`: Provider domain (default: `http://localhost:3333`)
- `total_bytes_offered`: Maximum storage space in bytes (default: `1092616192`)
- `chain_config.rpc_addr`: Blockchain RPC address
- `chain_config.grpc_addr`: Blockchain gRPC address

### Chain Configuration

Update the chain configuration in `docker-compose.yml` to match your network:

```yaml
environment:
  - chain_config.rpc_addr=http://your-rpc-node:26657
  - chain_config.grpc_addr=your-grpc-node:9090
  - chain_config.gas_price=0.02ujkl
  - chain_config.gas_adjustment=1.5
  - chain_config.bech32_prefix=jkl
```

## Commands

### Start Services
```bash
docker-compose up -d
```

### Stop Services
```bash
docker-compose down
```

### View Logs
```bash
docker-compose logs -f sequoia
```

### Execute Commands in Container
```bash
# Initialize configuration
docker-compose run --rm sequoia init

# Check version
docker-compose run --rm sequoia version

# Access shell
docker-compose exec sequoia sh
```

### Restart Services
```bash
docker-compose restart sequoia
```

## Health Checks

The container includes a health check that monitors the API endpoint at `http://localhost:3333/health`. You can check the health status with:

```bash
docker-compose ps
```

## Troubleshooting

### Permission Issues
If you encounter permission issues with the mounted volumes:

```bash
# Fix ownership
sudo chown -R 1000:1000 data blockstore config logs
```

### Port Conflicts
If ports 3333 or 4005 are already in use, update the port mappings in `docker-compose.yml`:

```yaml
ports:
  - "3334:3333"  # Map host port 3334 to container port 3333
  - "4006:4005"  # Map host port 4006 to container port 4005
```

### Data Persistence
All data is stored in the mounted volumes (`./data`, `./blockstore`, `./config`, `./logs`). To completely reset:

```bash
docker-compose down
rm -rf data blockstore config logs
mkdir -p data blockstore config logs
```

## Security Notes

- The container runs as a non-root user (UID 1000)
- Sensitive data (wallet keys) are stored in the `./config` volume
- Ensure proper file permissions on the host system
- Consider using Docker secrets for production deployments

## Production Considerations

For production deployments:

1. Use environment-specific configuration files
2. Set up proper logging and monitoring
3. Configure backup strategies for the data volumes
4. Use Docker secrets for sensitive configuration
5. Consider using a reverse proxy for the API endpoints
6. Set up proper network security and firewall rules
