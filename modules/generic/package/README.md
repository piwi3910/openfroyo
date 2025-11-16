# Package Module

Universal package management across Linux, macOS, and Windows. Automatically detects and uses the appropriate package manager.

## Supported Package Managers

| Platform | Package Manager | Command |
|----------|----------------|---------|
| **Debian/Ubuntu** | apt | `apt-get` |
| **RHEL/CentOS/Fedora** | dnf/yum | `dnf` or `yum` |
| **Arch Linux** | pacman | `pacman` |
| **macOS** | Homebrew | `brew` |
| **Windows** | Chocolatey | `choco` |
| **Windows** | winget | `winget` |

## Variables

- `name` (required): Package name to install/remove/upgrade
- `state` (optional): Desired state - `present` (default), `absent`, or `latest`
- `update_cache` (optional): Update package cache before operation (boolean, default: false)

## States

- **present**: Install the package if not already installed
- **absent**: Remove the package if installed
- **latest**: Update package cache and upgrade to latest version

## Examples

### Install a package

```yaml
- name: Install nginx
  module: generic/package
  hosts:
    - web-server
  vars:
    name: nginx
    state: present
```

### Install with cache update

```yaml
- name: Install git with cache update
  module: generic/package
  hosts:
    - "@group:servers"
  vars:
    name: git
    state: present
    update_cache: true
```

### Remove a package

```yaml
- name: Remove apache
  module: generic/package
  hosts:
    - web-server
  vars:
    name: apache2
    state: absent
```

### Upgrade to latest version

```yaml
- name: Upgrade nginx to latest
  module: generic/package
  hosts:
    - web-server
  vars:
    name: nginx
    state: latest
```

### Cross-platform examples

```yaml
# Install Python (works on all platforms)
- name: Install Python
  module: generic/package
  hosts:
    - "@group:all"
  vars:
    name: python3  # On Windows: python
    state: present

# Install Docker (platform-specific names)
- name: Install Docker on Linux
  module: generic/package
  hosts:
    - "@group:linux-servers"
  vars:
    name: docker.io
    state: present

- name: Install Docker on Mac
  module: generic/package
  hosts:
    - "@group:mac-servers"
  vars:
    name: docker
    state: present
```

## How It Works

1. **Auto-Detection**: Automatically detects which package manager is available
2. **Command Generation**: Generates the appropriate command for the detected package manager
3. **Execution**: Runs the package management command with sudo (when required)
4. **Result**: Returns success/failure status and output

## Output

Returns facts:
- `package_manager`: Detected package manager (apt, yum, brew, etc.)
- `package_name`: Package that was managed
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/package.wasm`.
