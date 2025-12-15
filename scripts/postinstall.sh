#!/bin/sh
set -e

# Create otlp-network user if it doesn't exist
if ! getent passwd otlp-network >/dev/null; then
    echo "Creating system user otlp-network"
    useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin otlp-network
fi
