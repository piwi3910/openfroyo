# Storage Management Modules

This document provides an overview of the OpenFroyo storage management modules.

## Modules

### 1. filesystems
**Location**: `modules/generic/filesystems/`
**Purpose**: Manage filesystem mounts and `/etc/fstab` entries

**Key Features**:
- Mount/unmount filesystems
- Manage `/etc/fstab` entries
- Create/remove mount points
- Support for various filesystems (ext4, xfs, btrfs, etc.)
- Mount by device path, UUID, or LABEL

**States**: `mounted`, `unmounted`, `present`, `absent`

**Example**:
```yaml
- name: Mount data partition
  module: generic/filesystems
  vars:
    device: "/dev/sdb1"
    path: "/mnt/data"
    fstype: "ext4"
    opts: "defaults,noatime"
    state: "mounted"
```

### 2. lvm
**Location**: `modules/generic/lvm/`
**Purpose**: Manage LVM (Physical Volumes, Volume Groups, Logical Volumes)

**Key Features**:
- Create/remove physical volumes
- Create/remove/extend volume groups
- Create/extend/remove logical volumes
- Flexible size specifications (fixed, percentage, free space)
- Automatic PV creation when creating VGs

**Operation Types**:
- Physical Volume: specify `pvs` only
- Volume Group: specify `vg` + `pvs`
- Logical Volume: specify `vg` + `lv`

**Example**:
```yaml
- name: Create 10GB logical volume
  module: generic/lvm
  vars:
    vg: "vg_data"
    lv: "lv_mysql"
    size: "10G"
    state: "present"
```

## Module Structure

Both modules follow the standard OpenFroyo module structure:

```
module-name/
├── README.md              # Complete documentation
├── module.ofy.yml         # Module definition
├── defaults.ofy.yml       # Default variable values
├── test.ofy               # Test/example stack
└── wasm/
    ├── main.go            # Go source code
    └── module.wasm        # Compiled WASM binary
```

## Build Information

Both modules were built using TinyGo:

```bash
# Build filesystems module
cd modules/generic/filesystems
tinygo build -o wasm/filesystems.wasm -target=wasi -no-debug wasm/main.go

# Build lvm module
cd modules/generic/lvm
tinygo build -o wasm/lvm.wasm -target=wasi -no-debug wasm/main.go
```

**Build results**:
- `filesystems.wasm`: 473 KB
- `lvm.wasm`: 477 KB

## Implementation Pattern

Both modules follow the shell_exec pattern:

1. **Input Validation**: Validate required variables and state values
2. **Command Generation**: Generate appropriate shell commands based on state
3. **Shell Execution**: Return `shell_exec` facts for the executor to run
4. **Idempotency**: Check existence before performing operations

### Shell Execution Facts

The modules return facts in this format:

```json
{
  "status": "ok",
  "message": "",
  "facts": {
    "shell_exec": [
      {
        "type": "shell",
        "command": "mount -t ext4 -o defaults /dev/sdb1 /mnt/data"
      }
    ]
  }
}
```

## Testing

Both modules include test files (`test.ofy`) demonstrating various use cases:

- **filesystems/test.ofy**: Mount, unmount, present, absent states
- **lvm/test.ofy**: PV, VG, LV creation/removal, size specifications

Run tests with:
```bash
froyo apply modules/generic/filesystems/test.ofy
froyo apply modules/generic/lvm/test.ofy
```

## Usage Together

These modules work well together for complete storage setup:

```yaml
name: Setup Storage Stack
version: 1.0.0

run:
  # 1. Create LVM structure
  - name: Create volume group
    module: generic/lvm
    vars:
      vg: "vg_data"
      pvs: ["/dev/sdb", "/dev/sdc"]
      state: "present"

  - name: Create logical volume
    module: generic/lvm
    vars:
      vg: "vg_data"
      lv: "lv_mysql"
      size: "20G"
      state: "present"

  # 2. Format and mount
  # (Assumes filesystem was created with mkfs)
  - name: Mount logical volume
    module: generic/filesystems
    vars:
      device: "/dev/vg_data/lv_mysql"
      path: "/var/lib/mysql"
      fstype: "ext4"
      opts: "defaults,noatime"
      state: "mounted"
```

## Requirements

- Root/sudo privileges on target hosts
- LVM tools installed (for lvm module): `lvm2` package
- Appropriate filesystems tools (for filesystems module): `e2fsprogs`, `xfsprogs`, etc.

## Notes

- Both modules are idempotent and safe to run multiple times
- The `filesystems` module does not format filesystems - they must exist
- Extending logical volumes requires separate filesystem resize operations
- Use `force: true` with caution as it can cause data loss
