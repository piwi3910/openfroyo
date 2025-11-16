# Line in File Module

Manage lines in text files - add, remove, or replace lines using pattern matching.

## Overview

The `lineinfile` module ensures that a particular line is present or absent in a file. It's useful for managing configuration files by adding, replacing, or removing specific lines.

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `path` | string | Yes | - | Path to the file to manage |
| `line` | string | No | - | Text of the line to ensure exists in the file |
| `regexp` | string | No | - | Regular expression pattern to match lines (for replacement or removal) |
| `state` | string | No | `present` | Whether the line should be `present` or `absent` |
| `insertafter` | string | No | - | Insert line after this regexp pattern (use 'EOF' for end of file) |
| `insertbefore` | string | No | - | Insert line before this regexp pattern (use 'BOF' for beginning of file) |
| `backup` | boolean | No | `false` | Create a backup of the file before modification |
| `create` | boolean | No | `false` | Create the file if it does not exist |

## Behavior

### Adding Lines

- If `state=present` and the line doesn't exist, it will be added
- By default, lines are appended to the end of the file
- Use `insertafter` or `insertbefore` to control placement

### Replacing Lines

- If `regexp` is specified with `state=present`, lines matching the pattern will be replaced
- If no match is found, the line will be inserted according to placement rules

### Removing Lines

- If `state=absent`, all lines matching `regexp` will be removed
- `regexp` parameter is required when `state=absent`

## Examples

### Add a configuration line

```yaml
- name: Set port configuration
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    line: "port=8080"
```

### Replace a configuration value

```yaml
- name: Update port to 9090
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    regexp: "^port="
    line: "port=9090"
```

### Insert after a specific pattern

```yaml
- name: Add database config after port
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    line: "database=postgres"
    insertafter: "^port="
```

### Insert at beginning of file

```yaml
- name: Add header comment
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    line: "# Application Configuration"
    insertbefore: BOF
```

### Remove lines matching pattern

```yaml
- name: Remove debug configuration
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    regexp: "^debug="
    state: absent
```

### Create file and add line

```yaml
- name: Ensure config exists with setting
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    line: "enabled=true"
    create: true
```

### Make backup before changes

```yaml
- name: Update critical config with backup
  module: generic/lineinfile
  vars:
    path: /etc/app/config.conf
    regexp: "^max_connections="
    line: "max_connections=100"
    backup: true
```

## Implementation

This module returns `shell_exec` facts containing sed, grep, and echo commands that:

1. Verify the file exists (or create it if `create=true`)
2. Create a backup if requested
3. Use sed for in-place line replacement/insertion
4. Use grep to verify the final state

The runner executes these shell commands on the target host.

## Testing

Build the module:

```bash
cd modules/generic/lineinfile
make build
```

Run tests:

```bash
froyo apply modules/generic/lineinfile/test.ofy
```

## Notes

- Single quotes in the line text are properly escaped for shell safety
- The module uses `sed -i.tmp` for in-place editing (creates .tmp backup automatically deleted)
- When using `insertafter="EOF"`, the line is appended to the end
- When using `insertbefore="BOF"`, the line is inserted at the beginning
