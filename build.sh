#!/bin/bash

# Get git version and commit
export VERSION=$(echo $(git describe --tags) | sed 's/^v//')
export COMMIT=$(git log -1 --format='%H')

echo "Building with VERSION: $VERSION"
echo "Building with COMMIT: $COMMIT"

# Run docker compose build with the environment variables
docker compose build

echo "Build complete. You can now run 'docker compose up' or 'docker compose restart'"
