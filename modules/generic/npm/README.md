# NPM Module

Manage Node.js packages using the npm package manager. Install, update, and remove packages globally or locally with support for versioning, production dependencies, and custom paths.

## Requirements

- Node.js and npm must be installed on the target system
- For global installations, appropriate permissions may be required
- For local installations, a valid directory (with or without package.json) is required

## Variables

- `name` (required): Package name to install/remove/upgrade
- `state` (optional): Desired state - `present` (default), `absent`, or `latest`
- `version` (optional): Specific version to install (e.g., "4.17.21")
- `global` (optional): Install globally (boolean, default: false)
- `path` (optional): Path to package.json directory for local installs
- `production` (optional): Install only production dependencies (boolean, default: false)

## States

- **present**: Install the package if not already installed
- **absent**: Remove the package if installed
- **latest**: Update package to the latest available version

## Examples

### Install a package globally

```yaml
- name: Install Express globally
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: express
    state: present
    global: true
```

### Install a specific version

```yaml
- name: Install lodash 4.17.21
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: lodash
    state: present
    version: "4.17.21"
```

### Install locally in a specific directory

```yaml
- name: Install React in project directory
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: react
    state: present
    path: "/var/www/myapp"
```

### Install production dependencies only

```yaml
- name: Install axios (production only)
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: axios
    state: present
    production: true
```

### Update to latest version

```yaml
- name: Update Express to latest
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: express
    state: latest
    global: true
```

### Remove a package globally

```yaml
- name: Remove old package globally
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: deprecated-package
    state: absent
    global: true
```

### Remove a package locally

```yaml
- name: Remove package from project
  module: generic/npm
  hosts:
    - node-server
  vars:
    name: unused-package
    state: absent
    path: "/var/www/myapp"
```

## How It Works

1. **NPM Detection**: Checks if the `npm` command is available
2. **Status Check**: Verifies if the package is installed (globally or locally)
3. **Command Generation**: Builds the appropriate `npm install`, `npm update`, or `npm uninstall` command
4. **Execution**: Runs the npm command with specified options
5. **Result**: Returns success/failure status and output

## Behavior Details

### Global vs Local Installation

#### Global Installation (`global: true`)
- Packages are installed in the global npm directory
- Available system-wide via command line
- Typically requires elevated permissions on Linux/macOS
- Use for CLI tools and utilities

#### Local Installation (`global: false`, default)
- Packages are installed in the `node_modules` directory
- Scoped to the project directory
- Recommended for project dependencies
- No special permissions required

### Path Handling
- If `path` is specified, the module changes to that directory before executing npm commands
- The directory must exist before attempting installation
- If `path` is not specified, operations occur in the current working directory

### Version Handling
- If `version` is specified with `state: present`, only that specific version will be installed
- Format: `name@version` (e.g., `lodash@4.17.21`)
- If the package is already installed, npm will verify or update to the specified version
- If `version` is not specified, the latest version is installed

### Production Dependencies
- When `production: true`, only dependencies (not devDependencies) are installed
- Useful for production deployments where development tools are not needed
- Only applies to local installations

### Checking Installation Status
- **Global**: Uses `npm list -g <package> --depth=0`
- **Local**: Uses `npm list <package> --depth=0` in the specified directory
- Returns exit code 0 if installed, non-zero otherwise

## Common Use Cases

### Development Environment Setup

```yaml
- name: Install global development tools
  module: generic/npm
  hosts:
    - dev-servers
  vars:
    name: "{{ item }}"
    state: present
    global: true
  loop:
    - nodemon
    - typescript
    - eslint
```

### Production Deployment

```yaml
- name: Install production dependencies
  module: generic/npm
  hosts:
    - prod-servers
  vars:
    name: "all"
    state: present
    path: "/var/www/app"
    production: true
```

### Package Version Pinning

```yaml
- name: Install specific React version
  module: generic/npm
  hosts:
    - web-servers
  vars:
    name: react
    state: present
    version: "18.2.0"
    path: "/var/www/myapp"
```

## Output

Returns facts:
- `package_name`: Name of the package managed
- `package_state`: State that was applied
- `package_scope`: "global" or "local"
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/npm.wasm`.
