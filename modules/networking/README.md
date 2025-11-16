# OpenFroyo Networking Modules

This directory contains modules for managing network switches and related infrastructure.

## Available Modules (2 Total)

### 1. os10 (Dell OS10 Switch Management)
**Purpose:** Complete management of Dell OS10 network switches

**Key Features:**
- 36 comprehensive commands
- VLAN management
- Interface configuration (access, trunk, LAG)
- Port channel/LAG management
- Layer 3 routing (static routes, VLAN interfaces)
- Spanning Tree Protocol (STP)
- System configuration
- User management
- Access Control Lists (ACLs)

**Compatibility:** Dell OS10 switches (S4048, S5048, Z9100, etc.)

**Size:** 512KB

**Transport Methods:**
- REST API (preferred) - Dell OS10 REST API
- SSH CLI (fallback) - via sshpass

---

### 2. sonic (SONiC Switch Management)
**Purpose:** Complete management of SONiC network switches

**Key Features:**
- 35+ operations
- VLAN management and membership
- Interface configuration (IP, MTU, speed)
- Port channel/LAG management
- BGP configuration and neighbor management
- ACL tables and rules
- Static routing
- System configuration

**Compatibility:** SONiC switches (any vendor: Arista, Dell, Edge-Core, etc.)

**Size:** 3.2MB

**Transport Methods:**
- REST API (preferred) - SONiC RESTCONF API
- CLI (fallback) - SONiC configuration commands

---

## Comparison Matrix

| Feature | os10 | sonic |
|---------|------|-------|
| VLAN Management | ✅ (5 ops) | ✅ (5 ops) |
| Interface Management | ✅ (8 ops) | ✅ (8 ops) |
| Port Channels/LAG | ✅ (4 ops) | ✅ (5 ops) |
| Layer 3 Routing | ✅ (5 ops) | ✅ (3 ops) |
| BGP Configuration | ❌ | ✅ (6 ops) |
| Spanning Tree | ✅ (3 ops) | ❌ |
| ACLs | ✅ (3 ops) | ✅ (4 ops) |
| System Config | ✅ (5 ops) | ✅ (4 ops) |
| User Management | ✅ (3 ops) | ❌ |
| REST API | ✅ | ✅ |
| CLI Fallback | ✅ | ✅ |

## When to Use Each Module

### Use **os10** when:
- You have Dell OS10 switches in your environment
- You need traditional L2/L3 switching features
- You require Spanning Tree Protocol configuration
- You want user management on the switch
- You need enterprise switch features

### Use **sonic** when:
- You have SONiC-based switches (any vendor)
- You're building a data center fabric
- You need BGP for routing (leaf-spine architecture)
- You want open-source network OS features
- You're implementing EVPN/VXLAN overlays
- You need multi-vendor switch management

## Architecture

All modules follow the OpenFroyo shell_exec pattern:

1. **REST API (preferred):** Use native switch REST APIs
   - Dell OS10: `/restconf/data/` endpoints
   - SONiC: `/restconf/data/` endpoints (openconfig models)

2. **CLI Fallback:** Generate configuration commands
   - Dell OS10: Native OS10 CLI commands
   - SONiC: `config` and `show` commands

3. **Authentication:** HTTP Basic Auth for REST, SSH for CLI

## Common Usage Patterns

### Basic VLAN Configuration

**Dell OS10:**
```yaml
- module: networking/os10
  vars:
    host: "192.168.1.10"
    username: "admin"
    password: "admin123"
    command: create_vlan
    vlan_id: 100
    vlan_name: "PRODUCTION"
```

**SONiC:**
```yaml
- module: networking/sonic
  vars:
    host: "192.168.1.20"
    username: "admin"
    password: "YourPaSsWoRd"
    command: create_vlan
    vlan_id: 100
```

### Interface Configuration

**Dell OS10 - Trunk Port:**
```yaml
- module: networking/os10
  vars:
    host: "192.168.1.10"
    username: "admin"
    password: "admin123"
    command: configure_interface
    interface: "ethernet 1/1/1"
    mode: "trunk"
    allowed_vlans: "10,20,30"
```

**SONiC - Interface IP:**
```yaml
- module: networking/sonic
  vars:
    host: "192.168.1.20"
    username: "admin"
    password: "YourPaSsWoRd"
    command: set_interface_ip
    interface: "Ethernet0"
    ip_address: "192.168.1.1"
    prefix_length: 24
```

### Port Channel/LAG

**Dell OS10:**
```yaml
- module: networking/os10
  vars:
    command: create_port_channel
    port_channel_id: 1
    mode: "static"

- module: networking/os10
  vars:
    command: add_interface_to_lag
    interface: "ethernet 1/1/1"
    port_channel_id: 1
```

**SONiC:**
```yaml
- module: networking/sonic
  vars:
    command: create_portchannel
    portchannel: "PortChannel0001"

- module: networking/sonic
  vars:
    command: add_portchannel_member
    portchannel: "PortChannel0001"
    interface: "Ethernet0"
```

### BGP Configuration (SONiC)

```yaml
# Configure BGP
- module: networking/sonic
  vars:
    command: configure_bgp
    bgp_asn: 65001
    router_id: "192.168.1.1"

# Add BGP neighbor
- module: networking/sonic
  vars:
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: "192.168.1.2"
    neighbor_asn: 65002
```

## Data Center Scenarios

### Leaf-Spine Architecture (SONiC)

```yaml
# Configure leaf switch
- module: networking/sonic
  vars:
    host: "{{ leaf_switch }}"
    command: configure_bgp
    bgp_asn: "{{ leaf_asn }}"
    router_id: "{{ leaf_loopback }}"

- module: networking/sonic
  vars:
    command: add_bgp_neighbor
    neighbor_ip: "{{ spine1_ip }}"
    neighbor_asn: "{{ spine_asn }}"

- module: networking/sonic
  vars:
    command: add_bgp_neighbor
    neighbor_ip: "{{ spine2_ip }}"
    neighbor_asn: "{{ spine_asn }}"
```

### Access Switch Configuration (Dell OS10)

```yaml
# Configure access VLANs
- module: networking/os10
  vars:
    command: create_vlan
    vlan_id: "{{ item }}"
  loop: [10, 20, 30, 40]

# Configure access ports
- module: networking/os10
  vars:
    command: configure_interface
    interface: "ethernet 1/1/{{ item }}"
    mode: "access"
    access_vlan: 10
  loop: [1, 2, 3, 4]

# Configure uplink trunk
- module: networking/os10
  vars:
    command: configure_interface
    interface: "ethernet 1/1/48"
    mode: "trunk"
    allowed_vlans: "10,20,30,40"
```

## Security Considerations

### Authentication
- Use strong passwords for switch management
- Enable certificate validation in production (`validate_certs: true`)
- Use dedicated automation user accounts with minimal privileges
- Rotate credentials regularly

### Network Security
- Place switch management on dedicated management VLAN
- Use firewall rules to restrict management access
- Enable HTTPS for REST API (disable HTTP)
- Use SSH key authentication where possible

### Configuration Management
- Always save configurations after changes (`save_config` command)
- Backup configurations regularly
- Use version control for network configurations
- Test changes in lab environment first

## Performance Considerations

### REST API
- Faster than CLI for most operations
- Supports parallel operations
- Better structured responses (JSON)
- Recommended for automation

### CLI Fallback
- Use when REST API unavailable
- Some advanced features may require CLI
- Slower for bulk operations
- Good for legacy switch support

### Rate Limiting
- Most switches limit API requests (100-1000 req/min)
- Implement delays between bulk operations
- Monitor switch CPU usage during automation

## Troubleshooting

### Connection Issues
- Verify switch IP and network connectivity
- Check management VLAN configuration
- Verify firewall rules (HTTPS: 443, SSH: 22)
- Ensure switch management interface is up

### Authentication Failures
- Verify username and password
- Check user privilege level (need admin/15)
- Verify account is not locked
- Check RADIUS/TACACS if using AAA

### Command Failures
- Verify switch model compatibility
- Check firmware version requirements
- Review switch logs for errors
- Validate command syntax in documentation

### REST API Issues
- Verify REST API is enabled on switch
- Check certificate validation settings
- Review API version compatibility
- Use CLI fallback if REST unavailable

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **COMMANDS.md** - Detailed command reference
- **QUICK_REFERENCE.md** - Quick command lookup
- **test.ofy** - Example test stack

## Module Statistics

- **2 networking modules**
- **71+ total operations**
- **~3.7MB total WASM binaries** (source only, binaries in .gitignore)
- **Multi-vendor support** (Dell, SONiC ecosystem)
- **Production-ready** with comprehensive documentation

## Future Enhancements

Potential additions:
- Cisco IOS/IOS-XE module
- Arista EOS module
- Juniper Junos module
- Cumulus Linux module
- NETCONF/YANG support
- Network device discovery
- Configuration diff/rollback
