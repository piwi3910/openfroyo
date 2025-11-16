# Dell OS10 Switch Management Module

## Overview

The **os10** module provides comprehensive management capabilities for Dell OS10 network switches. It supports both REST API and SSH CLI methods for configuration, offering 36+ commands across VLAN management, interface configuration, port channels, Layer 3 networking, spanning tree, system settings, user management, and ACL configuration.

## Features

- **Dual Transport Support**: REST API (preferred) with automatic CLI fallback
- **Comprehensive Operations**: 36 commands covering all major switch functions
- **Flexible Authentication**: Basic authentication for REST, password-based SSH
- **Production Ready**: Full error handling and validation
- **Standards Compliant**: Follows Dell OS10 API and CLI conventions

## Supported Operations

### VLAN Management (5 commands)
- `create_vlan` - Create a new VLAN
- `delete_vlan` - Delete an existing VLAN
- `configure_vlan` - Configure VLAN properties
- `show_vlan` - Display VLAN information
- `show_vlan_brief` - Display brief VLAN summary

### Interface Management (8 commands)
- `configure_interface` - Configure interface properties
- `shutdown_interface` - Disable an interface
- `no_shutdown_interface` - Enable an interface
- `set_description` - Set interface description
- `set_mtu` - Configure interface MTU
- `set_speed` - Set interface speed
- `show_interface` - Display interface details
- `show_interface_status` - Show interface status summary

### Port Channel/LAG (4 commands)
- `create_port_channel` - Create a port channel
- `delete_port_channel` - Delete a port channel
- `add_interface_to_lag` - Add interface to LAG
- `remove_interface_from_lag` - Remove interface from LAG

### Layer 3 Configuration (5 commands)
- `configure_ip_interface` - Configure IP address on interface
- `create_vlan_interface` - Create VLAN interface with IP
- `configure_static_route` - Add static route
- `show_ip_interface` - Display IP interface information
- `show_ip_route` - Show IP routing table

### Spanning Tree (3 commands)
- `configure_spanning_tree` - Configure STP mode
- `set_stp_priority` - Set STP priority
- `show_spanning_tree` - Display STP information

### System Configuration (5 commands)
- `set_hostname` - Configure switch hostname
- `configure_ntp` - Configure NTP server
- `configure_dns` - Configure DNS server
- `save_config` - Save running configuration
- `show_running_config` - Display running configuration

### User Management (3 commands)
- `create_user` - Create new user account
- `delete_user` - Delete user account
- `change_password` - Change user password

### ACL Configuration (3 commands)
- `create_acl` - Create access control list
- `delete_acl` - Delete access control list
- `show_acl` - Display ACL configuration

## Connection Parameters

All commands require these connection parameters:

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `host` | Yes | - | Switch IP address or hostname |
| `username` | Yes | - | Administrator username |
| `password` | Yes | - | Administrator password |
| `use_rest` | No | `true` | Use REST API (falls back to CLI if false) |
| `validate_certs` | No | `false` | Validate SSL certificates |
| `port` | No | `443` | HTTPS port for REST API |
| `ssh_port` | No | `22` | SSH port for CLI access |

## Quick Start

### Basic VLAN Creation

```yaml
- name: Create VLAN 100
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 100
    vlan_name: DATA_VLAN
```

### Configure Interface

```yaml
- name: Configure ethernet interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/1
    description: Uplink to Core
    switchport_mode: trunk
    trunk_allowed_vlans: 100,200,300
```

### Create Layer 3 Interface

```yaml
- name: Create VLAN interface with IP
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan_interface
    vlan_id: 10
    ip_address: 192.168.10.1
    netmask: 255.255.255.0
```

## Command-Specific Parameters

### VLAN Commands

**create_vlan / configure_vlan**
- `vlan_id` (required): VLAN ID (1-4094)
- `vlan_name` (optional): VLAN name
- `vlan_description` (optional): VLAN description

**delete_vlan / show_vlan**
- `vlan_id` (required): VLAN ID

### Interface Commands

**configure_interface**
- `interface` (required): Interface name (e.g., "ethernet 1/1/1")
- `description` (optional): Interface description
- `mtu` (optional): MTU size
- `speed` (optional): Interface speed (auto, 1000, 10000, etc.)
- `shutdown` (optional): Shutdown state (true/false)
- `switchport_mode` (optional): access or trunk
- `access_vlan` (optional): Access VLAN (for access mode)
- `trunk_allowed_vlans` (optional): Allowed VLANs (for trunk mode)

**shutdown_interface / no_shutdown_interface**
- `interface` (required): Interface name

**set_description / set_mtu / set_speed**
- `interface` (required): Interface name
- `description` / `mtu` / `speed` (required): Value to set

### Port Channel Commands

**create_port_channel**
- `port_channel_id` (required): Port channel number
- `lag_mode` (optional): LAG mode (active, passive, on)

**delete_port_channel**
- `port_channel_id` (required): Port channel number

**add_interface_to_lag / remove_interface_from_lag**
- `interface` (required): Interface name
- `port_channel_id` (required): Port channel number
- `lag_mode` (optional): LAG mode (for add operation)

### Layer 3 Commands

**configure_ip_interface**
- `interface` (required): Interface name
- `ip_address` (required): IP address
- `netmask` (required): Subnet mask

**create_vlan_interface**
- `vlan_id` (required): VLAN ID
- `ip_address` (required): IP address
- `netmask` (required): Subnet mask

**configure_static_route**
- `destination` (required): Destination network
- `netmask` (required): Network mask
- `next_hop` (required): Next hop IP address

### Spanning Tree Commands

**configure_spanning_tree**
- `stp_mode` (optional): STP mode (rstp, pvst, mstp)

**set_stp_priority**
- `stp_priority` (required): Priority value (0-61440, multiples of 4096)
- `vlan_id` (optional): VLAN ID for per-VLAN priority

### System Commands

**set_hostname**
- `hostname` (required): New hostname

**configure_ntp**
- `ntp_server` (required): NTP server IP or hostname

**configure_dns**
- `dns_server` (required): DNS server IP

### User Management Commands

**create_user**
- `user` (required): Username
- `new_password` (required): Password
- `role` (optional): User role (sysadmin, netadmin, etc.)

**delete_user**
- `user` (required): Username to delete

**change_password**
- `user` (required): Username
- `new_password` (required): New password

### ACL Commands

**create_acl**
- `acl_name` (required): ACL name
- `acl_type` (optional): standard or extended
- `sequence` (optional): Sequence number
- `action` (optional): permit or deny
- `protocol` (optional): Protocol (ip, tcp, udp, etc.)
- `source` (optional): Source address/network
- `destination_acl` (optional): Destination address/network

**delete_acl**
- `acl_name` (required): ACL name
- `acl_type` (optional): standard or extended

**show_acl**
- `acl_name` (optional): ACL name (shows all if omitted)

## Transport Methods

### REST API (Preferred)

When `use_rest: true` (default), the module uses Dell OS10 REST API:

```yaml
vars:
  use_rest: true
  port: 443
  validate_certs: false
```

**Advantages:**
- Structured JSON responses
- Better error handling
- Programmatic access
- No terminal emulation needed

**Requirements:**
- REST API enabled on switch
- HTTPS access (port 443)
- Valid SSL certificate (or validate_certs: false)

### SSH CLI (Fallback)

When `use_rest: false`, the module uses SSH CLI:

```yaml
vars:
  use_rest: false
  ssh_port: 22
```

**Advantages:**
- Works on all OS10 switches
- No REST API configuration needed
- Familiar CLI syntax

**Requirements:**
- SSH access enabled
- `sshpass` utility installed on host
- SSH key or password authentication

## Security Best Practices

1. **Credential Management**
   - Store credentials in vault or secure variables
   - Never commit passwords to version control
   - Use dedicated automation accounts with minimum privileges

2. **SSL Certificate Validation**
   - Set `validate_certs: true` in production
   - Install proper SSL certificates on switches
   - Use certificate pinning for critical infrastructure

3. **Network Security**
   - Restrict management access to secure networks
   - Use VPN or jump hosts for remote management
   - Enable audit logging on switches

4. **Access Control**
   - Create dedicated user accounts for automation
   - Use role-based access control (RBAC)
   - Regularly audit user accounts and permissions

5. **Configuration Backup**
   - Always run `save_config` after changes
   - Maintain configuration backups
   - Test changes in non-production environment first

## Error Handling

The module provides detailed error messages for common issues:

- **Connection Errors**: Invalid host, port, or credentials
- **Validation Errors**: Missing required parameters
- **Command Errors**: Invalid VLAN IDs, interface names, etc.
- **API Errors**: REST API failures with HTTP status codes

Example error output:
```json
{
  "status": "failed",
  "message": "Missing required variable: vlan_id",
  "facts": {}
}
```

## Dell OS10 API Reference

### REST API Endpoints

- **VLANs**: `/restconf/data/dell-vlan:vlans`
- **Interfaces**: `/restconf/data/ietf-interfaces:interfaces`
- **System**: `/restconf/data/dell-system:system`
- **Running Config**: `/restconf/data/dell-system:running-config`

### Authentication

REST API uses HTTP Basic Authentication:
```
Authorization: Basic base64(username:password)
```

### Content Types

- **Request**: `application/json`
- **Response**: `application/json`

## Troubleshooting

### REST API Not Working

1. **Verify REST API is enabled:**
   ```
   show rest-server
   ```

2. **Enable REST API:**
   ```
   configure terminal
   rest-server enable
   ```

3. **Check certificate:**
   - Use `validate_certs: false` for testing
   - Install valid certificate for production

### SSH Connection Failures

1. **Verify SSH access:**
   ```bash
   ssh admin@192.168.1.10
   ```

2. **Check sshpass installation:**
   ```bash
   which sshpass
   ```

3. **Install sshpass if needed:**
   ```bash
   # macOS
   brew install hudochenkov/sshpass/sshpass

   # Ubuntu/Debian
   apt-get install sshpass

   # RHEL/CentOS
   yum install sshpass
   ```

### Command Execution Issues

1. **Enable debugging:**
   - Check module output for detailed error messages
   - Verify switch CLI syntax matches OS10 version

2. **Verify permissions:**
   - Ensure user has appropriate role (sysadmin)
   - Check command authorization settings

3. **Check syntax:**
   - Refer to Dell OS10 documentation
   - Test commands manually via SSH

## Examples

See `test.ofy` for comprehensive examples covering:
- VLAN creation and management
- Interface configuration
- Port channel setup
- Layer 3 routing
- System configuration
- User management
- ACL configuration

## Building from Source

```bash
cd modules/networking/os10/wasm
make build
```

Requirements:
- TinyGo compiler
- WASI target support

## Version History

- **1.0.0** - Initial release with 36 commands
  - VLAN management
  - Interface configuration
  - Port channel/LAG support
  - Layer 3 routing
  - Spanning tree
  - System configuration
  - User management
  - ACL configuration

## License

Part of the OpenFroyo project.

## Support

For issues and questions:
- Review the command reference in `COMMANDS.md`
- Check quick reference in `QUICK_REFERENCE.md`
- Consult Dell OS10 documentation
- Review test examples in `test.ofy`
