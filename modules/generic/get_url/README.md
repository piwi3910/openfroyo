# get_url Module

Download files from HTTP/HTTPS URLs with checksum verification, perfect for downloading binaries, configuration files, or the OpenFroyo agent itself.

## Features

- Download files from HTTP/HTTPS URLs
- Checksum verification (MD5, SHA1, SHA256, SHA512)
- Set file permissions and ownership
- Idempotent (skip if file exists with correct checksum)
- Follow HTTP redirects
- Custom headers support
- HTTP basic authentication
- Configurable timeouts
- SSL/TLS certificate validation

## Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `url` | Yes | - | HTTP/HTTPS URL to download from |
| `dest` | Yes | - | Absolute path where file should be saved |
| `checksum` | No | - | Checksum in format "algorithm:hash" (e.g., `sha256:abc123...`) |
| `mode` | No | `0644` | File permissions in octal format |
| `owner` | No | - | Owner of the downloaded file |
| `group` | No | - | Group of the downloaded file |
| `force` | No | `false` | Force download even if file exists |
| `timeout` | No | `60` | Timeout in seconds for HTTP request |
| `headers` | No | `{}` | Custom HTTP headers as key-value pairs |
| `validate_certs` | No | `true` | Validate SSL/TLS certificates |
| `follow_redirects` | No | `true` | Follow HTTP redirects |
| `username` | No | - | Username for HTTP basic auth |
| `password` | No | - | Password for HTTP basic auth |

## Checksum Formats

Checksums should be specified in the format `algorithm:hash`:

- `md5:d8e8fca2dc0f896fd7cb4cb0031ba249`
- `sha1:da39a3ee5e6b4b0d3255bfef95601890afd80709`
- `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e`

## Examples

### Basic Download

```yaml
- name: Download file
  module: generic/get_url
  vars:
    url: https://example.com/file.tar.gz
    dest: /tmp/file.tar.gz
```

### Download with Checksum Verification

```yaml
- name: Download with checksum
  module: generic/get_url
  vars:
    url: https://example.com/file.tar.gz
    dest: /tmp/file.tar.gz
    checksum: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

### Download Binary with Execute Permissions

```yaml
- name: Download froyo-agent
  module: generic/get_url
  vars:
    url: https://releases.openfroyo.com/v2.0/froyo-agent-linux-amd64
    dest: /usr/local/bin/froyo-agent
    mode: '0755'
    owner: root
    group: root
    checksum: sha256:abc123def456...
```

### Download with Custom Headers

```yaml
- name: Download from API
  module: generic/get_url
  vars:
    url: https://api.github.com/repos/user/repo/releases/latest
    dest: /tmp/latest-release.json
    headers:
      Accept: application/json
      Authorization: "Bearer token123"
      User-Agent: OpenFroyo/2.0
```

### Download with Basic Authentication

```yaml
- name: Download protected file
  module: generic/get_url
  vars:
    url: https://example.com/protected/file.zip
    dest: /tmp/file.zip
    username: myuser
    password: mypassword
```

### Download and Force Overwrite

```yaml
- name: Force download
  module: generic/get_url
  vars:
    url: https://example.com/file.txt
    dest: /tmp/file.txt
    force: true  # Download even if file exists
```

### Download with Extended Timeout

```yaml
- name: Download large file
  module: generic/get_url
  vars:
    url: https://example.com/large-file.iso
    dest: /tmp/large-file.iso
    timeout: 300  # 5 minutes
```

## Real-World Use Cases

### Download OpenFroyo Agent

```yaml
- name: Download froyo-agent binary
  module: generic/get_url
  vars:
    url: https://github.com/piwi3910/openfroyo/releases/download/v2.0/froyo-agent-linux-amd64
    dest: /usr/local/bin/froyo-agent
    mode: '0755'
    owner: root
    group: root
    checksum: sha256:{{ agent_checksum }}
```

### Download Docker Compose

```yaml
- name: Download docker-compose
  module: generic/get_url
  vars:
    url: https://github.com/docker/compose/releases/download/v2.23.0/docker-compose-linux-x86_64
    dest: /usr/local/bin/docker-compose
    mode: '0755'
    checksum: sha256:916b5a0e1f1e77568f9e8ba5c7a9e8e96c5e7b9c8d9e8f7a6b5c4d3e2f1a0b9
```

### Download Configuration File

```yaml
- name: Download nginx config
  module: generic/get_url
  vars:
    url: https://config-server.example.com/nginx/{{ env }}/nginx.conf
    dest: /etc/nginx/nginx.conf
    mode: '0644'
    owner: root
    group: root
    headers:
      Authorization: "Bearer {{ config_token }}"
```

## Idempotency

The module is idempotent:
- If `force=false` (default) and the file exists:
  - If `checksum` is provided and matches, status is "ok" (no download)
  - If `checksum` is not provided, status is "ok" (no download)
  - If `checksum` doesn't match, file is re-downloaded
- If `force=true`, file is always downloaded

## Security Considerations

- **Checksum verification:** Always use checksums for binary downloads to prevent tampering
- **HTTPS:** Prefer HTTPS URLs over HTTP
- **Certificate validation:** Keep `validate_certs=true` in production
- **Credentials:** Use vault or secrets management for `username`/`password`, don't hardcode

## Return Values

The module returns the following facts:

```json
{
  "status": "changed|ok|failed",
  "message": "Download status message",
  "facts": {
    "shell_exec": [...],  // Shell commands executed
    "url": "source URL",
    "dest": "destination path",
    "checksum": "calculated checksum"
  }
}
```

## Notes

- Uses `curl` command on remote host (must be installed)
- Downloads to temporary file first, then moves to destination
- Automatically creates parent directories if they don't exist (via curl -o behavior)
- Supports all HTTP methods supported by curl

## See Also

- [file module](../file/) - Manage files and directories
- [copy module](../copy/) - Copy files from local to remote
- [template module](../template/) - Template file generation
