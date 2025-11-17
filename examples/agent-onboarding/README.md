# Agent Onboarding

This example demonstrates how to onboard servers to OpenFroyo pull mode by installing and configuring the froyo-agent.

## Overview

Agent onboarding converts servers from push mode (agentless/SSH) to pull mode (agent-based). This process:

1. ✅ Creates froyo user and group
2. ✅ Downloads froyo-agent and froyo-runner binaries
3. ✅ Configures NATS connection
4. ✅ Creates systemd service
5. ✅ Starts and enables the agent

## Prerequisites

- OpenFroyo CLI installed on your workstation
- NATS server deployed and accessible
- SSH access to target servers
- Root or sudo access on target servers

## Quick Start

### 1. Deploy NATS Server (If Not Already Done)

```bash
cd ../nats-deployment
# Update inventory.yml with your NATS server IP
froyo apply systemd-stack.ofy
```

### 2. Update Inventory

Edit `inventory.yml` and update with your server IPs:

```yaml
hosts:
  web-01:
    ssh_host: 192.168.1.10  # Your server IP
    ssh_user: root
  web-02:
    ssh_host: 192.168.1.11  # Your server IP
    ssh_user: root
```

### 3. Configure NATS Connection

Edit `onboard-agent.ofy` defaults section:

```yaml
defaults:
  nats_server: "nats://192.168.1.50:4222"  # Your NATS server
  datacenter: "dc1"
  agent_version: "1.0.0"
```

### 4. Run Onboarding

```bash
froyo apply onboard-agent.ofy
```

### 5. Verify Agents

SSH to each server and check:

```bash
# Check service status
systemctl status froyo-agent

# Check logs
journalctl -u froyo-agent -f

# Should see:
# "Connected to NATS: nats://192.168.1.50:4222"
# "Subscribed to: openfroyo.dc1.agents.web-01.tasks"
# "Health monitoring enabled"
```

## Architecture

### Push Mode (SSH-based)
```
┌─────────────┐           SSH            ┌────────────┐
│   froyo     │─────────────────────────>│  Server    │
│     CLI     │    Execute Commands      │  (Target)  │
└─────────────┘                          └────────────┘
```

### Pull Mode (Agent-based)
```
┌─────────────┐                          ┌────────────┐
│   froyo     │                          │  Server    │
│     CLI     │                          │            │
└──────┬──────┘                          │ ┌────────┐ │
       │                                 │ │ Agent  │ │
       │        ┌──────────────┐         │ └────┬───┘ │
       └───────>│     NATS     │<───────────────┘     │
                │ Message Queue│         └────────────┘
                └──────────────┘
                   Subscribe to:
              openfroyo.dc1.agents.>
```

## Configuration Details

### Agent Configuration (`/etc/froyo/agent.yml`)

The agent configuration includes:

- **Agent Identity**: Hostname-based ID and datacenter location
- **NATS Connection**: Server URLs, reconnection settings, subject prefixes
- **Execution Settings**: Working directories, concurrency limits, timeouts
- **Logging**: Log levels, file locations, rotation settings
- **Pull Mode**: Periodic stack execution configuration
- **Security**: TLS and authentication settings (optional)

### Systemd Service (`/etc/systemd/system/froyo-agent.service`)

The systemd unit includes:

- **Service Type**: Simple (foreground process)
- **User/Group**: Runs as `froyo` user for security
- **Auto-Restart**: Always restart on failure
- **Resource Limits**: File descriptors and process limits
- **Security Hardening**: PrivateTmp, ProtectSystem, ReadWrite paths

### Directory Structure

```
/etc/froyo/
  └── agent.yml                 # Agent configuration

/usr/local/bin/
  └── froyo-agent              # Agent binary (755)

/var/lib/froyo/
  ├── work/                    # Working directory for tasks
  └── modules/                 # Cached WASM modules

/var/log/froyo/
  └── agent.log                # Agent logs
```

## Agent Modes

### Pull Mode (Continuous Compliance)

The agent periodically executes configured stacks:

```yaml
pull:
  enabled: true
  interval: 60  # Poll every 60 seconds
  stacks:
    - name: compliance-check
      interval: 300  # Every 5 minutes
      enabled: true
```

### Push Mode (Ad-hoc Tasks)

The orchestrator can still push tasks to agents via NATS:

```bash
# froyo CLI publishes to NATS subject
froyo apply stack.ofy --target @agent:web-01
```

## NATS Subject Structure

Agents subscribe to datacenter-specific subjects:

```
openfroyo.<dc>.agents.>           # All agents in datacenter
openfroyo.<dc>.agents.<hostname>  # Specific agent
openfroyo.broadcast.>             # All agents (all DCs)
```

Example for `web-01` in `dc1`:
```
openfroyo.dc1.agents.web-01       # Direct messages to this agent
openfroyo.dc1.agents.>            # Messages to all dc1 agents
openfroyo.broadcast.>             # Global broadcasts
```

## Multi-Datacenter Setup

### NATS Super Cluster

```
┌──────────────────────────────────────────────────────────┐
│                     NATS Super Cluster                   │
│                                                          │
│  ┌──────────┐      Gateway      ┌──────────┐           │
│  │  DC1     │◄─────────────────►│  DC2     │           │
│  │  NATS    │                    │  NATS    │           │
│  └────┬─────┘                    └────┬─────┘           │
│       │                               │                 │
│       │ openfroyo.dc1.>               │ openfroyo.dc2.> │
│       │                               │                 │
└───────┼───────────────────────────────┼─────────────────┘
        │                               │
   ┌────▼─────┐                    ┌────▼─────┐
   │ Agents   │                    │ Agents   │
   │  (DC1)   │                    │  (DC2)   │
   └──────────┘                    └──────────┘
```

Agents only receive messages for their datacenter unless broadcasted globally.

## Security Considerations

### Network Security
- Ensure NATS server is properly secured (authentication, TLS)
- Use firewall rules to restrict NATS access
- Consider VPN or private networks for NATS traffic

### Agent Security
- Agent runs as non-root `froyo` user
- Systemd hardening (ProtectSystem, PrivateTmp)
- Use TLS for NATS connections in production
- Rotate credentials regularly

### Binary Integrity
- Always verify checksums when downloading agent binary
- Use HTTPS for binary downloads
- Consider signing binaries for additional security

## Troubleshooting

### Agent Won't Start

Check systemd status:
```bash
systemctl status froyo-agent
journalctl -u froyo-agent -n 50
```

Common issues:
- Binary not executable: `chmod +x /usr/local/bin/froyo-agent`
- Config file errors: Validate YAML syntax
- NATS unreachable: Check network connectivity

### Agent Not Connecting to NATS

Check logs:
```bash
journalctl -u froyo-agent | grep -i nats
```

Common issues:
- Wrong NATS URL in config
- Firewall blocking NATS port (4222)
- NATS server not running
- Authentication credentials incorrect

### Permissions Issues

Ensure correct ownership:
```bash
chown -R froyo:froyo /var/lib/froyo /var/log/froyo
chmod 755 /usr/local/bin/froyo-agent
```

## Uninstalling Agent

To remove the agent and revert to SSH mode:

```bash
# Stop and disable service
systemctl stop froyo-agent
systemctl disable froyo-agent

# Remove service file
rm /etc/systemd/system/froyo-agent.service
systemctl daemon-reload

# Remove agent binary and config
rm /usr/local/bin/froyo-agent
rm -rf /etc/froyo

# Optionally remove user and directories
userdel -r froyo
rm -rf /var/lib/froyo /var/log/froyo
```

## Next Steps

After onboarding agents:

1. **Create Pull Mode Stacks** - Define periodic compliance checks
2. **Set Up Monitoring** - Monitor agent health and NATS connectivity
3. **Configure Alerting** - Alert on agent failures or drift detection
4. **Implement Drift Detection** - Compare desired vs actual state
5. **Scale Out** - Onboard more servers to agent mode

## Related Documentation

- [get_url module](../../modules/generic/get_url/) - Download agent binary
- [service module](../../modules/generic/service/) - Manage agent service
- [NATS Architecture](../../docs/architecture/nats.md) - NATS setup and configuration
- [Agent Architecture](../../docs/architecture/agent.md) - Agent design and operation

## Example: Hybrid Deployment

You can run both push mode and pull mode simultaneously:

```yaml
inventory:
  hosts:
    # Push mode servers (SSH-based)
    - name: admin-server
      ssh_host: 10.0.1.10
      ssh_user: root
      mode: ssh

    # Pull mode servers (agent-based)
    - name: web-01
      mode: agent
      datacenter: dc1

    - name: web-02
      mode: agent
      datacenter: dc1
```
