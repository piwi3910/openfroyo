# OpenFroyo Quick Start Guide

This guide will help you get OpenFroyo up and running in minutes.

## Prerequisites

1. **Go 1.21+** - [Download](https://go.dev/dl/)
2. **TinyGo** - [Installation Guide](https://tinygo.org/getting-started/install/)
3. **SSH access to a Linux host** - For testing

### Installing TinyGo

**macOS (Homebrew):**
```bash
brew tap tinygo-org/tools
brew install tinygo
```

**Ubuntu/Debian:**
```bash
wget https://github.com/tinygo-org/tinygo/releases/download/v0.30.0/tinygo_0.30.0_amd64.deb
sudo dpkg -i tinygo_0.30.0_amd64.deb
```

**Other platforms:** See [TinyGo installation docs](https://tinygo.org/getting-started/install/)

## Step-by-Step Setup

### 1. Build OpenFroyo

```bash
# Build the CLI and runner binaries
make build

# Build the WASM modules
make build-wasm
```

You should now have:
- `bin/froyo` - The CLI tool (6.8MB)
- `bin/froyo-runner` - The runner binary (5.0MB, deployed to remote hosts)
- `modules/exec/wasm/exec.wasm` - The exec module (518KB)

### 2. Configure Your Inventory

Edit `examples/inventory/hosts.yml` with your SSH host details:

```yaml
hosts:
  test-host:
    ssh_host: 192.168.1.100  # Replace with your host IP
    ssh_port: 22
    ssh_user: root           # Replace with your SSH user
    ssh_key_file: ~/.ssh/id_rsa  # Path to your SSH key
    # Or use password authentication:
    # ssh_password: your-password
```

### 3. Run Your First Stack

```bash
./bin/froyo apply examples/stacks/test.ofy
```

**Expected output:**
```
Loading stack: examples/stacks/test.ofy
Loading inventory...
Loaded 1 hosts, 0 groups

Executing stack: Test Stack

[1/1] Check kernel version
  → test-host (192.168.1.100)
    ✓ ok: Command executed successfully: uname -r
      stdout: 5.15.0-91-generic

Stack execution completed successfully!
```

## What Just Happened?

1. **Parsed the stack** - Loaded `test.ofy` configuration
2. **Connected via SSH** - Connected to your host using SSH
3. **Deployed froyo-runner** - Transferred the runner binary to `/tmp/froyo/`
4. **Uploaded WASM module** - Transferred `exec.wasm` to the host
5. **Executed the module** - Ran `froyo-runner --module exec.wasm --input-base64 ...`
6. **Returned results** - Got the kernel version from `uname -r`

## Next Steps

### Create Your Own Stack

Create `my-stack.ofy`:

```yaml
name: My First Stack
inventory:
  - examples/inventory/hosts.yml

run:
  - name: Get system info
    module: exec
    hosts:
      - test-host
    vars:
      cmd: hostname

  - name: Check disk space
    module: exec
    hosts:
      - test-host
    vars:
      cmd: df -h /
```

Run it:
```bash
./bin/froyo apply my-stack.ofy
```

### Working with Multiple Hosts

Edit your inventory to add more hosts:

```yaml
hosts:
  web-01:
    ssh_host: 192.168.1.101
    ssh_port: 22
    ssh_user: ubuntu
    ssh_key_file: ~/.ssh/id_rsa

  web-02:
    ssh_host: 192.168.1.102
    ssh_port: 22
    ssh_user: ubuntu
    ssh_key_file: ~/.ssh/id_rsa

groups:
  webservers:
    hosts:
      - web-01
      - web-02
```

Target the group in your stack:

```yaml
- name: Check all web servers
  module: exec
  hosts:
    - "@group:webservers"
  vars:
    cmd: uptime
```

## Troubleshooting

### SSH Connection Fails

- Verify the host is reachable: `ping <host>`
- Test SSH manually: `ssh user@host`
- Check SSH key permissions: `chmod 600 ~/.ssh/id_rsa`
- Verify SSH key is in `~/.ssh/authorized_keys` on remote host

### Module Fails to Execute

- Check that the WASM module is built: `ls -l modules/exec/wasm/exec.wasm`
- Verify the remote host can execute binaries in `/tmp/froyo/`
- Check if `/tmp/froyo/froyo-runner` has execute permissions

### TinyGo Build Fails

- Ensure TinyGo is installed: `tinygo version`
- Check TinyGo is in your PATH
- Try building manually: `cd modules/exec && tinygo build -o wasm/exec.wasm -target=wasi wasm/main.go`

## Development

### Run Tests
```bash
go test ./...
```

### Format Code
```bash
go fmt ./...
```

### Clean Build Artifacts
```bash
make clean
```

## Architecture Overview

```
┌─────────────┐
│  froyo CLI  │  Parses stack, manages execution
└──────┬──────┘
       │ SSH
       ▼
┌─────────────────┐
│  Remote Host    │
│                 │
│ /tmp/froyo/     │
│  ├─ froyo-runner│  WASM runtime
│  └─ exec.wasm   │  WASM module
└─────────────────┘
```

The froyo CLI:
1. Connects to hosts via SSH
2. Transfers `froyo-runner` (once per host)
3. Uploads WASM modules as needed
4. Executes modules via froyo-runner
5. Collects and displays results

## Learn More

- Full specification: `openfroyo_mvp_spec.md`
- Module development: `modules/exec/README.md`
- Architecture details: `README.md`
