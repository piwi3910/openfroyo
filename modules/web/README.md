# OpenFroyo Web Server Modules

This directory contains modules for managing web servers and reverse proxies.

## Available Modules (2 Total)

### 1. apache (Apache HTTP Server)
**Purpose:** Complete Apache HTTP Server (httpd) management

**Key Features:**
- 34 operations across 8 categories
- Virtual host lifecycle management
- Apache module management (ssl, rewrite, headers, proxy)
- SSL/TLS certificate installation
- Directory and location configuration
- Service lifecycle (start, stop, restart, reload)
- Performance tuning (MPM, KeepAlive, timeouts)
- Security headers configuration

**Compatibility:** Apache 2.4+ (Debian/Ubuntu, RHEL/CentOS)

**Size:** 3.5MB

**Platform Support:** Cross-platform (Debian/Ubuntu and RHEL/CentOS)

---

### 2. nginx (Nginx Web Server)
**Purpose:** Complete Nginx web server and reverse proxy management

**Key Features:**
- 35+ operations across 7 categories
- Server block management
- Upstream load balancing (round_robin, least_conn, ip_hash)
- SSL/TLS with HTTP/2
- Location blocks and routing
- Reverse proxy configuration
- FastCGI/PHP-FPM integration
- Performance optimization (worker processes, connections)

**Compatibility:** Nginx 1.18+ (Debian/Ubuntu, RHEL/CentOS)

**Size:** 3.8MB

**Platform Support:** Cross-platform (Debian/Ubuntu and RHEL/CentOS)

---

## Comparison Matrix

| Feature | Apache | Nginx |
|---------|--------|-------|
| **Configuration** | Hierarchical | Block-based |
| Virtual Hosts/Server Blocks | ✅ (6 ops) | ✅ (6 ops) |
| Module Management | ✅ (5 ops) | ❌ (compiled-in) |
| SSL/TLS | ✅ (5 ops) | ✅ (6 ops) |
| HTTP/2 | ✅ (via module) | ✅ (1 op) |
| Load Balancing | ✅ (mod_proxy_balancer) | ✅ (6 ops, 5 algorithms) |
| Reverse Proxy | ✅ | ✅ (built-in) |
| FastCGI/PHP | ✅ (mod_fcgid) | ✅ (native) |
| Directory Blocks | ✅ (6 ops) | ✅ (via location) |
| Rewrite Rules | ✅ (mod_rewrite) | ✅ (native) |
| Performance Tuning | ✅ MPM (4 ops) | ✅ Workers (2 ops) |
| Service Management | ✅ (4 ops) | ✅ (4 ops) |
| Config Testing | ✅ | ✅ |
| Hot Reload | ✅ | ✅ |

## When to Use Each Module

### Use **apache** when:
- You need .htaccess support
- Running legacy applications
- Require mod_rewrite complexity
- Need extensive module ecosystem
- Want directory-level configuration
- Running shared hosting environments
- Prefer process-based architecture

### Use **nginx** when:
- Building high-performance sites
- Need efficient reverse proxy
- Serving static content at scale
- Want built-in load balancing
- Need WebSocket support
- Prefer event-driven architecture
- Building microservices gateway

## Common Usage Patterns

### Apache Virtual Host Setup

```yaml
# Create basic virtual host
- module: web/apache
  vars:
    apache_service: apache2
    command: create_vhost
    vhost_name: example.com
    server_name: example.com
    server_alias: www.example.com
    document_root: /var/www/example.com
    port: 80

# Enable SSL with HTTPS redirect
- module: web/apache
  vars:
    command: create_ssl_vhost
    vhost_name: secure.example.com
    server_name: secure.example.com
    document_root: /var/www/secure
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    hsts_enabled: true
    redirect_http_to_https: true

# Enable modules
- module: web/apache
  vars:
    command: enable_module
    module_name: rewrite

- module: web/apache
  vars:
    command: reload_config
```

### Nginx Server Block Setup

```yaml
# Create server block
- module: web/nginx
  vars:
    command: create_server_block
    server_name: example.com
    server_file: example.com.conf
    listen_port: 80
    root: /var/www/example.com
    index: index.html

# Enable SSL with HTTP/2
- module: web/nginx
  vars:
    command: create_ssl_server
    server_name: secure.example.com
    listen_port: 443
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    ssl_protocols: "TLSv1.2 TLSv1.3"
    http2: true

# Enable and reload
- module: web/nginx
  vars:
    command: enable_server_block
    server_file: secure.example.com.conf

- module: web/nginx
  vars:
    command: reload_config
```

## Advanced Scenarios

### Apache Reverse Proxy

```yaml
- module: web/apache
  vars:
    command: enable_module
    module_name: proxy

- module: web/apache
  vars:
    command: enable_module
    module_name: proxy_http

- module: web/apache
  vars:
    command: create_proxy_vhost
    vhost_name: app.example.com
    server_name: app.example.com
    proxy_pass: http://localhost:8080
    preserve_host: true
```

### Nginx Load Balancer

```yaml
# Create upstream with backend servers
- module: web/nginx
  vars:
    command: create_upstream
    upstream_name: backend_pool
    upstream_servers:
      - "192.168.1.10:8080"
      - "192.168.1.11:8080"
      - "192.168.1.12:8080"
    load_balance_method: least_conn

# Create server block with proxy
- module: web/nginx
  vars:
    command: create_server_block
    server_name: app.example.com
    listen_port: 80

- module: web/nginx
  vars:
    command: add_proxy_pass
    server_file: app.example.com.conf
    location_path: /
    proxy_pass: http://backend_pool
    proxy_headers:
      Host: $host
      X-Real-IP: $remote_addr
      X-Forwarded-For: $proxy_add_x_forwarded_for
```

### PHP Application (WordPress)

**Apache + mod_php:**
```yaml
- module: web/apache
  vars:
    command: create_vhost
    vhost_name: wordpress.example.com
    document_root: /var/www/wordpress
    allow_override: All

- module: web/apache
  vars:
    command: enable_module
    module_name: rewrite

- module: web/apache
  vars:
    command: enable_module
    module_name: php7.4
```

**Nginx + PHP-FPM:**
```yaml
- module: web/nginx
  vars:
    command: create_server_block
    server_name: wordpress.example.com
    root: /var/www/wordpress
    index: index.php

- module: web/nginx
  vars:
    command: add_fastcgi_pass
    server_file: wordpress.example.com.conf
    location_path: ~ \.php$
    fastcgi_pass: unix:/var/run/php/php7.4-fpm.sock
    fastcgi_index: index.php
```

## SSL/TLS Best Practices

### Apache SSL Configuration

```yaml
- module: web/apache
  vars:
    command: create_ssl_vhost
    vhost_name: secure.example.com
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    ssl_certificate_chain: /etc/ssl/certs/chain.crt
    ssl_protocols: "TLSv1.2 TLSv1.3"
    ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256"
    hsts_enabled: true
    hsts_max_age: 31536000
```

### Nginx SSL Configuration

```yaml
- module: web/nginx
  vars:
    command: create_ssl_server
    server_name: secure.example.com
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    ssl_protocols: "TLSv1.2 TLSv1.3"
    ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256"
    http2: true

- module: web/nginx
  vars:
    command: add_header
    server_file: secure.example.com.conf
    header_name: Strict-Transport-Security
    header_value: "max-age=31536000; includeSubDomains"
```

## Performance Tuning

### Apache Performance

```yaml
# Use Event MPM for better performance
- module: web/apache
  vars:
    command: set_mpm_config
    mpm_module: event
    start_servers: 3
    min_spare_threads: 75
    max_spare_threads: 250
    threads_per_child: 25
    max_request_workers: 400
    max_connections_per_child: 0

# Optimize KeepAlive
- module: web/apache
  vars:
    command: set_keepalive
    keep_alive: On
    max_keep_alive_requests: 100
    keep_alive_timeout: 5
```

### Nginx Performance

```yaml
# Optimize worker processes
- module: web/nginx
  vars:
    command: set_worker_processes
    worker_processes: auto

- module: web/nginx
  vars:
    command: set_worker_connections
    worker_connections: 2048
```

## Security Hardening

### Apache Security

```yaml
# Configure security headers
- module: web/apache
  vars:
    command: set_security_headers
    x_frame_options: DENY
    x_content_type_options: nosniff
    x_xss_protection: "1; mode=block"
    referrer_policy: strict-origin-when-cross-origin

# Set server tokens
- module: web/apache
  vars:
    command: set_server_tokens
    server_tokens: Prod
```

### Nginx Security

```yaml
# Add security headers
- module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: X-Frame-Options
    header_value: DENY

- module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: X-Content-Type-Options
    header_value: nosniff
```

## High Availability Setup

### Apache with HAProxy

```yaml
# Configure multiple Apache backends
# HAProxy handles load balancing
# Apache servers run on different ports or IPs
```

### Nginx Load Balancer

```yaml
# Nginx as load balancer in front of Apache/app servers
- module: web/nginx
  vars:
    command: create_upstream
    upstream_name: web_cluster
    upstream_servers:
      - "web1.internal:80"
      - "web2.internal:80"
      - "web3.internal:80"
    load_balance_method: least_conn
    health_check: true
```

## Monitoring and Logging

### Apache Monitoring

```yaml
# Enable server-status module
- module: web/apache
  vars:
    command: enable_module
    module_name: status

# Check service status
- module: web/apache
  vars:
    command: check_service_status
```

### Nginx Monitoring

```yaml
# Check service status
- module: web/nginx
  vars:
    command: check_service_status

# Review access logs
# tail -f /var/log/nginx/access.log

# Review error logs
# tail -f /var/log/nginx/error.log
```

## Troubleshooting

### Apache Issues

**Service Won't Start:**
- Test configuration: `command: test_config`
- Check error logs: `/var/log/apache2/error.log` or `/var/log/httpd/error_log`
- Verify port 80/443 not in use
- Check file permissions on DocumentRoot

**403 Forbidden:**
- Check Directory permissions
- Verify AllowOverride settings
- Check .htaccess files
- Review SELinux/AppArmor settings

### Nginx Issues

**Service Won't Start:**
- Test configuration: `command: test_config`
- Check error logs: `/var/log/nginx/error.log`
- Verify port 80/443 not in use
- Check file permissions on root directory

**502 Bad Gateway:**
- Verify upstream servers are running
- Check firewall rules to backend
- Review upstream configuration
- Check backend application logs

## Documentation

Each module includes comprehensive documentation:
- **README.md** - Complete usage guide
- **COMMANDS.md** - Detailed command reference
- **WORKFLOWS.md** - Real-world scenarios
- **test.ofy** - Example test stack

## Module Statistics

- **2 web server modules**
- **69+ total operations**
- **~7.3MB total WASM binaries** (source only, binaries in .gitignore)
- **Complete web stack** (traditional + modern architectures)
- **Production-ready** with comprehensive documentation
- **Cross-platform** (Debian/Ubuntu and RHEL/CentOS)

## Future Enhancements

Potential additions:
- HAProxy module
- Caddy module
- Traefik module
- Lighttpd module
- Apache Tomcat module
- IIS module (Windows)
- Let's Encrypt integration
- Web Application Firewall (ModSecurity)
