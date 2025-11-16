# LVM Module

Manage LVM (Logical Volume Manager) on remote hosts.

## Description

The `lvm` module allows you to:
- Create and remove physical volumes (PVs)
- Create and remove volume groups (VGs)
- Create, extend, and remove logical volumes (LVs)
- Manage LVM resources with flexible size specifications

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| vg | string | conditional | - | Volume group name |
| lv | string | conditional | - | Logical volume name |
| pvs | array | conditional | `[]` | List of physical volume device paths |
| size | string | conditional | - | Size specification (e.g., `10G`, `50%VG`, `100%FREE`) |
| state | string | no | `present` | Desired state: `present` or `absent` |
| force | bool | no | `false` | Force creation/deletion operations |
| opts | string | no | - | Additional options for LVM commands |

## Operation Types

The module determines the operation type based on which variables are provided:

1. **Logical Volume**: `vg` + `lv` specified
2. **Volume Group**: `vg` + `pvs` specified
3. **Physical Volume**: only `pvs` specified

## Size Specifications

For logical volumes, the `size` parameter supports:

- **Fixed size**: `10G`, `500M`, `1T` (size in gigabytes, megabytes, terabytes)
- **Percentage of VG**: `50%VG`, `75%VG` (percentage of volume group)
- **Percentage of free space**: `100%FREE`, `50%FREE` (percentage of available space)

## Examples

### Physical Volume Operations

#### Create physical volumes
```yaml
- name: Create physical volumes
  module: generic/lvm
  vars:
    pvs:
      - "/dev/sdb"
      - "/dev/sdc"
    state: "present"
```

#### Remove physical volumes
```yaml
- name: Remove physical volumes
  module: generic/lvm
  vars:
    pvs:
      - "/dev/sdd"
    state: "absent"
    force: true
```

### Volume Group Operations

#### Create volume group
```yaml
- name: Create volume group
  module: generic/lvm
  vars:
    vg: "vg_data"
    pvs:
      - "/dev/sdb"
      - "/dev/sdc"
    state: "present"
```

#### Extend volume group
```yaml
- name: Add PV to existing VG
  module: generic/lvm
  vars:
    vg: "vg_data"
    pvs:
      - "/dev/sdd"
    state: "present"
```

#### Remove volume group
```yaml
- name: Remove volume group
  module: generic/lvm
  vars:
    vg: "vg_old"
    pvs: []
    state: "absent"
    force: true
```

### Logical Volume Operations

#### Create logical volume with fixed size
```yaml
- name: Create 10GB logical volume
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_mysql"
    size: "10G"
    state: "present"
```

#### Create LV using percentage of VG
```yaml
- name: Create LV using 50% of VG
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_postgres"
    size: "50%VG"
    state: "present"
```

#### Create LV using all free space
```yaml
- name: Use all available space
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_backup"
    size: "100%FREE"
    state: "present"
```

#### Extend existing logical volume
```yaml
- name: Extend LV to 20GB
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_mysql"
    size: "20G"
    state: "present"
```

#### Remove logical volume
```yaml
- name: Remove logical volume
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_old"
    state: "absent"
    force: true
```

## Complete Example: Set up LVM storage

```yaml
name: Setup LVM Storage
version: 1.0.0

inventory:
  - hosts:
      - storage-server

defaults:
  ssh_user: root

run:
  # Step 1: Create physical volumes
  - name: Initialize disks as PVs
    module: generic/lvm
    vars:
      pvs:
        - "/dev/sdb"
        - "/dev/sdc"
        - "/dev/sdd"
      state: "present"

  # Step 2: Create volume group
  - name: Create data volume group
    module: generic/lvm
    vars:
      vg: "vg_data"
      pvs:
        - "/dev/sdb"
        - "/dev/sdc"
        - "/dev/sdd"
      state: "present"

  # Step 3: Create logical volumes
  - name: Create MySQL LV (20GB)
    module: generic/lvm
    vars:
      vg: "vg_data"
      lv: "lv_mysql"
      size: "20G"
      state: "present"

  - name: Create PostgreSQL LV (30GB)
    module: generic/lvm
    vars:
      vg: "vg_data"
      lv: "lv_postgres"
      size: "30G"
      state: "present"

  - name: Create backup LV (rest of space)
    module: generic/lvm
    vars:
      vg: "vg_data"
      lv: "lv_backup"
      size: "100%FREE"
      state: "present"
```

## Shell Commands Used

### Physical Volume commands
```bash
pvdisplay /dev/sdb           # Check PV
pvcreate /dev/sdb            # Create PV
pvcreate -f /dev/sdb         # Force create PV
pvremove /dev/sdb            # Remove PV
```

### Volume Group commands
```bash
vgdisplay vg_name            # Check VG
vgcreate vg_name /dev/sdb... # Create VG
vgextend vg_name /dev/sdc    # Add PV to VG
vgremove vg_name             # Remove VG
vgremove -f vg_name          # Force remove VG
```

### Logical Volume commands
```bash
lvdisplay vg/lv              # Check LV
lvcreate -n lv -L 10G vg     # Create LV (fixed size)
lvcreate -n lv -l 50%VG vg   # Create LV (percentage)
lvextend -L 20G vg/lv        # Extend LV
lvremove vg/lv               # Remove LV
lvremove -f vg/lv            # Force remove LV
```

## Notes

- The module is idempotent - running it multiple times produces the same result
- Root/sudo privileges are required for LVM operations
- The `force` option should be used with caution as it can cause data loss
- Extending a logical volume does not automatically resize the filesystem - use the `filesystems` module or separate resize commands
- When removing volume groups, all logical volumes must be removed first
- When removing physical volumes, they must not be part of any volume group
- The module checks for existence before performing operations to ensure idempotency
