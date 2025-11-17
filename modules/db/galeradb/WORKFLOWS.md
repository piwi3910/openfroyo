# Galera Cluster Workflows

Complete workflow examples for common Galera Cluster operations using the OpenFroyo galeradb module.

## Table of Contents

- [Initial Cluster Setup](#initial-cluster-setup)
- [Adding Nodes to Cluster](#adding-nodes-to-cluster)
- [Cluster Health Monitoring](#cluster-health-monitoring)
- [Cluster Recovery](#cluster-recovery)
- [Node Maintenance](#node-maintenance)
- [Performance Monitoring](#performance-monitoring)
- [Troubleshooting](#troubleshooting)
- [Split-Brain Recovery](#split-brain-recovery)

---

## Initial Cluster Setup

### Scenario: Setting up a new 3-node Galera cluster

**Prerequisites:**
- MySQL/MariaDB with Galera installed on all nodes
- Network connectivity between all nodes
- Galera configuration in my.cnf

**Step 1: Configure First Node**

```yaml
# File: stacks/galera_bootstrap.ofy
inventory: galera-cluster

defaults:
  galera_user: root
  galera_password: "{{ vault.mysql_root_password }}"
  mysql_service: mysql

run:
  - name: Bootstrap cluster on node1
    module: db/galeradb
    hosts: node1
    vars:
      command: bootstrap_cluster
      grastate_file: /var/lib/mysql/grastate.dat

  - name: Wait for bootstrap to complete
    module: system/sleep
    vars:
      seconds: 10

  - name: Verify node1 cluster status
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_status

  - name: Verify node1 is healthy
    module: db/galeradb
    hosts: node1
    vars:
      command: check_node_status
```

**Step 2: Add Second Node**

```yaml
  - name: Join node2 to cluster
    module: db/galeradb
    hosts: node2
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com

  - name: Wait for node2 to sync
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_synced
    loop: "{{ range(1, 30) }}"
    until: facts.synced == true
    delay: 10

  - name: Verify cluster size is 2
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size
```

**Step 3: Add Third Node**

```yaml
  - name: Join node3 to cluster
    module: db/galeradb
    hosts: node3
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com

  - name: Wait for node3 to sync
    module: db/galeradb
    hosts: node3
    vars:
      command: check_node_synced
    loop: "{{ range(1, 30) }}"
    until: facts.synced == true
    delay: 10

  - name: Verify final cluster status
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status

  - name: Verify cluster size is 3
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size
```

**Execution:**
```bash
froyo apply stacks/galera_bootstrap.ofy
```

**Expected Results:**
- Node1: Bootstrapped, cluster_size: 1 → 2 → 3
- Node2: Joined, state: Synced
- Node3: Joined, state: Synced
- All nodes: cluster_status: Primary

---

## Adding Nodes to Cluster

### Scenario: Adding a new node to existing 3-node cluster

**Step 1: Pre-Check Existing Cluster**

```yaml
# File: stacks/add_node4.ofy
inventory: galera-cluster

defaults:
  galera_user: root
  galera_password: "{{ vault.mysql_root_password }}"

run:
  - name: Check current cluster status
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_status

  - name: Verify current cluster size
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size

  - name: Check all nodes are synced
    module: db/galeradb
    hosts: @group:existing-nodes
    vars:
      command: check_node_synced
```

**Step 2: Add New Node**

```yaml
  - name: Join node4 to cluster
    module: db/galeradb
    hosts: node4
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com,node4.example.com
      mysql_service: mysql

  - name: Monitor node4 state during SST
    module: db/galeradb
    hosts: node4
    vars:
      command: show_node_state
    loop: "{{ range(1, 60) }}"
    delay: 30
```

**Step 3: Verify Addition**

```yaml
  - name: Wait for node4 to reach Synced state
    module: db/galeradb
    hosts: node4
    vars:
      command: check_node_synced
    loop: "{{ range(1, 60) }}"
    until: facts.synced == true
    delay: 10

  - name: Verify cluster size is 4
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size

  - name: Verify all nodes see same cluster
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status
```

**Step 4: Update Cluster Addresses on All Nodes**

```yaml
  - name: Update cluster address on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: set_cluster_address
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com,node4.example.com
```

---

## Cluster Health Monitoring

### Scenario: Continuous cluster health monitoring

**Complete Health Check Stack:**

```yaml
# File: stacks/galera_healthcheck.ofy
inventory: galera-cluster

defaults:
  galera_user: monitor
  galera_password: "{{ vault.monitor_password }}"

run:
  # Overall Cluster Health
  - name: Check cluster status on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status

  - name: Verify cluster size
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size

  # Node Health
  - name: Check each node status
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_status

  - name: Verify all nodes synced
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_synced

  # Performance Metrics
  - name: Check flow control on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_flow_control

  - name: Check apply lag on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_apply_lag

  - name: Check queue sizes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_queue_size

  # Replication Health
  - name: Check replication status
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_replication_status

  - name: Check conflicts
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_conflicts
```

**Automated Monitoring (Cron):**

```yaml
# Run every 5 minutes
*/5 * * * * froyo apply /opt/openfroyo/stacks/galera_healthcheck.ofy --quiet
```

**Alert Thresholds:**

```yaml
# File: stacks/galera_alerts.ofy
run:
  - name: Alert if cluster not Primary
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status
    alert:
      when: facts.cluster_status != "Primary"
      severity: critical
      message: "Node {{ host }} not in Primary cluster"

  - name: Alert if node not synced
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_synced
    alert:
      when: facts.synced == false
      severity: warning
      message: "Node {{ host }} not synchronized"

  - name: Alert on high flow control
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_flow_control
    alert:
      when: facts.wsrep_flow_control_paused > 0.1
      severity: warning
      message: "Node {{ host }} has high flow control: {{ facts.wsrep_flow_control_paused }}"

  - name: Alert on high queue size
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_queue_size
    alert:
      when: facts.recv_queue > 100
      severity: critical
      message: "Node {{ host }} has high receive queue: {{ facts.recv_queue }}"
```

---

## Cluster Recovery

### Scenario: Complete cluster failure - all nodes stopped

**Recovery Process:**

**Step 1: Assess Cluster State**

```yaml
# File: stacks/galera_recover.ofy
inventory: galera-cluster

defaults:
  mysql_service: mysql

run:
  - name: Check grastate.dat on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_grastate
      grastate_file: /var/lib/mysql/grastate.dat

  - name: Display seqno from all nodes
    module: system/command
    hosts: @group:all-nodes
    vars:
      command: "grep 'seqno:' /var/lib/mysql/grastate.dat"
```

**Step 2: Identify Most Advanced Node**

Manual analysis of seqno values:
```
node1: seqno: 12345
node2: seqno: 12340
node3: seqno: 12350  ← Highest (most advanced)
```

**Step 3: Recover from Most Advanced Node**

```yaml
  - name: Recover cluster from node with highest seqno
    module: db/galeradb
    hosts: node3  # Node with highest seqno
    vars:
      command: recover_cluster
      grastate_file: /var/lib/mysql/grastate.dat
      mysql_service: mysql

  - name: Wait for bootstrap
    module: system/sleep
    vars:
      seconds: 15

  - name: Verify node3 bootstrapped
    module: db/galeradb
    hosts: node3
    vars:
      command: show_cluster_status
```

**Step 4: Rejoin Other Nodes**

```yaml
  - name: Join node1 to recovered cluster
    module: db/galeradb
    hosts: node1
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com

  - name: Wait for node1 sync
    module: db/galeradb
    hosts: node1
    vars:
      command: check_node_synced
    loop: "{{ range(1, 60) }}"
    until: facts.synced == true
    delay: 10

  - name: Join node2 to recovered cluster
    module: db/galeradb
    hosts: node2
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com

  - name: Wait for node2 sync
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_synced
    loop: "{{ range(1, 60) }}"
    until: facts.synced == true
    delay: 10
```

**Step 5: Verify Recovery**

```yaml
  - name: Verify cluster size is 3
    module: db/galeradb
    hosts: node3
    vars:
      command: show_cluster_size

  - name: Verify all nodes healthy
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_status

  - name: Check final cluster status
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status
```

### Scenario: Recovery when seqno is -1 on all nodes

**Emergency Recovery with force_bootstrap:**

```yaml
run:
  - name: Force bootstrap on most recently updated node
    module: db/galeradb
    hosts: node3  # Choose most likely candidate
    vars:
      command: force_bootstrap
      grastate_file: /var/lib/mysql/grastate.dat
      mysql_service: mysql

  # Warning message
  - name: Log force bootstrap decision
    module: system/log
    vars:
      message: "Force bootstrapped cluster from node3. Some data loss possible."
      level: warning
```

---

## Node Maintenance

### Scenario: Performing maintenance on a cluster node

**Step 1: Pre-Maintenance Verification**

```yaml
# File: stacks/node_maintenance.ofy
inventory: galera-cluster

defaults:
  galera_user: root
  galera_password: "{{ vault.mysql_root_password }}"

run:
  - name: Verify cluster is healthy
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_status

  - name: Verify cluster has 3+ nodes
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size
    assert:
      that: facts.cluster_size >= 3
      fail_msg: "Cannot perform maintenance with less than 3 nodes"
```

**Step 2: Remove Node from Service**

```yaml
  - name: Remove node2 from load balancer
    module: network/haproxy
    hosts: load-balancer
    vars:
      command: disable_server
      backend: mysql-backend
      server: node2

  - name: Drain connections from node2
    module: system/sleep
    vars:
      seconds: 30

  - name: Stop MySQL on node2
    module: db/galeradb
    hosts: node2
    vars:
      command: leave_cluster
      mysql_service: mysql

  - name: Verify cluster size decreased
    module: db/galeradb
    hosts: node1
    vars:
      command: show_cluster_size
```

**Step 3: Perform Maintenance**

```yaml
  - name: Perform system updates
    module: system/apt
    hosts: node2
    vars:
      command: upgrade
      packages: all

  - name: Reboot if needed
    module: system/reboot
    hosts: node2
    when: facts.reboot_required == true
```

**Step 4: Return Node to Service**

```yaml
  - name: Rejoin node2 to cluster
    module: db/galeradb
    hosts: node2
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com
      mysql_service: mysql

  - name: Wait for node2 to sync
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_synced
    loop: "{{ range(1, 60) }}"
    until: facts.synced == true
    delay: 10

  - name: Verify node2 healthy
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_status

  - name: Add node2 back to load balancer
    module: network/haproxy
    hosts: load-balancer
    vars:
      command: enable_server
      backend: mysql-backend
      server: node2

  - name: Verify final cluster state
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status
```

### Scenario: Rolling restart of cluster

**Rolling Restart Stack:**

```yaml
# File: stacks/rolling_restart.ofy
inventory: galera-cluster

defaults:
  mysql_service: mysql

run:
  - name: Restart node1
    module: db/galeradb
    hosts: node1
    vars:
      command: restart_node

  - name: Wait for node1 sync
    module: db/galeradb
    hosts: node1
    vars:
      command: check_node_synced
    loop: "{{ range(1, 30) }}"
    until: facts.synced == true
    delay: 10

  - name: Restart node2
    module: db/galeradb
    hosts: node2
    vars:
      command: restart_node

  - name: Wait for node2 sync
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_synced
    loop: "{{ range(1, 30) }}"
    until: facts.synced == true
    delay: 10

  - name: Restart node3
    module: db/galeradb
    hosts: node3
    vars:
      command: restart_node

  - name: Wait for node3 sync
    module: db/galeradb
    hosts: node3
    vars:
      command: check_node_synced
    loop: "{{ range(1, 30) }}"
    until: facts.synced == true
    delay: 10

  - name: Verify all nodes healthy
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_node_status
```

---

## Performance Monitoring

### Scenario: Diagnose slow replication

**Performance Analysis Stack:**

```yaml
# File: stacks/performance_analysis.ofy
inventory: galera-cluster

defaults:
  galera_user: monitor
  galera_password: "{{ vault.monitor_password }}"

run:
  - name: Check flow control on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_flow_control

  - name: Identify nodes with flow control
    module: system/filter
    vars:
      results: "{{ previous.facts }}"
      condition: wsrep_flow_control_paused > 0.01

  - name: Check apply lag on problematic nodes
    module: db/galeradb
    hosts: "{{ filtered_hosts }}"
    vars:
      command: show_apply_lag

  - name: Check queue sizes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_queue_size

  - name: Analyze replication stats
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_replication_status

  - name: Check certification conflicts
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_conflicts

  - name: Get full wsrep status
    module: db/galeradb
    hosts: "{{ slowest_node }}"
    vars:
      command: show_wsrep_status
```

**Analysis Checklist:**

1. **Flow Control Active?**
   - wsrep_flow_control_paused > 0.01
   - Action: Increase wsrep_slave_threads

2. **High Receive Queue?**
   - wsrep_local_recv_queue > 100
   - Action: Check disk I/O, optimize queries

3. **High Send Queue?**
   - wsrep_local_send_queue > 100
   - Action: Check network, other nodes

4. **High Conflicts?**
   - wsrep_local_cert_failures increasing
   - Action: Review application transactions

---

## Troubleshooting

### Scenario: Node stuck in "Joining" state

**Diagnostic Stack:**

```yaml
# File: stacks/diagnose_joining.ofy
inventory: galera-cluster

run:
  - name: Check node state
    module: db/galeradb
    hosts: node2  # Stuck node
    vars:
      command: show_node_state

  - name: Check if donor available
    module: db/galeradb
    hosts: @group:other-nodes
    vars:
      command: show_node_state

  - name: Check disk space on joining node
    module: system/disk
    hosts: node2
    vars:
      command: check_usage
      path: /var/lib/mysql

  - name: Check disk space on potential donors
    module: system/disk
    hosts: @group:other-nodes
    vars:
      command: check_usage
      path: /var/lib/mysql

  - name: Check network connectivity
    module: network/ping
    hosts: node2
    vars:
      targets: "{{ other_nodes }}"

  - name: Check MySQL error log
    module: system/file
    hosts: node2
    vars:
      command: tail
      path: /var/log/mysql/error.log
      lines: 100
```

**Common Causes:**
1. Insufficient disk space (on donor or joiner)
2. Network issues between nodes
3. SST script problems
4. All other nodes in Donor state

**Resolution:**

```yaml
run:
  - name: Restart stuck node
    module: db/galeradb
    hosts: node2
    vars:
      command: restart_node
      mysql_service: mysql

  - name: Monitor state change
    module: db/galeradb
    hosts: node2
    vars:
      command: show_node_state
    loop: "{{ range(1, 60) }}"
    delay: 30
```

### Scenario: Node not accepting writes

**Diagnostic Stack:**

```yaml
run:
  - name: Check node ready status
    module: db/galeradb
    hosts: node2
    vars:
      command: check_node_status

  - name: Check node state
    module: db/galeradb
    hosts: node2
    vars:
      command: show_node_state

  - name: Check cluster status
    module: db/galeradb
    hosts: node2
    vars:
      command: show_cluster_status
```

**Possible Issues:**
1. Node in "Donor" state (read-only during SST)
2. wsrep_ready = OFF
3. Cluster status = Non-Primary (split-brain)

**Resolution depends on issue:**

```yaml
# If Donor state - wait for SST to complete
- name: Wait for Donor to finish
  module: db/galeradb
  hosts: node2
  vars:
    command: show_node_state
  loop: "{{ range(1, 60) }}"
  until: facts.state_comment == "Synced"
  delay: 30

# If Non-Primary - rejoin cluster
- name: Rejoin cluster
  module: db/galeradb
  hosts: node2
  vars:
    command: restart_node
```

---

## Split-Brain Recovery

### Scenario: Network partition caused split-brain

**Detection:**

```yaml
# File: stacks/detect_splitbrain.ofy
run:
  - name: Check cluster status on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status

  - name: Check cluster UUIDs
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_wsrep_status

  - name: Analyze results
    module: system/script
    vars:
      script: |
        # Check if multiple Primary clusters exist
        # or different cluster UUIDs
```

**Indicators of Split-Brain:**
- Multiple nodes with cluster_status = "Primary"
- Different cluster_uuid values
- Cluster_size doesn't match actual node count

**Recovery Process:**

```yaml
# File: stacks/recover_splitbrain.ofy
run:
  # Step 1: Stop all nodes
  - name: Stop all MySQL instances
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: leave_cluster
      mysql_service: mysql

  # Step 2: Check grastate on all nodes
  - name: Check seqno on all nodes
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: check_grastate
      grastate_file: /var/lib/mysql/grastate.dat

  # Step 3: Identify most advanced partition
  # Manual analysis needed here

  # Step 4: Bootstrap from most advanced node
  - name: Bootstrap from node with highest seqno
    module: db/galeradb
    hosts: node3  # Assuming node3 has highest seqno
    vars:
      command: recover_cluster
      grastate_file: /var/lib/mysql/grastate.dat
      mysql_service: mysql

  # Step 5: Rejoin all other nodes
  - name: Rejoin remaining nodes
    module: db/galeradb
    hosts: node1,node2
    vars:
      command: join_cluster
      cluster_address: gcomm://node1.example.com,node2.example.com,node3.example.com

  # Step 6: Verify recovery
  - name: Verify single Primary cluster
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_cluster_status

  - name: Verify all nodes have same UUID
    module: db/galeradb
    hosts: @group:all-nodes
    vars:
      command: show_wsrep_status
```

**Data Reconciliation:**

After recovery, check for data divergence:

```yaml
run:
  - name: Compare checksums across nodes
    module: db/checksum
    hosts: @group:all-nodes
    vars:
      databases: all

  - name: Identify inconsistencies
    module: system/compare
    vars:
      results: "{{ previous.facts }}"
```

---

## Best Practices

### Daily Operations

**Morning Health Check:**
```yaml
# Run every morning at 8 AM
0 8 * * * froyo apply /opt/openfroyo/stacks/galera_healthcheck.ofy
```

**Continuous Monitoring:**
```yaml
# Every 5 minutes
*/5 * * * * froyo apply /opt/openfroyo/stacks/galera_alerts.ofy
```

### Change Management

**Always:**
1. Check cluster health before changes
2. Verify quorum maintained (N/2 + 1 nodes)
3. One node at a time for changes
4. Wait for sync before next node
5. Document changes

**Never:**
1. Bootstrap multiple nodes simultaneously
2. Force changes without checking grastate
3. Perform maintenance on majority of nodes
4. Skip verification steps

### Backup Strategy

**Before Major Changes:**
```yaml
run:
  - name: Backup all nodes
    module: db/mysqldump
    hosts: @group:all-nodes
    vars:
      databases: all
      output: /backup/pre-change-{{ timestamp }}.sql
```

### Documentation

**Recovery Runbook:**
1. Keep printed copy of recovery procedures
2. Document node roles (which to bootstrap)
3. Maintain grastate.dat history
4. Log all cluster changes

---

## Emergency Procedures

### Quick Reference Card

**Cluster Won't Start:**
```bash
1. Check grastate.dat on all nodes
2. Find highest seqno
3. Bootstrap from that node
4. Join other nodes
```

**Node Stuck Joining:**
```bash
1. Check disk space
2. Check network
3. Check donor nodes
4. Restart if needed
```

**Split-Brain:**
```bash
1. Stop all nodes
2. Check grastate.dat
3. Choose most advanced
4. Bootstrap one node
5. Join others
```

**Performance Issues:**
```bash
1. Check flow control
2. Check queue sizes
3. Check conflicts
4. Tune wsrep_slave_threads
```

---

## See Also

- [README.md](README.md) - Module overview
- [COMMANDS.md](COMMANDS.md) - Detailed command reference
- [Galera Documentation](https://galeracluster.com/library/documentation/)
