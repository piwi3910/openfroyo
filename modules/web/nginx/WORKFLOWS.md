# Nginx Module Workflows

Complete end-to-end workflow examples for common Nginx deployment scenarios.

## Table of Contents

1. [Static Website Hosting](#1-static-website-hosting)
2. [Reverse Proxy for Web Application](#2-reverse-proxy-for-web-application)
3. [Load Balancer Setup](#3-load-balancer-setup)
4. [SSL/TLS HTTPS Configuration](#4-ssltls-https-configuration)
5. [WordPress/PHP Application](#5-wordpressphp-application)
6. [API Gateway Pattern](#6-api-gateway-pattern)
7. [WebSocket Proxy](#7-websocket-proxy)
8. [Multi-Tenant Hosting](#8-multi-tenant-hosting)
9. [Blue-Green Deployment](#9-blue-green-deployment)
10. [CDN Origin Server](#10-cdn-origin-server)
11. [Microservices Routing](#11-microservices-routing)
12. [Development Environment](#12-development-environment)

---

## 1. Static Website Hosting

Deploy a simple static HTML website with caching and compression.

### Scenario
Host a static website at `www.example.com` from `/var/www/example.com` with optimal performance settings.

### Stack File: `stacks/static-website.ofy`

```yaml
name: Static Website Deployment
description: Deploy static HTML website with performance optimization
inventory: "@group:webservers"

defaults:
  domain: example.com
  doc_root: /var/www/example.com

run:
  # Create server block
  - name: Create static website server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }} www.{{ var.domain }}"
      server_file: "{{ var.domain }}.conf"
      root: "{{ var.doc_root }}"
      index: index.html

  # Add caching for static assets
  - name: Add static asset caching
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2)$
      location_modifier: ~*
      additional_config: |
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;

  # Add security headers
  - name: Add X-Frame-Options header
    module: web/nginx
    vars:
      command: add_header
      server_file: "{{ var.domain }}.conf"
      header_name: X-Frame-Options
      header_value: DENY

  - name: Add X-Content-Type-Options header
    module: web/nginx
    vars:
      command: add_header
      server_file: "{{ var.domain }}.conf"
      header_name: X-Content-Type-Options
      header_value: nosniff

  # Enable the site
  - name: Enable website
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "{{ var.domain }}.conf"

  # Test and reload
  - name: Test configuration
    module: web/nginx
    vars:
      command: test_config

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- Website accessible at http://www.example.com
- Static assets cached for 1 year
- Security headers enabled
- Zero downtime deployment

---

## 2. Reverse Proxy for Web Application

Set up Nginx as a reverse proxy for a Node.js/Python/Ruby application.

### Scenario
Proxy requests from `app.example.com` to backend application running on `localhost:3000`.

### Stack File: `stacks/reverse-proxy.ofy`

```yaml
name: Reverse Proxy Setup
description: Configure Nginx as reverse proxy for web application
inventory: "@group:appservers"

defaults:
  domain: app.example.com
  backend_port: 3000

run:
  # Create basic server block
  - name: Create proxy server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: "{{ var.domain }}.conf"

  # Add proxy configuration
  - name: Configure reverse proxy
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: "{{ var.domain }}.conf"
      location_path: /
      proxy_pass: "http://127.0.0.1:{{ var.backend_port }}"

  # Add WebSocket support
  - name: Add WebSocket proxy headers
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: /ws
      proxy_pass: "http://127.0.0.1:{{ var.backend_port }}"
      additional_config: |
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;

  # Add client max body size for file uploads
  - name: Configure upload size
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: /upload
      proxy_pass: "http://127.0.0.1:{{ var.backend_port }}"
      additional_config: |
        client_max_body_size 100M;
        proxy_request_buffering off;

  # Enable and reload
  - name: Enable proxy server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "{{ var.domain }}.conf"

  - name: Reload configuration
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- All traffic to app.example.com proxied to localhost:3000
- WebSocket support on /ws path
- Large file uploads supported on /upload
- Proper proxy headers set

---

## 3. Load Balancer Setup

Configure Nginx as a load balancer for multiple backend servers.

### Scenario
Distribute traffic across 3 application servers with health checks and session persistence.

### Stack File: `stacks/load-balancer.ofy`

```yaml
name: Load Balancer Configuration
description: Set up Nginx load balancer with health checks
inventory: "@group:loadbalancers"

defaults:
  domain: app.example.com
  backend_servers:
    - "10.0.1.10:8080"
    - "10.0.1.11:8080"
    - "10.0.1.12:8080"

run:
  # Create upstream with least connections method
  - name: Create backend upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: app_backend
      upstream_servers: "{{ var.backend_servers }}"
      load_balance_method: least_conn
      upstream_keepalive: 32

  # Create load balancer server block
  - name: Create load balancer server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: lb-{{ var.domain }}.conf

  # Configure proxy to upstream
  - name: Add proxy to upstream
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: lb-{{ var.domain }}.conf
      location_path: /
      proxy_pass: http://app_backend

  # Add health check endpoint
  - name: Add health check location
    module: web/nginx
    vars:
      command: add_location_block
      server_file: lb-{{ var.domain }}.conf
      location_path: /health
      location_modifier: "="
      additional_config: |
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;

  # Enable load balancer
  - name: Enable load balancer
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: lb-{{ var.domain }}.conf

  # Performance tuning
  - name: Optimize worker processes
    module: web/nginx
    vars:
      command: set_worker_processes
      worker_processes: auto

  - name: Increase worker connections
    module: web/nginx
    vars:
      command: set_worker_connections
      worker_connections: 2048

  - name: Reload configuration
    module: web/nginx
    vars:
      command: reload_config
```

### Adding/Removing Backend Servers

```yaml
# Add new server to pool
- name: Add new backend server
  module: web/nginx
  vars:
    command: add_upstream_server
    upstream_name: app_backend
    upstream_server: "10.0.1.13:8080"

# Remove failed server
- name: Remove failed backend
  module: web/nginx
  vars:
    command: remove_upstream_server
    upstream_name: app_backend
    upstream_server: "10.0.1.11:8080"
```

### Expected Result
- Traffic distributed across backends using least connections
- Persistent keepalive connections to backends
- Health check endpoint at /health
- Optimized for high concurrency

---

## 4. SSL/TLS HTTPS Configuration

Deploy a secure HTTPS website with modern TLS settings.

### Scenario
Create an HTTPS website with Let's Encrypt certificates, HTTP/2, and HSTS.

### Stack File: `stacks/https-website.ofy`

```yaml
name: HTTPS Website with SSL
description: Deploy secure HTTPS website with modern TLS
inventory: "@group:webservers"

defaults:
  domain: secure.example.com
  doc_root: /var/www/secure
  ssl_cert: /etc/letsencrypt/live/secure.example.com/fullchain.pem
  ssl_key: /etc/letsencrypt/live/secure.example.com/privkey.pem

run:
  # Verify SSL certificates exist
  - name: Verify SSL certificates
    module: web/nginx
    vars:
      command: install_ssl_certificate
      ssl_certificate: "{{ var.ssl_cert }}"
      ssl_certificate_key: "{{ var.ssl_key }}"

  # Create SSL server block
  - name: Create HTTPS server
    module: web/nginx
    vars:
      command: create_ssl_server
      server_name: "{{ var.domain }}"
      server_file: "{{ var.domain }}.conf"
      ssl_certificate: "{{ var.ssl_cert }}"
      ssl_certificate_key: "{{ var.ssl_key }}"
      ssl_protocols: "TLSv1.2 TLSv1.3"
      ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384"
      http2: "true"
      hsts_enabled: "true"
      hsts_max_age: "31536000"
      root: "{{ var.doc_root }}"

  # Add additional security headers
  - name: Add X-Frame-Options
    module: web/nginx
    vars:
      command: add_header
      server_file: "{{ var.domain }}.conf"
      header_name: X-Frame-Options
      header_value: DENY

  - name: Add Content-Security-Policy
    module: web/nginx
    vars:
      command: add_header
      server_file: "{{ var.domain }}.conf"
      header_name: Content-Security-Policy
      header_value: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"

  - name: Add Referrer-Policy
    module: web/nginx
    vars:
      command: add_header
      server_file: "{{ var.domain }}.conf"
      header_name: Referrer-Policy
      header_value: strict-origin-when-cross-origin

  # Enable and test
  - name: Enable HTTPS server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "{{ var.domain }}.conf"

  - name: Test SSL configuration
    module: web/nginx
    vars:
      command: test_ssl_config

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- HTTPS site with A+ SSL Labs rating
- HTTP automatically redirects to HTTPS
- HTTP/2 enabled for better performance
- HSTS with 1-year max age
- Modern security headers configured

---

## 5. WordPress/PHP Application

Deploy a WordPress site with Nginx and PHP-FPM.

### Scenario
Host WordPress at `blog.example.com` with PHP-FPM and proper security.

### Stack File: `stacks/wordpress.ofy`

```yaml
name: WordPress Deployment
description: Deploy WordPress with Nginx and PHP-FPM
inventory: "@group:webservers"

defaults:
  domain: blog.example.com
  doc_root: /var/www/wordpress
  php_socket: unix:/var/run/php/php7.4-fpm.sock

run:
  # Create WordPress server block
  - name: Create WordPress server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: "{{ var.domain }}.conf"
      root: "{{ var.doc_root }}"
      index: index.php index.html

  # Add PHP processing
  - name: Configure PHP-FPM
    module: web/nginx
    vars:
      command: add_fastcgi_pass
      server_file: "{{ var.domain }}.conf"
      location_path: ~ \.php$
      location_modifier: ~
      fastcgi_pass: "{{ var.php_socket }}"

  # WordPress permalink support
  - name: Add permalink rewrite
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: /
      try_files: $uri $uri/ /index.php?$args

  # Block access to sensitive files
  - name: Block .htaccess access
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: ~ /\.ht
      location_modifier: ~
      additional_config: |
        deny all;

  - name: Block wp-config.php access
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: = /wp-config.php
      location_modifier: "="
      additional_config: |
        deny all;

  # Static file caching
  - name: Cache static assets
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$
      location_modifier: ~*
      additional_config: |
        expires 1y;
        add_header Cache-Control "public, immutable";

  # Upload size limit
  - name: Set upload size limit
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "{{ var.domain }}.conf"
      location_path: /wp-admin/
      additional_config: |
        client_max_body_size 64M;

  # Enable WordPress site
  - name: Enable WordPress
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "{{ var.domain }}.conf"

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- WordPress accessible with clean URLs
- PHP files processed via PHP-FPM
- Sensitive files blocked
- 64MB upload limit
- Static assets cached

---

## 6. API Gateway Pattern

Configure Nginx as an API gateway routing to microservices.

### Scenario
Route API requests to different backend services based on URL path.

### Stack File: `stacks/api-gateway.ofy`

```yaml
name: API Gateway
description: Configure Nginx as API gateway for microservices
inventory: "@group:gateways"

defaults:
  domain: api.example.com

run:
  # Create upstreams for each service
  - name: Create auth service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: auth_service
      upstream_servers:
        - "10.0.2.10:8001"
        - "10.0.2.11:8001"
      load_balance_method: round_robin

  - name: Create user service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: user_service
      upstream_servers:
        - "10.0.2.20:8002"
        - "10.0.2.21:8002"
      load_balance_method: least_conn

  - name: Create product service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: product_service
      upstream_servers:
        - "10.0.2.30:8003"
        - "10.0.2.31:8003"
      load_balance_method: least_conn

  # Create API gateway server
  - name: Create API gateway server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: api-gateway.conf

  # Route /auth/* to auth service
  - name: Route auth endpoints
    module: web/nginx
    vars:
      command: add_location_block
      server_file: api-gateway.conf
      location_path: /api/v1/auth/
      proxy_pass: http://auth_service
      additional_config: |
        rewrite ^/api/v1/auth/(.*)$ /$1 break;
        proxy_set_header X-Service-Name "auth";

  # Route /users/* to user service
  - name: Route user endpoints
    module: web/nginx
    vars:
      command: add_location_block
      server_file: api-gateway.conf
      location_path: /api/v1/users/
      proxy_pass: http://user_service
      additional_config: |
        rewrite ^/api/v1/users/(.*)$ /$1 break;
        proxy_set_header X-Service-Name "users";

  # Route /products/* to product service
  - name: Route product endpoints
    module: web/nginx
    vars:
      command: add_location_block
      server_file: api-gateway.conf
      location_path: /api/v1/products/
      proxy_pass: http://product_service
      additional_config: |
        rewrite ^/api/v1/products/(.*)$ /$1 break;
        proxy_set_header X-Service-Name "products";

  # Add rate limiting
  - name: Add rate limiting
    module: web/nginx
    vars:
      command: add_location_block
      server_file: api-gateway.conf
      location_path: /api/
      additional_config: |
        limit_req zone=api_limit burst=20 nodelay;

  # Enable API gateway
  - name: Enable API gateway
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: api-gateway.conf

  - name: Reload configuration
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- /api/v1/auth/* routes to auth service
- /api/v1/users/* routes to user service
- /api/v1/products/* routes to product service
- URL rewriting removes /api/v1/service prefix
- Rate limiting applied

---

## 7. WebSocket Proxy

Configure WebSocket proxying for real-time applications.

### Scenario
Proxy WebSocket connections for a chat or real-time application.

### Stack File: `stacks/websocket-proxy.ofy`

```yaml
name: WebSocket Proxy
description: Configure WebSocket proxying for real-time apps
inventory: "@group:webservers"

defaults:
  domain: ws.example.com
  ws_backend: 127.0.0.1:3000

run:
  # Create server block
  - name: Create WebSocket server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: websocket.conf

  # Configure WebSocket proxy
  - name: Add WebSocket proxy location
    module: web/nginx
    vars:
      command: add_location_block
      server_file: websocket.conf
      location_path: /
      proxy_pass: "http://{{ var.ws_backend }}"
      additional_config: |
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;

  # Add connection upgrade map
  - name: Configure connection upgrade
    module: web/nginx
    vars:
      command: add_location_block
      server_file: websocket.conf
      location_path: /socket.io/
      proxy_pass: "http://{{ var.ws_backend }}"
      additional_config: |
        proxy_http_version 1.1;
        proxy_buffering off;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";

  # Enable WebSocket server
  - name: Enable WebSocket server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: websocket.conf

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- WebSocket connections properly upgraded
- Long-lived connections supported (24h timeout)
- Compatible with Socket.IO and other WS libraries

---

## 8. Multi-Tenant Hosting

Host multiple customer websites on a single server.

### Scenario
Manage 10+ customer websites with isolated configurations.

### Stack File: `stacks/multi-tenant.ofy`

```yaml
name: Multi-Tenant Hosting
description: Deploy multiple customer websites
inventory: "@group:webservers"

defaults:
  customers:
    - name: customer1
      domain: customer1.com
      root: /var/www/customer1
    - name: customer2
      domain: customer2.com
      root: /var/www/customer2
    - name: customer3
      domain: customer3.com
      root: /var/www/customer3

run:
  # Create server blocks for each customer
  - name: Create customer sites
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ item.domain }} www.{{ item.domain }}"
      server_file: "{{ item.name }}.conf"
      root: "{{ item.root }}"
      index: index.html index.php
    loop: "{{ var.customers }}"

  # Add PHP for each customer
  - name: Add PHP processing
    module: web/nginx
    vars:
      command: add_fastcgi_pass
      server_file: "{{ item.name }}.conf"
      location_path: ~ \.php$
      location_modifier: ~
      fastcgi_pass: unix:/var/run/php/php7.4-fpm.sock
    loop: "{{ var.customers }}"

  # Enable all customer sites
  - name: Enable customer sites
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "{{ item.name }}.conf"
    loop: "{{ var.customers }}"

  # Reload once at the end
  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- Each customer has isolated server block
- Individual document roots
- All sites enabled and active

---

## 9. Blue-Green Deployment

Implement blue-green deployment pattern for zero-downtime updates.

### Scenario
Switch traffic between blue and green application versions.

### Stack File: `stacks/blue-green.ofy`

```yaml
name: Blue-Green Deployment
description: Switch between blue and green application versions
inventory: "@group:loadbalancers"

defaults:
  domain: app.example.com
  blue_servers:
    - "10.0.3.10:8080"
    - "10.0.3.11:8080"
  green_servers:
    - "10.0.3.20:8080"
    - "10.0.3.21:8080"
  active_version: blue  # or green

run:
  # Create blue upstream
  - name: Create blue upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: app_blue
      upstream_servers: "{{ var.blue_servers }}"
      load_balance_method: least_conn

  # Create green upstream
  - name: Create green upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: app_green
      upstream_servers: "{{ var.green_servers }}"
      load_balance_method: least_conn

  # Create main server pointing to active version
  - name: Create main server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: app-main.conf

  - name: Configure proxy to active upstream
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: app-main.conf
      location_path: /
      proxy_pass: "http://app_{{ var.active_version }}"

  # Enable main server
  - name: Enable main server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: app-main.conf

  - name: Reload configuration
    module: web/nginx
    vars:
      command: reload_config
```

### Switching Versions

```yaml
# Switch from blue to green
- name: Update to green version
  module: web/nginx
  vars:
    command: set_server_name  # This is simplified - actual implementation would modify proxy_pass
    server_file: app-main.conf
    # In real implementation, you'd update the proxy_pass line

- name: Reload after switch
  module: web/nginx
  vars:
    command: reload_config
```

### Expected Result
- Traffic served from active version
- Other version ready for deployment
- Instant switching capability
- Easy rollback by switching back

---

## 10. CDN Origin Server

Configure Nginx as an origin server for CDN.

### Scenario
Set up Nginx to serve content to CloudFlare/Cloudinary/AWS CloudFront.

### Stack File: `stacks/cdn-origin.ofy`

```yaml
name: CDN Origin Server
description: Configure Nginx as CDN origin with caching
inventory: "@group:origins"

defaults:
  domain: origin.example.com
  doc_root: /var/www/static

run:
  # Create origin server
  - name: Create origin server block
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.domain }}"
      server_file: cdn-origin.conf
      root: "{{ var.doc_root }}"

  # Configure aggressive caching
  - name: Cache images
    module: web/nginx
    vars:
      command: add_location_block
      server_file: cdn-origin.conf
      location_path: ~* \.(jpg|jpeg|png|gif|webp)$
      location_modifier: ~*
      additional_config: |
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header X-Content-Type-Options "nosniff";
        access_log off;

  - name: Cache video files
    module: web/nginx
    vars:
      command: add_location_block
      server_file: cdn-origin.conf
      location_path: ~* \.(mp4|webm|ogg)$
      location_modifier: ~*
      additional_config: |
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
        sendfile on;
        tcp_nopush on;

  # CORS headers for fonts
  - name: Configure font CORS
    module: web/nginx
    vars:
      command: add_location_block
      server_file: cdn-origin.conf
      location_path: ~* \.(woff|woff2|ttf|eot)$
      location_modifier: ~*
      additional_config: |
        add_header Access-Control-Allow-Origin "*";
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;

  # Restrict access to CDN only (optional)
  - name: Add CDN IP whitelist
    module: web/nginx
    vars:
      command: add_location_block
      server_file: cdn-origin.conf
      location_path: /
      additional_config: |
        # CloudFlare IPs (example)
        allow 173.245.48.0/20;
        allow 103.21.244.0/22;
        allow 103.22.200.0/22;
        deny all;

  # Enable origin server
  - name: Enable origin server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: cdn-origin.conf

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- Optimized for CDN pull requests
- Long cache times for static assets
- CORS enabled for fonts
- Optional IP whitelisting
- Sendfile optimization for large files

---

## 11. Microservices Routing

Advanced routing for microservices architecture.

### Scenario
Route requests to 5 different microservices with service discovery and health checks.

### Stack File: `stacks/microservices.ofy`

```yaml
name: Microservices Router
description: Advanced routing for microservices
inventory: "@group:routers"

run:
  # Frontend service
  - name: Create frontend upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: frontend
      upstream_servers:
        - "10.0.10.10:3000"
        - "10.0.10.11:3000"

  # API services
  - name: Create orders service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: orders_api
      upstream_servers:
        - "10.0.11.10:8001"
        - "10.0.11.11:8001"
      load_balance_method: least_conn

  - name: Create payments service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: payments_api
      upstream_servers:
        - "10.0.12.10:8002"
      load_balance_method: ip_hash  # Sticky for payment sessions

  - name: Create inventory service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: inventory_api
      upstream_servers:
        - "10.0.13.10:8003"
        - "10.0.13.11:8003"

  - name: Create notifications service upstream
    module: web/nginx
    vars:
      command: create_upstream
      upstream_name: notifications_api
      upstream_servers:
        - "10.0.14.10:8004"

  # Main router server
  - name: Create router server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: app.example.com
      server_file: microservices-router.conf

  # Route frontend
  - name: Route to frontend
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: microservices-router.conf
      location_path: /
      proxy_pass: http://frontend

  # Route API services
  - name: Route orders API
    module: web/nginx
    vars:
      command: add_location_block
      server_file: microservices-router.conf
      location_path: /api/orders
      proxy_pass: http://orders_api
      additional_config: |
        proxy_set_header X-Service "orders";

  - name: Route payments API
    module: web/nginx
    vars:
      command: add_location_block
      server_file: microservices-router.conf
      location_path: /api/payments
      proxy_pass: http://payments_api
      additional_config: |
        proxy_set_header X-Service "payments";

  - name: Route inventory API
    module: web/nginx
    vars:
      command: add_location_block
      server_file: microservices-router.conf
      location_path: /api/inventory
      proxy_pass: http://inventory_api
      additional_config: |
        proxy_set_header X-Service "inventory";

  - name: Route notifications API
    module: web/nginx
    vars:
      command: add_location_block
      server_file: microservices-router.conf
      location_path: /api/notifications
      proxy_pass: http://notifications_api
      additional_config: |
        proxy_set_header X-Service "notifications";

  # Enable router
  - name: Enable microservices router
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: microservices-router.conf

  - name: Reload configuration
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- Frontend served from /
- Each API service routed to correct path
- Load balancing per service requirements
- Service identification headers

---

## 12. Development Environment

Quick setup for local development environment.

### Scenario
Set up development environment with auto-reloading and debugging.

### Stack File: `stacks/dev-environment.ofy`

```yaml
name: Development Environment
description: Quick dev setup with debugging enabled
inventory: localhost

defaults:
  project_name: myapp
  dev_port: 3000

run:
  # Create dev server
  - name: Create development server
    module: web/nginx
    vars:
      command: create_server_block
      server_name: "{{ var.project_name }}.local"
      server_file: "dev-{{ var.project_name }}.conf"

  # Proxy to dev server with debugging
  - name: Configure development proxy
    module: web/nginx
    vars:
      command: add_proxy_pass
      server_file: "dev-{{ var.project_name }}.conf"
      location_path: /
      proxy_pass: "http://127.0.0.1:{{ var.dev_port }}"

  # Disable caching for development
  - name: Disable caching
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "dev-{{ var.project_name }}.conf"
      location_path: /
      additional_config: |
        proxy_buffering off;
        proxy_cache off;
        add_header Cache-Control "no-store, no-cache, must-revalidate";
        expires 0;

  # Enable detailed error pages
  - name: Enable error pages
    module: web/nginx
    vars:
      command: add_location_block
      server_file: "dev-{{ var.project_name }}.conf"
      location_path: /
      additional_config: |
        proxy_intercept_errors on;
        error_page 500 502 503 504 /50x.html;

  # Enable and test
  - name: Enable dev server
    module: web/nginx
    vars:
      command: enable_server_block
      server_file: "dev-{{ var.project_name }}.conf"

  - name: Test configuration
    module: web/nginx
    vars:
      command: test_config

  - name: Reload Nginx
    module: web/nginx
    vars:
      command: reload_config
```

### Expected Result
- Development site at http://myapp.local
- No caching for live updates
- Detailed error messages
- Quick setup and teardown

---

## Workflow Best Practices

### 1. Always Test Before Reload
```yaml
- name: Test configuration
  module: web/nginx
  vars:
    command: test_config

- name: Reload only if valid
  module: web/nginx
  vars:
    command: reload_config
```

### 2. Use Variables for Reusability
```yaml
defaults:
  domain: example.com
  backend: 127.0.0.1:3000
```

### 3. Modular Stack Organization
```yaml
# Base configuration
- import: stacks/nginx-base.ofy

# Application-specific
- import: stacks/app-proxy.ofy
```

### 4. Progressive Deployment
```yaml
# 1. Create configuration
- name: Create server block
  ...

# 2. Test locally
- name: Test config
  ...

# 3. Enable on one server
- name: Enable on canary
  inventory: "@host:web-01"
  ...

# 4. Monitor metrics
# (manual step)

# 5. Roll out to all servers
- name: Enable on production
  inventory: "@group:webservers"
  ...
```

### 5. Health Checks
```yaml
- name: Add health check
  module: web/nginx
  vars:
    command: add_location_block
    location_path: /health
    location_modifier: "="
    additional_config: |
      access_log off;
      return 200 "OK";
```

---

## Next Steps

- Review [README.md](README.md) for module overview
- Check [COMMANDS.md](COMMANDS.md) for detailed command reference
- See [defaults.ofy.yml](defaults.ofy.yml) for all configuration options
- Test workflows in your environment with [test.ofy](test.ofy)
