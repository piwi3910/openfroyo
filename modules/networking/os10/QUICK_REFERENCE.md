# Dell OS10 Module - Quick Reference

Fast reference guide for common Dell OS10 switch management workflows.

## Quick Command Lookup

| Task | Command |
|------|---------|
| Create VLAN | `create_vlan` |
| Delete VLAN | `delete_vlan` |
| Configure trunk port | `configure_interface` + `switchport_mode: trunk` |
| Configure access port | `configure_interface` + `switchport_mode: access` |
| Enable interface | `no_shutdown_interface` |
| Disable interface | `shutdown_interface` |
| Create LAG | `create_port_channel` |
| Add port to LAG | `add_interface_to_lag` |
| Configure IP | `configure_ip_interface` |
| Add static route | `configure_static_route` |
| Set hostname | `set_hostname` |
| Save config | `save_config` |
| Show VLANs | `show_vlan_brief` |
| Show interfaces | `show_interface_status` |
| Show routes | `show_ip_route` |

## Common Workflows

### 1. Basic Switch Setup

```yaml
# Set hostname
- name: Configure hostname
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_hostname
    hostname: CORE-SW-01

# Configure NTP
- name: Configure NTP server
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_ntp
    ntp_server: 10.0.0.1

# Configure DNS
- name: Configure DNS server
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_dns
    dns_server: 8.8.8.8

# Save configuration
- name: Save config
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: save_config
```

### 2. VLAN Configuration

```yaml
# Create multiple VLANs
- name: Create DATA VLAN
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 100
    vlan_name: DATA

- name: Create VOICE VLAN
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 200
    vlan_name: VOICE

- name: Create MGMT VLAN
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 999
    vlan_name: MGMT

# Verify VLANs
- name: Show all VLANs
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_vlan_brief
```

### 3. Access Port Configuration

```yaml
# Configure server access port
- name: Configure server port
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/10
    description: Production Server
    switchport_mode: access
    access_vlan: 100
    shutdown: false
```

### 4. Trunk Port Configuration

```yaml
# Configure uplink trunk
- name: Configure trunk to core
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/1
    description: Uplink to CORE-SW-01
    switchport_mode: trunk
    trunk_allowed_vlans: 100,200,999
    mtu: 9000
    shutdown: false
```

### 5. Port Channel (LAG) Setup

```yaml
# Create port channel
- name: Create LAG 10
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_port_channel
    port_channel_id: 10
    lag_mode: active

# Add first interface
- name: Add interface 1/1/1 to LAG
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: add_interface_to_lag
    interface: ethernet 1/1/1
    port_channel_id: 10
    lag_mode: active

# Add second interface
- name: Add interface 1/1/2 to LAG
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: add_interface_to_lag
    interface: ethernet 1/1/2
    port_channel_id: 10
    lag_mode: active

# Configure port channel as trunk
- name: Configure LAG as trunk
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: port-channel 10
    description: LAG to Distribution
    switchport_mode: trunk
    trunk_allowed_vlans: 100,200,999
```

### 6. Layer 3 Configuration

```yaml
# Create VLAN interfaces for routing
- name: Create VLAN 100 interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan_interface
    vlan_id: 100
    ip_address: 192.168.100.1
    netmask: 255.255.255.0

- name: Create VLAN 200 interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan_interface
    vlan_id: 200
    ip_address: 192.168.200.1
    netmask: 255.255.255.0

# Add default route
- name: Configure default route
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_static_route
    destination: 0.0.0.0
    netmask: 0.0.0.0
    next_hop: 192.168.1.1
```

### 7. Spanning Tree Configuration

```yaml
# Configure RSTP
- name: Enable RSTP
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_spanning_tree
    stp_mode: rstp

# Set priority (make this switch root)
- name: Set STP priority
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_stp_priority
    stp_priority: 4096
```

### 8. User Management

```yaml
# Create operator account
- name: Create operator user
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_user
    user: netops
    new_password: SecurePass123!
    role: netoperator

# Create admin account
- name: Create admin user
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_user
    user: netadmin
    new_password: AdminPass456!
    role: sysadmin
```

### 9. Access Control Lists

```yaml
# Create management ACL
- name: Create MGMT ACL
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_acl
    acl_name: MGMT_ACCESS
    acl_type: standard
    sequence: 10
    action: permit
    source: 192.168.1.0/24

# Create web server ACL
- name: Create WEB ACL
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_acl
    acl_name: WEB_ACCESS
    acl_type: extended
    sequence: 10
    action: permit
    protocol: tcp
    source: any
    destination_acl: 192.168.100.10/32
```

### 10. Complete Edge Port Setup

```yaml
# Complete configuration for an edge access port
- name: Configure edge port
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/20
    description: User Workstation - Desk 20
    switchport_mode: access
    access_vlan: 100
    speed: auto
    mtu: 1500
    shutdown: false
```

### 11. Monitoring and Verification

```yaml
# Show running configuration
- name: Get running config
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_running_config

# Show interface status
- name: Check interface status
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_interface_status

# Show VLAN summary
- name: Check VLANs
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_vlan_brief

# Show routing table
- name: Check routes
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_ip_route

# Show spanning tree
- name: Check STP
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_spanning_tree
```

## Stack File Examples

### Complete Stack: Basic Switch Configuration

```yaml
# File: stacks/configure_access_switch.ofy
inventory:
  hosts:
    - access-sw-01

defaults:
  host: 192.168.1.10
  username: admin
  password: "{{ vault.switch_password }}"

run:
  # System configuration
  - name: Initial Setup
    tasks:
      - name: Set hostname
        module: networking/os10
        vars:
          command: set_hostname
          hostname: ACCESS-SW-01

      - name: Configure NTP
        module: networking/os10
        vars:
          command: configure_ntp
          ntp_server: 10.0.0.1

  # VLAN configuration
  - name: Configure VLANs
    tasks:
      - name: Create DATA VLAN
        module: networking/os10
        vars:
          command: create_vlan
          vlan_id: 100
          vlan_name: DATA

      - name: Create VOICE VLAN
        module: networking/os10
        vars:
          command: create_vlan
          vlan_id: 200
          vlan_name: VOICE

  # Interface configuration
  - name: Configure Uplink
    tasks:
      - name: Configure uplink port
        module: networking/os10
        vars:
          command: configure_interface
          interface: ethernet 1/1/1
          description: Uplink to Core
          switchport_mode: trunk
          trunk_allowed_vlans: 100,200

  # Save configuration
  - name: Save Configuration
    tasks:
      - name: Write memory
        module: networking/os10
        vars:
          command: save_config
```

### Complete Stack: Multiple Switches

```yaml
# File: stacks/configure_distribution_layer.ofy
inventory:
  groups:
    - name: distribution
      hosts:
        - dist-sw-01
        - dist-sw-02

defaults:
  username: admin
  password: "{{ vault.switch_password }}"

run:
  # Configure first switch
  - name: Configure DIST-SW-01
    hosts: dist-sw-01
    vars:
      host: 192.168.1.11
    tasks:
      - name: Set hostname
        module: networking/os10
        vars:
          command: set_hostname
          hostname: DIST-SW-01

      - name: Set STP priority
        module: networking/os10
        vars:
          command: set_stp_priority
          stp_priority: 4096

  # Configure second switch
  - name: Configure DIST-SW-02
    hosts: dist-sw-02
    vars:
      host: 192.168.1.12
    tasks:
      - name: Set hostname
        module: networking/os10
        vars:
          command: set_hostname
          hostname: DIST-SW-02

      - name: Set STP priority
        module: networking/os10
        vars:
          command: set_stp_priority
          stp_priority: 8192

  # Configure VLANs on all switches
  - name: Configure VLANs
    hosts: "@group:distribution"
    tasks:
      - name: Create VLAN 100
        module: networking/os10
        vars:
          command: create_vlan
          vlan_id: 100
          vlan_name: DATA
```

## Tips and Best Practices

### 1. Always Save Configuration
```yaml
# Always end your configuration with save_config
- name: Save configuration
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: save_config
```

### 2. Use Descriptive Interface Descriptions
```yaml
# Good: Descriptive and informative
description: Web-Server-01 Port 1 - Production

# Bad: Not descriptive
description: server
```

### 3. Configure Trunk Ports with Specific VLANs
```yaml
# Good: Explicit VLAN list
trunk_allowed_vlans: 100,200,300

# Avoid: Allowing all VLANs (security risk)
trunk_allowed_vlans: all
```

### 4. Use REST API When Available
```yaml
# Preferred: Use REST API
vars:
  use_rest: true

# Fallback: CLI via SSH
vars:
  use_rest: false
```

### 5. Document Your Configurations
```yaml
# Use task names to document what you're doing
- name: Configure trunk port for inter-switch link to CORE-SW-01
  module: networking/os10
  vars:
    command: configure_interface
    interface: ethernet 1/1/1
    # ... configuration ...
```

### 6. Use Variables for Repetitive Values
```yaml
defaults:
  host: 192.168.1.10
  username: admin
  password: "{{ vault.switch_password }}"
  use_rest: true
  validate_certs: false

run:
  - name: Configure VLANs
    tasks:
      # Variables from defaults are automatically used
      - name: Create VLAN 100
        module: networking/os10
        vars:
          command: create_vlan
          vlan_id: 100
```

## Troubleshooting Quick Fixes

### Can't Connect via REST API
```yaml
# Try using CLI instead
vars:
  use_rest: false
```

### Certificate Validation Errors
```yaml
# Disable certificate validation (test only)
vars:
  validate_certs: false
```

### Command Timeouts
```yaml
# Break large configurations into smaller tasks
# Save after each major section
```

### Interface Not Coming Up
```yaml
# Explicitly enable the interface
- name: Enable interface
  module: networking/os10
  vars:
    command: no_shutdown_interface
    interface: ethernet 1/1/10
```

## Parameter Quick Reference

### Connection Parameters
```yaml
host: <switch_ip>           # Required
username: <admin_user>      # Required
password: <admin_pass>      # Required
use_rest: true|false        # Optional, default: true
validate_certs: true|false  # Optional, default: false
port: <https_port>          # Optional, default: 443
ssh_port: <ssh_port>        # Optional, default: 22
```

### Common Interface Parameters
```yaml
interface: ethernet 1/1/1   # Physical interface
interface: vlan 100         # VLAN interface
interface: port-channel 10  # Port channel
description: <text>         # Interface description
mtu: 1500|9000             # MTU size
speed: auto|1000|10000      # Interface speed
shutdown: true|false        # Admin state
```

### VLAN Parameters
```yaml
vlan_id: 1-4094            # VLAN number
vlan_name: <name>          # VLAN name
vlan_description: <text>   # VLAN description
```

### Switchport Parameters
```yaml
switchport_mode: access|trunk        # Switchport mode
access_vlan: <vlan_id>              # Access VLAN
trunk_allowed_vlans: 100,200,300    # Trunk VLANs
```

### Layer 3 Parameters
```yaml
ip_address: 192.168.1.1    # IP address
netmask: 255.255.255.0     # Subnet mask
destination: 0.0.0.0       # Route destination
next_hop: 192.168.1.254    # Gateway IP
```

## See Also

- **README.md** - Comprehensive module documentation
- **COMMANDS.md** - Detailed command reference
- **test.ofy** - Complete test examples
