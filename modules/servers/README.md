# OpenFroyo Server Management Modules

This directory contains modules for out-of-band server management via BMC (Baseboard Management Controller) interfaces.

## Available Modules (5 Total)

### 1. redfish (Generic Redfish API)
**Purpose:** Industry-standard REST API for server management

**Key Features:**
- Multi-vendor support (Dell, HP, Lenovo, Supermicro, etc.)
- Power management
- System inventory
- BIOS configuration
- Firmware updates
- Boot configuration

**Compatibility:** Any Redfish 1.0+ compliant BMC

**Size:** 475KB

---

### 2. idrac (Dell iDRAC Management)
**Purpose:** Complete Dell iDRAC management

**Key Features:**
- 27 comprehensive commands
- Power management
- Virtual media (ISO mounting)
- User management
- Network configuration
- Lifecycle Controller (firmware updates)
- RAID configuration
- System Event Log (SEL)

**Compatibility:** Dell iDRAC 7, 8, 9, 10+

**Size:** 503KB

---

### 3. idrac_server_config_profile (Dell SCP)
**Purpose:** Dell iDRAC Server Configuration Profile management

**Key Features:**
- Export server configuration (XML/JSON)
- Import server configuration
- Preview changes before applying
- Component selection (ALL, IDRAC, BIOS, NIC, RAID)
- Job tracking and monitoring
- Configurable shutdown behavior

**Compatibility:** Dell iDRAC 7+ with Lifecycle Controller

**Size:** 481KB

---

### 4. ilo (HP iLO Management)
**Purpose:** Complete HP iLO management

**Key Features:**
- 53 comprehensive operations
- Power management
- Virtual media (CD/DVD/Floppy/USB)
- User management
- Network configuration
- Firmware updates
- Logs (IML, iLO Event Log)
- License management
- Security features (CSR, certificates)

**Compatibility:** HP iLO 4, iLO 5, iLO 6

**Size:** 503KB

---

### 5. ipmi (Generic IPMI)
**Purpose:** Standard IPMI management via ipmitool

**Key Features:**
- 50+ IPMI commands
- Power management
- Chassis management
- Sensor monitoring
- System Event Log (SEL)
- FRU inventory
- Serial Over LAN (SOL)
- User management
- LAN configuration
- SDR repository

**Compatibility:** Any IPMI 1.5 or 2.0 compliant BMC

**Size:** 483KB

---

## Comparison Matrix

| Feature | Redfish | iDRAC | iDRAC SCP | iLO | IPMI |
|---------|---------|-------|-----------|-----|------|
| Power Management | ✅ | ✅ | ❌ | ✅ | ✅ |
| Virtual Media | ✅ | ✅ | ❌ | ✅ | ❌ |
| System Inventory | ✅ | ✅ | ❌ | ✅ | ✅ |
| User Management | ❌ | ✅ | ❌ | ✅ | ✅ |
| Firmware Updates | ✅ | ✅ | ❌ | ✅ | ❌ |
| BIOS Configuration | ✅ | ✅ | ✅ | ✅ | ✅ |
| Configuration Profiles | ❌ | ❌ | ✅ | ❌ | ❌ |
| RAID Management | ❌ | ✅ | ✅ | ❌ | ❌ |
| Sensor Monitoring | ❌ | ✅ | ❌ | ✅ | ✅ |
| Network Config | ❌ | ✅ | ✅ | ✅ | ✅ |

## When to Use Each Module

### Use **redfish** when:
- You need cross-vendor compatibility
- You want a modern REST API interface
- You have mixed server hardware (Dell, HP, Lenovo, etc.)
- You need basic power/inventory/firmware operations

### Use **idrac** when:
- You have Dell servers
- You need Dell-specific features (RAID, Lifecycle Controller)
- You want virtual media support
- You need detailed system management

### Use **idrac_server_config_profile** when:
- You need to backup/restore Dell server configurations
- You're cloning server configurations
- You want to automate BIOS/RAID/NIC settings
- You need to audit server configurations

### Use **ilo** when:
- You have HP servers
- You need HP-specific features (advanced licensing, etc.)
- You want comprehensive virtual media support
- You need IML (Integrated Management Log) access

### Use **ipmi** when:
- You need universal BMC support
- You have older servers without Redfish
- You need sensor monitoring
- You want SOL (Serial Over LAN) access
- You prefer command-line tools (ipmitool)

## Architecture

All modules follow the OpenFroyo shell_exec pattern:

1. **Redfish/iDRAC/iLO:** Use `curl` for REST API calls
2. **IPMI:** Use `ipmitool` for BMC commands
3. **All modules:** Return JSON with shell commands in `shell_exec` facts

## Common Usage Patterns

### Server Discovery
```yaml
# Get server information
- module: servers/redfish
  vars:
    baseuri: "https://{{server_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: GetSystemInfo
```

### Power Management
```yaml
# Graceful shutdown
- module: servers/idrac
  vars:
    idrac_ip: "{{server_ip}}"
    idrac_user: "{{bmc_user}}"
    idrac_password: "{{bmc_pass}}"
    command: graceful_shutdown
```

### Configuration Backup
```yaml
# Export Dell server config
- module: servers/idrac_server_config_profile
  vars:
    idrac_ip: "{{server_ip}}"
    idrac_user: "{{bmc_user}}"
    idrac_password: "{{bmc_pass}}"
    command: export
    export_format: XML
    scp_file: /backup/{{inventory_hostname}}_scp.xml
```

### Firmware Updates
```yaml
# Update server firmware
- module: servers/ilo
  vars:
    ilo_ip: "{{server_ip}}"
    ilo_user: "{{bmc_user}}"
    ilo_password: "{{bmc_pass}}"
    command: update_firmware
    firmware_uri: "http://firmware.local/firmware.bin"
```

### Monitoring
```yaml
# Monitor server sensors
- module: servers/ipmi
  vars:
    bmc_host: "{{server_ip}}"
    username: "{{bmc_user}}"
    password: "{{bmc_pass}}"
    command: get_all_sensors
```

## Security Considerations

### Authentication
- All modules support username/password authentication
- Use encrypted variables for credentials
- Consider using separate BMC users with minimal privileges

### Network Security
- BMCs should be on isolated management networks
- Use SSL/TLS for Redfish/iDRAC/iLO (validate_certs: true)
- Use IPMI 2.0 (lanplus) instead of 1.5 for better security

### Credential Management
- Store BMC credentials in vault or encrypted variables
- Rotate BMC passwords regularly
- Use role-based access control (RBAC)

## Dependencies

### Redfish/iDRAC/iLO Modules
- `curl` - HTTP client for REST API calls
- Network access to BMC on port 443 (HTTPS)

### IPMI Module
- `ipmitool` - IPMI command-line tool
- Network access to BMC on port 623 (IPMI)

## Troubleshooting

### Connection Issues
- Verify BMC network connectivity: `ping <bmc_ip>`
- Check firewall rules (HTTPS: 443, IPMI: 623)
- Verify credentials with direct access

### Certificate Errors
- Use `validate_certs: false` for self-signed certificates
- Import BMC certificate to trust store
- Use proper CA-signed certificates in production

### Timeout Issues
- Increase timeout value for slow operations
- Check network latency to BMC
- Verify BMC is not overloaded

### Command Failures
- Check module-specific README for command syntax
- Verify BMC firmware version compatibility
- Review BMC logs for error details

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **test.ofy** - Example test stack
- Module-specific guides (COMMANDS.md, QUICK_REFERENCE.md)

## Total Statistics

- **5 server management modules**
- **~2.4MB total WASM binaries**
- **180+ operations** across all modules
- **100% vendor compatibility** (Dell, HP, Lenovo, Supermicro, Generic)

## Future Enhancements

Potential additions:
- Lenovo XClarity module
- Cisco CIMC module
- Supermicro BMC module
- Generic BMC discovery
- Certificate management module
- Automated firmware update workflows
