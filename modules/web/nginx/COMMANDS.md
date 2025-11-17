# Nginx Module Command Reference

Complete reference for all 35+ commands available in the Nginx module.

## Table of Contents

- [Server Block Management](#server-block-management)
- [Upstream and Load Balancing](#upstream-and-load-balancing)
- [SSL/TLS Configuration](#ssltls-configuration)
- [Location Blocks](#location-blocks)
- [Configuration Management](#configuration-management)
- [Service Management](#service-management)
- [Performance Tuning](#performance-tuning)

---

## Server Block Management

### create_server_block

Create a new Nginx server block (virtual host) configuration file.

**Required Variables:**
- `server_name`: Domain name(s) for this server block

**Optional Variables:**
- `server_file`: Configuration file name (defaults to `{server_name}.conf`)
- `listen_port`: HTTP port to listen on (default: `80`)
- `root`: Document root directory
- `index`: Index file names (default: `index.html index.htm`)
- `access_log`: Access log file path
- `error_log`: Error log file path
- `additional_config`: Additional raw Nginx configuration

**Example:**
```yaml
- name: Create basic server block
  module: web/nginx
  vars:
    command: create_server_block
    server_name: example.com www.example.com
    server_file: example.com.conf
    root: /var/www/example.com
    index: index.html index.htm index.php
```

**Generated Configuration:**
```nginx
server {
    listen 80;
    server_name example.com www.example.com;
    root /var/www/example.com;
    index index.html index.htm index.php;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

**Output Facts:**
- `config_path`: Full path to created configuration file
- `server_name`: Server names configured

---

### delete_server_block

Delete a server block configuration file and disable it if currently enabled.

**Required Variables:**
- `server_file`: Configuration file name to delete

**Example:**
```yaml
- name: Delete old server block
  module: web/nginx
  vars:
    command: delete_server_block
    server_file: old-site.conf
```

**Behavior:**
- Automatically disables the server block if enabled
- Tests configuration after deletion
- Reloads Nginx if test passes

---

### enable_server_block

Enable a server block by creating a symlink from `sites-available` to `sites-enabled`.

**Required Variables:**
- `server_file`: Configuration file name to enable

**Example:**
```yaml
- name: Enable website
  module: web/nginx
  vars:
    command: enable_server_block
    server_file: example.com.conf
```

**Behavior:**
- Creates symlink: `sites-enabled/example.com.conf -> sites-available/example.com.conf`
- Tests configuration before enabling
- Rolls back if test fails
- Reloads Nginx on success

**Output Facts:**
- `enabled_path`: Path to created symlink

---

### disable_server_block

Disable a server block by removing the symlink from `sites-enabled`.

**Required Variables:**
- `server_file`: Configuration file name to disable

**Example:**
```yaml
- name: Disable maintenance site
  module: web/nginx
  vars:
    command: disable_server_block
    server_file: maintenance.conf
```

**Behavior:**
- Removes symlink from `sites-enabled`
- Reloads Nginx configuration
- Does not delete the configuration file

---

### list_server_blocks

List all available and enabled server blocks.

**No Required Variables**

**Example:**
```yaml
- name: Inventory server blocks
  module: web/nginx
  vars:
    command: list_server_blocks
```

**Output Facts:**
- `available`: Array of all configuration files in `sites-available`
- `enabled`: Array of all enabled server blocks

**Example Output:**
```json
{
  "available": ["default", "example.com.conf", "api.example.com.conf"],
  "enabled": ["default", "example.com.conf"]
}
```

---

### test_server_config

Test Nginx configuration validity using `nginx -t`.

**No Required Variables**

**Example:**
```yaml
- name: Validate configuration
  module: web/nginx
  vars:
    command: test_server_config
```

**Output Facts:**
- `valid`: Boolean indicating if configuration is valid

---

## Upstream and Load Balancing

### create_upstream

Create an upstream block defining a group of backend servers for load balancing.

**Required Variables:**
- `upstream_name`: Name for the upstream block
- `upstream_servers`: Array of backend server addresses (format: `host:port`)

**Optional Variables:**
- `load_balance_method`: Load balancing algorithm (default: `round_robin`)
  - Options: `round_robin`, `least_conn`, `ip_hash`, `hash`, `random`
- `hash_key`: Key for hash-based load balancing (default: `$request_uri`)
- `upstream_keepalive`: Number of keepalive connections to backends

**Example:**
```yaml
- name: Create application upstream
  module: web/nginx
  vars:
    command: create_upstream
    upstream_name: appservers
    upstream_servers:
      - "192.168.1.10:8080"
      - "192.168.1.11:8080"
      - "192.168.1.12:8080"
    load_balance_method: least_conn
    upstream_keepalive: 32
```

**Generated Configuration:**
```nginx
upstream appservers {
    least_conn;
    server 192.168.1.10:8080;
    server 192.168.1.11:8080;
    server 192.168.1.12:8080;

    keepalive 32;
}
```

**Load Balancing Methods:**

| Method | Description | Use Case |
|--------|-------------|----------|
| `round_robin` | Distribute requests evenly (default) | Balanced load, stateless apps |
| `least_conn` | Send to server with fewest connections | Long-running requests |
| `ip_hash` | Same client IP always goes to same server | Session persistence |
| `hash` | Hash a custom key to determine server | Custom sticky routing |
| `random` | Random server selection | Simple distribution |

**Output Facts:**
- `upstream_file`: Path to created upstream configuration
- `servers`: Array of configured server addresses

---

### delete_upstream

Delete an upstream configuration.

**Required Variables:**
- `upstream_name`: Name of upstream to delete

**Example:**
```yaml
- name: Remove old upstream
  module: web/nginx
  vars:
    command: delete_upstream
    upstream_name: old_backend
```

---

### add_upstream_server

Add a backend server to an existing upstream block.

**Required Variables:**
- `upstream_name`: Name of upstream to modify
- `upstream_server`: Server address to add (format: `host:port`)

**Example:**
```yaml
- name: Add new backend server
  module: web/nginx
  vars:
    command: add_upstream_server
    upstream_name: appservers
    upstream_server: "192.168.1.13:8080"
```

**Behavior:**
- Checks if server already exists
- Adds server line before closing brace
- Tests and reloads configuration

---

### remove_upstream_server

Remove a backend server from an upstream block.

**Required Variables:**
- `upstream_name`: Name of upstream to modify
- `upstream_server`: Server address to remove (format: `host:port`)

**Example:**
```yaml
- name: Remove failed backend
  module: web/nginx
  vars:
    command: remove_upstream_server
    upstream_name: appservers
    upstream_server: "192.168.1.10:8080"
```

---

### set_load_balance_method

Change the load balancing algorithm for an upstream.

**Required Variables:**
- `upstream_name`: Name of upstream to modify
- `load_balance_method`: New load balancing method

**Optional Variables:**
- `hash_key`: Key for hash-based methods (default: `$request_uri`)

**Example:**
```yaml
- name: Switch to least connections
  module: web/nginx
  vars:
    command: set_load_balance_method
    upstream_name: appservers
    load_balance_method: least_conn
```

**Hash-Based Example:**
```yaml
- name: Use cookie-based hashing
  module: web/nginx
  vars:
    command: set_load_balance_method
    upstream_name: appservers
    load_balance_method: hash
    hash_key: $cookie_session
```

---

### show_upstreams

List all configured upstream blocks.

**No Required Variables**

**Example:**
```yaml
- name: List all upstreams
  module: web/nginx
  vars:
    command: show_upstreams
```

**Output Facts:**
- `upstreams`: Array of upstream names

---

## SSL/TLS Configuration

### create_ssl_server

Create an HTTPS server block with SSL/TLS configuration and HTTP to HTTPS redirect.

**Required Variables:**
- `server_name`: Domain name(s)
- `ssl_certificate`: Path to SSL certificate file
- `ssl_certificate_key`: Path to SSL certificate key file

**Optional Variables:**
- `server_file`: Configuration file name (defaults to `{server_name}-ssl.conf`)
- `listen_port`: HTTP redirect port (default: `80`)
- `listen_ssl_port`: HTTPS port (default: `443`)
- `ssl_protocols`: Allowed SSL/TLS protocols (default: `TLSv1.2 TLSv1.3`)
- `ssl_ciphers`: Cipher suites (default: `HIGH:!aNULL:!MD5`)
- `http2`: Enable HTTP/2 (default: `false`)
- `hsts_enabled`: Enable HSTS (default: `false`)
- `hsts_max_age`: HSTS max-age in seconds (default: `31536000`)
- `root`: Document root
- `index`: Index files

**Example:**
```yaml
- name: Create SSL website
  module: web/nginx
  vars:
    command: create_ssl_server
    server_name: secure.example.com
    server_file: secure.example.com.conf
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
    ssl_protocols: "TLSv1.2 TLSv1.3"
    http2: "true"
    hsts_enabled: "true"
    root: /var/www/secure
```

**Generated Configuration:**
```nginx
# HTTP to HTTPS redirect
server {
    listen 80;
    server_name secure.example.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name secure.example.com;

    ssl_certificate /etc/ssl/certs/example.com.crt;
    ssl_certificate_key /etc/ssl/private/example.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    add_header Strict-Transport-Security "max-age=31536000" always;

    root /var/www/secure;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }
}
```

**Security Recommendations:**
- Use TLSv1.2 and TLSv1.3 only
- Enable HTTP/2 for better performance
- Enable HSTS for security
- Use strong cipher suites

---

### install_ssl_certificate

Verify that SSL certificate files exist and are readable.

**Required Variables:**
- `ssl_certificate`: Path to certificate file
- `ssl_certificate_key`: Path to key file

**Example:**
```yaml
- name: Verify SSL certificates
  module: web/nginx
  vars:
    command: install_ssl_certificate
    ssl_certificate: /etc/ssl/certs/example.com.crt
    ssl_certificate_key: /etc/ssl/private/example.com.key
```

**Output Facts:**
- `certificate`: Certificate file path
- `certificate_key`: Key file path

---

### set_ssl_protocols

Update SSL/TLS protocol versions for a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `ssl_protocols`: Space-separated list of protocols

**Example:**
```yaml
- name: Update SSL protocols
  module: web/nginx
  vars:
    command: set_ssl_protocols
    server_file: secure.example.com.conf
    ssl_protocols: "TLSv1.2 TLSv1.3"
```

**Protocol Options:**
- `TLSv1.3` - Modern, recommended
- `TLSv1.2` - Widely supported, secure
- `TLSv1.1` - Deprecated, avoid
- `TLSv1.0` - Deprecated, avoid

---

### set_ssl_ciphers

Update SSL cipher suites for a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `ssl_ciphers`: Cipher suite string

**Example:**
```yaml
- name: Set strong ciphers
  module: web/nginx
  vars:
    command: set_ssl_ciphers
    server_file: secure.example.com.conf
    ssl_ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256"
```

**Common Cipher Configurations:**
- `HIGH:!aNULL:!MD5` - Strong ciphers, no anonymous or MD5
- Mozilla Modern: `ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256`
- Mozilla Intermediate: `ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305`

---

### enable_http2

Enable HTTP/2 protocol support on an HTTPS server block.

**Required Variables:**
- `server_file`: Server configuration file to modify

**Example:**
```yaml
- name: Enable HTTP/2
  module: web/nginx
  vars:
    command: enable_http2
    server_file: secure.example.com.conf
```

**Requirements:**
- Nginx compiled with HTTP/2 support
- Must be used with SSL/TLS (HTTPS)

**Benefits:**
- Multiplexed connections
- Server push capability
- Header compression
- Better performance

---

### test_ssl_config

Test SSL configuration validity.

**No Required Variables**

**Example:**
```yaml
- name: Test SSL configuration
  module: web/nginx
  vars:
    command: test_ssl_config
```

---

## Location Blocks

### add_location_block

Add a location block to a server configuration.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `location_path`: URL path pattern for location

**Optional Variables:**
- `location_modifier`: Location modifier (`=`, `~`, `~*`, `^~`)
- `proxy_pass`: Proxy backend URL
- `fastcgi_pass`: FastCGI socket or address
- `alias`: Alias path for file serving
- `try_files`: Try files directive
- `additional_config`: Additional raw configuration

**Example (Static Files):**
```yaml
- name: Add static files location
  module: web/nginx
  vars:
    command: add_location_block
    server_file: example.com.conf
    location_path: /static/
    alias: /var/www/static/
    additional_config: |
      expires 1y;
      add_header Cache-Control "public, immutable";
```

**Example (Proxy):**
```yaml
- name: Add API proxy location
  module: web/nginx
  vars:
    command: add_location_block
    server_file: example.com.conf
    location_path: /api/
    proxy_pass: http://backend_api
```

**Location Modifiers:**
- `=` - Exact match (highest priority)
- `^~` - Prefix match, stop regex matching
- `~` - Case-sensitive regex
- `~*` - Case-insensitive regex
- (none) - Prefix match

---

### add_proxy_pass

Add a reverse proxy location block with standard proxy headers.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `proxy_pass`: Backend URL to proxy to

**Optional Variables:**
- `location_path`: URL path (default: `/`)
- `location_modifier`: Location modifier

**Example:**
```yaml
- name: Add reverse proxy
  module: web/nginx
  vars:
    command: add_proxy_pass
    server_file: app.example.com.conf
    location_path: /
    proxy_pass: http://backend_servers
```

**Generated Configuration:**
```nginx
location / {
    proxy_pass http://backend_servers;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

---

### add_fastcgi_pass

Add a FastCGI location block for PHP or other FastCGI applications.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `fastcgi_pass`: FastCGI socket or address

**Optional Variables:**
- `location_path`: URL path pattern (default: `~ \.php$`)
- `location_modifier`: Location modifier (default: `~`)

**Example:**
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

**Generated Configuration:**
```nginx
location ~ \.php$ {
    fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
    fastcgi_index index.php;
    include fastcgi_params;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
}
```

---

### add_static_location

Add a location block optimized for static file serving.

**Required Variables:**
- `server_file`: Server configuration file to modify

**Optional Variables:**
- `location_path`: URL path (default: `/`)
- `try_files`: Try files directive (default: `$uri $uri/ =404`)

**Example:**
```yaml
- name: Add static file serving
  module: web/nginx
  vars:
    command: add_static_location
    server_file: cdn.example.com.conf
    location_path: /
    try_files: $uri $uri/ =404
```

---

### add_rewrite_rule

Add a URL rewrite rule to a server or location block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `rewrite_regex`: Regex pattern to match
- `rewrite_replacement`: Replacement URL

**Optional Variables:**
- `rewrite_flag`: Rewrite flag (default: `last`)
  - Options: `last`, `break`, `redirect`, `permanent`
- `location_path`: Add inside specific location block

**Example (Permanent Redirect):**
```yaml
- name: Add redirect from old to new
  module: web/nginx
  vars:
    command: add_rewrite_rule
    server_file: example.com.conf
    rewrite_regex: "^/old/(.*)$"
    rewrite_replacement: /new/$1
    rewrite_flag: permanent
```

**Rewrite Flags:**
- `last` - Stop processing, search for new location
- `break` - Stop processing, use current location
- `redirect` - 302 temporary redirect
- `permanent` - 301 permanent redirect

---

## Configuration Management

### set_server_name

Update the server_name directive in a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `server_name`: New server name(s)

**Example:**
```yaml
- name: Update server name
  module: web/nginx
  vars:
    command: set_server_name
    server_file: example.com.conf
    server_name: example.com www.example.com example.org
```

---

### set_root

Update the root directory for a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `root`: New document root path

**Example:**
```yaml
- name: Update document root
  module: web/nginx
  vars:
    command: set_root
    server_file: example.com.conf
    root: /var/www/html/example.com/public
```

---

### set_index

Update the index file list for a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `index`: Space-separated list of index files

**Example:**
```yaml
- name: Update index files
  module: web/nginx
  vars:
    command: set_index
    server_file: example.com.conf
    index: index.php index.html index.htm
```

---

### add_header

Add a custom HTTP response header to a server block.

**Required Variables:**
- `server_file`: Server configuration file to modify
- `header_name`: HTTP header name
- `header_value`: Header value

**Optional Variables:**
- `header_action`: Action type (default: `add`)
  - Options: `add`, `set`

**Example (Security Headers):**
```yaml
- name: Add X-Frame-Options
  module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: X-Frame-Options
    header_value: SAMEORIGIN

- name: Add Content-Security-Policy
  module: web/nginx
  vars:
    command: add_header
    server_file: example.com.conf
    header_name: Content-Security-Policy
    header_value: "default-src 'self'"
```

**Common Security Headers:**
- `X-Frame-Options: SAMEORIGIN` - Prevent clickjacking
- `X-Content-Type-Options: nosniff` - Prevent MIME sniffing
- `X-XSS-Protection: 1; mode=block` - XSS protection
- `Referrer-Policy: no-referrer-when-downgrade` - Control referrer
- `Content-Security-Policy` - Control resource loading

---

### test_config

Test complete Nginx configuration for syntax errors.

**No Required Variables**

**Example:**
```yaml
- name: Test configuration
  module: web/nginx
  vars:
    command: test_config
```

**Output Facts:**
- `valid`: Boolean indicating if configuration is valid

---

### reload_config

Reload Nginx configuration gracefully (zero downtime).

**No Required Variables**

**Example:**
```yaml
- name: Reload Nginx
  module: web/nginx
  vars:
    command: reload_config
```

**Behavior:**
- Tests configuration first
- Only reloads if test passes
- Uses graceful reload (no connection drops)

---

## Service Management

### start_service

Start the Nginx service.

**No Required Variables**

**Example:**
```yaml
- name: Start Nginx
  module: web/nginx
  vars:
    command: start_service
```

**Uses:** `systemctl start nginx`

---

### stop_service

Stop the Nginx service.

**No Required Variables**

**Example:**
```yaml
- name: Stop Nginx
  module: web/nginx
  vars:
    command: stop_service
```

**Uses:** `systemctl stop nginx`

---

### restart_service

Restart the Nginx service.

**No Required Variables**

**Example:**
```yaml
- name: Restart Nginx
  module: web/nginx
  vars:
    command: restart_service
```

**Behavior:**
- Tests configuration first
- Only restarts if test passes
- Uses `systemctl restart nginx`

**Note:** Prefer `reload_config` for zero-downtime configuration changes.

---

### check_service_status

Check if Nginx service is running.

**No Required Variables**

**Example:**
```yaml
- name: Check Nginx status
  module: web/nginx
  vars:
    command: check_service_status
```

**Output Facts:**
- `running`: Boolean indicating if service is active
- `status`: Service status string

---

## Performance Tuning

### set_worker_processes

Configure the number of Nginx worker processes.

**Required Variables:**
- `worker_processes`: Number of workers or `auto`

**Example:**
```yaml
- name: Set worker processes to CPU count
  module: web/nginx
  vars:
    command: set_worker_processes
    worker_processes: auto
```

**Guidelines:**
- Use `auto` to match CPU core count
- Or set specific number (e.g., `4`)
- More workers = more concurrent connections
- Don't exceed CPU core count significantly

**Modifies:** `/etc/nginx/nginx.conf`

---

### set_worker_connections

Configure maximum connections per worker process.

**Required Variables:**
- `worker_connections`: Maximum connections per worker

**Example:**
```yaml
- name: Increase worker connections
  module: web/nginx
  vars:
    command: set_worker_connections
    worker_connections: 2048
```

**Guidelines:**
- Default: 1024
- Increase for high-traffic sites
- Must not exceed OS file descriptor limit
- Total connections = worker_processes × worker_connections

**Check system limits:**
```bash
ulimit -n  # Check file descriptor limit
```

**Modifies:** `/etc/nginx/nginx.conf`

---

## Command Summary Table

| Category | Command | Purpose |
|----------|---------|---------|
| **Server Blocks** | create_server_block | Create new virtual host |
| | delete_server_block | Remove virtual host |
| | enable_server_block | Activate virtual host |
| | disable_server_block | Deactivate virtual host |
| | list_server_blocks | Show all virtual hosts |
| | test_server_config | Validate configuration |
| **Upstreams** | create_upstream | Create backend pool |
| | delete_upstream | Remove backend pool |
| | add_upstream_server | Add backend server |
| | remove_upstream_server | Remove backend server |
| | set_load_balance_method | Change load balancing |
| | show_upstreams | List backend pools |
| **SSL/TLS** | create_ssl_server | Create HTTPS site |
| | install_ssl_certificate | Verify certificates |
| | set_ssl_protocols | Set TLS versions |
| | set_ssl_ciphers | Set cipher suites |
| | enable_http2 | Enable HTTP/2 |
| | test_ssl_config | Validate SSL config |
| **Locations** | add_location_block | Add URL route |
| | add_proxy_pass | Add reverse proxy |
| | add_fastcgi_pass | Add PHP/FastCGI |
| | add_static_location | Add static files |
| | add_rewrite_rule | Add URL rewrite |
| **Config** | set_server_name | Update domain names |
| | set_root | Update document root |
| | set_index | Update index files |
| | add_header | Add HTTP header |
| | test_config | Validate all config |
| | reload_config | Graceful reload |
| **Service** | start_service | Start Nginx |
| | stop_service | Stop Nginx |
| | restart_service | Restart Nginx |
| | check_service_status | Check if running |
| **Performance** | set_worker_processes | Set worker count |
| | set_worker_connections | Set connection limit |

---

## Common Variable Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `nginx_config_dir` | string | `/etc/nginx` | Main config directory |
| `nginx_sites_available` | string | `/etc/nginx/sites-available` | Available configs |
| `nginx_sites_enabled` | string | `/etc/nginx/sites-enabled` | Enabled configs |
| `server_name` | string | - | Domain name(s) |
| `server_file` | string | - | Config file name |
| `listen_port` | string | `80` | HTTP port |
| `listen_ssl_port` | string | `443` | HTTPS port |
| `root` | string | - | Document root |
| `index` | string | `index.html index.htm` | Index files |
| `ssl_certificate` | string | - | SSL cert path |
| `ssl_certificate_key` | string | - | SSL key path |
| `ssl_protocols` | string | `TLSv1.2 TLSv1.3` | TLS versions |
| `upstream_name` | string | - | Upstream name |
| `upstream_servers` | array | `[]` | Backend servers |
| `load_balance_method` | string | `round_robin` | LB algorithm |
| `worker_processes` | string | `auto` | Worker count |
| `worker_connections` | string | `1024` | Max connections |
