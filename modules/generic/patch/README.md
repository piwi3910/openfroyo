# Patch Module

Applies patch files to source code, similar to Ansible's `patch` module.

## Description

This module applies unified diff patches to files and directories. It supports both local patch files (uploaded from the control machine) and remote patch files (already on the target host).

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| src | string | yes | - | Path to the patch file (local or remote depending on remote_src) |
| dest | string | no | "." | Directory to apply the patch to |
| strip | integer | no | 0 | Number of leading path components to strip (-p flag) |
| remote_src | boolean | no | false | If true, src is on remote host; if false, src is a local file |

## Examples

### Apply a local patch file
```yaml
- name: Apply security patch
  module: generic/patch
  vars:
    src: ./patches/security-fix.patch
    dest: /opt/application
    strip: 1
```

### Apply a patch already on remote host
```yaml
- name: Apply remote patch
  module: generic/patch
  vars:
    src: /tmp/hotfix.patch
    dest: /opt/application
    remote_src: true
```

### Apply patch to current directory
```yaml
- name: Patch current directory
  module: generic/patch
  vars:
    src: ./myfix.patch
```

### Apply with different strip levels
```yaml
- name: Apply patch with strip level 2
  module: generic/patch
  vars:
    src: ./patches/feature.patch
    dest: /usr/src/myproject
    strip: 2
```

## Notes

- When `remote_src` is false (default), the executor reads the local patch file and base64-encodes it for transfer
- The `strip` parameter corresponds to the `-p` flag in the `patch` command
- Strip level 0: use full path from patch file
- Strip level 1: remove first path component (common for patches from `git diff`)
- The module creates a temporary patch file on the remote host when using local patches
- This module requires the `patch` command to be installed on the target host

## Implementation

This module uses the shell_exec pattern with special handling for local patch files:
- When `remote_src=false`, the executor reads the local patch file and adds it as `_patch_content` (base64-encoded)
- The WASM module writes this content to a temporary file, applies the patch, and cleans up
