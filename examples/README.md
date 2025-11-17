# OpenFroyo Examples

This directory contains practical examples demonstrating OpenFroyo's capabilities for infrastructure automation and orchestration.

## 📚 Table of Contents

1. [Getting Started](#getting-started)
2. [Push Mode Examples](#push-mode-examples)
3. [Pull Mode Setup](#pull-mode-setup)
4. [Hybrid Mode Examples](#hybrid-mode-examples)
5. [Module Examples](#module-examples)
6. [Learning Path](#learning-path)

## Getting Started

OpenFroyo supports two execution modes:

- **Push Mode (SSH)**: Direct SSH execution, agentless, great for < 50 hosts
- **Pull Mode (Agent)**: Agent-based execution via NATS, scalable to 1000+ hosts

You can use either mode independently or mix them in hybrid deployments.

## Push Mode Examples

Push mode uses direct SSH connections to execute tasks on remote hosts. No agents required!

### Basic Stack Execution

```bash
# Simple package installation
froyo apply examples/stacks/package-test.ofy

# Test script execution
froyo apply examples/stacks/script-test.ofy

# Full system test
froyo apply examples/stacks/test-openfroyo.ofy
```

**Example inventory (push mode):**
```yaml
hosts:
  web-01:
    mode: ssh                          # Push mode
    ssh_host: 192.168.1.10
    ssh_user: root
    ssh_key_file: ~/.ssh/id_rsa
```

**When to use push mode:**
- Small deployments (< 50 hosts)
- One-time tasks or ad-hoc changes
- Servers that can't run agents (embedded systems, network devices)
- Quick testing and experimentation

## Pull Mode Setup

Pull mode requires NATS infrastructure and agents on target hosts. Follow these steps:

### Step 1: Deploy NATS Server (15 min)

NATS is the message queue that enables agent communication.

```bash
cd nats-deployment

# Option A: Quick Docker deployment
./local-test.sh

# Option B: Production systemd deployment
# 1. Update inventory.yml with your NATS server IP
vim inventory.yml

# 2. Deploy NATS
froyo apply systemd-stack.ofy
```

**What you get:**
- NATS server running on port 4222
- HTTP monitoring on port 8222
- JetStream enabled for reliable messaging
- Systemd service with auto-restart

**Verify NATS:**
```bash
curl http://your-nats-server:8222/varz
```

See: [nats-deployment/README.md](nats-deployment/README.md)

### Step 2: Onboard Agents (10 min)

Install froyo-agent on your target servers.

```bash
cd agent-onboarding

# 1. Update inventory with your servers
vim inventory.yml

# 2. Configure NATS connection
vim onboard-agent.ofy  # Set nats_server URL

# 3. Run onboarding
froyo apply onboard-agent.ofy
```

**What you get:**
- froyo-agent binary installed
- froyo-runner for WASM execution
- Systemd service running as froyo user
- Agent connected to NATS and ready

**Verify agents:**
```bash
# SSH to a server
ssh root@192.168.1.10

# Check service
systemctl status froyo-agent

# Check logs
journalctl -u froyo-agent -f
# Should see: "Connected to NATS: nats://..."
```

See: [agent-onboarding/README.md](agent-onboarding/README.md)

### Step 3: Test Pull Mode (5 min)

**Update your inventory for agent mode:**
```yaml
hosts:
  web-01:
    mode: agent                        # Pull mode
    agent_id: web-01
    datacenter: dc1
```

**Run a test stack:**
```bash
cd agent-mode
froyo apply hybrid-stack.ofy --nats-server nats://192.168.1.50:4222
```

See: [agent-mode/README.md](agent-mode/README.md)

## Hybrid Mode Examples

Mix push and pull modes in the same deployment!

### Example: Hybrid Inventory

```yaml
---
hosts:
  # Admin server - push mode (direct SSH)
  admin-server:
    mode: ssh
    ssh_host: 192.168.1.5
    ssh_user: root

  # Web servers - pull mode (via agents)
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

### Example: Hybrid Stack

```yaml
name: Hybrid Deployment
inventory:
  - inventory.yml

run:
  # Push mode: Direct SSH to admin server
  - name: Configure admin server
    module: linux/package
    hosts:
      - admin-server
    vars:
      name: monitoring-tools
      state: present

  # Pull mode: Via NATS to web servers
  - name: Deploy application
    module: generic/copy
    hosts:
      - @group:webservers
    vars:
      src: app.tar.gz
      dest: /opt/app/

  # Mixed: Both modes in one task
  - name: Check all servers
    module: generic/command
    hosts:
      - admin-server
      - web-01
      - web-02
    vars:
      cmd: df -h
```

**Run hybrid stack:**
```bash
froyo apply hybrid-stack.ofy --nats-server nats://192.168.1.50:4222
```

**When to use hybrid mode:**
- Large deployments with mix of server types
- Phased migration from push to pull
- Different security zones (DMZ uses push, internal uses pull)
- Some servers can't run agents

See: [agent-mode/hybrid-stack.ofy](agent-mode/hybrid-stack.ofy)

## Module Examples

### Package Management

```bash
# Install packages
froyo apply examples/stacks/package-test.ofy

# Test package lifecycle
froyo apply examples/stacks/package-cycle-test.ofy
```

### Script Execution

```bash
# Run custom scripts
froyo apply examples/stacks/script-test.ofy
```

### File Operations

See module documentation in `modules/` directory for all available modules.

## Learning Path

### Path 1: Push Mode Only (Fastest)

Perfect for getting started quickly with small deployments.

1. ✅ **Install OpenFroyo CLI**
2. ✅ **Create inventory** with SSH hosts
3. ✅ **Run example stack**: `froyo apply examples/stacks/test-openfroyo.ofy`
4. ✅ **Create your own stack**

**Time:** 30 minutes
**Good for:** < 50 hosts, testing, learning

### Path 2: Pull Mode Setup (Production)

For larger deployments and continuous compliance.

1. ✅ **Install OpenFroyo CLI**
2. ✅ **Deploy NATS server**: [nats-deployment/](nats-deployment/)
3. ✅ **Onboard agents**: [agent-onboarding/](agent-onboarding/)
4. ✅ **Test pull mode**: [agent-mode/](agent-mode/)
5. ✅ **Enable continuous compliance**: Configure pull scheduler in agent.yml

**Time:** 60 minutes
**Good for:** 100+ hosts, continuous compliance, multi-datacenter

### Path 3: Hybrid Mode (Gradual Migration)

Start with push, gradually migrate to pull.

1. ✅ **Start with push mode** (Path 1)
2. ✅ **Deploy NATS server** when ready to scale
3. ✅ **Onboard first batch** of agents (10-20 servers)
4. ✅ **Run hybrid stacks** with both modes
5. ✅ **Migrate remaining servers** in phases
6. ✅ **Keep some servers on push mode** (if needed)

**Time:** Multiple phases
**Good for:** Large existing deployments, risk-averse migrations

## Directory Structure

```
examples/
├── README.md                    # This file
│
├── stacks/                      # Basic stack examples
│   ├── package-test.ofy         # Package installation
│   ├── script-test.ofy          # Script execution
│   └── test-openfroyo.ofy       # Comprehensive test
│
├── nats-deployment/             # NATS server setup
│   ├── docker-stack.ofy         # Docker deployment
│   ├── systemd-stack.ofy        # Production deployment
│   ├── local-test.sh            # Quick local test
│   └── README.md                # NATS deployment guide
│
├── agent-onboarding/            # Agent installation
│   ├── onboard-agent.ofy        # Onboarding stack
│   ├── inventory.yml            # Target servers
│   └── README.md                # Onboarding guide
│
└── agent-mode/                  # Hybrid mode examples
    ├── hybrid-stack.ofy         # Mixed push/pull execution
    ├── inventory.yml            # Hybrid inventory
    └── README.md                # Hybrid mode guide
```

## Common Patterns

### Pattern 1: Bootstrap with Push, Manage with Pull

Use push mode for initial setup, then switch to pull mode for ongoing management.

```yaml
run:
  # Initial setup (push mode)
  - name: Bootstrap server
    module: linux/package
    hosts:
      - new-server
    vars:
      name: python3

  # Deploy agent (push mode)
  - name: Install froyo-agent
    module: generic/get_url
    hosts:
      - new-server
    vars:
      url: "https://github.com/.../froyo-agent-linux-amd64"
      dest: /usr/local/bin/froyo-agent

  # Now server is in pull mode for future operations!
```

### Pattern 2: Critical via Push, Routine via Pull

Use push for immediate critical changes, pull for scheduled compliance.

```yaml
# inventory.yml
hosts:
  # Critical systems - use push for immediate response
  critical-db:
    mode: ssh

  # Web tier - use pull for scalability
  web-01:
    mode: agent
```

### Pattern 3: Multi-Datacenter

Use pull mode with datacenter isolation.

```yaml
hosts:
  # DC1 agents
  web-dc1-01:
    mode: agent
    datacenter: dc1

  # DC2 agents
  web-dc2-01:
    mode: agent
    datacenter: dc2
```

NATS routes messages to correct datacenter automatically!

## Troubleshooting

### Push Mode Issues

```bash
# Test SSH connection
ssh root@192.168.1.10

# Check SSH keys
ssh -i ~/.ssh/id_rsa root@192.168.1.10

# Enable SSH debug
froyo apply stack.ofy --verbose
```

### Pull Mode Issues

**Agents won't connect to NATS:**
```bash
# Check agent logs
journalctl -u froyo-agent -f

# Test NATS connectivity
telnet nats-server-ip 4222

# Verify NATS is running
curl http://nats-server-ip:8222/varz
```

**Tasks timeout:**
```bash
# Check NATS subscriptions
curl http://nats-server:8222/subsz | jq

# Should see:
# openfroyo.dc1.agents.web-01.tasks
```

**Agents not receiving tasks:**
```bash
# Check orchestrator NATS connection
# Look for: "Connected to NATS" in froyo CLI output

# Check agent is subscribed
journalctl -u froyo-agent | grep "Subscribed to"
```

## Next Steps

1. **Read the QuickStart**: [docs/QUICKSTART-PULL-MODE.md](../docs/QUICKSTART-PULL-MODE.md)
2. **Explore Modules**: See `modules/` directory for all available modules
3. **Join Community**: https://github.com/piwi3910/openfroyo/discussions
4. **Report Issues**: https://github.com/piwi3910/openfroyo/issues

## Performance Guidelines

### Push Mode (SSH)

- **Good for:** < 50 hosts
- **Max concurrent:** ~50 SSH connections
- **Latency:** Low (direct connection)
- **Resource usage:** High on orchestrator

### Pull Mode (Agent)

- **Good for:** 100+ hosts
- **Max concurrent:** 1000+ agents
- **Latency:** Slightly higher (via NATS)
- **Resource usage:** Distributed across agents

### Hybrid Mode

Best of both worlds! Use push for small/critical sets, pull for large/routine sets.

## Security Best Practices

### Push Mode

- Use SSH key authentication (not passwords)
- Rotate SSH keys regularly
- Use bastion/jump hosts for DMZ access
- Implement firewall rules

### Pull Mode

- Enable NATS authentication (token or JWT)
- Use TLS for NATS connections
- Run agents as unprivileged user (froyo)
- Implement network segmentation

### Both Modes

- Audit all stack executions
- Use version control for stacks
- Implement approval workflows for production
- Monitor all changes

## References

- [Main Documentation](../docs/)
- [Module Reference](../modules/)
- [Architecture](../docs/architecture/)
- [GitHub Repository](https://github.com/piwi3910/openfroyo)
