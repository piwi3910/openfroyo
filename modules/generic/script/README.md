# Script Module

Copy and execute local scripts on remote hosts. The script file is copied from your local machine to the remote host, made executable, and then run.

## Variables

- `script` (required): Path to local script file to execute
- `args` (optional): Arguments to pass to the script

## Examples

### Run a simple script

```yaml
- name: Run backup script
  module: generic/script
  hosts:
    - web-server
  vars:
    script: scripts/backup.sh
```

### Run script with arguments

```yaml
- name: Deploy application
  module: generic/script
  hosts:
    - app-servers
  vars:
    script: scripts/deploy.sh
    args: --env production --version 1.2.3
```

### Run Python script

```yaml
- name: Run data migration
  module: generic/script
  hosts:
    - database-server
  vars:
    script: scripts/migrate.py
    args: --database prod
```

## How It Works

1. **Upload**: The local script file is copied to the remote host at `/tmp/froyo/scripts/<scriptname>`
2. **Permission**: Script is made executable with `chmod +x`
3. **Execute**: Script runs with provided arguments
4. **Output**: Returns stdout/stderr and exit code

## Output

Returns facts:
- `script_path`: Remote path where script was uploaded
- `stdout`: Script output
- `stderr`: Script error output (if any)
- `exit_code`: Script exit code

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/script.wasm`.
