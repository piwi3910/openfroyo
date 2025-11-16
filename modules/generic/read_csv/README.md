# Read CSV Module

The **read_csv** module reads CSV files from remote hosts and returns their contents as structured data.

## Features

- Read CSV files with custom delimiters
- Parse CSV into JSON array format
- Support for dictionary output using a key column
- Handle initial whitespace in fields
- Cross-platform compatibility

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| path | string | yes | - | Path to the CSV file to read |
| delimiter | string | no | "," | Field delimiter character |
| key | string | no | - | Column name to use as dictionary key |
| skipinitialspace | boolean | no | false | Skip whitespace after delimiter |

## Implementation Details

The read_csv module follows the OpenFroyo pattern of returning `shell_exec` facts:

1. **WASM Module**: Builds shell commands to read and parse the CSV
2. **Shell Commands**: Returns commands that:
   - Check if the CSV file exists
   - Use awk to parse CSV into JSON format
   - Handle custom delimiters and whitespace options

## Usage Examples

### Read a basic CSV file
```yaml
- name: Read user data
  read_csv:
    path: /etc/users.csv
```

### Read TSV file (tab-delimited)
```yaml
- name: Read tab-delimited data
  read_csv:
    path: /tmp/data.tsv
    delimiter: "\t"
```

### Read CSV with key column
```yaml
- name: Read users as dictionary
  read_csv:
    path: /etc/users.csv
    key: username
```

### Read CSV with initial spaces
```yaml
- name: Read spaced CSV
  read_csv:
    path: /tmp/data.csv
    skipinitialspace: true
```

## Testing

Test the module with:

```bash
# Build the module
make build

# Create a test CSV file
echo -e "name,age,city\nAlice,30,NYC\nBob,25,LA" > /tmp/test.csv

# Test reading the CSV
echo -n '{"vars":{"path":"/tmp/test.csv"},"context":{"host":"localhost","task_name":"Test read_csv"}}' | base64 | \
./froyo-runner --module modules/generic/read_csv/wasm/read_csv.wasm --input-base64 -
```

## Output Format

The module returns `shell_exec` facts with awk commands that parse the CSV. The output is a JSON array of objects:

```json
[
  {"name":"Alice","age":"30","city":"NYC"},
  {"name":"Bob","age":"25","city":"LA"}
]
```

## Example Output

```json
{
  "status": "ok",
  "message": "",
  "facts": {
    "shell_exec": [
      {
        "type": "shell",
        "command": "[ -f '/tmp/test.csv' ] || (echo 'CSV_ERROR: File not found' && exit 1)"
      },
      {
        "type": "shell",
        "command": "awk -F',' '...' '/tmp/test.csv'"
      }
    ]
  }
}
```

## Parsing the Output

The executor should:
1. Execute the shell commands in sequence
2. Parse the JSON output from the awk command
3. Store the result in facts for subsequent tasks to use
4. If a `key` parameter was provided, convert the array to a dictionary using that column

## Notes

- CSV parsing is done using awk for maximum portability
- Empty lines in the CSV file are skipped
- The first row is always treated as the header row
- Field values are automatically trimmed if `skipinitialspace` is true
- Special characters in field values are properly escaped
