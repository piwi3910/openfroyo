# OpenFroyo

An agentless, Go-based automation and orchestration framework that combines concepts from Ansible and Terraform. It uses WebAssembly (WASM) as the execution engine for remote actions over SSH.

## Features

- **Agentless** - No agents required on target hosts, uses SSH
- **WASM-Based Execution** - Remote actions execute as WASM modules via lightweight `froyo-runner` binary
- **Stack-Based Orchestration** - Ordered execution model with modules and task blocks
- **Go Implementation** - Single language for all components

## Quick Start

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

## Project Structure

```
openfroyo/
├── cmd/
│   ├── froyo/          # Main CLI
│   └── froyo-runner/   # WASM runner (deployed to remote hosts)
├── internal/
│   ├── parser/         # Stack and inventory parsers
│   ├── executor/       # Execution engine
│   └── ssh/            # SSH client
├── modules/
│   └── exec/           # Execute shell commands module
│       ├── wasm/       # WASM binary
│       └── module.ofy.yml
└── examples/
    ├── inventory/      # Example inventory files
    └── stacks/         # Example stack files
```

## How It Works

1. **Parse Stack** - Load and parse the stack file (`.ofy.yml`)
2. **Load Inventory** - Load host and group definitions
3. **Execute Run Entries** - For each entry in the stack:
   - Connect to target hosts via SSH
   - Ensure `froyo-runner` binary exists on the host
   - Upload the WASM module
   - Execute: `froyo-runner --module <path>.wasm --input-base64 "<JSON>"`
   - Parse and display JSON output

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

## Available Modules

### exec

Execute shell commands on remote hosts.

**Variables:**
- `cmd` (required): The shell command to execute

**Example:**
```yaml
- name: Check kernel version
  module: exec
  hosts:
    - test-host
  vars:
    cmd: uname -r
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

This is an MVP implementation demonstrating the core workflow. See `openfroyo_mvp_spec.md` for the complete architecture specification.

## License

MIT
