## Module Registry Deployment

This directory contains deployment stacks for the OpenFroyo Module Registry - a centralized HTTP server that distributes WASM modules to agents.

## Overview

The Module Registry provides:
- ✅ Centralized module distribution
- ✅ Automatic checksum verification
- ✅ Module caching on agents
- ✅ Access control (API keys)
- ✅ Health monitoring
- ✅ Prometheus metrics

## Quick Start

### 1. Deploy Registry Server (10 min)

```bash
cd examples/registry-deployment

# Update inventory with your server IP
vim inventory.yml

# Deploy registry
froyo apply systemd-stack.ofy
```

**What you get:**
- froyo-registry binary installed
- Systemd service running as froyo-registry user
- HTTP server on port 8080
- Module storage in /var/lib/froyo-registry/modules

### 2. Upload Modules (5 min)

```bash
# Build modules first
cd ../../
make build-wasm

# Upload to registry
cd examples/registry-deployment
./upload-modules.sh 192.168.1.60
```

### 3. Configure Agents (5 min)

Update agent configuration to use registry:

**`/etc/froyo/agent.yml`:**
```yaml
execution:
  module_registry_url: "http://192.168.1.60:8080"
  module_registry_api_key: ""  # Optional, if auth enabled
```

Restart agents:
```bash
systemctl restart froyo-agent
```

## Architecture

```
┌──────────────┐
│ Orchestrator │
│   (froyo)    │
└──────┬───────┘
       │
   ┌───▼────────┐
   │    NATS    │
   └───┬────────┘
       │
       │ Task: {module: "generic/file"}
       │
   ┌───▼──────────┐
   │    Agent     │
   │              │
   │ 1. Receive task
   │ 2. Check cache for module
   │ 3. If not cached:
   │    GET http://registry:8080/modules/generic-file/metadata
   │    ├─► Get checksum
   │    GET http://registry:8080/modules/generic-file/download
   │    ├─► Download WASM
   │    └─► Verify checksum
   │ 4. Execute module
   │
   └──────┬───────┘
          │
          │ GET /modules/generic-file/download
          │
   ┌──────▼──────────┐
   │ Module Registry │
   │  HTTP Server    │
   │                 │
   │ /modules        │ ← List all modules
   │ /modules/{name}/metadata │ ← Get metadata
   │ /modules/{name}/download │ ← Download WASM
   │ /healthz        │ ← Health check
   │ /metrics        │ ← Prometheus metrics
   │
   └─────────────────┘
```

## Registry API

### List Modules

```bash
curl http://registry:8080/modules
```

**Response:**
```json
{
  "modules": [
    {
      "name": "generic-file",
      "version": "1.0.0",
      "checksum": "sha256:abc123...",
      "size": 1234567
    }
  ],
  "count": 1
}
```

### Get Module Metadata

```bash
curl http://registry:8080/modules/generic-file/metadata
```

**Response:**
```json
{
  "name": "generic-file",
  "version": "1.0.0",
  "description": "File operations module",
  "checksum": "sha256:abc123...",
  "size": 1234567,
  "tags": ["file", "generic"]
}
```

### Download Module

```bash
curl http://registry:8080/modules/generic-file/download -o generic-file.wasm
```

**Headers:**
```
Content-Type: application/wasm
X-Module-Checksum: sha256:abc123...
X-Module-Version: 1.0.0
```

### Health Check

```bash
curl http://registry:8080/healthz
```

**Response:**
```json
{
  "status": "healthy",
  "time": 1234567890
}
```

### Metrics

```bash
curl http://registry:8080/metrics
```

**Response (Prometheus format):**
```
# HELP froyo_registry_module_downloads_total Total module downloads
# TYPE froyo_registry_module_downloads_total counter
froyo_registry_module_downloads_total 42

# HELP froyo_registry_bytes_served_total Total bytes served
# TYPE froyo_registry_bytes_served_total counter
froyo_registry_bytes_served_total 12345678
```

## Configuration

### Registry Configuration (`/etc/froyo-registry/registry.yml`)

```yaml
server:
  listen: "0.0.0.0:8080"
  read_timeout: 30s
  write_timeout: 30s
  enable_cors: true
  enable_metrics: true
  enable_healthz: true

storage:
  modules_dir: "/var/lib/froyo-registry/modules"
  checksums_file: "/var/lib/froyo-registry/checksums.yml"
  metadata_file: "/var/lib/froyo-registry/metadata.yml"
  enable_cache: true
  cache_ttl: 3600

auth:
  enabled: false
  type: "none"
  # For authentication:
  # enabled: true
  # type: "api_key"
  # api_keys:
  #   - "secret-key-1"
  #   - "secret-key-2"

logging:
  level: "info"
  format: "text"
  access_log: true
  file: "/var/log/froyo-registry/registry.log"
```

### Agent Configuration (`/etc/froyo/agent.yml`)

```yaml
execution:
  runner_path: /usr/local/bin/froyo-runner
  modules_dir: /etc/froyo/modules
  max_concurrent: 5
  default_timeout: 300

  # Module Registry (new!)
  module_registry_url: "http://192.168.1.60:8080"
  module_registry_api_key: ""  # Optional
```

## Authentication

### Enable API Key Authentication

**1. Update registry config:**

```yaml
# /etc/froyo-registry/registry.yml
auth:
  enabled: true
  type: "api_key"
  api_keys:
    - "your-secret-api-key-1"
    - "your-secret-api-key-2"
```

**2. Restart registry:**

```bash
systemctl restart froyo-registry
```

**3. Configure agents with API key:**

```yaml
# /etc/froyo/agent.yml
execution:
  module_registry_url: "http://192.168.1.60:8080"
  module_registry_api_key: "your-secret-api-key-1"
```

**4. Restart agents:**

```bash
systemctl restart froyo-agent
```

## Module Management

### Upload New Modules

**Option 1: Using upload script**

```bash
./upload-modules.sh <registry-host> <modules-dir>
```

**Option 2: Manual upload**

```bash
# Copy modules
scp bin/modules/*.wasm root@registry:/var/lib/froyo-registry/modules/

# Set permissions
ssh root@registry "chown -R froyo-registry:froyo-registry /var/lib/froyo-registry/modules"

# Restart to pick up new modules
ssh root@registry "systemctl restart froyo-registry"
```

### Update Modules

Simply upload new versions and restart the registry. Agents will:
1. Check cache
2. Request metadata (gets new checksum)
3. Checksum mismatch triggers re-download
4. New version cached

### Remove Modules

```bash
ssh root@registry "rm /var/lib/froyo-registry/modules/module-name.wasm"
ssh root@registry "systemctl restart froyo-registry"
```

## Monitoring

### Health Checks

```bash
# Health endpoint
curl http://registry:8080/healthz

# Readiness endpoint
curl http://registry:8080/ready

# View service status
ssh root@registry "systemctl status froyo-registry"
```

### Logs

```bash
# Live logs
ssh root@registry "journalctl -u froyo-registry -f"

# Recent logs
ssh root@registry "journalctl -u froyo-registry -n 100"

# Log file
ssh root@registry "tail -f /var/log/froyo-registry/registry.log"
```

### Metrics (Prometheus)

Add to Prometheus config:

```yaml
scrape_configs:
  - job_name: 'froyo-registry'
    static_configs:
      - targets: ['registry:8080']
    metrics_path: '/metrics'
```

**Metrics available:**
- `froyo_registry_module_downloads_total` - Total downloads
- `froyo_registry_bytes_served_total` - Total bytes served
- `froyo_registry_module_list_requests_total` - List requests
- `froyo_registry_module_metadata_requests_total` - Metadata requests

## High Availability

### Multi-Registry Setup

Deploy multiple registry instances:

```yaml
# agents configure with multiple URLs
execution:
  module_registry_url: "http://registry1:8080,http://registry2:8080"
```

Agents will try each registry in order until success.

### Load Balancer

```
┌─────────────────────┐
│   Load Balancer     │
│  registry.acme.com  │
└──────────┬──────────┘
           │
     ┌─────┴──────┐
     │            │
┌────▼────┐  ┌────▼────┐
│Registry │  │Registry │
│ Node 1  │  │ Node 2  │
└────┬────┘  └────┬────┘
     │            │
     └──────┬─────┘
            │
    ┌───────▼────────┐
    │ Shared Storage │
    │   (NFS/S3)     │
    └────────────────┘
```

## Troubleshooting

### Registry won't start

```bash
# Check logs
journalctl -u froyo-registry -n 50

# Common issues:
# - Port already in use: netstat -tlnp | grep 8080
# - Config syntax error: froyo-registry --config /etc/froyo-registry/registry.yml
# - Permissions: chown -R froyo-registry:froyo-registry /var/lib/froyo-registry
```

### Agents can't download modules

```bash
# Test connectivity
curl http://registry:8080/healthz

# Test module download
curl http://registry:8080/modules/generic-file/download -o /tmp/test.wasm

# Check agent logs
journalctl -u froyo-agent | grep registry

# Check firewall
ufw status
```

### Slow downloads

```bash
# Check registry resource usage
ssh root@registry "top"

# Check network bandwidth
ssh root@registry "iftop"

# Increase cache TTL
# /etc/froyo-registry/registry.yml
storage:
  cache_ttl: 7200  # 2 hours
```

## Security

### Best Practices

1. **Enable Authentication**
   - Use API keys in production
   - Rotate keys regularly
   - Different keys per environment

2. **Use TLS**
   - Deploy behind reverse proxy (nginx)
   - Use Let's Encrypt certificates
   - Enforce HTTPS only

3. **Network Security**
   - Restrict access to registry port
   - Use VPN or private network
   - Firewall rules

4. **Access Control**
   - Separate keys for dev/staging/prod
   - Audit API key usage
   - Monitor download patterns

## Performance

### Caching

**Registry-side:**
```yaml
storage:
  enable_cache: true
  cache_ttl: 3600  # 1 hour
```

**Agent-side:**
Modules are cached in `/etc/froyo/modules/` and reused until checksum changes.

### Scaling

| Agents | Registry Nodes | Load Balancer |
|--------|----------------|---------------|
| < 100  | 1              | No            |
| 100-500| 2              | Yes           |
| 500+   | 3+             | Yes + CDN     |

## Next Steps

1. **Deploy Registry**: `froyo apply systemd-stack.ofy`
2. **Upload Modules**: `./upload-modules.sh`
3. **Configure Agents**: Update `/etc/froyo/agent.yml`
4. **Test Download**: Check agent logs for successful downloads
5. **Enable Auth**: Add API keys for production

## References

- [Agent Configuration](../agent-onboarding/README.md)
- [Module Development](../../modules/README.md)
- [Pull Mode Quickstart](../../docs/QUICKSTART-PULL-MODE.md)
- [Module Distribution Comparison](../../docs/MODULE-DISTRIBUTION-COMPARISON.md)
