# Networking and Storage Infrastructure Modules

This document provides comprehensive information about OpenFroyo's networking and storage infrastructure management modules.

## Overview

OpenFroyo includes 4 comprehensive modules for managing network switches and enterprise storage arrays:

**Networking (2 modules):**
1. **os10** - Dell OS10 switch management (36 operations)
2. **sonic** - SONiC switch management (35+ operations)

**Storage (2 modules):**
3. **powerstore** - Dell PowerStore block/file storage (40 operations)
4. **powerscale** - Dell PowerScale NAS (45+ operations)

These modules enable complete infrastructure-as-code management of network and storage resources.

---

# NETWORKING MODULES

## Dell OS10 Switch Module

**Purpose:** Complete management of Dell OS10 network switches

**Supported Operations (36 commands):**

### VLAN Management (5)
- **create_vlan** - Create a new VLAN
- **delete_vlan** - Delete an existing VLAN
- **configure_vlan** - Configure VLAN properties
- **show_vlan** - Display VLAN information
- **show_vlan_brief** - Display brief VLAN summary

### Interface Management (8)
- **configure_interface** - Configure interface (access/trunk/LAG mode)
- **shutdown_interface** - Administratively disable interface
- **no_shutdown_interface** - Enable interface
- **set_description** - Set interface description
- **set_mtu** - Configure MTU size
- **set_speed** - Set interface speed
- **show_interface** - Display interface details
- **show_interface_status** - Show interface status summary

### Port Channel/LAG (4)
- **create_port_channel** - Create port channel
- **delete_port_channel** - Delete port channel
- **add_interface_to_lag** - Add interface to LAG
- **remove_interface_from_lag** - Remove interface from LAG

### Layer 3 Configuration (5)
- **configure_ip_interface** - Configure IP address on interface
- **create_vlan_interface** - Create VLAN interface with IP
- **configure_static_route** - Add static route
- **show_ip_interface** - Display IP interface information
- **show_ip_route** - Show IP routing table

### Spanning Tree (3)
- **configure_spanning_tree** - Configure STP mode (RSTP/MSTP/PVST)
- **set_stp_priority** - Set STP priority
- **show_spanning_tree** - Display STP information

### System Configuration (5)
- **set_hostname** - Configure switch hostname
- **configure_ntp** - Configure NTP server
- **configure_dns** - Configure DNS server
- **save_config** - Save running configuration to startup
- **show_running_config** - Display running configuration

### User Management (3)
- **create_user** - Create new user account
- **delete_user** - Delete user account
- **change_password** - Change user password

### ACL Configuration (3)
- **create_acl** - Create access control list
- **delete_acl** - Delete access control list
- **show_acl** - Display ACL configuration

**Compatibility:** Dell OS10 switches (S4048, S5048, Z9100, MX series)

**Transport:** REST API (preferred) or SSH CLI (fallback)

**Binary Size:** 512KB

---

## SONiC Switch Module

**Purpose:** Complete management of SONiC-based network switches

**Supported Operations (35+ commands):**

### VLAN Management (5)
- **create_vlan** - Create a new VLAN
- **delete_vlan** - Delete VLAN
- **add_vlan_member** - Add interface to VLAN
- **remove_vlan_member** - Remove interface from VLAN
- **show_vlan** - Display VLAN configuration

### Interface Management (8)
- **configure_interface** - Configure interface properties
- **set_interface_ip** - Configure IP address on interface
- **set_interface_mtu** - Set interface MTU
- **set_interface_speed** - Configure interface speed
- **shutdown_interface** - Disable interface
- **startup_interface** - Enable interface
- **show_interface** - Display interface details
- **show_interface_status** - Show interface status

### Port Channel/LAG (5)
- **create_portchannel** - Create port channel (PortChannelXXXX)
- **delete_portchannel** - Delete port channel
- **add_portchannel_member** - Add interface to port channel
- **remove_portchannel_member** - Remove interface from port channel
- **show_portchannel** - Display port channel information

### BGP Configuration (6)
- **configure_bgp** - Configure BGP with AS number and router ID
- **add_bgp_neighbor** - Add BGP neighbor
- **remove_bgp_neighbor** - Remove BGP neighbor
- **configure_bgp_network** - Advertise network in BGP
- **show_bgp_summary** - Display BGP summary
- **show_bgp_neighbors** - Show BGP neighbor details

### ACL Configuration (4)
- **create_acl_table** - Create ACL table
- **delete_acl_table** - Delete ACL table
- **add_acl_rule** - Add ACL rule to table
- **show_acl** - Display ACL configuration

### Route Management (3)
- **add_static_route** - Add static route
- **delete_static_route** - Remove static route
- **show_ip_route** - Display IP routing table

### System Configuration (4)
- **set_hostname** - Configure switch hostname
- **configure_ntp** - Configure NTP server
- **save_config** - Save configuration
- **show_running_config** - Display running configuration

**Compatibility:** Any SONiC-based switch (Arista, Dell, Edge-Core, Celestica, etc.)

**Transport:** REST API (RESTCONF) or CLI (config commands)

**Binary Size:** 3.2MB

---

# STORAGE MODULES

## Dell PowerStore Module

**Purpose:** Complete management of Dell PowerStore unified block/file storage arrays

**Supported Operations (40 commands):**

### Volume Management (10)
- **create_volume** - Create new storage volume
- **delete_volume** - Delete volume
- **modify_volume** - Modify volume properties
- **clone_volume** - Clone existing volume
- **resize_volume** - Resize volume (expand only)
- **map_volume** - Map volume to host or host group
- **unmap_volume** - Unmap volume from host
- **show_volume** - Display volume details
- **show_volumes** - List all volumes
- **show_volume_details** - Show detailed volume information

### Host Management (6)
- **create_host** - Create host object
- **delete_host** - Delete host object
- **modify_host** - Modify host properties
- **add_initiator** - Add WWN/IQN to host
- **remove_initiator** - Remove initiator from host
- **show_hosts** - List all hosts

### Host Group Management (4)
- **create_host_group** - Create host group
- **delete_host_group** - Delete host group
- **add_host_to_group** - Add host to group
- **show_host_groups** - List host groups

### Snapshot Management (6)
- **create_snapshot** - Create volume snapshot
- **delete_snapshot** - Delete snapshot
- **restore_snapshot** - Restore volume from snapshot
- **create_snapshot_rule** - Create snapshot schedule/retention rule
- **delete_snapshot_rule** - Delete snapshot rule
- **show_snapshots** - List snapshots

### Performance Metrics (4)
- **show_volume_metrics** - Display volume performance metrics
- **show_appliance_metrics** - Show appliance metrics
- **show_node_metrics** - Display node metrics
- **show_cluster_metrics** - Show cluster performance

### Replication (5)
- **create_replication_session** - Create replication session
- **delete_replication_session** - Delete replication session
- **pause_replication** - Pause active replication
- **resume_replication** - Resume paused replication
- **show_replication_sessions** - List replication sessions

### System Information (5)
- **show_cluster_info** - Display cluster information
- **show_appliances** - List appliances in cluster
- **show_nodes** - Show cluster nodes
- **show_capacity** - Display capacity utilization
- **show_alerts** - Show system alerts

**Compatibility:** Dell PowerStore T, X, and Appliance models

**API Version:** PowerStore REST API v3.0+

**Binary Size:** 10MB

---

## Dell PowerScale Module

**Purpose:** Complete management of Dell PowerScale (Isilon) scale-out NAS systems

**Supported Operations (45+ commands):**

### SMB Share Management (8)
- **create_smb_share** - Create SMB/CIFS share
- **delete_smb_share** - Delete SMB share
- **modify_smb_share** - Modify share properties
- **show_smb_share** - Display share details
- **add_smb_permission** - Add share permissions
- **remove_smb_permission** - Remove share permissions
- **show_smb_sessions** - Display active SMB sessions
- **show_smb_shares** - List all SMB shares

### NFS Export Management (6)
- **create_nfs_export** - Create NFS export
- **delete_nfs_export** - Delete NFS export
- **modify_nfs_export** - Modify export properties
- **show_nfs_export** - Display export details
- **show_nfs_exports** - List all NFS exports
- **reload_nfs_exports** - Reload NFS exports

### Access Zone Management (5)
- **create_access_zone** - Create access zone (multi-tenancy)
- **delete_access_zone** - Delete access zone
- **modify_access_zone** - Modify zone properties
- **show_access_zone** - Display zone details
- **show_access_zones** - List all access zones

### Quota Management (6)
- **create_quota** - Create quota (directory/user/group)
- **delete_quota** - Delete quota
- **modify_quota** - Modify quota thresholds
- **show_quota** - Display quota details
- **show_quotas** - List all quotas
- **show_quota_reports** - Show quota utilization reports

### Snapshot Management (7)
- **create_snapshot** - Create filesystem snapshot
- **delete_snapshot** - Delete snapshot
- **restore_snapshot** - Restore from snapshot
- **create_snapshot_schedule** - Create snapshot schedule
- **delete_snapshot_schedule** - Delete schedule
- **show_snapshot** - Display snapshot details
- **show_snapshots** - List all snapshots

### Cluster Management (6)
- **show_cluster_config** - Display cluster configuration
- **show_cluster_nodes** - List cluster nodes
- **show_cluster_version** - Show OneFS version
- **show_cluster_capacity** - Display capacity utilization
- **show_cluster_performance** - Show cluster performance
- **show_cluster_health** - Display cluster health status

### User/Group Management (4)
- **create_user** - Create local user
- **delete_user** - Delete user
- **create_group** - Create local group
- **delete_group** - Delete group

### SyncIQ Replication (3)
- **create_synciq_policy** - Create replication policy
- **delete_synciq_policy** - Delete policy
- **show_synciq_policies** - List replication policies

**Compatibility:** Dell PowerScale (Isilon) OneFS 8.x and 9.x

**API Version:** PowerScale Platform API (PAPI) v1-v9

**Binary Size:** 3.3MB

---

# COMMON WORKFLOWS

## Network Infrastructure Setup

### Data Center Leaf-Spine (SONiC)

```yaml
# Configure leaf switch with BGP
- module: networking/sonic
  vars:
    host: "{{ leaf_ip }}"
    username: admin
    password: "{{ vault_password }}"
    command: configure_bgp
    bgp_asn: 65001
    router_id: "10.0.0.1"

# Add BGP neighbors (spine switches)
- module: networking/sonic
  vars:
    command: add_bgp_neighbor
    bgp_asn: 65001
    neighbor_ip: "{{ item }}"
    neighbor_asn: 65000
  loop:
    - "10.1.0.1"  # spine1
    - "10.1.0.2"  # spine2

# Configure VLANs for tenants
- module: networking/sonic
  vars:
    command: create_vlan
    vlan_id: "{{ item }}"
  loop: [100, 200, 300]

# Configure VLAN interfaces
- module: networking/sonic
  vars:
    command: set_interface_ip
    interface: "Vlan{{ item.vlan }}"
    ip_address: "{{ item.ip }}"
    prefix_length: 24
  loop:
    - {vlan: 100, ip: "192.168.100.1"}
    - {vlan: 200, ip: "192.168.200.1"}
    - {vlan: 300, ip: "192.168.300.1"}
```

### Access Switch Deployment (Dell OS10)

```yaml
# Configure access VLANs
- module: networking/os10
  vars:
    host: "{{ switch_ip }}"
    username: admin
    password: "{{ vault_password }}"
    command: create_vlan
    vlan_id: "{{ item.id }}"
    vlan_name: "{{ item.name }}"
  loop:
    - {id: 10, name: "DATA"}
    - {id: 20, name: "VOICE"}
    - {id: 30, name: "GUEST"}
    - {id: 99, name: "MANAGEMENT"}

# Configure access ports
- module: networking/os10
  vars:
    command: configure_interface
    interface: "ethernet 1/1/{{ item }}"
    mode: "access"
    access_vlan: 10
  loop: "{{ range(1, 49) }}"

# Configure uplink as trunk
- module: networking/os10
  vars:
    command: configure_interface
    interface: "ethernet 1/1/48"
    mode: "trunk"
    allowed_vlans: "10,20,30,99"
```

---

## Storage Infrastructure Provisioning

### VMware vSphere Storage (PowerStore)

```yaml
# Create datastore volumes
- module: storage/powerstore
  vars:
    powerstore_host: "192.168.1.100"
    powerstore_user: admin
    powerstore_password: "{{ vault_password }}"
    command: create_volume
    volume_name: "vmware-datastore-{{ item }}"
    volume_size: 2000  # 2TB
  loop: [1, 2, 3]

# Create host group for vSphere cluster
- module: storage/powerstore
  vars:
    command: create_host_group
    host_group_name: "vsphere-cluster-prod"

# Add ESXi hosts
- module: storage/powerstore
  vars:
    command: create_host
    host_name: "{{ item.name }}"
    os_type: "ESXi"
    initiators: ["{{ item.iqn }}"]
  loop:
    - {name: "esx01", iqn: "iqn.1998-01.com.vmware:esx01"}
    - {name: "esx02", iqn: "iqn.1998-01.com.vmware:esx02"}
    - {name: "esx03", iqn: "iqn.1998-01.com.vmware:esx03"}

# Add hosts to group
- module: storage/powerstore
  vars:
    command: add_host_to_group
    host_name: "{{ item }}"
    host_group_name: "vsphere-cluster-prod"
  loop: ["esx01", "esx02", "esx03"]

# Map datastores to cluster
- module: storage/powerstore
  vars:
    command: map_volume
    volume_name: "vmware-datastore-{{ item }}"
    host_group_name: "vsphere-cluster-prod"
  loop: [1, 2, 3]

# Create snapshot rules for datastores
- module: storage/powerstore
  vars:
    command: create_snapshot_rule
    snapshot_rule_name: "daily-backup"
    desired_retention: 168  # 7 days
```

### Multi-Tenant File Storage (PowerScale)

```yaml
# Create access zones for tenants
- module: storage/powerscale
  vars:
    powerscale_host: "192.168.1.200"
    powerscale_user: admin
    powerscale_password: "{{ vault_password }}"
    command: create_access_zone
    zone_name: "tenant-{{ item }}"
    zone_path: "/ifs/zones/{{ item }}"
  loop: ["acme", "globex", "initech"]

# Create SMB shares for each tenant
- module: storage/powerscale
  vars:
    command: create_smb_share
    share_name: "{{ item }}-data"
    path: "/ifs/zones/{{ item }}/data"
    access_zone: "tenant-{{ item }}"
    description: "{{ item | upper }} data share"
  loop: ["acme", "globex", "initech"]

# Set quotas for each tenant
- module: storage/powerscale
  vars:
    command: create_quota
    quota_path: "/ifs/zones/{{ item }}"
    quota_type: "directory"
    hard_threshold: 5497558138880  # 5TB
    soft_threshold: 4947802122240  # 4.5TB
  loop: ["acme", "globex", "initech"]

# Create snapshot schedules
- module: storage/powerscale
  vars:
    command: create_snapshot
    snapshot_name: "{{ item }}-daily"
    snapshot_path: "/ifs/zones/{{ item }}"
  loop: ["acme", "globex", "initech"]
```

### Database Storage (PowerStore)

```yaml
# Oracle RAC storage
run:
  # Create volumes
  - module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-{{ item.name }}"
      volume_size: "{{ item.size }}"
      description: "Oracle {{ item.desc }}"
    loop:
      - {name: "data", size: 2000, desc: "data files"}
      - {name: "redo", size: 200, desc: "redo logs"}
      - {name: "archive", size: 1000, desc: "archive logs"}
      - {name: "fra", size: 2000, desc: "fast recovery area"}

  # Create host group for RAC nodes
  - module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "oracle-rac-prod"

  # Add RAC nodes
  - module: storage/powerstore
    vars:
      command: create_host
      host_name: "rac-node-{{ item }}"
      os_type: "Linux"
      initiators: ["{{ hostvars['rac-node-' + item].iscsi_initiator }}"]
    loop: [1, 2, 3]

  # Add nodes to group
  - module: storage/powerstore
    vars:
      command: add_host_to_group
      host_name: "rac-node-{{ item }}"
      host_group_name: "oracle-rac-prod"
    loop: [1, 2, 3]

  # Map all volumes to RAC cluster
  - module: storage/powerstore
    vars:
      command: map_volume
      volume_name: "oracle-{{ item }}"
      host_group_name: "oracle-rac-prod"
    loop: ["data", "redo", "archive", "fra"]
```

---

# DISASTER RECOVERY

## Storage Replication (PowerStore)

```yaml
# Create replication sessions for critical volumes
- module: storage/powerstore
  vars:
    powerstore_host: "192.168.1.100"
    powerstore_user: admin
    powerstore_password: "{{ vault_password }}"
    command: create_replication_session
    volume_name: "{{ item }}"
    remote_system_id: "{{ dr_powerstore_id }}"
    replication_rule: "hourly-replication"
  loop:
    - "prod-database-01"
    - "prod-database-02"
    - "prod-application-data"
```

## File Replication (PowerScale)

```yaml
# Create SyncIQ replication policies
- module: storage/powerscale
  vars:
    powerscale_host: "192.168.1.200"
    powerscale_user: admin
    powerscale_password: "{{ vault_password }}"
    command: create_synciq_policy
    policy_name: "{{ item.name }}"
    source_path: "{{ item.source }}"
    target_host: "dr-powerscale.example.com"
    target_path: "{{ item.target }}"
    schedule: "{{ item.schedule }}"
  loop:
    - name: "critical-hourly"
      source: "/ifs/data/critical"
      target: "/ifs/dr/critical"
      schedule: "every 1 hours"
    - name: "important-daily"
      source: "/ifs/data/important"
      target: "/ifs/dr/important"
      schedule: "every 1 days at 2:00 AM"
```

---

# SECURITY BEST PRACTICES

## Network Security

### Authentication
- Use strong passwords for switch management
- Create dedicated automation user accounts
- Enable RADIUS/TACACS+ for centralized authentication
- Rotate credentials quarterly
- Use privilege level 15 for full access

### Network Isolation
- Place switch management on dedicated VLAN (e.g., VLAN 99)
- Use firewall rules to restrict management access
- Enable HTTPS for REST API access
- Disable HTTP, Telnet, and other insecure protocols
- Use SSH key authentication where possible

### Configuration Security
- Always save configurations after changes
- Backup configurations regularly
- Use version control for network configs
- Implement change approval process
- Test changes in lab environment first

## Storage Security

### Authentication & Authorization
- Use strong, unique passwords for storage arrays
- Create service accounts for automation
- Implement role-based access control (RBAC)
- Enable multi-factor authentication
- Rotate credentials every 90 days

### Data Protection
- Enable encryption at rest
- Use encryption in transit (HTTPS, encrypted SMB)
- Implement snapshot retention policies
- Configure replication to DR site
- Secure erase decommissioned volumes

### Network Security
- Place storage management on dedicated network
- Use separate networks for iSCSI/FC traffic
- Enable CHAP for iSCSI connections
- Use VLANs to isolate storage traffic
- Implement firewall rules for management access

### Access Control
- Use host groups for clustered applications (PowerStore)
- Implement access zones for multi-tenancy (PowerScale)
- Configure Windows ACLs for SMB shares
- Restrict NFS exports to specific networks
- Regular access audits

---

# MONITORING AND METRICS

## Network Monitoring

```yaml
# Monitor switch interfaces
- module: networking/os10
  vars:
    command: show_interface_status

# Check BGP status (SONiC)
- module: networking/sonic
  vars:
    command: show_bgp_summary

# View spanning tree status
- module: networking/os10
  vars:
    command: show_spanning_tree
```

## Storage Monitoring

```yaml
# PowerStore metrics
- module: storage/powerstore
  vars:
    command: show_cluster_metrics

- module: storage/powerstore
  vars:
    command: show_volume_metrics
    volume_name: "prod-db-01"

- module: storage/powerstore
  vars:
    command: show_capacity

# PowerScale metrics
- module: storage/powerscale
  vars:
    command: show_cluster_health

- module: storage/powerscale
  vars:
    command: show_cluster_capacity

- module: storage/powerscale
  vars:
    command: show_cluster_performance
```

---

# TROUBLESHOOTING

## Network Issues

### Connection Problems
- Verify switch management IP and network
- Check firewall rules (HTTPS: 443, SSH: 22)
- Ensure management interface is up
- Verify credentials and privilege level

### BGP Not Establishing (SONiC)
- Check BGP neighbor configuration
- Verify AS numbers match
- Ensure interfaces are up
- Check firewall rules for BGP (TCP 179)
- Review logs: `show_bgp_neighbors`

### Spanning Tree Issues (OS10)
- Verify STP mode matches on all switches
- Check priority values
- Look for topology loops
- Review port states

## Storage Issues

### Volume Mapping Failures (PowerStore)
- Verify host initiators are correct
- Check host OS type setting
- Ensure volume is not already mapped
- Verify host group membership

### SMB Share Access Issues (PowerScale)
- Check share permissions
- Verify access zone configuration
- Review Windows ACLs
- Ensure DNS is resolving correctly
- Check firewall rules (SMB: 445)

### Performance Degradation
- Check capacity utilization
- Review performance metrics
- Look for snapshot overhead
- Verify network bandwidth
- Check for workload spikes

---

# MODULE STATISTICS

## Networking Modules
- **2 modules** (os10, sonic)
- **71+ operations** total
- **~3.7MB WASM binaries**
- **Multi-vendor support** (Dell, SONiC ecosystem)

## Storage Modules
- **2 modules** (powerstore, powerscale)
- **85+ operations** total
- **~13.3MB WASM binaries**
- **Complete Dell storage portfolio**

## Combined Total
- **4 infrastructure modules**
- **156+ operations**
- **~17MB WASM binaries**
- **Production-ready** with comprehensive documentation

---

# GITHUB ISSUE

This work was tracked in: [Issue #17 - Add networking and storage infrastructure modules](https://github.com/piwi3910/openfroyo/issues/17)
