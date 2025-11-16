# IPMI Module

The IPMI (Intelligent Platform Management Interface) module provides comprehensive out-of-band server management capabilities through the ipmitool command-line utility. This module supports remote power management, sensor monitoring, user management, and network configuration of server BMCs (Baseboard Management Controllers).

## Requirements

This module requires `ipmitool` to be installed on the target host where OpenFroyo executes the module.

**Installation:**
```bash
# Ubuntu/Debian
sudo apt-get install ipmitool

# RHEL/CentOS/Rocky
sudo yum install ipmitool

# Alpine
sudo apk add ipmitool
```

## IPMI Version Compatibility

- **IPMI 1.5**: Use `interface: "lan"`
- **IPMI 2.0**: Use `interface: "lanplus"` (default, recommended)

## Required Variables

- **bmc_host** (string): IP address or hostname of the BMC
- **username** (string): IPMI username
- **password** (string): IPMI password
- **command** (string): IPMI command to execute (see commands below)

## Optional Variables

- **interface** (string, default: "lanplus"): IPMI interface type ("lan" or "lanplus")
- **privilege** (string, default: "ADMINISTRATOR"): Privilege level (CALLBACK, USER, OPERATOR, ADMINISTRATOR)

## Supported Commands

### Power Management

#### power_on
Power on the server chassis.

**Example:**
```yaml
- name: Power on server
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_on"
```

#### power_off
Immediately power off the server (hard power off).

**Example:**
```yaml
- name: Power off server
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_off"
```

#### power_cycle
Power cycle the server (hard reset with power off/on).

**Example:**
```yaml
- name: Power cycle server
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_cycle"
```

#### power_reset
Hard reset the server without power off.

**Example:**
```yaml
- name: Reset server
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_reset"
```

#### power_soft
Initiate a soft shutdown (ACPI shutdown).

**Example:**
```yaml
- name: Soft shutdown server
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_soft"
```

#### power_status
Get current power status.

**Example:**
```yaml
- name: Get power status
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "power_status"
```

### Chassis Management

#### chassis_status
Get detailed chassis status information.

**Example:**
```yaml
- name: Get chassis status
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "chassis_status"
```

#### chassis_identify_on
Turn on chassis identify LED for a specified duration.

**Variables:**
- **duration** (int, default: 15): Duration in seconds

**Example:**
```yaml
- name: Turn on chassis LED for 30 seconds
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "chassis_identify_on"
    duration: 30
```

#### chassis_identify_off
Turn off chassis identify LED.

**Example:**
```yaml
- name: Turn off chassis LED
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "chassis_identify_off"
```

#### chassis_selftest
Run chassis self-test.

**Example:**
```yaml
- name: Run chassis self-test
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "chassis_selftest"
```

#### set_boot_device
Set next boot device.

**Variables:**
- **boot_device** (string, required): Boot device (pxe, disk, cdrom, bios, safe, diag, floppy, none)
- **persistent** (bool, default: false): Make boot device persistent across reboots

**Example:**
```yaml
- name: Set PXE boot once
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_boot_device"
    boot_device: "pxe"
    persistent: false
```

#### get_boot_device
Get current boot device settings.

**Example:**
```yaml
- name: Get boot device
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_boot_device"
```

### Sensor Monitoring

#### get_all_sensors
List all sensor readings.

**Example:**
```yaml
- name: Get all sensor readings
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_all_sensors"
```

#### get_sensor
Get specific sensor reading.

**Variables:**
- **sensor_name** (string, required): Name of the sensor

**Example:**
```yaml
- name: Get CPU temperature
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sensor"
    sensor_name: "CPU Temp"
```

#### get_sensor_thresholds
Get sensor thresholds.

**Variables:**
- **sensor_name** (string, required): Name of the sensor

**Example:**
```yaml
- name: Get CPU temperature thresholds
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sensor_thresholds"
    sensor_name: "CPU Temp"
```

### SEL (System Event Log)

#### get_sel
List all SEL entries.

**Example:**
```yaml
- name: Get system event log
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sel"
```

#### get_sel_info
Get SEL information (size, free space, etc.).

**Example:**
```yaml
- name: Get SEL info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sel_info"
```

#### clear_sel
Clear all SEL entries.

**Example:**
```yaml
- name: Clear system event log
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "clear_sel"
```

#### save_sel
Save SEL to file.

**Variables:**
- **filepath** (string, default: "/tmp/sel.log"): Path to save SEL

**Example:**
```yaml
- name: Save SEL to file
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "save_sel"
    filepath: "/var/log/server-sel.log"
```

#### get_sel_elist
Get extended SEL list.

**Example:**
```yaml
- name: Get extended SEL
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sel_elist"
```

#### get_sel_time
Get SEL time.

**Example:**
```yaml
- name: Get SEL time
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sel_time"
```

### FRU (Field Replaceable Unit)

#### get_fru
Get all FRU inventory data.

**Example:**
```yaml
- name: Get FRU inventory
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_fru"
```

#### get_fru_id
Get specific FRU by ID.

**Variables:**
- **fru_id** (int, default: 0): FRU ID

**Example:**
```yaml
- name: Get FRU ID 0
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_fru_id"
    fru_id: 0
```

### SOL (Serial Over LAN)

#### sol_activate
Activate Serial Over LAN session.

**Example:**
```yaml
- name: Activate SOL
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "sol_activate"
```

#### sol_deactivate
Deactivate Serial Over LAN session.

**Example:**
```yaml
- name: Deactivate SOL
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "sol_deactivate"
```

#### sol_info
Get SOL configuration information.

**Variables:**
- **channel** (int, default: 1): LAN channel number

**Example:**
```yaml
- name: Get SOL info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "sol_info"
    channel: 1
```

#### sol_set
Set SOL parameter.

**Variables:**
- **channel** (int, default: 1): LAN channel number
- **param** (string, required): Parameter name
- **value** (string, required): Parameter value

**Example:**
```yaml
- name: Set SOL enabled
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "sol_set"
    channel: 1
    param: "enabled"
    value: "true"
```

### User Management

#### list_users
List all IPMI users.

**Variables:**
- **channel** (int, default: 1): LAN channel number

**Example:**
```yaml
- name: List IPMI users
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "list_users"
    channel: 1
```

#### set_user_name
Set user name.

**Variables:**
- **user_id** (int, required): User ID
- **user_name** (string, required): New username

**Example:**
```yaml
- name: Set user name
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_user_name"
    user_id: 2
    user_name: "operator"
```

#### set_user_password
Set user password.

**Variables:**
- **user_id** (int, required): User ID
- **user_password** (string, required): New password

**Example:**
```yaml
- name: Set user password
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_user_password"
    user_id: 2
    user_password: "newpassword123"
```

#### enable_user
Enable a user account.

**Variables:**
- **user_id** (int, required): User ID

**Example:**
```yaml
- name: Enable user
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "enable_user"
    user_id: 2
```

#### disable_user
Disable a user account.

**Variables:**
- **user_id** (int, required): User ID

**Example:**
```yaml
- name: Disable user
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "disable_user"
    user_id: 2
```

#### set_user_privilege
Set user privilege level.

**Variables:**
- **user_id** (int, required): User ID
- **channel** (int, default: 1): LAN channel number
- **user_privilege** (string, required): Privilege level (1=CALLBACK, 2=USER, 3=OPERATOR, 4=ADMINISTRATOR)

**Example:**
```yaml
- name: Set user privilege to OPERATOR
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_user_privilege"
    user_id: 2
    channel: 1
    user_privilege: "3"
```

### LAN Configuration

#### get_lan_config
Get LAN configuration.

**Variables:**
- **channel** (int, default: 1): LAN channel number

**Example:**
```yaml
- name: Get LAN configuration
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_lan_config"
    channel: 1
```

#### set_lan_ip_static
Configure static IP address.

**Variables:**
- **channel** (int, default: 1): LAN channel number
- **ip_address** (string, required): IP address
- **subnet_mask** (string, required): Subnet mask
- **gateway** (string, required): Default gateway

**Example:**
```yaml
- name: Set static IP
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_lan_ip_static"
    channel: 1
    ip_address: "192.168.1.101"
    subnet_mask: "255.255.255.0"
    gateway: "192.168.1.1"
```

#### set_lan_ip_dhcp
Configure DHCP for IP address.

**Variables:**
- **channel** (int, default: 1): LAN channel number

**Example:**
```yaml
- name: Set DHCP
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_lan_ip_dhcp"
    channel: 1
```

#### set_lan_vlan
Configure VLAN ID.

**Variables:**
- **channel** (int, default: 1): LAN channel number
- **vlan_id** (int, required): VLAN ID (0 to disable)

**Example:**
```yaml
- name: Set VLAN 100
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_lan_vlan"
    channel: 1
    vlan_id: 100
```

#### set_lan_auth
Set LAN authentication type.

**Variables:**
- **channel** (int, default: 1): LAN channel number
- **auth_level** (string, required): Authentication level (CALLBACK, USER, OPERATOR, ADMIN)
- **auth_type** (string, required): Authentication type (NONE, MD2, MD5, PASSWORD, OEM)

**Example:**
```yaml
- name: Set LAN auth
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_lan_auth"
    channel: 1
    auth_level: "ADMIN"
    auth_type: "MD5"
```

### SDR (Sensor Data Record)

#### get_sdr_info
Get SDR repository information.

**Example:**
```yaml
- name: Get SDR info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sdr_info"
```

#### get_sdr_list
List all SDR records.

**Example:**
```yaml
- name: List SDR records
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sdr_list"
```

#### get_sdr_type
Get SDR by type.

**Variables:**
- **sdr_type** (string, required): SDR type

**Example:**
```yaml
- name: Get temperature SDRs
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sdr_type"
    sdr_type: "Temperature"
```

#### get_sdr_entity
Get SDR by entity ID.

**Variables:**
- **entity_id** (string, required): Entity ID

**Example:**
```yaml
- name: Get CPU SDRs
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_sdr_entity"
    entity_id: "3.1"
```

### Firmware and BIOS

#### get_bmc_info
Get BMC information (version, device ID, etc.).

**Example:**
```yaml
- name: Get BMC info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_bmc_info"
```

#### get_device_id
Get device ID.

**Example:**
```yaml
- name: Get device ID
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_device_id"
```

#### get_firmware_version
Get firmware version.

**Example:**
```yaml
- name: Get firmware version
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_firmware_version"
```

#### reset_bmc_cold
Cold reset the BMC (full reboot).

**Example:**
```yaml
- name: Cold reset BMC
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "reset_bmc_cold"
```

#### reset_bmc_warm
Warm reset the BMC (soft restart).

**Example:**
```yaml
- name: Warm reset BMC
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "reset_bmc_warm"
```

#### get_guid
Get system GUID.

**Example:**
```yaml
- name: Get system GUID
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_guid"
```

#### get_selftest
Get BMC self-test results.

**Example:**
```yaml
- name: Get BMC self-test
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_selftest"
```

### Watchdog Timer

#### get_watchdog
Get watchdog timer status.

**Example:**
```yaml
- name: Get watchdog status
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_watchdog"
```

#### reset_watchdog
Reset watchdog timer.

**Example:**
```yaml
- name: Reset watchdog
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "reset_watchdog"
```

#### set_watchdog
Set watchdog timer.

**Variables:**
- **watchdog_action** (string, default: "reset"): Action on timeout (reset, power_down, power_cycle)
- **watchdog_timeout** (int, default: 60): Timeout in seconds

**Example:**
```yaml
- name: Set watchdog timer
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "set_watchdog"
    watchdog_action: "reset"
    watchdog_timeout: 120
```

#### off_watchdog
Turn off watchdog timer.

**Example:**
```yaml
- name: Turn off watchdog
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "off_watchdog"
```

### Advanced Commands

#### get_pef
Get Platform Event Filtering configuration.

**Example:**
```yaml
- name: Get PEF config
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_pef"
```

#### raw
Execute raw IPMI command.

**Variables:**
- **raw_command** (string, required): Raw IPMI command bytes

**Example:**
```yaml
- name: Execute raw IPMI command
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "raw"
    raw_command: "0x06 0x01"
```

#### get_channel_info
Get channel information.

**Variables:**
- **channel** (int, default: 1): Channel number

**Example:**
```yaml
- name: Get channel info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_channel_info"
    channel: 1
```

#### get_channel_authcap
Get channel authentication capabilities.

**Variables:**
- **channel** (int, default: 1): Channel number

**Example:**
```yaml
- name: Get channel auth capabilities
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_channel_authcap"
    channel: 1
```

#### get_session_info
Get active session information.

**Example:**
```yaml
- name: Get session info
  module: servers/ipmi
  vars:
    bmc_host: "192.168.1.100"
    username: "admin"
    password: "password"
    command: "get_session_info"
```

## Common IPMI Channels

Different BMC implementations use different channel numbers:

- **Channel 1**: Usually LAN/LAN+ (most common)
- **Channel 8**: Some Dell systems
- **Channel 2**: Some HP systems
- **Channel 3**: Some Supermicro systems

Use `get_channel_info` to discover available channels.

## Troubleshooting

### Authentication Failures

1. Verify BMC is accessible:
   ```bash
   ping <bmc_host>
   ```

2. Test credentials manually:
   ```bash
   ipmitool -I lanplus -H <bmc_host> -U <username> -P <password> chassis status
   ```

3. Try different interfaces:
   - Use `interface: "lan"` for IPMI 1.5
   - Use `interface: "lanplus"` for IPMI 2.0

### Connection Timeouts

1. Check network connectivity
2. Verify BMC IP address is correct
3. Check firewall rules (IPMI uses UDP port 623)
4. Verify BMC is powered on and responding

### Privilege Level Issues

If you get "Insufficient privilege level" errors, try:
- Setting `privilege: "ADMINISTRATOR"`
- Checking user privilege level with `list_users`
- Verifying the user has the required privilege for the operation

### VLAN Issues

When BMC is on a VLAN:
1. Ensure you can reach the VLAN from the execution host
2. Use the correct VLAN-tagged IP address
3. Consider using a management interface in the same VLAN

## Security Considerations

1. **Password Security**: IPMI passwords are sent in plaintext with IPMI 1.5. Always use IPMI 2.0 (lanplus) when possible.

2. **Network Isolation**: BMCs should be on a separate management network, not on the production network.

3. **Credential Storage**: Never store IPMI credentials in plain text in stack files. Use OpenFroyo's variable encryption or external credential management.

4. **User Management**: Regularly audit IPMI users and disable unused accounts.

5. **Firmware Updates**: Keep BMC firmware updated to patch security vulnerabilities.

6. **Change Default Credentials**: Always change default IPMI passwords during initial setup.

## Module Version

**Version:** 1.0.0

**Supported IPMI Versions:**
- IPMI 1.5 (lan interface)
- IPMI 2.0 (lanplus interface)

**Dependencies:**
- ipmitool (any recent version)

## License

Part of the OpenFroyo project.
