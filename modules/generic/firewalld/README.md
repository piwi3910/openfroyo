# Firewalld Module

Manage firewalld firewall services and ports on Linux systems using firewalld.

## Requirements

- firewalld must be installed and running
- Root or sudo access (firewalld commands require elevated privileges)

## Variables

- `service` (optional): Service name to manage (http, https, ssh, etc.)
- `port` (optional): Port/protocol combination (e.g., "8080/tcp", "53/udp")
- `state` (optional): Desired state - `enabled` (default) or `disabled`
- `zone` (optional): Firewall zone (default: "public")
- `permanent` (optional): Make changes permanent (boolean, default: true)
- `immediate` (optional): Apply changes immediately (boolean, default: true)

**Note**: Either `service` or `port` must be specified.

## States

- **enabled**: Enable the service or port in the specified zone
- **disabled**: Disable the service or port in the specified zone

## Examples

### Enable HTTP service

```yaml
- name: Enable HTTP
  module: generic/firewalld
  hosts:
    - web-server
  vars:
    service: http
    state: enabled
    zone: public
```

### Enable custom port

```yaml
- name: Enable custom application port
  module: generic/firewalld
  hosts:
    - app-server
  vars:
    port: 8080/tcp
    state: enabled
    zone: public
```

### Disable SSH service

```yaml
- name: Disable SSH
  module: generic/firewalld
  hosts:
    - secured-server
  vars:
    service: ssh
    state: disabled
    zone: public
```

### Enable HTTPS with immediate apply only

```yaml
- name: Enable HTTPS temporarily
  module: generic/firewalld
  hosts:
    - web-server
  vars:
    service: https
    state: enabled
    permanent: false
    immediate: true
```

### Enable multiple ports in DMZ zone

```yaml
- name: Enable app port in DMZ
  module: generic/firewalld
  hosts:
    - dmz-server
  vars:
    port: 3000/tcp
    state: enabled
    zone: dmz
    permanent: true
    immediate: true
```

## How It Works

1. **Verification**: Checks if firewalld is installed and running
2. **State Check**: Queries current state of service or port
3. **Configuration**: Adds or removes service/port based on desired state
4. **Reload**: Reloads firewalld if both permanent and immediate flags are set
5. **Result**: Returns success/failure status and output

## Common Services

- `http` - HTTP web traffic (port 80)
- `https` - HTTPS encrypted web traffic (port 443)
- `ssh` - SSH remote access (port 22)
- `ftp` - FTP file transfer (port 21)
- `smtp` - SMTP mail transfer (port 25)
- `dns` - DNS queries (port 53)
- `mysql` - MySQL database (port 3306)
- `postgresql` - PostgreSQL database (port 5432)

## Firewalld Zones

- `public` - Default zone for public networks
- `dmz` - Demilitarized zone for public-facing servers
- `internal` - Internal private network
- `trusted` - Fully trusted network
- `home` - Home network
- `work` - Work network
- `external` - External network with masquerading
- `block` - All incoming connections rejected
- `drop` - All incoming connections dropped

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/firewalld.wasm`.
