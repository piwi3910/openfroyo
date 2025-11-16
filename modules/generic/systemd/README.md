# Systemd Module

Manage systemd services on Linux systems.

## Description

The systemd module allows you to manage systemd services including starting, stopping, restarting, enabling, disabling, masking, and reloading the systemd daemon.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `name` | string | Yes | - | Service name (e.g., "nginx", "apache2") |
| `state` | string | No | - | Desired state: "started", "stopped", "restarted", "reloaded" |
| `enabled` | bool | No | - | Enable service at boot (true/false) |
| `daemon_reload` | bool | No | false | Reload systemd daemon before operations |
| `masked` | bool | No | - | Mask/unmask service (true/false) |

## Examples

### Start and enable a service

```yaml
- task: Start nginx
  module: systemd
  vars:
    name: nginx
    state: started
    enabled: true
```

### Restart a service

```yaml
- task: Restart apache2
  module: systemd
  vars:
    name: apache2
    state: restarted
```

### Stop and disable a service

```yaml
- task: Stop mysql
  module: systemd
  vars:
    name: mysql
    state: stopped
    enabled: false
```

### Reload systemd daemon and restart service

```yaml
- task: Reload daemon and restart docker
  module: systemd
  vars:
    name: docker
    daemon_reload: true
    state: restarted
```

### Mask a service

```yaml
- task: Mask unwanted service
  module: systemd
  vars:
    name: unwanted-service
    masked: true
```

### Just enable without state change

```yaml
- task: Enable SSH at boot
  module: systemd
  vars:
    name: ssh
    enabled: true
```

## Implementation

This module uses the shell_exec pattern to generate systemd commands. It:

1. Checks if systemd is available on the system
2. Executes daemon reload if requested
3. Handles masking/unmasking operations
4. Manages service state (start/stop/restart/reload)
5. Manages boot-time enablement
6. Reports service status after operations

All operations use `sudo` for privilege escalation and include idempotency checks where applicable.

## Requirements

- systemd must be installed and running
- User must have sudo privileges for systemctl commands
- Service must exist in systemd

## Return Values

The module returns a shell_exec command that performs the requested systemd operations. The executor handles the actual execution and reports success/failure based on exit codes.
