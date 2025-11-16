# Tempfile Module

Create temporary files or directories using the `mktemp` command.

## Purpose

This module creates temporary files or directories with unique names in a specified location. It's useful for:
- Creating temporary working directories for scripts
- Generating unique temporary files for data processing
- Ensuring safe temporary file creation without naming conflicts

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| state | string | No | file | Type of temporary item: `file` or `directory` |
| path | string | No | /tmp | Base path for temporary file/directory |
| prefix | string | No | ansible. | Prefix for temporary filename |
| suffix | string | No | "" | Suffix for temporary filename |

## Examples

### Create a temporary file with default settings

```yaml
- name: Create temporary file
  module: generic/tempfile
  vars:
    state: file
```

This creates a file like `/tmp/ansible.XXXXXX` where XXXXXX is a random string.

### Create a temporary file with custom prefix and suffix

```yaml
- name: Create temporary log file
  module: generic/tempfile
  vars:
    state: file
    prefix: "app-log-"
    suffix: ".log"
```

This creates a file like `/tmp/app-log-XXXXXX.log`.

### Create a temporary directory

```yaml
- name: Create temporary working directory
  module: generic/tempfile
  vars:
    state: directory
    prefix: "build-"
```

This creates a directory like `/tmp/build-XXXXXX/`.

### Create a temporary directory in custom path

```yaml
- name: Create temporary directory in /var/tmp
  module: generic/tempfile
  vars:
    state: directory
    path: /var/tmp
    prefix: "work-"
```

This creates a directory like `/var/tmp/work-XXXXXX/`.

## Implementation Details

This module uses the `mktemp` command to create temporary files and directories:
- For files: `mktemp -p {path} {prefix}XXXXXX{suffix}`
- For directories: `mktemp -d -p {path} {prefix}XXXXXX{suffix}`

The `mktemp` command ensures unique names by replacing XXXXXX with random characters.

## Return Value

The module returns the path to the created temporary file or directory in the shell command output.
