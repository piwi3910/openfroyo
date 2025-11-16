# Unarchive Module

Extract compressed archives in various formats (tar.gz, tar.bz2, tar.xz, zip).

## Description

The unarchive module extracts archive files on remote hosts. It supports both local archives (uploaded from the control machine) and remote archives (already present on the target host). The module automatically detects the archive format based on file extension.

## Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| src | string | yes | - | Archive file path (local or remote) |
| dest | string | yes | - | Extraction destination directory |
| remote_src | boolean | no | false | If true, src is on remote host; if false, src is local |
| creates | string | no | - | Skip extraction if this file exists (idempotency) |
| list_files | boolean | no | false | List archive contents before extraction |

## Supported Formats

The module automatically detects format from file extension:

- **.tar.gz, .tgz**: Gzip compressed tar archive
- **.tar.bz2, .tbz2**: Bzip2 compressed tar archive
- **.tar.xz, .txz**: XZ compressed tar archive
- **.tar**: Uncompressed tar archive
- **.zip**: ZIP archive

## Examples

### Extract remote archive
```yaml
- name: Extract application archive
  module: unarchive
  vars:
    src: /tmp/myapp-v1.2.tar.gz
    dest: /opt/myapp
    remote_src: true
```

### Upload and extract local archive
```yaml
- name: Deploy from local archive
  module: unarchive
  vars:
    src: ./dist/release.tar.gz
    dest: /var/www/html
    remote_src: false
```

### Extract with idempotency check
```yaml
- name: Extract configuration (only if not already done)
  module: unarchive
  vars:
    src: /backups/config.tar.gz
    dest: /etc/myapp
    remote_src: true
    creates: /etc/myapp/.extracted
```

### List files before extraction
```yaml
- name: Extract and list contents
  module: unarchive
  vars:
    src: /tmp/data.zip
    dest: /opt/data
    remote_src: true
    list_files: true
```

## Implementation Details

The unarchive module:
1. Checks if the `creates` file exists (if specified) and skips extraction if present
2. For local archives (`remote_src: false`), uploads the archive content to a temp file
3. Verifies the archive file exists on the remote host
4. Creates the destination directory if it doesn't exist
5. Optionally lists archive contents (first 20 entries)
6. Extracts the archive using the appropriate command
7. Verifies extraction by listing the destination directory
8. Cleans up temporary files if created

## Shell Commands Used

- **tar.gz**: `tar -xzf <src> -C <dest>`
- **tar.bz2**: `tar -xjf <src> -C <dest>`
- **tar.xz**: `tar -xJf <src> -C <dest>`
- **tar**: `tar -xf <src> -C <dest>`
- **zip**: `unzip -q <src> -d <dest>`

## Notes

- When `remote_src: false`, the archive is base64 encoded and uploaded to the remote host
- The `creates` parameter enables idempotency - extraction is skipped if the specified file exists
- Large archives may take time to upload when using `remote_src: false`
- The destination directory is created automatically if it doesn't exist
- Format detection is case-insensitive and based on file extension
- Use `list_files: true` to preview archive contents before extraction (shows first 20 entries)

## Executor Integration

The executor must handle local archive uploads when `remote_src: false`:
1. Read the local archive file
2. Base64 encode the content
3. Pass it as `_archive_content` variable to the WASM module
4. The module will decode and write it to a temp file on the remote host
