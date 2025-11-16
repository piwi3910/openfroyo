# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**OpenFroyo** is an agentless, Go-based automation and orchestration framework that combines concepts from Ansible and Terraform. It uses WebAssembly (WASM) as the execution engine for remote actions over SSH (no agents required).

**Current State:** Early development stage. The repository contains the MVP specification document that defines the complete architecture.

## Architecture Principles

1. **Agentless Push Model** - All execution happens via SSH; no persistent agents on target hosts
2. **WASM-Based Execution** - Remote actions execute as WASM modules via a lightweight `froyo-runner` binary
3. **Stack-Based Orchestration** - Ordered execution model with modules and task blocks
4. **Inventory Management** - Host and group-based targeting similar to Ansible
5. **Go-Only Implementation** - Single language for all components (CLI, orchestrator, runner)

## Key Concepts

### Project Structure
```
openfroyo-project/
  inventory/
    hosts.yml          # SSH connection details for hosts
    groups.yml         # Host group definitions
  stacks/
    *.ofy              # Orchestration definitions (what to run, where, in what order)
  modules/
    <name>/
      module.ofy.yml   # Module definition with steps
      defaults.ofy.yml # Default variables
      wasm/            # WASM binaries for this module
```

### Execution Flow
1. Parse stack file (defines inventory, defaults, and ordered `run:` list)
2. Load inventory (hosts + groups)
3. Execute each `run:` entry in order:
   - Module invocation: Load module and execute its steps
   - Task block: Execute inline WASM tasks
4. For each step on each host:
   - SSH to host
   - Ensure `froyo-runner` binary exists
   - Upload WASM module if needed
   - Execute: `froyo-runner --module <path>.wasm --input-base64 "<JSON>"`
   - Parse JSON output: `{"status": "ok|changed|failed", "message": "...", "facts": {...}}`

### WASM Module Contract

**Input (JSON):**
```json
{
  "vars": { /* user-provided variables */ },
  "context": {
    "host": "web-01",
    "task_name": "Install nginx"
  }
}
```

**Output (JSON):**
```json
{
  "status": "ok|changed|failed",
  "message": "descriptive message",
  "facts": { /* discovered facts */ }
}
```

**Host API** (exposed to WASM modules):
- `host_log(level, msg)`
- `host_exec(cmd, args, timeout)`
- `host_file_read(path)`
- `host_file_write(path, contents, mode)`
- `host_file_stat(path)`
- `host_http_request(method, url, headers, body)`
- `host_env_get(name)`

## CLI Commands (Planned)

```bash
# Run a stack
froyo apply stacks/web_stack.ofy

# Dry run (plan mode)
froyo plan stacks/web_stack.ofy

# Partial execution
froyo apply stacks/web_stack.ofy --until "Database layer"
froyo apply stacks/web_stack.ofy --from "Deploy app code"
```

## Development Commands

Since the codebase is in early stages, standard Go commands will be used:

```bash
# Build main CLI
go build -o froyo ./cmd/froyo

# Build froyo-runner (static binary for remote hosts)
CGO_ENABLED=0 go build -ldflags="-s -w" -o froyo-runner ./cmd/froyo-runner

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if golangci-lint is set up)
golangci-lint run
```

## Implementation Roadmap (MVP)

**Core Components to Implement:**
1. **Stack Parser** - Parse `.ofy.yml` stack files
2. **Inventory Loader** - Load and resolve hosts/groups from inventory files
3. **Execution Engine** - Execute ordered `run:` entries with proper strategy (serial/parallel)
4. **SSH Executor** - SSH connection management and command execution
5. **WASM Runner** - Go binary that loads and executes WASM modules on remote hosts
6. **Host API** - Provide host functions to WASM modules (file ops, exec, HTTP, etc.)
7. **Module Loader** - Load and execute module definitions
8. **Task Block Handler** - Execute inline task blocks with WASM modules

**Out of Scope for MVP:**
- Conditionals (`when:`)
- Loops (`loop:`)
- Pull/agent mode
- Diff mode for resources
- Providers (Terraform-style state management)

## File Formats

OpenFroyo uses different file extensions for different purposes:
- **Stack files** (`stacks/*.ofy`): Orchestration definitions (OpenFroyo-specific)
- **Inventory files** (`inventory/*.yml`): Host and group definitions (standard YAML)
- **Module files** (`modules/*/module.ofy.yml`): Module definitions (OpenFroyo-specific)
- **Variable files** (`defaults.ofy.yml`, `vars.ofy.yml`): Variable definitions

## Execution Behaviors

- **Parallelism**: Controlled via `strategy: parallel` or `strategy: serial`
- **Max Parallel**: `max_parallel: N` limits concurrent host operations
- **Failure Handling**: `continue_on_error: false` (default), `max_fail_percentage: X`
- **Host Targeting**: Use `@group:name` to target groups, or specific host names

## Technology Stack

- **Language**: Go (all components)
- **Execution Engine**: WebAssembly (WASM/WASI)
- **Transport**: SSH (no agents)
- **Config Format**: YAML (`.ofy.yml` files)
- **Variable Templating**: `{{ var.name | default('value') }}` syntax
