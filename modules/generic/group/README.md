# Group Module

Manage system groups on Linux systems. Create, modify, and delete group accounts.

## Supported Platforms

| Platform | Tool | Commands |
|----------|------|----------|
| **Linux** | groupadd/groupmod/groupdel | `groupadd`, `groupmod`, `groupdel` |

## Variables

- `name` (required): Group name to create/manage/delete
- `state` (optional): Desired state - `present` (default) or `absent`
- `gid` (optional): Group ID number
- `system` (optional): Create as system group (boolean, default: false)

## States

- **present**: Create the group if it doesn't exist, update if it does
- **absent**: Remove the group

## Examples

### Create a basic group

```yaml
- name: Create developers group
  module: generic/group
  hosts:
    - dev-server
  vars:
    name: developers
    state: present
```

### Create group with custom GID

```yaml
- name: Create application group
  module: generic/group
  hosts:
    - "@group:app-servers"
  vars:
    name: appgroup
    state: present
    gid: "1500"
```

### Create system group

```yaml
- name: Create nginx system group
  module: generic/group
  hosts:
    - web-server
  vars:
    name: nginx
    state: present
    system: true
```

### Remove a group

```yaml
- name: Remove old group
  module: generic/group
  hosts:
    - server-01
  vars:
    name: oldgroup
    state: absent
```

### Create multiple groups

```yaml
- name: Create project groups
  module: generic/group
  hosts:
    - "@group:dev-servers"
  vars:
    name: "{{ item }}"
    state: present
  loop:
    - frontend-team
    - backend-team
    - devops-team
```

### Create group before user

```yaml
# First create the group
- name: Create app group
  module: generic/group
  hosts:
    - app-server
  vars:
    name: appgroup
    state: present
    gid: "1500"

# Then create user with that group
- name: Create app user
  module: generic/user
  hosts:
    - app-server
  vars:
    name: appuser
    state: present
    group: appgroup
```

## How It Works

1. **Check Existence**: Uses `getent group` to check if group exists
2. **Create/Modify**:
   - New groups: Uses `groupadd` with specified options
   - Existing groups: Uses `groupmod` to update properties
3. **Remove**: Uses `groupdel` to remove the group

## Output

Returns facts:
- `groupname`: Group that was managed
- `action`: Action performed (created, updated, deleted, unchanged)
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/group.wasm`.

## Notes

- System groups typically have GIDs below 1000
- You cannot delete a group that is the primary group of any user
- Supplementary group membership is managed through the user module
- GID changes may affect file ownership
