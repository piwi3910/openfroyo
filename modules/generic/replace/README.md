# Replace Module

Replace text in files using regular expressions with the `sed` command.

## Purpose

This module performs in-place text replacement in files using regular expressions. It's useful for:
- Updating configuration file values
- Modifying source code or scripts
- Batch text replacements
- Configuration management

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| path | string | Yes | - | Path to the file to modify |
| regexp | string | Yes | - | Regular expression pattern to find |
| replace | string | Yes | - | Replacement text |
| backup | boolean | No | false | Create backup with .bak extension before modifying |

## Examples

### Simple text replacement

```yaml
- name: Replace Hello with Hi
  module: generic/replace
  vars:
    path: /tmp/test.txt
    regexp: "Hello"
    replace: "Hi"
```

This replaces all occurrences of "Hello" with "Hi" in the file.

### Replace with backup

```yaml
- name: Update port number with backup
  module: generic/replace
  vars:
    path: /etc/app/config.conf
    regexp: "port=8080"
    replace: "port=9090"
    backup: true
```

This replaces the port number and creates a backup file at `/etc/app/config.conf.bak`.

### Replace using regex patterns

```yaml
- name: Replace IP address pattern
  module: generic/replace
  vars:
    path: /etc/hosts
    regexp: "192\\.168\\.1\\..*"
    replace: "192.168.2.100"
```

This uses a regex pattern to match and replace IP addresses.

### Replace environment variable

```yaml
- name: Update database host
  module: generic/replace
  vars:
    path: /app/.env
    regexp: "DB_HOST=.*"
    replace: "DB_HOST=production-db.example.com"
```

This replaces the database host configuration.

## Implementation Details

This module uses the `sed` command for in-place text replacement:
- Without backup: `sed -i '' 's/{regexp}/{replace}/g' {path}`
- With backup: `sed -i.bak 's/{regexp}/{replace}/g' {path}`

The module automatically escapes special characters in the replacement text (/, \, &) to ensure proper sed syntax.

## Important Notes

1. **Regular Expressions**: The regexp parameter supports standard sed regular expressions
2. **Global Replacement**: All occurrences in the file are replaced (uses the `g` flag)
3. **Backup Files**: When backup is true, the original file is saved with a `.bak` extension
4. **Special Characters**: Forward slashes, backslashes, and ampersands in the replacement text are automatically escaped

## Return Value

The module returns shell commands that perform the text replacement and verify the file was modified successfully.
