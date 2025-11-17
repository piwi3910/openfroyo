#!/bin/bash
# Quick start registry with Docker for local testing

set -e

echo "==========================================="
echo "OpenFroyo Module Registry - Docker Test"
echo "==========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if registry container is already running
if docker ps | grep -q froyo-registry; then
    echo "Registry container is already running!"
    echo ""
    docker ps | grep froyo-registry
    echo ""
    echo "To stop: docker stop froyo-registry"
    echo "To remove: docker rm froyo-registry"
    echo "To view logs: docker logs -f froyo-registry"
    exit 0
fi

# Create local directories
echo "Creating local directories..."
mkdir -p ./modules ./data ./logs ./config

# Create config file
echo "Creating configuration..."
cat > ./config/registry.yml << 'EOF'
server:
  listen: "0.0.0.0:8080"
  enable_cors: true
  enable_metrics: true
  enable_healthz: true

storage:
  modules_dir: "/var/lib/froyo-registry/modules"
  checksums_file: "/var/lib/froyo-registry/data/checksums.yml"
  metadata_file: "/var/lib/froyo-registry/data/metadata.yml"
  enable_cache: true
  cache_ttl: 3600

auth:
  enabled: false
  type: "none"

logging:
  level: "info"
  format: "text"
  access_log: true
EOF

# Start registry container
echo ""
echo "Starting registry container..."
docker run -d \
    --name froyo-registry \
    --restart unless-stopped \
    -p 8080:8080 \
    -v "$(pwd)/modules:/var/lib/froyo-registry/modules:ro" \
    -v "$(pwd)/data:/var/lib/froyo-registry/data" \
    -v "$(pwd)/logs:/var/log/froyo-registry" \
    -v "$(pwd)/config:/etc/froyo-registry:ro" \
    golang:1.21-alpine \
    sh -c "apk add --no-cache ca-certificates wget && \
           wget -O /usr/local/bin/froyo-registry https://github.com/piwi3910/openfroyo/releases/download/v1.0.0/froyo-registry-linux-amd64 && \
           chmod +x /usr/local/bin/froyo-registry && \
           /usr/local/bin/froyo-registry --config /etc/froyo-registry/registry.yml"

echo ""
echo "Waiting for registry to be ready..."
sleep 5

# Test connection
if curl -f http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "✓ Registry is running!"
else
    echo "✗ Registry failed to start"
    echo "Check logs: docker logs froyo-registry"
    exit 1
fi

echo ""
echo "==========================================="
echo "Registry Ready!"
echo "==========================================="
echo "Registry URL: http://localhost:8080"
echo "Health: http://localhost:8080/healthz"
echo "Metrics: http://localhost:8080/metrics"
echo ""
echo "Upload modules:"
echo "  cp ../../bin/modules/*.wasm ./modules/"
echo "  docker restart froyo-registry"
echo ""
echo "View logs:"
echo "  docker logs -f froyo-registry"
echo ""
echo "Stop registry:"
echo "  docker stop froyo-registry"
echo ""
echo "Remove registry:"
echo "  docker rm froyo-registry"
echo "==========================================="
