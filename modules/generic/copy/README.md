# Copy Module

The **copy** module copies files from the local machine to remote hosts via SSH.

## Features

- Copy files from local filesystem to remote hosts
- Set file permissions (mode)
- Set file ownership (owner and group)
- Optional backup of existing destination files
- Automatic creation of destination directories

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| src | string | yes | - | Local source file path to copy |
| dest | string | yes | - | Remote destination path |
| mode | string | no | - | File permissions in octal (e.g., '0644', '0755') |
| owner | string | no | - | Owner of the destination file |
| group | string | no | - | Group owner of the destination file |
| backup | boolean | no | false | Create a backup of the destination file if it exists |

## Implementation Details

The copy module follows the OpenFroyo pattern of returning `shell_exec` facts:

1. **Local File Reading**: The OpenFroyo executor reads the source file from the local filesystem and base64-encodes it
2. **WASM Module**: Receives the encoded content via the `_file_content` variable
3. **Shell Commands**: Returns a list of shell commands to:
   - Create destination directory (if needed)
   - Backup existing file (if requested)
   - Write file content to destination
   - Set permissions (if specified)
   - Set ownership (if specified)
   - Verify file creation

## Usage Examples

### Basic file copy
```yaml
- name: Copy configuration file
  copy:
    src: /local/path/config.yml
    dest: /etc/myapp/config.yml
```

### Copy with permissions
```yaml
- name: Copy script with execute permissions
  copy:
    src: /local/scripts/deploy.sh
    dest: /usr/local/bin/deploy.sh
    mode: "0755"
```

### Copy with ownership
```yaml
- name: Copy web content
  copy:
    src: /local/www/index.html
    dest: /var/www/html/index.html
    owner: www-data
    group: www-data
    mode: "0644"
```

### Copy with backup
```yaml
- name: Copy config with backup
  copy:
    src: /local/config/app.conf
    dest: /etc/app/app.conf
    backup: true
```

## Testing

Test the module with:

```bash
# Build the module
make build

# Test with froyo-runner
echo -n '{"vars":{"src":"/tmp/test.txt","dest":"/tmp/dest.txt","_file_content":"SGVsbG8gV29ybGQh"},"context":{"host":"localhost","task_name":"Test copy"}}' | base64 | \
./froyo-runner --module modules/generic/copy/wasm/copy.wasm --input-base64 -
```

## Output

The module returns `shell_exec` facts with commands to execute on the remote host. The OpenFroyo runner executes these commands and reports the result.

Example output:
```json
{
  "status": "ok",
  "message": "",
  "facts": {
    "shell_exec": [
      {"type": "shell", "command": "mkdir -p '/path/to'"},
      {"type": "file_write", "path": "/path/to/file", "content": "...", "mode": 420},
      {"type": "shell", "command": "chmod 0644 '/path/to/file'"},
      {"type": "shell", "command": "[ -f '/path/to/file' ] && ls -l '/path/to/file'"}
    ]
  }
}
```
