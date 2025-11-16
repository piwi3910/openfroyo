# Dell iDRAC Module

Comprehensive Dell iDRAC management via Redfish API for OpenFroyo.

## Description

This module provides complete Dell iDRAC (integrated Dell Remote Access Controller) management capabilities using the Redfish API. It supports power management, virtual media, system configuration, user management, network configuration, lifecycle controller operations, and RAID management.

## Variables

### Required Variables

- `baseuri` (string): iDRAC IP address or hostname
- `username` (string): iDRAC username
- `password` (string): iDRAC password
- `command` (string): Operation to perform (see commands below)

### Optional Variables

- `timeout` (int): Request timeout in seconds (default: 30)
- `validate_certs` (bool): Validate SSL certificates (default: true)

### Command-Specific Variables

Different commands require additional variables as documented below.

## Supported Commands

### Power Management

**get_power_state**
- Get current power state of the system
- Returns: PowerState (On, Off, PoweringOn, PoweringOff)

**power_on**
- Power on the system
- Uses: ForceOn reset type

**power_off**
- Force power off the system
- Uses: ForceOff reset type

**graceful_shutdown**
- Gracefully shutdown the system
- Uses: GracefulShutdown reset type

**power_cycle**
- Force restart the system
- Uses: ForceRestart reset type

**force_restart**
- Force restart the system
- Uses: ForceRestart reset type

### Virtual Media

**insert_virtual_media**
- Insert virtual CD/DVD/Floppy media
- Required: `media_type` (CD, DVD, Floppy), `image_url` (HTTP/HTTPS/NFS URL)

**eject_virtual_media**
- Eject virtual media
- Required: `media_type` (CD, DVD, Floppy)

### System Configuration

**reset_idrac**
- Reset the iDRAC controller
- Uses: GracefulRestart

**get_system_inventory**
- Get detailed system inventory
- Returns: Manufacturer, Model, Serial Number, BIOS Version, Power State, CPU/Memory summary

**get_sel**
- Get System Event Log entries
- Returns: SEL entries with timestamps and messages

**clear_sel**
- Clear the System Event Log
- Removes all SEL entries

### User Management

**create_user**
- Create a new iDRAC user
- Required: `new_username`, `new_password`
- Optional: `privilege` (Administrator, Operator, ReadOnly; default: Operator)

**delete_user**
- Delete an iDRAC user
- Required: `target_username`

**change_password**
- Change password for an iDRAC user
- Required: `target_username`, `new_password`

### Network Configuration

**configure_network**
- Configure iDRAC network settings
- For static IP: `ip_address`, `netmask`, `gateway`
- For DHCP: `dhcp_enabled: true`

**configure_dns**
- Configure DNS settings
- For static DNS: `dns1`, `dns2`
- For DHCP DNS: `dns_from_dhcp: true`

**configure_ntp**
- Configure NTP settings
- Required: `ntp_enabled` (bool)
- Optional: `ntp1`, `ntp2` (NTP server addresses)

### Lifecycle Controller

**get_firmware_inventory**
- Get firmware inventory for all components
- Returns: Firmware versions for iDRAC, BIOS, NICs, storage, etc.

**update_firmware**
- Update firmware from a URI
- Required: `image_uri` (HTTP/HTTPS/NFS URL to firmware image)

**get_job_queue**
- Get current job queue
- Returns: All pending and active jobs

**delete_job**
- Delete a job from the queue
- Required: `job_id`

### RAID Configuration

**get_storage_controllers**
- Get list of storage controllers
- Returns: Controller IDs and details

**get_virtual_disks**
- Get virtual disks for a controller
- Optional: `controller_id` (default: RAID.Integrated.1-1)

**get_physical_disks**
- Get physical disks for a controller
- Optional: `controller_id` (default: RAID.Integrated.1-1)

**create_virtual_disk**
- Create a new virtual disk
- Required: `controller_id`, `raid_level` (0, 1, 5, 6, 10, 50, 60), `disks` (comma-separated disk IDs)
- Optional: `vd_name` (virtual disk name)

**delete_virtual_disk**
- Delete a virtual disk
- Required: `vd_id` (virtual disk ID)

## Usage Examples

### Power Management

```yaml
- name: Check power state
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "get_power_state"

- name: Power on server
  module: idrac
  vars:
    baseuri: "idrac-server01.example.com"
    username: "admin"
    password: "password"
    command: "power_on"

- name: Graceful shutdown
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "graceful_shutdown"
```

### Virtual Media

```yaml
- name: Insert Ubuntu ISO
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "insert_virtual_media"
    media_type: "CD"
    image_url: "http://mirror.example.com/ubuntu-22.04.iso"

- name: Eject virtual CD
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "eject_virtual_media"
    media_type: "CD"
```

### System Configuration

```yaml
- name: Get system inventory
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "get_system_inventory"

- name: Clear System Event Log
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "clear_sel"
```

### User Management

```yaml
- name: Create operator user
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "create_user"
    new_username: "operator1"
    new_password: "SecurePass123"
    privilege: "Operator"

- name: Change user password
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "change_password"
    target_username: "operator1"
    new_password: "NewSecurePass456"
```

### Network Configuration

```yaml
- name: Configure static IP
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "configure_network"
    ip_address: "192.168.1.150"
    netmask: "255.255.255.0"
    gateway: "192.168.1.1"

- name: Enable DHCP
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "configure_network"
    dhcp_enabled: true

- name: Configure NTP
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "configure_ntp"
    ntp_enabled: true
    ntp1: "time.nist.gov"
    ntp2: "time.google.com"
```

### Firmware Management

```yaml
- name: Get firmware inventory
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "get_firmware_inventory"

- name: Update iDRAC firmware
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "update_firmware"
    image_uri: "http://downloads.dell.com/idrac_fw.exe"
```

### RAID Management

```yaml
- name: List storage controllers
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "get_storage_controllers"

- name: Get physical disks
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "get_physical_disks"
    controller_id: "RAID.Integrated.1-1"

- name: Create RAID 1 virtual disk
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "create_virtual_disk"
    controller_id: "RAID.Integrated.1-1"
    raid_level: "1"
    disks: "Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1,Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1"
    vd_name: "OS-Volume"

- name: Delete virtual disk
  module: idrac
  vars:
    baseuri: "192.168.1.100"
    username: "root"
    password: "calvin"
    command: "delete_virtual_disk"
    vd_id: "Disk.Virtual.0:RAID.Integrated.1-1"
```

## Implementation Details

- Uses Dell Redfish API endpoints
- All operations return JSON responses
- HTTP status codes are checked for success/failure
- Supports Dell OEM extensions for iDRAC-specific features
- SSL certificate validation can be disabled for self-signed certificates
- Configurable timeouts for long-running operations

## Notes

- Default iDRAC credentials are typically `root`/`calvin`
- Many operations require Administrator privileges
- Some operations (firmware update, RAID changes) create jobs that run asynchronously
- Job status can be monitored using `get_job_queue` command
- Virtual disk creation/deletion may require a system reboot
- Network configuration changes may cause temporary loss of connectivity

## Return Values

All commands return:
- `status`: "ok", "changed", or "failed"
- `message`: Human-readable description of the operation
- `facts`: Command output and discovered information

## Building

```bash
cd modules/servers/idrac
make build
```

## Requirements

- Dell PowerEdge server with iDRAC 7, 8, or 9
- iDRAC Enterprise license (for some advanced features)
- Network connectivity to iDRAC interface
- Valid iDRAC credentials

## References

- [Dell iDRAC Redfish API Documentation](https://www.dell.com/support/article/en-us/sln310367/idrac-redfish-api-overview)
- [DMTF Redfish Specification](https://www.dmtf.org/standards/redfish)
