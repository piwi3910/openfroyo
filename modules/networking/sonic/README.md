# SONiC Network Switch Management Module

Comprehensive module for managing SONiC (Software for Open Networking in the Cloud) network switches. This module provides full support for VLAN management, interface configuration, port channels, BGP routing, ACLs, static routes, and system configuration.

## Overview

SONiC is a fully open-source network operating system based on Linux that runs on switches from multiple vendors. This module provides both REST API and CLI-based management capabilities for SONiC switches.

## Features

- **35+ Operations** covering all major SONiC configuration areas
- **Dual Mode Support**: REST API (preferred) and CLI fallback
- **VLAN Management**: Create, delete, and manage VLAN memberships
- **Interface Configuration**: IP addressing, MTU, speed, admin status
- **Port Channels/LAG**: Link aggregation configuration
- **BGP Routing**: Full BGP configuration and neighbor management
- **ACL Management**: Access control list tables and rules
- **Static Routing**: Static route configuration
- **System Configuration**: Hostname, NTP, config save/show

## Supported Commands

### VLAN Management (5 commands)
- `create_vlan` - Create a new VLAN
- `delete_vlan` - Delete an existing VLAN
- `add_vlan_member` - Add interface to VLAN (tagged/untagged)
- `remove_vlan_member` - Remove interface from VLAN
- `show_vlan` - Display VLAN configuration

### Interface Management (8 commands)
- `configure_interface` - Configure multiple interface settings
- `set_interface_ip` - Set IP address on interface
- `set_interface_mtu` - Set MTU on interface
- `set_interface_speed` - Set interface speed
- `shutdown_interface` - Administratively disable interface
- `startup_interface` - Administratively enable interface
- `show_interface` - Display interface information
- `show_interface_status` - Display interface status

### Port Channel/LAG (5 commands)
- `create_portchannel` - Create port channel (LAG)
- `delete_portchannel` - Delete port channel
- `add_portchannel_member` - Add member to port channel
- `remove_portchannel_member` - Remove member from port channel
- `show_portchannel` - Display port channel configuration

### BGP Configuration (6 commands)
- `configure_bgp` - Configure BGP routing instance
- `add_bgp_neighbor` - Add BGP neighbor
- `remove_bgp_neighbor` - Remove BGP neighbor
- `configure_bgp_network` - Advertise network via BGP
- `show_bgp_summary` - Display BGP summary
- `show_bgp_neighbors` - Display BGP neighbor details

### ACL Configuration (4 commands)
- `create_acl_table` - Create ACL table
- `delete_acl_table` - Delete ACL table
- `add_acl_rule` - Add rule to ACL table
- `show_acl` - Display ACL configuration

### Route Management (3 commands)
- `add_static_route` - Add static route
- `delete_static_route` - Delete static route
- `show_ip_route` - Display routing table

### System Configuration (4 commands)
- `set_hostname` - Set switch hostname
- `configure_ntp` - Configure NTP server
- `save_config` - Save running configuration
- `show_running_config` - Display running configuration

## Connection Configuration

### Required Variables
```yaml
host: <switch_ip_address>        # SONiC switch IP address
username: <admin_username>        # Admin username
password: <admin_password>        # Admin password
```

### Optional Variables
```yaml
port: 443                         # HTTPS port for REST API (default: 443)
use_rest: true                    # Use REST API vs CLI (default: true)
validate_certs: false             # Validate SSL certificates (default: false)
```

## Usage Examples

### VLAN Management

#### Create VLAN
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: create_vlan
    vlan_id: 100
```

#### Add Interface to VLAN (Untagged)
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_vlan_member
    vlan_id: 100
    interface: Ethernet0
    tagged: false
```

#### Add Interface to VLAN (Tagged)
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_vlan_member
    vlan_id: 200
    interface: Ethernet4
    tagged: true
```

### Interface Management

#### Configure Interface IP Address
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: Ethernet8
    ip_address: 10.0.0.1
    prefix_length: 24
```

#### Set Interface MTU
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: set_interface_mtu
    interface: Ethernet12
    mtu: 9000
```

#### Startup Interface
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: startup_interface
    interface: Ethernet16
```

### Port Channel (LAG) Configuration

#### Create Port Channel
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: create_portchannel
    portchannel: PortChannel0001
```

#### Add Member to Port Channel
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet0
```

### BGP Configuration

#### Configure BGP Instance
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_bgp
    bgp_asn: 65000
    router_id: 10.0.0.1
```

#### Add BGP Neighbor
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_bgp_neighbor
    bgp_asn: 65000
    neighbor_ip: 10.0.0.2
    neighbor_asn: 65001
```

#### Configure BGP Network
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_bgp_network
    bgp_asn: 65000
    network: 192.168.0.0/16
```

### ACL Configuration

#### Create ACL Table
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: create_acl_table
    table_name: DATA_ACL
    table_type: L3
    stage: INGRESS
```

#### Add ACL Rule
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_acl_rule
    table_name: DATA_ACL
    rule_name: RULE_1
    priority: 100
    action: DROP
    src_ip: 192.168.100.0/24
    dst_ip: 10.0.0.0/8
    protocol: TCP
    dst_port: 22
```

### Static Route Configuration

#### Add Static Route
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_static_route
    prefix: 172.16.0.0/16
    nexthop: 10.0.0.254
```

### System Configuration

#### Set Hostname
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: set_hostname
    hostname: dc1-leaf-01
```

#### Configure NTP Server
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_ntp
    ntp_server: 0.pool.ntp.org
```

#### Save Configuration
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

## REST API vs CLI Mode

### REST API Mode (Default)
When `use_rest: true` (default), the module uses SONiC's RESTCONF API:
- **Advantages**: Structured responses, idempotent operations, programmatic access
- **Requirements**: SONiC REST server must be running
- **Endpoint**: `https://<switch-ip>:443/restconf/data/`

### CLI Mode
When `use_rest: false`, the module uses SSH and SONiC CLI commands:
- **Advantages**: Always available, fallback option
- **Requirements**: SSH access with password authentication
- **Tools Used**: `config` command, `show` command, `vtysh` for BGP

## SONiC REST API Endpoints

The module uses the following RESTCONF paths:

- **VLANs**: `/restconf/data/sonic-vlan:sonic-vlan`
- **Interfaces**: `/restconf/data/openconfig-interfaces:interfaces`
- **Port Channels**: `/restconf/data/sonic-portchannel:sonic-portchannel`
- **BGP**: `/restconf/data/openconfig-bgp:bgp`
- **ACLs**: `/restconf/data/sonic-acl:sonic-acl`

## Requirements

### SONiC Switch Requirements
- SONiC OS (any recent version)
- REST API server enabled (for REST mode)
- SSH access enabled (for CLI mode or as fallback)

### Host Requirements
- `curl` (for REST API calls)
- `sshpass` (for SSH authentication in CLI mode)
- Network connectivity to SONiC switch management interface

## Security Considerations

1. **Credentials**: Store passwords securely, consider using vault or encrypted variables
2. **SSL Certificates**: Set `validate_certs: true` in production when using proper certificates
3. **Network Access**: Restrict management network access to authorized hosts
4. **SSH Keys**: Consider using SSH key authentication instead of passwords
5. **RBAC**: Use appropriate user roles with minimum required privileges

## Troubleshooting

### REST API Issues
1. **Connection Refused**
   - Verify REST API server is running: `systemctl status restserver`
   - Check firewall rules allow port 443

2. **Authentication Failed**
   - Verify username/password are correct
   - Check user has necessary privileges

3. **SSL Certificate Errors**
   - Set `validate_certs: false` for self-signed certificates
   - Install proper CA certificates if using validated SSL

### CLI Issues
1. **SSH Connection Failed**
   - Verify SSH is enabled and running
   - Check network connectivity
   - Verify `sshpass` is installed on orchestrator host

2. **Command Not Found**
   - Verify SONiC version compatibility
   - Some commands may vary between SONiC versions

## Performance Considerations

- **REST API**: Generally faster and more efficient
- **CLI Mode**: May be slower due to SSH overhead
- **Batch Operations**: Consider grouping related operations in task blocks
- **Parallel Execution**: Use parallel strategy for configuring multiple switches

## Version Compatibility

This module has been tested with:
- SONiC 202012 and later
- SONiC 202106 and later
- SONiC 202205 and later

Note: Some features may vary between SONiC versions. Always test in a non-production environment first.

## Related Modules

- `networking/cisco` - Cisco IOS/IOS-XE management
- `networking/juniper` - Juniper JunOS management
- `networking/arista` - Arista EOS management

## References

- [SONiC Official Documentation](https://github.com/sonic-net/SONiC/wiki)
- [SONiC Command Reference](https://github.com/sonic-net/sonic-utilities/blob/master/doc/Command-Reference.md)
- [SONiC REST API](https://github.com/sonic-net/SONiC/blob/master/doc/mgmt/SONiC_Management_Framework_Developer_Guide.md)
- [OpenConfig Models](https://github.com/openconfig/public)

## Support

For issues, questions, or contributions related to this module, please refer to the OpenFroyo project documentation.
