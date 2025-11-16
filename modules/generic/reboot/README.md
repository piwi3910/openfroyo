# reboot Module

Reboot, shutdown, or halt the system with optional delays and messages.

## Purpose

The reboot module manages system power states including rebooting, shutting down, halting, or cancelling scheduled power actions. It supports both immediate and delayed operations with optional user messages.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| state | string | No | reboot | Power action: "reboot", "shutdown", "halt", or "cancel" |
| delay | int | No | 0 | Delay in minutes before executing the action |
| message | string | No | "System is going down for maintenance" | Message to display to logged-in users |
| force | bool | No | false | Force immediate action without graceful shutdown |

## State Options

- **reboot**: Restart the system
- **shutdown**: Power off the system
- **halt**: Stop the CPU without powering off
- **cancel**: Cancel a previously scheduled shutdown/reboot

## Examples

### Immediate reboot
```yaml
- task: Reboot server now
  module: reboot
  vars:
    state: reboot
```

### Scheduled reboot with message
```yaml
- task: Schedule reboot in 10 minutes
  module: reboot
  vars:
    state: reboot
    delay: 10
    message: "System will reboot for updates"
```

### Graceful shutdown
```yaml
- task: Shutdown server
  module: reboot
  vars:
    state: shutdown
    delay: 5
    message: "System maintenance - shutting down"
```

### Force immediate reboot
```yaml
- task: Force reboot now
  module: reboot
  vars:
    state: reboot
    force: true
```

### Cancel scheduled action
```yaml
- task: Cancel scheduled reboot
  module: reboot
  vars:
    state: cancel
```

### System halt
```yaml
- task: Halt system
  module: reboot
  vars:
    state: halt
```

## Output Facts

The module returns the following facts:

- `exit_code`: Exit code of the shutdown/reboot command
- `output`: Raw output from the command
- `reboot_state`: The state that was executed
- `reboot_delay`: The delay in minutes
- `force`: Whether force mode was used

## Implementation Details

This module uses the `shell_exec` pattern to execute shutdown/reboot commands on remote hosts.

### Commands Used

**Reboot:**
- Graceful: `shutdown -r now` or `shutdown -r +{delay}`
- Force: `reboot -f`

**Shutdown:**
- Graceful: `shutdown -h now` or `shutdown -h +{delay}`
- Force: `poweroff -f`

**Halt:**
- `halt`

**Cancel:**
- `shutdown -c`

### Message Support

When a message is provided with delayed actions:
```bash
shutdown -r +10 "System will reboot for updates"
```

Logged-in users will see this message as a wall message.

## Platform Compatibility

- **Linux**: Full support (all distributions)
- **macOS**: Partial support (shutdown and reboot, no -f flag)
- **BSD**: Full support

Note: Force reboot/shutdown may cause data loss. Use with caution.

## Return Status

- **changed**: System power state will change (for reboot, shutdown, halt)
- **ok**: For cancel operations (no state change)
- **failed**:
  - Invalid state value
  - Negative delay value
  - Insufficient permissions to execute shutdown commands

## Security Considerations

**IMPORTANT:** This module requires root/sudo privileges to execute shutdown and reboot commands. Ensure your SSH connection has appropriate permissions.

Common approaches:
1. SSH as root user
2. Use sudo with NOPASSWD for shutdown commands
3. Configure sudoers to allow specific users to execute shutdown

### Example sudoers entry:
```
deployer ALL=(ALL) NOPASSWD: /sbin/shutdown, /sbin/reboot, /sbin/halt, /sbin/poweroff
```

## Warning

Use this module with caution in production environments:

- **Immediate reboots** will terminate all processes immediately
- **Force mode** bypasses graceful shutdown and may cause data corruption
- **Scheduled actions** can be cancelled, but only if you act before the delay expires
- Always ensure critical services are properly stopped before rebooting

## Connection Handling

When executing immediate reboots or shutdowns, the SSH connection will likely be terminated before the command completes. This is expected behavior. The OpenFroyo runner should handle connection drops gracefully.
