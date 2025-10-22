#!/bin/bash

# Get git version and commit
export VERSION=$(git describe --tags --always 2>/dev/null | sed 's/^v//')
export COMMIT=$(git log -1 --format='%H')

echo "Building with VERSION: $VERSION"
echo "Building with COMMIT: $COMMIT"

# Run docker compose build with the environment variables
docker compose build

echo "Build complete. You can now run 'docker compose up' or 'docker compose restart'"
