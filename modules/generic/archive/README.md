# Archive Module

Create compressed archives in various formats (tar.gz, tar.bz2, tar.xz, zip).

## Description

The archive module compresses files and directories into archive files. It supports multiple compression formats and can optionally remove source files after successful archiving.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| path | string | yes | - | Files or directories to archive |
| dest | string | yes | - | Archive file destination path |
| format | string | no | gz | Archive format (gz, bz2, xz, zip) |
| remove | boolean | no | false | Remove original files after archiving |

## Supported Formats

- **gz**: Gzip compressed tar archive (.tar.gz) - fast compression, good balance
- **bz2**: Bzip2 compressed tar archive (.tar.bz2) - better compression, slower
- **xz**: XZ compressed tar archive (.tar.xz) - best compression, slowest
- **zip**: ZIP archive (.zip) - cross-platform compatibility

## Examples

### Create gzip archive (default)
```yaml
- name: Archive application logs
  module: archive
  vars:
    path: /var/log/myapp
    dest: /backups/logs-2024-01-15.tar.gz
```

### Create bzip2 archive
```yaml
- name: Archive configuration with better compression
  module: archive
  vars:
    path: /etc/myapp
    dest: /backups/config.tar.bz2
    format: bz2
```

### Create zip archive for cross-platform use
```yaml
- name: Archive documents as zip
  module: archive
  vars:
    path: /home/user/project
    dest: /backups/project.zip
    format: zip
```

### Archive and remove originals
```yaml
- name: Archive and cleanup temp files
  module: archive
  vars:
    path: /tmp/build-artifacts
    dest: /archives/build-20240115.tar.gz
    remove: true
```

## Implementation Details

The archive module:
1. Validates that the source path exists
2. Creates the destination directory if it doesn't exist
3. Creates the archive using the appropriate command (tar or zip)
4. Verifies the archive was created successfully
5. Optionally removes the original files if `remove: true`

## Shell Commands Used

- **gz format**: `tar -czf <dest> <path>`
- **bz2 format**: `tar -cjf <dest> <path>`
- **xz format**: `tar -cJf <dest> <path>`
- **zip format**: `zip -r <dest> <path>`

## Notes

- The destination directory will be created automatically if it doesn't exist
- Use `remove: true` carefully - source files are deleted after successful archiving
- Compression ratios vary by format: xz > bz2 > gz > zip (generally)
- Compression speed varies: gz > zip > bz2 > xz (generally)
- Choose format based on your needs: speed vs. compression ratio vs. compatibility
