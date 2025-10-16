# Micro-Runner Protocol Specification

## Overview

The OpenFroyo micro-runner is a lightweight, ephemeral binary that executes commands on remote hosts via JSON-over-stdio communication. It is designed to be agentless, self-deleting, and secure.

## Key Characteristics

- **Static Binary**: Built with `CGO_ENABLED=0`, no external dependencies
- **Small Size**: < 10 MB (currently ~2.3 MB)
- **Self-Deleting**: Removes itself on exit
- **TTL Limited**: Automatically exits after 10 minutes
- **Secure**: Supports command signature verification
- **Cross-Platform**: Builds for Linux, macOS, Windows

## Communication Protocol

The runner uses a frame-based protocol over stdin/stdout with JSON-encoded messages, one per line.

### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `READY` | Runner → Controller | Runner is ready to receive commands |
| `CMD` | Controller → Runner | Command to execute |
| `EVENT` | Runner → Controller | Progress event during execution |
| `DONE` | Runner → Controller | Command completed successfully |
| `ERROR` | Runner → Controller | Error occurred |
| `EXIT` | Runner → Controller | Runner is exiting |

### Message Structure

All messages follow this base structure:

```json
{
  "type": "MESSAGE_TYPE",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": { ... }
}
```

## Message Details

### READY Message

Sent by the runner immediately after startup.

```json
{
  "type": "READY",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "version": "1.0.0",
    "platform": "linux",
    "arch": "amd64",
    "pid": 12345,
    "capabilities": {
      "exec": true,
      "file.write": true,
      "file.read": true,
      "pkg.ensure": true,
      "service.reload": true,
      "sudoers.ensure": true,
      "sshd.harden": true
    },
    "metadata": {
      "ttl": "10m0s"
    }
  }
}
```

### CMD (Command) Message

Sent by the controller to execute a command.

```json
{
  "type": "CMD",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "id": "cmd-abc-123",
    "type": "exec",
    "idempotency_key": "optional-key",
    "timeout": 300,
    "params": { ... },
    "signature": "optional-signature",
    "metadata": {}
  }
}
```

**Fields:**
- `id`: Unique command identifier
- `type`: Command type (see Command Types below)
- `idempotency_key`: Optional key for idempotent operations
- `timeout`: Command timeout in seconds
- `params`: Command-specific parameters (JSON object)
- `signature`: Optional cryptographic signature
- `metadata`: Optional metadata

### EVENT Message

Sent by the runner to report progress during command execution.

```json
{
  "type": "EVENT",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "command_id": "cmd-abc-123",
    "level": "info",
    "message": "Processing file 5 of 10",
    "progress": {
      "current": 5,
      "total": 10,
      "unit": "files"
    },
    "metadata": {}
  }
}
```

**Levels:** `info`, `warn`, `debug`

### DONE Message

Sent by the runner when a command completes successfully.

```json
{
  "type": "DONE",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "command_id": "cmd-abc-123",
    "result": { ... },
    "duration": 1.234,
    "metadata": {}
  }
}
```

### ERROR Message

Sent by the runner when an error occurs.

```json
{
  "type": "ERROR",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "command_id": "cmd-abc-123",
    "code": "EXEC_FAILED",
    "message": "Command execution failed: ...",
    "details": {},
    "retryable": false,
    "retry_after": 0
  }
}
```

**Error Codes:**
- `INIT_FAILED`: Runner initialization failed
- `READY_FAILED`: Failed to send READY message
- `EXEC_FAILED`: Command execution failed
- `TIMEOUT`: Operation timed out
- `INVALID_CMD`: Invalid command format

### EXIT Message

Sent by the runner before terminating.

```json
{
  "type": "EXIT",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": {
    "reason": "completed",
    "exit_code": 0,
    "self_deleted": true,
    "commands_total": 5
  }
}
```

**Exit Reasons:**
- `completed`: Normal completion
- `ttl_expired`: TTL timeout reached
- `stdin_closed`: Controller closed stdin
- `error`: Fatal error occurred

## Command Types

### 1. exec - Execute Shell Command

Execute a shell command on the remote host.

**Parameters:**
```json
{
  "command": "ls",
  "args": ["-la", "/tmp"],
  "work_dir": "/tmp",
  "env": {
    "VAR": "value"
  },
  "shell": "/bin/sh",
  "capture_out": true,
  "capture_err": true,
  "stream_lines": false
}
```

**Result:**
```json
{
  "exit_code": 0,
  "stdout": "...",
  "stderr": "...",
  "duration": 0.123
}
```

### 2. file.write - Write File

Write content to a file on the remote host.

**Parameters:**
```json
{
  "path": "/etc/myapp/config.yaml",
  "content": "key: value\n",
  "mode": "0644",
  "owner": "root",
  "group": "root",
  "backup": true,
  "create": true
}
```

**Result:**
```json
{
  "bytes_written": 123,
  "created": false,
  "backup_path": "/etc/myapp/config.yaml.bak",
  "checksum": "sha256:..."
}
```

### 3. file.read - Read File

Read content from a file on the remote host.

**Parameters:**
```json
{
  "path": "/etc/myapp/config.yaml",
  "max_bytes": 10485760
}
```

**Result:**
```json
{
  "content": "...",
  "size": 123,
  "mode": "0644",
  "owner": "1000",
  "group": "1000",
  "checksum": "sha256:...",
  "truncated": false
}
```

### 4. pkg.ensure - Package Management

Ensure a package is installed, removed, or up to date.

**Parameters:**
```json
{
  "name": "nginx",
  "version": "1.18.0",
  "state": "present",
  "manager": "apt",
  "options": ["--no-install-recommends"]
}
```

**States:** `present`, `absent`, `latest`
**Managers:** `apt`, `dnf`, `yum`, `zypper` (auto-detected if not specified)

**Result:**
```json
{
  "changed": true,
  "previous_version": "",
  "installed_version": "1.18.0-0ubuntu1",
  "action": "installed"
}
```

**Actions:** `installed`, `removed`, `upgraded`, `already_present`, `already_absent`

### 5. service.reload - Service Management

Manage systemd services.

**Parameters:**
```json
{
  "name": "nginx",
  "action": "reload"
}
```

**Actions:** `reload`, `restart`, `start`, `stop`, `enable`, `disable`

**Result:**
```json
{
  "changed": true,
  "action": "reloaded",
  "status": "active",
  "enabled": true,
  "sub_state": "running"
}
```

### 6. sudoers.ensure - Sudoers Management

Manage sudoers rules for a user.

**Parameters:**
```json
{
  "user": "froyo",
  "commands": [
    "/usr/bin/systemctl",
    "/usr/bin/apt"
  ],
  "no_passwd": true,
  "state": "present"
}
```

**Result:**
```json
{
  "changed": true,
  "file_path": "/etc/sudoers.d/froyo-user",
  "action": "created"
}
```

### 7. sshd.harden - SSH Hardening

Apply SSH hardening configuration.

**Parameters:**
```json
{
  "disable_password_auth": true,
  "disable_root_login": true,
  "allow_users": ["froyo", "admin"],
  "port": 22,
  "test_connection": true
}
```

**Result:**
```json
{
  "changed": true,
  "backup_path": "/etc/ssh/sshd_config.bak",
  "modified_keys": [
    "PasswordAuthentication",
    "PermitRootLogin",
    "AllowUsers"
  ],
  "service_action": "reloaded"
}
```

## Example Flow

Here's a complete example of a runner session:

### 1. Runner Startup

```bash
$ ./micro-runner
{"type":"READY","timestamp":"2024-01-01T10:00:00Z","data":{"version":"1.0.0","platform":"linux","arch":"amd64","pid":12345,"capabilities":{"exec":true},"metadata":{"ttl":"10m0s"}}}
```

### 2. Controller Sends Command

```json
{"type":"CMD","timestamp":"2024-01-01T10:00:01Z","data":{"id":"cmd-1","type":"exec","timeout":30,"params":{"command":"uptime","capture_out":true,"capture_err":true}}}
```

### 3. Runner Sends Event (Optional)

```json
{"type":"EVENT","timestamp":"2024-01-01T10:00:01Z","data":{"command_id":"cmd-1","level":"info","message":"Executing command"}}
```

### 4. Runner Sends Result

```json
{"type":"DONE","timestamp":"2024-01-01T10:00:01Z","data":{"command_id":"cmd-1","result":{"exit_code":0,"stdout":"10:00:01 up 5 days, 3:21, 1 user, load average: 0.00, 0.01, 0.05\n","stderr":"","duration":0.012},"duration":0.015}}
```

### 5. Controller Closes stdin

```bash
^D
```

### 6. Runner Exits

```json
{"type":"EXIT","timestamp":"2024-01-01T10:00:02Z","data":{"reason":"stdin_closed","exit_code":0,"self_deleted":true,"commands_total":1}}
```

## Building the Runner

```bash
# For Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o micro-runner-linux-amd64 ./cmd/micro-runner

# For Linux arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o micro-runner-linux-arm64 ./cmd/micro-runner

# For Darwin (macOS) arm64
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o micro-runner-darwin-arm64 ./cmd/micro-runner
```

## Security Considerations

1. **Signature Verification**: Commands can include a cryptographic signature that the runner verifies before execution
2. **Timeout Enforcement**: All commands have mandatory timeouts to prevent runaway processes
3. **TTL**: Runner automatically exits after 10 minutes to prevent long-running processes
4. **Self-Delete**: Runner removes itself on exit to avoid leaving artifacts
5. **Capability Checks**: Runner advertises its capabilities; controller should respect them
6. **Limited Scope**: Runner should be executed with the minimum required privileges

## Error Handling

All errors follow a consistent structure with:
- **Code**: Machine-readable error code
- **Message**: Human-readable description
- **Retryable**: Whether the operation can be retried
- **RetryAfter**: Suggested retry delay in seconds

## Testing

Run the protocol tests:

```bash
go test ./pkg/micro_runner/protocol/... -v
```

## Implementation Notes

- Messages are newline-delimited JSON (one message per line)
- Timestamps are in RFC3339 format with UTC timezone
- Binary size is kept minimal through static linking and stripped symbols
- Runner uses buffered I/O for efficient message handling
- All operations are context-aware for cancellation and timeout
