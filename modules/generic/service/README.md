# service Module

Manage system services across platforms (Linux systemd, Windows services). Essential for starting, stopping, enabling, and managing services like the froyo-agent.

## Features

- Start, stop, restart, and reload services
- Enable/disable services for automatic startup
- Cross-platform support (Linux systemd, Windows services)
- Idempotent operations (safe to run multiple times)
- Daemon reload support for systemd
- Service status verification
- Configurable wait times after state changes

## Supported Platforms

### Linux
- **Service Manager**: systemd
- **Commands Used**: `systemctl`
- **Requirements**: systemd installed (most modern Linux distributions)

### Windows
- **Service Manager**: Windows Service Control Manager
- **Commands Used**: `sc.exe`
- **Requirements**: Windows Vista or later

## Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `name` | Yes | - | Name of the service to manage |
| `state` | No | `started` | Desired state: `started`, `stopped`, `restarted`, `reloaded` |
| `enabled` | No | - | Enable (`true`) or disable (`false`) service startup on boot |
| `daemon_reload` | No | `false` | Run systemd daemon-reload before managing service (Linux only) |
| `sleep` | No | `0` | Seconds to sleep after state change (useful for service startup) |

## Service States

### started
- Starts the service if it's not running
- Idempotent: Does nothing if service is already running

### stopped
- Stops the service if it's running
- Idempotent: Does nothing if service is already stopped

### restarted
- Always restarts the service (stop + start)
- Non-idempotent: Always performs the restart

### reloaded
- Reloads service configuration without stopping
- Linux: Uses `systemctl reload`
- Windows: Performs restart (Windows doesn't support reload)

## Examples

### Basic Service Start

```yaml
- name: Start nginx
  module: generic/service
  vars:
    name: nginx
    state: started
```

### Enable Service on Boot

```yaml
- name: Enable and start sshd
  module: generic/service
  vars:
    name: sshd
    state: started
    enabled: true
```

### Restart Service

```yaml
- name: Restart nginx
  module: generic/service
  vars:
    name: nginx
    state: restarted
```

### Stop and Disable Service

```yaml
- name: Stop and disable service
  module: generic/service
  vars:
    name: nginx
    state: stopped
    enabled: false
```

### Reload Service Configuration

```yaml
- name: Reload nginx config
  module: generic/service
  vars:
    name: nginx
    state: reloaded
```

### Systemd Daemon Reload

```yaml
- name: Reload systemd and start service
  module: generic/service
  vars:
    name: my-custom-service
    state: started
    daemon_reload: true
    enabled: true
```

### Service with Startup Delay

```yaml
- name: Start database with delay
  module: generic/service
  vars:
    name: postgresql
    state: started
    sleep: 5  # Wait 5 seconds after starting
```

## Real-World Use Cases

### Deploy OpenFroyo Agent

```yaml
# After installing froyo-agent binary and systemd unit file
- name: Enable and start froyo-agent
  module: generic/service
  vars:
    name: froyo-agent
    state: started
    enabled: true
    daemon_reload: true
```

### Web Server Management

```yaml
# Update nginx configuration and reload
- name: Copy nginx config
  module: generic/copy
  vars:
    src: files/nginx.conf
    dest: /etc/nginx/nginx.conf

- name: Reload nginx
  module: generic/service
  vars:
    name: nginx
    state: reloaded
```

### Application Deployment

```yaml
# Deploy new version and restart
- name: Stop application
  module: generic/service
  vars:
    name: myapp
    state: stopped

- name: Update application binary
  module: generic/get_url
  vars:
    url: https://releases.example.com/myapp-v2.0
    dest: /usr/local/bin/myapp
    mode: '0755'

- name: Start application
  module: generic/service
  vars:
    name: myapp
    state: started
    sleep: 3
```

### Database Server Setup

```yaml
- name: Enable PostgreSQL
  module: generic/service
  vars:
    name: postgresql
    state: started
    enabled: true

- name: Wait for PostgreSQL to be ready
  module: generic/service
  vars:
    name: postgresql
    state: started
    sleep: 5
```

### Windows Service Management

```yaml
# Manage Windows services (same syntax!)
- name: Start IIS service
  module: generic/service
  vars:
    name: W3SVC
    state: started
    enabled: true
```

## Idempotency

The module is idempotent for most operations:

- **started**: Only starts if service is not running
- **stopped**: Only stops if service is running
- **enabled/disabled**: Only changes startup configuration if needed
- **restarted**: Always restarts (not idempotent)
- **reloaded**: Always reloads (not idempotent)

## Platform-Specific Behavior

### Linux (systemd)

```bash
# Check if service is active
systemctl is-active nginx

# Start service
systemctl start nginx

# Enable service
systemctl enable nginx

# Daemon reload (after unit file changes)
systemctl daemon-reload

# Status check
systemctl status nginx
```

### Windows

```bash
# Check service state
sc.exe query "nginx"

# Start service
sc.exe start "nginx"

# Set to automatic startup
sc.exe config "nginx" start= auto

# Set to manual startup (disabled)
sc.exe config "nginx" start= demand
```

## Error Handling

The module will fail if:
- Service name doesn't exist
- Insufficient permissions to manage service
- Invalid state specified
- Service manager not available (systemctl or sc.exe)

## Return Values

```json
{
  "status": "changed|ok|failed",
  "message": "Service management status",
  "facts": {
    "shell_exec": [...],
    "service_name": "nginx",
    "service_state": "started"
  }
}
```

## Common Service Names

### Linux (systemd)
- `nginx` - Nginx web server
- `apache2` / `httpd` - Apache web server
- `sshd` - SSH daemon
- `postgresql` - PostgreSQL database
- `mysql` / `mariadb` - MySQL/MariaDB database
- `docker` - Docker daemon
- `redis` - Redis server
- `froyo-agent` - OpenFroyo agent (custom service)

### Windows
- `W3SVC` - IIS World Wide Web Publishing Service
- `MSSQLSERVER` - Microsoft SQL Server
- `Spooler` - Print Spooler
- `WinRM` - Windows Remote Management
- `froyo-agent` - OpenFroyo agent (custom service)

## Best Practices

1. **Always enable critical services**
   ```yaml
   enabled: true
   ```

2. **Use daemon_reload after systemd unit changes**
   ```yaml
   daemon_reload: true
   ```

3. **Add delays for database services**
   ```yaml
   sleep: 5
   ```

4. **Reload instead of restart when possible**
   ```yaml
   state: reloaded  # Avoids downtime
   ```

5. **Check service status in orchestration**
   - The module automatically verifies service status after operations

## Security Considerations

- Requires elevated privileges (root/sudo on Linux, Administrator on Windows)
- Service management can affect system security
- Always validate service names to prevent unintended operations
- Be cautious with `enabled: false` on critical services like sshd

## Notes

- On Linux, requires systemd (most modern distributions)
- On Windows, requires sc.exe (built-in since Vista)
- Cross-platform stacks work seamlessly (module detects OS automatically)
- The `reloaded` state on Windows performs a restart (Windows limitation)

## See Also

- [get_url module](../get_url/) - Download service binaries and configurations
- [copy module](../copy/) - Copy service configuration files
- [template module](../template/) - Generate service unit files from templates
- [systemd module](../../linux/systemd/) - Advanced systemd unit management
