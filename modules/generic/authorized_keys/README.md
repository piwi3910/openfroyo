# Authorized Keys Module

Manage SSH authorized_keys files for users. Add, remove, or replace SSH public keys with support for key options and exclusive mode.

## Supported Platforms

| Platform | File | Location |
|----------|------|----------|
| **Linux** | authorized_keys | `~/.ssh/authorized_keys` |

## Variables

- `user` (required): Username whose authorized_keys file to manage
- `key` (required): SSH public key to add/remove
- `state` (optional): Desired state - `present` (default) or `absent`
- `key_options` (optional): SSH key options (e.g., `no-port-forwarding,command="/usr/bin/script"`)
- `exclusive` (optional): Replace all other keys (boolean, default: false)
- `path` (optional): Custom authorized_keys file path (default: `~/.ssh/authorized_keys`)

## States

- **present**: Add the SSH key if not already present
- **absent**: Remove the SSH key if present

## Examples

### Add a basic SSH key

```yaml
- name: Add SSH key for deploy user
  module: generic/authorized_keys
  hosts:
    - web-server
  vars:
    user: deploy
    key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7xK... deploy@laptop"
    state: present
```

### Add SSH key with options

```yaml
- name: Add restricted SSH key
  module: generic/authorized_keys
  hosts:
    - "@group:production"
  vars:
    user: backup
    key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFqP... backup@server"
    state: present
    key_options: "no-port-forwarding,no-X11-forwarding,no-agent-forwarding"
```

### Add SSH key with forced command

```yaml
- name: Add SSH key with forced command
  module: generic/authorized_keys
  hosts:
    - git-server
  vars:
    user: git
    key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQD... user@client"
    state: present
    key_options: 'command="/usr/local/bin/git-shell -c \"$SSH_ORIGINAL_COMMAND\""'
```

### Replace all keys (exclusive mode)

```yaml
- name: Set only one authorized key
  module: generic/authorized_keys
  hosts:
    - secure-server
  vars:
    user: admin
    key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIA... admin@secure"
    state: present
    exclusive: true
```

### Add key to custom path

```yaml
- name: Add key to custom location
  module: generic/authorized_keys
  hosts:
    - app-server
  vars:
    user: appuser
    key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... app@client"
    state: present
    path: /opt/app/.ssh/authorized_keys
```

### Remove an SSH key

```yaml
- name: Remove old SSH key
  module: generic/authorized_keys
  hosts:
    - "@group:all-servers"
  vars:
    user: deploy
    key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... oldkey@laptop"
    state: absent
```

### Manage multiple keys for a user

```yaml
- name: Add multiple SSH keys
  module: generic/authorized_keys
  hosts:
    - dev-server
  vars:
    user: developer
    key: "{{ item }}"
    state: present
  loop:
    - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... dev@laptop"
    - "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... dev@desktop"
    - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQD... dev@tablet"
```

## How It Works

1. **Determine Path**: Uses user's home directory to find `.ssh/authorized_keys` (or uses custom path)
2. **Create Directory**: Ensures `.ssh` directory exists with correct permissions (700)
3. **Create File**: Ensures `authorized_keys` file exists with correct permissions (600)
4. **Add Key**:
   - Checks if key already exists (avoids duplicates)
   - Prepends key options if specified
   - Adds key to file
   - Removes duplicate entries
5. **Remove Key**: Uses `grep -vF` to remove matching keys
6. **Set Ownership**: Ensures file is owned by the user
7. **Set Permissions**: Ensures correct permissions (600 for file, 700 for directory)

## SSH Key Options

Common key options you can use with `key_options`:

| Option | Description |
|--------|-------------|
| `no-port-forwarding` | Disable SSH port forwarding |
| `no-X11-forwarding` | Disable X11 forwarding |
| `no-agent-forwarding` | Disable SSH agent forwarding |
| `no-pty` | Disable PTY allocation |
| `command="cmd"` | Force specific command execution |
| `from="pattern"` | Restrict source IP addresses |
| `environment="NAME=value"` | Set environment variables |

Multiple options are comma-separated: `"no-port-forwarding,no-X11-forwarding,command=\"/usr/bin/script\""`

## Key Formats

Supported SSH key types:
- `ssh-rsa` (RSA)
- `ssh-ed25519` (Ed25519, recommended)
- `ecdsa-sha2-nistp256` (ECDSA)
- `ecdsa-sha2-nistp384` (ECDSA)
- `ecdsa-sha2-nistp521` (ECDSA)

## Output

Returns facts:
- `user`: User whose keys were managed
- `action`: Action performed (added, removed, unchanged)
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/authorized_keys.wasm`.

## Security Notes

- Always use strong SSH keys (Ed25519 recommended, minimum 2048-bit RSA)
- Use key options to restrict key usage when possible
- Never share private keys
- Regularly rotate SSH keys
- Use `exclusive: true` carefully - it removes all other keys
- Ensure `.ssh` directory is mode 700 and `authorized_keys` is mode 600
- The module automatically sets correct permissions and ownership
- Use forced commands for automation keys to limit what they can do

## Exclusive Mode Warning

When `exclusive: true` is set, ALL existing keys in the authorized_keys file will be removed and replaced with only the specified key. Use this option carefully, especially in production environments, as it can lock out other users or systems.
