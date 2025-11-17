# Agent Mode Example

This example demonstrates OpenFroyo's hybrid execution mode, where you can mix SSH-based (push mode) and agent-based (pull mode) hosts in the same stack.

## Overview

OpenFroyo supports two execution modes:

1. **SSH Mode (Push)**: Traditional approach where the orchestrator connects to hosts via SSH
2. **Agent Mode (Pull)**: Modern approach where agents connect to NATS message queue and pull tasks

Both modes can coexist in the same deployment!

## Example Files

- `inventory.yml` - Inventory with mixed SSH and agent hosts
- `hybrid-stack.ofy` - Stack that executes on both types of hosts

## Prerequisites

### For SSH Mode Hosts
- SSH access to target hosts
- SSH key or password authentication

### For Agent Mode Hosts
- NATS server running and accessible
- froyo-agent installed and running on target hosts
- Agent configuration with correct datacenter and agent_id

## Inventory Configuration

### SSH Mode Host
```yaml
legacy-server:
  mode: ssh
  ssh_host: 192.168.1.100
  ssh_user: root
  ssh_key_file: ~/.ssh/id_rsa
```

### Agent Mode Host
```yaml
web-01:
  mode: agent
  agent_id: web-01          # Must match agent's ID
  datacenter: dc1           # Must match agent's datacenter
```

## Running the Example

### 1. Set Up NATS Server

```bash
# Option 1: Using Docker
docker run -d --name nats -p 4222:4222 nats:latest

# Option 2: Download and run NATS server
# https://docs.nats.io/running-a-nats-service/introduction/installation
```

### 2. Set Up Agents

Agents must be running on the target hosts before you can execute tasks. See the [agent onboarding example](../agent-onboarding/) for details on setting up agents.

Example agent configuration (`/etc/froyo/agent.yml`):
```yaml
agent:
  id: web-01
  datacenter: dc1

nats:
  servers:
    - nats://your-nats-server:4222
```

Start the agent:
```bash
sudo systemctl start froyo-agent
```

### 3. Update Inventory

Edit `inventory.yml` to match your environment:
- For SSH hosts: Update IP addresses, usernames, credentials
- For agent hosts: Update agent_id and datacenter to match running agents

### 4. Execute the Stack

#### With only SSH hosts:
```bash
froyo apply hybrid-stack.ofy
```

#### With agent mode hosts:
```bash
froyo apply hybrid-stack.ofy --nats-server nats://localhost:4222
```

#### With NATS authentication:
```bash
froyo apply hybrid-stack.ofy \
  --nats-server nats://nats.example.com:4222 \
  --nats-token your-auth-token
```

## Execution Flow

When you run the stack:

1. **SSH Hosts**: Orchestrator connects via SSH and executes modules directly
2. **Agent Hosts**: Orchestrator publishes task messages to NATS, agents pull and execute

The output will show the execution mode for each host:

```
[1/4] Create file on SSH host
  → legacy-server (192.168.1.100) [SSH]
    ✓ changed: File created

[2/4] Create file on agent hosts
  → web-01 (agent: web-01) [AGENT]
    ✓ changed: File created
  → web-02 (agent: web-02) [AGENT]
    ✓ changed: File created
```

## Benefits of Hybrid Mode

### When to Use SSH Mode
- Legacy infrastructure without agent support
- One-off tasks on rarely managed hosts
- Hosts that cannot run agents (embedded systems, etc.)
- Initial bootstrapping before installing agents

### When to Use Agent Mode
- Large-scale deployments (100+ hosts)
- Continuous compliance and drift detection
- Reduced load on orchestrator
- Multi-datacenter deployments
- Hosts behind firewalls (agents can call out)

## Multi-Datacenter Setup

With agent mode, you can efficiently manage hosts across multiple datacenters:

```yaml
# DC1 hosts connect to NATS in DC1
web-01:
  mode: agent
  agent_id: web-01
  datacenter: dc1

# DC2 hosts connect to NATS in DC2
web-02:
  mode: agent
  agent_id: web-02
  datacenter: dc2
```

NATS super cluster handles routing messages to the correct datacenter automatically.

## Architecture Diagram

```
┌──────────────────┐
│  froyo CLI       │
│  (Orchestrator)  │
└────┬────────┬────┘
     │        │
SSH  │        │ NATS
     │        │
     │    ┌───▼──────────┐
     │    │ NATS Server  │
     │    └───┬──────────┘
     │        │
     │   ┌────▼────┐  ┌─────────┐
     │   │ web-01  │  │ web-02  │
     │   │ (agent) │  │ (agent) │
     │   └─────────┘  └─────────┘
     │
  ┌──▼─────────────┐  ┌────────────────┐
  │ legacy-server  │  │   db-server    │
  │   (SSH mode)   │  │   (SSH mode)   │
  └────────────────┘  └────────────────┘
```

## Troubleshooting

### "NATS server required for agent mode hosts"
- Ensure you're using the `--nats-server` flag
- Verify NATS server is running and accessible

### "agent_id is required for agent mode"
- Check that agent mode hosts have `agent_id` set in inventory

### "failed to execute on agent: task timed out"
- Verify agent is running: `systemctl status froyo-agent`
- Check agent logs: `journalctl -u froyo-agent -f`
- Ensure agent can connect to NATS server
- Check firewall rules for NATS port (4222)

### Tasks execute on SSH but not agents
- Verify agents are subscribed to correct NATS subjects
- Check datacenter and agent_id match between inventory and agent config
- Review NATS logs for connection issues

## See Also

- [Agent Architecture Documentation](../../docs/architecture/agent-architecture.md)
- [Agent Onboarding Example](../agent-onboarding/)
- [NATS Documentation](https://docs.nats.io/)
