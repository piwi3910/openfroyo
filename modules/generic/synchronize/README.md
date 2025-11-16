# Synchronize Module

Syncs files and directories using rsync, similar to Ansible's `synchronize` module.

## Description

This module is a wrapper around `rsync`, providing efficient file synchronization between source and destination paths. It preserves file attributes, handles deletions, and supports various rsync options.

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| src | string | yes | - | Source path to sync from |
| dest | string | yes | - | Destination path to sync to |
| delete | boolean | no | false | Delete extraneous files from destination directories |
| recursive | boolean | no | true | Recurse into directories |
| archive | boolean | no | true | Archive mode (preserve permissions, times, symlinks, etc.) |
| checksum | boolean | no | false | Skip files based on checksum instead of mod-time and size |

## Examples

### Basic directory sync
```yaml
- name: Sync web content
  module: generic/synchronize
  vars:
    src: /var/www/html/
    dest: /backup/www/
```

### Sync with delete option
```yaml
- name: Mirror directory (delete extra files)
  module: generic/synchronize
  vars:
    src: /source/data/
    dest: /mirror/data/
    delete: true
```

### Sync using checksums
```yaml
- name: Sync based on file checksums
  module: generic/synchronize
  vars:
    src: /data/source/
    dest: /data/destination/
    checksum: true
```

### Simple recursive copy without archive
```yaml
- name: Simple recursive copy
  module: generic/synchronize
  vars:
    src: /tmp/files/
    dest: /backup/files/
    archive: false
    recursive: true
```

## Notes

- When `archive` is true, it implies `-rlptgoD` flags (recursive, links, permissions, times, group, owner, devices)
- Always add trailing slashes to directory paths for consistent rsync behavior
- The `delete` option will remove files in destination that don't exist in source
- The `checksum` option is slower but more accurate for detecting changed files
- This module requires `rsync` to be installed on the target host

## Implementation

This module uses the shell_exec pattern, returning rsync commands for the runner to execute.
