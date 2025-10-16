# OpenFroyo CLI

The `froyo` command-line interface for OpenFroyo Infrastructure Orchestration Engine.

## Building

### Standard Build

```bash
make build
```

This creates the binary at `bin/froyo`.

### Build with Custom Version

```bash
make build VERSION=1.0.0
```

### Cross-Platform Build

```bash
make build-cross
```

This builds binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

## Command Structure

The CLI is organized into the following command hierarchy:

```
froyo
├── init              - Initialize workspace
├── validate          - Validate CUE configs
├── plan              - Generate execution plan
├── apply             - Execute plan
├── run               - Run action/runbook
├── drift
│   ├── detect        - Detect configuration drift
│   └── reconcile     - Reconcile drift
├── onboard
│   ├── ssh           - Onboard host via SSH
│   └── rollback      - Rollback onboarding
├── backup            - Backup data
├── restore           - Restore data
├── dev
│   ├── up            - Start dev environment
│   └── down          - Stop dev environment
└── facts
    ├── collect       - Collect facts
    ├── list          - List facts
    └── show          - Show fact details
```

## Command Details

### Initialize Workspace

```bash
froyo init --solo
```

Initializes a standalone workspace with SQLite database, local storage, and generates age keypair.

### Validate Configurations

```bash
froyo validate [path]
froyo validate --strict --schema ./schema.cue ./configs
```

Validates CUE configuration files against schemas and OPA policies.

### Plan and Apply

```bash
# Generate plan
froyo plan --out plan.json

# Generate plan with DOT graph
froyo plan --out plan.json --dot plan.dot

# Apply plan
froyo apply --plan plan.json

# Auto-approve and apply
froyo apply --plan plan.json --auto-approve
```

### Run Actions

```bash
# Run action
froyo run restart-nginx

# Run with parameters
froyo run deploy --param version=1.2.3 --param env=production

# Run on specific targets
froyo run health-check --target web1 --target web2
```

### Drift Detection

```bash
# Detect drift
froyo drift detect

# Detect drift on specific targets
froyo drift detect --target web1 --target web2

# Auto-reconcile drift
froyo drift detect --auto-reconcile

# Reconcile detected drift
froyo drift reconcile
```

### Host Onboarding

```bash
# Basic onboarding
froyo onboard ssh --host 10.0.0.42 --user root --password secret

# Full onboarding with hardening
froyo onboard ssh \
  --host 10.0.0.42 \
  --user root \
  --password s3cr3t \
  --key default-ed25519 \
  --create-user froyo \
  --sudo 'NOPASSWD: /usr/bin/systemctl,/usr/bin/apt' \
  --lock-down \
  --labels env=dev,role=web

# Rollback onboarding
froyo onboard rollback --host 10.0.0.42
```

### Backup and Restore

```bash
# Create backup
froyo backup --out backup.tar.gz

# Restore from backup
froyo restore --from backup.tar.gz

# Force restore without confirmation
froyo restore --from backup.tar.gz --force
```

### Development Mode

```bash
# Start controller and worker
froyo dev up

# Start controller only
froyo dev up --controller-only

# Start with multiple workers
froyo dev up --workers 3

# Stop dev environment
froyo dev down
```

### Facts Collection

```bash
# Collect all facts
froyo facts collect

# Collect from specific hosts
froyo facts collect --target web1 --target web2

# Collect specific fact types
froyo facts collect --type os.basic --type hw.cpu

# Collect using selector
froyo facts collect --selector 'env=prod,role=web'

# List collected facts
froyo facts list

# Show facts for a host
froyo facts show --target web1
```

## Global Flags

All commands support these global flags:

- `--config, -c`: Config file path
- `--verbose, -v`: Enable verbose output
- `--json`: Output in JSON format
- `--version`: Show version information

## Environment Variables

- `LOG_LEVEL`: Set logging level (debug, info, warn, error)

## Architecture

The CLI is built using:

- **Cobra**: Command-line framework
- **Zerolog**: Structured logging
- **Context**: Graceful shutdown and cancellation

### Directory Structure

```
cmd/froyo/
├── main.go              - Entry point with logging and signal handling
└── commands/
    ├── root.go          - Root command definition
    ├── init.go          - Init command
    ├── validate.go      - Validate command
    ├── plan.go          - Plan command
    ├── apply.go         - Apply command
    ├── run.go           - Run command
    ├── drift.go         - Drift commands (detect, reconcile)
    ├── onboard.go       - Onboarding commands (ssh, rollback)
    ├── backup.go        - Backup command
    ├── restore.go       - Restore command
    ├── dev.go           - Dev commands (up, down)
    └── facts.go         - Facts commands (collect, list, show)
```

## Implementation Status

Currently, all commands have:
- ✅ Full command structure and flag definitions
- ✅ Comprehensive help text and examples
- ✅ Proper error handling framework
- ✅ Structured logging integration
- ⏳ Placeholder implementations (show "Not implemented yet")

Each command is ready for business logic implementation according to the ENGINE_MVP.md specification.

## Next Steps

To implement command functionality, each command file needs:

1. Import relevant packages from `pkg/` (engine, stores, providers, etc.)
2. Replace placeholder implementation with actual logic
3. Add unit tests
4. Update this documentation with implementation notes

## Testing

```bash
# Build and test version info
make build
./bin/froyo --version

# Test help output
./bin/froyo --help
./bin/froyo init --help

# Test command execution (placeholder)
./bin/froyo init --solo
```
