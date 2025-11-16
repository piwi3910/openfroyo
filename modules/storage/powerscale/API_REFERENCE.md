# Dell PowerScale Platform API (PAPI) Reference

## Overview

This document provides detailed reference information for the Dell PowerScale Platform API endpoints used by the OpenFroyo PowerScale module.

## API Architecture

### Base URL
```
https://<powerscale-cluster-ip>:8080/platform/
```

### Authentication
All API requests use HTTP Basic Authentication:

```
Authorization: Basic <base64-encoded-credentials>
```

Where credentials are: `username:password` encoded in Base64.

### API Versioning

PowerScale PAPI uses version-specific endpoints:

```
/platform/<version>/<resource>
```

Example versions:
- `/platform/1/` - Foundation APIs (quotas, snapshots, statistics)
- `/platform/3/` - Core APIs (cluster, zones, sync policies)
- `/platform/4/` - Protocol APIs (NFS)
- `/platform/7/` - Advanced protocol APIs (SMB)

### Response Format

All responses are JSON formatted:

**Success Response:**
```json
{
  "id": "resource_id",
  "name": "resource_name",
  ...resource-specific fields...
}
```

**Error Response:**
```json
{
  "errors": [
    {
      "code": "AEC_ERROR_CODE",
      "message": "Detailed error message",
      "field": "field_name"
    }
  ]
}
```

### HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 200  | OK | Successful GET request |
| 201  | Created | Successful POST (creation) |
| 204  | No Content | Successful DELETE |
| 400  | Bad Request | Invalid parameters |
| 401  | Unauthorized | Authentication failed |
| 403  | Forbidden | Insufficient permissions |
| 404  | Not Found | Resource doesn't exist |
| 409  | Conflict | Resource already exists |
| 500  | Internal Server Error | Server-side error |

## API Endpoints by Category

### SMB Share Management

#### GET /platform/7/protocols/smb/shares

List all SMB shares.

**Query Parameters:**
- `zone` (string, optional): Access zone name (default: "System")
- `resume` (string, optional): Token for pagination
- `limit` (integer, optional): Maximum results per request

**Example Request:**
```bash
curl -X GET "https://192.168.1.100:8080/platform/7/protocols/smb/shares?zone=System" \
  -u "admin:password" \
  -H "Accept: application/json"
```

**Example Response:**
```json
{
  "shares": [
    {
      "id": "share_name",
      "name": "share_name",
      "path": "/ifs/data/share",
      "description": "Share description",
      "browsable": true,
      "ntfs_acl_support": true,
      "permissions": [
        {
          "trustee": {
            "name": "Everyone",
            "type": "wellknown"
          },
          "permission": "read",
          "permission_type": "allow"
        }
      ]
    }
  ]
}
```

#### GET /platform/7/protocols/smb/shares/{share-name}

Retrieve a specific SMB share.

**Path Parameters:**
- `share-name` (string, required): Name of the share

**Query Parameters:**
- `zone` (string, optional): Access zone

**Example Request:**
```bash
curl -X GET "https://192.168.1.100:8080/platform/7/protocols/smb/shares/data_share" \
  -u "admin:password"
```

#### POST /platform/7/protocols/smb/shares

Create a new SMB share.

**Query Parameters:**
- `zone` (string, optional): Access zone

**Request Body:**
```json
{
  "name": "share_name",
  "path": "/ifs/data/share",
  "description": "Share description",
  "browsable": true,
  "ntfs_acl_support": true,
  "access_based_enumeration": false,
  "permissions": [
    {
      "trustee": {
        "name": "Users",
        "type": "group"
      },
      "permission": "full",
      "permission_type": "allow"
    }
  ]
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/7/protocols/smb/shares" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "engineering",
    "path": "/ifs/data/engineering",
    "description": "Engineering share",
    "browsable": true,
    "ntfs_acl_support": true
  }'
```

**Response:**
```json
{
  "id": "engineering"
}
```

#### PUT /platform/7/protocols/smb/shares/{share-name}

Modify an existing SMB share.

**Path Parameters:**
- `share-name` (string, required): Share name

**Query Parameters:**
- `zone` (string, optional): Access zone

**Request Body:**
```json
{
  "description": "Updated description",
  "browsable": false,
  "permissions": [...]
}
```

#### DELETE /platform/7/protocols/smb/shares/{share-name}

Delete an SMB share.

**Path Parameters:**
- `share-name` (string, required): Share name

**Query Parameters:**
- `zone` (string, optional): Access zone

**Example Request:**
```bash
curl -X DELETE "https://192.168.1.100:8080/platform/7/protocols/smb/shares/old_share" \
  -u "admin:password"
```

#### GET /platform/3/protocols/smb/sessions

List active SMB sessions.

**Example Response:**
```json
{
  "sessions": [
    {
      "computer": "WORKSTATION01",
      "user": "DOMAIN\\user",
      "opens": 5,
      "idle_time": 120,
      "guest": false
    }
  ]
}
```

### NFS Export Management

#### GET /platform/4/protocols/nfs/exports

List all NFS exports.

**Query Parameters:**
- `zone` (string, optional): Access zone
- `path` (string, optional): Filter by path
- `resume` (string, optional): Pagination token

**Example Response:**
```json
{
  "exports": [
    {
      "id": 1,
      "paths": ["/ifs/data/export"],
      "clients": ["192.168.1.0/24"],
      "read_only": false,
      "root_clients": ["192.168.1.100"],
      "security_flavors": ["unix"]
    }
  ]
}
```

#### GET /platform/4/protocols/nfs/exports/{id}

Retrieve specific NFS export.

**Path Parameters:**
- `id` (integer, required): Export ID

#### POST /platform/4/protocols/nfs/exports

Create a new NFS export.

**Query Parameters:**
- `zone` (string, optional): Access zone

**Request Body:**
```json
{
  "paths": ["/ifs/data/export"],
  "clients": ["192.168.1.0/24", "10.0.0.50"],
  "read_only": false,
  "root_clients": ["192.168.1.100"],
  "security_flavors": ["unix"],
  "map_root": {
    "enabled": true,
    "user": "nobody"
  },
  "map_all": {
    "enabled": false
  }
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/4/protocols/nfs/exports" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "paths": ["/ifs/data/nfs_share"],
    "clients": ["192.168.1.0/24"],
    "read_only": false
  }'
```

#### PUT /platform/4/protocols/nfs/exports/{id}

Modify an NFS export.

**Path Parameters:**
- `id` (integer, required): Export ID

**Request Body:**
```json
{
  "clients": ["192.168.1.0/24", "192.168.2.0/24"],
  "read_only": true
}
```

#### DELETE /platform/4/protocols/nfs/exports/{id}

Delete an NFS export.

**Path Parameters:**
- `id` (integer, required): Export ID

#### POST /platform/4/protocols/nfs/reload

Reload NFS exports configuration.

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/4/protocols/nfs/reload" \
  -u "admin:password"
```

### Access Zone Management

#### GET /platform/3/zones

List all access zones.

**Example Response:**
```json
{
  "zones": [
    {
      "name": "System",
      "path": "/ifs",
      "groupnet": "groupnet0",
      "auth_providers": ["lsa-local-provider:System"]
    },
    {
      "name": "production",
      "path": "/ifs/zones/production",
      "groupnet": "groupnet0",
      "auth_providers": ["lsa-local-provider:production"]
    }
  ]
}
```

#### GET /platform/3/zones/{zone-name}

Retrieve specific access zone.

**Path Parameters:**
- `zone-name` (string, required): Zone name

#### POST /platform/3/zones

Create a new access zone.

**Request Body:**
```json
{
  "name": "zone_name",
  "path": "/ifs/zones/zone_name",
  "groupnet": "groupnet0",
  "auth_providers": ["lsa-local-provider:zone_name"]
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/3/zones" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "tenant_a",
    "path": "/ifs/zones/tenant_a",
    "groupnet": "groupnet0"
  }'
```

#### PUT /platform/3/zones/{zone-name}

Modify an access zone.

#### DELETE /platform/3/zones/{zone-name}

Delete an access zone.

**Note:** Cannot delete the System zone.

### Quota Management

#### GET /platform/1/quota/quotas

List all quotas.

**Query Parameters:**
- `type` (string, optional): Filter by type (directory, user, group)
- `path` (string, optional): Filter by path
- `resume` (string, optional): Pagination token

**Example Response:**
```json
{
  "quotas": [
    {
      "id": "quota_id",
      "type": "directory",
      "path": "/ifs/data/share",
      "enforced": true,
      "thresholds": {
        "hard": 1099511627776,
        "soft": 966367641600,
        "advisory": 858993459200
      },
      "usage": {
        "logical": 524288000000,
        "physical": 550502400000
      }
    }
  ]
}
```

#### GET /platform/1/quota/quotas/{quota-id}

Retrieve specific quota.

**Path Parameters:**
- `quota-id` (string, required): Quota ID

#### POST /platform/1/quota/quotas

Create a new quota.

**Request Body:**
```json
{
  "path": "/ifs/data/share",
  "type": "directory",
  "enforced": true,
  "thresholds": {
    "hard": 1099511627776,
    "soft": 966367641600,
    "advisory": 858993459200
  }
}
```

**For user quota:**
```json
{
  "path": "/ifs/home",
  "type": "user",
  "persona": {
    "name": "jdoe",
    "type": "user"
  },
  "enforced": true,
  "thresholds": {
    "hard": 107374182400
  }
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/1/quota/quotas" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/ifs/data/engineering",
    "type": "directory",
    "enforced": true,
    "thresholds": {
      "hard": 1099511627776
    }
  }'
```

#### PUT /platform/1/quota/quotas/{quota-id}

Modify a quota.

**Request Body:**
```json
{
  "enforced": true,
  "thresholds": {
    "hard": 2199023255552,
    "soft": 1979120230000
  }
}
```

#### DELETE /platform/1/quota/quotas/{quota-id}

Delete a quota.

#### GET /platform/1/quota/reports

Retrieve quota reports with usage information.

**Query Parameters:**
- `type` (string, optional): Report type
- `path` (string, optional): Path filter

### Snapshot Management

#### GET /platform/1/snapshot/snapshots

List all snapshots.

**Query Parameters:**
- `schedule` (string, optional): Filter by schedule
- `type` (string, optional): Filter by type
- `resume` (string, optional): Pagination token

**Example Response:**
```json
{
  "snapshots": [
    {
      "id": 12345,
      "name": "snapshot_name",
      "path": "/ifs/data",
      "created": 1705420800,
      "size": 1073741824,
      "alias": "latest",
      "expires": 1706025600
    }
  ]
}
```

#### GET /platform/1/snapshot/snapshots/{snapshot-id}

Retrieve specific snapshot.

**Path Parameters:**
- `snapshot-id` (integer, required): Snapshot ID or name

#### POST /platform/1/snapshot/snapshots

Create a snapshot.

**Request Body:**
```json
{
  "path": "/ifs/data",
  "name": "snapshot_name",
  "alias": "latest"
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/1/snapshot/snapshots" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/ifs/data/critical",
    "name": "pre_upgrade_snapshot",
    "alias": "before_maintenance"
  }'
```

**Response:**
```json
{
  "id": 12346
}
```

#### DELETE /platform/1/snapshot/snapshots/{snapshot-id}

Delete a snapshot.

**Path Parameters:**
- `snapshot-id` (integer, required): Snapshot ID or name

#### GET /platform/1/snapshot/schedules

List snapshot schedules.

**Example Response:**
```json
{
  "schedules": [
    {
      "id": 1,
      "name": "daily_backup",
      "path": "/ifs/data",
      "schedule": "0 2 * * *",
      "pattern": "daily-%Y%m%d",
      "retention": "7 days"
    }
  ]
}
```

#### POST /platform/1/snapshot/schedules

Create a snapshot schedule.

**Request Body:**
```json
{
  "name": "schedule_name",
  "path": "/ifs/data",
  "schedule": "0 2 * * *",
  "pattern": "daily-%Y%m%d",
  "duration": "1 week",
  "retention": "30 days"
}
```

#### DELETE /platform/1/snapshot/schedules/{schedule-id}

Delete a snapshot schedule.

### Cluster Management

#### GET /platform/3/cluster/config

Retrieve cluster configuration.

**Example Response:**
```json
{
  "config": {
    "name": "powerscale-cluster",
    "description": "Production PowerScale Cluster",
    "guid": "12345678-1234-1234-1234-123456789abc",
    "encoding": "utf-8"
  }
}
```

#### GET /platform/3/cluster/nodes

List all cluster nodes.

**Example Response:**
```json
{
  "nodes": [
    {
      "id": 1,
      "lnn": 1,
      "type": "storage",
      "class": "h500",
      "status": "online",
      "drives": 12,
      "serial_number": "SN123456"
    }
  ]
}
```

#### GET /platform/3/cluster/version

Retrieve OneFS version information.

**Example Response:**
```json
{
  "version": {
    "release": "9.5.0.0",
    "type": "Isilon OneFS",
    "build": "B_9_5_0_0(RELEASE)",
    "revision": "12345"
  }
}
```

#### GET /platform/1/statistics/current

Retrieve current cluster statistics.

**Query Parameters:**
- `key` (string, required): Statistics key(s) to retrieve

**Keys for Capacity:**
- `cluster.disk.bytes.used`
- `cluster.disk.bytes.total`
- `cluster.disk.bytes.avail`

**Keys for Performance:**
- `cluster.cpu.sys.avg`
- `cluster.disk.bytes.in.rate`
- `cluster.disk.bytes.out.rate`
- `node.ifs.ops.in.rate`
- `node.ifs.ops.out.rate`

**Example Request:**
```bash
curl -X GET "https://192.168.1.100:8080/platform/1/statistics/current?key=cluster.disk.bytes.used&key=cluster.disk.bytes.total" \
  -u "admin:password"
```

**Example Response:**
```json
{
  "stats": [
    {
      "key": "cluster.disk.bytes.used",
      "value": 5497558138880,
      "time": 1705420800
    },
    {
      "key": "cluster.disk.bytes.total",
      "value": 10995116277760,
      "time": 1705420800
    }
  ]
}
```

#### GET /platform/3/cluster/diagnostics

Retrieve cluster health diagnostics.

**Example Response:**
```json
{
  "diagnostics": {
    "status": "OK",
    "degraded": false,
    "node_count": 4,
    "nodes_online": 4,
    "drives_total": 48,
    "drives_failed": 0
  }
}
```

### User/Group Management

#### GET /platform/1/auth/users

List all users.

**Query Parameters:**
- `zone` (string, optional): Access zone
- `resume` (string, optional): Pagination token

**Example Response:**
```json
{
  "users": [
    {
      "uid": {
        "id": "UID:5001",
        "name": "jdoe",
        "type": "user"
      },
      "gid": {
        "id": "GID:100",
        "name": "users"
      },
      "home_directory": "/ifs/home/jdoe",
      "shell": "/bin/bash",
      "enabled": true
    }
  ]
}
```

#### POST /platform/1/auth/users

Create a new user.

**Query Parameters:**
- `zone` (string, optional): Access zone

**Request Body:**
```json
{
  "name": "jdoe",
  "uid": 5001,
  "primary_group": "users",
  "home_directory": "/ifs/home/jdoe",
  "shell": "/bin/bash",
  "enabled": true
}
```

#### DELETE /platform/1/auth/users/{user-name}

Delete a user.

#### GET /platform/1/auth/groups

List all groups.

**Example Response:**
```json
{
  "groups": [
    {
      "gid": {
        "id": "GID:6001",
        "name": "developers",
        "type": "group"
      },
      "members": ["jdoe", "asmith"]
    }
  ]
}
```

#### POST /platform/1/auth/groups

Create a new group.

**Request Body:**
```json
{
  "name": "developers",
  "gid": 6001
}
```

#### DELETE /platform/1/auth/groups/{group-name}

Delete a group.

### SyncIQ Replication

#### GET /platform/3/sync/policies

List all SyncIQ policies.

**Example Response:**
```json
{
  "policies": [
    {
      "id": "policy_id",
      "name": "dr_replication",
      "source_root_path": "/ifs/data",
      "target_host": "192.168.100.50",
      "target_path": "/ifs/dr/data",
      "action": "sync",
      "enabled": true,
      "schedule": "when-source-modified",
      "last_job_state": "finished"
    }
  ]
}
```

#### GET /platform/3/sync/policies/{policy-name}

Retrieve specific SyncIQ policy.

#### POST /platform/3/sync/policies

Create a SyncIQ policy.

**Request Body:**
```json
{
  "name": "policy_name",
  "source_root_path": "/ifs/data/source",
  "target_host": "target-cluster.example.com",
  "target_path": "/ifs/data/target",
  "action": "sync",
  "enabled": true,
  "schedule": "when-source-modified",
  "snapshot_sync_pattern": "*",
  "rpo_alert": 3600
}
```

**Example Request:**
```bash
curl -X POST "https://192.168.1.100:8080/platform/3/sync/policies" \
  -u "admin:password" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod_to_dr",
    "source_root_path": "/ifs/data/production",
    "target_host": "192.168.100.50",
    "target_path": "/ifs/dr/production",
    "action": "sync",
    "enabled": true
  }'
```

#### PUT /platform/3/sync/policies/{policy-name}

Modify a SyncIQ policy.

#### DELETE /platform/3/sync/policies/{policy-name}

Delete a SyncIQ policy.

#### POST /platform/3/sync/jobs

Start a SyncIQ job manually.

**Request Body:**
```json
{
  "id": "policy_name",
  "action": "run"
}
```

#### GET /platform/3/sync/reports

List SyncIQ reports.

**Example Response:**
```json
{
  "reports": [
    {
      "id": "report_id",
      "policy_name": "dr_replication",
      "state": "finished",
      "start_time": 1705420800,
      "end_time": 1705424400,
      "bytes_transferred": 10737418240
    }
  ]
}
```

## Error Handling

### Common Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| AEC_NOT_FOUND | Resource not found | Verify resource name/ID |
| AEC_ALREADY_EXISTS | Resource already exists | Use different name or modify existing |
| AEC_BAD_REQUEST | Invalid request | Check request format and parameters |
| AEC_UNAUTHORIZED | Authentication failed | Verify credentials |
| AEC_FORBIDDEN | Insufficient permissions | Check RBAC roles |
| AEC_CONFLICT | Resource conflict | Resolve conflicting resources |
| AEC_QUOTA_EXCEEDED | Quota limit reached | Increase quota or delete resources |

### Error Response Example

```json
{
  "errors": [
    {
      "code": "AEC_ALREADY_EXISTS",
      "message": "Share 'data_share' already exists in access zone 'System'",
      "field": "name",
      "status": 409
    }
  ]
}
```

## Rate Limiting

PowerScale PAPI implements rate limiting to prevent API abuse:

- **Default Limit:** 1000 requests per minute per IP
- **Burst Limit:** 2000 requests per minute
- **Rate Limit Headers:**
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Requests remaining
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

When rate limited:
```json
{
  "errors": [
    {
      "code": "AEC_RATE_LIMIT_EXCEEDED",
      "message": "Rate limit exceeded. Try again in 60 seconds.",
      "status": 429
    }
  ]
}
```

## Pagination

For endpoints returning large datasets, use pagination:

**Request:**
```bash
curl -X GET "https://192.168.1.100:8080/platform/7/protocols/smb/shares?limit=100" \
  -u "admin:password"
```

**Response:**
```json
{
  "shares": [...],
  "resume": "eyJsYXN0X2lkIjogMTAwfQ=="
}
```

**Next Page:**
```bash
curl -X GET "https://192.168.1.100:8080/platform/7/protocols/smb/shares?limit=100&resume=eyJsYXN0X2lkIjogMTAwfQ==" \
  -u "admin:password"
```

## Best Practices

1. **Use HTTPS:** Always use HTTPS for API calls
2. **Validate Certificates:** Enable certificate validation in production
3. **Error Handling:** Implement robust error handling for all API calls
4. **Rate Limiting:** Respect rate limits and implement backoff strategies
5. **Pagination:** Use pagination for large datasets
6. **Idempotency:** Design operations to be idempotent where possible
7. **Atomic Operations:** Use atomic operations for critical changes
8. **Monitoring:** Monitor API response times and error rates

## Additional Resources

- [Dell PowerScale API Documentation](https://www.dell.com/support/kbdoc/en-us/000020325)
- [OneFS API Reference](https://www.dell.com/support/home/en-us/product-support/product/isilon-onefs)
- [PowerScale SDK](https://github.com/dell/ansible-isilon)

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-01-16 | Initial API reference |
