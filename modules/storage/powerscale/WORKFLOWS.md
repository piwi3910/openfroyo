# Dell PowerScale Workflow Examples

This document provides comprehensive workflow examples for common Dell PowerScale (Isilon) NAS management scenarios using the OpenFroyo PowerScale module.

## Table of Contents

1. [Initial Cluster Setup](#initial-cluster-setup)
2. [Department File Sharing](#department-file-sharing)
3. [Multi-Tenant Configuration](#multi-tenant-configuration)
4. [Backup and Snapshot Management](#backup-and-snapshot-management)
5. [Disaster Recovery Setup](#disaster-recovery-setup)
6. [Home Directory Management](#home-directory-management)
7. [Application Integration](#application-integration)
8. [Capacity Management](#capacity-management)
9. [Performance Monitoring](#performance-monitoring)
10. [Migration Workflows](#migration-workflows)

---

## Initial Cluster Setup

### Workflow: First-Time Cluster Configuration

This workflow sets up a new PowerScale cluster with basic configuration.

```yaml
---
# Initial PowerScale cluster setup
inventory: local

defaults:
  powerscale_host: "192.168.1.100"
  powerscale_user: "admin"
  powerscale_password: "{{ vault.powerscale_password }}"

run:
  # 1. Verify cluster is accessible and gather info
  - name: "Verify cluster configuration"
    module: storage.powerscale
    vars:
      command: show_cluster_config

  - name: "Check OneFS version"
    module: storage.powerscale
    vars:
      command: show_cluster_version

  - name: "Check cluster capacity"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity

  - name: "List cluster nodes"
    module: storage.powerscale
    vars:
      command: show_cluster_nodes

  # 2. Create base directory structure
  - name: "Create base directories"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data /ifs/home /ifs/backups /ifs/zones"

  # 3. Create default access zone (optional, for multi-tenancy)
  - name: "Create production access zone"
    module: storage.powerscale
    vars:
      command: create_access_zone
      zone_name: "production"
      path: "/ifs/zones/production"
      groupnet: "groupnet0"

  # 4. Set up initial SMB share
  - name: "Create company-wide share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "company"
      path: "/ifs/data/company"
      description: "Company-wide shared folder"
      browsable: true

  # 5. Create initial NFS export for Unix systems
  - name: "Create NFS export for Unix clients"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/unix"
      clients:
        - "192.168.1.0/24"
      read_only: false

  # 6. Set up snapshot schedule
  - name: "Create daily snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "daily_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 2 * * *"  # 2 AM daily
      pattern: "daily-%Y%m%d"
      retention: "7 days"
```

---

## Department File Sharing

### Workflow: Create Department Shares with Quotas

This workflow creates isolated department shares with quotas and appropriate permissions.

```yaml
---
# Department file sharing setup
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

vars:
  departments:
    - name: "engineering"
      path: "/ifs/data/engineering"
      description: "Engineering Department"
      quota_gb: 1000
    - name: "marketing"
      path: "/ifs/data/marketing"
      description: "Marketing Department"
      quota_gb: 500
    - name: "finance"
      path: "/ifs/data/finance"
      description: "Finance Department"
      quota_gb: 200
    - name: "hr"
      path: "/ifs/data/hr"
      description: "Human Resources"
      quota_gb: 100

run:
  # Loop through departments (conceptual - OpenFroyo may handle loops differently)
  # Engineering Department
  - name: "Create engineering directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/engineering"

  - name: "Create engineering SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "engineering"
      path: "/ifs/data/engineering"
      description: "Engineering Department"
      browsable: true
      ntfs_acl_support: true

  - name: "Set engineering quota (1TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/engineering"
      quota_type: "directory"
      hard_threshold: 1099511627776  # 1TB
      soft_threshold: 989560053350   # 900GB
      enforce: true

  - name: "Create engineering NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/engineering"
      clients:
        - "192.168.10.0/24"  # Engineering VLAN
      read_only: false

  # Marketing Department
  - name: "Create marketing directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/marketing"

  - name: "Create marketing SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "marketing"
      path: "/ifs/data/marketing"
      description: "Marketing Department"
      browsable: true

  - name: "Set marketing quota (500GB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/marketing"
      quota_type: "directory"
      hard_threshold: 536870912000  # 500GB
      soft_threshold: 483183820800  # 450GB
      enforce: true

  # Finance Department (restricted access)
  - name: "Create finance directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/finance"

  - name: "Create finance SMB share (not browsable)"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "finance"
      path: "/ifs/data/finance"
      description: "Finance Department (Restricted)"
      browsable: false
      ntfs_acl_support: true

  - name: "Set finance quota (200GB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/finance"
      quota_type: "directory"
      hard_threshold: 214748364800  # 200GB
      enforce: true

  # Create snapshot schedule for all departments
  - name: "Create department snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "dept_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 1,13 * * *"  # Twice daily at 1 AM and 1 PM
      pattern: "dept-%Y%m%d-%H%M"
      retention: "14 days"
```

---

## Multi-Tenant Configuration

### Workflow: Complete Multi-Tenant Setup

This workflow demonstrates isolation using access zones for multiple tenants.

```yaml
---
# Multi-tenant PowerScale configuration
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # Tenant A Setup
  # ==================
  - name: "Create Tenant A access zone"
    module: storage.powerscale
    vars:
      command: create_access_zone
      zone_name: "tenant_a"
      path: "/ifs/zones/tenant_a"
      groupnet: "groupnet0"

  - name: "Create Tenant A base directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/zones/tenant_a/data"

  - name: "Create Tenant A SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "tenant_a_data"
      path: "/ifs/zones/tenant_a/data"
      access_zone: "tenant_a"
      description: "Tenant A Data Share"
      browsable: true

  - name: "Create Tenant A NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/zones/tenant_a/data"
      clients:
        - "10.100.1.0/24"  # Tenant A network
      access_zone: "tenant_a"
      read_only: false

  - name: "Set Tenant A quota (5TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/zones/tenant_a"
      quota_type: "directory"
      hard_threshold: 5497558138880  # 5TB
      soft_threshold: 4947802324992  # 4.5TB
      enforce: true

  - name: "Create Tenant A snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "tenant_a_snapshots"
      snapshot_path: "/ifs/zones/tenant_a"
      schedule: "0 */6 * * *"  # Every 6 hours
      pattern: "tenant_a-%Y%m%d-%H%M"
      retention: "7 days"

  # ==================
  # Tenant B Setup
  # ==================
  - name: "Create Tenant B access zone"
    module: storage.powerscale
    vars:
      command: create_access_zone
      zone_name: "tenant_b"
      path: "/ifs/zones/tenant_b"
      groupnet: "groupnet0"

  - name: "Create Tenant B base directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/zones/tenant_b/data"

  - name: "Create Tenant B SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "tenant_b_data"
      path: "/ifs/zones/tenant_b/data"
      access_zone: "tenant_b"
      description: "Tenant B Data Share"
      browsable: true

  - name: "Create Tenant B NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/zones/tenant_b/data"
      clients:
        - "10.100.2.0/24"  # Tenant B network
      access_zone: "tenant_b"
      read_only: false

  - name: "Set Tenant B quota (3TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/zones/tenant_b"
      quota_type: "directory"
      hard_threshold: 3298534883328  # 3TB
      soft_threshold: 2968681395200  # 2.7TB
      enforce: true

  - name: "Create Tenant B snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "tenant_b_snapshots"
      snapshot_path: "/ifs/zones/tenant_b"
      schedule: "0 */6 * * *"  # Every 6 hours
      pattern: "tenant_b-%Y%m%d-%H%M"
      retention: "7 days"

  # ==================
  # Verification
  # ==================
  - name: "List all access zones"
    module: storage.powerscale
    vars:
      command: show_access_zones

  - name: "List all quotas"
    module: storage.powerscale
    vars:
      command: show_quotas

  - name: "Check cluster capacity after setup"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity
```

---

## Backup and Snapshot Management

### Workflow: Comprehensive Backup Strategy

This workflow implements a multi-tiered snapshot strategy with different retention periods.

```yaml
---
# Backup and snapshot management
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # Hourly Snapshots (Short Retention)
  # ==================
  - name: "Create hourly snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "hourly_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 * * * *"  # Every hour
      pattern: "hourly-%Y%m%d-%H00"
      retention: "24 hours"

  # ==================
  # Daily Snapshots (Medium Retention)
  # ==================
  - name: "Create daily snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "daily_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 2 * * *"  # 2 AM daily
      pattern: "daily-%Y%m%d"
      retention: "30 days"

  # ==================
  # Weekly Snapshots (Long Retention)
  # ==================
  - name: "Create weekly snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "weekly_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 3 * * 0"  # 3 AM Sunday
      pattern: "weekly-%Y-W%U"
      retention: "90 days"

  # ==================
  # Monthly Snapshots (Archive)
  # ==================
  - name: "Create monthly snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "monthly_snapshots"
      snapshot_path: "/ifs/data"
      schedule: "0 4 1 * *"  # 4 AM on 1st of month
      pattern: "monthly-%Y-%m"
      retention: "365 days"

  # ==================
  # Create dedicated backup share
  # ==================
  - name: "Create backup directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/backups"

  - name: "Create backup SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "backups"
      path: "/ifs/backups"
      description: "Backup Target Share"
      browsable: false

  - name: "Create backup NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/backups"
      clients:
        - "192.168.1.50"  # Backup server
      read_only: false
      root_clients:
        - "192.168.1.50"

  - name: "Set backup quota (10TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/backups"
      quota_type: "directory"
      hard_threshold: 10995116277760  # 10TB
      soft_threshold: 9895604650112   # 9TB
      enforce: true

  # ==================
  # Create ad-hoc snapshot for maintenance
  # ==================
  - name: "Create pre-maintenance snapshot"
    module: storage.powerscale
    vars:
      command: create_snapshot
      snapshot_path: "/ifs/data"
      snapshot_name: "pre_maintenance_{{ var.date }}"
      snapshot_alias: "before_maint"

  # ==================
  # List current snapshots
  # ==================
  - name: "List all snapshots"
    module: storage.powerscale
    vars:
      command: show_snapshots
```

---

## Disaster Recovery Setup

### Workflow: SyncIQ Replication Configuration

This workflow sets up disaster recovery using SyncIQ replication.

```yaml
---
# Disaster recovery with SyncIQ
inventory: local

defaults:
  powerscale_host: "192.168.1.100"  # Primary cluster
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"
  dr_cluster: "192.168.100.50"  # DR cluster

run:
  # ==================
  # Primary Site Configuration
  # ==================
  - name: "Verify primary cluster status"
    module: storage.powerscale
    vars:
      command: show_cluster_config

  - name: "Check primary cluster capacity"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity

  # ==================
  # Create SyncIQ Policies for Critical Data
  # ==================
  - name: "Create SyncIQ policy for production data"
    module: storage.powerscale
    vars:
      command: create_synciq_policy
      policy_name: "prod_to_dr"
      source_root_path: "/ifs/data/production"
      target_host: "{{ var.dr_cluster }}"
      target_path: "/ifs/dr/production"
      action: "sync"
      enabled: true

  - name: "Create SyncIQ policy for engineering"
    module: storage.powerscale
    vars:
      command: create_synciq_policy
      policy_name: "eng_to_dr"
      source_root_path: "/ifs/data/engineering"
      target_host: "{{ var.dr_cluster }}"
      target_path: "/ifs/dr/engineering"
      action: "sync"
      enabled: true

  - name: "Create SyncIQ policy for finance"
    module: storage.powerscale
    vars:
      command: create_synciq_policy
      policy_name: "finance_to_dr"
      source_root_path: "/ifs/data/finance"
      target_host: "{{ var.dr_cluster }}"
      target_path: "/ifs/dr/finance"
      action: "sync"
      enabled: true

  # ==================
  # Create Snapshots Before Replication
  # ==================
  - name: "Create snapshot schedule for production"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "prod_pre_sync"
      snapshot_path: "/ifs/data/production"
      schedule: "0 */4 * * *"  # Every 4 hours
      pattern: "prod-sync-%Y%m%d-%H%M"
      retention: "7 days"

  # ==================
  # Verify Replication Setup
  # ==================
  - name: "List all SyncIQ policies"
    module: storage.powerscale
    vars:
      command: show_synciq_policies

  - name: "Check cluster health"
    module: storage.powerscale
    vars:
      command: show_cluster_health

  # ==================
  # DR Site Verification (switch to DR cluster)
  # ==================
  - name: "Verify DR cluster config"
    module: storage.powerscale
    vars:
      command: show_cluster_config
      powerscale_host: "{{ var.dr_cluster }}"

  - name: "Check DR cluster capacity"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity
      powerscale_host: "{{ var.dr_cluster }}"
```

---

## Home Directory Management

### Workflow: User Home Directories with Quotas

This workflow creates home directories with per-user quotas.

```yaml
---
# Home directory management
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"
  home_base: "/ifs/home"
  default_quota_gb: 50

run:
  # ==================
  # Setup Home Directory Infrastructure
  # ==================
  - name: "Create home directory base"
    module: generic.command
    vars:
      cmd: "mkdir -p {{ var.home_base }}"

  - name: "Create home SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "home"
      path: "{{ var.home_base }}"
      description: "User Home Directories"
      browsable: false

  - name: "Create home NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "{{ var.home_base }}"
      clients:
        - "192.168.1.0/24"
      read_only: false

  # ==================
  # Create Users and Home Directories
  # ==================
  - name: "Create user: jdoe"
    module: storage.powerscale
    vars:
      command: create_user
      user_name: "jdoe"
      user_id: "5001"
      primary_group: "users"
      home_directory: "/ifs/home/jdoe"
      shell: "/bin/bash"

  - name: "Create jdoe home directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/home/jdoe && chown 5001:100 /ifs/home/jdoe"

  - name: "Set quota for jdoe (50GB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/home"
      quota_type: "user"
      quota_user: "jdoe"
      hard_threshold: 53687091200  # 50GB
      soft_threshold: 48318382080  # 45GB
      enforce: true

  - name: "Create user: asmith"
    module: storage.powerscale
    vars:
      command: create_user
      user_name: "asmith"
      user_id: "5002"
      primary_group: "users"
      home_directory: "/ifs/home/asmith"
      shell: "/bin/bash"

  - name: "Create asmith home directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/home/asmith && chown 5002:100 /ifs/home/asmith"

  - name: "Set quota for asmith (50GB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/home"
      quota_type: "user"
      quota_user: "asmith"
      hard_threshold: 53687091200  # 50GB
      enforce: true

  # ==================
  # Create User Groups
  # ==================
  - name: "Create developers group"
    module: storage.powerscale
    vars:
      command: create_group
      group_name: "developers"
      group_id: "6001"

  - name: "Create managers group"
    module: storage.powerscale
    vars:
      command: create_group
      group_name: "managers"
      group_id: "6002"

  # ==================
  # Setup Home Directory Snapshots
  # ==================
  - name: "Create home directory snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "home_snapshots"
      snapshot_path: "/ifs/home"
      schedule: "0 */4 * * *"  # Every 4 hours
      pattern: "home-%Y%m%d-%H%M"
      retention: "14 days"

  # ==================
  # Verification
  # ==================
  - name: "List all user quotas"
    module: storage.powerscale
    vars:
      command: show_quotas
```

---

## Application Integration

### Workflow: Database Storage Configuration

This workflow configures PowerScale for database application workloads.

```yaml
---
# Database application storage
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # PostgreSQL Storage
  # ==================
  - name: "Create PostgreSQL data directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/postgresql"

  - name: "Create PostgreSQL NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/postgresql"
      clients:
        - "192.168.1.100"  # PostgreSQL server
      read_only: false
      root_clients:
        - "192.168.1.100"
      security_flavors:
        - "unix"

  - name: "Set PostgreSQL quota (2TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/postgresql"
      quota_type: "directory"
      hard_threshold: 2199023255552  # 2TB
      enforce: true

  - name: "Create PostgreSQL snapshot schedule"
    module: storage.powerscale
    vars:
      command: create_snapshot_schedule
      schedule_name: "postgresql_snapshots"
      snapshot_path: "/ifs/data/postgresql"
      schedule: "0 */2 * * *"  # Every 2 hours
      pattern: "pg-%Y%m%d-%H%M"
      retention: "7 days"

  # ==================
  # MongoDB Storage
  # ==================
  - name: "Create MongoDB data directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/mongodb"

  - name: "Create MongoDB NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/mongodb"
      clients:
        - "192.168.1.101"  # MongoDB server
      read_only: false
      root_clients:
        - "192.168.1.101"

  - name: "Set MongoDB quota (5TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/mongodb"
      quota_type: "directory"
      hard_threshold: 5497558138880  # 5TB
      enforce: true

  # ==================
  # Media Storage for Applications
  # ==================
  - name: "Create media storage directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/media"

  - name: "Create media SMB share"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "media"
      path: "/ifs/data/media"
      description: "Application Media Storage"
      browsable: true

  - name: "Create media NFS export"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/data/media"
      clients:
        - "192.168.1.0/24"
      read_only: false

  - name: "Set media quota (10TB)"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/media"
      quota_type: "directory"
      hard_threshold: 10995116277760  # 10TB
      soft_threshold: 9895604650112   # 9TB
      enforce: true
```

---

## Capacity Management

### Workflow: Monitor and Report on Storage Usage

This workflow monitors capacity and quota usage.

```yaml
---
# Capacity monitoring and reporting
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # Cluster Capacity Overview
  # ==================
  - name: "Get cluster capacity statistics"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity

  - name: "Get cluster performance metrics"
    module: storage.powerscale
    vars:
      command: show_cluster_performance

  - name: "Get cluster health status"
    module: storage.powerscale
    vars:
      command: show_cluster_health

  # ==================
  # Quota Reports
  # ==================
  - name: "List all quotas with usage"
    module: storage.powerscale
    vars:
      command: show_quotas

  - name: "Get detailed quota reports"
    module: storage.powerscale
    vars:
      command: show_quota_reports

  # ==================
  # Snapshot Usage
  # ==================
  - name: "List all snapshots"
    module: storage.powerscale
    vars:
      command: show_snapshots

  # ==================
  # Share Inventory
  # ==================
  - name: "List all SMB shares"
    module: storage.powerscale
    vars:
      command: show_smb_shares

  - name: "List all NFS exports"
    module: storage.powerscale
    vars:
      command: show_nfs_exports

  # ==================
  # Active Sessions
  # ==================
  - name: "Show active SMB sessions"
    module: storage.powerscale
    vars:
      command: show_smb_sessions

  # ==================
  # Node Status
  # ==================
  - name: "Get cluster node information"
    module: storage.powerscale
    vars:
      command: show_cluster_nodes
```

---

## Performance Monitoring

### Workflow: Performance Baseline and Monitoring

This workflow establishes performance baselines and monitoring.

```yaml
---
# Performance monitoring setup
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # Cluster Performance Metrics
  # ==================
  - name: "Collect cluster performance data"
    module: storage.powerscale
    vars:
      command: show_cluster_performance

  - name: "Get cluster capacity for IOPs calculation"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity

  # ==================
  # Protocol-Specific Metrics
  # ==================
  - name: "Check active SMB sessions"
    module: storage.powerscale
    vars:
      command: show_smb_sessions

  # ==================
  # Health Checks
  # ==================
  - name: "Verify cluster health"
    module: storage.powerscale
    vars:
      command: show_cluster_health

  - name: "Check all cluster nodes"
    module: storage.powerscale
    vars:
      command: show_cluster_nodes

  # Note: For detailed performance monitoring, use PowerScale InsightIQ
  # or integrate with monitoring tools like Prometheus/Grafana
```

---

## Migration Workflows

### Workflow: Data Migration to PowerScale

This workflow assists with migrating data to PowerScale.

```yaml
---
# Data migration to PowerScale
inventory: local

defaults:
  powerscale_host: "{{ var.powerscale_host }}"
  powerscale_user: "{{ var.powerscale_user }}"
  powerscale_password: "{{ var.powerscale_password }}"

run:
  # ==================
  # Pre-Migration Preparation
  # ==================
  - name: "Verify cluster capacity before migration"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity

  - name: "Create migration staging directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/migration/staging"

  # ==================
  # Create Target Structure
  # ==================
  - name: "Create target data directory"
    module: generic.command
    vars:
      cmd: "mkdir -p /ifs/data/migrated"

  - name: "Create SMB share for migration"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "migration"
      path: "/ifs/migration/staging"
      description: "Migration Staging Area"
      browsable: false

  - name: "Create NFS export for migration tools"
    module: storage.powerscale
    vars:
      command: create_nfs_export
      export_paths:
        - "/ifs/migration/staging"
      clients:
        - "192.168.1.0/24"
      read_only: false
      root_clients:
        - "192.168.1.50"  # Migration server

  # ==================
  # Pre-Migration Snapshot
  # ==================
  - name: "Create pre-migration snapshot"
    module: storage.powerscale
    vars:
      command: create_snapshot
      snapshot_path: "/ifs"
      snapshot_name: "pre_migration_{{ var.date }}"
      snapshot_alias: "before_data_migration"

  # ==================
  # Post-Migration Verification
  # ==================
  - name: "Create final share for migrated data"
    module: storage.powerscale
    vars:
      command: create_smb_share
      share_name: "migrated_data"
      path: "/ifs/data/migrated"
      description: "Migrated Data"
      browsable: true

  - name: "Set quota on migrated data"
    module: storage.powerscale
    vars:
      command: create_quota
      quota_path: "/ifs/data/migrated"
      quota_type: "directory"
      hard_threshold: 5497558138880  # 5TB
      enforce: true

  - name: "Create post-migration snapshot"
    module: storage.powerscale
    vars:
      command: create_snapshot
      snapshot_path: "/ifs/data/migrated"
      snapshot_name: "post_migration_{{ var.date }}"
      snapshot_alias: "after_data_migration"

  # ==================
  # Cleanup
  # ==================
  - name: "Delete migration staging share (after verification)"
    module: storage.powerscale
    vars:
      command: delete_smb_share
      share_name: "migration"

  - name: "Verify final cluster capacity"
    module: storage.powerscale
    vars:
      command: show_cluster_capacity
```

---

## Best Practices Summary

### General Workflow Guidelines

1. **Always verify cluster health** before making changes
2. **Take snapshots** before major configuration changes
3. **Use quotas** to prevent capacity exhaustion
4. **Implement tiered snapshot schedules** for different retention requirements
5. **Monitor capacity regularly** to plan for growth
6. **Use access zones** for multi-tenant isolation
7. **Document share purposes** with clear descriptions
8. **Test disaster recovery** procedures regularly
9. **Implement least privilege** access controls
10. **Use automation** for consistent deployments

### Performance Optimization

1. Use NFS for Unix/Linux workloads
2. Use SMB3 for Windows workloads
3. Enable appropriate security features without sacrificing performance
4. Monitor and tune based on actual workload patterns
5. Use SmartPools for tiered storage (configure via PowerScale CLI)

### Security Considerations

1. Disable browsable on sensitive shares
2. Use NTFS ACLs for fine-grained SMB permissions
3. Restrict NFS exports to specific client IPs
4. Implement regular security audits
5. Use access zones for tenant isolation

### Disaster Recovery

1. Implement SyncIQ for critical data
2. Maintain multiple snapshot schedules with different retention
3. Test restoration procedures regularly
4. Document failover procedures
5. Monitor replication lag and errors

---

## Additional Resources

- [PowerScale Module README](./README.md)
- [PowerScale API Reference](./API_REFERENCE.md)
- [Dell PowerScale Documentation](https://www.dell.com/support/home/en-us/product-support/product/isilon-onefs)

---

**Version:** 1.0.0
**Last Updated:** 2025-01-16
