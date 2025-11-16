# Redfish Module

Comprehensive WASM-based Redfish API module for server management operations in OpenFroyo.

## Overview

Redfish is the industry-standard REST API for server management, defined by the DMTF (Distributed Management Task Force). This module provides a complete interface to Redfish-enabled servers for power management, inventory collection, BIOS configuration, firmware updates, and boot configuration.

**Key Features:**
- Agentless operation via Redfish REST API
- Comprehensive power management (on, off, restart, graceful shutdown)
- Complete system inventory (CPU, memory, storage, network)
- BIOS configuration management
- Firmware inventory and updates
- Boot source and mode configuration
- SSL/TLS support with certificate validation
- Idempotent operations where applicable

## Module Location

```
modules/servers/redfish/
├── wasm/
│   └── main.go           # TinyGo implementation
├── module.ofy.yml        # Module definition
├── defaults.ofy.yml      # Default variable values
├── Makefile              # Build configuration
├── test.ofy              # Comprehensive test suite
└── README.md             # This file
```

## Requirements

- Redfish-enabled server (Dell iDRAC, HP iLO, Lenovo XClarity, Supermicro, etc.)
- Network access to the server's BMC/iLO/iDRAC
- Valid credentials with appropriate permissions
- `curl` available on the target host (for API calls)
- `jq` available on the target host (for JSON parsing)

## Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `baseuri` | string | Redfish API base URL (e.g., `https://192.168.1.100`) |
| `username` | string | Username for authentication |
| `password` | string | Password for authentication |
| `command` | string | Operation to perform (see Commands section) |

### Optional Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `timeout` | int | 30 | Request timeout in seconds |
| `validate_certs` | bool | true | Validate SSL certificates |
| `resource_id` | string | `System.Embedded.1` | System resource ID (vendor-specific) |
| `attributes` | map | {} | BIOS attributes for SetBiosAttributes |
| `boot_source` | string | "" | Boot source for SetBootSource |
| `boot_mode` | string | "" | Boot mode for SetBootMode (UEFI/Legacy) |
| `firmware_uri` | string | "" | Firmware image URI for UpdateFirmware |

## Supported Commands

### Power Management

#### PowerOn
Powers on the server if it's currently off.

```yaml
- name: Power on server
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: PowerOn
```

#### PowerOff
Forces an immediate power off (ungraceful shutdown).

```yaml
- name: Force power off
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: PowerOff
```

#### GracefulShutdown
Requests a graceful OS shutdown before powering off.

```yaml
- name: Graceful shutdown
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GracefulShutdown
```

#### ForceRestart
Forces an immediate restart (ungraceful).

```yaml
- name: Force restart
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: ForceRestart
```

#### PowerCycle
Performs a power cycle (off then on).

```yaml
- name: Power cycle server
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: PowerCycle
```

#### Nmi
Sends a Non-Maskable Interrupt (useful for crash dumps).

```yaml
- name: Send NMI
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: Nmi
```

#### GetPowerState
Retrieves the current power state of the server.

```yaml
- name: Check power state
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetPowerState
```

**Output:** Returns `On`, `Off`, `PoweringOn`, or `PoweringOff`

### System Inventory

#### GetSystemInfo
Retrieves comprehensive system information including manufacturer, model, serial number, BIOS version, processor count, memory, and health status.

```yaml
- name: Get system information
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetSystemInfo
```

**Output Example:**
```json
{
  "Manufacturer": "Dell Inc.",
  "Model": "PowerEdge R640",
  "SerialNumber": "ABC1234",
  "BiosVersion": "2.15.0",
  "ProcessorSummary": {
    "Count": 2,
    "Model": "Intel(R) Xeon(R) Gold 6140 CPU @ 2.30GHz"
  },
  "MemorySummary": {
    "TotalSystemMemoryGiB": 384
  },
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  }
}
```

#### GetProcessorInfo
Retrieves detailed CPU/processor information.

```yaml
- name: Get processor info
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetProcessorInfo
```

#### GetMemoryInfo
Retrieves detailed memory module information.

```yaml
- name: Get memory info
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetMemoryInfo
```

#### GetStorageInfo
Retrieves storage controller and drive information.

```yaml
- name: Get storage info
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetStorageInfo
```

#### GetNetworkInfo
Retrieves network interface information.

```yaml
- name: Get network info
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetNetworkInfo
```

### BIOS Configuration

#### GetBiosAttributes
Retrieves all BIOS configuration attributes.

```yaml
- name: Get BIOS attributes
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetBiosAttributes
```

**Output Example:**
```json
{
  "Attributes": {
    "BootMode": "Uefi",
    "QuietBoot": "Enabled",
    "Sriov": "Enabled",
    "VirtualizationTechnology": "Enabled"
  }
}
```

#### SetBiosAttributes
Sets one or more BIOS attributes. **Note:** Usually requires a reboot to take effect.

```yaml
- name: Configure BIOS settings
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: SetBiosAttributes
    attributes:
      BootMode: "Uefi"
      QuietBoot: "Enabled"
      VirtualizationTechnology: "Enabled"
```

**Important Notes:**
- Attribute names are vendor-specific
- Most changes require a system reboot to take effect
- Check your server's Redfish documentation for valid attribute names and values

#### ResetBiosToDefaults
Resets BIOS configuration to factory defaults.

```yaml
- name: Reset BIOS to defaults
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: ResetBiosToDefaults
```

### Boot Configuration

#### GetBootConfiguration
Retrieves current boot configuration including boot mode, boot order, and boot source override settings.

```yaml
- name: Get boot configuration
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetBootConfiguration
```

#### SetBootSource
Sets the boot source override for the next boot. The override is automatically disabled after one boot.

```yaml
- name: Boot from PXE once
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: SetBootSource
    boot_source: Pxe
```

**Valid boot sources:**
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

#### SetBootMode
Sets the boot mode (UEFI or Legacy BIOS).

```yaml
- name: Set UEFI boot mode
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: SetBootMode
    boot_mode: UEFI
```

**Valid boot modes:**
- `UEFI` - UEFI boot mode
- `Legacy` - Legacy BIOS boot mode

### Firmware Management

#### GetFirmwareInventory
Retrieves the current firmware inventory including BIOS, BMC, NIC, and other firmware versions.

```yaml
- name: Get firmware inventory
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: GetFirmwareInventory
```

**Output Example:**
```json
{
  "Members": [
    {
      "Name": "BIOS",
      "Version": "2.15.0",
      "Updateable": true
    },
    {
      "Name": "iDRAC",
      "Version": "5.10.00.00",
      "Updateable": true
    }
  ]
}
```

#### UpdateFirmware
Initiates a firmware update using a URI to the firmware image.

```yaml
- name: Update BIOS firmware
  module: servers/redfish
  hosts:
    - server-01
  vars:
    baseuri: "https://192.168.1.100"
    username: "admin"
    password: "password"
    command: UpdateFirmware
    firmware_uri: "http://firmware-repo.local/bios-2.16.0.bin"
```

**Important Notes:**
- The firmware URI must be accessible from the BMC/iLO
- Firmware updates may require a reboot
- Check update status separately (implementation-specific)
- Always verify firmware compatibility before updating

## Common Redfish Endpoints

This module uses standard Redfish endpoints. Here are the most common ones:

| Endpoint | Purpose |
|----------|---------|
| `/redfish/v1/` | Service root |
| `/redfish/v1/Systems/{id}` | System information |
| `/redfish/v1/Systems/{id}/Actions/ComputerSystem.Reset` | Power operations |
| `/redfish/v1/Systems/{id}/Bios` | BIOS attributes |
| `/redfish/v1/Systems/{id}/Bios/Settings` | BIOS settings (write) |
| `/redfish/v1/Systems/{id}/Processors` | Processor information |
| `/redfish/v1/Systems/{id}/Memory` | Memory information |
| `/redfish/v1/Systems/{id}/Storage` | Storage information |
| `/redfish/v1/Systems/{id}/EthernetInterfaces` | Network interfaces |
| `/redfish/v1/UpdateService/FirmwareInventory` | Firmware inventory |
| `/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate` | Firmware update |

## Resource IDs by Vendor

Different vendors use different resource IDs for their systems:

| Vendor | Default Resource ID | Notes |
|--------|---------------------|-------|
| Dell iDRAC | `System.Embedded.1` | Default for this module |
| HP iLO | `1` | Simple numeric ID |
| Lenovo XClarity | `1` | Simple numeric ID |
| Supermicro | `1` | Simple numeric ID |
| Cisco UCS | Check documentation | Varies by model |

**Finding your resource ID:**
```bash
curl -k -u admin:password https://your-server/redfish/v1/Systems/ | jq
```

Override the default with the `resource_id` variable:
```yaml
vars:
  resource_id: "1"  # For HP iLO
```

## SSL/TLS Certificate Validation

By default, this module validates SSL certificates. For self-signed certificates (common in BMC environments), disable validation:

```yaml
vars:
  validate_certs: false
```

**Production Recommendation:** Use proper certificates and keep validation enabled for security.

## Authentication

This module uses HTTP Basic Authentication, which is secure when used over HTTPS. Credentials are passed to `curl` using the `-u` flag.

**Security Best Practices:**
- Always use HTTPS (not HTTP) for the baseuri
- Use strong passwords for BMC accounts
- Rotate credentials regularly
- Use dedicated automation accounts with minimal required permissions
- Consider network segmentation for management interfaces

## Error Handling

The module handles various error conditions:

### HTTP Status Codes
- `200` - Success
- `202` - Accepted (async operation started)
- `400` - Bad request (invalid parameters)
- `401` - Unauthorized (invalid credentials)
- `403` - Forbidden (insufficient permissions)
- `404` - Not found (invalid endpoint or resource)
- `500` - Internal server error

### Common Errors

**Connection Refused:**
```
Failed to connect to <baseuri>
```
- Check network connectivity
- Verify baseuri is correct
- Ensure BMC/iLO is powered on and configured

**Authentication Failed:**
```
HTTP_STATUS:401
```
- Verify username and password
- Check account permissions
- Ensure account is not locked

**Invalid Resource:**
```
HTTP_STATUS:404
```
- Check resource_id matches your vendor
- Verify the endpoint is supported by your BMC firmware version

**SSL Certificate Error:**
```
SSL certificate problem
```
- Set `validate_certs: false` for self-signed certificates
- Or install proper certificates on the BMC

## Idempotency

The module provides idempotent operations where possible:

- **Power operations:** Check current state before changing (future enhancement)
- **BIOS settings:** Only changed attributes are updated
- **Boot configuration:** Idempotent when setting to the same value

## Performance Considerations

- **Timeout:** Default 30 seconds, adjust for slow networks or operations
- **Parallel execution:** Safe to run against multiple servers in parallel
- **API rate limits:** Some BMCs have rate limits (typically not an issue)
- **Firmware updates:** Can take 10-30 minutes, use appropriate timeouts

## Vendor-Specific Notes

### Dell iDRAC
- Resource ID: `System.Embedded.1`
- Excellent Redfish compliance
- BIOS changes require `CreateRebootJob` (not included in MVP)
- iDRAC 9 recommended for full Redfish support

### HP iLO
- Resource ID: `1`
- iLO 4 and newer support Redfish
- Some features require iLO Advanced license
- BIOS changes typically auto-commit on next boot

### Lenovo XClarity
- Resource ID: `1`
- Good Redfish support in recent firmware
- Some advanced features may require XCC enterprise license

### Supermicro
- Resource ID: `1`
- Redfish support varies by BMC firmware version
- Update to latest firmware for best compatibility

## Troubleshooting

### Debug Mode

To see the actual curl commands being executed, modify the shell script to remove `-s` (silent) flag and add `-v` (verbose):

```yaml
# This would require modifying the WASM module - for debugging
```

### Testing Connectivity

Use curl directly to test:
```bash
# Test connection
curl -k -u admin:password https://192.168.1.100/redfish/v1/

# Get system info
curl -k -u admin:password https://192.168.1.100/redfish/v1/Systems/System.Embedded.1
```

### Common Issues

**Problem:** Command times out
- **Solution:** Increase timeout value, check network latency

**Problem:** Invalid boot source
- **Solution:** Check `GetBootConfiguration` for supported boot sources

**Problem:** BIOS changes don't apply
- **Solution:** Reboot the server (most BIOS changes require a reboot)

**Problem:** Firmware update fails
- **Solution:** Ensure firmware URI is accessible from BMC network, verify compatibility

## Build Instructions

Build the WASM module:

```bash
cd modules/servers/redfish
make build
```

This compiles `wasm/main.go` to `wasm/redfish.wasm` using TinyGo with WASI target.

Clean build artifacts:
```bash
make clean
```

## Testing

Run the comprehensive test suite:

```bash
# Edit test.ofy to set your server details
froyo apply modules/servers/redfish/test.ofy
```

**Important:** Update the connection details in `test.ofy` before running tests. The test file includes examples of all supported operations.

## Examples

### Complete Server Provisioning Workflow

```yaml
name: Provision New Server
description: Configure and provision a new server via Redfish

inventory:
  - inventory/hosts.yml

defaults:
  redfish_baseuri: "https://{{ var.server_ipmi }}"
  redfish_username: "admin"
  redfish_password: "{{ var.ipmi_password }}"
  redfish_validate_certs: false

run:
  # Step 1: Verify server is accessible
  - name: Get system information
    module: servers/redfish
    hosts:
      - new-server
    vars:
      baseuri: "{{ var.redfish_baseuri }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: GetSystemInfo

  # Step 2: Configure BIOS for virtualization
  - name: Enable virtualization features
    module: servers/redfish
    hosts:
      - new-server
    vars:
      baseuri: "{{ var.redfish_baseuri }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: SetBiosAttributes
      attributes:
        VirtualizationTechnology: "Enabled"
        Sriov: "Enabled"
        BootMode: "Uefi"

  # Step 3: Set boot to PXE for OS installation
  - name: Set PXE boot for installation
    module: servers/redfish
    hosts:
      - new-server
    vars:
      baseuri: "{{ var.redfish_baseuri }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: SetBootSource
      boot_source: Pxe

  # Step 4: Power on server
  - name: Power on for OS installation
    module: servers/redfish
    hosts:
      - new-server
    vars:
      baseuri: "{{ var.redfish_baseuri }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: PowerOn
```

### Mass Inventory Collection

```yaml
name: Collect Server Inventory
description: Collect hardware inventory from all servers

inventory:
  - inventory/datacenters.yml

defaults:
  redfish_username: "readonly"
  redfish_password: "{{ var.readonly_password }}"
  redfish_validate_certs: true

run:
  - name: Get system information
    module: servers/redfish
    hosts:
      - @group:all_servers
    vars:
      baseuri: "https://{{ var.ipmi_address }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: GetSystemInfo

  - name: Get firmware inventory
    module: servers/redfish
    hosts:
      - @group:all_servers
    vars:
      baseuri: "https://{{ var.ipmi_address }}"
      username: "{{ var.redfish_username }}"
      password: "{{ var.redfish_password }}"
      command: GetFirmwareInventory
```

### Emergency Power Cycle

```yaml
name: Emergency Power Cycle
description: Force power cycle unresponsive servers

inventory:
  - inventory/hosts.yml

run:
  - name: Force power cycle
    module: servers/redfish
    hosts:
      - problem-server-01
    vars:
      baseuri: "https://192.168.1.100"
      username: "admin"
      password: "{{ var.emergency_password }}"
      command: PowerCycle
      validate_certs: false
```

## Reference Documentation

- **Redfish Specification:** https://www.dmtf.org/standards/redfish
- **Redfish API Reference:** https://redfish.dmtf.org/schemas/
- **Dell iDRAC Redfish:** https://www.dell.com/support/kbdoc/000177506
- **HP iLO Redfish:** https://www.hpe.com/us/en/servers/restful-api.html
- **Lenovo XClarity:** https://sysmgt.lenovofiles.com/help/topic/com.lenovo.systems.management.xcc.doc/dxcc_redfish_api.html

## Future Enhancements

Potential additions for future versions:
- Session-based authentication (more efficient for multiple calls)
- Event subscription support
- Virtual media management (ISO mounting)
- Job queue management (Dell-specific)
- Task monitoring for async operations
- Enhanced idempotency with state checking
- Support for vendor-specific extensions

## Contributing

When adding new commands or features:
1. Add the command handler in `wasm/main.go`
2. Update `module.ofy.yml` with new variables
3. Add examples to `test.ofy`
4. Document in this README
5. Test against multiple vendor implementations

## License

Part of the OpenFroyo project. See repository LICENSE file.

## Support

For issues or questions:
- Check troubleshooting section above
- Review vendor-specific Redfish documentation
- Ensure firmware is up to date
- Verify network connectivity and credentials

## Version History

- **1.0.0** - Initial release with comprehensive Redfish support
  - Power management operations
  - System inventory collection
  - BIOS configuration management
  - Firmware management
  - Boot configuration
