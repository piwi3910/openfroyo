# PowerStore REST API Reference

Complete reference for Dell PowerStore REST API endpoints used by this module.

## API Overview

- **Base URL**: `https://<powerstore-ip>:443/api/rest`
- **Authentication**: HTTP Basic Auth
- **Content-Type**: `application/json`
- **API Version**: v3.0+

## Authentication

All API requests require HTTP Basic Authentication:

```bash
curl -X GET https://192.168.1.100/api/rest/volume \
  -u "admin:password" \
  -H "Content-Type: application/json"
```

## Volume Management APIs

### Create Volume

**Endpoint**: `POST /api/rest/volume`

**Request Body**:
```json
{
  "name": "my-volume",
  "size": 107374182400,
  "description": "Application data volume",
  "protection_policy_id": "policy-uuid",
  "performance_policy_id": "perf-policy-uuid"
}
```

**Response** (201 Created):
```json
{
  "id": "vol-12345678-1234-1234-1234-123456789012",
  "name": "my-volume",
  "size": 107374182400,
  "state": "Ready",
  "type": "Primary",
  "wwn": "naa.68ccf09800918a9b2e5e4f5e5e5e5e5e"
}
```

**OpenFroyo Usage**:
```yaml
command: create_volume
volume_name: "my-volume"
volume_size: 100  # GB
```

### Delete Volume

**Endpoint**: `DELETE /api/rest/volume/{volume_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
```

### Modify Volume

**Endpoint**: `PATCH /api/rest/volume/{volume_id}`

**Request Body**:
```json
{
  "name": "new-volume-name",
  "description": "Updated description",
  "size": 214748364800
}
```

**OpenFroyo Usage**:
```yaml
command: modify_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
volume_name: "new-volume-name"
description: "Updated description"
```

### Clone Volume

**Endpoint**: `POST /api/rest/volume/{volume_id}/clone`

**Request Body**:
```json
{
  "name": "volume-clone",
  "description": "Cloned volume for testing"
}
```

**Response** (201 Created):
```json
{
  "id": "vol-87654321-4321-4321-4321-210987654321",
  "name": "volume-clone"
}
```

**OpenFroyo Usage**:
```yaml
command: clone_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
clone_name: "volume-clone"
```

### Resize Volume

**Endpoint**: `PATCH /api/rest/volume/{volume_id}`

**Request Body**:
```json
{
  "size": 214748364800
}
```

**Note**: Only expansion is supported; volumes cannot be shrunk.

**OpenFroyo Usage**:
```yaml
command: resize_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
new_size: 200  # GB
```

### Map Volume to Host

**Endpoint**: `POST /api/rest/volume/{volume_id}/attach`

**Request Body**:
```json
{
  "host_id": "host-12345678-1234-1234-1234-123456789012",
  "logical_unit_number": 0
}
```

**OpenFroyo Usage**:
```yaml
command: map_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
host_id: "host-12345678-1234-1234-1234-123456789012"
```

### Unmap Volume from Host

**Endpoint**: `POST /api/rest/volume/{volume_id}/detach`

**Request Body**:
```json
{
  "host_id": "host-12345678-1234-1234-1234-123456789012"
}
```

**OpenFroyo Usage**:
```yaml
command: unmap_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
host_id: "host-12345678-1234-1234-1234-123456789012"
```

### Get Volume

**Endpoint**: `GET /api/rest/volume/{volume_id}`

**Response** (200 OK):
```json
{
  "id": "vol-12345678-1234-1234-1234-123456789012",
  "name": "my-volume",
  "size": 107374182400,
  "state": "Ready",
  "type": "Primary",
  "wwn": "naa.68ccf09800918a9b2e5e4f5e5e5e5e5e",
  "appliance_id": "app-12345678",
  "host_mappings": []
}
```

**OpenFroyo Usage**:
```yaml
command: show_volume
volume_id: "vol-12345678-1234-1234-1234-123456789012"
```

### List Volumes

**Endpoint**: `GET /api/rest/volume`

**Query Parameters**:
- `select`: Fields to include (default: summary fields)
- `limit`: Maximum results to return
- `offset`: Pagination offset

**Response** (200 OK):
```json
[
  {
    "id": "vol-12345678",
    "name": "volume-1",
    "size": 107374182400
  },
  {
    "id": "vol-87654321",
    "name": "volume-2",
    "size": 214748364800
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_volumes
```

### Get Volume Details

**Endpoint**: `GET /api/rest/volume/{volume_id}?select=*`

**Response** (200 OK): Complete volume details including all attributes

**OpenFroyo Usage**:
```yaml
command: show_volume_details
volume_id: "vol-12345678-1234-1234-1234-123456789012"
```

## Host Management APIs

### Create Host

**Endpoint**: `POST /api/rest/host`

**Request Body**:
```json
{
  "name": "esxi-host-01",
  "os_type": "ESXi",
  "description": "VMware ESXi host",
  "initiators": [
    {
      "port_name": "iqn.1998-01.com.vmware:esxi-host-01-12345678",
      "port_type": "iSCSI"
    },
    {
      "port_name": "21:00:00:24:ff:12:34:56",
      "port_type": "FC"
    }
  ]
}
```

**Response** (201 Created):
```json
{
  "id": "host-12345678-1234-1234-1234-123456789012",
  "name": "esxi-host-01",
  "os_type": "ESXi"
}
```

**OS Types**:
- `Linux`
- `Windows`
- `ESXi`
- `AIX`
- `HP-UX`
- `Solaris`

**OpenFroyo Usage**:
```yaml
command: create_host
host_name: "esxi-host-01"
os_type: "ESXi"
initiators:
  - "iqn.1998-01.com.vmware:esxi-host-01-12345678"
  - "21:00:00:24:ff:12:34:56"
```

### Delete Host

**Endpoint**: `DELETE /api/rest/host/{host_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_host
host_id: "host-12345678-1234-1234-1234-123456789012"
```

### Modify Host

**Endpoint**: `PATCH /api/rest/host/{host_id}`

**Request Body**:
```json
{
  "name": "new-host-name",
  "description": "Updated description",
  "os_type": "Linux"
}
```

**OpenFroyo Usage**:
```yaml
command: modify_host
host_id: "host-12345678-1234-1234-1234-123456789012"
host_name: "new-host-name"
os_type: "Linux"
```

### Add Initiator to Host

**Endpoint**: `PATCH /api/rest/host/{host_id}`

**Request Body**:
```json
{
  "add": [
    {
      "port_name": "iqn.1993-08.org.debian:01:new-initiator"
    }
  ]
}
```

**OpenFroyo Usage**:
```yaml
command: add_initiator
host_id: "host-12345678-1234-1234-1234-123456789012"
initiator: "iqn.1993-08.org.debian:01:new-initiator"
```

### Remove Initiator from Host

**Endpoint**: `PATCH /api/rest/host/{host_id}`

**Request Body**:
```json
{
  "remove": [
    {
      "port_name": "iqn.1993-08.org.debian:01:old-initiator"
    }
  ]
}
```

**OpenFroyo Usage**:
```yaml
command: remove_initiator
host_id: "host-12345678-1234-1234-1234-123456789012"
initiator: "iqn.1993-08.org.debian:01:old-initiator"
```

### List Hosts

**Endpoint**: `GET /api/rest/host`

**Response** (200 OK):
```json
[
  {
    "id": "host-12345678",
    "name": "esxi-host-01",
    "os_type": "ESXi"
  },
  {
    "id": "host-87654321",
    "name": "linux-host-01",
    "os_type": "Linux"
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_hosts
```

## Host Group Management APIs

### Create Host Group

**Endpoint**: `POST /api/rest/host_group`

**Request Body**:
```json
{
  "name": "vsphere-cluster",
  "description": "VMware vSphere cluster hosts"
}
```

**Response** (201 Created):
```json
{
  "id": "hg-12345678-1234-1234-1234-123456789012",
  "name": "vsphere-cluster"
}
```

**OpenFroyo Usage**:
```yaml
command: create_host_group
host_group_name: "vsphere-cluster"
```

### Delete Host Group

**Endpoint**: `DELETE /api/rest/host_group/{host_group_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_host_group
host_group_id: "hg-12345678-1234-1234-1234-123456789012"
```

### Add Host to Group

**Endpoint**: `PATCH /api/rest/host_group/{host_group_id}`

**Request Body**:
```json
{
  "add": ["host-12345678-1234-1234-1234-123456789012"]
}
```

**OpenFroyo Usage**:
```yaml
command: add_host_to_group
host_group_id: "hg-12345678-1234-1234-1234-123456789012"
host_id: "host-12345678-1234-1234-1234-123456789012"
```

### List Host Groups

**Endpoint**: `GET /api/rest/host_group`

**Response** (200 OK):
```json
[
  {
    "id": "hg-12345678",
    "name": "vsphere-cluster",
    "host_count": 4
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_host_groups
```

## Snapshot Management APIs

### Create Snapshot

**Endpoint**: `POST /api/rest/volume_snapshot`

**Request Body**:
```json
{
  "name": "backup-snapshot-001",
  "volume_id": "vol-12345678-1234-1234-1234-123456789012",
  "description": "Pre-upgrade backup"
}
```

**Response** (201 Created):
```json
{
  "id": "snap-12345678-1234-1234-1234-123456789012",
  "name": "backup-snapshot-001",
  "creation_timestamp": "2024-01-15T10:30:00Z"
}
```

**OpenFroyo Usage**:
```yaml
command: create_snapshot
volume_id: "vol-12345678-1234-1234-1234-123456789012"
snapshot_name: "backup-snapshot-001"
```

### Delete Snapshot

**Endpoint**: `DELETE /api/rest/volume_snapshot/{snapshot_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_snapshot
snapshot_id: "snap-12345678-1234-1234-1234-123456789012"
```

### Restore Snapshot

**Endpoint**: `POST /api/rest/volume_snapshot/{snapshot_id}/restore`

**Request Body**:
```json
{
  "backup": true
}
```

**OpenFroyo Usage**:
```yaml
command: restore_snapshot
snapshot_id: "snap-12345678-1234-1234-1234-123456789012"
```

### Create Snapshot Rule

**Endpoint**: `POST /api/rest/snapshot_rule`

**Request Body**:
```json
{
  "name": "daily-snapshots",
  "desired_retention": 168,
  "interval": "Every_24_Hours",
  "time_of_day": "02:00"
}
```

**Interval Options**:
- `Five_Minutes`
- `Fifteen_Minutes`
- `Thirty_Minutes`
- `One_Hour`
- `Two_Hours`
- `Three_Hours`
- `Four_Hours`
- `Six_Hours`
- `Eight_Hours`
- `Twelve_Hours`
- `Every_24_Hours`

**Response** (201 Created):
```json
{
  "id": "rule-12345678-1234-1234-1234-123456789012",
  "name": "daily-snapshots"
}
```

**OpenFroyo Usage**:
```yaml
command: create_snapshot_rule
snapshot_rule_name: "daily-snapshots"
desired_retention: 168  # hours (7 days)
interval: "Every_24_Hours"
```

### Delete Snapshot Rule

**Endpoint**: `DELETE /api/rest/snapshot_rule/{snapshot_rule_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_snapshot_rule
snapshot_rule_id: "rule-12345678-1234-1234-1234-123456789012"
```

### List Snapshots

**Endpoint**: `GET /api/rest/volume_snapshot`

**Query Parameters**:
- `volume_id`: Filter by volume

**Response** (200 OK):
```json
[
  {
    "id": "snap-12345678",
    "name": "snapshot-1",
    "creation_timestamp": "2024-01-15T10:30:00Z",
    "size": 107374182400
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_snapshots
volume_id: "vol-12345678-1234-1234-1234-123456789012"  # Optional
```

## Performance Metrics APIs

### Generate Volume Metrics

**Endpoint**: `POST /api/rest/metrics/generate?entity=volume&entity_id={volume_id}`

**Query Parameters**:
- `entity`: Entity type (`volume`, `appliance`, `node`, `cluster`)
- `entity_id`: UUID of the entity
- `interval`: Data granularity (`Five_Mins`, `One_Hour`, `One_Day`)

**Response** (200 OK):
```json
{
  "result": [
    {
      "timestamp": "2024-01-15T10:00:00Z",
      "avg_read_iops": 1234.5,
      "avg_write_iops": 567.8,
      "avg_read_bandwidth": 52428800,
      "avg_write_bandwidth": 26214400,
      "avg_latency": 2.5
    }
  ]
}
```

**OpenFroyo Usage**:
```yaml
command: show_volume_metrics
volume_id: "vol-12345678-1234-1234-1234-123456789012"
interval: "Five_Mins"
```

### Generate Appliance Metrics

**Endpoint**: `POST /api/rest/metrics/generate?entity=appliance&entity_id={appliance_id}`

**OpenFroyo Usage**:
```yaml
command: show_appliance_metrics
appliance_id: "app-12345678"
interval: "Five_Mins"
```

### Generate Node Metrics

**Endpoint**: `POST /api/rest/metrics/generate?entity=node&entity_id={node_id}`

**OpenFroyo Usage**:
```yaml
command: show_node_metrics
node_id: "node-12345678"
interval: "Five_Mins"
```

### Generate Cluster Metrics

**Endpoint**: `POST /api/rest/metrics/generate?entity=cluster&entity_id={cluster_id}`

**OpenFroyo Usage**:
```yaml
command: show_cluster_metrics
cluster_id: "cluster-12345678"
interval: "Five_Mins"
```

## Replication APIs

### Create Replication Session

**Endpoint**: `POST /api/rest/replication_session`

**Request Body**:
```json
{
  "local_resource_id": "vol-12345678-1234-1234-1234-123456789012",
  "remote_system_id": "remote-12345678-1234-1234-1234-123456789012",
  "replication_rule": "async-replication-4hr",
  "state": "OK"
}
```

**Response** (201 Created):
```json
{
  "id": "repl-12345678-1234-1234-1234-123456789012",
  "local_resource_id": "vol-12345678",
  "remote_system_id": "remote-12345678",
  "state": "OK"
}
```

**OpenFroyo Usage**:
```yaml
command: create_replication_session
volume_id: "vol-12345678-1234-1234-1234-123456789012"
remote_system_id: "remote-12345678-1234-1234-1234-123456789012"
replication_rule: "async-replication-4hr"
```

### Delete Replication Session

**Endpoint**: `DELETE /api/rest/replication_session/{replication_session_id}`

**Response** (204 No Content)

**OpenFroyo Usage**:
```yaml
command: delete_replication_session
replication_session_id: "repl-12345678-1234-1234-1234-123456789012"
```

### Pause Replication

**Endpoint**: `PATCH /api/rest/replication_session/{replication_session_id}`

**Request Body**:
```json
{
  "pause": true
}
```

**OpenFroyo Usage**:
```yaml
command: pause_replication
replication_session_id: "repl-12345678-1234-1234-1234-123456789012"
```

### Resume Replication

**Endpoint**: `PATCH /api/rest/replication_session/{replication_session_id}`

**Request Body**:
```json
{
  "pause": false
}
```

**OpenFroyo Usage**:
```yaml
command: resume_replication
replication_session_id: "repl-12345678-1234-1234-1234-123456789012"
```

### List Replication Sessions

**Endpoint**: `GET /api/rest/replication_session`

**Response** (200 OK):
```json
[
  {
    "id": "repl-12345678",
    "local_resource_id": "vol-12345678",
    "remote_system_id": "remote-12345678",
    "state": "OK",
    "progress_percentage": 100
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_replication_sessions
```

## System Information APIs

### Get Cluster Information

**Endpoint**: `GET /api/rest/cluster`

**Response** (200 OK):
```json
[
  {
    "id": "cluster-12345678",
    "name": "PowerStore-Cluster-01",
    "management_address": "192.168.1.100",
    "appliance_count": 2,
    "state": "Configured",
    "compatibility_level": 10
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_cluster_info
```

### List Appliances

**Endpoint**: `GET /api/rest/appliance`

**Response** (200 OK):
```json
[
  {
    "id": "app-12345678",
    "name": "Appliance-A",
    "service_tag": "ABCD123",
    "model": "PowerStore 3000T",
    "node_count": 2
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_appliances
```

### List Nodes

**Endpoint**: `GET /api/rest/node`

**Response** (200 OK):
```json
[
  {
    "id": "node-12345678",
    "name": "Node-A-0",
    "appliance_id": "app-12345678",
    "slot": 0,
    "is_master": true
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_nodes
```

### Get Capacity Information

**Endpoint**: `GET /api/rest/capacity`

**Response** (200 OK):
```json
{
  "physical_total": 107374182400000,
  "physical_used": 53687091200000,
  "logical_used": 80530636800000,
  "data_reduction": 1.5,
  "snapshot_savings": 0.25
}
```

**OpenFroyo Usage**:
```yaml
command: show_capacity
```

### List Alerts

**Endpoint**: `GET /api/rest/alert`

**Query Parameters**:
- `severity`: Filter by severity (`Info`, `Minor`, `Major`, `Critical`)
- `state`: Filter by state (`Active`, `Cleared`)

**Response** (200 OK):
```json
[
  {
    "id": "alert-12345678",
    "severity": "Minor",
    "state": "Active",
    "message": "Disk usage above 80%",
    "timestamp": "2024-01-15T10:00:00Z"
  }
]
```

**OpenFroyo Usage**:
```yaml
command: show_alerts
```

## Error Responses

### Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **204 No Content**: Delete successful
- **400 Bad Request**: Invalid request parameters
- **401 Unauthorized**: Authentication failed
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource does not exist
- **409 Conflict**: Resource conflict (e.g., name already exists)
- **500 Internal Server Error**: Server-side error

### Error Response Format

```json
{
  "messages": [
    {
      "code": "0xE0A08001000F",
      "severity": "Error",
      "message_l10n": "The specified volume name already exists.",
      "arguments": ["my-volume"]
    }
  ]
}
```

## Rate Limiting

PowerStore implements API rate limiting:

- **Per-user limit**: 100 requests/minute
- **Burst allowance**: 200 requests
- **Rate limit headers**:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

When rate limit is exceeded:
- **Status Code**: 429 Too Many Requests
- **Retry-After**: Seconds to wait before retrying

## API Versioning

PowerStore REST API uses URI versioning:

- Current: `/api/rest` (v3.x)
- Legacy: `/api/rest/v1` (deprecated)

Always use the latest version for new implementations.
