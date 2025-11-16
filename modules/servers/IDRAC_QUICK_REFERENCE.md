# Dell iDRAC Modules - Quick Reference

## Module: idrac

### Basic Usage
```yaml
module: idrac
vars:
  baseuri: "192.168.1.100"
  username: "root"
  password: "calvin"
  command: "<command_name>"
```

### Common Commands

**Power Management**
```yaml
command: "get_power_state"         # Check power status
command: "power_on"                # Power on
command: "power_off"               # Force power off
command: "graceful_shutdown"       # Graceful shutdown
command: "power_cycle"             # Restart
```

**System Info**
```yaml
command: "get_system_inventory"    # Hardware details
command: "get_firmware_inventory"  # Firmware versions
command: "get_sel"                 # System Event Log
command: "clear_sel"               # Clear Event Log
```

**Virtual Media**
```yaml
command: "insert_virtual_media"
media_type: "CD"                   # CD, DVD, Floppy
image_url: "http://server/image.iso"

command: "eject_virtual_media"
media_type: "CD"
```

**RAID Management**
```yaml
command: "get_storage_controllers"
command: "get_physical_disks"
controller_id: "RAID.Integrated.1-1"

command: "create_virtual_disk"
controller_id: "RAID.Integrated.1-1"
raid_level: "1"                    # 0, 1, 5, 6, 10, 50, 60
disks: "Disk.Bay.0:...,Disk.Bay.1:..."
vd_name: "OS-Volume"
```

**User Management**
```yaml
command: "create_user"
new_username: "operator1"
new_password: "SecurePass123"
privilege: "Operator"              # Administrator, Operator, ReadOnly
```

**Network Config**
```yaml
command: "configure_network"
ip_address: "192.168.1.150"
netmask: "255.255.255.0"
gateway: "192.168.1.1"

# OR for DHCP
dhcp_enabled: true
```

## Module: idrac_server_config_profile

### Basic Usage
```yaml
module: idrac_server_config_profile
vars:
  idrac_ip: "192.168.1.100"
  idrac_user: "root"
  idrac_password: "calvin"
  command: "<command_name>"
```

### Commands

**Export Configuration**
```yaml
command: "export"
export_format: "XML"               # XML or JSON
scp_components: "ALL"              # ALL, IDRAC, BIOS, NIC, RAID
share_name: "/backup/config.xml"
```

**Preview Import**
```yaml
command: "preview"
scp_file: "/backup/config.xml"
```

**Import Configuration**
```yaml
command: "import"
scp_file: "/backup/config.xml"
shutdown_type: "Graceful"          # Graceful, Forced, NoReboot
end_host_power_state: "On"         # On, Off
job_wait: true
job_wait_timeout: 3600
```

## Complete Example Stack

```yaml
name: Dell Server Management
inventory: inventory/hosts.yml

defaults:
  username: "root"
  password: "calvin"
  validate_certs: false

run:
  # Backup all server configurations
  - name: Export configurations
    module: idrac_server_config_profile
    hosts: "@group:dell_servers"
    vars:
      idrac_ip: "{{ host.idrac_ip }}"
      idrac_user: "{{ var.username }}"
      idrac_password: "{{ var.password }}"
      command: "export"
      export_format: "XML"
      scp_components: "ALL"
      share_name: "/backup/{{ host.name }}_{{ date }}.xml"

  # Power management
  - name: Check power states
    module: idrac
    hosts: "@group:dell_servers"
    vars:
      baseuri: "{{ host.idrac_ip }}"
      username: "{{ var.username }}"
      password: "{{ var.password }}"
      command: "get_power_state"

  # Get system inventory
  - name: Collect hardware inventory
    module: idrac
    hosts: "@group:dell_servers"
    vars:
      baseuri: "{{ host.idrac_ip }}"
      username: "{{ var.username }}"
      password: "{{ var.password }}"
      command: "get_system_inventory"
```

## Build Commands

```bash
# Build idrac module
cd modules/servers/idrac
make build

# Build idrac_server_config_profile module
cd modules/servers/idrac_server_config_profile
make build
```

## File Locations

```
modules/servers/idrac/
  wasm/idrac.wasm                    # 503KB binary
  README.md                          # Full documentation
  test_power.ofy                     # Power test
  test_system.ofy                    # System test

modules/servers/idrac_server_config_profile/
  wasm/idrac_server_config_profile.wasm  # 481KB binary
  README.md                          # Full documentation
  test_export.ofy                    # Export test
  test_import.ofy                    # Import test
```

## Default Credentials

- Username: `root`
- Password: `calvin` (Dell factory default)

## Common iDRAC IPs

Typically assigned via:
- DHCP (check router)
- Static: 192.168.0.120 (common default)
- Via server LCD/BIOS configuration

## Tips

1. Always run `preview` before `import` in production
2. Use `validate_certs: false` for self-signed certificates
3. Export before making changes (backup)
4. RAID operations may require system reboot
5. SCP import can take 5-30 minutes
6. Job tracking shows progress every 15 seconds
7. Keep `job_wait: true` for configuration changes
8. Use `NoReboot` shutdown type for iDRAC-only changes

## Error Codes

- HTTP 200: Success
- HTTP 202: Accepted (job created)
- HTTP 204: Success (no content)
- HTTP 401: Authentication failed
- HTTP 404: Resource not found
- HTTP 500: Internal server error

## Support Matrix

| iDRAC Version | idrac Module | SCP Module |
|---------------|--------------|------------|
| iDRAC 7       | ✓            | ✓          |
| iDRAC 8       | ✓            | ✓          |
| iDRAC 9       | ✓            | ✓          |

## Requirements

- Dell PowerEdge server
- iDRAC Enterprise license (for SCP)
- Network access to iDRAC
- Administrator privileges

## Getting Help

See full documentation:
- `modules/servers/idrac/README.md`
- `modules/servers/idrac_server_config_profile/README.md`
- `modules/servers/IDRAC_MODULES_SUMMARY.md`
