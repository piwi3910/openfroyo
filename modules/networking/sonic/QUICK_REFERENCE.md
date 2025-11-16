# SONiC Module Quick Reference

Fast lookup guide for common SONiC switch management tasks.

## Data Center Leaf-Spine Scenarios

### Scenario 1: Configure Leaf Switch

Complete configuration for a data center leaf switch.

```yaml
# Set hostname
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_hostname
    hostname: dc1-leaf-01

# Configure uplink interfaces to spine
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: Ethernet48
    ip_address: 10.255.1.1
    prefix_length: 31

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: Ethernet52
    ip_address: 10.255.1.3
    prefix_length: 31

# Configure BGP
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: configure_bgp
    bgp_asn: 65001
    router_id: 10.0.0.1

# Add spine neighbors
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: 10.255.1.0
    neighbor_asn: 65100

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: 10.255.1.2
    neighbor_asn: 65100

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 2: VLAN Configuration for Server Access

Configure VLANs for server connectivity.

```yaml
# Create VLANs
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_vlan
    vlan_id: 100

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_vlan
    vlan_id: 200

# Add server ports as untagged members
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_vlan_member
    vlan_id: 100
    interface: Ethernet0
    tagged: false

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_vlan_member
    vlan_id: 200
    interface: Ethernet4
    tagged: false

# Configure VLAN interfaces
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: Vlan100
    ip_address: 192.168.100.1
    prefix_length: 24

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: Vlan200
    ip_address: 192.168.200.1
    prefix_length: 24

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 3: Port Channel (LACP) Configuration

Configure link aggregation for high-bandwidth server connections.

```yaml
# Create port channel
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_portchannel
    portchannel: PortChannel0001

# Add member interfaces
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet0

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet4

# Configure IP on port channel
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_ip
    interface: PortChannel0001
    ip_address: 10.1.1.1
    prefix_length: 30

# Enable interfaces
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: startup_interface
    interface: Ethernet0

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: startup_interface
    interface: Ethernet4

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: startup_interface
    interface: PortChannel0001

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 4: Security - ACL Configuration

Configure ACLs for network segmentation and security.

```yaml
# Create ACL table
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_acl_table
    table_name: SECURITY_ACL
    table_type: L3
    stage: INGRESS

# Block SSH from untrusted network
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_acl_rule
    table_name: SECURITY_ACL
    rule_name: BLOCK_SSH_UNTRUSTED
    priority: 10
    action: DROP
    src_ip: 192.168.100.0/24
    protocol: TCP
    dst_port: 22

# Allow SSH from management network
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_acl_rule
    table_name: SECURITY_ACL
    rule_name: ALLOW_SSH_MGMT
    priority: 5
    action: FORWARD
    src_ip: 10.0.0.0/24
    protocol: TCP
    dst_port: 22

# Block inter-VLAN traffic
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_acl_rule
    table_name: SECURITY_ACL
    rule_name: BLOCK_VLAN_100_TO_200
    priority: 20
    action: DROP
    src_ip: 192.168.100.0/24
    dst_ip: 192.168.200.0/24

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 5: BGP EVPN VXLAN Configuration

Configure BGP for EVPN/VXLAN overlay network.

```yaml
# Configure BGP with loopback as router ID
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: configure_bgp
    bgp_asn: 65001
    router_id: 10.0.0.1

# Add spine neighbors for underlay
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: 10.255.1.0
    neighbor_asn: 65100

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: 10.255.1.2
    neighbor_asn: 65100

# Advertise loopback network
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: configure_bgp_network
    bgp_asn: 65001
    network: 10.0.0.1/32

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 6: Multi-Chassis LAG (MLAG) Configuration

Configure MLAG for dual-homed server connectivity.

```yaml
# Configure peer link port channel
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_portchannel
    portchannel: PortChannel0010

# Add peer link members
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0010
    interface: Ethernet56

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0010
    interface: Ethernet60

# Configure MLAG server-facing port channel
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: create_portchannel
    portchannel: PortChannel0001

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet0

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: add_portchannel_member
    portchannel: PortChannel0001
    interface: Ethernet4

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

### Scenario 7: Jumbo Frames Configuration

Enable jumbo frames for storage network.

```yaml
# Set MTU on storage VLAN interfaces
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_mtu
    interface: Ethernet0
    mtu: 9216

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_mtu
    interface: Ethernet4
    mtu: 9216

- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_mtu
    interface: Ethernet8
    mtu: 9216

# Set MTU on VLAN interface
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: set_interface_mtu
    interface: Vlan300
    mtu: 9216

# Save configuration
- module: networking/sonic
  vars:
    host: 192.168.1.10
    username: admin
    password: YourPaSsWoRd
    command: save_config
```

---

## Quick Command Lookup

### VLAN Operations
```yaml
create_vlan        # Create VLAN
delete_vlan        # Delete VLAN
add_vlan_member    # Add interface to VLAN
remove_vlan_member # Remove interface from VLAN
show_vlan          # Show VLAN info
```

### Interface Operations
```yaml
configure_interface    # Configure multiple settings
set_interface_ip       # Set IP address
set_interface_mtu      # Set MTU
set_interface_speed    # Set speed
shutdown_interface     # Disable interface
startup_interface      # Enable interface
show_interface         # Show interface info
show_interface_status  # Show interface status
```

### Port Channel Operations
```yaml
create_portchannel         # Create port channel
delete_portchannel         # Delete port channel
add_portchannel_member     # Add member
remove_portchannel_member  # Remove member
show_portchannel           # Show port channel info
```

### BGP Operations
```yaml
configure_bgp         # Configure BGP instance
add_bgp_neighbor      # Add neighbor
remove_bgp_neighbor   # Remove neighbor
configure_bgp_network # Advertise network
show_bgp_summary      # Show BGP summary
show_bgp_neighbors    # Show neighbors
```

### ACL Operations
```yaml
create_acl_table # Create ACL table
delete_acl_table # Delete ACL table
add_acl_rule     # Add ACL rule
show_acl         # Show ACL config
```

### Route Operations
```yaml
add_static_route    # Add static route
delete_static_route # Delete static route
show_ip_route       # Show routing table
```

### System Operations
```yaml
set_hostname         # Set hostname
configure_ntp        # Configure NTP
save_config          # Save configuration
show_running_config  # Show running config
```

---

## Common Variable Patterns

### Connection Variables (Always Required)
```yaml
host: 192.168.1.1
username: admin
password: YourPaSsWoRd
```

### VLAN Variables
```yaml
vlan_id: 100
interface: Ethernet0
tagged: false
```

### Interface Variables
```yaml
interface: Ethernet0
ip_address: 10.0.0.1
prefix_length: 24
mtu: 9000
speed: 100000
```

### BGP Variables
```yaml
bgp_asn: 65000
router_id: 10.0.0.1
neighbor_ip: 10.0.0.2
neighbor_asn: 65001
network: 192.168.0.0/16
```

### ACL Variables
```yaml
table_name: DATA_ACL
table_type: L3
stage: INGRESS
rule_name: RULE_1
priority: 100
action: DROP
src_ip: 192.168.100.0/24
dst_ip: 10.0.0.0/8
protocol: TCP
dst_port: 22
```

---

## Best Practices

1. **Always save configuration** after making changes
2. **Use descriptive names** for port channels, ACLs, and rules
3. **Document changes** in task names for audit trail
4. **Test in non-production** before deploying to production
5. **Use consistent naming conventions** across all switches
6. **Implement ACLs** for security segmentation
7. **Configure NTP** for accurate time synchronization
8. **Use jumbo frames** for storage and high-throughput networks
9. **Enable LACP** for port channels to ensure proper failover
10. **Monitor BGP neighbors** to ensure routing stability

---

## Troubleshooting Quick Reference

### Check Interface Status
```yaml
command: show_interface_status
```

### Verify VLAN Configuration
```yaml
command: show_vlan
```

### Check BGP Peering
```yaml
command: show_bgp_summary
command: show_bgp_neighbors
```

### Verify Port Channel Status
```yaml
command: show_portchannel
```

### Check Routing Table
```yaml
command: show_ip_route
```

### View ACL Configuration
```yaml
command: show_acl
```

### Display Running Config
```yaml
command: show_running_config
```
