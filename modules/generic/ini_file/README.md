# INI File Module

Manages INI file sections and options, similar to Ansible's `ini_file` module.

## Description

This module allows you to manage INI configuration files by adding, modifying, or removing sections and options. It tries to use `crudini` if available, otherwise falls back to `sed`/`awk` commands.

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| path | string | yes | - | Path to the INI file to manage |
| section | string | yes | - | Section name in the INI file |
| option | string | no | - | Option name within the section |
| value | string | no | - | Value for the option |
| state | string | no | "present" | Desired state: "present" or "absent" |
| no_extra_spaces | boolean | no | false | Remove spaces around the = delimiter |
| create | boolean | no | true | Create the file if it doesn't exist |

## Examples

### Create a section
```yaml
- name: Ensure database section exists
  module: generic/ini_file
  vars:
    path: /etc/myapp/config.ini
    section: database
```

### Set an option value
```yaml
- name: Set database host
  module: generic/ini_file
  vars:
    path: /etc/myapp/config.ini
    section: database
    option: host
    value: localhost
```

### Set option without extra spaces
```yaml
- name: Set debug flag
  module: generic/ini_file
  vars:
    path: /etc/myapp/config.ini
    section: application
    option: debug
    value: "true"
    no_extra_spaces: true
```

### Remove an option
```yaml
- name: Remove deprecated option
  module: generic/ini_file
  vars:
    path: /etc/myapp/config.ini
    section: database
    option: old_setting
    state: absent
```

### Remove an entire section
```yaml
- name: Remove deprecated section
  module: generic/ini_file
  vars:
    path: /etc/myapp/config.ini
    section: legacy
    state: absent
```

## Behavior

### State: present
- If only `section` is specified: Ensures the section exists
- If `section` and `option` are specified: Sets the option to the given value
- Creates the file if it doesn't exist (when `create: true`)

### State: absent
- If only `section` is specified: Removes the entire section
- If `section` and `option` are specified: Removes only that option

### Spacing
- By default, options are formatted as `option = value` (with spaces)
- Set `no_extra_spaces: true` for `option=value` (no spaces)

## Notes

- The module first tries to use `crudini` (if installed) for INI manipulation
- Falls back to `sed`/`awk` commands if `crudini` is not available
- When creating new sections/options, they are appended to the file
- Quote numeric values to ensure they are treated as strings
- This module is idempotent - running it multiple times produces the same result

## Implementation

This module uses the shell_exec pattern, generating either `crudini` commands or fallback `sed`/`awk` scripts depending on what's available on the target host.
