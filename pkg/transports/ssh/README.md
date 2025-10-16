# SSH Transport Layer

The SSH transport layer provides secure remote operations for OpenFroyo infrastructure orchestration engine.

## Features

- **Connection Management**: Efficient connection pooling with automatic reconnection
- **Multiple Authentication Methods**: Password, private key, and SSH agent support
- **Command Execution**: Execute commands with sudo support and timeout handling
- **File Transfer**: SFTP-based upload/download with checksum verification
- **Interactive Sessions**: Support for micro-runner and interactive operations
- **Proxy/Jump Host**: Connect through bastion/jump hosts
- **Keep-Alive**: Automatic connection keep-alive mechanism
- **Context Support**: All operations support context cancellation

## Architecture

### Components

1. **transport.go**: Defines the `Transport` interface for all remote operations
2. **config.go**: Configuration management with validation and defaults
3. **ssh_client.go**: Core SSH client with connection pooling
4. **executor.go**: Command execution with sudo and batch support
5. **file_transfer.go**: SFTP-based file operations

### Connection Lifecycle

```
┌─────────────┐
│   Create    │
│   Config    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    New      │
│  SSHClient  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Connect   │──────┐
└──────┬──────┘      │
       │             │ Auto-reconnect
       │             │ on failure
       ▼             │
┌─────────────┐      │
│  Operations │◄─────┘
│ (exec/sftp) │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Disconnect  │
└─────────────┘
```

## Usage Examples

### Basic Connection

```go
import (
    "context"
    "github.com/openfroyo/openfroyo/pkg/transports/ssh"
)

// Create configuration
config := ssh.DefaultConfig("example.com", "user")
config.AuthMethod = ssh.AuthMethodPassword
config.Password = "secret"

// Create client
client, err := ssh.NewSSHClient(config)
if err != nil {
    log.Fatal(err)
}

// Connect
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

// Check connection health
if err := client.HealthCheck(ctx); err != nil {
    log.Fatal(err)
}
```

### Key-Based Authentication

```go
config := ssh.DefaultConfig("example.com", "user")
config.AuthMethod = ssh.AuthMethodKey
config.PrivateKeyPath = "/home/user/.ssh/id_ed25519"
config.PrivateKeyPassphrase = "keypassphrase" // Optional

client, err := ssh.NewSSHClient(config)
// ... connect and use
```

### Execute Commands

```go
// Simple command
stdout, stderr, err := client.ExecuteCommand(ctx, "ls -la /tmp")
if err != nil {
    log.Printf("Command failed: %v", err)
}
fmt.Println(stdout)

// Command with sudo
stdout, stderr, err = client.ExecuteCommandWithSudo(ctx, "systemctl restart nginx", "")
if err != nil {
    log.Printf("Sudo command failed: %v", err)
}

// Batch execution
results, err := client.executor.ExecuteBatch(ctx, []string{
    "echo 'step 1'",
    "echo 'step 2'",
    "echo 'step 3'",
}, true, false, "")
```

### File Operations

```go
// Upload file
err := client.UploadFile(ctx, "/local/path/file.txt", "/remote/path/file.txt", 0644)
if err != nil {
    log.Fatal(err)
}

// Download file
err = client.DownloadFile(ctx, "/remote/path/file.txt", "/local/path/file.txt")
if err != nil {
    log.Fatal(err)
}

// Upload directory
err = client.UploadDirectory(ctx, "/local/dir", "/remote/dir")
if err != nil {
    log.Fatal(err)
}

// Set permissions
err = client.SetFilePermissions(ctx, "/remote/path/file.txt", 0755)

// Set ownership (requires sudo)
err = client.SetFileOwnership(ctx, "/remote/path/file.txt", 1000, 1000)

// Verify checksum
checksum, err := client.ComputeChecksum(ctx, "/remote/path/file.txt")
```

### Interactive Session (for Micro-Runner)

```go
stdin, stdout, stderr, cleanup, err := client.StartInteractiveSession(ctx)
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// Write JSON command
cmd := `{"command": "exec", "args": ["ls", "-la"]}`
stdin.Write([]byte(cmd + "\n"))

// Read response
buf := make([]byte, 4096)
n, _ := stdout.Read(buf)
fmt.Println(string(buf[:n]))
```

### Proxy/Jump Host

```go
config := ssh.DefaultConfig("target.internal.com", "user")
config.AuthMethod = ssh.AuthMethodKey
config.PrivateKeyPath = "/home/user/.ssh/id_rsa"

// Configure proxy/jump host
config.ProxyHost = "bastion.example.com"
config.ProxyPort = 22
config.ProxyUser = "jumpuser"
config.ProxyAuthMethod = ssh.AuthMethodKey
config.ProxyPrivateKeyPath = "/home/user/.ssh/bastion_key"

client, err := ssh.NewSSHClient(config)
// Connection will go through bastion.example.com to reach target.internal.com
```

### Connection Pooling

```go
config := ssh.DefaultConfig("example.com", "user")
config.EnableConnectionPooling = true
config.MaxPoolSize = 10
config.PoolIdleTimeout = 10 * time.Minute

client, err := ssh.NewSSHClient(config)
// Connections are pooled and reused automatically
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `Host` | Remote hostname or IP | Required |
| `Port` | SSH port | 22 |
| `User` | SSH username | Required |
| `AuthMethod` | Authentication method (password/key/agent) | key |
| `Password` | Password for auth | "" |
| `PrivateKeyPath` | Path to private key | ~/.ssh/id_* |
| `KnownHostsPath` | Path to known_hosts | ~/.ssh/known_hosts |
| `StrictHostKeyChecking` | Enable strict host key verification | true |
| `ConnectionTimeout` | Timeout for connection | 30s |
| `CommandTimeout` | Default command timeout | 5m |
| `KeepAliveInterval` | Keep-alive interval | 30s |
| `MaxKeepAliveRetries` | Max keep-alive retries | 3 |
| `EnableConnectionPooling` | Enable connection pooling | true |
| `MaxPoolSize` | Max pool size | 5 |
| `PoolIdleTimeout` | Idle connection timeout | 10m |
| `ProxyHost` | Jump host hostname | "" |
| `ProxyPort` | Jump host port | 22 |

## Error Handling

All errors implement the `TransportError` interface:

```go
type TransportError struct {
    Op          string  // Operation that failed
    Err         error   // Underlying error
    IsTemporary bool    // Can be retried
    IsAuthError bool    // Authentication related
}
```

Example error handling:

```go
_, _, err := client.ExecuteCommand(ctx, "some-command")
if err != nil {
    if transportErr, ok := err.(*ssh.TransportError); ok {
        if transportErr.IsTemporary {
            // Retry the operation
        }
        if transportErr.IsAuthError {
            // Handle authentication failure
        }
    }
}
```

## Testing

The package includes comprehensive tests with a built-in SSH test server:

```bash
go test ./pkg/transports/ssh/...
```

## Best Practices

1. **Always use context**: Pass context to all operations for proper cancellation
2. **Close connections**: Always defer `Disconnect()` after connecting
3. **Handle errors**: Check and handle `TransportError` appropriately
4. **Use connection pooling**: Enable pooling for multiple operations
5. **Verify host keys**: Use `StrictHostKeyChecking` in production
6. **Timeout operations**: Set appropriate timeouts for long-running commands
7. **Secure credentials**: Never hardcode passwords or keys in code

## Security Considerations

- Host key verification is enabled by default
- Private keys should have 0600 permissions
- Use key-based auth over passwords when possible
- Keep connections encrypted (no plaintext fallback)
- Validate checksums after file transfers
- Use sudo with NOPASSWD rules for automation
- Rotate keys regularly
- Use jump hosts for accessing internal networks

## Dependencies

- `golang.org/x/crypto/ssh`: SSH protocol implementation
- `github.com/pkg/sftp`: SFTP file transfer
- `github.com/rs/zerolog`: Structured logging

## Integration with OpenFroyo

The SSH transport is used by:
- **Onboarding workflow**: Initial host setup
- **Micro-runner**: Uploading and executing the runner
- **Facts collection**: Gathering system information
- **Provider operations**: Executing provider commands
- **State management**: Verifying remote state

## Performance Considerations

- Connection pooling reduces overhead for multiple operations
- Keep-alive prevents connection timeouts
- SFTP is more efficient than scp for multiple files
- Batch command execution reduces round-trips
- Use compression for large file transfers (future enhancement)

## Future Enhancements

- SSH agent authentication support
- Connection multiplexing (SSH ControlMaster)
- Parallel file transfers
- Progress callbacks for long operations
- Bandwidth throttling for transfers
- Port forwarding support
- X11 forwarding for GUI applications
