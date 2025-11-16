# Dell PowerStore Storage Management Module

Comprehensive OpenFroyo module for managing Dell PowerStore storage arrays through the REST API.

## Overview

Dell PowerStore is a next-generation midrange storage platform that delivers container-ready, end-to-end NVMe, and multi-cloud storage for modern workloads. This module provides complete automation capabilities for PowerStore management including:

- Volume provisioning and lifecycle management
- Host and host group configuration
- Snapshot creation and management
- Performance monitoring and metrics
- Replication session management
- System information retrieval

## Architecture

### PowerStore Platform

Dell PowerStore is a unified storage platform that supports:

- **Block Storage**: iSCSI and Fibre Channel protocols
- **File Storage**: NFS and SMB/CIFS protocols
- **vVols**: VMware Virtual Volumes
- **Container Storage**: CSI driver integration

### Key Components

1. **Management API**: RESTful API v3.0+ for all operations
2. **Appliances**: Physical or virtual storage nodes
3. **Cluster**: Collection of appliances managed as a single system
4. **Volumes**: Block storage objects that can be mapped to hosts
5. **Hosts**: Server definitions with initiator information
6. **Protection Policies**: Snapshot rules and replication settings

## Module Capabilities

### Supported Operations (40+)

#### Volume Management (10 operations)
- `create_volume` - Create new storage volumes
- `delete_volume` - Remove volumes
- `modify_volume` - Update volume properties
- `clone_volume` - Create volume clones
- `resize_volume` - Expand volume capacity
- `map_volume` - Attach volumes to hosts
- `unmap_volume` - Detach volumes from hosts
- `show_volume` - Get volume information
- `show_volumes` - List all volumes
- `show_volume_details` - Get detailed volume information

#### Host Management (6 operations)
- `create_host` - Register new hosts
- `delete_host` - Remove host definitions
- `modify_host` - Update host properties
- `add_initiator` - Add WWN/IQN to host
- `remove_initiator` - Remove WWN/IQN from host
- `show_hosts` - List all hosts

#### Host Group Management (4 operations)
- `create_host_group` - Create host groups
- `delete_host_group` - Remove host groups
- `add_host_to_group` - Add host to group
- `show_host_groups` - List all host groups

#### Snapshot Management (6 operations)
- `create_snapshot` - Create point-in-time snapshots
- `delete_snapshot` - Remove snapshots
- `restore_snapshot` - Restore volume from snapshot
- `create_snapshot_rule` - Define snapshot policies
- `delete_snapshot_rule` - Remove snapshot policies
- `show_snapshots` - List snapshots

#### Performance Metrics (4 operations)
- `show_volume_metrics` - Volume performance data
- `show_appliance_metrics` - Appliance performance data
- `show_node_metrics` - Node performance data
- `show_cluster_metrics` - Cluster performance data

#### Replication (5 operations)
- `create_replication_session` - Set up replication
- `delete_replication_session` - Remove replication
- `pause_replication` - Pause replication session
- `resume_replication` - Resume replication session
- `show_replication_sessions` - List replication sessions

#### System Information (5 operations)
- `show_cluster_info` - Cluster details
- `show_appliances` - List appliances
- `show_nodes` - List nodes
- `show_capacity` - Storage capacity information
- `show_alerts` - Active system alerts

## Installation

1. Ensure the module is in your OpenFroyo modules directory:
   ```
   modules/storage/powerstore/
   ```

2. Build the WASM module:
   ```bash
   cd modules/storage/powerstore/wasm
   make build
   ```

3. Verify the build:
   ```bash
   ls -lh powerstore.wasm
   ```

## Configuration

### Connection Parameters

```yaml
powerstore_host: "192.168.1.100"      # PowerStore management IP
powerstore_user: "admin"               # Username with appropriate permissions
powerstore_password: "SecurePass123!"  # User password
validate_certs: false                  # SSL certificate validation
port: 443                              # REST API port (default: 443)
```

### Required Permissions

The user account must have appropriate role-based permissions:

- **Storage Administrator**: Full volume and snapshot management
- **VM Administrator**: Host and initiator management
- **Operator**: Read-only access for monitoring

## Usage Examples

### Basic Volume Provisioning

```yaml
- name: Create application volume
  module: storage/powerstore
  vars:
    command: create_volume
    volume_name: "app-data-001"
    volume_size: 500  # GB
    description: "Application data volume"
```

### Host Registration

```yaml
- name: Register ESXi host
  module: storage/powerstore
  vars:
    command: create_host
    host_name: "esxi-host-01"
    os_type: "ESXi"
    initiators:
      - "iqn.1998-01.com.vmware:esxi-host-01-12345678"
      - "21:00:00:24:ff:12:34:56"
```

### Volume Mapping

```yaml
- name: Map volume to host
  module: storage/powerstore
  vars:
    command: map_volume
    volume_id: "{{ facts.volume_id }}"
    host_id: "{{ facts.host_id }}"
```

### Snapshot Protection

```yaml
- name: Create backup snapshot
  module: storage/powerstore
  vars:
    command: create_snapshot
    volume_id: "vol-12345678"
    snapshot_name: "weekly-backup-2024-01-15"
    description: "Weekly production backup"
```

### Performance Monitoring

```yaml
- name: Monitor volume performance
  module: storage/powerstore
  vars:
    command: show_volume_metrics
    volume_id: "vol-12345678"
    interval: "Five_Mins"
```

## Common Workflows

### 1. Provision Storage for New Application

```yaml
run:
  - name: Create host definition
    module: storage/powerstore
    vars:
      command: create_host
      host_name: "app-server-01"
      os_type: "Linux"
      initiators:
        - "iqn.1993-08.org.debian:01:app-server-01"

  - name: Create data volume
    module: storage/powerstore
    vars:
      command: create_volume
      volume_name: "app-data-vol"
      volume_size: 1000

  - name: Map volume to host
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.volume_id }}"
      host_id: "{{ facts.host_id }}"

  - name: Create snapshot policy
    module: storage/powerstore
    vars:
      command: create_snapshot_rule
      snapshot_rule_name: "app-hourly-snapshots"
      desired_retention: 72
      interval: "Every_Hour"
```

### 2. Clone Volume for Testing

```yaml
run:
  - name: Create production volume clone
    module: storage/powerstore
    vars:
      command: clone_volume
      volume_id: "prod-vol-12345678"
      clone_name: "test-vol-clone-001"
      description: "Clone for QA testing"

  - name: Map clone to test host
    module: storage/powerstore
    vars:
      command: map_volume
      volume_id: "{{ facts.clone_id }}"
      host_id: "test-host-id"
```

### 3. Disaster Recovery Setup

```yaml
run:
  - name: Create replication session
    module: storage/powerstore
    vars:
      command: create_replication_session
      volume_id: "prod-vol-12345678"
      remote_system_id: "dr-powerstore-uuid"
      replication_rule: "async-4hour-rpo"

  - name: Verify replication status
    module: storage/powerstore
    vars:
      command: show_replication_sessions
```

### 4. Capacity Management

```yaml
run:
  - name: Check overall capacity
    module: storage/powerstore
    vars:
      command: show_capacity

  - name: List all volumes
    module: storage/powerstore
    vars:
      command: show_volumes

  - name: Expand critical volume
    module: storage/powerstore
    vars:
      command: resize_volume
      volume_id: "vol-12345678"
      new_size: 2000  # Expand to 2TB
```

## Best Practices

### Security

1. **Credential Management**
   - Store credentials in encrypted vaults (not in plain text)
   - Use least-privilege accounts
   - Rotate passwords regularly
   - Enable audit logging

2. **Network Security**
   - Use dedicated management networks
   - Enable SSL/TLS for API connections
   - Implement firewall rules restricting API access
   - Consider VPN for remote management

### Performance

1. **Volume Design**
   - Align volume sizes with workload requirements
   - Use appropriate volume types (block vs file)
   - Monitor IOPS and throughput metrics
   - Implement QoS policies for critical workloads

2. **Host Configuration**
   - Use multipathing for redundancy
   - Configure proper queue depths
   - Optimize initiator settings
   - Balance load across appliances

### Data Protection

1. **Snapshots**
   - Implement regular snapshot schedules
   - Define appropriate retention periods
   - Test snapshot restore procedures
   - Monitor snapshot capacity consumption

2. **Replication**
   - Configure replication for critical volumes
   - Verify replication health regularly
   - Test failover procedures
   - Document RPO/RTO requirements

### Operations

1. **Monitoring**
   - Track capacity trends
   - Monitor performance metrics
   - Review system alerts
   - Maintain inventory documentation

2. **Change Management**
   - Test changes in development first
   - Use volume cloning for testing
   - Maintain rollback procedures
   - Document all configuration changes

## Troubleshooting

### Common Issues

#### Connection Failures

```
Error: API request failed: connection refused
```

**Solutions:**
- Verify PowerStore management IP is correct
- Check network connectivity to management interface
- Ensure firewall allows HTTPS (443) traffic
- Verify PowerStore management services are running

#### Authentication Errors

```
Error: API returned status 401: Unauthorized
```

**Solutions:**
- Verify username and password are correct
- Check user account is not locked
- Ensure user has appropriate role assignments
- Verify account has not expired

#### Volume Creation Failures

```
Error: API returned status 400: Insufficient capacity
```

**Solutions:**
- Check available capacity with `show_capacity`
- Review storage pool allocation
- Consider reducing volume size
- Review and delete unused volumes

#### Mapping Errors

```
Error: Volume already mapped to host
```

**Solutions:**
- Check current volume mappings with `show_volume_details`
- Unmap volume from existing host if appropriate
- Verify host and volume IDs are correct

### Debug Mode

Enable verbose logging by checking API responses in the facts output:

```yaml
- name: Debug volume creation
  module: storage/powerstore
  vars:
    command: create_volume
    volume_name: "debug-test-vol"
    volume_size: 10
```

Review the `facts` output for detailed API responses and error messages.

## API Rate Limits

PowerStore REST API implements rate limiting:

- **Default limit**: 100 requests per minute per user
- **Burst allowance**: 200 requests
- **Recovery time**: 60 seconds

For bulk operations, implement appropriate delays between requests.

## Integration Examples

### VMware vSphere Integration

```yaml
- name: Provision VMFS datastore volume
  module: storage/powerstore
  vars:
    command: create_volume
    volume_name: "vmfs-datastore-01"
    volume_size: 4000

- name: Register vSphere cluster host group
  module: storage/powerstore
  vars:
    command: create_host_group
    host_group_name: "vsphere-cluster-01"

- name: Map datastore to cluster
  module: storage/powerstore
  vars:
    command: map_volume
    volume_id: "{{ facts.volume_id }}"
    host_id: "{{ facts.host_group_id }}"
```

### Kubernetes CSI Integration

```yaml
- name: Create volume for persistent volume claim
  module: storage/powerstore
  vars:
    command: create_volume
    volume_name: "pvc-database-storage"
    volume_size: 200
    description: "K8s PVC for database pod"
```

## Version Compatibility

- **PowerStore OS**: 2.0.0.0 and later
- **REST API**: v3.0 and later
- **OpenFroyo**: 1.0+

## Support and Documentation

- **Dell PowerStore Documentation**: [Dell Support Portal](https://www.dell.com/support/powerstore)
- **REST API Reference**: Available in PowerStore UI under Settings > API
- **OpenFroyo Documentation**: See project README.md

## License

This module is part of the OpenFroyo project and follows the same license terms.
