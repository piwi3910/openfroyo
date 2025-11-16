# Stat Module

The **stat** module retrieves file or directory status information from remote hosts.

## Features

- Check if a file or directory exists
- Retrieve file size, permissions, ownership, and timestamps
- Cross-platform support (GNU stat, BSD stat, and ls fallback)
- Works with files, directories, and symlinks

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| path | string | yes | - | Path to the file or directory to stat |

## Implementation Details

The stat module follows the OpenFroyo pattern of returning `shell_exec` facts:

1. **WASM Module**: Builds a shell command to stat the file
2. **Shell Commands**: Returns commands that:
   - Check if the path exists
   - Try GNU stat format first (Linux)
   - Fall back to BSD stat format (macOS/BSD)
   - Fall back to `ls -la` if stat is not available
   - Report existence status

## Usage Examples

### Stat an existing file
```yaml
- name: Check /etc/hosts file
  stat:
    path: /etc/hosts
```

### Stat a directory
```yaml
- name: Check /var/log directory
  stat:
    path: /var/log
```

### Check if file exists
```yaml
- name: Check if config exists
  stat:
    path: /etc/myapp/config.yml
```

## Testing

Test the module with:

```bash
# Build the module
make build

# Test with existing file
echo -n '{"vars":{"path":"/etc/hosts"},"context":{"host":"localhost","task_name":"Test stat"}}' | base64 | \
./froyo-runner --module modules/generic/stat/wasm/stat.wasm --input-base64 -

# Test with non-existent file
echo -n '{"vars":{"path":"/tmp/does_not_exist"},"context":{"host":"localhost","task_name":"Test stat"}}' | base64 | \
./froyo-runner --module modules/generic/stat/wasm/stat.wasm --input-base64 -
```

## Output Format

The module returns `shell_exec` facts with the stat command. The output varies by platform:

### GNU stat (Linux)
```
213|100644|root|wheel|1761700864|1761700864
EXISTS=true
```

Format: `size|mode|owner|group|mtime|atime`

### BSD stat (macOS/BSD)
```
213|100644|root|wheel|1761700864|1761700864
EXISTS=true
```

Format: `size|permissions|owner|group|mtime|atime`

### ls fallback
```
-rw-r--r--  1 root  wheel  213 Jan 16 10:30 /etc/hosts
EXISTS=true
```

### Non-existent file
```
EXISTS=false
```

## Example Output

```json
{
  "status": "ok",
  "message": "Command executed: ...",
  "facts": {
    "stdout": "213|100644|root|wheel|1761700864|1761700864\nEXISTS=true"
  }
}
```

## Parsing the Output

The executor should parse the output to extract:
- **exists**: Whether the file/directory exists (true/false)
- **size**: File size in bytes
- **mode**: File permissions
- **owner**: File owner username
- **group**: File group name
- **mtime**: Modification timestamp (Unix time)
- **atime**: Access timestamp (Unix time)

The format is designed to be parsed by splitting on `|` when EXISTS=true.
