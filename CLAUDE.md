# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**OpenFroyo** is a next-generation Infrastructure-as-Code (IaC) orchestration engine written in Go. It combines declarative state management (like Terraform) with procedural configuration capabilities (like Ansible), modernized with:

- Typed configurations via CUE
- Light procedural scripting via Starlark
- WASM-based provider system for security and portability
- Ephemeral micro-runner for complex local operations
- SQLite-based persistence for solo deployments

**Current State:** Repository contains design documents only. No code has been implemented yet.

## Architecture Principles

1. **Start local, scale later** - Begin with a single binary using SQLite and filesystem, designed to scale to distributed systems
2. **Go-only core** - Single language for CLI, controller, worker, and micro-runner (no polyglot complexity in MVP)
3. **WASM for providers** - Safe, sandboxed execution of plugins with capability-based security
4. **Agentless-first** - Default to ephemeral micro-runner over persistent agents
5. **Typed everything** - CUE for config schema validation, strong typing throughout

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go | All core components |
| Plugins | WASM/WASI via Wasmtime | Provider execution sandbox |
| Config DSL | CUE + Starlark | Declarative configs + minimal scripting |
| Persistence | SQLite (WAL mode) | State, metadata, events (solo profile) |
| Queue | Badger/Pebble | Embedded at-least-once delivery queue |
| Secrets | age keypair | Envelope encryption |
| Policy | OPA (rego) | Policy-as-code enforcement |
| Telemetry | OpenTelemetry + Zerolog | Structured logs and distributed traces |

## Repository Structure

```
openfroyo/
├── cmd/
│   └── froyo/              # Main CLI entry point (controller/worker/cli)
├── pkg/
│   ├── engine/             # Core: planner, DAG builder, scheduler
│   ├── stores/             # Persistence adapters (SQLite, FS, queue)
│   ├── providers/host/     # WASM host runtime for providers
│   ├── transports/         # SSH, WinRM, API communication
│   ├── micro_runner/       # Micro-runner client and protocol
│   ├── policy/             # OPA integration
│   ├── telemetry/          # OpenTelemetry and logging setup
│   └── api/                # gRPC/REST service definitions
├── providers/              # WASM provider implementations
│   ├── linux.pkg/          # Package management provider
│   ├── linux.service/      # Service management provider
│   ├── linux.file/         # File operations provider
│   └── probe.http/         # HTTP probe provider
├── examples/               # Demo configurations (Apache module, etc.)
└── docs/                   # Design documents and guides
```

## Development Commands

### Build
```bash
# Build main CLI
go build -o froyo ./cmd/froyo

# Build with all platforms
GOOS=linux GOARCH=amd64 go build -o froyo-linux-amd64 ./cmd/froyo
GOOS=darwin GOARCH=arm64 go build -o froyo-darwin-arm64 ./cmd/froyo

# Build micro-runner (static binary)
CGO_ENABLED=0 go build -ldflags="-s -w" -o micro-runner ./cmd/micro-runner
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./pkg/engine/...

# Run with race detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Linting and Quality
```bash
# Run golangci-lint
golangci-lint run

# Run with all linters
golangci-lint run --enable-all

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### WASM Provider Development
```bash
# Build a provider to WASM
tinygo build -o providers/linux.pkg/plugin.wasm -target=wasi providers/linux.pkg/main.go

# Validate provider manifest
go run ./cmd/froyo validate-provider ./providers/linux.pkg
```

### Local Development
```bash
# Initialize local workspace
./froyo init --solo

# Run controller + worker locally
./froyo dev up

# Validate CUE configs
./froyo validate ./examples/apache/

# Generate plan
./froyo plan --out plan.json ./examples/apache/

# Apply plan
./froyo apply --plan plan.json
```

## Core Execution Flow

The engine follows a strict 6-phase workflow:

1. **Evaluate** - Parse CUE configs, run Starlark helpers, validate schemas and policies
2. **Discover** - Collect facts from targets via API/SSH/WinRM/micro-runner
3. **Plan** - Compute diffs (Desired vs Actual), build DAG of Plan Units (PUs)
4. **Apply** - Execute DAG in parallel, respecting dependencies (`require`/`notify`)
5. **Post-Apply** - Trigger handlers/actions, run smoke tests
6. **Drift** - Periodically compare Actual vs State, auto-reconcile or alert

## Micro-Runner Protocol

The micro-runner is a critical component - a tiny (<10 MB) static Go binary that:

- Runs via SSH/WinRM with JSON-over-stdio communication
- Self-deletes after execution (no persistent agent)
- Supports commands: `exec`, `file.write`, `pkg.ensure`, `service.reload`, etc.
- Uses frame-based protocol: `READY → CMD → EVENT → DONE/ERROR → EXIT`
- Has built-in signature verification (cosign) and TTL (10 min)

**Important:** This is the default path for any "complex local" operations. Design with micro-runner delegation in mind.

## Provider System

Providers are packaged as OCI images containing:
```
/plugin/plugin.wasm
/plugin/manifest.yaml
/schemas/*.json
/LICENSE
/SBOM.json
```

**Required interface:**
```go
Init()    error
Read()    (Actual, error)
Plan()    (Ops, error)
Apply()   (Result, error)
Destroy() (Result, error)
```

**Key concepts:**
- Capability-based security declared in manifest
- JSON Schema validation for inputs/outputs
- Error classification: `transient`, `throttled`, `conflict`, `permanent`
- All execution happens in WASM sandbox with host-enforced limits

## Database Schema (Solo Profile)

Core tables in SQLite:
- `runs` - Execution runs and their status
- `plan_units` - Individual plan unit records
- `events` - Append-only event log
- `resource_state` - Current state of managed resources
- `facts` - Discovered facts with TTL
- `audit` - Audit trail for compliance

**Important:** Use WAL mode for SQLite, advisory locks for concurrency.

## MVP Scope Exclusions

The following are **explicitly excluded** from MVP:
- Web GUI/Console (CLI only)
- Multi-cluster controller
- Cloud providers (AWS/GCP/Azure)
- Windows support
- Persistent agents (pull mode)
- Centralized policy service
- OIDC/SSO/multi-user auth
- Metrics dashboards

Focus on solo-profile, agentless, Linux-only for MVP.

## Development Milestones

**M1 (weeks 1-3) - Core Engine:**
- Implement CLI scaffolding and command structure
- Build CUE parser and Starlark integration
- Create SQLite store and basic DAG planner
- Develop WASM runtime and initial Linux providers
- Implement SSH transport and micro-runner protocol
- Create working Apache module demo

**M2 (weeks 4-6) - Reliability & Policy:**
- Add retry logic, idempotency, and error handling
- Implement drift detection
- Integrate OPA policy enforcement
- Add facts caching with TTL
- Set up logging and OpenTelemetry tracing

**M3 (weeks 7-8) - Polish:**
- Create provider SDK and documentation
- Implement backup/restore functionality
- Refine error classification and structured events
- Package binaries for all platforms
- Write comprehensive examples and guides

## Code Style Guidelines

- Follow standard Go conventions and idiomatic patterns
- Use `golangci-lint` with strict settings
- All public APIs must have godoc comments
- Prefer explicit error handling over panics
- Use context.Context for cancellation and deadlines
- Keep packages focused and single-purpose
- Minimize dependencies (vendor critical libs)

## Security Considerations

- All micro-runner binaries must be signed with cosign
- Provider WASM modules run in capability-restricted sandbox
- Use `age` for all secret encryption at rest
- SSH key-based auth only (password auth only for initial onboarding)
- Validate all CUE schemas before execution
- Enforce OPA policies before any state changes
- Audit log all privileged operations

## Testing Strategy

- Unit tests for all core engine logic
- Integration tests for provider execution
- End-to-end tests for CLI workflows
- Mock WASM providers for testing host runtime
- Test SSH/WinRM transports with test containers
- Validate micro-runner protocol with golden files
- Policy tests using OPA test framework
