package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Input represents the module input parameters
type Input struct {
	URL             string            `json:"url"`
	Dest            string            `json:"dest"`
	Checksum        string            `json:"checksum"`
	Mode            string            `json:"mode"`
	Owner           string            `json:"owner"`
	Group           string            `json:"group"`
	Force           bool              `json:"force"`
	Timeout         int               `json:"timeout"`
	Headers         map[string]string `json:"headers"`
	ValidateCerts   bool              `json:"validate_certs"`
	FollowRedirects bool              `json:"follow_redirects"`
	Username        string            `json:"username"`
	Password        string            `json:"password"`
}

// Output represents the module output
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// ShellExec represents a shell command to execute
type ShellExec struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func main() {
	// Read input from stdin
	var input Input
	decoder := json.NewDecoder(strings.NewReader(getInput()))
	if err := decoder.Decode(&input); err != nil {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input: %v", err),
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	// Validate required parameters
	if input.URL == "" {
		output := Output{
			Status:  "failed",
			Message: "URL is required",
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	if input.Dest == "" {
		output := Output{
			Status:  "failed",
			Message: "Destination path (dest) is required",
			Facts:   map[string]interface{}{},
		}
		printOutput(output)
		return
	}

	// Set defaults
	if input.Mode == "" {
		input.Mode = "0644"
	}
	if input.Timeout == 0 {
		input.Timeout = 60
	}

	// Build commands to download file (includes idempotency check)
	commands := buildDownloadCommands(input)

	output := Output{
		Status:  "changed",
		Message: fmt.Sprintf("Downloading %s to %s", input.URL, input.Dest),
		Facts: map[string]interface{}{
			"shell_exec": commands,
			"url":        input.URL,
			"dest":       input.Dest,
		},
	}

	printOutput(output)
}

func buildDownloadCommands(input Input) []ShellExec {
	var commands []ShellExec

	// Check if file exists (idempotency)
	if !input.Force {
		checkCmd := fmt.Sprintf("[ -f '%s' ]", escapeShell(input.Dest))
		if input.Checksum != "" {
			algorithm, expectedHash := parseChecksum(input.Checksum)
			var checksumCmd string
			switch algorithm {
			case "md5":
				checksumCmd = fmt.Sprintf("md5sum '%s' | awk '{print $1}'", escapeShell(input.Dest))
			case "sha1":
				checksumCmd = fmt.Sprintf("sha1sum '%s' | awk '{print $1}'", escapeShell(input.Dest))
			case "sha256":
				checksumCmd = fmt.Sprintf("sha256sum '%s' | awk '{print $1}'", escapeShell(input.Dest))
			case "sha512":
				checksumCmd = fmt.Sprintf("sha512sum '%s' | awk '{print $1}'", escapeShell(input.Dest))
			}
			checkCmd = fmt.Sprintf("[ -f '%s' ] && [ \"$(%s)\" = \"%s\" ] && echo 'File exists with correct checksum, skipping download' && exit 0 || true",
				escapeShell(input.Dest), checksumCmd, expectedHash)
		} else {
			checkCmd = fmt.Sprintf("[ -f '%s' ] && echo 'File exists, skipping download (use force=true to override)' && exit 0 || true",
				escapeShell(input.Dest))
		}
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: checkCmd,
		})
	}

	// Build curl command
	curlCmd := fmt.Sprintf("curl -fsSL --max-time %d", input.Timeout)

	// Add follow redirects
	if input.FollowRedirects {
		curlCmd += " -L"
	}

	// Add SSL validation
	if !input.ValidateCerts {
		curlCmd += " -k"
	}

	// Add custom headers
	for key, value := range input.Headers {
		curlCmd += fmt.Sprintf(" -H '%s: %s'", escapeShell(key), escapeShell(value))
	}

	// Add basic auth if provided
	if input.Username != "" {
		if input.Password != "" {
			curlCmd += fmt.Sprintf(" -u '%s:%s'", escapeShell(input.Username), escapeShell(input.Password))
		} else {
			curlCmd += fmt.Sprintf(" -u '%s'", escapeShell(input.Username))
		}
	}

	// Output to temporary file first
	tmpFile := input.Dest + ".tmp"
	curlCmd += fmt.Sprintf(" -o '%s' '%s'", escapeShell(tmpFile), escapeShell(input.URL))

	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: curlCmd,
	})

	// Verify checksum if provided
	if input.Checksum != "" {
		algorithm, expectedHash := parseChecksum(input.Checksum)
		var checksumCmd string

		switch algorithm {
		case "md5":
			checksumCmd = fmt.Sprintf("echo '%s  %s' | md5sum -c", expectedHash, tmpFile)
		case "sha1":
			checksumCmd = fmt.Sprintf("echo '%s  %s' | sha1sum -c", expectedHash, tmpFile)
		case "sha256":
			checksumCmd = fmt.Sprintf("echo '%s  %s' | sha256sum -c", expectedHash, tmpFile)
		case "sha512":
			checksumCmd = fmt.Sprintf("echo '%s  %s' | sha512sum -c", expectedHash, tmpFile)
		default:
			checksumCmd = fmt.Sprintf("echo 'Warning: Unknown checksum algorithm: %s'", algorithm)
		}

		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: checksumCmd,
		})
	}

	// Move temp file to destination
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: fmt.Sprintf("mv '%s' '%s'", escapeShell(tmpFile), escapeShell(input.Dest)),
	})

	// Set permissions
	commands = append(commands, ShellExec{
		Type:    "shell",
		Command: fmt.Sprintf("chmod %s '%s'", input.Mode, escapeShell(input.Dest)),
	})

	// Set ownership if specified
	if input.Owner != "" && input.Group != "" {
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("chown %s:%s '%s'", escapeShell(input.Owner), escapeShell(input.Group), escapeShell(input.Dest)),
		})
	} else if input.Owner != "" {
		commands = append(commands, ShellExec{
			Type:    "shell",
			Command: fmt.Sprintf("chown %s '%s'", escapeShell(input.Owner), escapeShell(input.Dest)),
		})
	}

	return commands
}

func parseChecksum(checksum string) (algorithm string, hash string) {
	parts := strings.SplitN(checksum, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// Default to sha256 if no algorithm specified
	return "sha256", checksum
}

func extractChecksumValue(checksum string) string {
	_, hash := parseChecksum(checksum)
	return hash
}

func escapeShell(s string) string {
	// Escape single quotes for shell command
	return strings.ReplaceAll(s, "'", "'\\''")
}

func getInput() string {
	// This would normally read from stdin, but in WASM we get it differently
	// For now, return empty string and let the host provide it
	return ""
}

func printOutput(output Output) {
	data, _ := json.Marshal(output)
	fmt.Println(string(data))
}
