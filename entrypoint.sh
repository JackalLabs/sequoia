#!/bin/sh

# Only run permission fixes if we're root (during container startup)
if [ "$(id -u)" = "0" ]; then
    # Ensure the sequoia user owns the mounted volumes
    if [ -d "/home/sequoia/.sequoia/data" ]; then
        chown -R sequoia:sequoia /home/sequoia/.sequoia/data
    fi

    if [ -d "/home/sequoia/.sequoia/blockstore" ]; then
        chown -R sequoia:sequoia /home/sequoia/.sequoia/blockstore
    fi

    if [ -d "/home/sequoia/.sequoia/config" ]; then
        chown -R sequoia:sequoia /home/sequoia/.sequoia/config
    fi

    if [ -d "/home/sequoia/.sequoia/logs" ]; then
        chown -R sequoia:sequoia /home/sequoia/.sequoia/logs
    fi

    # Set proper permissions
    chmod -R 755 /home/sequoia/.sequoia

    # Switch to sequoia user and execute the command
    exec gosu sequoia "$@"
else
    # Already running as sequoia user, execute directly
    exec "$@"
fi
