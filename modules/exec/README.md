# Exec Module

Execute shell commands on remote hosts.

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/) to be installed.

```bash
make build
```

This will produce `wasm/exec.wasm`.

## Variables

- `cmd` (required): The shell command to execute

## Example Usage

```yaml
- name: Check kernel version
  module: exec
  vars:
    cmd: uname -r
```

## Output

Returns the command output in `facts.stdout`.
