# System Administration Modules

This document describes the 17 system administration modules added to OpenFroyo for comprehensive server management.

## Overview

OpenFroyo now includes a complete set of system administration modules for managing users, firewalls, services, and system configuration. These modules enable infrastructure-as-code for Linux server administration.

## Module Categories

### User & Group Management (3 modules)

#### user
**Purpose:** Manage system users

**Key Features:**
- Create, modify, and delete users
- Set UID, groups, shell, home directory
- Password management
- System user support

**Common Operations:**
```yaml
# Create user
- module: generic/user
  vars:
    name: webadmin
    groups: [sudo, www-data]
    shell: /bin/bash
    create_home: true

# Remove user
- module: generic/user
  vars:
    name: olduser
    state: absent
```

#### group
**Purpose:** Manage system groups

**Key Features:**
- Create, modify, and delete groups
- Set GID
- System group support

#### authorized_keys
**Purpose:** Manage SSH authorized_keys files

**Key Features:**
- Add/remove SSH public keys
- Key options support
- Exclusive mode (replace all keys)
- Automatic permission management

---

### Firewall Management (3 modules)

#### firewalld
**Purpose:** Manage firewalld (RHEL/CentOS/Fedora)

**Key Features:**
- Manage services and ports
- Zone-based configuration
- Permanent and immediate changes
- Automatic reload

**Example:**
```yaml
- module: generic/firewalld
  vars:
    service: http
    zone: public
    state: enabled
    permanent: true
    immediate: true
```

#### iptables
**Purpose:** Manage iptables firewall rules

**Key Features:**
- Full control over chains (INPUT, OUTPUT, FORWARD)
- Protocol, port, source/destination filtering
- Multiple table support (filter, nat, mangle)
- Automatic persistence

**Example:**
```yaml
- module: generic/iptables
  vars:
    chain: INPUT
    protocol: tcp
    destination_port: "22"
    jump: ACCEPT
    state: present
```

#### ufw
**Purpose:** Manage UFW (Ubuntu/Debian)

**Key Features:**
- Simple allow/deny/reject/limit rules
- Port and protocol filtering
- Source/destination IP filtering
- Rate limiting

**Example:**
```yaml
- module: generic/ufw
  vars:
    rule: allow
    port: "443"
    proto: tcp
```

---

### System Services (2 modules)

#### systemd
**Purpose:** Manage systemd services

**Key Features:**
- Start/stop/restart/reload services
- Enable/disable at boot
- Mask/unmask services
- Daemon reload

**Example:**
```yaml
- module: generic/systemd
  vars:
    name: nginx
    state: started
    enabled: true
```

#### cron
**Purpose:** Manage cron jobs

**Key Features:**
- Add/update/remove cron jobs
- Traditional schedule syntax
- Special time strings (@hourly, @daily, etc.)
- Per-user crontab management

**Example:**
```yaml
- module: generic/cron
  vars:
    name: backup-job
    job: /usr/local/bin/backup.sh
    hour: "2"
    minute: "0"
    user: root
```

---

### System Configuration (4 modules)

#### hostname
**Purpose:** Set system hostname

**Key Features:**
- Set hostname using hostnamectl
- Update /etc/hostname

**Example:**
```yaml
- module: generic/hostname
  vars:
    name: webserver-01.example.com
```

#### timezone
**Purpose:** Configure system timezone

**Key Features:**
- Set timezone using timedatectl
- Symlink /etc/localtime

**Example:**
```yaml
- module: generic/timezone
  vars:
    name: America/New_York
```

#### sysctl
**Purpose:** Manage kernel parameters

**Key Features:**
- Set kernel parameters
- Persist to /etc/sysctl.conf
- Immediate and permanent changes
- Reload support

**Example:**
```yaml
- module: generic/sysctl
  vars:
    name: net.ipv4.ip_forward
    value: "1"
    state: present
    reload: true
```

#### selinux
**Purpose:** Configure SELinux

**Key Features:**
- Set enforcing/permissive/disabled
- Update config file
- Policy selection

**Example:**
```yaml
- module: generic/selinux
  vars:
    state: enforcing
    policy: targeted
```

---

### Storage Management (2 modules)

#### filesystems
**Purpose:** Manage filesystem mounts

**Key Features:**
- Mount/unmount filesystems
- Manage /etc/fstab entries
- Multiple states (mounted, unmounted, present, absent)
- Automatic mount point creation

**Example:**
```yaml
- module: generic/filesystems
  vars:
    device: /dev/sdb1
    path: /mnt/data
    fstype: ext4
    opts: defaults
    state: mounted
    fstab: true
```

#### lvm
**Purpose:** Manage LVM (Logical Volume Manager)

**Key Features:**
- Create/remove physical volumes
- Create/extend/remove volume groups
- Create/extend/remove logical volumes
- Flexible size specifications

**Example:**
```yaml
# Create VG and LV
- module: generic/lvm
  vars:
    vg: data_vg
    lv: data_lv
    pvs: [/dev/sdb, /dev/sdc]
    size: 100G
    state: present
```

---

### Utilities (2 modules)

#### ping
**Purpose:** Test network connectivity

**Key Features:**
- ICMP ping test
- Configurable count and timeout
- Interface binding
- RTT statistics

**Example:**
```yaml
- module: generic/ping
  vars:
    host: 8.8.8.8
    count: 4
    timeout: 5
```

#### reboot
**Purpose:** Reboot or shutdown systems

**Key Features:**
- Reboot, shutdown, or halt
- Delayed execution
- User messages
- Force mode

**Example:**
```yaml
- module: generic/reboot
  vars:
    state: reboot
    delay: 1
    message: "System rebooting for updates"
```

---

## Build Information

All modules compiled successfully to WebAssembly:

```
authorized_keys.wasm  462K
cron.wasm             463K
filesystems.wasm      473K
firewalld.wasm        462K
group.wasm            458K
hostname.wasm         458K
iptables.wasm         461K
lvm.wasm              477K
ping.wasm             464K
reboot.wasm           467K
selinux.wasm          465K
sysctl.wasm           463K
systemd.wasm          464K
timezone.wasm         458K
ufw.wasm              460K
user.wasm             464K
```

**Total:** 7.5MB for all 16 modules

## Testing

Comprehensive test stack: `stacks/test_sysadmin_modules.ofy`

**Test Results:**
✅ All modules tested successfully
✅ User/group creation and removal
✅ Cron job management
✅ Network connectivity tests
✅ System information retrieval

## Usage Patterns

### Server Hardening
```yaml
# Secure SSH configuration
- module: generic/user
  vars:
    name: admin
    groups: [sudo]
    shell: /bin/bash

- module: generic/authorized_keys
  vars:
    user: admin
    key: "ssh-rsa AAAA..."
    exclusive: true

- module: generic/firewalld
  vars:
    service: ssh
    state: enabled
    permanent: true
```

### Web Server Setup
```yaml
# Install and configure nginx
- module: generic/package
  vars:
    name: nginx
    state: present

- module: generic/systemd
  vars:
    name: nginx
    state: started
    enabled: true

- module: generic/firewalld
  vars:
    service: http
    state: enabled

- module: generic/firewalld
  vars:
    service: https
    state: enabled
```

### Backup Automation
```yaml
# Schedule nightly backups
- module: generic/cron
  vars:
    name: nightly-backup
    job: /usr/local/bin/backup.sh
    special_time: daily
    user: root

# Mount backup volume
- module: generic/filesystems
  vars:
    device: /dev/sdb1
    path: /mnt/backups
    fstype: ext4
    state: mounted
```

### Performance Tuning
```yaml
# Enable IP forwarding for routing
- module: generic/sysctl
  vars:
    name: net.ipv4.ip_forward
    value: "1"

# Increase file descriptor limits
- module: generic/sysctl
  vars:
    name: fs.file-max
    value: "65536"
```

## Security Considerations

### User Management
- Always use strong passwords (encrypted)
- Set appropriate shell and home directory
- Use `authorized_keys` with exclusive mode for better security
- Remove unused users promptly

### Firewall Management
- Default deny policy recommended
- Open only necessary ports
- Use zones (firewalld) or chains (iptables) appropriately
- Test rules before making permanent

### System Services
- Enable only required services
- Use `masked` to prevent service starts
- Regular service audits

### Storage Management
- Use UUIDs instead of device paths in fstab
- Test mount operations before making permanent
- Always backup data before LVM operations

## Platform Support

| Module | RHEL/CentOS | Ubuntu/Debian | Notes |
|--------|-------------|---------------|-------|
| user | ✅ | ✅ | Universal |
| group | ✅ | ✅ | Universal |
| authorized_keys | ✅ | ✅ | Universal |
| firewalld | ✅ | ⚠️ | RHEL family default |
| iptables | ✅ | ✅ | Universal |
| ufw | ⚠️ | ✅ | Ubuntu/Debian default |
| systemd | ✅ | ✅ | Modern systems |
| cron | ✅ | ✅ | Universal |
| hostname | ✅ | ✅ | Universal |
| timezone | ✅ | ✅ | Universal |
| sysctl | ✅ | ✅ | Universal |
| selinux | ✅ | ⚠️ | RHEL family |
| filesystems | ✅ | ✅ | Universal |
| lvm | ✅ | ✅ | Universal |
| ping | ✅ | ✅ | Universal |
| reboot | ✅ | ✅ | Universal |

## GitHub Issue

This work was tracked in: [Issue #14 - Add system administration modules](https://github.com/piwi3910/openfroyo/issues/14)

## Related Documentation

- [Generic Modules README](../modules/generic/README.md)
- [Package Manager Modules](PACKAGE_MANAGER_MODULES.md)
- [Module Creation Summary](MODULE_CREATION_SUMMARY.md)

## Future Enhancements

Potential additions:
- **service** - Generic service management (SysV init, upstart)
- **mount** - Additional mount options and features
- **lvg** / **lvol** - Separate LVM group and volume modules
- **parted** - Partition management
- **sudoers** - Manage sudo configuration
- **modprobe** - Manage kernel modules
- **syslog** - Manage syslog configuration
