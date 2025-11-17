# Galera Cluster Management Module

Comprehensive Galera Cluster management module for OpenFroyo that provides complete control over Galera/Percona XtraDB/MariaDB Galera clusters including monitoring, node management, replication tracking, and cluster recovery operations.

## Overview

Galera Cluster is a synchronous multi-master cluster solution for MySQL, MariaDB, and Percona XtraDB. It provides:

- **Synchronous Replication**: True multi-master with no slave lag
- **Active-Active**: Read and write to any node
- **Automatic Node Provisioning**: New nodes automatically sync
- **True Parallel Replication**: Parallel apply on slave nodes
- **Automatic Node Failure Handling**: Failed nodes automatically removed

This module enables comprehensive management of Galera clusters through standardized operations.

## Galera Architecture

### Cluster Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Galera Cluster                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐      ┌──────────┐      ┌──────────┐         │
│  │  Node 1  │◄────►│  Node 2  │◄────►│  Node 3  │         │
│  │          │      │          │      │          │         │
│  │  MySQL/  │      │  MySQL/  │      │  MySQL/  │         │
│  │ MariaDB  │      │ MariaDB  │      │ MariaDB  │         │
│  │          │      │          │      │          │         │
│  │  wsrep   │      │  wsrep   │      │  wsrep   │         │
│  └──────────┘      └──────────┘      └──────────┘         │
│       ▲                 ▲                 ▲                │
│       │                 │                 │                │
│       └─────────────────┴─────────────────┘                │
│              Galera Replication                            │
│              (Group Communication)                         │
└─────────────────────────────────────────────────────────────┘
```

### Key Concepts

**Write-Set Replication (wsrep)**
- Changes are replicated as write-sets
- Certification-based replication ensures consistency
- Parallel application of write-sets on all nodes

**Quorum**
- Cluster requires majority of nodes to operate
- Split-brain prevention through quorum
- Primary component serves requests

**SST (State Snapshot Transfer)**
- Full data copy for new or significantly lagged nodes
- Methods: rsync, mysqldump, xtrabackup, mariabackup

**IST (Incremental State Transfer)**
- Transfers only missing transactions
- Faster than SST for temporarily disconnected nodes

## Node States

| State | Value | Description |
|-------|-------|-------------|
| Joining | 1 | Node joining cluster, receiving SST/IST |
| Donor/Desynced | 2 | Node serving SST/IST to another node |
| Joined | 3 | Node received full state, preparing to sync |
| Synced | 4 | Node synchronized and ready for queries |

## Features

### Cluster Management (6 Operations)
- Bootstrap new clusters
- Monitor cluster status and size
- Check node health
- Configure cluster addresses
- View wsrep status variables

### Node Management (5 Operations)
- Join nodes to clusters
- Gracefully remove nodes
- Restart nodes
- Check synchronization status
- Monitor node states

### Replication Monitoring (5 Operations)
- Track replication statistics
- Monitor flow control
- Measure apply lag
- Detect conflicts
- Analyze queue sizes

### Recovery Operations (4 Operations)
- Recover failed clusters
- Force bootstrap
- Manage safe_to_bootstrap flag
- Inspect grastate.dat

## Installation

### Prerequisites

**Required:**
- MySQL/MariaDB with Galera support installed
- `mysql` command-line client
- SSH access to cluster nodes
- Appropriate database user credentials

**Supported Distributions:**
- MariaDB Galera Cluster 10.x+
- Percona XtraDB Cluster 5.7+, 8.0+
- MySQL Galera Cluster (via Codership)

### Module Installation

The module is located at `modules/db/galeradb/` and includes:
- `module.ofy.yml` - Module definition
- `defaults.ofy.yml` - Default variables
- `wasm/galeradb.wasm` - Compiled WASM module

## Usage

### Basic Cluster Status Check

```yaml
- name: Check Galera cluster status
  module: db/galeradb
  vars:
    command: show_cluster_status
    galera_host: db1.example.com
    galera_user: monitor
    galera_password: monitor_pass
```

### Monitor Cluster Size

```yaml
- name: Check cluster size
  module: db/galeradb
  vars:
    command: show_cluster_size
    galera_host: db1.example.com
    galera_user: monitor
```

### Check Node Health

```yaml
- name: Verify node is healthy
  module: db/galeradb
  vars:
    command: check_node_status
    galera_host: localhost
    galera_user: root
    galera_password: secret
```

### Bootstrap Cluster

```yaml
- name: Bootstrap first node
  module: db/galeradb
  vars:
    command: bootstrap_cluster
    mysql_service: mysql
    grastate_file: /var/lib/mysql/grastate.dat
```

### Join Node to Cluster

```yaml
- name: Add node to cluster
  module: db/galeradb
  vars:
    command: join_cluster
    cluster_address: gcomm://node1.example.com,node2.example.com
    mysql_service: mysql
```

## Configuration

### Connection Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `galera_host` | localhost | MySQL server hostname |
| `galera_port` | 3306 | MySQL server port |
| `galera_user` | root | Database user |
| `galera_password` | "" | Database password |

### Cluster Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `cluster_address` | "" | Cluster address (gcomm://host1,host2,...) |
| `cluster_name` | galera_cluster | Cluster name |
| `node_address` | "" | This node's IP address |
| `node_name` | "" | This node's name |

### Service Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `mysql_service` | mysql | MySQL service name (mysql/mariadb/mysqld) |
| `grastate_file` | /var/lib/mysql/grastate.dat | Path to grastate.dat |
| `command_timeout` | 30 | Command timeout in seconds |

### Bootstrap Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `safe_to_bootstrap` | 0 | Safe to bootstrap flag (0 or 1) |

## Commands Reference

### Cluster Management Commands

#### bootstrap_cluster
Bootstrap a new Galera cluster (first node).

```yaml
vars:
  command: bootstrap_cluster
  mysql_service: mysql
```

**Prerequisites:**
- MySQL service should be stopped
- Node should have safe_to_bootstrap: 1 in grastate.dat
- Only run on the most advanced node during recovery

#### show_cluster_status
Display comprehensive cluster status information.

```yaml
vars:
  command: show_cluster_status
```

**Returns:**
- Cluster size, status, UUID
- Node state and readiness
- Connection status
- Provider version

#### show_cluster_size
Show the number of nodes in the cluster.

```yaml
vars:
  command: show_cluster_size
```

#### check_node_status
Check if node is healthy and ready for queries.

```yaml
vars:
  command: check_node_status
```

**Health Criteria:**
- wsrep_ready = ON
- wsrep_connected = ON
- wsrep_local_state_comment = Synced

#### set_cluster_address
Update the cluster address (gcomm:// URL).

```yaml
vars:
  command: set_cluster_address
  cluster_address: gcomm://node1,node2,node3
```

#### show_wsrep_status
Display all wsrep_* status variables.

```yaml
vars:
  command: show_wsrep_status
```

### Node Management Commands

#### join_cluster
Join this node to an existing cluster.

```yaml
vars:
  command: join_cluster
  cluster_address: gcomm://node1,node2,node3
  mysql_service: mysql
```

#### leave_cluster
Gracefully remove this node from the cluster.

```yaml
vars:
  command: leave_cluster
  mysql_service: mysql
```

#### restart_node
Restart the MySQL service on this node.

```yaml
vars:
  command: restart_node
  mysql_service: mysql
```

#### check_node_synced
Verify if node is synchronized with the cluster.

```yaml
vars:
  command: check_node_synced
```

#### show_node_state
Display detailed node state information.

```yaml
vars:
  command: show_node_state
```

### Replication Monitoring Commands

#### show_replication_status
Display comprehensive replication statistics.

```yaml
vars:
  command: show_replication_status
```

**Metrics:**
- Transactions committed, replicated, received
- Bytes replicated and received
- Local commits and certification failures
- Brute force aborts

#### show_flow_control
Monitor flow control events.

```yaml
vars:
  command: show_flow_control
```

**Flow Control:**
- Paused time and count
- Sent and received FC messages
- Indicates replication performance issues

#### show_apply_lag
Check replication apply lag (receive queue).

```yaml
vars:
  command: show_apply_lag
```

**Metrics:**
- Current receive queue size
- Average, max, min queue sizes
- High values indicate apply lag

#### show_conflicts
Display certification conflict statistics.

```yaml
vars:
  command: show_conflicts
```

**Conflicts:**
- Certification failures
- Brute force aborts
- Certification dependency distance

#### show_queue_size
Show send and receive queue sizes.

```yaml
vars:
  command: show_queue_size
```

### Recovery Operations

#### recover_cluster
Recover a completely failed cluster.

```yaml
vars:
  command: recover_cluster
  grastate_file: /var/lib/mysql/grastate.dat
```

**Process:**
1. Check grastate.dat
2. Set safe_to_bootstrap: 1
3. Bootstrap cluster

#### force_bootstrap
Force bootstrap even if grastate indicates unsafe.

```yaml
vars:
  command: force_bootstrap
  grastate_file: /var/lib/mysql/grastate.dat
```

**Warning:** Use only when certain this node has the most recent data.

#### set_safe_to_bootstrap
Manually set the safe_to_bootstrap flag.

```yaml
vars:
  command: set_safe_to_bootstrap
  safe_to_bootstrap: 1
  grastate_file: /var/lib/mysql/grastate.dat
```

#### check_grastate
Read and parse the grastate.dat file.

```yaml
vars:
  command: check_grastate
  grastate_file: /var/lib/mysql/grastate.dat
```

## Important Status Variables

### Cluster Health

| Variable | Description | Healthy Value |
|----------|-------------|---------------|
| wsrep_cluster_status | Cluster status | Primary |
| wsrep_cluster_size | Number of nodes | ≥ 1 |
| wsrep_ready | Node ready for queries | ON |
| wsrep_connected | Connected to cluster | ON |
| wsrep_local_state_comment | Node state | Synced |

### Performance Monitoring

| Variable | Description | Good Value |
|----------|-------------|------------|
| wsrep_flow_control_paused | Time paused by FC | Close to 0 |
| wsrep_cert_deps_distance | Parallel apply window | Higher is better |
| wsrep_local_recv_queue | Apply queue size | Close to 0 |
| wsrep_local_send_queue | Send queue size | Close to 0 |

### Replication Health

| Variable | Description | Monitor For |
|----------|-------------|-------------|
| wsrep_local_cert_failures | Certification failures | Sudden increases |
| wsrep_local_bf_aborts | Brute force aborts | High values |
| wsrep_local_commits | Local commits | Transaction rate |
| wsrep_replicated | Replicated transactions | Outbound traffic |
| wsrep_received | Received transactions | Inbound traffic |

## Security Considerations

### Database Credentials
- Use dedicated monitoring user with minimal privileges
- Store passwords in vault or encrypted variables
- Rotate credentials regularly

### Required Privileges

**For Monitoring:**
```sql
GRANT PROCESS, REPLICATION CLIENT ON *.* TO 'monitor'@'%';
```

**For Cluster Operations:**
```sql
GRANT SUPER, PROCESS, REPLICATION CLIENT ON *.* TO 'admin'@'%';
```

### SSH Access
- Ensure SSH keys are properly configured
- Use dedicated service account for operations
- Limit sudo access to specific commands

## Troubleshooting

### Cluster Won't Start

**Check grastate.dat:**
```yaml
- name: Check grastate
  module: db/galeradb
  vars:
    command: check_grastate
```

Look for:
- `seqno: -1` indicates crash
- `safe_to_bootstrap: 0` prevents bootstrap

**Solution:**
1. Find node with highest seqno
2. Set safe_to_bootstrap: 1 on that node
3. Bootstrap from that node

### Node Stuck in Joining

**Check:**
- SST method configuration
- Network connectivity
- Disk space on donor and joiner

**Monitor SST Progress:**
Check MySQL error log for SST messages.

### Flow Control Active

**Symptoms:**
- wsrep_flow_control_paused > 0
- High receive queue

**Solutions:**
- Increase wsrep_slave_threads
- Improve disk I/O
- Optimize slow queries
- Consider hardware upgrades

### Split-Brain Scenario

**Detection:**
- Multiple nodes with wsrep_cluster_status = Primary
- Different cluster UUIDs

**Resolution:**
1. Stop all nodes
2. Identify most advanced node (highest seqno)
3. Set safe_to_bootstrap: 1 on that node
4. Bootstrap from that node
5. Join other nodes one by one

## Examples

See [WORKFLOWS.md](WORKFLOWS.md) for complete workflow examples including:
- Setting up a new cluster
- Adding nodes to existing cluster
- Monitoring cluster health
- Performing cluster recovery
- Handling node failures

## Performance Best Practices

### Configuration Recommendations

**For Write-Heavy Workloads:**
```
wsrep_slave_threads = 4-8 (CPU cores)
wsrep_certification_rules = optimized
```

**For Large Transactions:**
```
wsrep_max_ws_size = 2G
wsrep_max_ws_rows = 0 (unlimited)
```

**For WAN Clusters:**
```
wsrep_provider_options = "evs.keepalive_period=PT3S; evs.inactive_timeout=PT30S"
```

### Monitoring Recommendations

**Monitor These Metrics:**
1. Flow control frequency and duration
2. Receive and send queue sizes
3. Certification failure rate
4. Apply lag
5. Cluster size changes

**Alert On:**
- wsrep_ready = OFF
- wsrep_cluster_status != Primary
- wsrep_local_state_comment != Synced
- Flow control active > 0.1 (10%)
- High certification failure rate

## Building

To build the WASM module:

```bash
cd modules/db/galeradb/wasm
make build
```

Requirements:
- Go 1.21+ with WASI support
- GOOS=wasip1 GOARCH=wasm

## Testing

Run the test stack:

```bash
froyo apply modules/db/galeradb/test.ofy
```

Note: Some operations in the test file are commented out as they modify cluster state.

## References

- [Galera Cluster Documentation](https://galeracluster.com/library/documentation/)
- [MariaDB Galera Cluster](https://mariadb.com/kb/en/galera-cluster/)
- [Percona XtraDB Cluster](https://www.percona.com/software/mysql-database/percona-xtradb-cluster)
- [Galera Status Variables](https://galeracluster.com/library/documentation/galera-status-variables.html)

## License

Part of the OpenFroyo project.

## Support

For issues, questions, or contributions, please refer to the main OpenFroyo repository.
