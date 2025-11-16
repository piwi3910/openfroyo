# Changelog

All notable changes to the Redfish module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-16

### Added
- Initial release of the Redfish module for OpenFroyo
- Power management operations:
  - PowerOn, PowerOff, GracefulShutdown, ForceRestart
  - PowerCycle, Nmi
  - GetPowerState
- System inventory collection:
  - GetSystemInfo - manufacturer, model, serial, BIOS version
  - GetProcessorInfo - CPU details
  - GetMemoryInfo - memory module details
  - GetStorageInfo - storage controller and drive information
  - GetNetworkInfo - network interface information
- BIOS configuration management:
  - GetBiosAttributes - retrieve all BIOS settings
  - SetBiosAttributes - configure BIOS settings
  - ResetBiosToDefaults - factory reset BIOS
- Boot configuration:
  - GetBootConfiguration - current boot settings
  - SetBootSource - one-time boot source override (Pxe, Hdd, Cd, etc.)
  - SetBootMode - set UEFI or Legacy boot mode
- Firmware management:
  - GetFirmwareInventory - list all firmware versions
  - UpdateFirmware - update firmware from URI
- Security features:
  - SSL/TLS certificate validation (configurable)
  - HTTP Basic Authentication
  - Password escaping for shell safety
- Configuration options:
  - Configurable timeout (default: 30 seconds)
  - Vendor-specific resource ID support
  - Certificate validation toggle
- Comprehensive documentation:
  - README.md with full usage examples
  - COMMANDS.md quick reference guide
  - Vendor-specific notes (Dell, HP, Lenovo, Supermicro)
  - Troubleshooting guide
- Complete test suite in test.ofy
- Build system with Makefile
- WASM implementation using TinyGo
- Shell execution pattern using curl for API calls

### Implementation Details
- Built with TinyGo
- WASM binary size: 475 KB
- Total Go code: 454 lines
- Test coverage: 266 lines across all commands
- Documentation: 844 lines in README

### Supported Vendors
- Dell iDRAC (9+)
- HP iLO (4+)
- Lenovo XClarity
- Supermicro BMC
- Any Redfish 1.0+ compliant BMC

### Known Limitations
- Session-based authentication not yet implemented (uses Basic Auth)
- Dell-specific job queue management not included
- Virtual media management not included
- Event subscriptions not supported
- Async operation monitoring is manual
- State-based idempotency not fully implemented

### Future Enhancements Planned
- Session token support for better performance
- Enhanced idempotency with state checking
- Virtual media mount/unmount operations
- Dell job queue integration
- HP iLO license detection
- Task/job monitoring for async operations
- Vendor extension support
- Event subscription and monitoring

## [Unreleased]

### Planned for 1.1.0
- Session-based authentication for improved performance
- State checking for full idempotency
- Virtual media operations
- Enhanced error messages with suggested fixes

### Planned for 1.2.0
- Dell iDRAC job queue management
- HP iLO specific extensions
- Lenovo XClarity integration improvements

### Planned for 2.0.0
- Event subscription support
- Real-time monitoring capabilities
- Multi-vendor abstraction layer
- Advanced task orchestration

---

## Version History

| Version | Date | Key Features |
|---------|------|--------------|
| 1.0.0 | 2025-11-16 | Initial release with comprehensive Redfish support |
