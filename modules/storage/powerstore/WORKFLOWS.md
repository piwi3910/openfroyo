# PowerStore Storage Management Workflows

Common storage management workflows and patterns for Dell PowerStore using OpenFroyo.

## Table of Contents

1. [Basic Storage Provisioning](#basic-storage-provisioning)
2. [VMware Integration](#vmware-integration)
3. [Database Storage Setup](#database-storage-setup)
4. [Backup and Recovery](#backup-and-recovery)
5. [Disaster Recovery](#disaster-recovery)
6. [Performance Optimization](#performance-optimization)
7. [Capacity Management](#capacity-management)
8. [Multi-Tier Storage](#multi-tier-storage)

---

## Basic Storage Provisioning

### Simple Volume Provisioning

Provision a basic storage volume and map it to a host.

```yaml
# stack: provision-basic-volume.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "{{ var.powerstore_ip }}"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"

run:
  - name: Create 500GB application volume
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "app-data-001"
      volume_size: 500
      description: "Application data volume"

  - name: Create host definition
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "app-server-01"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:app-server-01"

  - name: Map volume to host
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Verify mapping
    module: storage/powerstore
    vars:
      command: show_volume_details
      volume_id: "{{ facts.volume_id }}"
```

### Multi-Volume Provisioning

Provision multiple volumes with different characteristics.

```yaml
run:
  - name: Create host for database server
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "db-server-01"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:db-server-01"

  - name: Create database data volume (1TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "db-data-001"
      volume_size: 1000
      description: "Database data files"

  - name: Create database log volume (200GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "db-log-001"
      volume_size: 200
      description: "Database transaction logs"

  - name: Create backup volume (2TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "db-backup-001"
      volume_size: 2000
      description: "Database backups"

  - name: Map data volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.db_data_volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Map log volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.db_log_volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Map backup volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.db_backup_volume_id }}"
      host_id: "{{ facts.host_id }}"
```

---

## VMware Integration

### vSphere Cluster Datastore Setup

Configure shared storage for a VMware vSphere cluster.

```yaml
# stack: vmware-cluster-storage.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"
  cluster_name: "vsphere-prod-cluster"

run:
  - name: Create host group for vSphere cluster
    module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "{{ var.cluster_name }}"
      description: "VMware vSphere production cluster"

  - name: Register ESXi host 1
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "esxi-01.example.com"
      os_type: "ESXi"
      initiators:
        - "iqn.1998-01.com.vmware:esxi-01-12345678"
        - "21:00:00:24:ff:11:11:11"
        - "21:00:00:24:ff:11:11:12"

  - name: Register ESXi host 2
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "esxi-02.example.com"
      os_type: "ESXi"
      initiators:
        - "iqn.1998-01.com.vmware:esxi-02-87654321"
        - "21:00:00:24:ff:22:22:21"
        - "21:00:00:24:ff:22:22:22"

  - name: Register ESXi host 3
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "esxi-03.example.com"
      os_type: "ESXi"
      initiators:
        - "iqn.1998-01.com.vmware:esxi-03-11223344"
        - "21:00:00:24:ff:33:33:31"
        - "21:00:00:24:ff:33:33:32"

  - name: Add host 1 to cluster group
    module: storage/powerstore
    vars:
      command: add_host_to_group
      host_group_id: "{{ facts.host_group_id }}"
      host_id: "{{ facts.esxi01_host_id }}"

  - name: Add host 2 to cluster group
    module: storage/powerstore
    vars:
      command: add_host_to_group
      host_group_id: "{{ facts.host_group_id }}"
      host_id: "{{ facts.esxi02_host_id }}"

  - name: Add host 3 to cluster group
    module: storage/powerstore
    vars:
      command: add_host_to_group
      host_group_id: "{{ facts.host_group_id }}"
      host_id: "{{ facts.esxi03_host_id }}"

  - name: Create VMFS datastore 1 (4TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vmfs-prod-datastore-01"
      volume_size: 4000
      description: "Production VM datastore 1"

  - name: Create VMFS datastore 2 (4TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vmfs-prod-datastore-02"
      volume_size: 4000
      description: "Production VM datastore 2"

  - name: Map datastore 1 to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.datastore1_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map datastore 2 to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.datastore2_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Verify cluster configuration
    module: storage/powerstore
    vars:
      command: show_host_groups
```

### VDI Storage Setup

Configure storage for Virtual Desktop Infrastructure (VDI).

```yaml
run:
  - name: Create host group for VDI hosts
    module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "vdi-infrastructure"
      description: "VDI host cluster"

  - name: Create VDI desktop pool volume (8TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vdi-desktop-pool-01"
      volume_size: 8000
      description: "VDI persistent desktop pool"

  - name: Create VDI replica volume (2TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vdi-replica-pool-01"
      volume_size: 2000
      description: "VDI replica pool"

  - name: Create VDI user data volume (4TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "vdi-user-data-01"
      volume_size: 4000
      description: "VDI user profiles and data"

  - name: Map desktop pool to VDI cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.desktop_pool_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map replica pool to VDI cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.replica_pool_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map user data to VDI cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.user_data_volume_id }}"
      host_id: "{{ facts.host_group_id }}"
```

---

## Database Storage Setup

### Oracle RAC Cluster

Configure shared storage for Oracle Real Application Clusters.

```yaml
# stack: oracle-rac-storage.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"

run:
  - name: Create host group for Oracle RAC
    module: storage/powerstore
    vars:
      command: create_host_group
      host_group_name: "oracle-rac-cluster"
      description: "Oracle RAC database cluster"

  - name: Register RAC node 1
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "rac-node-01"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:rac-node-01"

  - name: Register RAC node 2
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "rac-node-02"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:rac-node-02"

  - name: Add node 1 to RAC group
    module: storage/powerstore
    vars:
      command: add_host_to_group
      host_group_id: "{{ facts.host_group_id }}"
      host_id: "{{ facts.node1_host_id }}"

  - name: Add node 2 to RAC group
    module: storage/powerstore
    vars:
      command: add_host_to_group
      host_group_id: "{{ facts.host_group_id }}"
      host_id: "{{ facts.node2_host_id }}"

  - name: Create OCR volume (20GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-ocr-01"
      volume_size: 20
      description: "Oracle Cluster Registry"

  - name: Create voting disk volume (10GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-voting-01"
      volume_size: 10
      description: "Oracle voting disks"

  - name: Create ASM data volume (2TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-asm-data-01"
      volume_size: 2000
      description: "Oracle ASM data diskgroup"

  - name: Create ASM FRA volume (1TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "oracle-asm-fra-01"
      volume_size: 1000
      description: "Oracle ASM fast recovery area"

  - name: Map OCR to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.ocr_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map voting disk to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.voting_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map data diskgroup to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.data_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Map FRA to cluster
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.fra_volume_id }}"
      host_id: "{{ facts.host_group_id }}"

  - name: Create snapshot rule for database backups
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "oracle-hourly-snapshots"
      desired_retention: 168  # 7 days
      interval: "One_Hour"
```

### Microsoft SQL Server

Configure storage for SQL Server Always On cluster.

```yaml
run:
  - name: Create SQL Server primary host
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "sql-primary-01"
      os_type: "Windows"
      initiators:
        - "iqn.1991-05.com.microsoft:sql-primary-01"

  - name: Create SQL Server secondary host
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "sql-secondary-01"
      os_type: "Windows"
      initiators:
        - "iqn.1991-05.com.microsoft:sql-secondary-01"

  - name: Create SQL data volume (1TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "sql-data-001"
      volume_size: 1000
      description: "SQL Server data files"

  - name: Create SQL log volume (300GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "sql-log-001"
      volume_size: 300
      description: "SQL Server transaction logs"

  - name: Create SQL TempDB volume (200GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "sql-tempdb-001"
      volume_size: 200
      description: "SQL Server TempDB"

  - name: Map data volume to primary
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.data_volume_id }}"
      host_id: "{{ facts.primary_host_id }}"

  - name: Map log volume to primary
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.log_volume_id }}"
      host_id: "{{ facts.primary_host_id }}"

  - name: Map TempDB to primary
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.tempdb_volume_id }}"
      host_id: "{{ facts.primary_host_id }}"
```

---

## Backup and Recovery

### Snapshot-Based Backup

Implement scheduled snapshot-based backups.

```yaml
# stack: snapshot-backup-setup.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"

run:
  - name: Create hourly snapshot rule
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "production-hourly"
      desired_retention: 72  # 3 days
      interval: "One_Hour"

  - name: Create daily snapshot rule
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "production-daily"
      desired_retention: 720  # 30 days
      interval: "Every_24_Hours"

  - name: Create weekly snapshot rule
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "production-weekly"
      desired_retention: 4320  # 180 days (6 months)
      interval: "Every_24_Hours"

  - name: List all volumes for protection
    module: storage/powerstore
    vars:
      command: show_volumes

  - name: Take manual pre-maintenance snapshot
    module: storage/powerstore
    vars:
      command: create_snapshot
      volume_id: "{{ var.critical_volume_id }}"
      snapshot_name: "pre-maintenance-{{ var.date }}"
      description: "Snapshot before scheduled maintenance"
```

### Volume Recovery Workflow

Recover from a snapshot.

```yaml
run:
  - name: List available snapshots
    module: storage/powerstore
    vars:
      command: show_snapshots
      volume_id: "{{ var.volume_id }}"

  - name: Verify snapshot details
    module: storage/powerstore
    vars:
      command: show_volume_details
      volume_id: "{{ var.snapshot_id }}"

  - name: Unmap volume from production host
    module: storage/powerstore
    vars:
      command: unmap_volume
      volume_id: "{{ var.volume_id }}"
      host_id: "{{ var.host_id }}"

  - name: Restore volume from snapshot
    module: storage/powerstore
    vars:
      command: restore_snapshot
      snapshot_id: "{{ var.snapshot_id }}"

  - name: Remap volume to production host
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ var.volume_id }}"
      host_id: "{{ var.host_id }}"

  - name: Verify volume status
    module: storage/powerstore
    vars:
      command: show_volume_details
      volume_id: "{{ var.volume_id }}"
```

### Clone for Testing

Create production clones for testing and development.

```yaml
run:
  - name: Create snapshot of production volume
    module: storage/powerstore
    vars:
      command: create_snapshot
      volume_id: "{{ var.prod_volume_id }}"
      snapshot_name: "prod-clone-source-{{ var.date }}"

  - name: Clone production volume
    module: storage/powerstore
    vars:
      command: clone_volume
      volume_id: "{{ var.prod_volume_id }}"
      clone_name: "dev-test-clone-{{ var.date }}"
      description: "Development testing clone"

  - name: Map clone to test environment
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.clone_id }}"
      host_id: "{{ var.test_host_id }}"
```

---

## Disaster Recovery

### Configure Asynchronous Replication

Set up replication to a remote PowerStore system.

```yaml
# stack: setup-dr-replication.ofy
inventory:
  - powerstore-primary

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"
  remote_system_id: "remote-powerstore-uuid"

run:
  - name: List critical volumes for DR
    module: storage/powerstore
    vars:
      command: show_volumes

  - name: Create replication session for volume 1
    module: storage/powerstore
    vars:
      command: create_replication_session
      volume_id: "{{ var.critical_vol_1_id }}"
      remote_system_id: "{{ var.remote_system_id }}"
      replication_rule: "async-replication-4hr"

  - name: Create replication session for volume 2
    module: storage/powerstore
    vars:
      command: create_replication_session
      volume_id: "{{ var.critical_vol_2_id }}"
      remote_system_id: "{{ var.remote_system_id }}"
      replication_rule: "async-replication-4hr"

  - name: Create replication session for volume 3
    module: storage/powerstore
    vars:
      command: create_replication_session
      volume_id: "{{ var.critical_vol_3_id }}"
      remote_system_id: "{{ var.remote_system_id }}"
      replication_rule: "async-replication-4hr"

  - name: Verify replication status
    module: storage/powerstore
    vars:
      command: show_replication_sessions
```

### DR Testing Workflow

Test disaster recovery procedures without impacting production.

```yaml
run:
  - name: Pause production replication
    module: storage/powerstore
    vars:
      command: pause_replication
      replication_session_id: "{{ var.replication_session_id }}"

  - name: Clone replicated volume on DR site
    module: storage/powerstore
    vars:
      command: clone_volume
      volume_id: "{{ var.replicated_volume_id }}"
      clone_name: "dr-test-clone-{{ var.date }}"
      description: "DR test clone"

  - name: Map clone to DR test host
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.clone_id }}"
      host_id: "{{ var.dr_test_host_id }}"

  - name: Resume production replication
    module: storage/powerstore
    vars:
      command: resume_replication
      replication_session_id: "{{ var.replication_session_id }}"
```

---

## Performance Optimization

### Performance Monitoring Setup

Monitor and analyze storage performance.

```yaml
# stack: performance-monitoring.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"
  monitoring_interval: "Five_Mins"

run:
  - name: Get cluster performance baseline
    module: storage/powerstore
    vars:
      command: show_cluster_metrics
      cluster_id: "{{ var.cluster_id }}"
      interval: "{{ var.monitoring_interval }}"

  - name: Monitor critical volume 1
    module: storage/powerstore
    vars:
      command: show_volume_metrics
      volume_id: "{{ var.critical_vol_1_id }}"
      interval: "{{ var.monitoring_interval }}"

  - name: Monitor critical volume 2
    module: storage/powerstore
    vars:
      command: show_volume_metrics
      volume_id: "{{ var.critical_vol_2_id }}"
      interval: "{{ var.monitoring_interval }}"

  - name: Monitor appliance performance
    module: storage/powerstore
    vars:
      command: show_appliance_metrics
      appliance_id: "{{ var.appliance_id }}"
      interval: "{{ var.monitoring_interval }}"

  - name: Check for performance alerts
    module: storage/powerstore
    vars:
      command: show_alerts
```

---

## Capacity Management

### Capacity Planning Workflow

Analyze and plan for capacity growth.

```yaml
# stack: capacity-planning.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"

run:
  - name: Get overall capacity information
    module: storage/powerstore
    vars:
      command: show_capacity

  - name: List all volumes with sizes
    module: storage/powerstore
    vars:
      command: show_volumes

  - name: List all snapshots
    module: storage/powerstore
    vars:
      command: show_snapshots

  - name: Check cluster configuration
    module: storage/powerstore
    vars:
      command: show_cluster_info

  - name: Check capacity alerts
    module: storage/powerstore
    vars:
      command: show_alerts
```

### Volume Expansion

Expand volumes to meet growing capacity needs.

```yaml
run:
  - name: Check current volume size
    module: storage/powerstore
    vars:
      command: show_volume_details
      volume_id: "{{ var.volume_id }}"

  - name: Expand volume by 500GB
    module: storage/powerstore
    vars:
      command: resize_volume
      volume_id: "{{ var.volume_id }}"
      new_size: 1500  # Current + 500GB

  - name: Verify new size
    module: storage/powerstore
    vars:
      command: show_volume_details
      volume_id: "{{ var.volume_id }}"
```

---

## Multi-Tier Storage

### Tiered Application Storage

Configure storage tiers for different performance requirements.

```yaml
# stack: multi-tier-app-storage.ofy
inventory:
  - powerstore-mgmt

defaults:
  powerstore_host: "192.168.1.100"
  powerstore_user: "admin"
  powerstore_password: "{{ vault.powerstore_password }}"

run:
  - name: Create host for app server
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "app-tier-server-01"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:app-tier-01"

  # Tier 1: High-performance database storage
  - name: Create Tier 1 volume (500GB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "app-tier1-db-001"
      volume_size: 500
      description: "Tier 1 - High-performance database"

  # Tier 2: Standard application storage
  - name: Create Tier 2 volume (1TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "app-tier2-data-001"
      volume_size: 1000
      description: "Tier 2 - Standard application data"

  # Tier 3: Archive/backup storage
  - name: Create Tier 3 volume (5TB)
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "app-tier3-archive-001"
      volume_size: 5000
      description: "Tier 3 - Archive and backup data"

  - name: Map Tier 1 volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.tier1_volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Map Tier 2 volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.tier2_volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Map Tier 3 volume
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.tier3_volume_id }}"
      host_id: "{{ facts.host_id }}"

  # Different snapshot policies per tier
  - name: Create Tier 1 snapshot rule (hourly)
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "tier1-hourly-snapshots"
      desired_retention: 48
      interval: "One_Hour"

  - name: Create Tier 2 snapshot rule (daily)
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "tier2-daily-snapshots"
      desired_retention: 720
      interval: "Every_24_Hours"

  - name: Create Tier 3 snapshot rule (weekly)
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "tier3-weekly-snapshots"
      desired_retention: 4320
      interval: "Every_24_Hours"
```

This workflows document provides comprehensive examples for common PowerStore storage management scenarios using OpenFroyo automation.
