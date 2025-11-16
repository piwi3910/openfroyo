# Fetch Module

Fetch files from remote hosts to the local machine.

## Description

The fetch module retrieves files from remote hosts via SSH and stores them locally. It uses base64 encoding to safely transfer file contents and preserves file metadata when possible.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| src | string | yes | - | Remote file path to fetch |
| dest | string | yes | - | Local destination path |
| flat | boolean | no | false | Flatten directory structure (store without subdirectories) |
| fail_on_missing | boolean | no | true | Fail if source file doesn't exist |

## Examples

### Basic file fetch
```yaml
- name: Fetch configuration file
  module: fetch
  vars:
    src: /etc/myapp/config.yml
    dest: ./backups/config.yml
```

### Fetch without failing on missing file
```yaml
- name: Fetch optional log file
  module: fetch
  vars:
    src: /var/log/application.log
    dest: ./logs/app.log
    fail_on_missing: false
```

### Fetch with flat structure
```yaml
- name: Fetch file to flat directory
  module: fetch
  vars:
    src: /var/www/html/index.html
    dest: ./web-backups/
    flat: true
```

## Implementation Details

The fetch module:
1. Checks if the source file exists on the remote host
2. Reads the file and base64 encodes it for safe transfer
3. Retrieves file metadata (size, permissions, owner, group)
4. Returns the encoded content via facts for the executor to decode and write locally

## Notes

- The executor handles writing the fetched file to the local filesystem
- File permissions and ownership information is captured but may not be preserved on the local copy
- Large files will be base64 encoded, increasing transfer size by ~33%
- The flat option controls whether directory structure is preserved in the destination path
