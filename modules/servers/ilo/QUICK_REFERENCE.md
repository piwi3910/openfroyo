# HP iLO Module - Quick Reference

## Most Common Operations

### Power Control
```yaml
# Power On
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "power_on"}

# Graceful Shutdown
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "graceful_shutdown"}

# Force Restart
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "force_restart"}
```

### Virtual Media Boot
```yaml
# 1. Insert ISO
- module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "admin"
    ilo_password: "pass"
    command: "insert_virtual_cd"
    iso_url: "http://repo.local/ubuntu.iso"

# 2. Set Boot Device
- module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "admin"
    ilo_password: "pass"
    command: "set_onetime_boot"
    boot_device: "Cd"

# 3. Power On
- module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "admin"
    ilo_password: "pass"
    command: "power_on"
```

### System Information
```yaml
# Full System Info
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "get_system_info"}

# Serial Number
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "get_serial_number"}

# Hardware Inventory
- module: ilo
  vars: {ilo_ip: "10.0.1.100", ilo_user: "admin", ilo_password: "pass", command: "get_hardware_inventory"}
```

### User Management
```yaml
# Create User
- module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "admin"
    ilo_password: "pass"
    command: "create_user"
    new_username: "operator"
    new_password: "SecurePass123!"
    privilege: "Operator"  # Administrator, Operator, or ReadOnly
```

## Command Quick List

### Power Management
- `power_on` - Power on server
- `power_off` - Force power off
- `graceful_shutdown` - OS shutdown
- `force_restart` - Force restart
- `power_button` - Press power button
- `cold_boot` - Cold boot
- `get_power_state` - Get power state

### Virtual Media
- `insert_virtual_cd` - Mount ISO (requires `iso_url`)
- `eject_virtual_cd` - Eject CD
- `insert_virtual_floppy` - Mount floppy (requires `iso_url`)
- `eject_virtual_floppy` - Eject floppy
- `get_virtual_media_status` - Get media status

### System Info
- `get_system_info` - Complete system info
- `get_health_status` - Health status
- `get_hardware_inventory` - Hardware summary
- `get_firmware_versions` - All firmware versions
- `get_serial_number` - Serial number
- `get_product_info` - Model and UUID

### Users
- `get_users` - List all users
- `create_user` - Create user (requires `new_username`, `new_password`, `privilege`)
- `modify_user` - Modify user (requires `user_id`, `new_password`)
- `delete_user` - Delete user (requires `user_id`)

### Network
- `get_network_settings` - Network config
- `set_network_static` - Static IP (requires `ip_address`, `subnet_mask`, `gateway`)
- `set_network_dhcp` - Enable DHCP
- `set_hostname` - Set hostname (requires `hostname`)
- `set_dns_servers` - Set DNS (requires `dns_servers`)

### Firmware
- `get_firmware_inventory` - Firmware versions
- `update_firmware` - Update firmware (requires `firmware_url`)
- `get_update_progress` - Update status

### Boot
- `set_onetime_boot` - One-time boot (requires `boot_device`: Cd, Hdd, Pxe, BiosSetup)
- `set_persistent_boot` - Persistent boot (requires `boot_device`)
- `get_boot_config` - Boot configuration
- `set_uefi_boot` - Enable UEFI (requires `enabled`: true/false)

### Logs
- `get_iml_log` - Get IML
- `clear_iml_log` - Clear IML
- `get_ilo_event_log` - Get iLO log
- `clear_ilo_event_log` - Clear iLO log

### License
- `get_license_status` - License info
- `install_license` - Install license (requires `license_key`)

### iLO Settings
- `reset_ilo` - Reset iLO
- `get_ilo_datetime` - Get date/time
- `set_ilo_datetime` - Set date/time (requires `datetime`)
- `get_ilo_info` - iLO information
- `ping_from_ilo` - Network test (requires `target_host`)

### Security
- `get_security_dashboard` - Security status
- `generate_csr` - Generate CSR (requires `common_name`)
- `import_certificate` - Import cert (requires `certificate`)

## Common Variable Combinations

### Required Variables (All Commands)
```yaml
ilo_ip: "10.0.1.100"        # iLO IP or hostname
ilo_user: "Administrator"   # iLO username
ilo_password: "password"    # iLO password
command: "command_name"     # Operation to perform
```

### Optional Variables
```yaml
timeout: 60                 # Request timeout (default: 60)
validate_certs: false       # SSL validation (default: true)
```

## Boot Device Values
- `Cd` - CD/DVD (virtual or physical)
- `Hdd` - Hard disk
- `Pxe` - Network boot
- `BiosSetup` - BIOS setup
- `UefiShell` - UEFI shell

## User Privilege Levels
- `Administrator` - Full access
- `Operator` - Limited configuration
- `ReadOnly` - View only

## Common Workflows

### PXE Boot Installation
```yaml
- {command: "set_onetime_boot", boot_device: "Pxe"}
- {command: "power_on"}
```

### ISO Installation with Virtual Media
```yaml
- {command: "insert_virtual_cd", iso_url: "http://repo/os.iso"}
- {command: "set_onetime_boot", boot_device: "Cd"}
- {command: "power_on"}
# After installation:
- {command: "eject_virtual_cd"}
```

### Gather Server Facts
```yaml
- {command: "get_system_info"}
- {command: "get_hardware_inventory"}
- {command: "get_firmware_versions"}
- {command: "get_serial_number"}
```

### Network Reconfiguration
```yaml
- command: "set_network_static"
  ip_address: "10.0.1.101"
  subnet_mask: "255.255.255.0"
  gateway: "10.0.1.1"
- command: "set_hostname"
  hostname: "ilo-server01"
- command: "set_dns_servers"
  dns_servers: '"8.8.8.8","8.8.4.4"'
```

## Troubleshooting

### Test iLO Connectivity
```bash
curl -k https://ILO_IP/redfish/v1/
```

### Disable SSL Validation
```yaml
validate_certs: false
```

### Increase Timeout for Long Operations
```yaml
timeout: 300  # 5 minutes for firmware updates
```

## HTTP Status Codes
- `200` - Success
- `201` - Created (user, resource)
- `202` - Accepted (async operation started)
- `400` - Bad request (check parameters)
- `401` - Authentication failed
- `403` - Forbidden (insufficient privileges)
- `404` - Not found (endpoint or resource)
- `500` - iLO internal error
