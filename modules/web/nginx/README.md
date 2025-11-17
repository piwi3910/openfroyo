# Nginx Web Server Management Module

Comprehensive Nginx web server management module for OpenFroyo that provides 35+ operations covering server blocks, upstreams, SSL/TLS, location blocks, configuration management, service management, and performance tuning.

## Overview

This module enables complete control over Nginx web server configuration and management through OpenFroyo stacks. It supports:

- **Server Block Management** - Create, configure, enable, and manage virtual hosts
- **Upstream and Load Balancing** - Configure backend servers and load balancing strategies
- **SSL/TLS Configuration** - Set up HTTPS with modern security settings
- **Location Blocks** - Define URL routing, proxying, and static file serving
- **Configuration Management** - Modify server settings and test configurations
- **Service Management** - Control Nginx service lifecycle
- **Performance Tuning** - Optimize worker processes and connections

## Nginx Architecture Concepts

### Server Blocks (Virtual Hosts)

Server blocks are Nginx's equivalent to Apache's virtual hosts. Each server block defines how to handle requests for specific domain names or IP addresses.

```nginx
server {
    listen 80;
    server_name example.com www.example.com;
    root /var/www/example.com;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

**Key Concepts:**
- **listen**: Port and protocol (80 for HTTP, 443 for HTTPS)
- **server_name**: Domain names this server block handles
- **root**: Document root directory for files
- **index**: Default files to serve for directories

### Upstreams (Backend Servers)

Upstreams define groups of backend servers for load balancing and proxying.

```nginx
upstream backend {
    least_conn;
    server 192.168.1.10:8080;
    server 192.168.1.11:8080;
    server 192.168.1.12:8080;
}

server {
    location / {
        proxy_pass http://backend;
    }
}
```

**Load Balancing Methods:**
- **round_robin** (default): Requests distributed evenly
- **least_conn**: Send to server with fewest active connections
- **ip_hash**: Client IP determines which server (sticky sessions)
- **hash**: Custom key-based distribution
- **random**: Random selection

### Location Blocks (URL Routing)

Location blocks define how to process requests for specific URL paths.

```nginx
location / {
    # Static files
    try_files $uri $uri/ =404;
}

location ~ \.php$ {
    # PHP files via FastCGI
    fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
}

location /api/ {
    # Proxy to backend
    proxy_pass http://backend;
}
```

**Location Modifiers:**
- `=` : Exact match
- `~` : Case-sensitive regex
- `~*` : Case-insensitive regex
- `^~` : Prefix match (stops regex checking)
- (none) : Prefix match

### SSL/TLS Configuration

Modern SSL/TLS setup with security best practices.

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # HSTS
    add_header Strict-Transport-Security "max-age=31536000" always;
}
```

### Configuration Directory Structure

**Debian/Ubuntu:**
```
/etc/nginx/
├── nginx.conf              # Main configuration
├── sites-available/        # All available server blocks
│   ├── default
│   ├── example.com.conf
│   └── api.example.com.conf
├── sites-enabled/          # Enabled server blocks (symlinks)
│   ├── default -> ../sites-available/default
│   └── example.com.conf -> ../sites-available/example.com.conf
└── conf.d/                 # Additional configs (upstreams, etc.)
    └── upstream-backend.conf
```

**RHEL/CentOS:**
```
/etc/nginx/
├── nginx.conf              # Main configuration
└── conf.d/                 # All server blocks and configs
    ├── example.com.conf
    ├── api.example.com.conf
    └── upstream-backend.conf
```

## Available Operations

### Server Block Management (6 commands)

| Command | Description |
|---------|-------------|
| `create_server_block` | Create a new server block configuration |
| `delete_server_block` | Remove a server block configuration |
| `enable_server_block` | Enable a server block (create symlink) |
| `disable_server_block` | Disable a server block (remove symlink) |
| `list_server_blocks` | List all available and enabled server blocks |
| `test_server_config` | Test server block configuration validity |

### Upstream and Load Balancing (6 commands)

| Command | Description |
|---------|-------------|
| `create_upstream` | Create a new upstream block with backend servers |
| `delete_upstream` | Remove an upstream configuration |
| `add_upstream_server` | Add a server to an existing upstream |
| `remove_upstream_server` | Remove a server from an upstream |
| `set_load_balance_method` | Change load balancing algorithm |
| `show_upstreams` | List all configured upstreams |

### SSL/TLS Configuration (6 commands)

| Command | Description |
|---------|-------------|
| `create_ssl_server` | Create an HTTPS server block with SSL config |
| `install_ssl_certificate` | Verify SSL certificate installation |
| `set_ssl_protocols` | Configure allowed SSL/TLS protocol versions |
| `set_ssl_ciphers` | Set SSL cipher suites |
| `enable_http2` | Enable HTTP/2 protocol support |
| `test_ssl_config` | Test SSL configuration validity |

### Location Blocks (5 commands)

| Command | Description |
|---------|-------------|
| `add_location_block` | Add a location block to server configuration |
| `add_proxy_pass` | Add reverse proxy location with proxy_pass |
| `add_fastcgi_pass` | Add PHP/FastCGI location block |
| `add_static_location` | Add static file serving location |
| `add_rewrite_rule` | Add URL rewrite rule |

### Configuration Management (6 commands)

| Command | Description |
|---------|-------------|
| `set_server_name` | Update server_name directive |
| `set_root` | Update document root directory |
| `set_index` | Update default index files |
| `add_header` | Add custom HTTP response header |
| `test_config` | Test entire Nginx configuration |
| `reload_config` | Reload Nginx configuration (graceful) |

### Service Management (4 commands)

| Command | Description |
|---------|-------------|
| `start_service` | Start Nginx service |
| `stop_service` | Stop Nginx service |
| `restart_service` | Restart Nginx service |
| `check_service_status` | Check if Nginx is running |

### Performance Tuning (2 commands)

| Command | Description |
|---------|-------------|
| `set_worker_processes` | Set number of worker processes |
| `set_worker_connections` | Set maximum connections per worker |

## Quick Start Examples

### Create a Basic Web Server

```yaml
# stacks/simple-web.ofy
name: Simple Web Server
inventory: "@group:webservers"

run:
  - name: Create example.com server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: example.com www.example.com
      server_file: example.com.conf
      root: /var/www/example.com
      index: index.html index.htm

  - name: Enable the server block
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: example.com.conf
```

### Set Up HTTPS with SSL

```yaml
# stacks/ssl-web.ofy
name: SSL Web Server
inventory: "@group:webservers"

run:
  - name: Create SSL server block
    module: web/nginx
    vars:
      command: create_ssl_server
      server_name: secure.example.com
      server_file: secure.example.com.conf
      ssl_certificate: /etc/ssl/certs/example.com.crt
      ssl_certificate_key: /etc/ssl/private/example.com.key
      ssl_protocols: "TLSv1.2 TLSv1.3"
      http2: "true"
      root: /var/www/secure

  - name: Enable SSL server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: secure.example.com.conf
```

### Configure Load Balancer

```yaml
# stacks/load-balancer.ofy
name: Load Balancer Setup
inventory: "@group:loadbalancers"

run:
  - name: Create backend upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: appservers
      upstream_servers:
        - "192.168.1.10:8080"
        - "192.168.1.11:8080"
        - "192.168.1.12:8080"
      load_balance_method: least_conn

  - name: Create proxy server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: app.example.com
      server_file: app.example.com.conf

  - name: Add proxy location
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: app.example.com.conf
      location_path: /
      proxy_pass: http://appservers

  - name: Enable proxy server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: app.example.com.conf
```

## Configuration Variables

See [defaults.ofy.yml](defaults.ofy.yml) for all available configuration variables with defaults and descriptions.

### Common Variables

```yaml
# Service configuration
nginx_config_dir: /etc/nginx
nginx_sites_available: /etc/nginx/sites-available
nginx_sites_enabled: /etc/nginx/sites-enabled

# Server block
server_name: example.com
server_file: example.com.conf
listen_port: 80
root: /var/www/example.com
index: index.html index.htm

# SSL/TLS
ssl_certificate: /etc/ssl/certs/example.com.crt
ssl_certificate_key: /etc/ssl/private/example.com.key
ssl_protocols: "TLSv1.2 TLSv1.3"
http2: false

# Upstream
upstream_name: backend
upstream_servers: []
load_balance_method: round_robin

# Performance
worker_processes: auto
worker_connections: 1024
```

## Security Best Practices

### SSL/TLS Security

1. **Use Modern Protocols Only**
   ```yaml
   ssl_protocols: "TLSv1.2 TLSv1.3"  # Disable TLSv1.0 and TLSv1.1
   ```

2. **Strong Cipher Suites**
   ```yaml
   ssl_ciphers: "HIGH:!aNULL:!MD5"
   ssl_prefer_server_ciphers: "on"
   ```

3. **Enable HTTP/2**
   ```yaml
   http2: "true"
   ```

4. **Enable HSTS**
   ```yaml
   hsts_enabled: "true"
   hsts_max_age: 31536000  # 1 year
   ```

### HTTP Security Headers

Add security headers to protect against common attacks:

```yaml
- name: Add security headers
  module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: X-Frame-Options
    header_value: SAMEORIGIN

- name: Add XSS protection
  module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: X-XSS-Protection
    header_value: "1; mode=block"
```

### Configuration Testing

Always test configuration before reloading:

```yaml
- name: Test configuration
  module: web/nginx
  vars:
    command: test_config
```

The module automatically tests configuration before reload operations.

## Performance Tuning

### Worker Optimization

Adjust worker processes and connections based on server resources:

```yaml
# Set workers to number of CPU cores
- name: Optimize worker processes
  module: web/nginx
  vars:
    command: set_worker_processes
    worker_processes: auto  # or specific number like "4"

# Increase connections per worker
- name: Optimize worker connections
  module: web/nginx
  vars:
    command: set_worker_connections
    worker_connections: 2048
```

**Guidelines:**
- `worker_processes`: Set to number of CPU cores or use "auto"
- `worker_connections`: 1024-4096 typical, depends on available file descriptors
- Total connections = `worker_processes * worker_connections`

### Caching and Compression

Configure in `defaults.ofy.yml`:

```yaml
# Gzip compression
gzip_enabled: true
gzip_comp_level: 6
gzip_types:
  - text/plain
  - text/css
  - application/json
  - application/javascript

# File caching
open_file_cache_max: 1000
open_file_cache_valid: 30s
```

### Keepalive Connections

```yaml
keepalive_timeout: 65
keepalive_requests: 100
upstream_keepalive: 32  # For upstream connections
```

## Common Patterns

### Static Site with CDN

```yaml
- name: Create static site
  module: web/nginx
  vars:
    command: create_server_block
    server_name: static.example.com
    root: /var/www/static

- name: Add cache headers
  module: web/nginx
  vars:
    command: add_header
    server_file: static.example.com.conf
    header_name: Cache-Control
    header_value: "public, max-age=31536000"
```

### Reverse Proxy with Caching

```yaml
- name: Create caching proxy
  module: web/nginx
  vars:
    command: add_location_block
    server_file: proxy.conf
    location_path: /
    proxy_pass: http://backend
    additional_config: |
      proxy_cache_valid 200 1h;
      proxy_cache_bypass $http_pragma;
```

### PHP Application (WordPress, Laravel, etc.)

```yaml
- name: Add PHP processing
  module: web/nginx
  vars:
    command: add_fastcgi_pass
    server_file: wordpress.conf
    location_path: ~ \.php$
    location_modifier: ~
    fastcgi_pass: unix:/var/run/php/php7.4-fpm.sock
```

### WebSocket Proxy

```yaml
- name: WebSocket proxy location
  module: web/nginx
  vars:
    command: add_location_block
    server_file: websocket.conf
    location_path: /ws
    proxy_pass: http://websocket_backend
    additional_config: |
      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
```

## Troubleshooting

### Configuration Test Failed

If `nginx -t` fails:

1. Check syntax errors in generated configs
2. Verify file paths exist (root, ssl certificates)
3. Check for duplicate server_name declarations
4. Validate upstream server addresses

### Service Won't Start

Common issues:

1. **Port already in use**: Check if another process is using port 80/443
   ```bash
   netstat -tlnp | grep :80
   ```

2. **Permission issues**: Ensure Nginx has read access to root directories

3. **Missing dependencies**: Verify SSL certificates exist

### Proxy Connection Issues

For proxy_pass problems:

1. Verify backend servers are reachable
2. Check upstream health
3. Review proxy timeout settings
4. Examine error logs: `/var/log/nginx/error.log`

## Additional Resources

- [COMMANDS.md](COMMANDS.md) - Detailed command reference
- [WORKFLOWS.md](WORKFLOWS.md) - Complete workflow examples
- [Nginx Official Documentation](https://nginx.org/en/docs/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)

## Module Structure

```
modules/web/nginx/
├── module.ofy.yml      # Module definition
├── defaults.ofy.yml    # Default variables
├── wasm/
│   ├── main.go         # WASM implementation
│   ├── nginx.wasm      # Compiled WASM binary
│   └── Makefile        # Build configuration
├── test.ofy            # Test scenarios
├── README.md           # This file
├── COMMANDS.md         # Command reference
└── WORKFLOWS.md        # Workflow examples
```

## License

Part of the OpenFroyo project.
