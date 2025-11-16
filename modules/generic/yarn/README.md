# Yarn Module

Manage Node.js packages using the Yarn package manager.

## Description

The yarn module allows you to install, update, and remove Node.js packages using Yarn. It supports both global and local (project-specific) installations, version specifications, and production dependency filtering.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| name | string | yes | - | Package name to manage |
| state | string | no | present | Desired state: `present`, `absent`, `latest` |
| version | string | no | - | Specific version to install (e.g., "2.1.0") |
| global | boolean | no | false | Install package globally |
| path | string | no | - | Working directory for local installs |
| production | boolean | no | false | Install only production dependencies |

## States

- **present**: Ensure the package is installed
- **absent**: Ensure the package is not installed
- **latest**: Update the package to the latest version

## Examples

### Install a global package

```yaml
- name: Install lodash globally
  module: generic/yarn
  vars:
    name: lodash
    global: true
  hosts:
    - webserver
```

### Install a specific version

```yaml
- name: Install specific version of prettier
  module: generic/yarn
  vars:
    name: prettier
    version: "2.8.0"
    global: true
  hosts:
    - webserver
```

### Install a local package

```yaml
- name: Install express in project
  module: generic/yarn
  vars:
    name: express
    path: /var/www/myapp
  hosts:
    - webserver
```

### Install all dependencies from package.json

```yaml
- name: Install all dependencies
  module: generic/yarn
  vars:
    path: /var/www/myapp
  hosts:
    - webserver
```

### Install production dependencies only

```yaml
- name: Install production dependencies
  module: generic/yarn
  vars:
    path: /var/www/myapp
    production: true
  hosts:
    - webserver
```

### Update package to latest version

```yaml
- name: Update lodash to latest
  module: generic/yarn
  vars:
    name: lodash
    state: latest
    global: true
  hosts:
    - webserver
```

### Remove a package

```yaml
- name: Remove prettier
  module: generic/yarn
  vars:
    name: prettier
    state: absent
    global: true
  hosts:
    - webserver
```

## Implementation Details

This module uses the shell_exec pattern to generate bash scripts that:

1. Verify yarn is installed
2. Check if the package is already installed (globally or locally)
3. Execute the appropriate yarn command based on the desired state
4. Handle both global and local scopes correctly

### Commands Used

- **Check global**: `yarn global list --pattern "{name}"`
- **Check local**: `yarn list --pattern "{name}"` (in specified path)
- **Install global**: `yarn global add {name}[@version]`
- **Install local**: `yarn add {name}[@version]`
- **Remove global**: `yarn global remove {name}`
- **Remove local**: `yarn remove {name}`
- **Update global**: `yarn global upgrade {name}`
- **Update local**: `yarn upgrade {name}`
- **Install all**: `yarn install [--production]`

## Idempotency

The module is idempotent - it checks whether a package is already installed before attempting to install it, and only makes changes when necessary.

## Return Values

The module returns shell command output via the `shell_exec` fact, which the froyo-runner executes. The output includes:
- Installation/removal status messages
- Change indicators (CHANGED=true/false)
- Success indicators (SUCCESS=true)

## Requirements

- Yarn must be installed on the remote host
- For local installations, the path must exist and contain a package.json file (for most operations)
- Appropriate permissions to install packages (globally or locally)

## Notes

- Global installations require appropriate permissions (may need sudo)
- Local installations are scoped to the specified path's node_modules directory
- The module uses `set -e` to fail fast on errors
- Production mode installs only dependencies (not devDependencies)

## See Also

- [npm module](../npm/README.md) - Alternative Node.js package manager
- [package module](../package/README.md) - Generic system package manager
