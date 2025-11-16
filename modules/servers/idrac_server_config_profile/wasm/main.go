package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// TaskInput represents the input to this module
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output from this module
type TaskOutput struct {
	Status  string                 `json:"status"`  // "ok", "changed", "failed"
	Message string                 `json:"message"` // Human-readable message
	Facts   map[string]interface{} `json:"facts"`   // Discovered facts
}

// SCPConfig holds the Server Configuration Profile parameters
type SCPConfig struct {
	IDRACAddress  string
	Username      string
	Password      string
	Command       string
	Timeout       int
	ValidateCerts bool

	// Share configuration
	ShareName     string
	ShareUser     string
	SharePassword string

	// Export configuration
	ExportFormat    string
	ExportUse       string
	SCPComponents   string

	// Import configuration
	SCPFile            string
	ShutdownType       string
	EndHostPowerState  string
	JobWait            bool
	JobWaitTimeout     int
}

func main() {
	// Read input from stdin
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		outputError(fmt.Sprintf("Failed to read stdin: %v", err))
		return
	}

	var input TaskInput
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		outputError(fmt.Sprintf("Failed to parse input JSON: %v", err))
		return
	}

	// Parse configuration from vars
	config, err := parseConfig(input.Vars)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Build the appropriate command based on the operation
	result := buildCommand(config, input.Vars)
	printOutput(result)
}

func parseConfig(vars map[string]interface{}) (*SCPConfig, error) {
	config := &SCPConfig{
		Timeout:            120,
		ValidateCerts:      true,
		ExportFormat:       "XML",
		ExportUse:          "Default",
		SCPComponents:      "ALL",
		ShutdownType:       "Graceful",
		EndHostPowerState:  "On",
		JobWait:            true,
		JobWaitTimeout:     3600,
	}

	// Required parameters
	idracIP, ok := vars["idrac_ip"].(string)
	if !ok || idracIP == "" {
		return nil, fmt.Errorf("missing required variable: idrac_ip")
	}
	config.IDRACAddress = idracIP

	username, ok := vars["idrac_user"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("missing required variable: idrac_user")
	}
	config.Username = username

	password, ok := vars["idrac_password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("missing required variable: idrac_password")
	}
	config.Password = password

	command, ok := vars["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("missing required variable: command (export, import, preview)")
	}
	config.Command = command

	// Optional parameters
	if shareName, ok := vars["share_name"].(string); ok {
		config.ShareName = shareName
	}
	if shareUser, ok := vars["share_user"].(string); ok {
		config.ShareUser = shareUser
	}
	if sharePassword, ok := vars["share_password"].(string); ok {
		config.SharePassword = sharePassword
	}
	if exportFormat, ok := vars["export_format"].(string); ok {
		config.ExportFormat = exportFormat
	}
	if exportUse, ok := vars["export_use"].(string); ok {
		config.ExportUse = exportUse
	}
	if scpComponents, ok := vars["scp_components"].(string); ok {
		config.SCPComponents = scpComponents
	}
	if scpFile, ok := vars["scp_file"].(string); ok {
		config.SCPFile = scpFile
	}
	if shutdownType, ok := vars["shutdown_type"].(string); ok {
		config.ShutdownType = shutdownType
	}
	if endHostPowerState, ok := vars["end_host_power_state"].(string); ok {
		config.EndHostPowerState = endHostPowerState
	}
	if jobWait, ok := vars["job_wait"].(bool); ok {
		config.JobWait = jobWait
	}
	if timeout, ok := vars["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}
	if jobWaitTimeout, ok := vars["job_wait_timeout"].(float64); ok {
		config.JobWaitTimeout = int(jobWaitTimeout)
	}
	if validateCerts, ok := vars["validate_certs"].(bool); ok {
		config.ValidateCerts = validateCerts
	}

	return config, nil
}

func buildCommand(config *SCPConfig, vars map[string]interface{}) TaskOutput {
	// Ensure HTTPS
	baseURL := config.IDRACAddress
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	// Build curl options
	curlOpts := []string{"-s", "-w", "\\nHTTP_CODE:%{http_code}"}
	if !config.ValidateCerts {
		curlOpts = append(curlOpts, "-k")
	}
	curlOpts = append(curlOpts, "-u", fmt.Sprintf("%s:%s", config.Username, config.Password))
	curlOpts = append(curlOpts, "--connect-timeout", strconv.Itoa(config.Timeout))
	curlOpts = append(curlOpts, "--max-time", strconv.Itoa(config.Timeout))

	var script string
	var message string

	switch config.Command {
	case "export":
		script = buildExportSCPCommand(baseURL, curlOpts, config)
		message = fmt.Sprintf("Exporting Server Configuration Profile (%s format, components: %s)", config.ExportFormat, config.SCPComponents)

	case "import":
		if config.SCPFile == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "import command requires scp_file parameter",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildImportSCPCommand(baseURL, curlOpts, config)
		message = fmt.Sprintf("Importing Server Configuration Profile from: %s", config.SCPFile)

	case "preview":
		if config.SCPFile == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "preview command requires scp_file parameter",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildPreviewSCPCommand(baseURL, curlOpts, config)
		message = fmt.Sprintf("Previewing Server Configuration Profile changes from: %s", config.SCPFile)

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s (valid: export, import, preview)", config.Command),
			Facts:   make(map[string]interface{}),
		}
	}

	shellExec := []map[string]interface{}{
		{
			"type":    "shell",
			"command": script,
		},
	}

	return TaskOutput{
		Status:  "ok",
		Message: message,
		Facts: map[string]interface{}{
			"shell_exec": shellExec,
		},
	}
}

func buildExportSCPCommand(baseURL string, curlOpts []string, config *SCPConfig) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration", baseURL)
	opts := strings.Join(curlOpts, " ")

	// Build the export payload
	payload := fmt.Sprintf(`{
  "ExportFormat": "%s",
  "ExportUse": "%s",
  "ShareParameters": {
    "Target": "%s"
  }
}`, config.ExportFormat, config.ExportUse, config.SCPComponents)

	// Determine output file name
	outputFile := config.ShareName
	if outputFile == "" {
		outputFile = fmt.Sprintf("scp_export_%s.%s", config.SCPComponents, strings.ToLower(config.ExportFormat))
	}

	return fmt.Sprintf(`
# Export Server Configuration Profile
echo "Exporting SCP: Format=%s, Components=%s, Use=%s"

# Send export request
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '%s' \
  "%s" 2>&1)

HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" != "202" ] && [ "$HTTP_CODE" != "200" ]; then
    echo "Failed to initiate SCP export. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi

# Extract job URI from response
JOB_URI=$(echo "$BODY" | grep -o '"@odata.id":"[^"]*Jobs/[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$JOB_URI" ]; then
    echo "Warning: Could not extract job URI from response"
    echo "$BODY"
    # Try to extract the SCP content directly if available
    SCP_CONTENT=$(echo "$BODY" | grep -o '"SystemConfiguration"' || echo "$BODY")
    if [ -n "$SCP_CONTENT" ]; then
        echo "SCP export completed inline"
        echo "$BODY" > "%s"
        echo "SCP saved to: %s"
        exit 0
    fi
    exit 1
fi

echo "Job created: $JOB_URI"

# Poll job status
MAX_WAIT=%d
ELAPSED=0
POLL_INTERVAL=10

while [ $ELAPSED -lt $MAX_WAIT ]; do
    sleep $POLL_INTERVAL
    ELAPSED=$((ELAPSED + POLL_INTERVAL))

    JOB_RESPONSE=$(curl %s -H "Content-Type: application/json" "%s$JOB_URI" 2>&1)
    JOB_HTTP_CODE=$(echo "$JOB_RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    JOB_BODY=$(echo "$JOB_RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

    if [ "$JOB_HTTP_CODE" != "200" ]; then
        echo "Failed to get job status. HTTP Code: $JOB_HTTP_CODE"
        continue
    fi

    JOB_STATE=$(echo "$JOB_BODY" | grep -o '"JobState":"[^"]*"' | cut -d'"' -f4)
    PERCENT_COMPLETE=$(echo "$JOB_BODY" | grep -o '"PercentComplete":[0-9]*' | cut -d: -f2)
    MESSAGE=$(echo "$JOB_BODY" | grep -o '"Message":"[^"]*"' | cut -d'"' -f4)

    echo "Job Status: $JOB_STATE ($PERCENT_COMPLETE%%) - $MESSAGE"

    if [ "$JOB_STATE" = "Completed" ]; then
        echo "Export job completed successfully"
        # Get the SCP file URI or content
        FILE_URI=$(echo "$JOB_BODY" | grep -o '"File":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$FILE_URI" ]; then
            # Download the SCP file
            SCP_RESPONSE=$(curl %s -H "Content-Type: application/json" "%s$FILE_URI" 2>&1)
            SCP_HTTP_CODE=$(echo "$SCP_RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
            SCP_BODY=$(echo "$SCP_RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

            if [ "$SCP_HTTP_CODE" = "200" ]; then
                echo "$SCP_BODY" > "%s"
                echo "SCP file downloaded to: %s"
                echo "$SCP_BODY"
            else
                echo "Failed to download SCP file. HTTP Code: $SCP_HTTP_CODE"
                exit 1
            fi
        else
            # SCP content might be in the job response
            echo "$JOB_BODY" > "%s"
            echo "SCP content saved to: %s"
            echo "$JOB_BODY"
        fi
        exit 0
    elif [ "$JOB_STATE" = "Failed" ] || [ "$JOB_STATE" = "CompletedWithErrors" ]; then
        echo "Export job failed: $MESSAGE"
        echo "$JOB_BODY"
        exit 1
    fi
done

echo "Export job timed out after $MAX_WAIT seconds"
exit 1
`, config.ExportFormat, config.SCPComponents, config.ExportUse,
		opts, payload, endpoint,
		outputFile, outputFile,
		config.JobWaitTimeout,
		opts, baseURL,
		opts, baseURL, outputFile, outputFile,
		outputFile, outputFile)
}

func buildImportSCPCommand(baseURL string, curlOpts []string, config *SCPConfig) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration", baseURL)
	opts := strings.Join(curlOpts, " ")

	// Map shutdown types to Dell values
	shutdownType := config.ShutdownType
	switch strings.ToLower(shutdownType) {
	case "graceful":
		shutdownType = "Graceful"
	case "forced":
		shutdownType = "Forced"
	case "noreboot":
		shutdownType = "NoReboot"
	default:
		shutdownType = "Graceful"
	}

	// Map power states
	powerState := config.EndHostPowerState
	switch strings.ToLower(powerState) {
	case "on":
		powerState = "On"
	case "off":
		powerState = "Off"
	default:
		powerState = "On"
	}

	return fmt.Sprintf(`
# Import Server Configuration Profile
echo "Importing SCP from: %s"
echo "Shutdown Type: %s, End Power State: %s"

# Check if SCP file exists
if [ ! -f "%s" ]; then
    echo "Error: SCP file not found: %s"
    exit 1
fi

# Read and base64 encode the SCP file content
SCP_CONTENT=$(cat "%s" | base64 | tr -d '\n')

if [ -z "$SCP_CONTENT" ]; then
    echo "Error: Failed to read SCP file"
    exit 1
fi

# Build import payload
PAYLOAD=$(cat <<'EOF'
{
  "ImportBuffer": "BASE64_PLACEHOLDER",
  "ShareParameters": {
    "Target": "ALL"
  },
  "ShutdownType": "%s",
  "HostPowerState": "%s"
}
EOF
)

# Replace placeholder with actual base64 content
# Using a temporary approach since we can't easily inject large base64 in JSON
TEMP_PAYLOAD=$(mktemp)
cat > "$TEMP_PAYLOAD" <<EOFPAYLOAD
{
  "ImportBuffer": "$SCP_CONTENT",
  "ShareParameters": {
    "Target": "ALL"
  },
  "ShutdownType": "%s",
  "HostPowerState": "%s"
}
EOFPAYLOAD

# Send import request
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d "@$TEMP_PAYLOAD" \
  "%s" 2>&1)

rm -f "$TEMP_PAYLOAD"

HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" != "202" ] && [ "$HTTP_CODE" != "200" ]; then
    echo "Failed to initiate SCP import. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi

echo "Import request submitted successfully"

# Extract job URI from response
JOB_URI=$(echo "$BODY" | grep -o '"@odata.id":"[^"]*Jobs/[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$JOB_URI" ]; then
    echo "Warning: Could not extract job URI from response"
    echo "$BODY"
    # If no job tracking, consider it successful
    if [ "$HTTP_CODE" = "200" ]; then
        echo "Import completed successfully (no job tracking)"
        exit 0
    fi
    exit 1
fi

echo "Configuration job created: $JOB_URI"
`, config.SCPFile, shutdownType, powerState,
		config.SCPFile, config.SCPFile,
		config.SCPFile,
		shutdownType, powerState,
		shutdownType, powerState,
		opts, endpoint) + buildJobWaitScript(baseURL, curlOpts, config)
}

func buildPreviewSCPCommand(baseURL string, curlOpts []string, config *SCPConfig) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview", baseURL)
	opts := strings.Join(curlOpts, " ")

	return fmt.Sprintf(`
# Preview Server Configuration Profile Import
echo "Previewing SCP import from: %s"

# Check if SCP file exists
if [ ! -f "%s" ]; then
    echo "Error: SCP file not found: %s"
    exit 1
fi

# Read and base64 encode the SCP file content
SCP_CONTENT=$(cat "%s" | base64 | tr -d '\n')

if [ -z "$SCP_CONTENT" ]; then
    echo "Error: Failed to read SCP file"
    exit 1
fi

# Build preview payload
TEMP_PAYLOAD=$(mktemp)
cat > "$TEMP_PAYLOAD" <<EOFPAYLOAD
{
  "ImportBuffer": "$SCP_CONTENT",
  "ShareParameters": {
    "Target": "ALL"
  }
}
EOFPAYLOAD

# Send preview request
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d "@$TEMP_PAYLOAD" \
  "%s" 2>&1)

rm -f "$TEMP_PAYLOAD"

HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" != "202" ] && [ "$HTTP_CODE" != "200" ]; then
    echo "Failed to preview SCP import. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi

echo "Preview request submitted successfully"
echo "$BODY"

# Extract job URI if available
JOB_URI=$(echo "$BODY" | grep -o '"@odata.id":"[^"]*Jobs/[^"]*"' | head -1 | cut -d'"' -f4)
if [ -n "$JOB_URI" ]; then
    echo "Preview job created: $JOB_URI"
    # For preview, we still want to wait for results
    `, config.SCPFile,
		config.SCPFile, config.SCPFile,
		config.SCPFile,
		opts, endpoint) + buildJobWaitScript(baseURL, curlOpts, config) + `
else
    echo "Preview completed (no job created)"
    exit 0
fi
`
}

func buildJobWaitScript(baseURL string, curlOpts []string, config *SCPConfig) string {
	if !config.JobWait {
		return `
echo "Job tracking disabled (job_wait=false)"
echo "Job URI: $JOB_URI"
echo "Check job status manually using get_job_queue command"
exit 0
`
	}

	opts := strings.Join(curlOpts, " ")

	return fmt.Sprintf(`
# Wait for job completion
MAX_WAIT=%d
ELAPSED=0
POLL_INTERVAL=15

echo "Waiting for configuration job to complete (timeout: ${MAX_WAIT}s)..."

while [ $ELAPSED -lt $MAX_WAIT ]; do
    sleep $POLL_INTERVAL
    ELAPSED=$((ELAPSED + POLL_INTERVAL))

    JOB_RESPONSE=$(curl %s -H "Content-Type: application/json" "%s$JOB_URI" 2>&1)
    JOB_HTTP_CODE=$(echo "$JOB_RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    JOB_BODY=$(echo "$JOB_RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

    if [ "$JOB_HTTP_CODE" != "200" ]; then
        echo "Failed to get job status. HTTP Code: $JOB_HTTP_CODE"
        continue
    fi

    JOB_STATE=$(echo "$JOB_BODY" | grep -o '"JobState":"[^"]*"' | cut -d'"' -f4)
    PERCENT_COMPLETE=$(echo "$JOB_BODY" | grep -o '"PercentComplete":[0-9]*' | cut -d: -f2)
    MESSAGE=$(echo "$JOB_BODY" | grep -o '"Message":"[^"]*"' | cut -d'"' -f4)

    echo "[${ELAPSED}s] Job Status: $JOB_STATE ($PERCENT_COMPLETE%%) - $MESSAGE"

    if [ "$JOB_STATE" = "Completed" ]; then
        echo ""
        echo "Configuration job completed successfully!"
        echo "Final job details:"
        echo "$JOB_BODY"
        exit 0
    elif [ "$JOB_STATE" = "Failed" ] || [ "$JOB_STATE" = "CompletedWithErrors" ]; then
        echo ""
        echo "Configuration job failed!"
        echo "Error details:"
        echo "$JOB_BODY"
        exit 1
    elif [ "$JOB_STATE" = "Scheduled" ] || [ "$JOB_STATE" = "Running" ] || [ "$JOB_STATE" = "Starting" ]; then
        # Job is still in progress, continue waiting
        continue
    else
        echo "Unexpected job state: $JOB_STATE"
    fi
done

echo ""
echo "Configuration job timed out after $MAX_WAIT seconds"
echo "Job may still be running on iDRAC"
echo "Check job status manually: $JOB_URI"
exit 1
`, config.JobWaitTimeout, opts, baseURL)
}

func outputError(msg string) {
	result := TaskOutput{
		Status:  "failed",
		Message: msg,
		Facts:   make(map[string]interface{}),
	}
	printOutput(result)
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.Marshal(output)
	fmt.Println(string(outputJSON))
}
