# Galera Cluster Commands Reference

Complete reference for all 20 commands available in the Galera Cluster management module.

## Command Categories

- [Cluster Management](#cluster-management) (6 commands)
- [Node Management](#node-management) (5 commands)
- [Replication Monitoring](#replication-monitoring) (5 commands)
- [Recovery Operations](#recovery-operations) (4 commands)

---

## Cluster Management

### bootstrap_cluster

Bootstrap a new Galera cluster. This should only be executed on the first node when starting a new cluster or recovering from a complete cluster failure.

**Usage:**
```yaml
- name: Bootstrap Galera cluster
  module: db/galeradb
  vars:
    command: bootstrap_cluster
    mysql_service: mysql
    grastate_file: /var/lib/mysql/grastate.dat
```

**Parameters:**
- `mysql_service` (required): MySQL service name (mysql, mariadb, mysqld)
- `grastate_file` (required): Path to grastate.dat file

**Process:**
1. Stops MySQL service
2. Sets safe_to_bootstrap: 1 in grastate.dat
3. Starts MySQL with `galera_new_cluster` or `mysqld_safe --wsrep-new-cluster`

**Output:**
```json
{
  "status": "changed",
  "message": "Galera cluster bootstrapped successfully",
  "facts": {
    "bootstrapped": true
  }
}
```

**Prerequisites:**
- MySQL service must be stopped first
- Node should be the most advanced (highest seqno)
- Only one node should bootstrap

**Warning:** Never bootstrap multiple nodes simultaneously. This causes split-brain.

---

### show_cluster_status

Display comprehensive cluster status including size, health, and node state.

**Usage:**
```yaml
- name: Check cluster status
  module: db/galeradb
  vars:
    command: show_cluster_status
    galera_host: localhost
    galera_user: monitor
    galera_password: monitor_pass
```

**Parameters:**
- `galera_host`: MySQL server hostname (default: localhost)
- `galera_port`: MySQL server port (default: 3306)
- `galera_user`: Database username (default: root)
- `galera_password`: Database password

**Output:**
```json
{
  "status": "ok",
  "message": "Cluster status: Primary, Size: 3, Node state: Synced",
  "facts": {
    "cluster_size": "3",
    "cluster_status": "Primary",
    "local_state_comment": "Synced",
    "ready": "ON",
    "connected": "ON",
    "local_state": "4",
    "cluster_uuid": "abc-123-def",
    "provider_version": "4.x"
  }
}
```

**Key Facts Explained:**
- `cluster_size`: Number of nodes in cluster
- `cluster_status`: Primary (healthy) or Non-Primary (split-brain)
- `local_state_comment`: Current node state (Synced = healthy)
- `ready`: ON means node can accept queries
- `connected`: ON means connected to cluster
- `local_state`: Numeric state (4 = Synced)

**Health Check:**
A healthy cluster shows:
- cluster_status = "Primary"
- cluster_size ≥ 1
- local_state_comment = "Synced"
- ready = "ON"
- connected = "ON"

---

### show_cluster_size

Show the number of nodes currently in the cluster.

**Usage:**
```yaml
- name: Get cluster size
  module: db/galeradb
  vars:
    command: show_cluster_size
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Cluster size: 3 nodes",
  "facts": {
    "cluster_size": "3"
  }
}
```

**Use Cases:**
- Quick cluster size check
- Monitoring node count
- Detecting node failures (size decreases)
- Validating node additions (size increases)

**Expected Values:**
- Minimum: 1 (single node cluster)
- Recommended: 3+ (for quorum and high availability)
- Odd numbers preferred (for quorum voting)

---

### check_node_status

Check if the node is healthy and ready to accept queries.

**Usage:**
```yaml
- name: Verify node health
  module: db/galeradb
  vars:
    command: check_node_status
    galera_host: db1.example.com
    galera_user: monitor
    galera_password: secret
```

**Output:**
```json
{
  "status": "ok",
  "message": "Node status - Ready: ON, Connected: ON, State: Synced, Healthy: true",
  "facts": {
    "ready": "ON",
    "connected": "ON",
    "state": "Synced",
    "healthy": true
  }
}
```

**Health Criteria:**
A node is considered healthy when:
- `ready` = "ON" (can accept queries)
- `connected` = "ON" (connected to cluster)
- `state` = "Synced" (synchronized with cluster)

**Unhealthy States:**
- ready = "OFF": Node cannot accept queries
- connected = "OFF": Node disconnected from cluster
- state = "Joining": Node syncing (SST/IST in progress)
- state = "Donor": Node providing SST (temporarily read-only)

**Use Cases:**
- Pre-maintenance health check
- Load balancer health checks
- Automated monitoring
- Before performing writes

---

### set_cluster_address

Update the cluster address (gcomm:// connection string) dynamically.

**Usage:**
```yaml
- name: Update cluster address
  module: db/galeradb
  vars:
    command: set_cluster_address
    cluster_address: gcomm://node1,node2,node3
    galera_host: localhost
    galera_user: root
    galera_password: secret
```

**Parameters:**
- `cluster_address` (required): Cluster address in gcomm:// format

**Output:**
```json
{
  "status": "changed",
  "message": "Cluster address set to: gcomm://node1,node2,node3",
  "facts": {
    "cluster_address": "gcomm://node1,node2,node3"
  }
}
```

**Address Formats:**

**Join Existing Cluster:**
```
gcomm://node1.example.com,node2.example.com,node3.example.com
```

**Bootstrap (Empty):**
```
gcomm://
```

**With Ports:**
```
gcomm://192.168.1.10:4567,192.168.1.11:4567,192.168.1.12:4567
```

**Notes:**
- Changes take effect after MySQL restart
- Use for cluster reconfigurations
- Update all nodes consistently

---

### show_wsrep_status

Display all wsrep_* status variables for detailed cluster diagnostics.

**Usage:**
```yaml
- name: Get all wsrep status
  module: db/galeradb
  vars:
    command: show_wsrep_status
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Retrieved 50+ wsrep status variables",
  "facts": {
    "wsrep_cluster_size": "3",
    "wsrep_cluster_status": "Primary",
    "wsrep_ready": "ON",
    "wsrep_connected": "ON",
    "wsrep_local_state_comment": "Synced",
    "wsrep_flow_control_paused": "0.0",
    "wsrep_cert_deps_distance": "12.5",
    "wsrep_local_recv_queue": "0",
    "wsrep_local_send_queue": "0",
    "wsrep_replicated": "12345",
    "wsrep_received": "12340",
    "... (50+ more variables) ...": "..."
  }
}
```

**Important Variable Categories:**

**Cluster State:**
- wsrep_cluster_size, wsrep_cluster_status, wsrep_cluster_uuid

**Node State:**
- wsrep_ready, wsrep_connected, wsrep_local_state_comment

**Replication:**
- wsrep_replicated, wsrep_received, wsrep_replicated_bytes

**Performance:**
- wsrep_flow_control_paused, wsrep_cert_deps_distance

**Queues:**
- wsrep_local_recv_queue, wsrep_local_send_queue

**Use Cases:**
- Deep diagnostics
- Performance analysis
- Troubleshooting
- Capacity planning

---

## Node Management

### join_cluster

Join this node to an existing Galera cluster.

**Usage:**
```yaml
- name: Join node to cluster
  module: db/galeradb
  vars:
    command: join_cluster
    cluster_address: gcomm://node1,node2,node3
    mysql_service: mysql
    galera_host: localhost
    galera_user: root
    galera_password: secret
```

**Parameters:**
- `cluster_address` (required): Cluster address with existing nodes
- `mysql_service` (required): MySQL service name

**Process:**
1. Sets wsrep_cluster_address to specified value
2. Restarts MySQL service
3. Node performs SST or IST to synchronize

**Output:**
```json
{
  "status": "changed",
  "message": "Node joining cluster at gcomm://node1,node2,node3",
  "facts": {
    "cluster_address": "gcomm://node1,node2,node3",
    "action": "join"
  }
}
```

**SST vs IST:**
- **SST (State Snapshot Transfer)**: Full data copy (new node or large gap)
- **IST (Incremental State Transfer)**: Only missing transactions (small gap)

**Prerequisites:**
- At least one node in cluster_address must be running
- Sufficient disk space for SST if needed
- Network connectivity to cluster nodes

**Monitoring Join:**
After executing, monitor node state:
1. Initially: "Joining" (receiving SST/IST)
2. Then: "Joined" (preparing to sync)
3. Finally: "Synced" (ready for queries)

---

### leave_cluster

Gracefully remove this node from the cluster.

**Usage:**
```yaml
- name: Remove node from cluster
  module: db/galeradb
  vars:
    command: leave_cluster
    mysql_service: mysql
```

**Parameters:**
- `mysql_service` (required): MySQL service name

**Process:**
1. Stops MySQL service gracefully
2. Node removed from cluster
3. Other nodes detect removal

**Output:**
```json
{
  "status": "changed",
  "message": "Node left cluster (mysql service stopped)",
  "facts": {
    "action": "leave"
  }
}
```

**Use Cases:**
- Maintenance windows
- Node decommissioning
- Cluster downsizing
- Before hardware replacement

**Impact:**
- Cluster size decreases by 1
- No data loss (other nodes retain data)
- Quorum may be affected if removing too many nodes

**Best Practices:**
- Don't remove too many nodes simultaneously (maintain quorum)
- For odd-sized clusters, keep at least (N/2)+1 nodes
- Drain connections before leaving (remove from load balancer)

---

### restart_node

Restart the MySQL service on this node.

**Usage:**
```yaml
- name: Restart MySQL node
  module: db/galeradb
  vars:
    command: restart_node
    mysql_service: mysql
```

**Parameters:**
- `mysql_service` (required): MySQL service name

**Process:**
1. Restarts MySQL service (systemctl restart)
2. Node rejoins cluster automatically
3. Performs IST if possible, SST if needed

**Output:**
```json
{
  "status": "changed",
  "message": "Node restarted successfully",
  "facts": {
    "action": "restart"
  }
}
```

**Use Cases:**
- Apply configuration changes
- Recover from errors
- Upgrade MySQL version
- Clear stuck states

**Downtime:**
- Brief interruption during restart
- IST: seconds to minutes
- SST: minutes to hours (depending on data size)

**After Restart:**
Monitor node state until "Synced":
```yaml
- name: Wait for sync
  module: db/galeradb
  vars:
    command: check_node_synced
```

---

### check_node_synced

Verify if the node is fully synchronized with the cluster.

**Usage:**
```yaml
- name: Check if node is synced
  module: db/galeradb
  vars:
    command: check_node_synced
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Node sync status: Synced (synced: true)",
  "facts": {
    "state": "Synced",
    "synced": true
  }
}
```

**Possible States:**
- **Synced** (synced: true): Fully synchronized, ready
- **Joining** (synced: false): Receiving SST/IST
- **Joined** (synced: false): Preparing to sync
- **Donor** (synced: false): Providing SST to another node

**Use Cases:**
- Post-join verification
- Post-restart verification
- Before performing writes
- Load balancer decisions

**Wait for Sync Loop:**
```yaml
- name: Wait for node to sync
  module: db/galeradb
  vars:
    command: check_node_synced
  loop: "{{ range(1, 60) }}"
  until: facts.synced == true
  delay: 10
```

---

### show_node_state

Display detailed node state information including numeric state value and description.

**Usage:**
```yaml
- name: Show node state details
  module: db/galeradb
  vars:
    command: show_node_state
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Node state: Synced (Synced)",
  "facts": {
    "state": "4",
    "state_comment": "Synced",
    "state_desc": "Synced",
    "state_uuid": "abc-123-def"
  }
}
```

**State Values:**
| Value | Name | Description |
|-------|------|-------------|
| 1 | Joining | Node joining cluster, receiving state transfer |
| 2 | Donor/Desynced | Node sending state transfer to another node |
| 3 | Joined | Node received state, preparing to synchronize |
| 4 | Synced | Node synchronized and ready for queries |

**State Transitions:**
```
Joining (1) → Joined (3) → Synced (4)
                ↑              ↓
           Donor/Desynced (2) ←
```

**Use Cases:**
- Detailed diagnostics
- Understanding node behavior
- Troubleshooting sync issues
- Monitoring state changes

---

## Replication Monitoring

### show_replication_status

Display comprehensive replication statistics including transactions and bytes.

**Usage:**
```yaml
- name: Check replication stats
  module: db/galeradb
  vars:
    command: show_replication_status
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Replication status retrieved successfully",
  "facts": {
    "wsrep_last_committed": "12345",
    "wsrep_replicated": "12340",
    "wsrep_replicated_bytes": "1234567890",
    "wsrep_received": "12338",
    "wsrep_received_bytes": "1234567800",
    "wsrep_local_commits": "12340",
    "wsrep_local_cert_failures": "5",
    "wsrep_local_bf_aborts": "2"
  }
}
```

**Key Metrics:**

**Transaction Counts:**
- `wsrep_last_committed`: Last GTID committed
- `wsrep_replicated`: Transactions replicated TO other nodes
- `wsrep_received`: Transactions received FROM other nodes
- `wsrep_local_commits`: Local transactions committed

**Data Volume:**
- `wsrep_replicated_bytes`: Bytes sent to other nodes
- `wsrep_received_bytes`: Bytes received from other nodes

**Issues:**
- `wsrep_local_cert_failures`: Certification failures (conflicts)
- `wsrep_local_bf_aborts`: Brute force aborts

**Analysis:**
- High cert_failures → Write conflicts between nodes
- High bf_aborts → Local transactions aborted by cluster
- Replicated ≈ Received → Balanced write load
- Replicated >> Received → This node handles most writes

---

### show_flow_control

Monitor flow control events which indicate replication performance issues.

**Usage:**
```yaml
- name: Check flow control
  module: db/galeradb
  vars:
    command: show_flow_control
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Flow control status retrieved successfully",
  "facts": {
    "wsrep_flow_control_paused": "0.0",
    "wsrep_flow_control_paused_ns": "0",
    "wsrep_flow_control_sent": "0",
    "wsrep_flow_control_recv": "0"
  }
}
```

**What is Flow Control?**
Flow control pauses replication when a node can't keep up with write rate. It prevents slow nodes from falling too far behind.

**Key Metrics:**
- `wsrep_flow_control_paused`: Fraction of time paused (0.0 - 1.0)
- `wsrep_flow_control_paused_ns`: Nanoseconds paused
- `wsrep_flow_control_sent`: FC messages sent by this node
- `wsrep_flow_control_recv`: FC messages received from other nodes

**Thresholds:**
- **Good**: paused < 0.01 (1%)
- **Warning**: paused 0.01 - 0.1 (1-10%)
- **Critical**: paused > 0.1 (>10%)

**Causes:**
- Slow disk I/O
- Insufficient wsrep_slave_threads
- Large transactions
- Network latency

**Solutions:**
1. Increase wsrep_slave_threads
2. Improve disk I/O (SSDs, RAID)
3. Optimize slow queries
4. Add more memory
5. Reduce transaction sizes

---

### show_apply_lag

Check replication apply lag by monitoring the receive queue size.

**Usage:**
```yaml
- name: Check apply lag
  module: db/galeradb
  vars:
    command: show_apply_lag
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Apply lag - Current queue: 0",
  "facts": {
    "recv_queue": "0",
    "recv_queue_avg": "0.5",
    "recv_queue_max": "5",
    "recv_queue_min": "0"
  }
}
```

**What is Apply Lag?**
Apply lag occurs when a node receives transactions faster than it can apply them. Transactions queue up in the receive queue.

**Key Metrics:**
- `recv_queue`: Current receive queue size
- `recv_queue_avg`: Average queue size
- `recv_queue_max`: Maximum queue size observed
- `recv_queue_min`: Minimum queue size

**Thresholds:**
- **Good**: recv_queue < 10
- **Warning**: recv_queue 10-100
- **Critical**: recv_queue > 100

**Impact:**
- High queue → Node falling behind
- Triggers flow control when threshold reached
- Eventually causes SST if gap too large

**Solutions:**
- Increase wsrep_slave_threads (more parallel apply)
- Improve disk I/O performance
- Reduce write rate
- Optimize database configuration

**Related Variables:**
```yaml
wsrep_slave_threads: 1-8  # Adjust based on CPU cores
```

---

### show_conflicts

Display certification conflict and abort statistics.

**Usage:**
```yaml
- name: Check conflicts
  module: db/galeradb
  vars:
    command: show_conflicts
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Conflict statistics retrieved successfully",
  "facts": {
    "wsrep_local_cert_failures": "5",
    "wsrep_local_bf_aborts": "2",
    "wsrep_cert_deps_distance": "12.5"
  }
}
```

**Conflict Types:**

**Certification Failures:**
- Occur during commit certification
- Transaction conflicts with concurrent transaction on another node
- Transaction rolled back, client receives error

**Brute Force Aborts:**
- Occur when applying remote transaction
- Remote transaction conflicts with local uncommitted transaction
- Local transaction forcibly aborted

**Key Metrics:**
- `wsrep_local_cert_failures`: Certification failures on this node
- `wsrep_local_bf_aborts`: Local transactions aborted by cluster
- `wsrep_cert_deps_distance`: Parallel apply window (higher = better)

**Normal Rates:**
- **Low**: < 1% of transactions
- **Medium**: 1-5% (acceptable for some workloads)
- **High**: > 5% (indicates design issues)

**Causes:**
- Hot rows (many nodes updating same rows)
- Long-running transactions
- Optimistic locking in application
- Missing WHERE clauses (full table updates)

**Solutions:**
1. **Application Design:**
   - Avoid hot rows
   - Use shorter transactions
   - Implement retry logic
   - Add conflict resolution

2. **Database Design:**
   - Partition hot tables
   - Use different schemas per node
   - Implement versioning

3. **Configuration:**
   - Tune wsrep_certification_rules
   - Adjust transaction isolation level

---

### show_queue_size

Show both send and receive queue sizes to identify replication bottlenecks.

**Usage:**
```yaml
- name: Check queue sizes
  module: db/galeradb
  vars:
    command: show_queue_size
    galera_host: localhost
    galera_user: monitor
```

**Output:**
```json
{
  "status": "ok",
  "message": "Queue sizes - Send: 0, Recv: 0",
  "facts": {
    "send_queue": "0",
    "recv_queue": "0",
    "send_queue_avg": "0.2",
    "recv_queue_avg": "0.5"
  }
}
```

**Queue Types:**

**Send Queue (wsrep_local_send_queue):**
- Transactions waiting to be sent to cluster
- High value indicates network issues or slow cluster

**Receive Queue (wsrep_local_recv_queue):**
- Transactions waiting to be applied
- High value indicates apply lag

**Analysis:**

| Send Queue | Recv Queue | Diagnosis |
|------------|------------|-----------|
| High | Low | Network issues, slow other nodes |
| Low | High | This node slow (disk I/O, CPU) |
| High | High | Cluster-wide overload |
| Low | Low | Healthy |

**Thresholds:**
- **Good**: Both < 10
- **Warning**: Either 10-100
- **Critical**: Either > 100

**Monitoring:**
Monitor these continuously to detect:
- Performance degradation
- Capacity limits
- Hardware issues
- Network problems

---

## Recovery Operations

### recover_cluster

Recover a completely failed Galera cluster.

**Usage:**
```yaml
- name: Recover failed cluster
  module: db/galeradb
  vars:
    command: recover_cluster
    grastate_file: /var/lib/mysql/grastate.dat
    mysql_service: mysql
```

**Parameters:**
- `grastate_file` (required): Path to grastate.dat
- `mysql_service` (required): MySQL service name

**Process:**
1. Checks grastate.dat file
2. Sets safe_to_bootstrap: 1
3. Bootstraps cluster

**Output:**
```json
{
  "status": "changed",
  "message": "Galera cluster bootstrapped successfully",
  "facts": {
    "bootstrapped": true
  }
}
```

**When to Use:**
- All nodes stopped/crashed
- Cannot start cluster normally
- After power outage
- After network partition resolved

**Manual Recovery Steps:**
1. **Stop all nodes** (if any running)
2. **Check seqno on all nodes:**
   ```bash
   cat /var/lib/mysql/grastate.dat | grep seqno
   ```
3. **Find highest seqno** (most advanced node)
4. **Run recover_cluster on that node**
5. **Start other nodes normally** (they will join via IST/SST)

**Important:**
- Always recover from node with highest seqno
- Never bootstrap multiple nodes
- Check grastate.dat before recovery

---

### force_bootstrap

Force bootstrap even if grastate.dat indicates it's not safe.

**Usage:**
```yaml
- name: Force bootstrap cluster
  module: db/galeradb
  vars:
    command: force_bootstrap
    grastate_file: /var/lib/mysql/grastate.dat
    mysql_service: mysql
```

**Parameters:**
- `grastate_file` (required): Path to grastate.dat
- `mysql_service` (required): MySQL service name

**Process:**
1. Forces safe_to_bootstrap: 1 (overrides safety check)
2. Bootstraps cluster

**Output:**
```json
{
  "status": "changed",
  "message": "Galera cluster bootstrapped successfully",
  "facts": {
    "bootstrapped": true
  }
}
```

**When to Use:**
- Last resort when normal recovery fails
- Certain this node has most recent data
- Acceptable to lose some transactions

**Risks:**
- **Data loss possible** if wrong node bootstrapped
- May create inconsistencies
- Should only be used when certain

**Safety Check:**
Before force_bootstrap:
1. Check all nodes' seqno values
2. Verify this node has highest seqno
3. Confirm data loss acceptable
4. Document decision

**Alternative:**
If unsure, use `recover_cluster` instead (safer).

---

### set_safe_to_bootstrap

Manually set the safe_to_bootstrap flag in grastate.dat.

**Usage:**
```yaml
- name: Set safe to bootstrap
  module: db/galeradb
  vars:
    command: set_safe_to_bootstrap
    safe_to_bootstrap: 1
    grastate_file: /var/lib/mysql/grastate.dat
```

**Parameters:**
- `safe_to_bootstrap` (required): 0 or 1
- `grastate_file` (required): Path to grastate.dat

**Output:**
```json
{
  "status": "changed",
  "message": "safe_to_bootstrap set to 1",
  "facts": {
    "safe_to_bootstrap": 1,
    "grastate_file": "/var/lib/mysql/grastate.dat"
  }
}
```

**Flag Values:**
- **0**: Not safe to bootstrap (default after shutdown)
- **1**: Safe to bootstrap (node can start new cluster)

**Use Cases:**

**Set to 1:**
- Preparing to bootstrap cluster
- After verifying this is most advanced node
- Manual recovery procedure

**Set to 0:**
- Prevent accidental bootstrap
- After recovery complete
- Security measure

**Manual Edit:**
```bash
sed -i 's/safe_to_bootstrap: 0/safe_to_bootstrap: 1/' /var/lib/mysql/grastate.dat
```

**Important:**
- Only set to 1 on one node at a time
- Verify node has latest data before setting
- Reset to 0 after bootstrap complete

---

### check_grastate

Read and parse the grastate.dat file to inspect cluster state information.

**Usage:**
```yaml
- name: Inspect grastate file
  module: db/galeradb
  vars:
    command: check_grastate
    grastate_file: /var/lib/mysql/grastate.dat
```

**Parameters:**
- `grastate_file` (required): Path to grastate.dat

**Output:**
```json
{
  "status": "ok",
  "message": "Grastate file read successfully",
  "facts": {
    "content": "# full file content...",
    "version": "2.1",
    "uuid": "abc-123-def-456",
    "seqno": "12345",
    "safe_to_bootstrap": "0"
  }
}
```

**Grastate.dat Fields:**

**version:**
- Grastate file format version
- Usually "2.1"

**uuid:**
- Cluster UUID
- All nodes in cluster have same UUID
- Different UUID = was in different cluster

**seqno:**
- Sequence number (last committed transaction)
- Higher = more recent data
- `-1` = unclean shutdown
- Special value indicating crash

**safe_to_bootstrap:**
- `0` = Not safe to bootstrap
- `1` = Safe to bootstrap

**Interpreting seqno:**

| seqno Value | Meaning | Action |
|-------------|---------|--------|
| Positive number | Clean shutdown | Can join cluster normally |
| -1 | Unclean shutdown/crash | Need recovery, compare seqno across nodes |
| 0 | New node | Needs SST |

**Recovery Decision Tree:**

```
All nodes down?
├─ Yes → Check seqno on all nodes
│  ├─ All have clean seqno → Bootstrap highest
│  └─ Some have seqno -1 → Run recovery procedure
└─ No → Join existing cluster normally
```

**Example:**
```bash
# GALERA saved state
version: 2.1
uuid:    a1b2c3d4-e5f6-11eb-ba80-0242ac130004
seqno:   12345
safe_to_bootstrap: 0
```

**Use Cases:**
- Pre-recovery diagnostics
- Determining most advanced node
- Troubleshooting cluster issues
- Verifying clean shutdown
- Audit trail

---

## Command Quick Reference

### By Frequency of Use

**Daily Operations:**
1. `show_cluster_status` - Overall health check
2. `check_node_status` - Individual node health
3. `show_replication_status` - Replication monitoring

**Performance Monitoring:**
1. `show_flow_control` - Performance issues
2. `show_apply_lag` - Apply queue size
3. `show_queue_size` - Send/receive queues
4. `show_conflicts` - Certification conflicts

**Cluster Changes:**
1. `join_cluster` - Add nodes
2. `leave_cluster` - Remove nodes
3. `restart_node` - Apply changes

**Emergency Operations:**
1. `check_grastate` - Diagnose before recovery
2. `recover_cluster` - Standard recovery
3. `force_bootstrap` - Emergency bootstrap

### By Risk Level

**Safe (Read-Only):**
- show_cluster_status
- show_cluster_size
- check_node_status
- show_wsrep_status
- show_node_state
- show_replication_status
- show_flow_control
- show_apply_lag
- show_conflicts
- show_queue_size
- check_grastate

**Low Risk (Reversible):**
- set_cluster_address
- check_node_synced

**Medium Risk (Causes Downtime):**
- join_cluster
- leave_cluster
- restart_node
- set_safe_to_bootstrap

**High Risk (Cluster Impact):**
- bootstrap_cluster
- recover_cluster
- force_bootstrap

### By Required Privileges

**PROCESS + REPLICATION CLIENT:**
- All show_* commands
- All check_* commands

**SUPER:**
- set_cluster_address
- join_cluster
- leave_cluster
- restart_node
- bootstrap_cluster
- recover_cluster
- force_bootstrap

**File System Access:**
- check_grastate
- set_safe_to_bootstrap

---

## Best Practices

### Monitoring
- Check cluster_status every 5 minutes
- Monitor flow_control continuously
- Alert on queue_size > 100
- Track conflict rates

### Operations
- Always check_node_status before maintenance
- Use check_grastate before recovery
- Verify check_node_synced after changes
- Document recovery decisions

### Safety
- Never bootstrap multiple nodes
- Always find highest seqno for recovery
- Test recovery procedures regularly
- Maintain quorum during operations

---

## See Also

- [README.md](README.md) - Module overview and architecture
- [WORKFLOWS.md](WORKFLOWS.md) - Complete workflow examples
- [Galera Documentation](https://galeracluster.com/library/documentation/)
