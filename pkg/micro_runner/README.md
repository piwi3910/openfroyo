# Micro-Runner

The OpenFroyo micro-runner is a lightweight, ephemeral execution agent that runs on remote hosts to execute commands via JSON-over-stdio communication.

## Features

- **Agentless**: Self-contained binary that doesn't require installation
- **Self-Deleting**: Automatically removes itself after execution
- **Secure**: Static binary with signature verification support
- **Minimal**: < 10 MB binary size (~2.3 MB)
- **Time-Limited**: Automatic TTL (10 minutes) to prevent runaway processes
- **Cross-Platform**: Builds for Linux, macOS, Windows

## Components

### `/protocol`
Protocol definitions and message codecs for JSON-over-stdio communication.

- `types.go` - Message and command type definitions
- `codec.go` - Encoder/decoder for protocol messages
- Tests with >90% coverage

### `/handlers`
Command handlers for each operation type:

- `exec.go` - Shell command execution
- `file.go` - File read/write operations
- `package.go` - Package management (apt, dnf, yum, zypper)
- `service.go` - Systemd service management
- `sudoers.go` - Sudoers rule management
- `sshd.go` - SSH configuration hardening

### `/client`
Client library for communicating with the runner from Go applications.

- `client.go` - High-level client with lifecycle management
- Transport abstraction for SSH, WinRM, etc.

## Quick Start

### Build the Runner

```bash
# For Linux amd64 (most common)
make build-runner-linux-amd64

# For all platforms
make build-runner-all
```

### Use the Client Library

```go
import (
    "github.com/openfroyo/openfroyo/pkg/micro_runner/client"
    "github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// Create and start client
cfg := &client.Config{
    Transport:  mySSHTransport,
    RunnerPath: "./micro-runner-linux-amd64",
    RemotePath: "/tmp/micro-runner",
}

c, _ := client.NewClient(cfg)
c.Start(ctx, cfg)
defer c.Close(ctx, cfg.RemotePath)

// Execute command
cmd := &protocol.CommandMessage{
    ID:      "cmd-1",
    Type:    protocol.CommandTypeExec,
    Timeout: 30,
    Params:  mustMarshal(protocol.ExecParams{
        Command:    "uptime",
        CaptureOut: true,
    }),
}

done, _ := c.Execute(ctx, cmd)
```

## Documentation

- [Protocol Specification](../../docs/micro-runner-protocol.md) - Complete protocol documentation
- [Usage Guide](../../docs/micro-runner-usage.md) - Examples and best practices

## Architecture

```
Controller                  Remote Host
----------                  -----------
    |                           |
    | 1. Upload runner         |
    |------------------------->|
    |                           |
    | 2. Start runner          |
    |------------------------->| ./micro-runner
    |                           |
    | 3. READY message         |
    |<-------------------------|
    |                           |
    | 4. CMD message           |
    |------------------------->|
    |                           |
    | 5. EVENT messages        |
    |<-------------------------| (optional progress)
    |                           |
    | 6. DONE/ERROR message    |
    |<-------------------------|
    |                           |
    | 7. Close stdin           |
    |------------------------->|
    |                           |
    | 8. EXIT message          |
    |<-------------------------|
    |                           | (self-delete)
```

## Command Types

1. **exec** - Execute shell commands
2. **file.write** - Write files with permissions
3. **file.read** - Read file content
4. **pkg.ensure** - Install/remove packages
5. **service.reload** - Manage systemd services
6. **sudoers.ensure** - Configure sudoers
7. **sshd.harden** - Apply SSH hardening

## Security

- Static binary with no dynamic dependencies
- Signature verification for commands (optional)
- TTL enforcement (10 minutes max)
- Self-delete on exit
- Timeouts on all operations
- Capability-based security model

## Testing

```bash
# Run all tests
go test ./pkg/micro_runner/... -v

# Run with coverage
go test ./pkg/micro_runner/... -cover

# Run specific test
go test ./pkg/micro_runner/protocol -run TestEncoder
```

## Building

The runner is built as a static binary with CGO disabled:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o micro-runner-linux-amd64 \
  ./cmd/micro-runner
```

Build flags:
- `CGO_ENABLED=0` - No C dependencies
- `-ldflags="-s -w"` - Strip debug info and symbols
- Static linking for portability

## Performance

- Binary size: ~2.3 MB
- Startup time: < 100ms
- Memory usage: < 10 MB
- Command execution overhead: < 5ms

## Limitations

- Maximum runtime: 10 minutes (TTL)
- Single command at a time (serial execution)
- No persistent state between runs
- Requires write access to temporary directory

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines.

## License

See [LICENSE](../../LICENSE) file.
