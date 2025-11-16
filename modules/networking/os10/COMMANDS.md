# Dell OS10 Module - Command Reference

Complete command reference for all 36 operations supported by the os10 module.

## Table of Contents

- [VLAN Management](#vlan-management)
- [Interface Management](#interface-management)
- [Port Channel/LAG](#port-channellag)
- [Layer 3 Configuration](#layer-3-configuration)
- [Spanning Tree](#spanning-tree)
- [System Configuration](#system-configuration)
- [User Management](#user-management)
- [ACL Configuration](#acl-configuration)

---

## VLAN Management

### create_vlan

Creates a new VLAN on the switch.

**Required Parameters:**
- `vlan_id` - VLAN ID (1-4094)

**Optional Parameters:**
- `vlan_name` - VLAN name

**Example:**
```yaml
- name: Create VLAN 100
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 100
    vlan_name: PRODUCTION
```

**CLI Equivalent:**
```
configure terminal
interface vlan 100
name PRODUCTION
exit
write memory
```

**REST API Equivalent:**
```bash
POST /restconf/data/dell-vlan:vlans
{
  "dell-vlan:vlan": [{
    "vlan-id": 100,
    "name": "PRODUCTION"
  }]
}
```

---

### delete_vlan

Deletes an existing VLAN.

**Required Parameters:**
- `vlan_id` - VLAN ID to delete

**Example:**
```yaml
- name: Delete VLAN 100
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: delete_vlan
    vlan_id: 100
```

**CLI Equivalent:**
```
configure terminal
no interface vlan 100
write memory
```

---

### configure_vlan

Configures properties of an existing VLAN.

**Required Parameters:**
- `vlan_id` - VLAN ID

**Optional Parameters:**
- `vlan_name` - VLAN name
- `vlan_description` - VLAN description

**Example:**
```yaml
- name: Configure VLAN 100
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_vlan
    vlan_id: 100
    vlan_name: PROD_DATA
    vlan_description: Production data network
```

**CLI Equivalent:**
```
configure terminal
interface vlan 100
name PROD_DATA
description Production data network
exit
write memory
```

---

### show_vlan

Displays VLAN information.

**Optional Parameters:**
- `vlan_id` - Specific VLAN ID (shows all if omitted)

**Example:**
```yaml
- name: Show all VLANs
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_vlan

- name: Show specific VLAN
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_vlan
    vlan_id: 100
```

**CLI Equivalent:**
```
show vlan
show vlan id 100
```

---

### show_vlan_brief

Displays brief VLAN summary.

**Example:**
```yaml
- name: Show VLAN brief
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_vlan_brief
```

**CLI Equivalent:**
```
show vlan brief
```

---

## Interface Management

### configure_interface

Configures interface properties.

**Required Parameters:**
- `interface` - Interface name (e.g., "ethernet 1/1/1")

**Optional Parameters:**
- `description` - Interface description
- `mtu` - MTU size
- `speed` - Interface speed (auto, 1000, 10000, 25000, 40000, 100000)
- `shutdown` - Shutdown state (true/false)
- `switchport_mode` - access or trunk
- `access_vlan` - Access VLAN ID (for access mode)
- `trunk_allowed_vlans` - Comma-separated VLAN list (for trunk mode)

**Example - Access Port:**
```yaml
- name: Configure access port
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/10
    description: Server Port
    switchport_mode: access
    access_vlan: 100
```

**Example - Trunk Port:**
```yaml
- name: Configure trunk port
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
    mtu: 9000
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
description Uplink to Core
mtu 9000
switchport mode trunk
switchport trunk allowed vlan 100,200,300
no shutdown
exit
write memory
```

---

### shutdown_interface

Disables an interface.

**Required Parameters:**
- `interface` - Interface name

**Example:**
```yaml
- name: Shutdown interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: shutdown_interface
    interface: ethernet 1/1/10
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/10
shutdown
exit
write memory
```

---

### no_shutdown_interface

Enables an interface.

**Required Parameters:**
- `interface` - Interface name

**Example:**
```yaml
- name: Enable interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: no_shutdown_interface
    interface: ethernet 1/1/10
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/10
no shutdown
exit
write memory
```

---

### set_description

Sets interface description.

**Required Parameters:**
- `interface` - Interface name
- `description` - Description text

**Example:**
```yaml
- name: Set interface description
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_description
    interface: ethernet 1/1/10
    description: Web Server Connection
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/10
description Web Server Connection
exit
write memory
```

---

### set_mtu

Configures interface MTU.

**Required Parameters:**
- `interface` - Interface name
- `mtu` - MTU size (1280-9216)

**Example:**
```yaml
- name: Set jumbo frames
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_mtu
    interface: ethernet 1/1/1
    mtu: 9000
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
mtu 9000
exit
write memory
```

---

### set_speed

Configures interface speed.

**Required Parameters:**
- `interface` - Interface name
- `speed` - Speed value (auto, 1000, 10000, 25000, 40000, 100000)

**Example:**
```yaml
- name: Set interface speed
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_speed
    interface: ethernet 1/1/1
    speed: 10000
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
speed 10000
exit
write memory
```

---

### show_interface

Displays interface information.

**Optional Parameters:**
- `interface` - Specific interface (shows all if omitted)

**Example:**
```yaml
- name: Show all interfaces
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_interface

- name: Show specific interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_interface
    interface: ethernet 1/1/1
```

**CLI Equivalent:**
```
show interfaces
show interface ethernet 1/1/1
```

---

### show_interface_status

Displays interface status summary.

**Example:**
```yaml
- name: Show interface status
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_interface_status
```

**CLI Equivalent:**
```
show interface status
```

---

## Port Channel/LAG

### create_port_channel

Creates a port channel (LAG).

**Required Parameters:**
- `port_channel_id` - Port channel number

**Optional Parameters:**
- `lag_mode` - LAG mode (active, passive, on)

**Example:**
```yaml
- name: Create port channel
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_port_channel
    port_channel_id: 10
    lag_mode: active
```

**CLI Equivalent:**
```
configure terminal
interface port-channel 10
channel-group mode active
exit
write memory
```

---

### delete_port_channel

Deletes a port channel.

**Required Parameters:**
- `port_channel_id` - Port channel number

**Example:**
```yaml
- name: Delete port channel
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: delete_port_channel
    port_channel_id: 10
```

**CLI Equivalent:**
```
configure terminal
no interface port-channel 10
write memory
```

---

### add_interface_to_lag

Adds an interface to a LAG.

**Required Parameters:**
- `interface` - Interface name
- `port_channel_id` - Port channel number

**Optional Parameters:**
- `lag_mode` - LAG mode (active, passive, on) - default: active

**Example:**
```yaml
- name: Add interface to LAG
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: add_interface_to_lag
    interface: ethernet 1/1/1
    port_channel_id: 10
    lag_mode: active
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
channel-group 10 mode active
exit
write memory
```

---

### remove_interface_from_lag

Removes an interface from a LAG.

**Required Parameters:**
- `interface` - Interface name
- `port_channel_id` - Port channel number

**Example:**
```yaml
- name: Remove interface from LAG
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: remove_interface_from_lag
    interface: ethernet 1/1/1
    port_channel_id: 10
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
no channel-group 10
exit
write memory
```

---

## Layer 3 Configuration

### configure_ip_interface

Configures IP address on an interface.

**Required Parameters:**
- `interface` - Interface name
- `ip_address` - IP address
- `netmask` - Subnet mask

**Example:**
```yaml
- name: Configure IP on interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_ip_interface
    interface: ethernet 1/1/1
    ip_address: 10.0.0.1
    netmask: 255.255.255.0
```

**CLI Equivalent:**
```
configure terminal
interface ethernet 1/1/1
no switchport
ip address 10.0.0.1 255.255.255.0
exit
write memory
```

---

### create_vlan_interface

Creates a VLAN interface with IP address.

**Required Parameters:**
- `vlan_id` - VLAN ID
- `ip_address` - IP address
- `netmask` - Subnet mask

**Example:**
```yaml
- name: Create VLAN interface
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan_interface
    vlan_id: 100
    ip_address: 192.168.100.1
    netmask: 255.255.255.0
```

**CLI Equivalent:**
```
configure terminal
interface vlan 100
ip address 192.168.100.1 255.255.255.0
no shutdown
exit
write memory
```

---

### configure_static_route

Configures a static route.

**Required Parameters:**
- `destination` - Destination network
- `netmask` - Network mask
- `next_hop` - Next hop IP address

**Example:**
```yaml
- name: Add static route
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_static_route
    destination: 10.0.0.0
    netmask: 255.255.255.0
    next_hop: 192.168.1.1
```

**CLI Equivalent:**
```
configure terminal
ip route 10.0.0.0 255.255.255.0 192.168.1.1
write memory
```

---

### show_ip_interface

Displays IP interface information.

**Optional Parameters:**
- `interface` - Specific interface (shows all if omitted)

**Example:**
```yaml
- name: Show IP interfaces
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_ip_interface
```

**CLI Equivalent:**
```
show ip interface
show ip interface ethernet 1/1/1
```

---

### show_ip_route

Displays IP routing table.

**Example:**
```yaml
- name: Show routing table
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_ip_route
```

**CLI Equivalent:**
```
show ip route
```

---

## Spanning Tree

### configure_spanning_tree

Configures spanning tree mode.

**Optional Parameters:**
- `stp_mode` - STP mode (rstp, pvst, mstp) - default: rstp

**Example:**
```yaml
- name: Configure STP mode
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_spanning_tree
    stp_mode: rstp
```

**CLI Equivalent:**
```
configure terminal
spanning-tree mode rstp
write memory
```

---

### set_stp_priority

Sets spanning tree priority.

**Required Parameters:**
- `stp_priority` - Priority value (0-61440, multiples of 4096)

**Optional Parameters:**
- `vlan_id` - VLAN ID for per-VLAN priority

**Example - Global:**
```yaml
- name: Set global STP priority
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_stp_priority
    stp_priority: 4096
```

**Example - Per-VLAN:**
```yaml
- name: Set VLAN STP priority
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_stp_priority
    stp_priority: 4096
    vlan_id: 100
```

**CLI Equivalent:**
```
configure terminal
spanning-tree priority 4096
spanning-tree vlan 100 priority 4096
write memory
```

---

### show_spanning_tree

Displays spanning tree information.

**Optional Parameters:**
- `vlan_id` - Specific VLAN (shows all if omitted)

**Example:**
```yaml
- name: Show spanning tree
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_spanning_tree
```

**CLI Equivalent:**
```
show spanning-tree
show spanning-tree vlan 100
```

---

## System Configuration

### set_hostname

Configures switch hostname.

**Required Parameters:**
- `hostname` - New hostname

**Example:**
```yaml
- name: Set hostname
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: set_hostname
    hostname: CORE-SW-01
```

**CLI Equivalent:**
```
configure terminal
hostname CORE-SW-01
write memory
```

---

### configure_ntp

Configures NTP server.

**Required Parameters:**
- `ntp_server` - NTP server IP or hostname

**Example:**
```yaml
- name: Configure NTP
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_ntp
    ntp_server: time.nist.gov
```

**CLI Equivalent:**
```
configure terminal
ntp server time.nist.gov
write memory
```

---

### configure_dns

Configures DNS server.

**Required Parameters:**
- `dns_server` - DNS server IP

**Example:**
```yaml
- name: Configure DNS
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_dns
    dns_server: 8.8.8.8
```

**CLI Equivalent:**
```
configure terminal
ip name-server 8.8.8.8
write memory
```

---

### save_config

Saves running configuration to startup configuration.

**Example:**
```yaml
- name: Save configuration
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: save_config
```

**CLI Equivalent:**
```
write memory
```

---

### show_running_config

Displays running configuration.

**Example:**
```yaml
- name: Show running config
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_running_config
```

**CLI Equivalent:**
```
show running-config
```

---

## User Management

### create_user

Creates a new user account.

**Required Parameters:**
- `user` - Username
- `new_password` - Password

**Optional Parameters:**
- `role` - User role (sysadmin, netadmin, netoperator) - default: sysadmin

**Example:**
```yaml
- name: Create user
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_user
    user: operator1
    new_password: securepass123
    role: netoperator
```

**CLI Equivalent:**
```
configure terminal
username operator1 password securepass123 role netoperator
write memory
```

---

### delete_user

Deletes a user account.

**Required Parameters:**
- `user` - Username to delete

**Example:**
```yaml
- name: Delete user
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: delete_user
    user: operator1
```

**CLI Equivalent:**
```
configure terminal
no username operator1
write memory
```

---

### change_password

Changes user password.

**Required Parameters:**
- `user` - Username
- `new_password` - New password

**Example:**
```yaml
- name: Change password
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: change_password
    user: operator1
    new_password: newpass456
```

**CLI Equivalent:**
```
configure terminal
username operator1 password newpass456
write memory
```

---

## ACL Configuration

### create_acl

Creates an access control list.

**Required Parameters:**
- `acl_name` - ACL name

**Optional Parameters:**
- `acl_type` - standard or extended (default: standard)
- `sequence` - Sequence number
- `action` - permit or deny (default: permit)
- `protocol` - Protocol (ip, tcp, udp, icmp, etc.)
- `source` - Source address/network
- `destination_acl` - Destination address/network

**Example - Standard ACL:**
```yaml
- name: Create standard ACL
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_acl
    acl_name: MGMT_ACCESS
    acl_type: standard
    action: permit
    source: 192.168.1.0/24
```

**Example - Extended ACL:**
```yaml
- name: Create extended ACL
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

**CLI Equivalent:**
```
configure terminal
ip access-list standard MGMT_ACCESS
permit 192.168.1.0/24
exit
write memory
```

---

### delete_acl

Deletes an access control list.

**Required Parameters:**
- `acl_name` - ACL name

**Optional Parameters:**
- `acl_type` - standard or extended (default: standard)

**Example:**
```yaml
- name: Delete ACL
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: delete_acl
    acl_name: MGMT_ACCESS
    acl_type: standard
```

**CLI Equivalent:**
```
configure terminal
no ip access-list standard MGMT_ACCESS
write memory
```

---

### show_acl

Displays ACL configuration.

**Optional Parameters:**
- `acl_name` - Specific ACL name (shows all if omitted)

**Example:**
```yaml
- name: Show all ACLs
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_acl

- name: Show specific ACL
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: show_acl
    acl_name: MGMT_ACCESS
```

**CLI Equivalent:**
```
show ip access-lists
show ip access-lists MGMT_ACCESS
```

---

## Common Patterns

### Configuration Workflow

```yaml
# 1. Create VLANs
- name: Create data VLAN
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: create_vlan
    vlan_id: 100
    vlan_name: DATA

# 2. Configure interfaces
- name: Configure trunk port
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: configure_interface
    interface: ethernet 1/1/1
    switchport_mode: trunk
    trunk_allowed_vlans: 100

# 3. Save configuration
- name: Save config
  module: networking/os10
  vars:
    host: 192.168.1.10
    username: admin
    password: admin123
    command: save_config
```

### Error Checking

Always check the status field in the output:
- `ok` - Command executed successfully, no changes
- `changed` - Configuration was modified
- `failed` - Command failed, see message for details
