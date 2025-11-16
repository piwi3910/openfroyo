# Dell iDRAC Server Configuration Profile Module

Import and export Dell iDRAC Server Configuration Profiles (SCP) for OpenFroyo.

## Description

This module provides Server Configuration Profile (SCP) management for Dell iDRAC. SCP files allow you to backup, restore, and replicate complete server configurations including BIOS settings, iDRAC settings, RAID configuration, network interface settings, and more.

## Variables

### Required Variables

- `idrac_ip` (string): iDRAC IP address or hostname
- `idrac_user` (string): iDRAC username
- `idrac_password` (string): iDRAC password
- `command` (string): Operation to perform (`export`, `import`, or `preview`)

### Common Optional Variables

- `timeout` (int): Request timeout in seconds (default: 120)
- `validate_certs` (bool): Validate SSL certificates (default: true)
- `share_name` (string): Network share path or local file path for SCP file
- `share_user` (string): Network share username (if applicable)
- `share_password` (string): Network share password (if applicable)

### Export-Specific Variables

- `export_format` (string): Export file format - `XML` or `JSON` (default: `XML`)
- `export_use` (string): Export profile use case (default: `Default`)
  - `Default`: Standard configuration export
  - `Clone`: Export for cloning to another server
  - `Replace`: Export for replacing current configuration
- `scp_components` (string): Components to export (default: `ALL`)
  - `ALL`: All components
  - `IDRAC`: iDRAC configuration only
  - `BIOS`: BIOS settings only
  - `NIC`: Network interface settings only
  - `RAID`: RAID controller settings only

### Import-Specific Variables

- `scp_file` (string): Path to SCP file to import (required for import/preview)
- `shutdown_type` (string): System shutdown behavior (default: `Graceful`)
  - `Graceful`: Graceful shutdown before applying changes
  - `Forced`: Force shutdown before applying changes
  - `NoReboot`: Apply changes without rebooting (limited settings)
- `end_host_power_state` (string): Desired power state after import (default: `On`)
  - `On`: Power on system after configuration
  - `Off`: Leave system powered off after configuration
- `job_wait` (bool): Wait for configuration job to complete (default: true)
- `job_wait_timeout` (int): Maximum time to wait for job completion in seconds (default: 3600)

## Supported Commands

### export

Export the current server configuration profile to a file.

**Features:**
- Export complete system configuration or specific components
- Multiple format support (XML, JSON)
- Different export profiles for different use cases
- Automatic job tracking and file download

**Returns:**
- SCP file content
- Export job status and completion details

### import

Import a server configuration profile from a file.

**Features:**
- Upload and apply SCP configuration
- Configurable shutdown behavior
- Control final power state
- Automatic job tracking with progress updates
- Handles long-running configuration jobs
- Timeout protection

**Important Notes:**
- Import may require system reboot depending on settings changed
- Some settings require specific shutdown types
- Job can take several minutes for complex configurations
- System may be unavailable during configuration

### preview

Preview the changes that would be made by importing an SCP file without actually applying them.

**Features:**
- Preview configuration changes before applying
- Identify which settings will be modified
- Validate SCP file compatibility
- No system changes made

**Use Cases:**
- Validate SCP file before import
- Review configuration differences
- Test SCP files in safe mode

## Usage Examples

### Export Full Configuration

```yaml
- name: Export complete server configuration
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "export"
    export_format: "XML"
    scp_components: "ALL"
    share_name: "/tmp/server01_full_config.xml"
```

### Export BIOS Settings Only

```yaml
- name: Export BIOS configuration
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "export"
    export_format: "JSON"
    scp_components: "BIOS"
    share_name: "/tmp/server01_bios.json"
```

### Export for Cloning

```yaml
- name: Export configuration for cloning to other servers
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "export"
    export_format: "XML"
    export_use: "Clone"
    scp_components: "ALL"
    share_name: "/tmp/clone_template.xml"
```

### Export RAID Configuration

```yaml
- name: Export RAID controller configuration
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "export"
    scp_components: "RAID"
    share_name: "/tmp/raid_config.xml"
```

### Preview Import

```yaml
- name: Preview configuration changes
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "preview"
    scp_file: "/tmp/new_config.xml"
```

### Import with Graceful Shutdown

```yaml
- name: Import server configuration with graceful shutdown
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "import"
    scp_file: "/tmp/server01_config.xml"
    shutdown_type: "Graceful"
    end_host_power_state: "On"
    job_wait: true
    job_wait_timeout: 3600
```

### Import Without Reboot

```yaml
- name: Import configuration without rebooting
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "import"
    scp_file: "/tmp/idrac_settings.xml"
    shutdown_type: "NoReboot"
    job_wait: true
```

### Import and Leave Powered Off

```yaml
- name: Import configuration and leave server off
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "import"
    scp_file: "/tmp/config.xml"
    shutdown_type: "Forced"
    end_host_power_state: "Off"
    job_wait: true
```

### Import with Extended Timeout

```yaml
- name: Import large configuration with extended timeout
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "import"
    scp_file: "/tmp/full_config.xml"
    shutdown_type: "Graceful"
    end_host_power_state: "On"
    job_wait: true
    job_wait_timeout: 7200
```

### Import Without Waiting for Completion

```yaml
- name: Import configuration asynchronously
  module: idrac_server_config_profile
  vars:
    idrac_ip: "192.168.1.100"
    idrac_user: "root"
    idrac_password: "calvin"
    command: "import"
    scp_file: "/tmp/config.xml"
    job_wait: false
```

## Complete Workflow Example

```yaml
# Stack for server configuration management
name: Server Configuration Backup and Restore
inventory: inventory/datacenter.yml

defaults:
  idrac_user: "root"
  idrac_password: "calvin"
  validate_certs: false

run:
  # Export configurations from all servers
  - name: Backup server configurations
    module: idrac_server_config_profile
    hosts: "@group:production_servers"
    vars:
      idrac_ip: "{{ host.idrac_ip }}"
      command: "export"
      export_format: "XML"
      scp_components: "ALL"
      share_name: "/backup/{{ host.name }}_config_{{ date }}.xml"

  # Clone configuration to new server
  - name: Preview configuration on new server
    module: idrac_server_config_profile
    hosts: "new-server-01"
    vars:
      idrac_ip: "192.168.1.200"
      command: "preview"
      scp_file: "/backup/template_server_config.xml"

  - name: Apply configuration to new server
    module: idrac_server_config_profile
    hosts: "new-server-01"
    vars:
      idrac_ip: "192.168.1.200"
      command: "import"
      scp_file: "/backup/template_server_config.xml"
      shutdown_type: "Graceful"
      end_host_power_state: "On"
      job_wait: true
```

## Implementation Details

### Export Process

1. Send export request to iDRAC Redfish API
2. iDRAC creates a job to generate the SCP file
3. Poll job status until completion
4. Download generated SCP file content
5. Save to specified file path

### Import Process

1. Read SCP file from local filesystem
2. Base64 encode SCP content
3. Send import request with encoded content
4. iDRAC creates configuration job
5. Poll job status with progress updates
6. Wait for job completion (if job_wait=true)
7. System may reboot during process

### Job Tracking

- Polls job status every 15 seconds
- Reports progress percentage and messages
- Handles job states: Scheduled, Running, Completed, Failed
- Configurable timeout to prevent infinite waiting
- Detailed error reporting on failure

## SCP File Format

### XML Format Example

```xml
<SystemConfiguration>
  <Component FQDD="iDRAC.Embedded.1">
    <Attribute Name="IPv4.1.Address">192.168.1.100</Attribute>
    <Attribute Name="IPv4.1.Netmask">255.255.255.0</Attribute>
  </Component>
  <Component FQDD="BIOS.Setup.1-1">
    <Attribute Name="BootMode">Uefi</Attribute>
    <Attribute Name="SriovGlobalEnable">Enabled</Attribute>
  </Component>
</SystemConfiguration>
```

### JSON Format Example

```json
{
  "SystemConfiguration": {
    "Components": [
      {
        "FQDD": "iDRAC.Embedded.1",
        "Attributes": [
          {
            "Name": "IPv4.1.Address",
            "Value": "192.168.1.100"
          }
        ]
      }
    ]
  }
}
```

## Return Values

All commands return:
- `status`: "ok", "changed", or "failed"
- `message`: Human-readable description of the operation
- `facts`: Operation details including:
  - Job URI and status
  - SCP file content (export)
  - Progress updates (import)
  - Error details (if failed)

## Error Handling

The module handles various error scenarios:
- Invalid SCP file path or format
- Network connectivity issues
- Authentication failures
- Job timeouts
- Configuration conflicts
- System state incompatibilities

## Performance Considerations

- **Export**: Typically completes in 30-60 seconds
- **Import**: Can take 5-30 minutes depending on:
  - Number of settings changed
  - Whether reboot is required
  - RAID configuration changes
  - Firmware updates triggered
- **Preview**: Similar to export, 30-60 seconds

## Building

```bash
cd modules/servers/idrac_server_config_profile
make build
```

## Requirements

- Dell PowerEdge server with iDRAC 7, 8, or 9
- iDRAC Enterprise license (required for SCP operations)
- Network connectivity to iDRAC interface
- Valid iDRAC credentials with configuration privileges
- Sufficient disk space for SCP files

## Best Practices

1. **Always preview before importing** in production environments
2. **Backup current configuration** before importing new settings
3. **Use appropriate shutdown type** - Graceful for production, Forced only when necessary
4. **Set adequate timeouts** for complex configurations
5. **Monitor job progress** by keeping job_wait=true
6. **Test SCP files** on non-production systems first
7. **Use Clone export type** when replicating configurations
8. **Store SCP files securely** as they contain sensitive configuration data

## Troubleshooting

### Import Fails with "NoReboot" Shutdown Type

Some settings require a system reboot. Use "Graceful" or "Forced" shutdown type instead.

### Job Timeout

Increase `job_wait_timeout` for complex configurations or slow systems.

### Invalid SCP File

Ensure SCP file:
- Is valid XML or JSON format
- Matches target server hardware
- Contains compatible firmware versions
- Was exported with appropriate `export_use` setting

### Authentication Errors

Verify:
- iDRAC credentials are correct
- User has Administrator privileges
- iDRAC is accessible over network

## References

- [Dell iDRAC Server Configuration Profile Reference](https://www.dell.com/support/article/en-us/sln296511/idrac-server-configuration-profile)
- [Dell Redfish API Guide](https://www.dell.com/support/article/en-us/sln310367/idrac-redfish-api-overview)
- [Ansible idrac_server_config_profile Module](https://docs.ansible.com/ansible/2.9/modules/idrac_server_config_profile_module.html)
