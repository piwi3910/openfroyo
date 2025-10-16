# Apache Web Server Demo

This example demonstrates using OpenFroyo to manage an Apache web server installation.

## Overview

The configuration manages:
- Apache 2.4 HTTP Server
- Apache utilities (ab, htpasswd, etc.)
- mod_ssl for HTTPS support

## Prerequisites

- Target hosts with:
  - Labels: `role=web` and `env=production`
  - Debian/Ubuntu OS (for apt package manager)
  - SSH access configured
  - Sudo privileges

## Usage

### 1. Validate Configuration

```bash
froyo validate examples/apache/apache.cue
```

### 2. Generate Execution Plan

```bash
froyo plan --out apache-plan.json examples/apache/apache.cue
```

### 3. Review the Plan

```bash
cat apache-plan.json | jq .
```

### 4. Apply the Configuration

```bash
froyo apply --plan apache-plan.json
```

### 5. Verify Installation

On the target host:

```bash
# Check Apache is installed
apache2 -v

# Check utilities are available
ab -V

# Check SSL module
apache2ctl -M | grep ssl
```

## Configuration Details

### Resources

The configuration defines three resources:

1. **apache2-package** - Base Apache HTTP Server
2. **apache2-utils** - Apache utilities
3. **apache2-ssl** - mod_ssl for HTTPS

### Dependencies

The configuration uses `require` dependencies to ensure proper installation order:

```
apache2-package
    ↓
apache2-utils
apache2-ssl
```

### Target Selection

Resources are deployed to hosts matching:
```cue
labels: {
    role: "web"
    env:  "production"
}
```

## Expected Output

During execution, you'll see:

```
[INFO] Starting run run-abc123
[INFO] Evaluating configuration
[INFO] Discovering facts from targets
[INFO] Building execution plan (3 units)
[INFO] Executing plan units in parallel
[INFO] [apache2-package] Installing apache2...
[INFO] [apache2-package] Completed (10.2s)
[INFO] [apache2-utils] Installing apache2-utils...
[INFO] [apache2-utils] Completed (3.1s)
[INFO] [apache2-ssl] Installing libapache2-mod-ssl...
[INFO] [apache2-ssl] Completed (2.8s)
[INFO] Run completed successfully (16.1s total)
```

## Testing Changes

To test configuration changes without applying:

```bash
froyo plan --out apache-plan.json --dry-run examples/apache/apache.cue
```

## Drift Detection

To check if the actual state matches desired state:

```bash
froyo drift detect examples/apache/apache.cue
```

## Removing Apache

Create a removal configuration:

```cue
resources: {
    apache_pkg: {
        // ... same config
        config: {
            package: "apache2"
            state:   "absent"  // Changed from "present"
        }
    }
}
```

Then plan and apply as usual.

## Troubleshooting

### SSH Connection Issues

If you see SSH connection errors:
1. Verify SSH key is configured: `ssh user@target-host`
2. Check target labels match: `froyo facts list`

### Package Installation Failures

If package installation fails:
1. Check package name for your distribution
2. Verify sudo permissions on target
3. Check package manager logs on target host

### Policy Violations

If policies block execution:
1. Review policy violations: Check plan output
2. Update labels or configuration to comply
3. Contact admin if policy should be updated

## Related Examples

- **examples/configs/web-stack.cue** - Full LAMP stack
- **examples/configs/multi-environment.cue** - Multi-env deployments
- **providers/linux.pkg/examples/** - More package management examples
