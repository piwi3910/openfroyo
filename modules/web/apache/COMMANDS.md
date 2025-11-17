# Apache Module - Command Reference

Complete reference for all 30+ operations supported by the Apache module.

## Table of Contents

1. [Virtual Host Management](#virtual-host-management)
2. [Module Management](#module-management)
3. [SSL/TLS Configuration](#ssltls-configuration)
4. [Configuration Management](#configuration-management)
5. [Service Management](#service-management)
6. [Performance Tuning](#performance-tuning)
7. [Additional Operations](#additional-operations)

---

## Virtual Host Management

### create_vhost

Create a new virtual host configuration file.

**Required Variables:**
- `vhost_name` (string): Name for the virtual host (used in filenames and logs)
- `document_root` (string): Path to website files

**Optional Variables:**
- `vhost_file` (string): Configuration filename (default: `{vhost_name}.conf`)
- `server_name` (string): Primary domain name (default: vhost_name)
- `server_alias` (string): Additional domain names (space-separated)
- `port` (string): Port number (default: 80)
- `server_admin` (string): Administrator email (default: from defaults)
- `directory_options` (string): Directory Options directive (default: "Indexes FollowSymLinks")
- `allow_override` (string): AllowOverride directive (default: "None")
- `require` (string): Require directive (default: "all granted")
- `error_log` (string): Error log path
- `access_log` (string): Access log path
- `custom_config` (string): Additional configuration directives

**Example:**
```yaml
- name: Create example.com virtual host
  module: apache
  vars:
    command: create_vhost
    vhost_name: example.com
    document_root: /var/www/example.com
    server_name: example.com
    server_alias: www.example.com
    port: 80
    allow_override: All
```

**Output Facts:**
- `vhost_name`: Name of the virtual host
- `vhost_file`: Full path to configuration file
- `document_root`: Document root path

---

### delete_vhost

Delete a virtual host configuration.

**Required Variables:**
- `vhost_name` (string): Name of the virtual host to delete

**Optional Variables:**
- `vhost_file` (string): Configuration filename (default: `{vhost_name}.conf`)
- `remove_document_root` (bool): Also remove document root directory (default: false)

**Example:**
```yaml
- name: Delete old virtual host
  module: apache
  vars:
    command: delete_vhost
    vhost_name: old-site.com
    remove_document_root: true
```

---

### enable_vhost

Enable a virtual host by creating a symlink in sites-enabled.

**Required Variables:**
- `vhost_name` (string): Name of the virtual host to enable

**Optional Variables:**
- `vhost_file` (string): Configuration filename (default: `{vhost_name}.conf`)
- `reload_service` (bool): Reload Apache after enabling (default: true)

**Example:**
```yaml
- name: Enable example.com
  module: apache
  vars:
    command: enable_vhost
    vhost_name: example.com
    reload_service: true
```

---

### disable_vhost

Disable a virtual host by removing the symlink from sites-enabled.

**Required Variables:**
- `vhost_name` (string): Name of the virtual host to disable

**Optional Variables:**
- `vhost_file` (string): Configuration filename (default: `{vhost_name}.conf`)
- `reload_service` (bool): Reload Apache after disabling (default: true)

**Example:**
```yaml
- name: Disable maintenance site
  module: apache
  vars:
    command: disable_vhost
    vhost_name: maintenance.example.com
```

---

### list_vhosts

List all available and enabled virtual hosts.

**Optional Variables:**
- `show_details` (bool): Show detailed virtual host information (default: false)

**Example:**
```yaml
- name: List all virtual hosts
  module: apache
  vars:
    command: list_vhosts
    show_details: true
```

**Output Facts:**
- `available_vhosts`: Array of available virtual host files
- `enabled_vhosts`: Array of enabled virtual host files

---

### test_vhost_config

Test a virtual host configuration for syntax errors.

**Required Variables:**
- `vhost_file` (string): Virtual host configuration file to test

**Optional Variables:**
- `verbose` (bool): Show detailed test output (default: false)

**Example:**
```yaml
- name: Test virtual host configuration
  module: apache
  vars:
    command: test_vhost_config
    vhost_file: example.com.conf
    verbose: true
```

---

## Module Management

### enable_module

Enable an Apache module.

**Required Variables:**
- `module_name` (string): Name of the module to enable (e.g., ssl, rewrite, headers)

**Optional Variables:**
- `reload_service` (bool): Reload Apache after enabling (default: true)

**Example:**
```yaml
- name: Enable SSL module
  module: apache
  vars:
    command: enable_module
    module_name: ssl
    reload_service: true
```

---

### disable_module

Disable an Apache module.

**Required Variables:**
- `module_name` (string): Name of the module to disable

**Optional Variables:**
- `reload_service` (bool): Reload Apache after disabling (default: true)

**Example:**
```yaml
- name: Disable status module
  module: apache
  vars:
    command: disable_module
    module_name: status
```

---

### list_modules

List all currently loaded/enabled modules.

**Optional Variables:**
- `format` (string): Output format - "text" or "json" (default: "text")

**Example:**
```yaml
- name: List loaded modules
  module: apache
  vars:
    command: list_modules
    format: json
```

**Output Facts:**
- `loaded_modules`: Array of loaded module names
- `module_count`: Number of loaded modules

---

### list_available_modules

List all available modules (not necessarily loaded).

**Optional Variables:**
- `format` (string): Output format - "text" or "json" (default: "text")

**Example:**
```yaml
- name: List available modules
  module: apache
  vars:
    command: list_available_modules
```

**Output Facts:**
- `available_modules`: Array of available module names
- `module_count`: Number of available modules

---

### check_module_loaded

Check if a specific module is currently loaded.

**Required Variables:**
- `module_name` (string): Name of the module to check

**Example:**
```yaml
- name: Check if SSL is loaded
  module: apache
  vars:
    command: check_module_loaded
    module_name: ssl
```

**Output Facts:**
- `module_name`: Name of the checked module
- `is_loaded`: Boolean indicating if module is loaded

---

## SSL/TLS Configuration

### create_ssl_vhost

Create a new SSL/TLS enabled virtual host.

**Required Variables:**
- `vhost_name` (string): Name for the virtual host
- `document_root` (string): Path to website files
- `ssl_certificate` (string): Path to SSL certificate file
- `ssl_certificate_key` (string): Path to SSL private key file

**Optional Variables:**
- `vhost_file` (string): Configuration filename (default: `{vhost_name}-ssl.conf`)
- `server_name` (string): Primary domain name
- `server_alias` (string): Additional domain names
- `ssl_port` (string): SSL port number (default: 443)
- `ssl_certificate_chain` (string): Path to certificate chain file
- `ssl_protocols` (string): SSL/TLS protocols (default: "TLSv1.2 TLSv1.3")
- `ssl_ciphers` (string): SSL cipher suite
- `hsts_enabled` (bool): Enable HSTS header (default: true)
- `redirect_http_to_https` (bool): Add HTTP to HTTPS redirect (default: true)

**Example:**
```yaml
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
```

**Output Facts:**
- `vhost_name`: Name of the virtual host
- `vhost_file`: Full path to configuration file
- `document_root`: Document root path
- `ssl_enabled`: Boolean (true)
- `hsts_enabled`: Boolean indicating HSTS status
- `http_redirected`: Boolean indicating HTTP redirect status

---

### install_ssl_certificate

Install or update SSL certificate for an existing virtual host.

**Required Variables:**
- `vhost_file` (string): Virtual host configuration file
- `ssl_certificate` (string): Path to SSL certificate file
- `ssl_certificate_key` (string): Path to SSL private key file

**Optional Variables:**
- `ssl_certificate_chain` (string): Path to certificate chain file

**Example:**
```yaml
- name: Update SSL certificate
  module: apache
  vars:
    command: install_ssl_certificate
    vhost_file: example.com-ssl.conf
    ssl_certificate: /etc/ssl/certs/example.com-new.crt
    ssl_certificate_key: /etc/ssl/private/example.com-new.key
```

---

### set_ssl_protocols

Configure SSL/TLS protocol versions.

**Required Variables:**
- `ssl_protocols` (string): Space-separated list of protocols (e.g., "TLSv1.2 TLSv1.3")

**Optional Variables:**
- `vhost_file` (string): Virtual host file (for vhost scope)
- `config_scope` (string): "global" or "vhost" (default: "global")

**Example:**
```yaml
- name: Set SSL protocols globally
  module: apache
  vars:
    command: set_ssl_protocols
    ssl_protocols: "TLSv1.2 TLSv1.3"
    config_scope: global
```

---

### set_ssl_ciphers

Configure SSL/TLS cipher suites.

**Required Variables:**
- `ssl_ciphers` (string): Cipher suite specification

**Optional Variables:**
- `vhost_file` (string): Virtual host file (for vhost scope)
- `ssl_prefer_server_ciphers` (string): "On" or "Off" (default: "On")
- `config_scope` (string): "global" or "vhost" (default: "global")

**Example:**
```yaml
- name: Configure SSL ciphers
  module: apache
  vars:
    command: set_ssl_ciphers
    ssl_ciphers: "HIGH:!aNULL:!MD5:!RC4"
    ssl_prefer_server_ciphers: "On"
```

---

### test_ssl_config

Test SSL/TLS configuration.

**Optional Variables:**
- `vhost_name` (string): Specific virtual host to test
- `check_certificates` (bool): Validate certificate files (default: false)

**Example:**
```yaml
- name: Test SSL configuration
  module: apache
  vars:
    command: test_ssl_config
    check_certificates: true
```

---

## Configuration Management

### set_server_name

Set the ServerName directive.

**Required Variables:**
- `server_name` (string): Server name (FQDN)

**Optional Variables:**
- `vhost_file` (string): Virtual host file (for vhost scope)
- `config_scope` (string): "global" or "vhost" (default: "global")

**Example:**
```yaml
- name: Set global server name
  module: apache
  vars:
    command: set_server_name
    server_name: www.example.com
    config_scope: global
```

---

### set_document_root

Set the DocumentRoot directive.

**Required Variables:**
- `document_root` (string): Path to document root

**Optional Variables:**
- `vhost_file` (string): Virtual host configuration file (required for vhost scope)
- `create_directory` (bool): Create directory if it doesn't exist (default: true)

**Example:**
```yaml
- name: Set document root
  module: apache
  vars:
    command: set_document_root
    vhost_file: example.com.conf
    document_root: /var/www/example.com/public
    create_directory: true
```

---

### set_directory_options

Configure directory options for an existing Directory block.

**Required Variables:**
- `directory_path` (string): Path to the directory

**Optional Variables:**
- `vhost_file` (string): Virtual host configuration file
- `options` (string): Options directive value
- `allow_override` (string): AllowOverride directive value
- `require` (string): Require directive value
- `index_options` (string): DirectoryIndex directive value

**Example:**
```yaml
- name: Configure directory options
  module: apache
  vars:
    command: set_directory_options
    directory_path: /var/www/example.com
    vhost_file: example.com.conf
    options: "FollowSymLinks"
    allow_override: "All"
```

---

### add_directory_block

Add a new Directory configuration block.

**Required Variables:**
- `directory_path` (string): Path to the directory
- `vhost_file` (string): Virtual host configuration file

**Optional Variables:**
- `options` (string): Options directive (default: "Indexes FollowSymLinks")
- `allow_override` (string): AllowOverride directive (default: "None")
- `require` (string): Require directive (default: "all granted")
- `custom_directives` (string): Additional custom directives

**Example:**
```yaml
- name: Add directory block
  module: apache
  vars:
    command: add_directory_block
    directory_path: /var/www/example.com/admin
    vhost_file: example.com.conf
    options: "None"
    allow_override: "None"
    require: "ip 192.168.1.0/24"
```

---

### test_config

Test Apache configuration syntax.

**Optional Variables:**
- `verbose` (bool): Show detailed output (default: false)
- `show_vhosts` (bool): Show virtual host summary (default: false)

**Example:**
```yaml
- name: Test Apache configuration
  module: apache
  vars:
    command: test_config
    verbose: true
    show_vhosts: true
```

---

### reload_config

Reload Apache configuration (graceful restart).

**Optional Variables:**
- `test_first` (bool): Test configuration before reloading (default: true)

**Example:**
```yaml
- name: Reload Apache configuration
  module: apache
  vars:
    command: reload_config
    test_first: true
```

---

## Service Management

### start_service

Start the Apache service.

**Optional Variables:**
- `test_config_first` (bool): Test configuration before starting (default: true)

**Example:**
```yaml
- name: Start Apache
  module: apache
  vars:
    command: start_service
    test_config_first: true
```

---

### stop_service

Stop the Apache service.

**Optional Variables:**
- `graceful` (bool): Perform graceful stop (default: false)

**Example:**
```yaml
- name: Stop Apache gracefully
  module: apache
  vars:
    command: stop_service
    graceful: true
```

---

### restart_service

Restart the Apache service.

**Optional Variables:**
- `graceful` (bool): Perform graceful restart (default: true)
- `test_config_first` (bool): Test configuration before restarting (default: true)

**Example:**
```yaml
- name: Restart Apache
  module: apache
  vars:
    command: restart_service
    graceful: true
    test_config_first: true
```

---

### check_service_status

Check Apache service status.

**Optional Variables:**
- `verbose` (bool): Show detailed status information (default: false)

**Example:**
```yaml
- name: Check Apache status
  module: apache
  vars:
    command: check_service_status
    verbose: true
```

**Output Facts:**
- `service_name`: Name of the Apache service
- `is_running`: Boolean indicating if service is running

---

## Performance Tuning

### set_mpm_config

Configure Multi-Processing Module (MPM) settings.

**Required Variables:**
- `mpm_module` (string): MPM module name - "prefork", "worker", or "event"

**Optional Variables:**
- `start_servers` (int): Number of child server processes to start
- `min_spare_threads` (int): Minimum number of idle threads
- `max_spare_threads` (int): Maximum number of idle threads
- `threads_per_child` (int): Number of threads per child process
- `max_request_workers` (int): Maximum number of request workers
- `max_connections_per_child` (int): Max connections per child (0 = unlimited)
- `config_file` (string): Custom configuration file path

**Example:**
```yaml
- name: Configure event MPM
  module: apache
  vars:
    command: set_mpm_config
    mpm_module: event
    start_servers: 3
    min_spare_threads: 75
    max_spare_threads: 250
    threads_per_child: 25
    max_request_workers: 400
```

---

### set_keepalive

Configure KeepAlive settings.

**Required Variables:**
- `keep_alive` (string): "On" or "Off"

**Optional Variables:**
- `keep_alive_timeout` (int): Timeout in seconds
- `max_keep_alive_requests` (int): Maximum requests per connection
- `config_scope` (string): "global" or "vhost" (default: "global")
- `vhost_file` (string): Virtual host file (for vhost scope)

**Example:**
```yaml
- name: Configure KeepAlive
  module: apache
  vars:
    command: set_keepalive
    keep_alive: "On"
    keep_alive_timeout: 5
    max_keep_alive_requests: 100
```

---

### set_timeout

Configure timeout settings.

**Required Variables:**
- `timeout` (int): Request timeout in seconds

**Optional Variables:**
- `request_read_timeout` (int): Request read timeout
- `config_scope` (string): "global" or "vhost" (default: "global")

**Example:**
```yaml
- name: Set timeout
  module: apache
  vars:
    command: set_timeout
    timeout: 300
    request_read_timeout: 20
```

---

### set_max_clients

Configure maximum concurrent client connections.

**Required Variables:**
- `max_clients` (int): Maximum number of concurrent connections

**Optional Variables:**
- `mpm_module` (string): MPM module name (default: "prefork")

**Example:**
```yaml
- name: Set max clients
  module: apache
  vars:
    command: set_max_clients
    max_clients: 256
    mpm_module: event
```

---

## Additional Operations

### set_log_level

Set Apache log level.

**Required Variables:**
- `log_level` (string): Log level - "debug", "info", "notice", "warn", "error", "crit", "alert", "emerg"

**Optional Variables:**
- `vhost_file` (string): Virtual host file (for vhost scope)
- `config_scope` (string): "global" or "vhost" (default: "global")

**Example:**
```yaml
- name: Set log level to info
  module: apache
  vars:
    command: set_log_level
    log_level: info
    config_scope: global
```

---

### enable_conf

Enable a configuration file.

**Required Variables:**
- `conf_name` (string): Configuration name (without .conf extension)

**Optional Variables:**
- `reload_service` (bool): Reload Apache after enabling (default: true)

**Example:**
```yaml
- name: Enable security configuration
  module: apache
  vars:
    command: enable_conf
    conf_name: security
```

---

### disable_conf

Disable a configuration file.

**Required Variables:**
- `conf_name` (string): Configuration name (without .conf extension)

**Optional Variables:**
- `reload_service` (bool): Reload Apache after disabling (default: true)

**Example:**
```yaml
- name: Disable charset configuration
  module: apache
  vars:
    command: disable_conf
    conf_name: charset
```

---

### create_proxy_vhost

Create a reverse proxy virtual host.

**Required Variables:**
- `vhost_name` (string): Name for the virtual host
- `proxy_target` (string): Backend server URL (e.g., http://localhost:8080)

**Optional Variables:**
- `vhost_file` (string): Configuration filename
- `server_name` (string): Primary domain name
- `port` (string): Port number (default: 80)
- `proxy_preserve_host` (bool): Preserve Host header (default: true)
- `proxy_timeout` (int): Proxy timeout in seconds (default: 300)
- `ssl_enabled` (bool): Enable SSL for proxy (default: false)

**Example:**
```yaml
- name: Create reverse proxy
  module: apache
  vars:
    command: create_proxy_vhost
    vhost_name: app.example.com
    proxy_target: http://localhost:3000
    server_name: app.example.com
    proxy_preserve_host: true
    proxy_timeout: 300
```

**Output Facts:**
- `vhost_name`: Name of the virtual host
- `vhost_file`: Full path to configuration file
- `proxy_target`: Backend target URL

---

### set_security_headers

Configure security headers.

**Optional Variables:**
- `vhost_file` (string): Virtual host file (for vhost scope)
- `config_scope` (string): "global" or "vhost" (default: "global")
- `x_frame_options` (string): X-Frame-Options header value (default: "SAMEORIGIN")
- `x_content_type_options` (string): X-Content-Type-Options header (default: "nosniff")
- `x_xss_protection` (string): X-XSS-Protection header (default: "1; mode=block")
- `strict_transport_security` (string): HSTS header value
- `content_security_policy` (string): CSP header value

**Example:**
```yaml
- name: Configure security headers
  module: apache
  vars:
    command: set_security_headers
    config_scope: global
    x_frame_options: "DENY"
    x_content_type_options: "nosniff"
    x_xss_protection: "1; mode=block"
    strict_transport_security: "max-age=31536000; includeSubDomains"
```

---

## Common Variable Patterns

### Boolean Values
Use `true` or `false` (lowercase) for boolean variables:
```yaml
reload_service: true
graceful: false
```

### String Values
Quote strings that might be interpreted as other types:
```yaml
port: "80"
timeout: "300"
keep_alive: "On"
```

### Array/List Values
Use space-separated strings for multiple values:
```yaml
ssl_protocols: "TLSv1.2 TLSv1.3"
server_alias: "www.example.com blog.example.com"
```

### File Paths
Always use absolute paths:
```yaml
document_root: /var/www/example.com
ssl_certificate: /etc/ssl/certs/example.com.crt
```

---

## Status Values

All operations return a status in their output:

- **ok**: Operation completed successfully, no changes made
- **changed**: Operation completed successfully, system state was modified
- **failed**: Operation failed, see message for details

## Error Handling

When an operation fails:
1. Check the returned message for specific error details
2. Verify all required variables are provided
3. Test configuration syntax with `test_config`
4. Check Apache error logs
5. Verify file permissions on configuration files
6. Ensure the Apache service is running

## Platform Differences

Some operations behave differently on Debian/Ubuntu vs RHEL/CentOS:

- **Module/Site Management**: Debian uses `a2ensite/a2dissite`, RHEL uses manual symlinks
- **Configuration Paths**: `/etc/apache2` vs `/etc/httpd`
- **Service Names**: `apache2` vs `httpd`

The module automatically detects the platform and adjusts behavior accordingly.
