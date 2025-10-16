# linux.pkg - OpenFroyo Linux Package Management Provider

A WASM-based provider for managing Linux packages across multiple package managers including apt, dnf, yum, and zypper.

## Overview

The `linux.pkg` provider enables declarative management of system packages on Linux distributions. It supports the most common package managers and automatically detects the appropriate one for your system.

### Supported Package Managers

- **apt** - Debian, Ubuntu, and derivatives
- **dnf** - Fedora, RHEL 8+, CentOS Stream
- **yum** - RHEL 7, CentOS 7
- **zypper** - openSUSE, SLES (SUSE Linux Enterprise Server)

## Architecture

This provider is compiled to WebAssembly (WASM) for secure, sandboxed execution. It delegates actual package operations to the OpenFroyo micro-runner using the `exec:micro-runner` capability, ensuring:

- **Security**: WASM sandbox prevents unauthorized system access
- **Portability**: Same WASM module works across all platforms
- **Efficiency**: Lightweight execution with minimal overhead
- **Isolation**: Provider cannot directly modify system state

### Provider Flow

```
┌─────────────┐
│  OpenFroyo  │
│   Engine    │
└──────┬──────┘
       │
       ▼
┌─────────────┐     WASM         ┌──────────────┐
│   WASM      │────────────────▶│  linux.pkg   │
│   Runtime   │                  │  Provider    │
└─────────────┘                  └──────┬───────┘
                                        │
                                        │ micro-runner
                                        │ protocol
                                        ▼
                                 ┌──────────────┐
                                 │ Micro-Runner │
                                 │  (on host)   │
                                 └──────┬───────┘
                                        │
                                        │ pkg.ensure
                                        ▼
                                 ┌──────────────┐
                                 │   apt/dnf/   │
                                 │  yum/zypper  │
                                 └──────────────┘
```

## Installation

### Build from Source

```bash
# Requires TinyGo (https://tinygo.org/getting-started/install/)
cd providers/linux.pkg
make build
```

### Package and Install

```bash
# Create distributable package
make package

# Install in OpenFroyo
froyo provider install linux.pkg-1.0.0.tar.gz
```

## Resource Configuration

### Basic Schema

```cue
resource "nginx_package" {
  type: "linux.pkg::package"

  config: {
    package: "nginx"
    state:   "present"  // present, absent, or latest
  }
}
```

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `package` | string | Yes | Name of the package to manage |
| `state` | string | No | Desired state: `present`, `absent`, or `latest` (default: `present`) |
| `version` | string | No | Specific version to install (incompatible with `latest` or `absent`) |
| `repository` | string | No | Specific repository to use |
| `manager` | string | No | Package manager to use (auto-detected if not specified) |
| `options` | array | No | Additional package manager options |

### State Values

- **present**: Package should be installed (default behavior)
- **absent**: Package should be removed
- **latest**: Package should be upgraded to the latest available version

## Examples

### Example 1: Install a Package

```cue
resource "install_nginx" {
  type: "linux.pkg::package"

  config: {
    package: "nginx"
    state:   "present"
  }
}
```

### Example 2: Install Specific Version

```cue
resource "install_postgresql" {
  type: "linux.pkg::package"

  config: {
    package: "postgresql-14"
    state:   "present"
    version: "14.5-1ubuntu1"
    manager: "apt"
  }
}
```

### Example 3: Keep Package Up-to-Date

```cue
resource "update_security_tools" {
  type: "linux.pkg::package"

  config: {
    package: "openssl"
    state:   "latest"
  }
}
```

### Example 4: Remove a Package

```cue
resource "remove_apache" {
  type: "linux.pkg::package"

  config: {
    package: "apache2"
    state:   "absent"
  }
}
```

### Example 5: Install from Specific Repository

```cue
resource "install_docker" {
  type: "linux.pkg::package"

  config: {
    package:    "docker-ce"
    state:      "present"
    repository: "docker-ce-stable"
    manager:    "dnf"
    options:    ["--enablerepo=docker-ce-stable"]
  }
}
```

### Example 6: Install with Custom Options

```cue
resource "install_minimal_nginx" {
  type: "linux.pkg::package"

  config: {
    package: "nginx"
    state:   "present"
    manager: "apt"
    options: ["--no-install-recommends"]
  }
}
```

### Example 7: Multi-Package Configuration

```cue
// Web server stack
resource "nginx_package" {
  type: "linux.pkg::package"
  config: {
    package: "nginx"
    state:   "latest"
  }
}

resource "php_fpm" {
  type: "linux.pkg::package"
  config: {
    package: "php-fpm"
    state:   "present"
  }
}

resource "postgresql" {
  type: "linux.pkg::package"
  config: {
    package: "postgresql-14"
    state:   "present"
  }
}

// Dependency relationships
resource "php_fpm" {
  depends_on: ["nginx_package"]
}

resource "postgresql" {
  depends_on: ["php_fpm"]
}
```

## Provider Configuration

You can configure provider defaults in your OpenFroyo configuration:

```cue
providers: {
  "linux.pkg": {
    config: {
      default_manager: "apt"
      update_cache: true
      cache_validity_minutes: 60
    }
  }
}
```

### Provider Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_manager` | string | auto-detect | Default package manager to use |
| `update_cache` | boolean | true | Whether to update package cache before operations |
| `cache_validity_minutes` | integer | 60 | How long package cache is considered fresh |

## Resource State

After applying a resource, OpenFroyo tracks the following state:

```json
{
  "package": "nginx",
  "installed": true,
  "version": "1.18.0-1ubuntu1",
  "manager": "apt",
  "available_version": "1.18.0-1ubuntu1",
  "repository": "main",
  "architecture": "amd64"
}
```

## Operations

### Read Operation

Queries the current package state:
- Checks if package is installed
- Retrieves installed version
- Identifies managing package manager
- Discovers available version

### Plan Operation

Determines required changes:
- **No-op**: Package already in desired state
- **Create**: Install package (when state is `present` and package not installed)
- **Update**: Change package version or upgrade to latest
- **Delete**: Remove package (when state is `absent`)

### Apply Operation

Executes the planned changes:
- Sends `pkg.ensure` command to micro-runner
- Micro-runner executes package manager commands
- Returns result and new state
- Updates OpenFroyo resource state

### Destroy Operation

Removes the package completely, equivalent to setting state to `absent`.

## Error Handling

The provider classifies errors for proper retry behavior:

- **Transient Errors**: Network issues, repository temporarily unavailable
- **Throttled Errors**: Rate limiting from package repositories
- **Permanent Errors**: Package not found, invalid version, permission denied
- **Conflict Errors**: Dependency conflicts, locked database

## Package Manager Specific Details

### APT (Debian/Ubuntu)

```bash
# Detection
command -v apt

# Install
apt install -y <package>[=<version>]

# Remove
apt remove -y <package>

# Upgrade
apt upgrade -y <package>

# Query
dpkg-query -W -f='${Version}' <package>
```

### DNF (Fedora/RHEL 8+)

```bash
# Detection
command -v dnf

# Install
dnf install -y <package>[-<version>]

# Remove
dnf remove -y <package>

# Upgrade
dnf upgrade -y <package>

# Query
rpm -q --queryformat '%{VERSION}-%{RELEASE}' <package>
```

### YUM (RHEL 7/CentOS 7)

```bash
# Detection
command -v yum

# Install
yum install -y <package>[-<version>]

# Remove
yum remove -y <package>

# Upgrade
yum upgrade -y <package>

# Query
rpm -q --queryformat '%{VERSION}-%{RELEASE}' <package>
```

### Zypper (openSUSE/SLES)

```bash
# Detection
command -v zypper

# Install
zypper install -y <package>[=<version>]

# Remove
zypper remove -y <package>

# Update
zypper update -y <package>

# Query
rpm -q --queryformat '%{VERSION}-%{RELEASE}' <package>
```

## Development

### Prerequisites

- **TinyGo**: 0.30.0 or later
- **Go**: 1.21 or later (for testing)
- **wabt**: WebAssembly Binary Toolkit (optional, for validation)
- **binaryen**: wasm-opt optimizer (optional)

### Building

```bash
# Standard build
make build

# Development build (with debug symbols)
make dev

# Optimized build
make optimize

# Run tests
make test

# Validate WASM module
make validate

# Package for distribution
make package
```

### Testing

```bash
# Run Go tests (non-WASM)
go test -v ./...

# Test with OpenFroyo
froyo plan --config examples/nginx.cue
froyo apply --plan plan.json
```

### Project Structure

```
linux.pkg/
├── main.go                    # Provider implementation
├── manifest.yaml              # Provider manifest
├── schemas/
│   ├── package.json          # Resource configuration schema
│   └── package-state.json    # Resource state schema
├── build.sh                   # Build script
├── Makefile                   # Build automation
├── README.md                  # This file
├── LICENSE                    # Apache 2.0 license
└── plugin.wasm               # Compiled WASM module (generated)
```

## Capabilities

This provider requires the following capability:

- **exec:micro-runner**: Delegate package operations to micro-runner

## Limitations

- Package version syntax varies by package manager
- Repository management is package-manager specific
- Some advanced package manager features may not be exposed
- Cache management is simplified compared to native tools

## Security Considerations

- WASM sandbox prevents direct system access
- All privileged operations delegated to micro-runner
- Micro-runner runs with controlled permissions
- Package verification depends on underlying package manager
- No credential management within provider (handled by system)

## Performance

- **Binary Size**: ~500KB (WASM module)
- **Memory**: <10MB runtime allocation
- **Execution**: Minimal overhead, actual time dominated by package operations
- **Caching**: Package cache updated based on configuration

## Troubleshooting

### Package Not Found

```
Error: Package 'xyz' not found
```

**Solution**: Verify package name and repository configuration. Update package cache.

### Permission Denied

```
Error: Permission denied when installing package
```

**Solution**: Ensure micro-runner has appropriate sudo privileges configured during onboarding.

### Version Conflict

```
Error: Cannot install version X.Y.Z, conflicts with dependency
```

**Solution**: Check package dependencies and consider removing conflicting packages first.

### Repository Unavailable

```
Error: Failed to fetch package metadata
```

**Solution**: Check network connectivity and repository configuration. This is typically a transient error and will be retried.

## Changelog

### Version 1.0.0 (Initial Release)

- Support for apt, dnf, yum, zypper package managers
- Install, remove, and upgrade operations
- Version pinning support
- Repository configuration
- Custom package manager options
- Comprehensive error handling and retry logic
- WASM compilation for secure execution
- Full OpenFroyo Provider interface implementation

## License

Apache License 2.0 - See LICENSE file for details

## Contributing

Contributions are welcome! Please see the main OpenFroyo repository for contribution guidelines.

## Support

- Documentation: https://docs.openfroyo.io
- Issues: https://github.com/openfroyo/openfroyo/issues
- Discussions: https://github.com/openfroyo/openfroyo/discussions
