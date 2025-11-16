# Package Manager Modules

This document describes the language-specific package manager modules added to OpenFroyo.

## Overview

OpenFroyo now includes dedicated package manager modules for the most popular programming language ecosystems:
- **gem** - Ruby (RubyGems)
- **npm** - Node.js (npm)
- **pip** - Python (pip)
- **yarn** - Node.js (Yarn)

These modules complement the existing **package** module, which handles system-level package managers (apt, dnf, yum, pacman, brew, etc.).

## Module Comparison

| Module | Package Manager | Language | Global Support | Version Pinning | User Install |
|--------|----------------|----------|----------------|-----------------|--------------|
| gem | RubyGems | Ruby | No | ✅ Yes | ✅ Yes (default) |
| npm | npm | Node.js | ✅ Yes | ✅ Yes | ❌ No |
| pip | pip | Python | ❌ No* | ✅ Yes | ✅ Yes |
| yarn | Yarn | Node.js | ✅ Yes | ✅ Yes | ❌ No |

*pip has a --user flag instead of true global install

## Common Features

All package manager modules share these features:

### States
- **present** - Ensure package is installed
- **absent** - Ensure package is removed
- **latest** - Update package to latest version

### Idempotency
All modules check if a package is already installed before attempting installation, ensuring idempotent operations.

### Error Handling
All modules verify the package manager exists before attempting operations and provide clear error messages.

### Shell_exec Pattern
All modules follow the established shell_exec pattern, returning shell commands for the froyo-runner to execute.

## Module Details

### gem Module

**Purpose:** Manage Ruby gems using RubyGems

**Key Variables:**
- `name` (required): Gem name
- `state`: present, absent, latest (default: present)
- `version`: Specific version (e.g., "2.1.0")
- `user_install`: Install in user directory (default: true)
- `source`: Custom gem repository URL
- `install_dir`: Custom installation directory

**Example:**
```yaml
- name: Install specific version of rails
  module: generic/gem
  vars:
    name: rails
    version: "7.0.0"
    user_install: true
  hosts:
    - webserver
```

**Commands:**
- Check: `gem list -i '^{name}$'`
- Install: `gem install {name} [--user-install] [-v {version}]`
- Uninstall: `gem uninstall {name} -x -I`
- Update: `gem update {name}`

---

### npm Module

**Purpose:** Manage Node.js packages using npm

**Key Variables:**
- `name` (required): Package name
- `state`: present, absent, latest (default: present)
- `version`: Specific version (e.g., "2.1.0")
- `global`: Install globally (default: false)
- `path`: Working directory for local installs
- `production`: Install only production dependencies (default: false)

**Example:**
```yaml
- name: Install express in project
  module: generic/npm
  vars:
    name: express
    version: "4.18.0"
    path: /var/www/myapp
  hosts:
    - webserver
```

**Commands:**
- Check global: `npm list -g {name} --depth=0`
- Check local: `npm list {name} --depth=0`
- Install global: `npm install -g {name}[@{version}]`
- Install local: `npm install {name}[@{version}] [--production]`
- Uninstall: `npm uninstall [-g] {name}`
- Update: `npm update [-g] {name}`

---

### pip Module

**Purpose:** Manage Python packages using pip

**Key Variables:**
- `name` (required): Package name
- `state`: present, absent, latest (default: present)
- `version`: Specific version (e.g., "2.1.0")
- `requirements`: Path to requirements.txt file
- `virtualenv`: Path to virtualenv
- `executable`: pip executable to use (default: "pip3")
- `user`: Install to user site-packages (default: false)
- `extra_args`: Additional arguments to pass to pip

**Example:**
```yaml
- name: Install from requirements file
  module: generic/pip
  vars:
    requirements: /var/www/myapp/requirements.txt
    virtualenv: /var/www/myapp/venv
  hosts:
    - webserver
```

**Commands:**
- Check: `{executable} show {name}`
- Install: `{executable} install {name}[=={version}] [--user]`
- Uninstall: `{executable} uninstall -y {name}`
- Update: `{executable} install --upgrade {name}`
- From requirements: `{executable} install -r {requirements}`
- In virtualenv: `{virtualenv}/bin/pip install {name}`

---

### yarn Module

**Purpose:** Manage Node.js packages using Yarn

**Key Variables:**
- `name` (required): Package name
- `state`: present, absent, latest (default: present)
- `version`: Specific version (e.g., "2.1.0")
- `global`: Install globally (default: false)
- `path`: Working directory for local installs
- `production`: Install only production dependencies (default: false)

**Example:**
```yaml
- name: Install dependencies from package.json
  module: generic/yarn
  vars:
    path: /var/www/myapp
    production: true
  hosts:
    - webserver
```

**Commands:**
- Check global: `yarn global list --pattern {name}`
- Check local: `yarn list --pattern {name}`
- Install global: `yarn global add {name}[@{version}]`
- Install local: `yarn add {name}[@{version}]`
- Uninstall: `yarn [global] remove {name}`
- Update: `yarn [global] upgrade {name}`
- Install all: `yarn install [--production]`

## Build Information

All modules successfully compiled to WebAssembly:

```
gem.wasm:  461KB (472,390 bytes)
npm.wasm:  463KB (473,088 bytes)
pip.wasm:  464KB (475,136 bytes)
yarn.wasm: 457KB (468,992 bytes)
```

Total size: ~1.82MB for all 4 modules

## Usage Patterns

### Development Environment Setup

```yaml
- name: Install Ruby gems
  module: generic/gem
  vars:
    name: "{{ item }}"
    user_install: true
  loop:
    - bundler
    - rake
    - rspec

- name: Install Node.js packages
  module: generic/npm
  vars:
    path: /var/www/myapp
  hosts:
    - webserver

- name: Install Python packages
  module: generic/pip
  vars:
    requirements: requirements.txt
    virtualenv: /var/www/myapp/venv
  hosts:
    - webserver
```

### Production Deployment

```yaml
- name: Install production Node dependencies
  module: generic/yarn
  vars:
    path: /var/www/myapp
    production: true
  hosts:
    - webserver

- name: Install specific Python package versions
  module: generic/pip
  vars:
    name: django
    version: "4.2.0"
    virtualenv: /var/www/myapp/venv
  hosts:
    - webserver
```

### Package Updates

```yaml
- name: Update all gems
  module: generic/gem
  vars:
    name: "{{ item }}"
    state: latest
  loop:
    - rails
    - puma

- name: Update npm packages
  module: generic/npm
  vars:
    name: "{{ item }}"
    state: latest
    global: true
  loop:
    - npm
    - typescript
```

## Implementation Notes

### Script Generation
All modules generate bash scripts that:
1. Verify the package manager exists
2. Check current installation status
3. Execute appropriate commands based on state
4. Provide idempotent operations

### Error Handling
Modules fail gracefully with clear messages when:
- Package manager is not installed
- Required directories don't exist (for path-based operations)
- Invalid state is specified
- Package installation fails

### Platform Support
- **gem**: Works on any system with Ruby and RubyGems installed
- **npm**: Works on any system with Node.js and npm installed
- **pip**: Works on any system with Python and pip installed
- **yarn**: Works on any system with Node.js and Yarn installed

## Testing

Test stacks are provided in each module's `test.ofy` file:
- `modules/generic/gem/test.ofy`
- `modules/generic/npm/test.ofy`
- `modules/generic/pip/test.ofy`
- `modules/generic/yarn/test.ofy`

Comprehensive test stack: `stacks/test_package_managers.ofy`

**Note:** Tests require the respective package managers to be installed on target hosts.

## GitHub Issue

This work was tracked in: [Issue #13 - Add language-specific package manager modules](https://github.com/piwi3910/openfroyo/issues/13)

## Related Modules

- [package](../modules/generic/package/README.md) - System package manager (apt, dnf, yum, etc.)
- [command](../modules/generic/command/README.md) - Execute shell commands
- [script](../modules/generic/script/README.md) - Execute scripts

## Future Enhancements

Potential additions:
- **composer** - PHP package manager
- **cargo** - Rust package manager
- **go get** - Go package manager
- **maven/gradle** - Java build/dependency tools
- **nuget** - .NET package manager
