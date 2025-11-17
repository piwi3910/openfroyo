# Apache HTTP Server Module

Comprehensive Apache HTTP Server management module for OpenFroyo with 30+ operations covering virtual host management, SSL/TLS configuration, module management, service control, and performance tuning.

## Overview

The Apache module provides complete lifecycle management for Apache HTTP Server across both Debian/Ubuntu (apache2) and RHEL/CentOS (httpd) platforms. It handles configuration management, virtual host setup, SSL certificate installation, module management, and performance optimization.

## Features

- **Virtual Host Management**: Create, delete, enable, and disable virtual hosts
- **SSL/TLS Configuration**: Full SSL certificate management and security settings
- **Module Management**: Enable/disable Apache modules with automatic reloads
- **Service Control**: Start, stop, restart, and check service status
- **Performance Tuning**: Configure MPM settings, KeepAlive, timeouts, and connection limits
- **Security Headers**: Configure modern security headers (X-Frame-Options, HSTS, CSP)
- **Reverse Proxy**: Create reverse proxy configurations
- **Configuration Testing**: Validate configurations before applying changes
- **Cross-Platform**: Automatic detection of Debian vs RHEL-based systems

## Apache HTTP Server Concepts

### Virtual Hosts

Virtual hosts allow Apache to serve multiple websites from a single server. Each virtual host has its own configuration including:

- **ServerName**: Primary domain name
- **ServerAlias**: Additional domain names
- **DocumentRoot**: Directory containing website files
- **Directory Options**: Access controls and feature settings
- **Logging**: Separate error and access logs

### MPM (Multi-Processing Modules)

Apache uses MPM modules to handle client connections. Three main types:

- **prefork**: Traditional forking model (best for mod_php)
- **worker**: Hybrid multi-process multi-threaded model
- **event**: Similar to worker but optimized for keep-alive connections

### Module System

Apache functionality is extended through modules:

- **Core Modules**: Built into Apache (http_core)
- **Static Modules**: Compiled into binary
- **Shared Modules**: Loaded dynamically (.so files)

Common modules: ssl, rewrite, headers, proxy, proxy_http, auth_basic

### Configuration Hierarchy

Debian/Ubuntu structure:
```
/etc/apache2/
├── apache2.conf           # Main configuration
├── ports.conf             # Port listening configuration
├── sites-available/       # Virtual host definitions
├── sites-enabled/         # Active virtual hosts (symlinks)
├── mods-available/        # Available modules
├── mods-enabled/          # Active modules (symlinks)
├── conf-available/        # Additional configurations
└── conf-enabled/          # Active configurations (symlinks)
```

RHEL/CentOS structure:
```
/etc/httpd/
├── conf/
│   └── httpd.conf         # Main configuration
├── conf.d/                # Additional configurations
├── conf.modules.d/        # Module loading
└── sites-enabled/         # Virtual hosts (manual setup)
```

## Installation

1. Ensure Go 1.21+ is installed
2. Build the WASM module:
   ```bash
   cd modules/web/apache/wasm
   make build
   ```

## Quick Start

### Basic Virtual Host Setup

```yaml
# Create a simple virtual host
- name: Create website virtual host
  module: apache
  vars:
    command: create_vhost
    vhost_name: example.com
    document_root: /var/www/example.com
    server_name: example.com
    server_alias: www.example.com

# Enable the virtual host
- name: Enable virtual host
  module: apache
  vars:
    command: enable_vhost
    vhost_name: example.com
    reload_service: true
```

### SSL/TLS Virtual Host

```yaml
# Create SSL-enabled virtual host
- name: Create SSL virtual host
  module: apache
  vars:
    command: create_ssl_vhost
    vhost_name: secure.example.com
    document_root: /var/www/secure
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    ssl_certificate_chain: /etc/ssl/certs/chain.crt
    hsts_enabled: true
    redirect_http_to_https: true

# Enable SSL virtual host
- name: Enable SSL virtual host
  module: apache
  vars:
    command: enable_vhost
    vhost_name: secure.example.com
    reload_service: true
```

### Module Management

```yaml
# Enable required modules
- name: Enable SSL module
  module: apache
  vars:
    command: enable_module
    module_name: ssl
    reload_service: true

# Enable rewrite module
- name: Enable rewrite module
  module: apache
  vars:
    command: enable_module
    module_name: rewrite
    reload_service: true
```

## Operations Reference

### Virtual Host Management (6 operations)

| Operation | Description |
|-----------|-------------|
| `create_vhost` | Create a new virtual host configuration |
| `delete_vhost` | Delete a virtual host configuration |
| `enable_vhost` | Enable a virtual host (create symlink) |
| `disable_vhost` | Disable a virtual host (remove symlink) |
| `list_vhosts` | List all available and enabled virtual hosts |
| `test_vhost_config` | Test virtual host configuration syntax |

### Module Management (5 operations)

| Operation | Description |
|-----------|-------------|
| `enable_module` | Enable an Apache module |
| `disable_module` | Disable an Apache module |
| `list_modules` | List all loaded/enabled modules |
| `list_available_modules` | List all available modules |
| `check_module_loaded` | Check if a specific module is loaded |

### SSL/TLS Configuration (5 operations)

| Operation | Description |
|-----------|-------------|
| `create_ssl_vhost` | Create SSL/TLS enabled virtual host |
| `install_ssl_certificate` | Install SSL certificate for existing vhost |
| `set_ssl_protocols` | Configure SSL/TLS protocols |
| `set_ssl_ciphers` | Configure SSL/TLS cipher suites |
| `test_ssl_config` | Test SSL/TLS configuration |

### Configuration Management (6 operations)

| Operation | Description |
|-----------|-------------|
| `set_server_name` | Set ServerName directive |
| `set_document_root` | Set DocumentRoot directive |
| `set_directory_options` | Configure directory options |
| `add_directory_block` | Add a Directory configuration block |
| `test_config` | Test Apache configuration syntax |
| `reload_config` | Reload Apache configuration (graceful restart) |

### Service Management (4 operations)

| Operation | Description |
|-----------|-------------|
| `start_service` | Start Apache service |
| `stop_service` | Stop Apache service |
| `restart_service` | Restart Apache service |
| `check_service_status` | Check Apache service status |

### Performance Tuning (4 operations)

| Operation | Description |
|-----------|-------------|
| `set_mpm_config` | Configure Multi-Processing Module settings |
| `set_keepalive` | Configure KeepAlive settings |
| `set_timeout` | Configure timeout settings |
| `set_max_clients` | Configure maximum client connections |

### Additional Operations (5 operations)

| Operation | Description |
|-----------|-------------|
| `set_log_level` | Set Apache log level |
| `enable_conf` | Enable a configuration file |
| `disable_conf` | Disable a configuration file |
| `create_proxy_vhost` | Create a reverse proxy virtual host |
| `set_security_headers` | Configure security headers |

See [COMMANDS.md](COMMANDS.md) for detailed parameter documentation.

## Configuration Variables

The module uses variables from `defaults.ofy.yml` with platform-specific defaults. Key variables include:

### Service Configuration
- `apache_service`: Service name (apache2 or httpd)
- `apache_binary`: Binary name (apache2ctl or apachectl)
- `apache_config_dir`: Main configuration directory

### Virtual Host Settings
- `default_port`: Default HTTP port (80)
- `default_ssl_port`: Default HTTPS port (443)
- `default_document_root`: Default document root path

### SSL/TLS Settings
- `ssl_protocols`: Supported SSL/TLS protocols
- `ssl_ciphers`: SSL cipher suite configuration
- `ssl_prefer_server_ciphers`: Prefer server cipher order

### Performance Settings
- `default_mpm`: Default MPM module (prefork, worker, event)
- `default_keep_alive`: KeepAlive setting (On/Off)
- `default_timeout`: Request timeout in seconds
- `default_max_clients`: Maximum concurrent connections

### Security Settings
- `server_tokens`: Server information disclosure level
- `server_signature`: Include server signature in pages
- `trace_enable`: Enable TRACE method

See [defaults.ofy.yml](defaults.ofy.yml) for complete variable reference.

## Platform Support

### Debian/Ubuntu
- Uses `apache2` service name
- Uses `a2ensite`, `a2dissite`, `a2enmod`, `a2dismod` commands
- Configuration in `/etc/apache2/`
- Symlink-based site/module management

### RHEL/CentOS
- Uses `httpd` service name
- Manual symlink management for sites
- Configuration in `/etc/httpd/`
- Module configuration in `conf.modules.d/`

The module automatically detects the platform and uses appropriate commands.

## Best Practices

### Virtual Host Configuration
1. Always test configuration before reloading: `test_config: true`
2. Use separate virtual hosts for HTTP and HTTPS
3. Implement proper directory permissions (755 for directories, 644 for files)
4. Use descriptive virtual host names
5. Keep virtual host files organized in sites-available

### SSL/TLS Security
1. Use modern TLS versions only (TLSv1.2, TLSv1.3)
2. Configure strong cipher suites
3. Enable HSTS for HTTPS sites
4. Redirect HTTP to HTTPS for secure sites
5. Keep SSL certificates up to date

### Performance Optimization
1. Choose appropriate MPM module for workload
2. Configure KeepAlive for persistent connections
3. Set reasonable timeouts to free resources
4. Monitor and adjust MaxRequestWorkers based on traffic
5. Enable compression (mod_deflate)

### Security Hardening
1. Disable unnecessary modules
2. Configure security headers (X-Frame-Options, CSP)
3. Set ServerTokens to Prod
4. Disable ServerSignature
5. Disable TRACE method
6. Use mod_security for additional protection

### Module Management
1. Only enable required modules
2. Reload service after module changes
3. Test configuration after enabling modules
4. Document module dependencies

## Troubleshooting

### Configuration Test Failures
```yaml
# Always test configuration before applying
- name: Test Apache configuration
  module: apache
  vars:
    command: test_config
    verbose: true
    show_vhosts: true
```

### Service Not Starting
1. Check configuration syntax: `test_config`
2. Check port conflicts: `netstat -tlnp | grep :80`
3. Check file permissions on document roots
4. Check error logs: `/var/log/apache2/error.log`

### Virtual Host Not Working
1. Verify virtual host is enabled: `list_vhosts`
2. Check ServerName matches request
3. Verify DNS resolution
4. Check firewall rules
5. Review access logs

### SSL Certificate Issues
1. Verify certificate files exist and are readable
2. Check certificate validity: `openssl x509 -in cert.crt -text -noout`
3. Verify certificate matches key
4. Check certificate chain is complete
5. Enable SSL module: `enable_module module_name=ssl`

### Performance Issues
1. Monitor current connections: `check_service_status`
2. Review MPM configuration
3. Check for slow backend services
4. Enable caching (mod_cache)
5. Monitor system resources (CPU, memory, disk I/O)

## Examples

See [WORKFLOWS.md](WORKFLOWS.md) for complete workflow examples including:
- Multi-site hosting setup
- SSL certificate installation
- Reverse proxy configuration
- Load balancer setup
- Security hardening
- Performance tuning

## Testing

A test stack is provided in `test.ofy` demonstrating all major operations:

```bash
froyo apply modules/web/apache/test.ofy
```

## Security Considerations

1. **File Permissions**: Ensure proper ownership and permissions on configuration files
2. **Certificate Storage**: Store private keys securely with restricted permissions (600)
3. **Configuration Backups**: Backup configurations before making changes
4. **Access Controls**: Implement proper authentication for sensitive directories
5. **Regular Updates**: Keep Apache updated with security patches
6. **Audit Logs**: Monitor access and error logs for suspicious activity

## Contributing

When contributing new operations:

1. Add operation definition to `module.ofy.yml`
2. Implement operation handler in `wasm/main.go`
3. Update `COMMANDS.md` with parameter documentation
4. Add examples to `WORKFLOWS.md`
5. Update `test.ofy` with test cases
6. Rebuild WASM module: `make build`

## Resources

- [Apache HTTP Server Documentation](https://httpd.apache.org/docs/)
- [Apache Security Tips](https://httpd.apache.org/docs/2.4/misc/security_tips.html)
- [Let's Encrypt SSL Certificates](https://letsencrypt.org/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- [Apache Performance Tuning](https://httpd.apache.org/docs/2.4/misc/perf-tuning.html)

## License

This module is part of the OpenFroyo project.
