# Dell PowerScale (Isilon) NAS Management Module

## Overview

The PowerScale module provides comprehensive management capabilities for Dell PowerScale (formerly EMC Isilon) NAS systems through the Platform API (PAPI). This module enables automation of file sharing, quota management, snapshots, access zones, replication, and cluster operations.

## Architecture

### Dell PowerScale/Isilon Overview

Dell PowerScale is a scale-out NAS platform designed for unstructured data storage. Key architectural components:

- **OneFS Operating System**: Distributed file system that runs on every node
- **Scale-Out Architecture**: Linear performance scaling by adding nodes
- **Single Namespace**: All storage presented as a unified file system under `/ifs`
- **Access Zones**: Logical partitions for multi-tenancy and authentication isolation
- **SmartPools**: Automated tiering across node types (SSDs, HDDs, archive)
- **SyncIQ**: Asynchronous replication for disaster recovery
- **SnapshotIQ**: Point-in-time snapshots with minimal space overhead

### Module Architecture

This module communicates with PowerScale clusters via the Platform API (PAPI):

```
┌─────────────────┐
│  OpenFroyo CLI  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ PowerScale WASM │
│     Module      │
└────────┬────────┘
         │ HTTPS (Port 8080)
         ▼
┌─────────────────┐
│  PowerScale API │
│   (PAPI v9+)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ PowerScale      │
│ Cluster (OneFS) │
└─────────────────┘
```

## Features

### 45+ Operations Across 9 Categories

1. **SMB Share Management** (8 operations)
2. **NFS Export Management** (6 operations)
3. **Access Zone Management** (5 operations)
4. **Quota Management** (6 operations)
5. **Snapshot Management** (7 operations)
6. **Cluster Management** (6 operations)
7. **User/Group Management** (4 operations)
8. **SyncIQ Replication** (3 operations)

## Installation

This module is included in the OpenFroyo `modules/storage/powerscale/` directory.

### Prerequisites

- Dell PowerScale cluster running OneFS 8.x or 9.x
- Network connectivity to PowerScale on port 8080 (HTTPS)
- Admin credentials for the PowerScale cluster
- PowerScale Platform API (PAPI) enabled

### Building the Module

```bash
cd modules/storage/powerscale/wasm
make build
```

This creates `powerscale.wasm` which is referenced by the module definition.

## Quick Start

### Basic Configuration

```yaml
# In your stack file
defaults:
  powerscale_host: "192.168.1.100"     # PowerScale cluster IP
  powerscale_user: "admin"              # Admin username
  powerscale_password: "SecurePass123"  # Admin password
  powerscale_port: 8080                 # API port (default: 8080)
  validate_certs: false                 # Set true for production
```

### Example: Create SMB Share

```yaml
run:
  - name: "Create department share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "engineering"
      path: "/ifs/data/engineering"
      description: "Engineering department share"
      browsable: true
      ntfs_acl_support: true
```

### Example: Create NFS Export

```yaml
run:
  - name: "Create NFS export for backup"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/backup"
      clients:
        - "192.168.1.0/24"
        - "10.0.0.50"
      read_only: false
      root_clients:
        - "192.168.1.10"
```

### Example: Create Quota

```yaml
run:
  - name: "Set 1TB quota on share"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/engineering"
      quota_type: "directory"
      hard_threshold: 1099511627776  # 1TB in bytes
      soft_threshold: 966367641600   # 900GB
      enforce: true
```

### Example: Create Snapshot

```yaml
run:
  - name: "Create daily snapshot"
    module: storage.powerscale
    vars:
      command: create_snapshot
      snapshot_path: "/ifs/data"
      snapshot_name: "daily_backup"
      snapshot_alias: "latest"
```

## Command Reference

### SMB Share Management

#### create_smb_share
Creates a new SMB (CIFS) share.

**Required Parameters:**
- `share_name`: Share name
- `path`: File system path (e.g., `/ifs/data/share`)

**Optional Parameters:**
- `access_zone`: Access zone (default: "System")
- `description`: Share description
- `browsable`: Show in browse list (default: true)
- `ntfs_acl_support`: Enable NTFS ACLs (default: true)
- `permissions`: Permission objects array

**Example:**
```yaml
- name: "Create sales share"
  module: storage.powerscale
  vars:
    command: create_smb_share
    share_name: "sales"
    path: "/ifs/data/sales"
    description: "Sales department data"
    browsable: true
```

#### delete_smb_share
Deletes an existing SMB share.

**Required Parameters:**
- `share_name`: Share name to delete

**Example:**
```yaml
- name: "Remove old share"
  module: storage.powerscale
  vars:
    command: delete_smb_share
    share_name: "old_share"
```

#### modify_smb_share
Modifies an existing SMB share.

**Required Parameters:**
- `share_name`: Share name to modify

**Optional Parameters:**
- `description`: New description
- `browsable`: Browsable setting
- `permissions`: Updated permissions

#### show_smb_share
Retrieves details of a specific SMB share.

**Required Parameters:**
- `share_name`: Share name

#### show_smb_shares
Lists all SMB shares in the specified access zone.

**Optional Parameters:**
- `access_zone`: Access zone (default: "System")

#### add_smb_permission
Adds permissions to an SMB share.

**Required Parameters:**
- `share_name`: Share name
- `permissions`: Permission objects

#### remove_smb_permission
Removes permissions from an SMB share.

**Required Parameters:**
- `share_name`: Share name
- `permissions`: Permission objects to remove

#### show_smb_sessions
Shows active SMB sessions.

### NFS Export Management

#### create_nfs_export
Creates a new NFS export.

**Required Parameters:**
- `export_path` or `export_paths`: Path(s) to export

**Optional Parameters:**
- `clients`: Client IP addresses/networks
- `read_only`: Read-only export (default: false)
- `root_clients`: Clients with root access
- `security_flavors`: Security types (e.g., ["unix", "krb5"])
- `map_root`: Map root user
- `map_all`: Map all users

**Example:**
```yaml
- name: "Create production NFS export"
  module: storage.powerscale
  vars:
    command: create_nfs_export
    export_paths:
      - "/ifs/data/prod"
    clients:
      - "192.168.10.0/24"
    read_only: false
    root_clients:
      - "192.168.10.100"
    security_flavors:
      - "unix"
```

#### delete_nfs_export
Deletes an NFS export.

**Required Parameters:**
- `export_id`: Export ID to delete

#### modify_nfs_export
Modifies an existing NFS export.

**Required Parameters:**
- `export_id`: Export ID to modify

**Optional Parameters:**
- `clients`: Updated client list
- `read_only`: Read-only setting
- `root_clients`: Updated root clients

#### show_nfs_export
Shows details of a specific NFS export.

**Required Parameters:**
- `export_id`: Export ID

#### show_nfs_exports
Lists all NFS exports.

**Optional Parameters:**
- `access_zone`: Access zone filter

#### reload_nfs_exports
Reloads NFS export configuration.

### Access Zone Management

#### create_access_zone
Creates a new access zone for multi-tenancy.

**Required Parameters:**
- `zone_name`: Zone name
- `path`: File system path for the zone

**Optional Parameters:**
- `groupnet`: Groupnet name (default: "groupnet0")
- `auth_providers`: Authentication providers

**Example:**
```yaml
- name: "Create production zone"
  module: storage.powerscale
  vars:
    command: create_access_zone
    zone_name: "production"
    path: "/ifs/zones/production"
    groupnet: "groupnet0"
```

#### delete_access_zone
Deletes an access zone.

**Required Parameters:**
- `zone_name`: Zone name to delete

#### modify_access_zone
Modifies an access zone.

**Required Parameters:**
- `zone_name`: Zone name

**Optional Parameters:**
- `auth_providers`: Updated auth providers

#### show_access_zone
Shows details of a specific access zone.

**Required Parameters:**
- `zone_name`: Zone name

#### show_access_zones
Lists all access zones.

### Quota Management

#### create_quota
Creates a storage quota.

**Required Parameters:**
- `quota_path`: Path for quota
- `quota_type`: Type (directory, user, group, default-user, default-group)

**Optional Parameters:**
- `hard_threshold`: Hard limit in bytes
- `soft_threshold`: Soft limit in bytes
- `advisory_threshold`: Advisory threshold in bytes
- `quota_user`: User name (for user quotas)
- `quota_group`: Group name (for group quotas)
- `enforce`: Enforce quota (default: true)

**Example:**
```yaml
- name: "Create user quota"
  module: storage.powerscale
  vars:
    command: create_quota
    quota_path: "/ifs/home"
    quota_type: "user"
    quota_user: "jdoe"
    hard_threshold: 107374182400  # 100GB
    soft_threshold: 96636764160   # 90GB
    enforce: true
```

#### delete_quota
Deletes a quota.

**Required Parameters:**
- `quota_id`: Quota ID to delete

#### modify_quota
Modifies an existing quota.

**Required Parameters:**
- `quota_id`: Quota ID

**Optional Parameters:**
- `hard_threshold`: New hard limit
- `soft_threshold`: New soft limit
- `enforce`: Enforce setting

#### show_quota
Shows details of a specific quota.

**Required Parameters:**
- `quota_id`: Quota ID

#### show_quotas
Lists all quotas.

#### show_quota_reports
Shows quota usage reports.

### Snapshot Management

#### create_snapshot
Creates a point-in-time snapshot.

**Required Parameters:**
- `snapshot_path`: Path to snapshot

**Optional Parameters:**
- `snapshot_name`: Snapshot name
- `snapshot_alias`: Snapshot alias

**Example:**
```yaml
- name: "Create pre-maintenance snapshot"
  module: storage.powerscale
  vars:
    command: create_snapshot
    snapshot_path: "/ifs/data/critical"
    snapshot_name: "pre_maint_20250116"
    snapshot_alias: "before_upgrade"
```

#### delete_snapshot
Deletes a snapshot.

**Required Parameters:**
- `snapshot_id` or `snapshot_name`: Snapshot identifier

#### restore_snapshot
Restores files from a snapshot (provides guidance).

**Required Parameters:**
- `snapshot_id`: Snapshot ID

#### create_snapshot_schedule
Creates a snapshot schedule.

**Required Parameters:**
- `schedule_name`: Schedule name
- `snapshot_path`: Path to snapshot
- `schedule`: Cron expression

**Optional Parameters:**
- `pattern`: Naming pattern
- `duration`: Snapshot duration
- `retention`: Retention period

**Example:**
```yaml
- name: "Create daily snapshot schedule"
  module: storage.powerscale
  vars:
    command: create_snapshot_schedule
    schedule_name: "daily_backup"
    snapshot_path: "/ifs/data"
    schedule: "0 2 * * *"  # 2 AM daily
    pattern: "daily-%Y%m%d"
    retention: "30 days"
```

#### delete_snapshot_schedule
Deletes a snapshot schedule.

**Required Parameters:**
- `schedule_name`: Schedule name

#### show_snapshot
Shows details of a specific snapshot.

**Required Parameters:**
- `snapshot_id`: Snapshot ID

#### show_snapshots
Lists all snapshots.

### Cluster Management

#### show_cluster_config
Shows cluster configuration.

#### show_cluster_nodes
Shows all cluster nodes.

#### show_cluster_version
Shows OneFS version information.

#### show_cluster_capacity
Shows cluster capacity statistics.

#### show_cluster_performance
Shows cluster performance metrics.

#### show_cluster_health
Shows cluster health status.

### User/Group Management

#### create_user
Creates a user account.

**Required Parameters:**
- `user_name`: Username

**Optional Parameters:**
- `user_id`: UID
- `primary_group`: Primary group
- `home_directory`: Home directory path
- `shell`: User shell (default: /bin/bash)

#### delete_user
Deletes a user account.

**Required Parameters:**
- `user_name`: Username

#### create_group
Creates a group.

**Required Parameters:**
- `group_name`: Group name

**Optional Parameters:**
- `group_id`: GID

#### delete_group
Deletes a group.

**Required Parameters:**
- `group_name`: Group name

### SyncIQ Replication

#### create_synciq_policy
Creates a replication policy.

**Required Parameters:**
- `policy_name`: Policy name
- `source_root_path`: Source path
- `target_host`: Target cluster IP
- `target_path`: Target path

**Optional Parameters:**
- `action`: Action type (sync, copy) (default: sync)
- `enabled`: Enable policy (default: true)

**Example:**
```yaml
- name: "Setup DR replication"
  module: storage.powerscale
  vars:
    command: create_synciq_policy
    policy_name: "prod_to_dr"
    source_root_path: "/ifs/data/production"
    target_host: "192.168.100.50"
    target_path: "/ifs/dr/production"
    action: "sync"
    enabled: true
```

#### delete_synciq_policy
Deletes a replication policy.

**Required Parameters:**
- `policy_name`: Policy name

#### show_synciq_policies
Lists all replication policies.

## Best Practices

### Security

1. **Use HTTPS with Valid Certificates**: In production, enable `validate_certs: true`
2. **Credential Management**: Store credentials securely, not in plain text
3. **Access Zones**: Use access zones for multi-tenant isolation
4. **Least Privilege**: Create service accounts with minimal required permissions
5. **Audit Logging**: Enable audit logging for compliance

### Performance

1. **SmartPools**: Use SmartPools to tier data across different node types
2. **L3 Cache**: Enable L3 cache for frequently accessed data
3. **Node Pools**: Organize nodes into pools by workload
4. **Network Optimization**: Use multiple 10GbE or higher connections
5. **Client Tuning**: Optimize SMB/NFS client settings for workload

### Capacity Management

1. **Quotas**: Implement directory and user quotas to prevent space exhaustion
2. **Monitoring**: Regularly monitor capacity with `show_cluster_capacity`
3. **Snapshot Overhead**: Monitor snapshot space consumption
4. **Data Tiering**: Use SmartPools to archive cold data
5. **Deduplication**: Consider enabling dedupe for appropriate workloads

### Data Protection

1. **Snapshots**: Implement regular snapshot schedules
2. **Retention Policies**: Define appropriate retention periods
3. **SyncIQ Replication**: Set up DR replication for critical data
4. **Testing**: Regularly test snapshot restores and failover
5. **Monitoring**: Monitor replication status and lag

### Operational

1. **Change Management**: Test changes in non-production zones first
2. **Maintenance Windows**: Schedule maintenance during low-usage periods
3. **Documentation**: Document share structures and quota policies
4. **Automation**: Use OpenFroyo stacks for consistent deployments
5. **Monitoring**: Implement comprehensive monitoring and alerting

## Troubleshooting

### Common Issues

#### Connection Refused
**Symptom**: Cannot connect to PowerScale API
**Solutions**:
- Verify network connectivity: `ping <powerscale_host>`
- Check firewall rules for port 8080
- Verify PAPI is enabled on PowerScale
- Check cluster is online: `isi status`

#### Authentication Failed
**Symptom**: 401 Unauthorized error
**Solutions**:
- Verify credentials are correct
- Check user has admin privileges
- Verify account is not locked
- Check access zone if using non-System zone

#### Permission Denied
**Symptom**: 403 Forbidden error
**Solutions**:
- Verify user has required RBAC privileges
- Check access zone permissions
- Verify path exists and is accessible
- Check file system permissions

#### Quota Creation Failed
**Symptom**: Quota cannot be created
**Solutions**:
- Verify path exists
- Check quota type matches usage (user/group/directory)
- Ensure no conflicting quotas exist
- Verify threshold values are valid

#### Snapshot Creation Failed
**Symptom**: Snapshot creation fails
**Solutions**:
- Check available space (snapshots need ~5% free space)
- Verify path is valid
- Check snapshot limits haven't been reached
- Review cluster health

### Debugging

Enable detailed logging in your stack:

```yaml
run:
  - name: "Debug cluster status"
    module: storage.powerscale
    vars:
      command: show_cluster_config
      powerscale_host: "{{ var.powerscale_host }}"
      powerscale_user: "{{ var.powerscale_user }}"
      powerscale_password: "{{ var.powerscale_password }}"
```

Check PowerScale logs:
```bash
isi statistics client show --protocols=smb,nfs
isi status
isi events list
```

## Performance Tuning

### SMB Performance

```yaml
# Enable SMB3 multichannel
- name: "Configure SMB3 settings"
  module: storage.powerscale
  vars:
    command: modify_smb_share
    share_name: "high_performance"
    # Note: Additional SMB tuning done via CLI
```

PowerScale CLI:
```bash
isi smb settings global modify --enable-smb3-encryption no
isi smb settings global modify --enable-smb3-multichannel yes
```

### NFS Performance

```yaml
# Optimize NFS settings
- name: "Create optimized NFS export"
  module: storage.powerscale
  vars:
    command: create_nfs_export
    export_paths:
      - "/ifs/data/fast"
    clients:
      - "192.168.1.0/24"
    read_only: false
    security_flavors:
      - "unix"
```

PowerScale CLI:
```bash
isi nfs settings global modify --nfsv4-enabled yes
isi nfs settings global modify --nfsv3-rdma-enabled yes
```

### Cache Optimization

PowerScale CLI for L3 cache:
```bash
isi l3 cache create l3cache-1 --data-cache-size 800G
isi l3 cache enable l3cache-1
```

## Integration Examples

### Backup Integration

```yaml
# Create backup share and schedule
run:
  - name: "Create backup share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "backups"
      path: "/ifs/data/backups"
      description: "Backup target"

  - name: "Create backup snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "backup_snapshots"
      snapshot_path: "/ifs/data/backups"
      schedule: "0 */6 * * *"  # Every 6 hours
      retention: "7 days"
```

### Multi-Tenant Configuration

```yaml
# Setup access zones for tenants
run:
  - name: "Create tenant A zone"
    module: storage.powerscale
    vars:
      command: create_access_zone
      zone_name: "tenant_a"
      path: "/ifs/tenants/a"

  - name: "Create tenant A share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "tenant_a_data"
      path: "/ifs/tenants/a/data"
      access_zone: "tenant_a"

  - name: "Set tenant A quota"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/tenants/a"
      quota_type: "directory"
      hard_threshold: 5497558138880  # 5TB
```

### Disaster Recovery Setup

```yaml
# Configure DR replication
run:
  - name: "Create DR policy for critical data"
    module: storage.powerscale
    vars:
      command: create_synciq_policy
      policy_name: "critical_data_dr"
      source_root_path: "/ifs/data/critical"
      target_host: "dr-powerscale.example.com"
      target_path: "/ifs/dr/critical"
      action: "sync"
      enabled: true

  - name: "Create DR snapshots"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "dr_snapshots"
      snapshot_path: "/ifs/data/critical"
      schedule: "0 */4 * * *"  # Every 4 hours
      retention: "30 days"
```

## Version Compatibility

| OneFS Version | PAPI Version | Supported | Notes |
|---------------|--------------|-----------|-------|
| 9.5.x         | v9           | Yes       | Full support |
| 9.4.x         | v9           | Yes       | Full support |
| 9.3.x         | v9           | Yes       | Full support |
| 9.2.x         | v9           | Yes       | Full support |
| 9.1.x         | v9           | Yes       | Full support |
| 8.2.x         | v7-v9        | Yes       | Most features |
| 8.1.x         | v7-v9        | Partial   | Limited features |
| 8.0.x         | v7           | Partial   | Limited features |

## API Reference

See [API_REFERENCE.md](./API_REFERENCE.md) for detailed API endpoint documentation.

## Workflow Examples

See [WORKFLOWS.md](./WORKFLOWS.md) for common workflow scenarios and advanced configurations.

## License

This module is part of the OpenFroyo project.

## Support

For PowerScale-specific issues:
- Dell PowerScale Documentation: https://www.dell.com/support/home/en-us/product-support/product/isilon-onefs
- Dell Support: https://www.dell.com/support

For OpenFroyo module issues:
- GitHub Issues: https://github.com/openfroyo/openfroyo/issues
