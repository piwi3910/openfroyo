# Find Files Module

Search for files and directories matching specified criteria.

## Overview

The `find` module searches for files, directories, and symbolic links in a directory tree based on various criteria such as name patterns, file type, age, size, and more.

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `path` | string | Yes | - | Directory path to search in |
| `patterns` | string | No | - | File name patterns to match (shell glob, e.g., '*.txt' or '*.log') |
| `file_type` | string | No | `any` | Type of files to find: `file`, `directory`, `link`, or `any` |
| `age` | string | No | - | Find files by age: `+N` (older than N days), `-N` (newer than N days) |
| `age_stamp` | string | No | `mtime` | Age stamp to check: `mtime` (modified), `atime` (accessed), `ctime` (changed) |
| `size` | string | No | - | Find files by size: `+N` (larger than), `-N` (smaller than), `N` (exact). Units: c(bytes), k(KB), M(MB), G(GB) |
| `recurse` | boolean | No | `true` | Recursively search subdirectories |
| `hidden` | boolean | No | `false` | Include hidden files (starting with .) |
| `depth` | string | No | - | Maximum depth to search (e.g., '2' for two levels deep) |

## Behavior

### File Type Filtering

- `file` - Find regular files only
- `directory` - Find directories only
- `link` - Find symbolic links only
- `any` - Find all types (default)

### Age Filtering

- `+7` - Files older than 7 days
- `-7` - Files newer than 7 days
- Age is measured in days (24-hour periods)

### Size Filtering

- `+1M` - Files larger than 1 megabyte
- `-100k` - Files smaller than 100 kilobytes
- `1G` - Files exactly 1 gigabyte
- Units: `c` (bytes), `k` (KB), `M` (MB), `G` (GB)

## Examples

### Find all files recursively

```yaml
- name: Find all files
  module: generic/find
  vars:
    path: /var/log
    file_type: file
```

### Find specific file patterns

```yaml
- name: Find all log files
  module: generic/find
  vars:
    path: /var/log
    patterns: "*.log"
    file_type: file
```

### Find directories only

```yaml
- name: Find all subdirectories
  module: generic/find
  vars:
    path: /var
    file_type: directory
```

### Non-recursive search

```yaml
- name: Find files in current directory only
  module: generic/find
  vars:
    path: /var/log
    file_type: file
    recurse: false
```

### Search with depth limit

```yaml
- name: Find files up to 2 levels deep
  module: generic/find
  vars:
    path: /var
    file_type: file
    depth: "2"
```

### Find files by age

```yaml
- name: Find recently modified files
  module: generic/find
  vars:
    path: /var/log
    file_type: file
    age: "-7"
    age_stamp: mtime
```

### Find files by size

```yaml
- name: Find large files
  module: generic/find
  vars:
    path: /var/log
    file_type: file
    size: "+100M"
```

### Find old, large log files

```yaml
- name: Find old large logs for cleanup
  module: generic/find
  vars:
    path: /var/log
    patterns: "*.log"
    file_type: file
    age: "+30"
    size: "+50M"
```

### Include hidden files

```yaml
- name: Find all files including hidden
  module: generic/find
  vars:
    path: /home/user
    file_type: file
    hidden: true
```

### Find symbolic links

```yaml
- name: Find all symbolic links
  module: generic/find
  vars:
    path: /usr/bin
    file_type: link
```

## Implementation

This module returns `shell_exec` facts containing a `find` command with appropriate flags:

1. Verifies the search path exists
2. Constructs find command with filters for:
   - File type (`-type f|d|l`)
   - Name patterns (`-name 'pattern'`)
   - Age (`-mtime`, `-atime`, `-ctime`)
   - Size (`-size`)
   - Depth (`-maxdepth`)
   - Hidden files (exclusion pattern)
3. Outputs the list of matching files
4. Counts the total matches

The runner executes these shell commands on the target host.

## Testing

Build the module:

```bash
cd modules/generic/find
make build
```

Run tests:

```bash
froyo apply modules/generic/find/test.ofy
```

## Notes

- Pattern matching uses shell glob syntax (`*`, `?`, `[]`)
- Hidden files are excluded by default unless `hidden=true`
- When `recurse=false`, only searches the immediate directory (equivalent to `-maxdepth 1`)
- Age is always measured in 24-hour periods (days)
- Size units follow standard find conventions: c (bytes), k (kilobytes), M (megabytes), G (gigabytes)
- The module outputs both the file list and a count of matches
