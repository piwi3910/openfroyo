# NATS Server Deployment

This directory contains examples for deploying NATS server infrastructure needed for OpenFroyo pull mode (agent-based) execution.

## Overview

NATS is a lightweight, high-performance messaging system that enables communication between the OpenFroyo orchestrator and agents. You need at least one NATS server to use pull mode.

## Deployment Options

### 1. Docker Deployment (Quick Start)

**Best for:** Testing, development, quick evaluation

```bash
# Update inventory.yml with your server IP
vim inventory.yml

# Deploy NATS in Docker
froyo apply docker-stack.ofy
```

**Pros:**
- Quick to deploy (< 1 minute)
- Easy to remove/restart
- No system dependencies

**Cons:**
- Requires Docker
- Less suitable for production

### 2. Systemd Deployment (Production)

**Best for:** Production environments, enterprise deployments

```bash
# Update inventory.yml with your server IP
vim inventory.yml

# Deploy NATS as systemd service
froyo apply systemd-stack.ofy
```

**Pros:**
- Production-ready
- Systemd integration
- Security hardening
- Automatic startup

**Cons:**
- More complex
- Requires root access

### 3. Local Testing (No Deployment)

**Best for:** Local development on your workstation

```bash
# Option A: Docker (easiest)
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:latest

# Option B: Download binary
curl -L https://github.com/nats-io/nats-server/releases/download/v2.10.22/nats-server-v2.10.22-linux-amd64.tar.gz | tar xz
cd nats-server-v2.10.22-linux-amd64
./nats-server

# Test connection
curl http://localhost:8222/varz
```

## Architecture

### Single Server (Simplest)

```
┌──────────────┐
│  froyo CLI   │
└──────┬───────┘
       │
   ┌───▼────────┐
   │    NATS    │
   │   Server   │
   └───┬────────┘
       │
  ┌────▼─────┐  ┌──────────┐  ┌──────────┐
  │  Agent   │  │  Agent   │  │  Agent   │
  │  (DC1)   │  │  (DC1)   │  │  (DC1)   │
  └──────────┘  └──────────┘  └──────────┘
```

**Use when:**
- Small deployments (< 100 agents)
- Single datacenter
- Testing/development

### Multi-Server Cluster (Recommended for Production)

```
┌──────────────┐
│  froyo CLI   │
└──────┬───────┘
       │
   ┌───▼───────────────────────────┐
   │     NATS Cluster (3 nodes)    │
   │  ┌──────┐ ┌──────┐ ┌──────┐  │
   │  │NATS-1│ │NATS-2│ │NATS-3│  │
   │  └──┬───┘ └──┬───┘ └──┬───┘  │
   └─────┼────────┼────────┼───────┘
         │        │        │
    ┌────▼───┐ ┌─▼────┐ ┌─▼────┐
    │Agents  │ │Agents│ │Agents│
    │ (DC1)  │ │(DC1) │ │(DC1) │
    └────────┘ └──────┘ └──────┘
```

**Use when:**
- Production environments
- High availability required
- 100+ agents

### Multi-Datacenter Super Cluster

```
      ┌──────────────┐
      │  froyo CLI   │
      └──────┬───────┘
             │
    ┌────────▼────────────────┐
    │  NATS Super Cluster     │
    │                         │
    │  DC1 Cluster ◄──────┐   │
    │  ┌──────┐  Gateway  │   │
    │  │NATS-1│◄──────────┼───┼──► DC2 Cluster
    │  └──┬───┘           │   │     ┌──────┐
    │     │               │   │     │NATS-4│
    │  Agents             │   │     └──┬───┘
    │  (DC1)              │   │        │
    └─────────────────────┘   │     Agents
                              │     (DC2)
                              └────────────
```

**Use when:**
- Multi-datacenter deployments
- Geographic distribution
- Latency optimization

## Configuration

### Basic Configuration

```conf
# /etc/nats/nats-server.conf
port: 4222
http_port: 8222

jetstream {
  store_dir: /var/lib/nats/jetstream
  max_memory_store: 1GB
  max_file_store: 10GB
}
```

### With Authentication

```conf
port: 4222
http_port: 8222

# Token-based auth (simplest)
authorization {
  token: "your-secret-token"
}

# Or user/password
authorization {
  users = [
    {user: froyo, password: "secure-password"}
  ]
}

# Or JWT/NKeys (most secure)
```

### Cluster Configuration

```conf
port: 4222
http_port: 8222

cluster {
  name: openfroyo
  listen: 0.0.0.0:6222
  routes: [
    nats://nats-server-1:6222
    nats://nats-server-2:6222
    nats://nats-server-3:6222
  ]
}
```

## Post-Deployment

### 1. Verify NATS is Running

```bash
# Check service
systemctl status nats-server

# Check HTTP endpoint
curl http://your-nats-server:8222/varz

# View logs
journalctl -u nats-server -f
```

### 2. Configure Agents

Update agent configuration to point to NATS:

```yaml
# /etc/froyo/agent.yml
nats:
  servers:
    - nats://your-nats-server:4222
```

Restart agents:
```bash
systemctl restart froyo-agent
```

### 3. Test with froyo CLI

```bash
# Execute stack using agent mode
froyo apply examples/agent-mode/hybrid-stack.ofy \
  --nats-server nats://your-nats-server:4222
```

## Monitoring

### HTTP Monitoring Endpoint

NATS provides a built-in monitoring endpoint:

```bash
# Server stats
curl http://localhost:8222/varz | jq

# Connection info
curl http://localhost:8222/connz | jq

# Subscription info
curl http://localhost:8222/subsz | jq

# JetStream info
curl http://localhost:8222/jsz | jq
```

### Key Metrics to Monitor

- **Connections**: Number of connected clients (agents + orchestrator)
- **Messages In/Out**: Message throughput
- **Memory Usage**: Should stay reasonable
- **Slow Consumers**: Indicates agent performance issues

## Firewall Configuration

Open these ports:

```bash
# NATS client port
ufw allow 4222/tcp

# NATS HTTP monitoring
ufw allow 8222/tcp

# NATS cluster port (for multi-server)
ufw allow 6222/tcp
```

## Troubleshooting

### NATS won't start

```bash
# Check logs
journalctl -u nats-server -n 50

# Common issues:
# - Port already in use: lsof -i :4222
# - Config syntax error: nats-server -c /etc/nats/nats-server.conf -D
# - Permissions: chown -R nats:nats /var/lib/nats
```

### Agents can't connect

```bash
# Test connectivity from agent host
telnet nats-server 4222

# Check firewall
ufw status

# Verify NATS is listening
netstat -tlnp | grep 4222
```

### High memory usage

```bash
# Check JetStream limits in config
# Reduce max_memory_store and max_file_store

# Restart NATS
systemctl restart nats-server
```

## Upgrading NATS

### Docker Deployment

```bash
# Update nats_version in docker-stack.ofy
# Re-run deployment
froyo apply docker-stack.ofy
```

### Systemd Deployment

```bash
# Update nats_version in systemd-stack.ofy
# Re-run deployment
froyo apply systemd-stack.ofy
```

## Security Best Practices

1. **Enable Authentication**
   - Use tokens or JWT for production
   - Rotate credentials regularly

2. **Use TLS**
   - Encrypt NATS traffic
   - Especially important over WAN

3. **Firewall Rules**
   - Only allow necessary ports
   - Restrict to known IP ranges

4. **Resource Limits**
   - Set max_connections
   - Set max_payload
   - Monitor resource usage

5. **Regular Updates**
   - Keep NATS up to date
   - Subscribe to security advisories

## Next Steps

After deploying NATS:

1. **Onboard Agents**
   ```bash
   froyo apply examples/agent-onboarding/onboard-agent.ofy
   ```

2. **Test Hybrid Mode**
   ```bash
   froyo apply examples/agent-mode/hybrid-stack.ofy \
     --nats-server nats://your-nats-server:4222
   ```

3. **Set Up Monitoring**
   - Configure Prometheus/Grafana
   - Set up alerts for NATS health

4. **Plan for HA**
   - Deploy 3-node cluster
   - Test failover scenarios

## References

- [NATS Official Documentation](https://docs.nats.io/)
- [NATS Server Configuration](https://docs.nats.io/running-a-nats-service/configuration)
- [NATS Clustering](https://docs.nats.io/running-a-nats-service/configuration/clustering)
- [JetStream](https://docs.nats.io/nats-concepts/jetstream)
- [NATS Security](https://docs.nats.io/running-a-nats-service/configuration/securing_nats)
