# Pip Module

Manage Python packages using pip package manager. Supports installing, removing, and upgrading Python packages from PyPI or requirements files.

## Variables

- `name` (required if no requirements): Package name to install/remove/upgrade
- `state` (optional): Desired state - `present` (default), `absent`, or `latest`
- `version` (optional): Specific version to install (e.g., "2.3.0")
- `requirements` (optional): Path to requirements.txt file
- `virtualenv` (optional): Path to virtualenv directory
- `extra_args` (optional): Additional arguments to pass to pip
- `executable` (optional): pip executable to use - `pip3` (default), `pip`, or `pip2`
- `user` (optional): Install to user site-packages (boolean, default: false)

## States

- **present**: Install the package if not already installed
- **absent**: Remove the package if installed
- **latest**: Upgrade to the latest version (installs if not present)

## Examples

### Install a package

```yaml
- name: Install requests
  module: generic/pip
  hosts:
    - python-server
  vars:
    name: requests
    state: present
```

### Install specific version

```yaml
- name: Install Flask 2.3.0
  module: generic/pip
  hosts:
    - web-server
  vars:
    name: flask
    state: present
    version: "2.3.0"
```

### Install to user site-packages

```yaml
- name: Install black for user
  module: generic/pip
  hosts:
    - dev-machine
  vars:
    name: black
    state: present
    user: true
```

### Upgrade to latest version

```yaml
- name: Upgrade numpy to latest
  module: generic/pip
  hosts:
    - ml-server
  vars:
    name: numpy
    state: latest
```

### Install from requirements.txt

```yaml
- name: Install dependencies
  module: generic/pip
  hosts:
    - app-server
  vars:
    requirements: /opt/myapp/requirements.txt
    state: present
```

### Install in virtualenv

```yaml
- name: Install Django in virtualenv
  module: generic/pip
  hosts:
    - web-server
  vars:
    name: django
    state: present
    virtualenv: /opt/myapp/venv
```

### Remove a package

```yaml
- name: Remove pytest
  module: generic/pip
  hosts:
    - dev-machine
  vars:
    name: pytest
    state: absent
```

### Use with extra arguments

```yaml
- name: Install with extra index URL
  module: generic/pip
  hosts:
    - app-server
  vars:
    name: private-package
    state: present
    extra_args: "--extra-index-url https://pypi.company.com/simple"
```

## How It Works

1. **Package Check**: Uses `pip show <package>` to check if package is installed
2. **Command Generation**: Generates appropriate pip command based on state and options
3. **Execution**: Runs the pip command with specified parameters
4. **Idempotency**: Only makes changes if needed (checks current state first)

## Supported Operations

- Install package: `pip install <package>`
- Install specific version: `pip install <package>==<version>`
- Install to user: `pip install --user <package>`
- Upgrade package: `pip install --upgrade <package>`
- Remove package: `pip uninstall -y <package>`
- Install from requirements: `pip install -r requirements.txt`
- Install in virtualenv: `<venv>/bin/pip install <package>`

## Output

Returns facts:
- `package_name`: Package that was managed
- `pip_executable`: Pip executable used
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
cd modules/generic/pip/wasm
make build
```

This produces `pip.wasm` in the wasm directory.

## Notes

- Default executable is `pip3` - use `executable: pip` for Python 2
- When using virtualenv, the module automatically uses `<venv>/bin/pip`
- The `user` flag installs to user site-packages (no sudo required)
- Requirements file takes precedence over individual package name
- All operations are idempotent - they check current state before making changes
