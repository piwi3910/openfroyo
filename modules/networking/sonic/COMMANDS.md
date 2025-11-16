# SONiC Module Command Reference

Complete reference for all 35+ commands supported by the SONiC module.

## Command Index

- [VLAN Management](#vlan-management)
- [Interface Management](#interface-management)
- [Port Channel/LAG](#port-channellag)
- [BGP Configuration](#bgp-configuration)
- [ACL Configuration](#acl-configuration)
- [Route Management](#route-management)
- [System Configuration](#system-configuration)

---

## VLAN Management

### create_vlan

Create a new VLAN on the switch.

**Required Variables:**
- `command: create_vlan`
- `vlan_id: <1-4094>` - VLAN ID

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: create_vlan
    vlan_id: 100
```

**REST API Call:**
```
POST /restconf/data/sonic-vlan:sonic-vlan/VLAN/VLAN_LIST
{"sonic-vlan:VLAN_LIST":[{"name":"Vlan100"}]}
```

**CLI Equivalent:**
```bash
config vlan add 100
```

---

### delete_vlan

Delete an existing VLAN from the switch.

**Required Variables:**
- `command: delete_vlan`
- `vlan_id: <1-4094>` - VLAN ID to delete

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: delete_vlan
    vlan_id: 100
```

**REST API Call:**
```
DELETE /restconf/data/sonic-vlan:sonic-vlan/VLAN/VLAN_LIST=Vlan100
```

**CLI Equivalent:**
```bash
config vlan del 100
```

---

### add_vlan_member

Add an interface to a VLAN as either tagged or untagged member.

**Required Variables:**
- `command: add_vlan_member`
- `vlan_id: <1-4094>` - VLAN ID
- `interface: <interface_name>` - Interface name (e.g., Ethernet0)

**Optional Variables:**
- `tagged: <true|false>` - Tagged (true) or untagged (false) member (default: false)

**Example (Untagged):**
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

**Example (Tagged):**
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

**CLI Equivalent:**
```bash
# Untagged
config vlan member add 100 Ethernet0

# Tagged
config vlan member add 200 Ethernet4 -t
```

---

### remove_vlan_member

Remove an interface from a VLAN.

**Required Variables:**
- `command: remove_vlan_member`
- `vlan_id: <1-4094>` - VLAN ID
- `interface: <interface_name>` - Interface name

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: remove_vlan_member
    vlan_id: 100
    interface: Ethernet0
```

**CLI Equivalent:**
```bash
config vlan member del 100 Ethernet0
```

---

### show_vlan

Display VLAN configuration.

**Required Variables:**
- `command: show_vlan`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_vlan
```

**CLI Equivalent:**
```bash
show vlan brief
```

---

## Interface Management

### configure_interface

Configure multiple interface settings in one operation.

**Required Variables:**
- `command: configure_interface`
- `interface: <interface_name>` - Interface name

**Optional Variables:**
- `mtu: <integer>` - MTU size
- `speed: <speed>` - Interface speed (e.g., 100000, 40000, 10000)
- `admin_status: <up|down>` - Administrative status

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_interface
    interface: Ethernet8
    mtu: 9000
    admin_status: up
```

---

### set_interface_ip

Configure IP address on an interface.

**Required Variables:**
- `command: set_interface_ip`
- `interface: <interface_name>` - Interface name
- `ip_address: <ip_address>` - IP address
- `prefix_length: <1-32>` - Prefix length (CIDR notation)

**Example:**
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

**CLI Equivalent:**
```bash
config interface ip add Ethernet8 10.0.0.1/24
```

---

### set_interface_mtu

Set MTU (Maximum Transmission Unit) on an interface.

**Required Variables:**
- `command: set_interface_mtu`
- `interface: <interface_name>` - Interface name
- `mtu: <integer>` - MTU size (typically 1500-9216)

**Example:**
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

**CLI Equivalent:**
```bash
config interface mtu Ethernet12 9000
```

---

### set_interface_speed

Set interface speed.

**Required Variables:**
- `command: set_interface_speed`
- `interface: <interface_name>` - Interface name
- `speed: <speed>` - Speed value (e.g., 100000 for 100G, 40000 for 40G)

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: set_interface_speed
    interface: Ethernet0
    speed: 100000
```

**CLI Equivalent:**
```bash
config interface speed Ethernet0 100000
```

---

### shutdown_interface

Administratively disable an interface.

**Required Variables:**
- `command: shutdown_interface`
- `interface: <interface_name>` - Interface name

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: shutdown_interface
    interface: Ethernet16
```

**CLI Equivalent:**
```bash
config interface shutdown Ethernet16
```

---

### startup_interface

Administratively enable an interface.

**Required Variables:**
- `command: startup_interface`
- `interface: <interface_name>` - Interface name

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: startup_interface
    interface: Ethernet16
```

**CLI Equivalent:**
```bash
config interface startup Ethernet16
```

---

### show_interface

Display interface information.

**Required Variables:**
- `command: show_interface`

**Optional Variables:**
- `interface: <interface_name>` - Specific interface (omit for all interfaces)

**Example (All interfaces):**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_interface
```

**Example (Specific interface):**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_interface
    interface: Ethernet0
```

**CLI Equivalent:**
```bash
show interfaces status
show interfaces status Ethernet0
```

---

### show_interface_status

Display interface status summary.

**Required Variables:**
- `command: show_interface_status`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_interface_status
```

**CLI Equivalent:**
```bash
show interfaces status
```

---

## Port Channel/LAG

### create_portchannel

Create a port channel (Link Aggregation Group).

**Required Variables:**
- `command: create_portchannel`
- `portchannel: <name>` - Port channel name (e.g., PortChannel0001)

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: create_portchannel
    portchannel: PortChannel0001
```

**CLI Equivalent:**
```bash
config portchannel add PortChannel0001
```

---

### delete_portchannel

Delete a port channel.

**Required Variables:**
- `command: delete_portchannel`
- `portchannel: <name>` - Port channel name

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: delete_portchannel
    portchannel: PortChannel0001
```

**CLI Equivalent:**
```bash
config portchannel del PortChannel0001
```

---

### add_portchannel_member

Add member interface to a port channel.

**Required Variables:**
- `command: add_portchannel_member`
- `portchannel: <name>` - Port channel name
- `interface: <interface_name>` - Member interface

**Example:**
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

**CLI Equivalent:**
```bash
config portchannel member add PortChannel0001 Ethernet0
```

---

### remove_portchannel_member

Remove member interface from a port channel.

**Required Variables:**
- `command: remove_portchannel_member`
- `portchannel: <name>` - Port channel name
- `interface: <interface_name>` - Member interface

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: remove_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet0
```

**CLI Equivalent:**
```bash
config portchannel member del PortChannel0001 Ethernet0
```

---

### show_portchannel

Display port channel configuration.

**Required Variables:**
- `command: show_portchannel`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_portchannel
```

**CLI Equivalent:**
```bash
show interfaces portchannel
```

---

## BGP Configuration

### configure_bgp

Configure BGP routing instance.

**Required Variables:**
- `command: configure_bgp`
- `bgp_asn: <integer>` - BGP Autonomous System Number

**Optional Variables:**
- `router_id: <ip_address>` - BGP router ID

**Example:**
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

**CLI Equivalent:**
```bash
vtysh -c "configure terminal; router bgp 65000; bgp router-id 10.0.0.1"
```

---

### add_bgp_neighbor

Add BGP neighbor peering.

**Required Variables:**
- `command: add_bgp_neighbor`
- `bgp_asn: <integer>` - Local BGP AS number
- `neighbor_ip: <ip_address>` - Neighbor IP address
- `neighbor_asn: <integer>` - Neighbor AS number

**Example:**
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

**CLI Equivalent:**
```bash
vtysh -c "configure terminal; router bgp 65000; neighbor 10.0.0.2 remote-as 65001"
```

---

### remove_bgp_neighbor

Remove BGP neighbor peering.

**Required Variables:**
- `command: remove_bgp_neighbor`
- `bgp_asn: <integer>` - Local BGP AS number
- `neighbor_ip: <ip_address>` - Neighbor IP address

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: remove_bgp_neighbor
    bgp_asn: 65000
    neighbor_ip: 10.0.0.2
```

**CLI Equivalent:**
```bash
vtysh -c "configure terminal; router bgp 65000; no neighbor 10.0.0.2"
```

---

### configure_bgp_network

Advertise network prefix via BGP.

**Required Variables:**
- `command: configure_bgp_network`
- `bgp_asn: <integer>` - BGP AS number
- `network: <prefix>` - Network prefix in CIDR notation

**Example:**
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

**CLI Equivalent:**
```bash
vtysh -c "configure terminal; router bgp 65000; network 192.168.0.0/16"
```

---

### show_bgp_summary

Display BGP summary information.

**Required Variables:**
- `command: show_bgp_summary`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_bgp_summary
```

**CLI Equivalent:**
```bash
show ip bgp summary
```

---

### show_bgp_neighbors

Display detailed BGP neighbor information.

**Required Variables:**
- `command: show_bgp_neighbors`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_bgp_neighbors
```

**CLI Equivalent:**
```bash
show ip bgp neighbors
```

---

## ACL Configuration

### create_acl_table

Create an Access Control List table.

**Required Variables:**
- `command: create_acl_table`
- `table_name: <name>` - ACL table name

**Optional Variables:**
- `table_type: <L3|L3V6|MIRROR>` - Table type (default: L3)
- `stage: <INGRESS|EGRESS>` - Pipeline stage (default: INGRESS)

**Example:**
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

**CLI Equivalent:**
```bash
config acl add table DATA_ACL L3 -s INGRESS
```

---

### delete_acl_table

Delete an ACL table.

**Required Variables:**
- `command: delete_acl_table`
- `table_name: <name>` - ACL table name

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: delete_acl_table
    table_name: DATA_ACL
```

**CLI Equivalent:**
```bash
config acl remove table DATA_ACL
```

---

### add_acl_rule

Add a rule to an ACL table.

**Required Variables:**
- `command: add_acl_rule`
- `table_name: <name>` - ACL table name
- `rule_name: <name>` - Rule name

**Optional Variables:**
- `priority: <integer>` - Rule priority (default: 100)
- `action: <FORWARD|DROP|REDIRECT>` - Action (default: FORWARD)
- `src_ip: <prefix>` - Source IP prefix
- `dst_ip: <prefix>` - Destination IP prefix
- `protocol: <TCP|UDP|ICMP>` - IP protocol
- `src_port: <integer>` - Source port number
- `dst_port: <integer>` - Destination port number

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: add_acl_rule
    table_name: DATA_ACL
    rule_name: BLOCK_SSH
    priority: 100
    action: DROP
    src_ip: 192.168.100.0/24
    protocol: TCP
    dst_port: 22
```

**CLI Equivalent:**
```bash
config acl add rule DATA_ACL BLOCK_SSH --priority 100 --action DROP \
  --src-ip 192.168.100.0/24 --protocol TCP --dst-port 22
```

---

### show_acl

Display ACL configuration.

**Required Variables:**
- `command: show_acl`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_acl
```

**CLI Equivalent:**
```bash
show acl table
```

---

## Route Management

### add_static_route

Add a static route.

**Required Variables:**
- `command: add_static_route`
- `prefix: <network_prefix>` - Destination network in CIDR notation
- `nexthop: <ip_address>` - Next-hop IP address

**Example:**
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

**CLI Equivalent:**
```bash
config route add prefix 172.16.0.0/16 nexthop 10.0.0.254
```

---

### delete_static_route

Delete a static route.

**Required Variables:**
- `command: delete_static_route`
- `prefix: <network_prefix>` - Destination network in CIDR notation
- `nexthop: <ip_address>` - Next-hop IP address

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: delete_static_route
    prefix: 172.16.0.0/16
    nexthop: 10.0.0.254
```

**CLI Equivalent:**
```bash
config route del prefix 172.16.0.0/16 nexthop 10.0.0.254
```

---

### show_ip_route

Display IP routing table.

**Required Variables:**
- `command: show_ip_route`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_ip_route
```

**CLI Equivalent:**
```bash
show ip route
```

---

## System Configuration

### set_hostname

Set switch hostname.

**Required Variables:**
- `command: set_hostname`
- `hostname: <name>` - New hostname

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: set_hostname
    hostname: dc1-leaf-01
```

**CLI Equivalent:**
```bash
config hostname dc1-leaf-01
```

---

### configure_ntp

Configure NTP server.

**Required Variables:**
- `command: configure_ntp`
- `ntp_server: <server_address>` - NTP server IP or hostname

**Optional Variables:**
- `remove: <true|false>` - Remove server instead of adding (default: false)

**Example (Add):**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_ntp
    ntp_server: 0.pool.ntp.org
```

**Example (Remove):**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: configure_ntp
    ntp_server: 0.pool.ntp.org
    remove: true
```

**CLI Equivalent:**
```bash
# Add
config ntp add 0.pool.ntp.org

# Remove
config ntp del 0.pool.ntp.org
```

---

### save_config

Save running configuration to startup configuration.

**Required Variables:**
- `command: save_config`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

**CLI Equivalent:**
```bash
config save -y
```

---

### show_running_config

Display running configuration.

**Required Variables:**
- `command: show_running_config`

**Example:**
```yaml
- module: networking/sonic
  vars:
    host: 192.168.1.1
    username: admin
    password: YourPaSsWoRd
    command: show_running_config
```

**CLI Equivalent:**
```bash
show runningconfiguration all
```

---

## Notes

- All commands require authentication (`host`, `username`, `password`)
- REST API mode is used by default (`use_rest: true`)
- CLI mode can be used as fallback (`use_rest: false`)
- Always save configuration after making changes using `save_config`
- Some commands may require specific privileges or SONiC versions
