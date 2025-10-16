# Micro-Runner Usage Guide

## Overview

This guide demonstrates how to use the OpenFroyo micro-runner client library to execute commands on remote hosts.

## Client Library

The client library (`pkg/micro_runner/client`) provides a high-level interface for:
- Uploading the runner binary to remote hosts
- Starting the runner process
- Sending commands and receiving results
- Managing the runner lifecycle

## Basic Usage

### 1. Create a Transport

First, implement the `Transport` interface for your connection method (SSH, WinRM, etc.):

```go
type Transport interface {
    Upload(ctx context.Context, localPath, remotePath string) error
    Execute(ctx context.Context, remotePath string) (stdin io.WriteCloser, stdout io.ReadCloser, err error)
    Cleanup(ctx context.Context, remotePath string) error
}
```

### 2. Initialize the Client

```go
import (
    "context"
    "time"
    "github.com/openfroyo/openfroyo/pkg/micro_runner/client"
    "github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

// Create client configuration
cfg := &client.Config{
    Transport:      myTransport,
    RunnerPath:     "./micro-runner-linux-amd64",
    RemotePath:     "/tmp/micro-runner",
    StartupTimeout: 10 * time.Second,
}

// Create and start the client
c, err := client.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}

// Start the runner (uploads and executes)
if err := c.Start(context.Background(), cfg); err != nil {
    log.Fatal(err)
}
defer c.Close(context.Background(), cfg.RemotePath)

// Check runner capabilities
ready := c.Ready()
fmt.Printf("Runner version: %s\n", ready.Version)
fmt.Printf("Platform: %s/%s\n", ready.Platform, ready.Arch)
```

### 3. Execute Commands

```go
// Execute a simple shell command
cmd := &protocol.CommandMessage{
    ID:      "cmd-001",
    Type:    protocol.CommandTypeExec,
    Timeout: 30,
    Params:  mustMarshal(protocol.ExecParams{
        Command:    "uptime",
        CaptureOut: true,
        CaptureErr: true,
    }),
}

done, err := c.Execute(context.Background(), cmd)
if err != nil {
    log.Fatal(err)
}

var result protocol.ExecResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Exit Code: %d\n", result.ExitCode)
fmt.Printf("Output: %s\n", result.Stdout)
```

## Command Examples

### Execute Shell Command

```go
import "encoding/json"

params := protocol.ExecParams{
    Command:    "apt-get",
    Args:       []string{"update"},
    CaptureOut: true,
    CaptureErr: true,
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-exec-1",
    Type:    protocol.CommandTypeExec,
    Timeout: 300, // 5 minutes
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)
```

### Write a File

```go
params := protocol.FileWriteParams{
    Path:    "/etc/myapp/config.yaml",
    Content: "server:\n  port: 8080\n",
    Mode:    "0644",
    Owner:   "root",
    Group:   "root",
    Backup:  true,
    Create:  true,
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-write-1",
    Type:    protocol.CommandTypeFileWrite,
    Timeout: 30,
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.FileWriteResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Wrote %d bytes, checksum: %s\n", result.BytesWritten, result.Checksum)
```

### Read a File

```go
params := protocol.FileReadParams{
    Path:     "/etc/hostname",
    MaxBytes: 1024,
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-read-1",
    Type:    protocol.CommandTypeFileRead,
    Timeout: 10,
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.FileReadResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Content: %s\n", result.Content)
```

### Install a Package

```go
params := protocol.PkgEnsureParams{
    Name:    "nginx",
    State:   "present",
    Manager: "apt", // or leave empty for auto-detection
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-pkg-1",
    Type:    protocol.CommandTypePkgEnsure,
    Timeout: 600, // 10 minutes for package installation
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.PkgEnsureResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Action: %s, Version: %s\n", result.Action, result.InstalledVersion)
```

### Reload a Service

```go
params := protocol.ServiceReloadParams{
    Name:   "nginx",
    Action: "reload",
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-service-1",
    Type:    protocol.CommandTypeServiceReload,
    Timeout: 30,
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.ServiceReloadResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Service %s, status: %s\n", result.Action, result.Status)
```

### Configure Sudoers

```go
params := protocol.SudoersEnsureParams{
    User: "froyo",
    Commands: []string{
        "/usr/bin/systemctl",
        "/usr/bin/apt",
    },
    NoPasswd: true,
    State:    "present",
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-sudoers-1",
    Type:    protocol.CommandTypeSudoersEnsure,
    Timeout: 10,
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.SudoersEnsureResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Sudoers file: %s, action: %s\n", result.FilePath, result.Action)
```

### Harden SSH Configuration

```go
params := protocol.SSHDHardenParams{
    DisablePasswordAuth: true,
    DisableRootLogin:    true,
    AllowUsers:          []string{"froyo", "admin"},
    TestConnection:      true,
}

paramsJSON, _ := json.Marshal(params)

cmd := &protocol.CommandMessage{
    ID:      "cmd-sshd-1",
    Type:    protocol.CommandTypeSSHDHarden,
    Timeout: 30,
    Params:  paramsJSON,
}

done, err := c.Execute(ctx, cmd)

var result protocol.SSHDHardenResult
json.Unmarshal(done.Result, &result)
fmt.Printf("Modified keys: %v\n", result.ModifiedKeys)
```

## Handling Events

To receive progress events during command execution:

```go
eventCh := make(chan *protocol.EventMessage, 10)

go func() {
    for evt := range eventCh {
        fmt.Printf("[%s] %s\n", evt.Level, evt.Message)
        if evt.Progress != nil {
            fmt.Printf("Progress: %d/%d %s\n",
                evt.Progress.Current,
                evt.Progress.Total,
                evt.Progress.Unit)
        }
    }
}()

done, err := c.ExecuteWithEvents(ctx, cmd, eventCh)
close(eventCh)
```

## Error Handling

```go
done, err := c.Execute(ctx, cmd)
if err != nil {
    // Command failed or communication error
    fmt.Printf("Error: %v\n", err)
    return
}

// Check result for command-specific errors
var result protocol.ExecResult
if err := json.Unmarshal(done.Result, &result); err != nil {
    fmt.Printf("Failed to parse result: %v\n", err)
    return
}

if result.ExitCode != 0 {
    fmt.Printf("Command failed with exit code %d\n", result.ExitCode)
    fmt.Printf("stderr: %s\n", result.Stderr)
}
```

## Idempotency

For idempotent operations, use the idempotency key:

```go
cmd := &protocol.CommandMessage{
    ID:             "cmd-001",
    Type:           protocol.CommandTypeFileWrite,
    IdempotencyKey: "config-update-v1",
    Timeout:        30,
    Params:         paramsJSON,
}
```

If the same idempotency key is used again, the runner can:
1. Skip execution if the operation was already performed
2. Return cached result from previous execution

## Complete Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/openfroyo/openfroyo/pkg/micro_runner/client"
    "github.com/openfroyo/openfroyo/pkg/micro_runner/protocol"
)

func main() {
    ctx := context.Background()

    // Setup transport (implement for your use case)
    transport := NewSSHTransport("user@host")

    // Create client
    cfg := &client.Config{
        Transport:      transport,
        RunnerPath:     "./micro-runner-linux-amd64",
        RemotePath:     "/tmp/micro-runner",
        StartupTimeout: 10 * time.Second,
    }

    c, err := client.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Start runner
    if err := c.Start(ctx, cfg); err != nil {
        log.Fatal(err)
    }
    defer c.Close(ctx, cfg.RemotePath)

    // Execute command
    params := protocol.ExecParams{
        Command:    "hostname",
        CaptureOut: true,
        CaptureErr: true,
    }
    paramsJSON, _ := json.Marshal(params)

    cmd := &protocol.CommandMessage{
        ID:      "cmd-001",
        Type:    protocol.CommandTypeExec,
        Timeout: 10,
        Params:  paramsJSON,
    }

    done, err := c.Execute(ctx, cmd)
    if err != nil {
        log.Fatal(err)
    }

    var result protocol.ExecResult
    json.Unmarshal(done.Result, &result)

    fmt.Printf("Hostname: %s\n", result.Stdout)
    fmt.Printf("Duration: %.3fs\n", done.Duration)
}
```

## Best Practices

1. **Always set timeouts**: Every command should have a reasonable timeout
2. **Handle errors gracefully**: Check both the error return and command exit codes
3. **Use context**: Pass context for cancellation and deadline propagation
4. **Close the client**: Always defer `Close()` to clean up resources
5. **Check capabilities**: Verify the runner supports the command type before sending
6. **Limit concurrency**: Don't send multiple commands in parallel to a single runner
7. **Monitor events**: Use `ExecuteWithEvents` for long-running operations
8. **Secure transport**: Always use encrypted transports (SSH, WinRM with TLS)
9. **Verify runner binary**: Check SHA256 hash before upload
10. **Log operations**: Log all commands and results for audit trail

## Troubleshooting

### Runner doesn't start

- Check runner binary has execute permissions
- Verify remote path is writable
- Check transport connection is working
- Increase startup timeout if needed

### Commands timeout

- Increase command timeout
- Check remote host performance
- Verify operation isn't blocked (e.g., waiting for input)

### Self-delete fails

- Runner may not have permissions to delete itself
- Some platforms don't allow active binaries to be deleted
- This is non-fatal; the runner will still exit

### Connection lost

- Runner may have hit TTL (10 minutes)
- Remote host may have terminated the process
- Network connection may have dropped
- Implement reconnection logic in your transport
