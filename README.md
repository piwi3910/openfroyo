# OpenFroyo

An agentless, Go-based automation and orchestration framework that combines concepts from Ansible and Terraform. It uses WebAssembly (WASM) as the execution engine for remote actions over SSH.

## Features

- **Dual Operation Modes** - Push mode (agentless SSH) or Pull mode (persistent agents with NATS)
- **WASM-Based Execution** - Remote actions execute as WASM modules via lightweight `froyo-runner` binary
- **Stack-Based Orchestration** - Ordered execution model with modules and task blocks
- **Module Registry** - Centralized HTTP server for distributing WASM modules to agents
- **Go Implementation** - Single language for all components
- **Production-Ready** - Includes systemd services, Docker deployment, monitoring, and CI/CD

## Operational Modes

OpenFroyo supports two operational modes to fit different infrastructure needs:

### Push Mode (Agentless)

Traditional Ansible-style operation via SSH. The CLI connects directly to hosts, transfers binaries, and executes tasks.

**Use Cases:**
- Quick automation tasks
- Small environments (< 50 hosts)
- Ad-hoc operations
- Development and testing

**Pros:**
- No infrastructure setup required
- Simple deployment
- Direct control

**Cons:**
- CLI must have network access to all hosts
- Not suitable for NAT/firewall scenarios
- Limited scalability

**Quick Start:** See [Push Mode Quick Start](#quick-start-push-mode) below.

### Pull Mode (Agent-Based)

Production-grade operation with persistent agents that connect to a NATS message broker.

**Architecture:**
```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  froyo CLI  │────▶│ NATS Broker  │◀────│  froyo-agent    │
│  (Control)  │     │  (Messages)  │     │  (Target Host)  │
└─────────────┘     └──────────────┘     └─────────────────┘
                            │                      │
                            │                      ▼
                    ┌───────────────┐     ┌─────────────────┐
                    │ Module        │────▶│  froyo-runner   │
                    │ Registry      │     │  (WASM Exec)    │
                    │ (HTTP)        │     └─────────────────┘
                    └───────────────┘
```

**Components:**
- **froyo-agent** - Persistent daemon on each host (systemd service)
- **NATS** - Message broker for command distribution
- **Module Registry** - HTTP server for WASM module distribution
- **froyo-runner** - WASM execution engine (invoked by agent)

**Use Cases:**
- Production environments
- Large-scale deployments (100s-1000s of hosts)
- Hosts behind NAT/firewalls
- Edge computing scenarios
- Continuous configuration management

**Pros:**
- Agents connect outbound (works through NAT/firewalls)
- Highly scalable (NATS can handle 10k+ agents)
- Automatic module distribution
- Self-healing (agents auto-reconnect)
- Real-time task execution

**Cons:**
- Requires NATS infrastructure
- More complex setup
- Persistent daemon on hosts

**Quick Start:** See [Pull Mode Quick Start](#quick-start-pull-mode) below.

## Quick Start (Push Mode)

### Prerequisites

- Go 1.21 or later
- [TinyGo](https://tinygo.org/getting-started/install/) (for building WASM modules)
- SSH access to a Linux host

### Build

```bash
# Build CLI and runner
make build

# Build WASM modules
make build-wasm
```

### Configure

Edit `examples/inventory/hosts.yml` with your SSH host details:

```yaml
hosts:
  test-host:
    ssh_host: 192.168.1.100  # Your Linux host IP
    ssh_port: 22
    ssh_user: root
    ssh_key_file: ~/.ssh/id_rsa
```

### Run

```bash
./bin/froyo apply examples/stacks/test.ofy
```

This will:
1. Connect to the host via SSH
2. Transfer the `froyo-runner` binary
3. Upload the WASM module
4. Execute `uname -r` on the remote host
5. Return and display the kernel version

### Binaries Location

All built binaries are placed in the `bin/` directory:
- `bin/froyo` - CLI tool (6.8MB)
- `bin/froyo-runner-{os}-{arch}` - Platform-specific WASM runners (~5MB each)

**Supported Platforms:**
- `linux-amd64` - Linux x86_64 (most servers)
- `linux-arm64` - Linux ARM64 (Raspberry Pi, ARM servers)
- `darwin-amd64` - macOS Intel
- `darwin-arm64` - macOS Apple Silicon
- `windows-amd64` - Windows x64

**Auto-Detection:** OpenFroyo automatically detects the remote host's OS and architecture, then uploads the correct binary. No manual selection needed!

The `bin/` directory is git-ignored to keep the repository clean.

## Quick Start (Pull Mode)

For production deployments with persistent agents and module registry.

### Prerequisites

- Go 1.21 or later
- [TinyGo](https://tinygo.org/getting-started/install/) (for building WASM modules)
- NATS server (can be deployed with included stack)
- Linux hosts for agents

### 1. Build Binaries

```bash
# Build all components
make build

# Build WASM modules
make build-wasm
```

This creates:
- `bin/froyo` - CLI tool
- `bin/froyo-agent-{os}-{arch}` - Agent binaries for multiple platforms
- `bin/froyo-runner-{os}-{arch}` - WASM runners for multiple platforms
- `bin/froyo-registry-{os}-{arch}` - Module registry binaries
- `bin/modules/*.wasm` - WASM modules

### 2. Deploy NATS Server

```bash
# Edit inventory with your NATS host details
vim examples/nats-deployment/inventory.yml

# Deploy NATS
./bin/froyo apply examples/nats-deployment/nats-stack.ofy
```

This installs NATS as a systemd service with monitoring.

### 3. Deploy Module Registry

**Option A: Systemd Service (Recommended for Production)**

```bash
# Edit inventory with your registry host details
vim examples/registry-deployment/inventory.yml

# Deploy registry
./bin/froyo apply examples/registry-deployment/systemd-stack.ofy

# Upload WASM modules to registry
./examples/registry-deployment/upload-modules.sh 192.168.1.60
```

**Option B: Docker Compose (Quick Testing)**

```bash
cd examples/registry-deployment

# Start registry with Docker
./local-docker-test.sh

# Upload modules
cp ../../bin/modules/*.wasm ./modules/
docker restart froyo-registry
```

**Option C: Docker via OpenFroyo**

```bash
# Edit inventory
vim examples/registry-deployment/inventory.yml

# Deploy registry in Docker
./bin/froyo apply examples/registry-deployment/docker-stack.ofy
```

### 4. Deploy Agents

```bash
# Edit inventory with your agent host details
vim examples/agent-onboarding/inventory.yml

# Edit agent configuration in the stack
vim examples/agent-onboarding/onboard-agent.ofy
# Update:
#   - nats_url: "nats://192.168.1.50:4222"  # Your NATS server
#   - module_registry_url: "http://192.168.1.60:8080"  # Your registry

# Deploy agents
./bin/froyo apply examples/agent-onboarding/onboard-agent.ofy
```

This will:
1. Install `froyo-agent` and `froyo-runner` binaries on each host
2. Create systemd service
3. Configure NATS connection
4. Configure module registry
5. Start the agent service
6. Verify agent connectivity

### 5. Run Tasks

Create a stack file (e.g., `my-tasks.ofy`):

```yaml
name: My Production Tasks
description: Example tasks using pull mode

inventory:
  - inventory.yml

defaults:
  strategy: parallel
  max_parallel: 10

run:
  # Execute on all web servers
  - name: Update packages
    module: package/apt
    hosts:
      - "@group:webservers"
    vars:
      name: "*"
      state: latest
      update_cache: true

  # Configure nginx
  - name: Install nginx
    module: package/apt
    hosts:
      - "@group:webservers"
    vars:
      name: nginx
      state: present

  # Manage service
  - name: Ensure nginx is running
    module: service
    hosts:
      - "@group:webservers"
    vars:
      name: nginx
      state: started
      enabled: true
```

Execute the stack:

```bash
./bin/froyo apply my-tasks.ofy
```

The CLI will:
1. Connect to NATS
2. Send task messages to agents
3. Agents automatically download required modules from registry (if not cached)
4. Agents execute tasks and return results
5. CLI displays results in real-time

### Monitoring

**NATS Monitoring:**
```bash
# View NATS server stats
curl http://192.168.1.50:8222/varz

# View connections
curl http://192.168.1.50:8222/connz
```

**Registry Monitoring:**
```bash
# Health check
curl http://192.168.1.60:8080/healthz

# Prometheus metrics
curl http://192.168.1.60:8080/metrics

# List available modules
curl http://192.168.1.60:8080/modules
```

**Agent Status:**
```bash
# On agent host
systemctl status froyo-agent

# View logs
journalctl -u froyo-agent -f
```

### Architecture Benefits

With pull mode deployed:

- **Scalability**: NATS can handle 10,000+ concurrent agents
- **Firewall-Friendly**: Agents connect outbound to NATS (no inbound SSH required)
- **Automatic Updates**: Agents auto-download new module versions
- **Self-Healing**: Agents auto-reconnect if NATS restarts
- **Centralized Control**: Single NATS server for entire infrastructure
- **Module Caching**: Modules downloaded once per host, cached locally
- **Secure**: SHA256 checksum verification on all module downloads
- **Observable**: Prometheus metrics on registry, NATS monitoring endpoints

### Deployment Examples

All deployment examples include:
- Full automation stacks (`.ofy` files)
- Systemd service configurations
- Security hardening
- Monitoring setup
- Comprehensive documentation

**Available Examples:**
- `examples/nats-deployment/` - NATS message broker setup
- `examples/registry-deployment/` - Module registry (systemd, Docker, Docker Compose)
- `examples/agent-onboarding/` - Agent deployment and configuration

## Project Structure

```
openfroyo/
├── cmd/
│   ├── froyo/          # Main CLI (push and pull mode)
│   ├── froyo-agent/    # Persistent agent daemon
│   ├── froyo-runner/   # WASM runner (deployed to remote hosts)
│   └── froyo-registry/ # Module registry HTTP server
├── internal/
│   ├── parser/         # Stack and inventory parsers
│   ├── executor/       # Execution engine (push mode)
│   ├── ssh/            # SSH client (push mode)
│   ├── agent/          # Agent implementation (pull mode)
│   │   ├── executor/   # Agent task executor
│   │   ├── nats/       # NATS client
│   │   ├── registry/   # Module registry client
│   │   └── config/     # Agent configuration
│   ├── registry/       # Registry server implementation
│   │   ├── server.go   # HTTP API server
│   │   ├── storage.go  # Module storage layer
│   │   └── config.go   # Registry configuration
│   └── cli/            # CLI implementation
│       ├── push/       # Push mode execution
│       └── pull/       # Pull mode execution (NATS)
├── modules/            # 58 built-in modules organized by category
│   ├── package/        # Package management (apt, yum, dnf, etc.)
│   ├── system/         # System operations (user, group, cron, etc.)
│   ├── network/        # Network configuration (firewall, DNS, etc.)
│   ├── storage/        # Storage management (filesystem, LVM, etc.)
│   ├── service/        # Service management
│   ├── generic/        # Generic operations (file, copy, command, etc.)
│   ├── security/       # Security (SELinux, AppArmor, fail2ban, etc.)
│   ├── database/       # Database management (MySQL, PostgreSQL, etc.)
│   ├── monitoring/     # Monitoring (Prometheus, Grafana, etc.)
│   ├── web/            # Web servers (Apache, Nginx, etc.)
│   ├── container/      # Container management (Docker, Podman)
│   └── cloud/          # Cloud providers (AWS, GCP, Azure, etc.)
└── examples/
    ├── inventory/              # Example inventory files
    ├── stacks/                 # Example stack files (push mode)
    ├── nats-deployment/        # NATS server deployment
    ├── registry-deployment/    # Module registry deployment
    │   ├── systemd-stack.ofy   # Production deployment
    │   ├── docker-stack.ofy    # Docker deployment
    │   ├── docker-compose.yml  # Docker Compose
    │   ├── local-docker-test.sh # Quick Docker test
    │   └── README.md           # Full documentation
    └── agent-onboarding/       # Agent deployment
```

## How It Works

### Push Mode (SSH-Based)

1. **Parse Stack** - Load and parse the stack file (`.ofy`)
2. **Load Inventory** - Load host and group definitions
3. **Execute Run Entries** - For each entry in the stack:
   - Connect to target hosts via SSH
   - Ensure `froyo-runner` binary exists on the host
   - Upload the WASM module
   - Execute: `froyo-runner --module <path>.wasm --input-base64 "<JSON>"`
   - Parse and display JSON output

### Pull Mode (Agent-Based)

1. **Parse Stack** - Load and parse the stack file (`.ofy`)
2. **Connect to NATS** - CLI connects to NATS message broker
3. **Publish Tasks** - For each task:
   - CLI publishes task message to `froyo.tasks.{hostname}` subject
   - Message includes module name, variables, and task context
4. **Agent Receives Task** - Agent subscribed to its hostname subject receives the task
5. **Module Download** - Agent checks if module is cached locally:
   - If not cached or checksum changed: Download from module registry
   - Verify SHA256 checksum
   - Cache module locally
6. **Execute Task** - Agent executes: `froyo-runner --module <path>.wasm --input-base64 "<JSON>"`
7. **Return Results** - Agent publishes results to `froyo.results.{taskid}` subject
8. **CLI Receives Results** - CLI receives and displays results in real-time

## WASM Module Contract

Modules receive JSON input via stdin and output JSON results:

**Input:**
```json
{
  "vars": {"cmd": "uname -r"},
  "context": {
    "host": "test-host",
    "task_name": "Check kernel version"
  }
}
```

**Output:**
```json
{
  "status": "ok|changed|failed",
  "message": "Command executed successfully: uname -r",
  "facts": {
    "stdout": "5.15.0-91-generic"
  }
}
```

## Module Registry API

The Module Registry is a RESTful HTTP server that distributes WASM modules to agents.

### Endpoints

**List Modules**
```bash
GET /modules
```
Returns JSON array of available module names.

**Module Metadata**
```bash
GET /modules/{name}/metadata
```
Returns JSON with module information:
```json
{
  "name": "package/apt",
  "size": 245678,
  "checksum": "a3b2c1d4e5f6...",
  "modified": "2025-01-17T10:30:00Z"
}
```

**Download Module**
```bash
GET /modules/{name}/download
```
Returns WASM binary with headers:
- `Content-Type: application/wasm`
- `X-Module-Checksum: {sha256}`

**Health Check**
```bash
GET /healthz
GET /ready
```
Returns 200 OK if registry is healthy.

**Metrics (Prometheus)**
```bash
GET /metrics
```
Returns Prometheus-formatted metrics.

### Authentication

Registry supports API key authentication (optional):

```yaml
# registry.yml
auth:
  enabled: true
  type: "api_key"
  api_keys:
    - "secret-key-1"
    - "secret-key-2"
```

Clients send API key in header:
```bash
curl -H "X-API-Key: secret-key-1" http://registry:8080/modules
```

### Agent Integration

Agents automatically use the registry if configured:

```yaml
# agent.yml
execution:
  module_registry_url: "http://192.168.1.60:8080"
  module_registry_api_key: "secret-key-1"  # Optional
  module_cache: "/var/cache/froyo/modules"
```

When a task requires a module, the agent:
1. Checks local cache
2. If not cached, fetches metadata from registry
3. Downloads module if checksum changed
4. Verifies SHA256 checksum
5. Caches module locally
6. Executes task

### Deployment Options

See `examples/registry-deployment/README.md` for complete documentation.

**Production (Systemd):**
```bash
./bin/froyo apply examples/registry-deployment/systemd-stack.ofy
```

**Docker Compose:**
```bash
cd examples/registry-deployment
./local-docker-test.sh
```

**Docker (via OpenFroyo):**
```bash
./bin/froyo apply examples/registry-deployment/docker-stack.ofy
```

## Available Modules

OpenFroyo includes 58 production-ready WASM modules organized by category:

### Package Management
- `package/apt` - Debian/Ubuntu package management
- `package/yum` - RHEL/CentOS package management (legacy)
- `package/dnf` - Modern Fedora/RHEL package management
- `package/zypper` - SUSE/openSUSE package management
- `package/apk` - Alpine Linux package management
- `package/pacman` - Arch Linux package management
- `package/gem` - Ruby gem management
- `package/npm` - Node.js package management
- `package/pip` - Python package management
- `package/yarn` - Yarn package management

### System Management
- `system/user` - User account management
- `system/group` - Group management
- `system/cron` - Cron job management
- `system/sysctl` - Kernel parameter configuration
- `system/hostname` - Hostname configuration
- `system/timezone` - Timezone configuration
- `system/locale` - Locale configuration
- `system/kernel_module` - Kernel module management
- `system/reboot` - System reboot management
- `system/mount` - Filesystem mount management
- `system/swap` - Swap space management
- `system/systemd_unit` - Systemd unit management
- `system/bootloader` - Bootloader configuration
- `system/limits` - Resource limits configuration
- `system/tuned` - Performance tuning profiles
- `system/ntp` - Time synchronization

### Network Management
- `network/firewall` - Firewall configuration (iptables/nftables)
- `network/dns` - DNS configuration
- `network/interface` - Network interface configuration
- `network/route` - Routing table management
- `network/bond` - Network bonding configuration
- `network/bridge` - Network bridge configuration
- `network/vlan` - VLAN configuration
- `network/vpn` - VPN configuration

### Storage Management
- `storage/filesystem` - Filesystem management
- `storage/lvm` - LVM volume management
- `storage/disk` - Disk management
- `storage/raid` - Software RAID management
- `storage/partition` - Partition management
- `storage/quota` - Disk quota management
- `storage/nfs` - NFS client/server configuration
- `storage/iscsi` - iSCSI configuration

### Service Management
- `service` - Service state management (systemd/init)
- `service/get_url` - Download files from HTTP/HTTPS

### Generic Operations
- `generic/file` - File/directory management
- `generic/copy` - Copy files between hosts
- `generic/template` - Template file generation
- `generic/command` - Execute shell commands
- `generic/script` - Execute shell scripts
- `generic/archive` - Archive/extract files
- `generic/lineinfile` - Manage lines in files
- `generic/find` - Find files

### Security
- `security/selinux` - SELinux configuration
- `security/apparmor` - AppArmor configuration
- `security/fail2ban` - Fail2ban configuration
- `security/audit` - Audit daemon configuration
- `security/aide` - AIDE intrusion detection
- `security/openssl` - SSL/TLS certificate management

### Database Management
- `database/mysql` - MySQL/MariaDB management
- `database/postgresql` - PostgreSQL management
- `database/redis` - Redis management
- `database/mongodb` - MongoDB management

### Monitoring
- `monitoring/prometheus` - Prometheus monitoring
- `monitoring/grafana` - Grafana dashboards
- `monitoring/node_exporter` - Prometheus node exporter

### Web Servers
- `web/apache` - Apache HTTP server
- `web/nginx` - Nginx web server

### Container Management
- `container/docker` - Docker management
- `container/podman` - Podman management

### Cloud Providers
- `cloud/aws` - AWS resource management
- `cloud/gcp` - Google Cloud Platform
- `cloud/azure` - Microsoft Azure

### Example: Package Management

**Using apt module:**
```yaml
- name: Install nginx
  module: package/apt
  hosts:
    - "@group:webservers"
  vars:
    name: nginx
    state: present
    update_cache: true
```

**Using yum module:**
```yaml
- name: Install httpd
  module: package/yum
  hosts:
    - "@group:rhel-servers"
  vars:
    name: httpd
    state: latest
```

### Example: Service Management

```yaml
- name: Ensure nginx is running
  module: service
  hosts:
    - "@group:webservers"
  vars:
    name: nginx
    state: started
    enabled: true
```

### Example: File Management

```yaml
- name: Create application directory
  module: generic/file
  hosts:
    - "@group:appservers"
  vars:
    path: /opt/myapp
    state: directory
    mode: '0755'
    owner: appuser
    group: appgroup
```

### Example: User Management

```yaml
- name: Create deployment user
  module: system/user
  hosts:
    - "@group:all"
  vars:
    name: deployer
    state: present
    shell: /bin/bash
    groups: ["docker", "sudo"]
    create_home: true
```

## Development

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run tests
make test

# Clean build artifacts
make clean
```

## Current Status

**Production-Ready Features:**
- ✅ Push mode (SSH-based, agentless)
- ✅ Pull mode (Agent-based with NATS)
- ✅ Module Registry (HTTP server for WASM distribution)
- ✅ 58 production-ready WASM modules
- ✅ Multi-platform builds (Linux, macOS, Windows - amd64/arm64)
- ✅ Systemd service deployments (agent, registry, NATS)
- ✅ Docker deployment options
- ✅ GitHub Actions CI/CD
- ✅ Prometheus metrics and monitoring
- ✅ API key authentication for registry
- ✅ SHA256 checksum verification
- ✅ Module caching on agents
- ✅ Automated deployment stacks

**Coming Soon:**
- ⏳ Conditionals (`when:` blocks)
- ⏳ Loops (`loop:` directives)
- ⏳ Diff mode for resource changes
- ⏳ Terraform-style state management
- ⏳ Vault integration for secrets
- ⏳ Web UI for monitoring

See `openfroyo_mvp_spec.md` for the complete architecture specification.

## License

MIT
