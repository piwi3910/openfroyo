# UFW Module

Manage UFW (Uncomplicated Firewall) rules on Linux systems, primarily Ubuntu and Debian.

## Requirements

- UFW must be installed
- Root or sudo access (UFW commands require elevated privileges)

## Variables

- `rule` (optional): Rule action (default: "allow") - "allow", "deny", "reject", "limit"
- `port` (optional): Port number or range
- `proto` (optional): Protocol - "tcp", "udp", or "any"
- `from_ip` (optional): Source IP address or network (CIDR notation)
- `to_ip` (optional): Destination IP address or network (CIDR notation)
- `direction` (optional): Direction - "in" or "out"
- `interface` (optional): Network interface name (eth0, wlan0, etc.)
- `state` (optional): UFW state (default: "enabled") - "enabled" or "disabled"
- `delete` (optional): Delete the rule (boolean, default: false)

## Rule Types

- **allow**: Allow traffic matching the rule
- **deny**: Silently deny traffic matching the rule
- **reject**: Deny and send rejection response
- **limit**: Rate-limit connections (useful for SSH protection)

## States

- **enabled**: Enable UFW firewall
- **disabled**: Disable UFW firewall

## Examples

### Enable UFW

```yaml
- name: Enable UFW
  module: generic/ufw
  hosts:
    - ubuntu-server
  vars:
    state: enabled
```

### Allow SSH

```yaml
- name: Allow SSH
  module: generic/ufw
  hosts:
    - ubuntu-server
  vars:
    rule: allow
    port: 22
    proto: tcp
```

### Allow HTTP and HTTPS

```yaml
- name: Allow HTTP
  module: generic/ufw
  hosts:
    - web-server
  vars:
    rule: allow
    port: 80
    proto: tcp

- name: Allow HTTPS
  module: generic/ufw
  hosts:
    - web-server
  vars:
    rule: allow
    port: 443
    proto: tcp
```

### Allow from specific IP

```yaml
- name: Allow from admin network
  module: generic/ufw
  hosts:
    - app-server
  vars:
    rule: allow
    from_ip: 192.168.1.0/24
```

### Allow port from specific IP

```yaml
- name: Allow MySQL from app server
  module: generic/ufw
  hosts:
    - db-server
  vars:
    rule: allow
    port: 3306
    proto: tcp
    from_ip: 10.0.1.100
```

### Limit SSH (rate limiting)

```yaml
- name: Limit SSH to prevent brute force
  module: generic/ufw
  hosts:
    - "@group:servers"
  vars:
    rule: limit
    port: 22
    proto: tcp
```

### Allow outgoing traffic on specific interface

```yaml
- name: Allow outgoing on eth1
  module: generic/ufw
  hosts:
    - gateway
  vars:
    rule: allow
    direction: out
    interface: eth1
```

### Deny traffic from specific IP

```yaml
- name: Block malicious IP
  module: generic/ufw
  hosts:
    - web-server
  vars:
    rule: deny
    from_ip: 10.0.0.50
```

### Delete a rule

```yaml
- name: Remove old firewall rule
  module: generic/ufw
  hosts:
    - server-01
  vars:
    rule: allow
    port: 8080
    proto: tcp
    delete: true
```

### Disable UFW

```yaml
- name: Disable UFW
  module: generic/ufw
  hosts:
    - test-server
  vars:
    state: disabled
```

## How It Works

1. **Verification**: Checks if UFW is installed
2. **State Management**: Enables or disables UFW based on state parameter
3. **Rule Check**: Verifies if the rule already exists (for non-delete operations)
4. **Add/Remove**: Adds, removes, or modifies rules based on parameters
5. **Result**: Returns success/failure status and output

## Common Use Cases

### Basic Server Setup

```yaml
# Default deny incoming, allow outgoing
- name: Set default policies
  module: generic/ufw
  hosts:
    - server
  vars:
    state: enabled

- name: Allow SSH
  module: generic/ufw
  hosts:
    - server
  vars:
    rule: limit
    port: 22
    proto: tcp

- name: Allow HTTP
  module: generic/ufw
  hosts:
    - server
  vars:
    rule: allow
    port: 80
    proto: tcp

- name: Allow HTTPS
  module: generic/ufw
  hosts:
    - server
  vars:
    rule: allow
    port: 443
    proto: tcp
```

### Application Server

```yaml
- name: Allow application port
  module: generic/ufw
  hosts:
    - app-server
  vars:
    rule: allow
    port: 3000
    proto: tcp
    from_ip: 10.0.1.0/24
```

### Database Server

```yaml
- name: Allow PostgreSQL from app servers
  module: generic/ufw
  hosts:
    - db-server
  vars:
    rule: allow
    port: 5432
    proto: tcp
    from_ip: 10.0.2.0/24
```

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/ufw.wasm`.
