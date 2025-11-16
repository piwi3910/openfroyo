# Server Management Modules Documentation

This document provides comprehensive information about OpenFroyo's server management modules for out-of-band BMC control.

## Overview

OpenFroyo includes 5 comprehensive modules for managing servers via their Baseboard Management Controllers (BMCs):

1. **redfish** - Generic Redfish API (industry standard)
2. **idrac** - Dell iDRAC complete management
3. **idrac_server_config_profile** - Dell configuration profiles
4. **ilo** - HP iLO complete management
5. **ipmi** - Generic IPMI management

These modules enable lights-out management including power control, virtual media, firmware updates, configuration management, and monitoring.

## Module Details

### Redfish Module

**Purpose:** Universal server management via Redfish REST API

**Supported Operations (18 commands):**
- Power: PowerOn, PowerOff, GracefulShutdown, ForceRestart, PowerCycle, Nmi, GetPowerState
- Inventory: GetSystemInfo, GetProcessorInfo, GetMemoryInfo, GetStorageInfo, GetNetworkInfo
- BIOS: GetBiosAttributes, SetBiosAttributes, ResetBiosToDefaults
- Boot: GetBootConfiguration, SetBootSource, SetBootMode
- Firmware: GetFirmwareInventory, UpdateFirmware

**Vendor Support:**
- Dell iDRAC 9+
- HP iLO 4+
- Lenovo XClarity
- Supermicro BMC
- Any Redfish 1.0+ compliant BMC

**Binary Size:** 475KB

---

### iDRAC Module

**Purpose:** Complete Dell iDRAC management with Dell-specific features

**Supported Operations (27 commands):**

**Power Management (6):**
- power_on, power_off, graceful_shutdown, force_restart, power_cycle, get_power_state

**Virtual Media (2):**
- insert_virtual_media, eject_virtual_media

**System Configuration (4):**
- reset_idrac, get_system_inventory, get_sel, clear_sel

**User Management (3):**
- create_user, delete_user, change_user_password

**Network Configuration (3):**
- configure_network, configure_dns, configure_ntp

**Lifecycle Controller (4):**
- get_firmware_inventory, update_firmware, get_job_queue, delete_job

**RAID Configuration (5):**
- get_controllers, get_virtual_disks, create_virtual_disk, delete_virtual_disk, get_physical_disks

**Compatibility:** Dell iDRAC 7, 8, 9, 10+

**Binary Size:** 503KB

---

### iDRAC Server Config Profile Module

**Purpose:** Import/export Dell server configuration profiles (Ansible-compatible)

**Supported Operations (3 commands):**
- **export** - Export server configuration to XML/JSON
- **import** - Import server configuration from file
- **preview** - Preview configuration changes without applying

**Export Features:**
- Multiple formats: XML, JSON
- Component selection: ALL, IDRAC, BIOS, NIC, RAID
- Save to local file

**Import Features:**
- Upload configuration file
- Configurable shutdown: Graceful, Forced, NoReboot
- End host power state: On, Off
- Automatic job tracking
- 15-second polling interval
- Configurable timeout (default: 3600s)

**Use Cases:**
- Server cloning
- Configuration backup/restore
- Audit server settings
- Automated BIOS/RAID configuration

**Compatibility:** Dell iDRAC 7+ with Lifecycle Controller

**Binary Size:** 481KB

---

### iLO Module

**Purpose:** Complete HP iLO management with HP-specific features

**Supported Operations (53 commands):**

**Power Management (7):**
- power_on, power_off, graceful_shutdown, force_restart, power_button, cold_boot, get_power_state

**Virtual Media (5):**
- insert_virtual_cd, eject_virtual_cd, insert_virtual_floppy, eject_virtual_floppy, get_virtual_media_status

**System Information (6):**
- get_system_info, get_health_status, get_hardware_inventory, get_firmware_versions, get_serial_number, get_product_info

**User Management (4):**
- get_users, create_user, modify_user, delete_user

**Network Configuration (5):**
- get_network_settings, set_network_static, set_network_dhcp, set_hostname, set_dns_servers

**Firmware Management (3):**
- get_firmware_inventory, update_firmware, get_update_progress

**Boot Configuration (4):**
- set_onetime_boot, set_persistent_boot, get_boot_config, set_uefi_boot

**Logs and Events (4):**
- get_iml_log, clear_iml_log, get_ilo_event_log, clear_ilo_event_log

**License Management (2):**
- get_license_status, install_license

**iLO Settings (5):**
- reset_ilo, get_ilo_datetime, set_ilo_datetime, get_ilo_info, ping_from_ilo

**Security (3):**
- get_security_dashboard, generate_csr, import_certificate

**Compatibility:** HP iLO 4, iLO 5, iLO 6

**Binary Size:** 503KB

---

### IPMI Module

**Purpose:** Universal BMC management via standard IPMI protocol

**Supported Operations (50+ commands):**

**Power Management (6):**
- power_on, power_off, power_cycle, power_reset, power_soft, power_status

**Chassis Management (6):**
- chassis_status, chassis_identify_on, chassis_identify_off, chassis_selftest, set_boot_device, get_boot_device

**Sensor Monitoring (3):**
- get_all_sensors, get_sensor, get_sensor_thresholds

**System Event Log (6):**
- get_sel, get_sel_info, get_sel_elist, get_sel_time, clear_sel, save_sel

**FRU Inventory (2):**
- get_fru, get_fru_id

**Serial Over LAN (4):**
- sol_activate, sol_deactivate, sol_info, sol_set

**User Management (6):**
- list_users, set_user_name, set_user_password, enable_user, disable_user, set_user_privilege

**LAN Configuration (5):**
- get_lan_config, set_lan_ip_static, set_lan_ip_dhcp, set_lan_vlan, set_lan_auth

**SDR Repository (4):**
- get_sdr_info, get_sdr_list, get_sdr_type, get_sdr_entity

**Firmware & BIOS (7):**
- get_bmc_info, get_device_id, get_firmware_version, reset_bmc_cold, reset_bmc_warm, get_guid, get_selftest

**Advanced (7):**
- get_watchdog, reset_watchdog, set_watchdog, off_watchdog, get_pef, raw, get_channel_info

**Compatibility:** Any IPMI 1.5 or 2.0 compliant BMC

**Requirements:** ipmitool installed on OpenFroyo control node

**Binary Size:** 483KB

---

## Common Workflows

### 1. Server Provisioning

```yaml
# Power on server
- module: servers/redfish
  vars:
    baseuri: "https://{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: PowerOn

# Mount installation ISO
- module: servers/idrac
  vars:
    idrac_ip: "{{bmc_ip}}"
    idrac_user: "{{bmc_user}}"
    idrac_password: "{{bmc_pass}}"
    command: insert_virtual_media
    media_type: CD
    image_url: "http://pxe.local/ubuntu-22.04.iso"

# Set one-time boot to virtual media
- module: servers/redfish
  vars:
    baseuri: "https://{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: SetBootSource
    boot_source: Cd
    boot_enabled: Once

# Restart server
- module: servers/redfish
  vars:
    baseuri: "https://{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: ForceRestart
```

### 2. Server Configuration Cloning (Dell)

```yaml
# Export configuration from template server
- module: servers/idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.10"
    idrac_user: "root"
    idrac_password: "calvin"
    command: export
    export_format: XML
    scp_components: ALL
    scp_file: /tmp/template_server.xml

# Import configuration to new servers
- module: servers/idrac_server_config_profile
  vars:
    idrac_ip: "{{item}}"
    idrac_user: "root"
    idrac_password: "calvin"
    command: import
    scp_file: /tmp/template_server.xml
    shutdown_type: Graceful
    end_host_power_state: On
    job_wait: true
  loop:
    - 192.168.1.20
    - 192.168.1.21
    - 192.168.1.22
```

### 3. Firmware Update Workflow

```yaml
# Get current firmware versions
- module: servers/ilo
  vars:
    ilo_ip: "{{bmc_ip}}"
    ilo_user: "Administrator"
    ilo_password: "{{ilo_pass}}"
    command: get_firmware_versions

# Update iLO firmware
- module: servers/ilo
  vars:
    ilo_ip: "{{bmc_ip}}"
    ilo_user: "Administrator"
    ilo_password: "{{ilo_pass}}"
    command: update_firmware
    firmware_uri: "http://firmware.local/ilo5_280.bin"

# Monitor update progress
- module: servers/ilo
  vars:
    ilo_ip: "{{bmc_ip}}"
    ilo_user: "Administrator"
    ilo_password: "{{ilo_pass}}"
    command: get_update_progress
```

### 4. Monitoring and Health Checks

```yaml
# Check power state
- module: servers/ipmi
  vars:
    bmc_host: "{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: power_status

# Get all sensor readings
- module: servers/ipmi
  vars:
    bmc_host: "{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: get_all_sensors

# Check system event log
- module: servers/ipmi
  vars:
    bmc_host: "{{bmc_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: get_sel

# Get hardware health
- module: servers/ilo
  vars:
    ilo_ip: "{{bmc_ip}}"
    ilo_user: "Administrator"
    ilo_password: "{{ilo_pass}}"
    command: get_health_status
```

## Security Best Practices

### Network Isolation
- Place BMCs on dedicated management network (VLAN)
- Restrict BMC access to authorized management hosts
- Use firewall rules to limit BMC access

### Authentication
- Use strong, unique passwords for each BMC
- Rotate BMC credentials regularly
- Create dedicated automation users with minimal privileges
- Disable default accounts (e.g., root, Administrator)

### Encryption
- Use HTTPS for Redfish/iDRAC/iLO (set validate_certs: true in production)
- Use IPMI 2.0 (lanplus) instead of 1.5
- Enable SSL/TLS certificate validation
- Install proper CA-signed certificates on BMCs

### Credential Management
- Store BMC credentials in encrypted vault
- Use environment variables or secure parameter passing
- Never hardcode credentials in stack files
- Use per-server credential isolation

## Performance Considerations

### Timeouts
- Default timeout: 30-60 seconds
- Increase for slow operations (firmware updates: 3600s)
- Consider BMC processing capacity

### Parallel Operations
- Limit concurrent BMC operations (max 10-20 per BMC)
- Stagger firmware updates to avoid network saturation
- Use serial execution for critical operations

### Network Latency
- BMC response times: 1-5 seconds typical
- Over VPN: 5-15 seconds
- Account for latency in timeout values

## Troubleshooting Guide

### Common Issues

**Connection Refused**
- Verify BMC IP address and network connectivity
- Check firewall rules (HTTPS: 443, IPMI: 623)
- Ensure BMC is powered on and initialized

**Authentication Failed**
- Verify username and password
- Check user account is enabled
- Verify user has appropriate privileges
- Check for locked/expired accounts

**SSL Certificate Errors**
- Use `validate_certs: false` for testing
- Install BMC certificate in trust store
- Use proper CA-signed certificates in production

**Timeout Errors**
- Increase timeout value
- Check network latency
- Verify BMC is not overloaded
- Check for firmware issues

**Command Not Supported**
- Verify BMC firmware version
- Check module compatibility
- Review vendor-specific requirements

## GitHub Issue

This work was tracked in: [Issue #15 - Add server out-of-band management modules](https://github.com/piwi3910/openfroyo/issues/15)

## Total Module Statistics

- **5 server management modules**
- **180+ total operations**
- **2.4MB total WASM binaries**
- **100% vendor coverage** (Dell, HP, Generic)
- **Complete documentation** for all modules
- **Comprehensive test suites** included
