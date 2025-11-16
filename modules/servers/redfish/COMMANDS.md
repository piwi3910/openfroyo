# Redfish Module - Quick Command Reference

## Power Management Commands

| Command | Description | Additional Variables |
|---------|-------------|---------------------|
| `PowerOn` | Power on the server | - |
| `PowerOff` | Force power off (ungraceful) | - |
| `GracefulShutdown` | Graceful OS shutdown | - |
| `ForceRestart` | Force restart (ungraceful) | - |
| `PowerCycle` | Power cycle (off then on) | - |
| `Nmi` | Send Non-Maskable Interrupt | - |
| `GetPowerState` | Get current power state | - |

## System Inventory Commands

| Command | Description | Additional Variables |
|---------|-------------|---------------------|
| `GetSystemInfo` | Get system information (model, serial, BIOS, etc.) | - |
| `GetProcessorInfo` | Get CPU/processor details | - |
| `GetMemoryInfo` | Get memory module details | - |
| `GetStorageInfo` | Get storage controller/drive info | - |
| `GetNetworkInfo` | Get network interface info | - |

## BIOS Configuration Commands

| Command | Description | Additional Variables |
|---------|-------------|---------------------|
| `GetBiosAttributes` | Get all BIOS attributes | - |
| `SetBiosAttributes` | Set BIOS attributes | `attributes` (map) |
| `ResetBiosToDefaults` | Reset BIOS to factory defaults | - |

## Boot Configuration Commands

| Command | Description | Additional Variables |
|---------|-------------|---------------------|
| `GetBootConfiguration` | Get boot configuration | - |
| `SetBootSource` | Set boot source override (one-time) | `boot_source` (string) |
| `SetBootMode` | Set boot mode (UEFI/Legacy) | `boot_mode` (string) |

## Firmware Management Commands

| Command | Description | Additional Variables |
|---------|-------------|---------------------|
| `GetFirmwareInventory` | Get firmware inventory | - |
| `UpdateFirmware` | Update firmware from URI | `firmware_uri` (string) |

## Valid Boot Sources

For use with `SetBootSource` command:
- `None` - No override (use boot order)
- `Pxe` - Network boot
- `Hdd` - Hard disk
- `Cd` - CD/DVD
- `BiosSetup` - Enter BIOS setup
- `UefiShell` - UEFI Shell
- `Usb` - USB device
- `UefiTarget` - Specific UEFI target
- `SDCard` - SD Card
- `UefiHttp` - HTTP boot

## Valid Boot Modes

For use with `SetBootMode` command:
- `UEFI` - UEFI boot mode
- `Legacy` - Legacy BIOS boot mode

## Quick Examples

### Power Operations
```yaml
# Power on
vars:
  command: PowerOn

# Graceful shutdown
vars:
  command: GracefulShutdown
```

### Inventory
```yaml
# Get system info
vars:
  command: GetSystemInfo

# Get all hardware details
vars:
  command: GetProcessorInfo
```

### BIOS Configuration
```yaml
# Get BIOS settings
vars:
  command: GetBiosAttributes

# Set BIOS settings
vars:
  command: SetBiosAttributes
  attributes:
    BootMode: "Uefi"
    VirtualizationTechnology: "Enabled"
```

### Boot Configuration
```yaml
# PXE boot once
vars:
  command: SetBootSource
  boot_source: Pxe

# Set UEFI mode
vars:
  command: SetBootMode
  boot_mode: UEFI
```

### Firmware
```yaml
# Check firmware versions
vars:
  command: GetFirmwareInventory

# Update firmware
vars:
  command: UpdateFirmware
  firmware_uri: "http://repo.local/firmware.bin"
```

## Common Workflows

### Server Provisioning
1. `GetSystemInfo` - Verify server
2. `SetBiosAttributes` - Configure BIOS
3. `SetBootSource` - Set PXE boot
4. `PowerOn` - Start installation

### Inventory Collection
1. `GetSystemInfo` - Basic info
2. `GetProcessorInfo` - CPU details
3. `GetMemoryInfo` - RAM details
4. `GetStorageInfo` - Storage details
5. `GetFirmwareInventory` - Firmware versions

### Emergency Recovery
1. `GetPowerState` - Check current state
2. `ForceRestart` - Force reboot if hung
3. `SetBootSource` - Set recovery boot source
4. `PowerCycle` - Full power cycle

## HTTP Status Codes

| Code | Meaning | Common Cause |
|------|---------|--------------|
| 200 | Success | Operation completed |
| 202 | Accepted | Async operation started |
| 400 | Bad Request | Invalid parameters |
| 401 | Unauthorized | Wrong credentials |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Invalid endpoint/resource |
| 500 | Server Error | BMC internal error |
