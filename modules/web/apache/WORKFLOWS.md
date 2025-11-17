# Apache Module - Common Workflows

This guide provides complete, production-ready workflows for common Apache HTTP Server scenarios.

## Table of Contents

1. [Basic Web Hosting Setup](#basic-web-hosting-setup)
2. [SSL/TLS Certificate Installation](#ssltls-certificate-installation)
3. [Multi-Site Hosting](#multi-site-hosting)
4. [Reverse Proxy Configuration](#reverse-proxy-configuration)
5. [Load Balancer Setup](#load-balancer-setup)
6. [WordPress Hosting](#wordpress-hosting)
7. [Security Hardening](#security-hardening)
8. [Performance Optimization](#performance-optimization)
9. [High-Availability Configuration](#high-availability-configuration)
10. [Migration Workflows](#migration-workflows)

---

## Basic Web Hosting Setup

Complete setup for hosting a simple website.

```yaml
name: Basic Web Hosting Stack
description: Setup Apache with a basic website

inventory:
  hosts:
    - web-01

defaults:
  apache_service: apache2
  apache_config_dir: /etc/apache2

run:
  - name: Create website virtual host
    module: apache
    hosts: web-01
    vars:
      command: create_vhost
      vhost_name: example.com
      document_root: /var/www/example.com
      server_name: example.com
      server_alias: www.example.com
      allow_override: All

  - name: Enable virtual host
    module: apache
    hosts: web-01
    vars:
      command: enable_vhost
      vhost_name: example.com
      reload_service: true

  - name: Enable required modules
    module: apache
    hosts: web-01
    vars:
      command: enable_module
      module_name: rewrite
      reload_service: true

  - name: Configure security headers
    module: apache
    hosts: web-01
    vars:
      command: set_security_headers
      config_scope: vhost
      vhost_file: example.com.conf
      x_frame_options: SAMEORIGIN
      x_content_type_options: nosniff

  - name: Test configuration
    module: apache
    hosts: web-01
    vars:
      command: test_config
      verbose: true

  - name: Restart Apache
    module: apache
    hosts: web-01
    vars:
      command: restart_service
      graceful: true
```

---

## SSL/TLS Certificate Installation

Complete SSL certificate installation workflow with Let's Encrypt or commercial certificates.

### Option 1: New SSL Virtual Host

```yaml
name: SSL Certificate Installation
description: Install SSL certificate for secure website

inventory:
  hosts:
    - web-01

run:
  - name: Enable SSL module
    module: apache
    hosts: web-01
    vars:
      command: enable_module
      module_name: ssl
      reload_service: false

  - name: Enable headers module
    module: apache
    hosts: web-01
    vars:
      command: enable_module
      module_name: headers
      reload_service: true

  - name: Create SSL virtual host
    module: apache
    hosts: web-01
    vars:
      command: create_ssl_vhost
      vhost_name: secure.example.com
      document_root: /var/www/secure
      server_name: secure.example.com
      ssl_certificate: /etc/letsencrypt/live/secure.example.com/fullchain.pem
      ssl_certificate_key: /etc/letsencrypt/live/secure.example.com/privkey.pem
      ssl_protocols: "TLSv1.2 TLSv1.3"
      ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384"
      hsts_enabled: true
      redirect_http_to_https: true

  - name: Enable SSL virtual host
    module: apache
    hosts: web-01
    vars:
      command: enable_vhost
      vhost_name: secure.example.com
      reload_service: true

  - name: Test SSL configuration
    module: apache
    hosts: web-01
    vars:
      command: test_ssl_config
      check_certificates: true
```

### Option 2: Update Existing Virtual Host

```yaml
name: Update SSL Certificate
description: Update SSL certificate for existing virtual host

inventory:
  hosts:
    - web-01

run:
  - name: Install updated SSL certificate
    module: apache
    hosts: web-01
    vars:
      command: install_ssl_certificate
      vhost_file: example.com-ssl.conf
      ssl_certificate: /etc/ssl/certs/example.com-2024.crt
      ssl_certificate_key: /etc/ssl/private/example.com-2024.key
      ssl_certificate_chain: /etc/ssl/certs/chain.crt

  - name: Test configuration
    module: apache
    hosts: web-01
    vars:
      command: test_config

  - name: Reload Apache
    module: apache
    hosts: web-01
    vars:
      command: reload_config
      test_first: true
```

---

## Multi-Site Hosting

Configure Apache to host multiple websites on a single server.

```yaml
name: Multi-Site Hosting Setup
description: Configure multiple websites on one Apache server

inventory:
  hosts:
    - web-server

defaults:
  document_root_base: /var/www

run:
  # Site 1: Blog
  - name: Create blog virtual host
    module: apache
    hosts: web-server
    vars:
      command: create_vhost
      vhost_name: blog.example.com
      document_root: /var/www/blog
      server_name: blog.example.com
      port: 80

  - name: Enable blog virtual host
    module: apache
    hosts: web-server
    vars:
      command: enable_vhost
      vhost_name: blog.example.com
      reload_service: false

  # Site 2: Shop
  - name: Create shop virtual host
    module: apache
    hosts: web-server
    vars:
      command: create_ssl_vhost
      vhost_name: shop.example.com
      document_root: /var/www/shop
      server_name: shop.example.com
      ssl_certificate: /etc/ssl/certs/shop.crt
      ssl_certificate_key: /etc/ssl/private/shop.key
      redirect_http_to_https: true

  - name: Enable shop virtual host
    module: apache
    hosts: web-server
    vars:
      command: enable_vhost
      vhost_name: shop.example.com
      reload_service: false

  # Site 3: API
  - name: Create API virtual host
    module: apache
    hosts: web-server
    vars:
      command: create_vhost
      vhost_name: api.example.com
      document_root: /var/www/api
      server_name: api.example.com
      port: 80

  - name: Add API directory block with restrictions
    module: apache
    hosts: web-server
    vars:
      command: add_directory_block
      directory_path: /var/www/api
      vhost_file: api.example.com.conf
      options: "None"
      allow_override: "None"
      require: "all granted"

  - name: Enable API virtual host
    module: apache
    hosts: web-server
    vars:
      command: enable_vhost
      vhost_name: api.example.com
      reload_service: false

  # Configure security for all sites
  - name: Configure global security headers
    module: apache
    hosts: web-server
    vars:
      command: set_security_headers
      config_scope: global
      x_frame_options: SAMEORIGIN
      x_content_type_options: nosniff
      x_xss_protection: "1; mode=block"

  # Final reload
  - name: Reload Apache configuration
    module: apache
    hosts: web-server
    vars:
      command: reload_config
      test_first: true

  - name: List all virtual hosts
    module: apache
    hosts: web-server
    vars:
      command: list_vhosts
      show_details: true
```

---

## Reverse Proxy Configuration

Configure Apache as a reverse proxy for backend applications.

### Single Backend Application

```yaml
name: Reverse Proxy Setup
description: Configure Apache as reverse proxy for Node.js app

inventory:
  hosts:
    - proxy-01

run:
  - name: Enable proxy modules
    module: apache
    hosts: proxy-01
    vars:
      command: enable_module
      module_name: proxy
      reload_service: false

  - name: Enable proxy_http module
    module: apache
    hosts: proxy-01
    vars:
      command: enable_module
      module_name: proxy_http
      reload_service: false

  - name: Enable proxy_wstunnel for WebSockets
    module: apache
    hosts: proxy-01
    vars:
      command: enable_module
      module_name: proxy_wstunnel
      reload_service: true

  - name: Create proxy virtual host
    module: apache
    hosts: proxy-01
    vars:
      command: create_proxy_vhost
      vhost_name: app.example.com
      proxy_target: http://localhost:3000
      server_name: app.example.com
      proxy_preserve_host: true
      proxy_timeout: 600

  - name: Enable proxy virtual host
    module: apache
    hosts: proxy-01
    vars:
      command: enable_vhost
      vhost_name: app.example.com
      reload_service: true
```

### Multiple Backend Applications

```yaml
name: Multi-Backend Reverse Proxy
description: Proxy multiple backend services

inventory:
  hosts:
    - proxy-server

run:
  - name: Enable proxy modules
    module: apache
    hosts: proxy-server
    vars:
      command: enable_module
      module_name: proxy
      reload_service: false

  - name: Enable proxy_http
    module: apache
    hosts: proxy-server
    vars:
      command: enable_module
      module_name: proxy_http
      reload_service: true

  # Backend 1: Node.js API
  - name: Create API proxy
    module: apache
    hosts: proxy-server
    vars:
      command: create_proxy_vhost
      vhost_name: api.example.com
      proxy_target: http://localhost:3000
      server_name: api.example.com

  - name: Enable API proxy
    module: apache
    hosts: proxy-server
    vars:
      command: enable_vhost
      vhost_name: api.example.com
      reload_service: false

  # Backend 2: Python Flask App
  - name: Create Flask app proxy
    module: apache
    hosts: proxy-server
    vars:
      command: create_proxy_vhost
      vhost_name: app.example.com
      proxy_target: http://localhost:5000
      server_name: app.example.com

  - name: Enable Flask proxy
    module: apache
    hosts: proxy-server
    vars:
      command: enable_vhost
      vhost_name: app.example.com
      reload_service: false

  # Backend 3: Java Application
  - name: Create Java app proxy
    module: apache
    hosts: proxy-server
    vars:
      command: create_proxy_vhost
      vhost_name: java.example.com
      proxy_target: http://localhost:8080
      server_name: java.example.com

  - name: Enable Java proxy
    module: apache
    hosts: proxy-server
    vars:
      command: enable_vhost
      vhost_name: java.example.com
      reload_service: true
```

---

## Load Balancer Setup

Configure Apache as a load balancer for multiple backend servers.

```yaml
name: Load Balancer Configuration
description: Configure Apache with mod_proxy_balancer

inventory:
  hosts:
    - loadbalancer-01

run:
  - name: Enable proxy module
    module: apache
    hosts: loadbalancer-01
    vars:
      command: enable_module
      module_name: proxy
      reload_service: false

  - name: Enable proxy_http module
    module: apache
    hosts: loadbalancer-01
    vars:
      command: enable_module
      module_name: proxy_http
      reload_service: false

  - name: Enable proxy_balancer module
    module: apache
    hosts: loadbalancer-01
    vars:
      command: enable_module
      module_name: proxy_balancer
      reload_service: false

  - name: Enable lbmethod_byrequests
    module: apache
    hosts: loadbalancer-01
    vars:
      command: enable_module
      module_name: lbmethod_byrequests
      reload_service: true

  # Note: The load balancer configuration requires custom configuration
  # This would be done via create_vhost with custom_config parameter
  - name: Create load balancer virtual host
    module: apache
    hosts: loadbalancer-01
    vars:
      command: create_vhost
      vhost_name: lb.example.com
      document_root: /var/www/html
      server_name: lb.example.com
      custom_config: |
        <Proxy balancer://mycluster>
            BalancerMember http://backend1.example.com:8080
            BalancerMember http://backend2.example.com:8080
            BalancerMember http://backend3.example.com:8080
            ProxySet lbmethod=byrequests
        </Proxy>

        ProxyPass / balancer://mycluster/
        ProxyPassReverse / balancer://mycluster/

  - name: Enable load balancer
    module: apache
    hosts: loadbalancer-01
    vars:
      command: enable_vhost
      vhost_name: lb.example.com
      reload_service: true
```

---

## WordPress Hosting

Complete WordPress hosting setup with optimal Apache configuration.

```yaml
name: WordPress Hosting Setup
description: Configure Apache for WordPress with performance and security

inventory:
  hosts:
    - wordpress-server

run:
  - name: Enable required modules
    module: apache
    hosts: wordpress-server
    vars:
      command: enable_module
      module_name: rewrite
      reload_service: false

  - name: Enable headers module
    module: apache
    hosts: wordpress-server
    vars:
      command: enable_module
      module_name: headers
      reload_service: false

  - name: Enable expires module
    module: apache
    hosts: wordpress-server
    vars:
      command: enable_module
      module_name: expires
      reload_service: true

  - name: Create WordPress virtual host
    module: apache
    hosts: wordpress-server
    vars:
      command: create_vhost
      vhost_name: myblog.com
      document_root: /var/www/wordpress
      server_name: myblog.com
      server_alias: www.myblog.com
      allow_override: All
      custom_config: |
        # WordPress pretty permalinks
        <Directory /var/www/wordpress>
            <IfModule mod_rewrite.c>
                RewriteEngine On
                RewriteBase /
                RewriteRule ^index\.php$ - [L]
                RewriteCond %{REQUEST_FILENAME} !-f
                RewriteCond %{REQUEST_FILENAME} !-d
                RewriteRule . /index.php [L]
            </IfModule>
        </Directory>

        # Browser caching
        <IfModule mod_expires.c>
            ExpiresActive On
            ExpiresByType image/jpg "access plus 1 year"
            ExpiresByType image/jpeg "access plus 1 year"
            ExpiresByType image/gif "access plus 1 year"
            ExpiresByType image/png "access plus 1 year"
            ExpiresByType text/css "access plus 1 month"
            ExpiresByType application/javascript "access plus 1 month"
            ExpiresByType application/pdf "access plus 1 month"
        </IfModule>

  - name: Enable WordPress site
    module: apache
    hosts: wordpress-server
    vars:
      command: enable_vhost
      vhost_name: myblog.com
      reload_service: false

  - name: Configure PHP memory and upload limits
    module: apache
    hosts: wordpress-server
    vars:
      command: add_directory_block
      directory_path: /var/www/wordpress
      vhost_file: myblog.com.conf
      custom_directives: |
        php_value upload_max_filesize 64M
        php_value post_max_size 64M
        php_value memory_limit 256M
        php_value max_execution_time 300

  - name: Configure security headers
    module: apache
    hosts: wordpress-server
    vars:
      command: set_security_headers
      config_scope: vhost
      vhost_file: myblog.com.conf
      x_frame_options: SAMEORIGIN
      x_content_type_options: nosniff
      x_xss_protection: "1; mode=block"

  - name: Reload Apache
    module: apache
    hosts: wordpress-server
    vars:
      command: reload_config
      test_first: true
```

---

## Security Hardening

Complete Apache security hardening workflow.

```yaml
name: Apache Security Hardening
description: Implement comprehensive Apache security measures

inventory:
  hosts:
    - web-server

run:
  # Enable security modules
  - name: Enable SSL module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: ssl
      reload_service: false

  - name: Enable headers module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: headers
      reload_service: true

  # Configure SSL/TLS security
  - name: Set secure SSL protocols
    module: apache
    hosts: web-server
    vars:
      command: set_ssl_protocols
      ssl_protocols: "TLSv1.2 TLSv1.3"
      config_scope: global

  - name: Set secure SSL ciphers
    module: apache
    hosts: web-server
    vars:
      command: set_ssl_ciphers
      ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"
      ssl_prefer_server_ciphers: "Off"
      config_scope: global

  # Configure security headers
  - name: Set comprehensive security headers
    module: apache
    hosts: web-server
    vars:
      command: set_security_headers
      config_scope: global
      x_frame_options: DENY
      x_content_type_options: nosniff
      x_xss_protection: "1; mode=block"
      strict_transport_security: "max-age=31536000; includeSubDomains; preload"
      content_security_policy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"

  # Disable unnecessary modules
  - name: Disable autoindex module
    module: apache
    hosts: web-server
    vars:
      command: disable_module
      module_name: autoindex
      reload_service: false

  - name: Disable status module
    module: apache
    hosts: web-server
    vars:
      command: disable_module
      module_name: status
      reload_service: false

  # Configure logging for security monitoring
  - name: Set log level for security events
    module: apache
    hosts: web-server
    vars:
      command: set_log_level
      log_level: notice
      config_scope: global

  # Reload configuration
  - name: Reload Apache with security settings
    module: apache
    hosts: web-server
    vars:
      command: reload_config
      test_first: true
```

---

## Performance Optimization

Optimize Apache for high-traffic scenarios.

```yaml
name: Apache Performance Optimization
description: Configure Apache for optimal performance

inventory:
  hosts:
    - web-server

run:
  # Configure Event MPM for best performance
  - name: Disable prefork MPM
    module: apache
    hosts: web-server
    vars:
      command: disable_module
      module_name: mpm_prefork
      reload_service: false

  - name: Configure Event MPM
    module: apache
    hosts: web-server
    vars:
      command: set_mpm_config
      mpm_module: event
      start_servers: 3
      min_spare_threads: 75
      max_spare_threads: 250
      threads_per_child: 25
      max_request_workers: 400
      max_connections_per_child: 10000

  # Configure KeepAlive
  - name: Enable KeepAlive
    module: apache
    hosts: web-server
    vars:
      command: set_keepalive
      keep_alive: "On"
      keep_alive_timeout: 5
      max_keep_alive_requests: 100

  # Set optimized timeouts
  - name: Configure timeouts
    module: apache
    hosts: web-server
    vars:
      command: set_timeout
      timeout: 60
      request_read_timeout: 20

  # Enable compression
  - name: Enable deflate module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: deflate
      reload_service: false

  # Enable caching
  - name: Enable cache module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: cache
      reload_service: false

  - name: Enable cache_disk module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: cache_disk
      reload_service: false

  # Enable expires for browser caching
  - name: Enable expires module
    module: apache
    hosts: web-server
    vars:
      command: enable_module
      module_name: expires
      reload_service: true

  # Test and reload
  - name: Test performance configuration
    module: apache
    hosts: web-server
    vars:
      command: test_config
      verbose: true

  - name: Restart Apache with new configuration
    module: apache
    hosts: web-server
    vars:
      command: restart_service
      graceful: true
```

---

## High-Availability Configuration

Configure Apache for high availability with failover.

```yaml
name: High-Availability Apache Setup
description: Configure Apache cluster with shared configuration

inventory:
  groups:
    web_cluster:
      - web-01
      - web-02
      - web-03

defaults:
  apache_service: apache2

run:
  # Configure all nodes identically
  - name: Enable required modules on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: enable_module
      module_name: ssl
      reload_service: false

  - name: Enable headers module
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: enable_module
      module_name: headers
      reload_service: true

  # Create identical virtual host on all nodes
  - name: Create application virtual host
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: create_ssl_vhost
      vhost_name: app.example.com
      document_root: /var/www/app
      server_name: app.example.com
      ssl_certificate: /etc/ssl/certs/app.example.com.crt
      ssl_certificate_key: /etc/ssl/private/app.example.com.key
      hsts_enabled: true

  - name: Enable virtual host on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: enable_vhost
      vhost_name: app.example.com
      reload_service: false

  # Configure performance settings
  - name: Configure Event MPM on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: set_mpm_config
      mpm_module: event
      start_servers: 2
      min_spare_threads: 50
      max_spare_threads: 150
      threads_per_child: 25
      max_request_workers: 300

  # Configure KeepAlive
  - name: Configure KeepAlive on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: set_keepalive
      keep_alive: "On"
      keep_alive_timeout: 5

  # Test and reload all nodes
  - name: Test configuration on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: parallel
    vars:
      command: test_config

  - name: Reload Apache on all nodes
    module: apache
    hosts: @group:web_cluster
    strategy: serial  # One at a time for zero-downtime
    vars:
      command: reload_config
      test_first: true
```

---

## Migration Workflows

### Migrate from Old Server to New Server

```yaml
name: Apache Migration
description: Migrate Apache configuration to new server

inventory:
  hosts:
    - new-web-server

run:
  # Step 1: Install and configure Apache
  - name: Create main site virtual host
    module: apache
    hosts: new-web-server
    vars:
      command: create_vhost
      vhost_name: example.com
      document_root: /var/www/example.com
      server_name: example.com
      server_alias: www.example.com
      allow_override: All

  - name: Enable main site
    module: apache
    hosts: new-web-server
    vars:
      command: enable_vhost
      vhost_name: example.com
      reload_service: false

  # Step 2: Install SSL certificate
  - name: Install SSL certificate
    module: apache
    hosts: new-web-server
    vars:
      command: create_ssl_vhost
      vhost_name: example.com
      document_root: /var/www/example.com
      server_name: example.com
      ssl_certificate: /etc/ssl/certs/example.com.crt
      ssl_certificate_key: /etc/ssl/private/example.com.key
      redirect_http_to_https: true

  - name: Enable SSL site
    module: apache
    hosts: new-web-server
    vars:
      command: enable_vhost
      vhost_name: example.com
      reload_service: false

  # Step 3: Enable required modules
  - name: Enable rewrite module
    module: apache
    hosts: new-web-server
    vars:
      command: enable_module
      module_name: rewrite
      reload_service: false

  - name: Enable headers module
    module: apache
    hosts: new-web-server
    vars:
      command: enable_module
      module_name: headers
      reload_service: true

  # Step 4: Configure performance (match old server)
  - name: Configure MPM settings
    module: apache
    hosts: new-web-server
    vars:
      command: set_mpm_config
      mpm_module: event
      start_servers: 3
      max_request_workers: 300

  # Step 5: Test everything
  - name: Test final configuration
    module: apache
    hosts: new-web-server
    vars:
      command: test_config
      verbose: true
      show_vhosts: true

  - name: Start Apache service
    module: apache
    hosts: new-web-server
    vars:
      command: start_service
      test_config_first: true
```

---

## Troubleshooting Workflows

### Diagnose Configuration Issues

```yaml
name: Apache Configuration Diagnostics
description: Comprehensive Apache health check

inventory:
  hosts:
    - web-server

run:
  - name: Check service status
    module: apache
    hosts: web-server
    vars:
      command: check_service_status
      verbose: true

  - name: Test configuration syntax
    module: apache
    hosts: web-server
    vars:
      command: test_config
      verbose: true
      show_vhosts: true

  - name: List enabled virtual hosts
    module: apache
    hosts: web-server
    vars:
      command: list_vhosts
      show_details: true

  - name: List loaded modules
    module: apache
    hosts: web-server
    vars:
      command: list_modules
      format: json

  - name: Test SSL configuration
    module: apache
    hosts: web-server
    vars:
      command: test_ssl_config
      check_certificates: true
```

---

## Best Practices Summary

1. **Always test before reload**: Use `test_first: true` or explicit `test_config` steps
2. **Graceful restarts**: Use `graceful: true` for restarts to avoid dropping connections
3. **Parallel execution**: Use `strategy: parallel` for multi-node operations
4. **Security first**: Enable security headers and configure SSL properly
5. **Performance monitoring**: After optimization, monitor metrics to validate improvements
6. **Backup configurations**: Always backup before making changes
7. **Documentation**: Document custom configurations in the stack file

## Additional Resources

- [Apache Performance Tuning Guide](https://httpd.apache.org/docs/2.4/misc/perf-tuning.html)
- [Apache Security Tips](https://httpd.apache.org/docs/2.4/misc/security_tips.html)
- [SSL Configuration Best Practices](https://ssl-config.mozilla.org/)
