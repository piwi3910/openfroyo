# Exec Module

Execute shell commands on remote hosts. Works cross-platform on Linux, macOS, and Windows.

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/) to be installed.

```bash
make build
```

This will produce `wasm/exec.wasm`.

## Variables

- `cmd` (required): The shell command to execute

## Cross-Platform Support

The module automatically adapts to the target platform:
- **Linux/macOS/FreeBSD**: Executes commands directly
- **Windows**: Uses `cmd.exe /c` for proper command execution

## Example Usage

```yaml
# Linux/macOS
- name: Check kernel version
  module: generic/exec
  hosts:
    - linux-server
  vars:
    cmd: uname -r

# Windows
- name: Check Windows version
  module: generic/exec
  hosts:
    - windows-server
  vars:
    cmd: ver

# Cross-platform hostname
- name: Get hostname
  module: generic/exec
  hosts:
    - "@group:all_servers"
  vars:
    cmd: hostname
```

## Output

Returns the command output in `facts.stdout`.
