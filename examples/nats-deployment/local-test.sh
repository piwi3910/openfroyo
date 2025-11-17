#!/bin/bash
# Quick start NATS server for local testing

set -e

echo "=========================================="
echo "OpenFroyo NATS Local Test Setup"
echo "=========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if NATS container is already running
if docker ps | grep -q nats-local; then
    echo "NATS container is already running!"
    echo ""
    docker ps | grep nats-local
    echo ""
    echo "To stop: docker stop nats-local"
    echo "To remove: docker rm nats-local"
    exit 0
fi

# Start NATS server
echo "Starting NATS server..."
docker run -d \
    --name nats-local \
    --restart unless-stopped \
    -p 4222:4222 \
    -p 8222:8222 \
    -p 6222:6222 \
    nats:latest \
    -js \
    --http_port 8222

echo ""
echo "Waiting for NATS to be ready..."
sleep 3

# Test connection
if curl -f http://localhost:8222/varz > /dev/null 2>&1; then
    echo "✓ NATS server is running!"
else
    echo "✗ NATS server failed to start"
    echo "Check logs: docker logs nats-local"
    exit 1
fi

echo ""
echo "=========================================="
echo "NATS Server Ready!"
echo "=========================================="
echo "NATS URL: nats://localhost:4222"
echo "HTTP Monitoring: http://localhost:8222"
echo ""
echo "Test the connection:"
echo "  curl http://localhost:8222/varz"
echo ""
echo "View logs:"
echo "  docker logs -f nats-local"
echo ""
echo "Stop server:"
echo "  docker stop nats-local"
echo ""
echo "Remove server:"
echo "  docker rm nats-local"
echo ""
echo "Use with froyo CLI:"
echo "  froyo apply stack.ofy --nats-server nats://localhost:4222"
echo "=========================================="
