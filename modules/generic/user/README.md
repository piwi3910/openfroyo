# User Module

Manage system users on Linux systems. Create, modify, and delete user accounts with full control over user properties.

## Supported Platforms

| Platform | Tool | Commands |
|----------|------|----------|
| **Linux** | useradd/usermod/userdel | `useradd`, `usermod`, `userdel` |

## Variables

- `name` (required): Username to create/manage/delete
- `state` (optional): Desired state - `present` (default) or `absent`
- `uid` (optional): User ID number
- `group` (optional): Primary group name or GID
- `groups` (optional): Array of supplementary group names
- `shell` (optional): Login shell (default: `/bin/bash`)
- `home` (optional): Home directory path (default: system default)
- `create_home` (optional): Create home directory (boolean, default: true)
- `password` (optional): Encrypted password (use `openssl passwd` or similar to generate)
- `comment` (optional): User comment/GECOS field (full name, description)
- `system` (optional): Create as system user (boolean, default: false)

## States

- **present**: Create the user if it doesn't exist, update if it does
- **absent**: Remove the user and their home directory

## Examples

### Create a basic user

```yaml
- name: Create developer user
  module: generic/user
  hosts:
    - dev-server
  vars:
    name: john
    state: present
```

### Create user with custom settings

```yaml
- name: Create application user
  module: generic/user
  hosts:
    - "@group:app-servers"
  vars:
    name: appuser
    state: present
    uid: "1500"
    group: appgroup
    groups:
      - docker
      - sudo
    shell: /bin/bash
    home: /opt/app
    create_home: true
    comment: "Application Service User"
```

### Create system user (for services)

```yaml
- name: Create nginx service user
  module: generic/user
  hosts:
    - web-server
  vars:
    name: nginx
    state: present
    system: true
    create_home: false
    shell: /usr/sbin/nologin
    comment: "Nginx Web Server"
```

### Create user with password

```yaml
- name: Create user with password
  module: generic/user
  hosts:
    - server-01
  vars:
    name: testuser
    state: present
    password: "$6$rounds=656000$..."  # Use encrypted password hash
    comment: "Test User"
```

### Remove a user

```yaml
- name: Remove old user account
  module: generic/user
  hosts:
    - server-01
  vars:
    name: olduser
    state: absent
```

### Create multiple users

```yaml
- name: Create developer users
  module: generic/user
  hosts:
    - "@group:dev-servers"
  vars:
    name: "{{ item }}"
    state: present
    groups:
      - developers
      - docker
  loop:
    - alice
    - bob
    - charlie
```

## How It Works

1. **Check Existence**: Uses `id` command to check if user exists
2. **Create/Modify**:
   - New users: Uses `useradd` with specified options
   - Existing users: Uses `usermod` to update properties
3. **Set Password**: Uses `chpasswd -e` for encrypted passwords
4. **Remove**: Uses `userdel -r` to remove user and home directory

## Password Generation

To generate encrypted passwords for the `password` variable:

```bash
# Using openssl (SHA-512)
openssl passwd -6 -salt xyz yourpassword

# Using mkpasswd (if available)
mkpasswd --method=sha-512

# Using Python
python3 -c 'import crypt; print(crypt.crypt("yourpassword", crypt.mksalt(crypt.METHOD_SHA512)))'
```

## Output

Returns facts:
- `username`: User that was managed
- `action`: Action performed (created, updated, deleted, unchanged)
- `stdout`: Command output

## Building

Requires [TinyGo](https://tinygo.org/getting-started/install/):

```bash
make build
```

This produces `wasm/user.wasm`.

## Security Notes

- Always use encrypted passwords (never plain text)
- System users should use `/usr/sbin/nologin` or `/bin/false` as shell
- Be careful with sudo group membership
- Consider using SSH keys instead of passwords
- The `userdel -r` command removes the home directory and mail spool
