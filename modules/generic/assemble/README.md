# Assemble Module

Assemble configuration files from multiple fragments into a single file.

## Purpose

This module combines multiple configuration fragments from a directory into a single assembled file. It's useful for:
- Building configuration files from modular components
- Assembling scripts from multiple parts
- Creating composite files from organized fragments
- Managing complex configurations in smaller, maintainable pieces

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| src | string | Yes | - | Directory containing configuration fragments |
| dest | string | Yes | - | Destination file path for assembled configuration |
| delimiter | string | No | "" | Delimiter to insert between fragments |
| regexp | string | No | "" | Only include files matching this regular expression |
| ignore_hidden | boolean | No | true | Ignore files starting with . (hidden files) |

## Examples

### Basic assembly of all fragments

```yaml
- name: Assemble nginx config from fragments
  module: generic/assemble
  vars:
    src: /etc/nginx/conf.d/fragments
    dest: /etc/nginx/nginx.conf
```

This assembles all non-hidden files from the fragments directory into a single config file.

### Assembly with delimiter

```yaml
- name: Assemble script with section markers
  module: generic/assemble
  vars:
    src: /app/script-parts
    dest: /app/deploy.sh
    delimiter: "\n# ---\n"
```

This assembles files with a comment delimiter between each fragment.

### Assembly with regex filter

```yaml
- name: Assemble only .conf files
  module: generic/assemble
  vars:
    src: /etc/app/config-fragments
    dest: /etc/app/app.conf
    regexp: ".*\\.conf$"
```

This only includes files with the .conf extension.

### Include hidden files

```yaml
- name: Assemble all files including hidden
  module: generic/assemble
  vars:
    src: /config/fragments
    dest: /config/complete.conf
    ignore_hidden: false
```

This includes hidden files (starting with .) in the assembly.

### Complex example with all options

```yaml
- name: Assemble service configuration
  module: generic/assemble
  vars:
    src: /etc/service/conf.d
    dest: /etc/service/service.conf
    delimiter: "\n\n# --- Next Section ---\n\n"
    regexp: ".*\\.(conf|config)$"
    ignore_hidden: true
```

This assembles only .conf and .config files, adds section delimiters, and ignores hidden files.

## Implementation Details

This module uses shell commands to find and concatenate files:

**Without delimiter:**
```bash
find {src} -type f ! -name '.*' | sort | xargs cat > {dest}
```

**With delimiter:**
```bash
{ first=true; for f in $(find {src} -type f ! -name '.*' | sort); do
  if [ "$first" = true ]; then first=false; else echo '{delimiter}'; fi;
  cat "$f";
done; } > {dest}
```

**With regex:**
```bash
find {src} -type f -regex '{regexp}' | sort | xargs cat > {dest}
```

## File Ordering

Files are assembled in sorted alphabetical order. To control the assembly order, use numbered prefixes:
- `00-header.conf`
- `10-main.conf`
- `20-footer.conf`

## Important Notes

1. **File Order**: Files are sorted alphabetically before assembly - use numeric prefixes to control order
2. **Hidden Files**: By default, files starting with `.` are ignored
3. **Delimiter**: The delimiter is inserted between fragments, not at the beginning or end
4. **Regex Matching**: The regexp parameter uses find's `-regex` option (extended regex)
5. **Verification**: The module includes a verification step that checks the assembled file exists and shows line count

## Return Value

The module returns shell commands that find fragments, assemble them, and verify the destination file was created.
