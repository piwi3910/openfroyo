# OpenFroyo Storage Modules

This directory contains modules for managing enterprise storage arrays and NAS systems.

## Available Modules (2 Total)

### 1. powerstore (Dell PowerStore Storage Array)
**Purpose:** Complete management of Dell PowerStore block and file storage arrays

**Key Features:**
- 40 comprehensive operations
- Block storage (iSCSI, FC) and file storage (NFS, SMB)
- Volume management (create, delete, modify, clone, resize, map)
- Host and host group management
- Snapshot management with retention rules
- Performance metrics collection
- Replication session management
- Cluster and system information

**Compatibility:** Dell PowerStore T, X, and Appliance models

**Size:** 10MB

**API:** Dell PowerStore REST API v3.0+

---

### 2. powerscale (Dell PowerScale NAS)
**Purpose:** Complete management of Dell PowerScale (Isilon) scale-out NAS systems

**Key Features:**
- 45+ operations
- SMB share management with permissions
- NFS export management with client access control
- Access zone multi-tenancy
- Quota management (directory, user, group)
- Snapshot management with schedules
- SyncIQ replication for disaster recovery
- Cluster management and monitoring
- User and group management

**Compatibility:** Dell PowerScale (Isilon) running OneFS 8.x and 9.x

**Size:** 3.3MB

**API:** Dell PowerScale Platform API (PAPI)

---

## Comparison Matrix

| Feature | PowerStore | PowerScale |
|---------|------------|------------|
| **Storage Type** | Block + File | File (NAS) |
| Volume Management | ✅ (10 ops) | ❌ |
| Host Management | ✅ (6 ops) | ❌ |
| SMB Shares | ✅ | ✅ (8 ops) |
| NFS Exports | ✅ | ✅ (6 ops) |
| Snapshots | ✅ (6 ops) | ✅ (7 ops) |
| Replication | ✅ (5 ops) | ✅ (3 ops - SyncIQ) |
| Multi-tenancy | ❌ | ✅ (Access Zones) |
| Quotas | ❌ | ✅ (6 ops) |
| Performance Metrics | ✅ (4 ops) | ✅ (in cluster ops) |
| Cluster Management | ✅ (5 ops) | ✅ (6 ops) |

## When to Use Each Module

### Use **powerstore** when:
- You need block storage for VMs or databases
- You require SAN connectivity (iSCSI, Fibre Channel)
- You want unified block and file storage
- You need volume cloning and snapshots
- You're managing ESXi datastores
- You need storage replication
- You have Dell PowerStore arrays

### Use **powerscale** when:
- You need scale-out NAS storage
- You require multi-protocol file sharing (SMB + NFS)
- You need multi-tenant file storage (Access Zones)
- You want directory/user/group quotas
- You need large-scale file system (petabyte-scale)
- You require disaster recovery replication (SyncIQ)
- You have Dell PowerScale (Isilon) clusters

## Architecture

All modules follow the OpenFroyo shell_exec pattern:

1. **REST API:** Use native storage REST APIs
   - PowerStore: `https://<ip>/api/rest/` (v3.0+)
   - PowerScale: `https://<ip>:8080/platform/` (PAPI)

2. **Authentication:** HTTP Basic Auth
3. **Response Format:** JSON
4. **Error Handling:** HTTP status codes + JSON error messages

## Common Usage Patterns

### Volume Provisioning (PowerStore)

```yaml
# Create volume
- module: storage/powerstore
  vars:
    powerstore_host: "192.168.1.100"
    powerstore_user: "admin"
    powerstore_password: "Password123!"
    command: create_volume
    volume_name: "prod-db-01"
    volume_size: 500  # GB
    description: "Production database volume"

# Create host
- module: storage/powerstore
  vars:
    command: create_host
    host_name: "esx-host-01"
    os_type: "ESXi"
    initiators:
      - "iqn.1998-01.com.vmware:esx-host-01"

# Map volume to host
- module: storage/powerstore
  vars:
    command: map_volume
    volume_name: "prod-db-01"
    host_name: "esx-host-01"
```

### File Share Creation (PowerScale)

```yaml
# Create SMB share
- module: storage/powerscale
  vars:
    powerscale_host: "192.168.1.200"
    powerscale_user: "admin"
    powerscale_password: "admin123"
    command: create_smb_share
    share_name: "engineering"
    path: "/ifs/data/engineering"
    description: "Engineering Department Share"

# Create quota
- module: storage/powerscale
  vars:
    command: create_quota
    quota_path: "/ifs/data/engineering"
    quota_type: "directory"
    hard_threshold: 1099511627776  # 1TB

# Create snapshot schedule
- module: storage/powerscale
  vars:
    command: create_snapshot
    snapshot_name: "eng-daily"
    snapshot_path: "/ifs/data/engineering"
```

### Snapshot Management

**PowerStore:**
```yaml
# Create snapshot
- module: storage/powerstore
  vars:
    command: create_snapshot
    volume_name: "prod-db-01"
    snapshot_name: "before-upgrade"
    description: "Pre-upgrade backup"

# Create snapshot rule (schedule)
- module: storage/powerstore
  vars:
    command: create_snapshot_rule
    snapshot_rule_name: "daily-backup"
    desired_retention: 168  # 7 days in hours
```

**PowerScale:**
```yaml
# Create snapshot
- module: storage/powerscale
  vars:
    command: create_snapshot
    snapshot_name: "daily-2024-11-16"
    snapshot_path: "/ifs/data"
    snapshot_alias: "daily"
```

### Replication Configuration

**PowerStore (Block Replication):**
```yaml
- module: storage/powerstore
  vars:
    command: create_replication_session
    volume_name: "prod-db-01"
    remote_system_id: "{{ dr_system_id }}"
    replication_rule: "hourly-replication"
```

**PowerScale (SyncIQ File Replication):**
```yaml
- module: storage/powerscale
  vars:
    command: create_synciq_policy
    policy_name: "dr-replication"
    source_path: "/ifs/data/critical"
    target_path: "/ifs/data/critical-replica"
    target_host: "dr-cluster.example.com"
    schedule: "every 1 hours"
```

## Enterprise Workflows

### VMware vSphere Integration (PowerStore)

```yaml
run:
  # Create datastore volume
  - module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vmware-datastore-01"
      volume_size: 2000  # 2TB

  # Create host group for cluster
  - module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "vmware-cluster-01"

  # Add ESXi hosts to group
  - module: storage/powerstore
    vars:
      command: add_host_to_group
      host_name: "{{ item }}"
      host_group_name: "vmware-cluster-01"
    loop:
      - "esx-host-01"
      - "esx-host-02"
      - "esx-host-03"

  # Map volume to host group
  - module: storage/powerstore
    vars:
      command: map_volume
      volume_name: "vmware-datastore-01"
      host_group_name: "vmware-cluster-01"
```

### Multi-Tenant File Sharing (PowerScale)

```yaml
run:
  # Create access zone for tenant
  - module: storage/powerscale
    vars:
      command: create_access_zone
      zone_name: "tenant-acme"
      zone_path: "/ifs/zones/acme"

  # Create SMB share in zone
  - module: storage/powerscale
    vars:
      command: create_smb_share
      share_name: "acme-data"
      path: "/ifs/zones/acme/data"
      access_zone: "tenant-acme"

  # Set quota for tenant
  - module: storage/powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/zones/acme"
      quota_type: "directory"
      hard_threshold: 5497558138880  # 5TB
```

### Database Storage Provisioning (PowerStore)

```yaml
# Oracle RAC shared storage
run:
  # Data volume
  - module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-data"
      volume_size: 1000

  # Redo logs volume (high performance)
  - module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-redo"
      volume_size: 100

  # Archive logs volume
  - module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-archive"
      volume_size: 500

  # Create host group for RAC nodes
  - module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "oracle-rac-cluster"

  # Map volumes to cluster
  - module: storage/powerstore
    vars:
      command: map_volume
      volume_name: "{{ item }}"
      host_group_name: "oracle-rac-cluster"
    loop:
      - "oracle-data"
      - "oracle-redo"
      - "oracle-archive"
```

### Backup and Recovery (PowerScale)

```yaml
# 4-tier snapshot retention
run:
  # Hourly snapshots (24 hours)
  - module: storage/powerscale
    vars:
      command: create_snapshot
      snapshot_name: "hourly-{{ ansible_date_time.hour }}"
      snapshot_path: "/ifs/data"

  # Daily snapshots (7 days)
  - module: storage/powerscale
    vars:
      command: create_snapshot
      snapshot_name: "daily-{{ ansible_date_time.date }}"
      snapshot_path: "/ifs/data"

  # Replicate to DR site
  - module: storage/powerscale
    vars:
      command: create_synciq_policy
      policy_name: "backup-replication"
      source_path: "/ifs/data"
      target_host: "dr-powerscale.example.com"
      target_path: "/ifs/backups/prod"
      schedule: "every 6 hours"
```

## Monitoring and Metrics

### PowerStore Performance Monitoring

```yaml
# Volume performance
- module: storage/powerstore
  vars:
    command: show_volume_metrics
    volume_name: "prod-db-01"

# Appliance health
- module: storage/powerstore
  vars:
    command: show_appliance_metrics

# Cluster capacity
- module: storage/powerstore
  vars:
    command: show_capacity
```

### PowerScale Cluster Monitoring

```yaml
# Cluster information
- module: storage/powerscale
  vars:
    command: show_cluster_info

# Cluster health
- module: storage/powerscale
  vars:
    command: show_cluster_health

# Capacity utilization
- module: storage/powerscale
  vars:
    command: show_cluster_capacity
```

## Security Best Practices

### Authentication
- Use strong, unique passwords for storage arrays
- Rotate credentials regularly (quarterly)
- Create dedicated automation user accounts
- Use role-based access control (RBAC)
- Enable multi-factor authentication where available

### Network Security
- Place storage management on dedicated VLAN
- Use firewall rules to restrict management access
- Enable HTTPS with valid certificates
- Disable HTTP management access
- Use VPN for remote management

### Data Protection
- Enable encryption at rest
- Use secure erase for decommissioned volumes
- Implement snapshot retention policies
- Configure replication to DR site
- Regular backup verification

### Access Control
- Use host groups for multi-host access (PowerStore)
- Implement access zones for multi-tenancy (PowerScale)
- Configure CHAP for iSCSI connections
- Use Windows ACLs for SMB shares
- Configure NFS export restrictions

## Performance Considerations

### PowerStore
- Use SSDs for high-performance workloads
- Configure host groups for clustered applications
- Enable compression for space efficiency
- Monitor IOPs and latency metrics
- Size volumes appropriately

### PowerScale
- Use SSD pools for metadata and hot data
- Configure SmartPools for tiered storage
- Enable L3 cache for read-intensive workloads
- Monitor client connections and throughput
- Use multiple access zones for isolation

## Capacity Management

### PowerStore
```yaml
# Check cluster capacity
- module: storage/powerstore
  vars:
    command: show_capacity

# Monitor volume usage
- module: storage/powerstore
  vars:
    command: show_volume_details
    volume_name: "{{ volume }}"
```

### PowerScale
```yaml
# Check cluster capacity
- module: storage/powerscale
  vars:
    command: show_cluster_capacity

# Review quota reports
- module: storage/powerscale
  vars:
    command: show_quota_reports
```

## Troubleshooting

### Connection Issues
- Verify storage management IP and network
- Check firewall rules (HTTPS: 443, PowerScale: 8080)
- Ensure management interface is configured
- Verify DNS resolution if using hostnames

### Authentication Failures
- Verify username and password
- Check user account status (not locked)
- Verify user has appropriate role/privileges
- Check for expired passwords

### REST API Errors
- Review HTTP status codes (401, 403, 404, 500)
- Check API version compatibility
- Verify request format (JSON)
- Review storage array logs

### Performance Issues
- Monitor array performance metrics
- Check for capacity constraints
- Verify network bandwidth
- Review workload patterns
- Check for snapshot overhead

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **API_REFERENCE.md** - REST API endpoint reference
- **WORKFLOWS.md** - Common storage workflows
- **test.ofy** - Example test stack

## Module Statistics

- **2 storage modules**
- **85+ total operations**
- **~13.3MB total WASM binaries** (source only, binaries in .gitignore)
- **Complete Dell storage portfolio** (Block + File)
- **Production-ready** with comprehensive documentation
- **Enterprise features** (replication, multi-tenancy, snapshots)

## Future Enhancements

Potential additions:
- Dell Unity module
- Pure Storage FlashArray module
- NetApp ONTAP module
- HPE Nimble module
- Automated storage tiering
- Capacity forecasting
- Performance analytics
