# ISO Extract Module

The **iso_extract** module extracts files from ISO images on remote hosts.

## Features

- Extract entire ISO images or specific files
- Cross-platform support (Linux, macOS, BSD)
- Multiple extraction methods (7z, bsdtar, mount loop)
- Idempotency support with `creates` parameter
- Automatic tool detection and fallback

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| image | string | yes | - | Path to the ISO image file |
| dest | string | yes | - | Destination directory for extracted files |
| files | array | no | all files | List of specific files to extract |
| creates | string | no | - | Skip extraction if this file exists |

## Implementation Details

The iso_extract module follows the OpenFroyo pattern of returning `shell_exec` facts:

1. **WASM Module**: Builds shell commands to extract the ISO
2. **Shell Commands**: Returns commands that:
   - Check if ISO image exists
   - Create destination directory
   - Try multiple extraction tools in order:
     - `7z` (7-Zip command line)
     - `bsdtar` (BSD tar with ISO support)
     - `mount -o loop` (Linux mount method)
   - Verify extraction succeeded

## Usage Examples

### Extract entire ISO
```yaml
- name: Extract Alpine Linux ISO
  iso_extract:
    image: /tmp/alpine-linux.iso
    dest: /tmp/alpine-extracted
```

### Extract specific files
```yaml
- name: Extract kernel and initrd
  iso_extract:
    image: /tmp/ubuntu.iso
    dest: /boot/ubuntu
    files:
      - boot/vmlinuz
      - boot/initrd.img
```

### Idempotent extraction
```yaml
- name: Extract ISO only if not already done
  iso_extract:
    image: /mnt/debian.iso
    dest: /var/cache/debian
    creates: /var/cache/debian/.extracted
```

## Testing

Test the module with:

```bash
# Build the module
make build

# Download a test ISO (Alpine Linux is small ~50MB)
curl -o /tmp/alpine.iso https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-virt-3.19.0-x86_64.iso

# Test extraction
echo -n '{"vars":{"image":"/tmp/alpine.iso","dest":"/tmp/alpine-test"},"context":{"host":"localhost","task_name":"Test ISO extract"}}' | base64 | \
./froyo-runner --module modules/generic/iso_extract/wasm/iso_extract.wasm --input-base64 -
```

## Extraction Methods

The module tries multiple extraction methods in order of preference:

### 1. 7-Zip (7z)
- **Best for**: All platforms with 7z installed
- **Install**:
  - Linux: `apt install p7zip-full` or `yum install p7zip`
  - macOS: `brew install p7zip`
- **Command**: `7z x image.iso -o/dest`

### 2. BSD tar (bsdtar)
- **Best for**: macOS and BSD systems (pre-installed)
- **Install**: Usually pre-installed on macOS/BSD
  - Linux: `apt install libarchive-tools`
- **Command**: `bsdtar -xf image.iso -C /dest`

### 3. Mount Loop
- **Best for**: Linux systems with root access
- **Requires**: Root/sudo privileges
- **Command**: `mount -o loop image.iso /mnt && cp -r /mnt/* /dest`

## Output Format

The module returns `shell_exec` facts with extraction commands:

```json
{
  "status": "ok",
  "message": "",
  "facts": {
    "shell_exec": [
      {
        "type": "shell",
        "command": "[ -f '/tmp/image.iso' ] || ..."
      },
      {
        "type": "shell",
        "command": "mkdir -p '/tmp/dest'"
      },
      {
        "type": "shell",
        "command": "7z x ... || bsdtar -xf ... || mount ..."
      },
      {
        "type": "shell",
        "command": "[ -d '/tmp/dest' ] && ..."
      }
    ]
  }
}
```

## Example Output

Success case:
```json
{
  "status": "ok",
  "message": "Command executed successfully",
  "facts": {
    "stdout": "Extraction complete\n"
  }
}
```

With `creates` parameter (already extracted):
```json
{
  "status": "ok",
  "message": "Command executed successfully",
  "facts": {
    "stdout": "SKIP: Already extracted\n"
  }
}
```

## Error Handling

The module handles several error cases:

- **ISO not found**: Returns error if image file doesn't exist
- **Extraction failed**: Returns error if all extraction methods fail
- **No tools available**: Falls back through multiple tools
- **Permission denied**: May fail if mount method used without root

## Performance Considerations

- Extracting large ISOs can take significant time
- Use `files` parameter to extract only needed files
- Use `creates` parameter to avoid re-extraction
- Consider disk space requirements (ISO + extracted files)

## Platform Support

| Platform | Recommended Tool | Notes |
|----------|------------------|-------|
| Linux | 7z or bsdtar | mount loop requires root |
| macOS | bsdtar (built-in) | 7z also works well |
| BSD | bsdtar (built-in) | Native support |
| Windows | 7z | Requires 7-Zip installation |
