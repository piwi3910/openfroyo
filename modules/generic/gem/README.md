# Gem Module

Manage Ruby gems using the RubyGems package manager. Install, update, and remove gems with support for versioning, user installations, and custom sources.

## Requirements

- Ruby and RubyGems must be installed on the target system
- Appropriate permissions for gem installation (use `user_install: true` for non-root users)

## Variables

- `name` (required): Gem name to install/remove/upgrade
- `state` (optional): Desired state - `present` (default), `absent`, or `latest`
- `version` (optional): Specific version to install (e.g., "2.1.0")
- `user_install` (optional): Install in user's home directory (boolean, default: true)
- `source` (optional): Gem repository URL
- `install_dir` (optional): Custom installation directory

## States

- **present**: Install the gem if not already installed
- **absent**: Remove the gem if installed (removes all versions)
- **latest**: Update gem to the latest available version

## Examples

### Install a gem

```yaml
- name: Install bundler
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: bundler
    state: present
```

### Install a specific version

```yaml
- name: Install Rails 6.1.0
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: rails
    state: present
    version: "6.1.0"
```

### Install with system-wide installation

```yaml
- name: Install gem system-wide
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: json
    state: present
    user_install: false
```

### Install from a custom source

```yaml
- name: Install from custom gem server
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: my-private-gem
    state: present
    source: "https://gems.company.com"
```

### Install to a specific directory

```yaml
- name: Install to custom directory
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: rake
    state: present
    install_dir: "/opt/custom-gems"
```

### Update to latest version

```yaml
- name: Update bundler to latest
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: bundler
    state: latest
```

### Remove a gem

```yaml
- name: Remove old gem
  module: generic/gem
  hosts:
    - ruby-server
  vars:
    name: deprecated-gem
    state: absent
```

## How It Works

1. **Gem Detection**: Checks if the `gem` command is available
2. **Status Check**: Verifies if the gem (and specific version) is installed
3. **Command Generation**: Builds the appropriate `gem install`, `gem update`, or `gem uninstall` command
4. **Execution**: Runs the gem command with specified options
5. **Result**: Returns success/failure status and output

## Behavior Details

### User Install (default: true)
- When `user_install: true`, gems are installed in the user's home directory (~/.gem)
- No sudo/root access required
- Recommended for development environments

### System Install (user_install: false)
- Installs gems in the system Ruby directory
- May require sudo/root access
- Recommended for production servers

### Version Handling
- If `version` is specified with `state: present`, only that specific version will be installed
- If the version is already installed, no action is taken
- If `version` is not specified, the latest version is installed

### Uninstall Behavior
- Removes all versions of the gem
- Removes executables (`-x` flag)
- Ignores dependencies (`-I` flag)

## Output

Returns facts:
- `gem_name`: Name of the gem managed
- `gem_state`: State that was applied
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/gem.wasm`.
