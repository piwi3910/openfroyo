# Filesystems Module

Manage filesystem mounts on remote hosts.

## Description

The `filesystems` module allows you to:
- Mount and unmount filesystems
- Add/remove entries from `/etc/fstab`
- Create mount points
- Manage filesystem state (mounted, unmounted, present, absent)

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| device | string | yes | - | Device path (e.g., `/dev/sdb1`, `UUID=...`, `LABEL=...`) |
| path | string | yes | - | Mount point path |
| fstype | string | yes | - | Filesystem type (ext4, xfs, btrfs, etc.) |
| opts | string | no | `defaults` | Mount options (comma-separated) |
| state | string | no | `mounted` | Desired state: `mounted`, `unmounted`, `present`, `absent` |
| fstab | bool | no | `true` | Add/remove entry from `/etc/fstab` |
| dump | int | no | `0` | Dump field in fstab (0 or 1) |
| passno | int | no | `0` | Fsck pass number in fstab (0, 1, or 2) |

## States

### mounted
- Creates mount point if it doesn't exist
- Mounts the filesystem
- Adds entry to `/etc/fstab` (if `fstab: true`)
- Idempotent: safe to run multiple times

### unmounted
- Unmounts the filesystem if mounted
- Does not remove fstab entry or mount point

### present
- Creates mount point
- Adds entry to `/etc/fstab` (if `fstab: true`)
- Does not mount the filesystem

### absent
- Unmounts the filesystem if mounted
- Removes entry from `/etc/fstab`
- Removes mount point directory

## Examples

### Basic filesystem mount

```yaml
- name: Mount data partition
  module: generic/filesystems
  vars:
    device: "/dev/sdb1"
    path: "/mnt/data"
    fstype: "ext4"
    state: "mounted"
```

### Mount with custom options

```yaml
- name: Mount with noatime
  module: generic/filesystems
  vars:
    device: "/dev/sdc1"
    path: "/mnt/fast"
    fstype: "xfs"
    opts: "defaults,noatime,nodiratime"
    state: "mounted"
```

### Mount by UUID

```yaml
- name: Mount by UUID
  module: generic/filesystems
  vars:
    device: "UUID=12345678-1234-1234-1234-123456789abc"
    path: "/mnt/backup"
    fstype: "ext4"
    state: "mounted"
```

### Add to fstab without mounting

```yaml
- name: Add to fstab only
  module: generic/filesystems
  vars:
    device: "/dev/sdd1"
    path: "/mnt/archive"
    fstype: "btrfs"
    state: "present"
    fstab: true
```

### Unmount filesystem

```yaml
- name: Unmount temporary filesystem
  module: generic/filesystems
  vars:
    device: "/dev/sde1"
    path: "/mnt/temp"
    fstype: "ext4"
    state: "unmounted"
```

### Complete removal

```yaml
- name: Remove filesystem completely
  module: generic/filesystems
  vars:
    device: "/dev/sdf1"
    path: "/mnt/old"
    fstype: "ext4"
    state: "absent"
```

## Shell Commands Used

The module generates the following shell commands based on the state:

### For state: mounted
```bash
mount | grep ' /path/to/mount '  # Check if mounted
mkdir -p /path/to/mount          # Create mount point
mount -t fstype -o opts device /path/to/mount  # Mount
grep '^/path/to/mount ' /etc/fstab || echo '...' >> /etc/fstab  # Add to fstab
```

### For state: unmounted
```bash
umount /path/to/mount
```

### For state: absent
```bash
umount /path/to/mount
sed -i.bak '\|^/path/to/mount |d' /etc/fstab
rmdir /path/to/mount
```

## Notes

- The module is idempotent - running it multiple times produces the same result
- Root/sudo privileges are required for filesystem operations
- The module creates a backup of `/etc/fstab` as `/etc/fstab.bak` when removing entries
- Paths with spaces are properly escaped
- The module does not format filesystems - the filesystem must already exist on the device
