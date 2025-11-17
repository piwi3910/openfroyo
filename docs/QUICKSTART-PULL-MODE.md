# QuickStart: OpenFroyo Pull Mode (Agent-Based)

This guide walks you through setting up OpenFroyo's pull mode (agent-based) execution from scratch.

## What You'll Build

```
┌──────────────┐
│  Your Laptop │  Push mode: Direct SSH to admin-server
│  (froyo CLI) │  Pull mode: Via NATS to web-servers
└──────┬───────┘
       │
   ┌───┴────────────────────┐
   │                        │
SSH│                    ┌───▼────────┐
   │                    │    NATS    │
   │                    │   Server   │
   │                    └───┬────────┘
   │                        │
   │                   ┌────▼─────┐  ┌──────────┐
   │                   │  web-01  │  │  web-02  │
   │                   │ (agent)  │  │ (agent)  │
   │                   └──────────┘  └──────────┘
   │
┌──▼──────────────┐
│  admin-server   │
│   (SSH only)    │
└─────────────────┘
```

## Time Required

- **Quick test (Docker)**: 15 minutes
- **Production setup**: 45 minutes

## Prerequisites

- 3+ Linux servers (can be VMs or cloud instances)
- SSH access to all servers
- OpenFroyo CLI installed on your laptop

## Step 1: Deploy NATS Server (5 min)

Choose one server to be your NATS server.

### Option A: Docker (Fastest)

```bash
# On NATS server
docker run -d \
  --name nats-local \
  --restart unless-stopped \
  -p 4222:4222 \
  -p 8222:8222 \
  nats:latest -js --http_port 8222

# Verify
curl http://localhost:8222/varz
```

### Option B: Using OpenFroyo (Automated)

```bash
# On your laptop
cd examples/nats-deployment

# Update inventory.yml with your NATS server IP
vim inventory.yml

# Deploy (choose one)
froyo apply docker-stack.ofy    # Docker deployment
froyo apply systemd-stack.ofy   # Production deployment
```

**Result:** NATS server running on `nats://your-nats-server:4222`

## Step 2: Onboard Agents (10 min)

Convert your web servers from push mode to pull mode.

```bash
cd examples/agent-onboarding

# Update inventory with your web servers
vim inventory.yml
```

**inventory.yml:**
```yaml
hosts:
  web-01:
    ssh_host: 192.168.1.10
    ssh_user: root
  web-02:
    ssh_host: 192.168.1.11
    ssh_user: root
```

**onboard-agent.ofy defaults (update these):**
```yaml
defaults:
  nats_server: "nats://192.168.1.50:4222"  # Your NATS server
  datacenter: "dc1"
  agent_version: "1.0.0"
```

**Run onboarding:**
```bash
froyo apply onboard-agent.ofy
```

This will:
1. ✓ Create froyo user
2. ✓ Download froyo-agent binary
3. ✓ Configure NATS connection
4. ✓ Install systemd service
5. ✓ Start agent

**Verify agents are running:**
```bash
# SSH to web-01
ssh root@192.168.1.10
systemctl status froyo-agent
journalctl -u froyo-agent | grep "connected to NATS"
```

## Step 3: Create Hybrid Inventory (5 min)

Create an inventory mixing push and pull mode hosts.

**inventory/hybrid.yml:**
```yaml
---
hosts:
  # Push mode - direct SSH
  admin-server:
    mode: ssh
    ssh_host: 192.168.1.5
    ssh_user: root

  # Pull mode - via agents
  web-01:
    mode: agent
    agent_id: web-01
    datacenter: dc1

  web-02:
    mode: agent
    agent_id: web-02
    datacenter: dc1

groups:
  webservers:
    hosts:
      - web-01
      - web-02
```

## Step 4: Create Your First Hybrid Stack (5 min)

**stacks/test-hybrid.ofy:**
```yaml
name: Test Hybrid Mode
description: Test both push and pull execution

inventory:
  - ../inventory/hybrid.yml

run:
  # Push mode: Direct SSH to admin server
  - name: Create test file on admin server
    module: generic/file
    hosts:
      - admin-server
    vars:
      path: /tmp/test-push.txt
      state: touch

  # Pull mode: Via NATS to agents
  - name: Create test file on web servers
    module: generic/file
    hosts:
      - @group:webservers
    vars:
      path: /tmp/test-pull.txt
      state: touch

  # Mixed: Both modes in one task
  - name: Check system on all hosts
    module: generic/command
    hosts:
      - admin-server
      - web-01
      - web-02
    vars:
      cmd: uname -a
```

## Step 5: Run Your First Hybrid Stack! (2 min)

```bash
froyo apply stacks/test-hybrid.ofy \
  --nats-server nats://192.168.1.50:4222
```

**Expected output:**
```
Loading stack: stacks/test-hybrid.ofy
Loading inventory...
Loaded 3 hosts, 1 groups
Connecting to NATS: nats://192.168.1.50:4222

Executing stack: Test Hybrid Mode

[1/3] Create test file on admin server
  → admin-server (192.168.1.5) [SSH]
    ✓ changed: File created

[2/3] Create test file on web servers
  → web-01 (agent: web-01) [AGENT]
    ✓ changed: File created
  → web-02 (agent: web-02) [AGENT]
    ✓ changed: File created

[3/3] Check system on all hosts
  → admin-server (192.168.1.5) [SSH]
    ✓ ok: Command executed
  → web-01 (agent: web-01) [AGENT]
    ✓ ok: Command executed
  → web-02 (agent: web-02) [AGENT]
    ✓ ok: Command executed

Stack execution completed successfully!
```

## Congratulations! 🎉

You now have a working hybrid mode setup!

## What Just Happened?

1. **Push Mode (admin-server)**:
   - froyo CLI connected directly via SSH
   - Executed module immediately
   - Same as traditional mode

2. **Pull Mode (web-01, web-02)**:
   - froyo CLI published task to NATS
   - Agents pulled task from queue
   - Agents executed locally
   - Results published back to NATS
   - froyo CLI received results

## Next Steps

### 1. Try Continuous Compliance

Enable pull mode scheduler for continuous compliance:

**/etc/froyo/agent.yml on web-01:**
```yaml
pull:
  enabled: true
  stacks:
    - name: security-baseline
      interval: 300  # Every 5 minutes
      enabled: true
```

Agents will now execute this stack every 5 minutes automatically!

### 2. Monitor Agent Health

NATS provides agent health visibility:

```bash
# Subscribe to health updates
nats sub "openfroyo.dc1.agents.*.health"
```

### 3. Scale Out

Add more agents:
```bash
# Onboard more servers
froyo apply examples/agent-onboarding/onboard-agent.ofy
```

### 4. Multi-Datacenter

Deploy NATS in each datacenter and configure gateways:

```yaml
# DC1 agents
web-dc1-01:
  mode: agent
  datacenter: dc1

# DC2 agents
web-dc2-01:
  mode: agent
  datacenter: dc2
```

Messages route automatically to correct datacenter!

## Common Patterns

### Pattern 1: Bootstrap with Push, Manage with Pull

```yaml
run:
  # Initial setup (push mode - needs SSH)
  - name: Initial package install
    module: linux/package
    hosts:
      - new-server
    vars:
      name: python3

  # Then onboard to pull mode
  - name: Install agent
    module: generic/get_url
    hosts:
      - new-server
    # ... onboarding steps ...
```

### Pattern 2: Critical Tasks via Push, Routine via Pull

```yaml
# Critical changes: Use push mode for immediate feedback
admin-server:
  mode: ssh

# Routine compliance: Use pull mode for scalability
web-servers:
  mode: agent
```

### Pattern 3: Hybrid Environments

Some hosts can't run agents (embedded systems, network devices):

```yaml
hosts:
  # Can run agents
  linux-servers:
    mode: agent

  # Can't run agents
  network-switches:
    mode: ssh
```

## Troubleshooting

### Agents won't connect to NATS

```bash
# Check agent logs
journalctl -u froyo-agent -f

# Test NATS connectivity from agent host
telnet nats-server-ip 4222

# Verify NATS is running
curl http://nats-server-ip:8222/varz
```

### Tasks timeout on agents

```bash
# Check agent is subscribed
# On NATS server
curl http://localhost:8222/subsz | jq

# Should see subscriptions to:
# openfroyo.dc1.agents.web-01.tasks
```

### Results not received

```bash
# Check orchestrator subscribed to results
# In froyo CLI logs, look for:
# "Subscribed to results: openfroyo.dc1.agents.*.results"

# Verify agent can publish
# Check agent logs for publish errors
```

## Architecture Deep Dive

### How Push Mode Works

```
1. froyo parses stack
2. Connects to host via SSH
3. Uploads froyo-runner if needed
4. Uploads WASM module
5. Executes: froyo-runner --module x.wasm
6. Receives result
7. Displays output
```

### How Pull Mode Works

```
1. froyo parses stack
2. Connects to NATS
3. Generates task ID
4. Publishes task message to subject
5. Subscribes to result subject
6. Waits for agent to respond
7. Receives result from NATS
8. Displays output

Meanwhile on agent:
1. Subscribed to openfroyo.dc1.agents.web-01.tasks
2. Receives task message
3. Executes with froyo-runner locally
4. Publishes result to openfroyo.dc1.agents.web-01.results
```

### Message Flow

```
Orchestrator                 NATS                    Agent
     │                         │                       │
     ├─ Publish Task ─────────►│                       │
     │  (task_id: abc123)       ├─ Deliver Task ──────►│
     │                          │                       ├─ Execute
     │                          │                       │  Module
     │                          │◄── Publish Result ───┤
     │◄─ Receive Result ────────┤  (task_id: abc123)   │
     ├─ Display                 │                       │
```

## Performance Comparison

### Push Mode (SSH)

**Pros:**
- Immediate execution
- Direct connection
- Simple troubleshooting

**Cons:**
- Limited by orchestrator resources
- SSH connections for every task
- Doesn't scale beyond ~50 concurrent

**Good for:** < 50 hosts, immediate tasks

### Pull Mode (Agent)

**Pros:**
- Highly scalable (1000+ agents)
- Asynchronous execution
- Reduced orchestrator load
- Continuous compliance

**Cons:**
- Requires NATS infrastructure
- Agent deployment needed
- Slightly more complex

**Good for:** 100+ hosts, continuous compliance

## Cost Analysis

### Push Mode
- No infrastructure cost
- SSH bandwidth cost
- Orchestrator compute cost

### Pull Mode
- NATS server cost (minimal)
- Agent compute cost (minimal)
- NATS bandwidth cost (minimal)
- Reduced orchestrator cost

**Break-even:** ~50 hosts

## Security Considerations

### Push Mode
- Requires inbound SSH (port 22)
- Firewall rules needed
- SSH key management

### Pull Mode
- Only outbound connections (agents → NATS)
- Firewall-friendly
- NATS authentication
- No inbound ports on agents

## References

- [Agent Architecture](architecture/agent-architecture.md)
- [NATS Deployment](../examples/nats-deployment/)
- [Agent Onboarding](../examples/agent-onboarding/)
- [Hybrid Mode Example](../examples/agent-mode/)
- [NATS Documentation](https://docs.nats.io/)

## Support

Questions? Issues?
- GitHub: https://github.com/piwi3910/openfroyo/issues
- Documentation: https://github.com/piwi3910/openfroyo/tree/main/docs
