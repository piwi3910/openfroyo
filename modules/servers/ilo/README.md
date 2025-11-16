# HP iLO Module

Comprehensive HP iLO (Integrated Lights-Out) management module for OpenFroyo. Supports all major iLO operations via Redfish API and HP OEM extensions.

## Features

- **Power Management**: Full control over server power states
- **Virtual Media**: Mount/unmount ISO images and virtual devices
- **System Information**: Hardware inventory, health status, firmware versions
- **User Management**: Create, modify, delete iLO users with role-based privileges
- **Network Configuration**: Configure IP settings, DNS, NTP, hostname
- **Firmware Management**: Update iLO and system firmware
- **Boot Configuration**: Control boot order and boot devices
- **Logs and Events**: Access IML and iLO event logs
- **License Management**: Install and check iLO Advanced licenses
- **Security**: Certificate management, security settings

## Compatibility

- **iLO 4**: Full support via Redfish API
- **iLO 5**: Full support with enhanced features
- **iLO 6**: Full support with latest Redfish extensions

## Requirements

- Network access to iLO interface
- Valid iLO credentials (username and password)
- `curl` and `jq` installed on the control host
- For firmware updates: accessible HTTP/HTTPS firmware repository

## Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `ilo_ip` | string | iLO IP address or hostname |
| `ilo_user` | string | iLO username (must have appropriate privileges) |
| `ilo_password` | string | iLO password |
| `command` | string | Operation to perform (see Commands below) |

### Optional Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `timeout` | int | 60 | Request timeout in seconds |
| `validate_certs` | bool | true | Validate SSL certificates |

### Command-Specific Variables

#### Power Management
- `power_state`: Target power state (for legacy operations)

#### Virtual Media
- `iso_url`: URL to ISO image (HTTP/HTTPS accessible from iLO)

#### User Management
- `new_username`: Username for new account
- `new_password`: Password for new/modified account
- `user_id`: User ID (1-12) for modify/delete operations
- `privilege`: User privilege level (`Administrator`, `Operator`, `ReadOnly`)

#### Network Configuration
- `ip_address`: Static IP address
- `subnet_mask`: Subnet mask
- `gateway`: Default gateway
- `hostname`: iLO hostname
- `dns_servers`: Comma-separated DNS server IPs

#### Firmware Management
- `firmware_url`: URL to firmware file (accessible from iLO)

#### Boot Configuration
- `boot_device`: Boot device (`Cd`, `Hdd`, `Pxe`, `BiosSetup`, `UefiShell`)
- `enabled`: Boolean for UEFI boot mode

#### License Management
- `license_key`: iLO Advanced license key

#### Security
- `common_name`: Common Name for CSR
- `organization`: Organization name
- `organizational_unit`: Organizational unit
- `locality`: City/locality
- `state`: State/province
- `country`: Two-letter country code
- `certificate`: PEM-encoded certificate

#### iLO Settings
- `datetime`: ISO8601 datetime string
- `target_host`: Target host for network test

## Supported Commands

### Power Management Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `power_on` | Power on the server | None |
| `power_off` | Force power off (immediate) | None |
| `graceful_shutdown` | Graceful OS shutdown | None |
| `force_restart` | Force restart (immediate) | None |
| `power_button` | Momentary press power button | None |
| `cold_boot` | Cold boot (power cycle) | None |
| `get_power_state` | Get current power state | None |

### Virtual Media Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `insert_virtual_cd` | Mount ISO as virtual CD/DVD | `iso_url` |
| `eject_virtual_cd` | Eject virtual CD/DVD | None |
| `insert_virtual_floppy` | Mount image as virtual floppy | `iso_url` |
| `eject_virtual_floppy` | Eject virtual floppy | None |
| `get_virtual_media_status` | Get all virtual media status | None |

### System Information Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_system_info` | Get complete system information | None |
| `get_health_status` | Get system health status | None |
| `get_hardware_inventory` | Get hardware component summary | None |
| `get_firmware_versions` | Get all firmware versions | None |
| `get_serial_number` | Get system serial number | None |
| `get_product_info` | Get model, manufacturer, UUID | None |

### User Management Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_users` | List all iLO user accounts | None |
| `create_user` | Create new iLO user | `new_username`, `new_password`, `privilege` |
| `modify_user` | Modify existing user | `user_id`, `new_password` |
| `delete_user` | Delete user account | `user_id` |

### Network Configuration Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_network_settings` | Get iLO network configuration | None |
| `set_network_static` | Configure static IP | `ip_address`, `subnet_mask`, `gateway` |
| `set_network_dhcp` | Enable DHCP | None |
| `set_hostname` | Set iLO hostname | `hostname` |
| `set_dns_servers` | Configure DNS servers | `dns_servers` |

### Firmware Management Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_firmware_inventory` | Get all firmware versions | None |
| `update_firmware` | Update firmware from URL | `firmware_url` |
| `get_update_progress` | Check firmware update status | None |

### Boot Configuration Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `set_onetime_boot` | Set one-time boot device | `boot_device` |
| `set_persistent_boot` | Set persistent boot device | `boot_device` |
| `get_boot_config` | Get boot configuration | None |
| `set_uefi_boot` | Enable/disable UEFI boot mode | `enabled` |

### Logs and Events Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_iml_log` | Get Integrated Management Log | None |
| `clear_iml_log` | Clear IML | None |
| `get_ilo_event_log` | Get iLO Event Log | None |
| `clear_ilo_event_log` | Clear iLO Event Log | None |

### License Management Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_license_status` | Get iLO license status | None |
| `install_license` | Install iLO Advanced license | `license_key` |

### iLO Settings Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `reset_ilo` | Reset iLO (graceful restart) | None |
| `get_ilo_datetime` | Get iLO date/time | None |
| `set_ilo_datetime` | Set iLO date/time | `datetime` |
| `get_ilo_info` | Get iLO manager information | None |
| `ping_from_ilo` | Ping host from iLO | `target_host` |

### Security Operations

| Command | Description | Variables |
|---------|-------------|-----------|
| `get_security_dashboard` | Get security status | None |
| `generate_csr` | Generate SSL certificate CSR | `common_name`, `organization`, etc. |
| `import_certificate` | Import SSL certificate | `certificate` |

## Usage Examples

### Basic Power Management

```yaml
# Power on a server
- name: Power on web server
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "power_on"

# Graceful shutdown
- name: Graceful shutdown
  module: ilo
  vars:
    ilo_ip: "{{ ilo_ip }}"
    ilo_user: "{{ ilo_user }}"
    ilo_password: "{{ ilo_password }}"
    command: "graceful_shutdown"
```

### Virtual Media Operations

```yaml
# Mount installation ISO
- name: Insert installation ISO
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "insert_virtual_cd"
    iso_url: "http://repo.example.com/images/ubuntu-22.04.iso"

# Set one-time boot from CD
- name: Boot from virtual CD once
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "set_onetime_boot"
    boot_device: "Cd"

# Eject virtual media
- name: Eject virtual CD
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "eject_virtual_cd"
```

### User Management

```yaml
# Create operator user
- name: Create monitoring user
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "create_user"
    new_username: "monitor"
    new_password: "{{ vault.monitor_password }}"
    privilege: "Operator"

# Delete user
- name: Remove old user
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "delete_user"
    user_id: "5"
```

### System Information

```yaml
# Get system information
- name: Gather server facts
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "get_system_info"

# Get hardware inventory
- name: Check hardware configuration
  module: ilo
  vars:
    ilo_ip: "{{ ilo_ip }}"
    ilo_user: "{{ ilo_user }}"
    ilo_password: "{{ ilo_password }}"
    command: "get_hardware_inventory"
```

### Network Configuration

```yaml
# Configure static IP
- name: Set iLO static IP
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"  # Current IP
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "set_network_static"
    ip_address: "10.0.1.101"
    subnet_mask: "255.255.255.0"
    gateway: "10.0.1.1"

# Set hostname
- name: Configure iLO hostname
  module: ilo
  vars:
    ilo_ip: "10.0.1.101"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "set_hostname"
    hostname: "ilo-web01"
```

### Firmware Management

```yaml
# Update iLO firmware
- name: Update iLO firmware
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "update_firmware"
    firmware_url: "http://repo.example.com/firmware/ilo5_270.bin"
    timeout: 300

# Check update progress
- name: Check firmware update status
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "get_update_progress"
```

### Log Management

```yaml
# Get IML entries
- name: Retrieve server logs
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "get_iml_log"

# Clear IML
- name: Clear hardware logs
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "clear_iml_log"
```

### License Management

```yaml
# Install iLO Advanced license
- name: Activate iLO Advanced features
  module: ilo
  vars:
    ilo_ip: "10.0.1.100"
    ilo_user: "Administrator"
    ilo_password: "{{ vault.ilo_password }}"
    command: "install_license"
    license_key: "{{ vault.ilo_license_key }}"
```

## HP Redfish API Endpoints

The module uses the following Redfish endpoints:

### Systems
- `/redfish/v1/Systems/1` - System information and control
- `/redfish/v1/Systems/1/Actions/ComputerSystem.Reset` - Power control
- `/redfish/v1/Systems/1/LogServices/IML/Entries` - Integrated Management Log

### Managers (iLO)
- `/redfish/v1/Managers/1` - iLO information and settings
- `/redfish/v1/Managers/1/VirtualMedia/1` - Virtual floppy
- `/redfish/v1/Managers/1/VirtualMedia/2` - Virtual CD/DVD
- `/redfish/v1/Managers/1/EthernetInterfaces/1` - Network configuration
- `/redfish/v1/Managers/1/LogServices/IEL/Entries` - iLO Event Log
- `/redfish/v1/Managers/1/LicenseService` - License management
- `/redfish/v1/Managers/1/SecurityService` - Security settings

### Account Service
- `/redfish/v1/AccountService/Accounts` - User management

### Update Service
- `/redfish/v1/UpdateService/FirmwareInventory` - Firmware inventory
- `/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate` - Firmware updates

### HP OEM Extensions
- `/redfish/v1/Systems/1/Oem/Hp/` - HP-specific system settings
- `/redfish/v1/Managers/1/Oem/Hp/` - HP-specific iLO settings

## Troubleshooting

### SSL Certificate Errors

If you encounter SSL certificate validation errors:

```yaml
vars:
  validate_certs: false
```

Note: Only disable certificate validation in trusted environments.

### Authentication Failures

Ensure the iLO user has appropriate privileges:
- **Administrator**: Full access to all iLO functions
- **Operator**: Limited configuration access
- **ReadOnly**: View-only access

For most operations, Administrator privileges are required.

### Network Connectivity

Test iLO accessibility:

```bash
curl -k https://YOUR_ILO_IP/redfish/v1/
```

### Virtual Media Issues

1. Ensure the ISO URL is accessible from the iLO network
2. iLO must be able to reach the HTTP/HTTPS server
3. Check that virtual media is enabled in iLO settings
4. Verify that another virtual media session isn't active

### Firmware Update Failures

1. Ensure firmware file is compatible with your hardware
2. Check that iLO has network access to the firmware URL
3. Use appropriate timeout values (300+ seconds recommended)
4. Only one firmware update can run at a time

## HP-Specific Considerations

### iLO Versions
- **iLO 4**: Redfish support added in firmware 2.40+
- **iLO 5**: Full Redfish support from initial release
- **iLO 6**: Enhanced Redfish with additional features

### OEM Extensions
HP provides additional functionality through OEM extensions in the `Oem.Hp` or `Oem.Hpe` namespace. This module uses these extensions where necessary for HP-specific features.

### Default Credentials
Never use default credentials. Always change the default Administrator password after initial setup.

### Security Best Practices
1. Use unique, strong passwords for each iLO
2. Enable HTTPS and disable HTTP
3. Configure authentication failure logging
4. Regularly update iLO firmware for security patches
5. Use dedicated management network for iLO interfaces
6. Enable two-factor authentication when available (iLO 5+)

## Building the Module

```bash
cd modules/servers/ilo
make build
```

Or manually:

```bash
cd modules/servers/ilo/wasm
tinygo build -o ilo.wasm -target=wasi -no-debug main.go
```

## Return Values

All operations return JSON output with:
- `status`: "ok", "changed", or "failed"
- `message`: Human-readable description
- `facts`: Operation-specific data (varies by command)

For GET operations, the facts will contain the retrieved data from the iLO.

## License

Part of the OpenFroyo project.
