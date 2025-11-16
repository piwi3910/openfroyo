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

// iDRACConfig holds the connection parameters
type iDRACConfig struct {
	BaseURI       string
	Username      string
	Password      string
	Command       string
	Timeout       int
	ValidateCerts bool
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

func parseConfig(vars map[string]interface{}) (*iDRACConfig, error) {
	config := &iDRACConfig{
		Timeout:       30,
		ValidateCerts: true,
	}

	// Required parameters
	baseURI, ok := vars["baseuri"].(string)
	if !ok || baseURI == "" {
		return nil, fmt.Errorf("missing required variable: baseuri (iDRAC IP/hostname)")
	}
	config.BaseURI = baseURI

	username, ok := vars["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("missing required variable: username")
	}
	config.Username = username

	password, ok := vars["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("missing required variable: password")
	}
	config.Password = password

	command, ok := vars["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("missing required variable: command")
	}
	config.Command = command

	// Optional parameters
	if timeout, ok := vars["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}

	if validateCerts, ok := vars["validate_certs"].(bool); ok {
		config.ValidateCerts = validateCerts
	}

	return config, nil
}

func buildCommand(config *iDRACConfig, vars map[string]interface{}) TaskOutput {
	// Ensure HTTPS
	baseURL := config.BaseURI
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

	var script string
	var message string

	switch config.Command {
	// Power Management
	case "get_power_state":
		script = buildGetPowerStateCommand(baseURL, curlOpts)
		message = "Getting power state from iDRAC"

	case "power_on":
		script = buildPowerCommand(baseURL, curlOpts, "On", "ForceOn")
		message = "Powering on system"

	case "power_off":
		script = buildPowerCommand(baseURL, curlOpts, "ForceOff", "ForceOff")
		message = "Powering off system"

	case "graceful_shutdown":
		script = buildPowerCommand(baseURL, curlOpts, "GracefulShutdown", "GracefulShutdown")
		message = "Gracefully shutting down system"

	case "power_cycle":
		script = buildPowerCommand(baseURL, curlOpts, "ForceRestart", "ForceRestart")
		message = "Power cycling system"

	case "force_restart":
		script = buildPowerCommand(baseURL, curlOpts, "ForceRestart", "ForceRestart")
		message = "Force restarting system"

	// Virtual Media
	case "insert_virtual_media":
		mediaType, _ := vars["media_type"].(string)
		imageURL, _ := vars["image_url"].(string)
		if mediaType == "" || imageURL == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "insert_virtual_media requires media_type (CD, DVD, Floppy) and image_url",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildInsertVirtualMediaCommand(baseURL, curlOpts, mediaType, imageURL)
		message = fmt.Sprintf("Inserting virtual media: %s from %s", mediaType, imageURL)

	case "eject_virtual_media":
		mediaType, _ := vars["media_type"].(string)
		if mediaType == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "eject_virtual_media requires media_type (CD, DVD, Floppy)",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildEjectVirtualMediaCommand(baseURL, curlOpts, mediaType)
		message = fmt.Sprintf("Ejecting virtual media: %s", mediaType)

	// System Configuration
	case "reset_idrac":
		script = buildResetIDRACCommand(baseURL, curlOpts)
		message = "Resetting iDRAC"

	case "get_system_inventory":
		script = buildGetSystemInventoryCommand(baseURL, curlOpts)
		message = "Getting system inventory"

	case "get_sel":
		script = buildGetSELCommand(baseURL, curlOpts)
		message = "Getting System Event Log"

	case "clear_sel":
		script = buildClearSELCommand(baseURL, curlOpts)
		message = "Clearing System Event Log"

	// User Management
	case "create_user":
		username, _ := vars["new_username"].(string)
		newPassword, _ := vars["new_password"].(string)
		privilege, _ := vars["privilege"].(string)
		if username == "" || newPassword == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "create_user requires new_username, new_password, and optionally privilege",
				Facts:   make(map[string]interface{}),
			}
		}
		if privilege == "" {
			privilege = "Operator"
		}
		script = buildCreateUserCommand(baseURL, curlOpts, username, newPassword, privilege)
		message = fmt.Sprintf("Creating user: %s with privilege: %s", username, privilege)

	case "delete_user":
		username, _ := vars["target_username"].(string)
		if username == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "delete_user requires target_username",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildDeleteUserCommand(baseURL, curlOpts, username)
		message = fmt.Sprintf("Deleting user: %s", username)

	case "change_password":
		username, _ := vars["target_username"].(string)
		newPassword, _ := vars["new_password"].(string)
		if username == "" || newPassword == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "change_password requires target_username and new_password",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildChangePasswordCommand(baseURL, curlOpts, username, newPassword)
		message = fmt.Sprintf("Changing password for user: %s", username)

	// Network Configuration
	case "configure_network":
		ipAddress, _ := vars["ip_address"].(string)
		netmask, _ := vars["netmask"].(string)
		gateway, _ := vars["gateway"].(string)
		dhcpEnabled, _ := vars["dhcp_enabled"].(bool)
		script = buildConfigureNetworkCommand(baseURL, curlOpts, ipAddress, netmask, gateway, dhcpEnabled)
		if dhcpEnabled {
			message = "Configuring iDRAC network: DHCP enabled"
		} else {
			message = fmt.Sprintf("Configuring iDRAC network: IP=%s, Netmask=%s, Gateway=%s", ipAddress, netmask, gateway)
		}

	case "configure_dns":
		dns1, _ := vars["dns1"].(string)
		dns2, _ := vars["dns2"].(string)
		dnsFromDHCP, _ := vars["dns_from_dhcp"].(bool)
		script = buildConfigureDNSCommand(baseURL, curlOpts, dns1, dns2, dnsFromDHCP)
		message = "Configuring DNS settings"

	case "configure_ntp":
		ntp1, _ := vars["ntp1"].(string)
		ntp2, _ := vars["ntp2"].(string)
		ntpEnabled, _ := vars["ntp_enabled"].(bool)
		script = buildConfigureNTPCommand(baseURL, curlOpts, ntp1, ntp2, ntpEnabled)
		message = "Configuring NTP settings"

	// Lifecycle Controller
	case "get_firmware_inventory":
		script = buildGetFirmwareInventoryCommand(baseURL, curlOpts)
		message = "Getting firmware inventory"

	case "update_firmware":
		imageURI, _ := vars["image_uri"].(string)
		if imageURI == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "update_firmware requires image_uri",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildUpdateFirmwareCommand(baseURL, curlOpts, imageURI)
		message = fmt.Sprintf("Updating firmware from: %s", imageURI)

	case "get_job_queue":
		script = buildGetJobQueueCommand(baseURL, curlOpts)
		message = "Getting job queue"

	case "delete_job":
		jobID, _ := vars["job_id"].(string)
		if jobID == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "delete_job requires job_id",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildDeleteJobCommand(baseURL, curlOpts, jobID)
		message = fmt.Sprintf("Deleting job: %s", jobID)

	// RAID Configuration
	case "get_storage_controllers":
		script = buildGetStorageControllersCommand(baseURL, curlOpts)
		message = "Getting storage controllers"

	case "get_virtual_disks":
		controllerID, _ := vars["controller_id"].(string)
		if controllerID == "" {
			controllerID = "RAID.Integrated.1-1"
		}
		script = buildGetVirtualDisksCommand(baseURL, curlOpts, controllerID)
		message = fmt.Sprintf("Getting virtual disks for controller: %s", controllerID)

	case "get_physical_disks":
		controllerID, _ := vars["controller_id"].(string)
		if controllerID == "" {
			controllerID = "RAID.Integrated.1-1"
		}
		script = buildGetPhysicalDisksCommand(baseURL, curlOpts, controllerID)
		message = fmt.Sprintf("Getting physical disks for controller: %s", controllerID)

	case "create_virtual_disk":
		controllerID, _ := vars["controller_id"].(string)
		raidLevel, _ := vars["raid_level"].(string)
		disks, _ := vars["disks"].(string)
		vdName, _ := vars["vd_name"].(string)
		if controllerID == "" || raidLevel == "" || disks == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "create_virtual_disk requires controller_id, raid_level, and disks (comma-separated)",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildCreateVirtualDiskCommand(baseURL, curlOpts, controllerID, raidLevel, disks, vdName)
		message = fmt.Sprintf("Creating virtual disk: RAID %s on controller %s", raidLevel, controllerID)

	case "delete_virtual_disk":
		vdID, _ := vars["vd_id"].(string)
		if vdID == "" {
			return TaskOutput{
				Status:  "failed",
				Message: "delete_virtual_disk requires vd_id",
				Facts:   make(map[string]interface{}),
			}
		}
		script = buildDeleteVirtualDiskCommand(baseURL, curlOpts, vdID)
		message = fmt.Sprintf("Deleting virtual disk: %s", vdID)

	default:
		return TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", config.Command),
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

// Power Management Commands

func buildGetPowerStateCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get power state from iDRAC
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    POWER_STATE=$(echo "$BODY" | grep -o '"PowerState":"[^"]*' | cut -d'"' -f4)
    echo "Power State: $POWER_STATE"
    echo "$BODY"
else
    echo "Failed to get power state. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildPowerCommand(baseURL string, curlOpts []string, resetType, targetState string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Execute power command: %s
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType":"%s"}' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "Power command '%s' executed successfully"
else
    echo "Failed to execute power command. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, resetType, opts, resetType, endpoint, resetType)
}

// Virtual Media Commands

func buildInsertVirtualMediaCommand(baseURL string, curlOpts []string, mediaType, imageURL string) string {
	// Map media type to Redfish virtual media ID
	var mediaID string
	switch strings.ToUpper(mediaType) {
	case "CD", "DVD":
		mediaID = "CD"
	case "FLOPPY":
		mediaID = "Floppy"
	default:
		mediaID = "CD"
	}

	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia/%s/Actions/VirtualMedia.InsertMedia", baseURL, mediaID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Insert virtual media: %s
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '{"Image":"%s","Inserted":true}' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "Virtual media inserted successfully: %s"
else
    echo "Failed to insert virtual media. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, mediaType, opts, imageURL, endpoint, imageURL)
}

func buildEjectVirtualMediaCommand(baseURL string, curlOpts []string, mediaType string) string {
	var mediaID string
	switch strings.ToUpper(mediaType) {
	case "CD", "DVD":
		mediaID = "CD"
	case "FLOPPY":
		mediaID = "Floppy"
	default:
		mediaID = "CD"
	}

	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia/%s/Actions/VirtualMedia.EjectMedia", baseURL, mediaID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Eject virtual media: %s
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "Virtual media ejected successfully"
else
    echo "Failed to eject virtual media. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, mediaType, opts, endpoint)
}

// System Configuration Commands

func buildResetIDRACCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Reset iDRAC
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType":"GracefulRestart"}' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "iDRAC reset initiated successfully"
else
    echo "Failed to reset iDRAC. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildGetSystemInventoryCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get system inventory
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "System Inventory:"
    echo "$BODY" | grep -E '"(Manufacturer|Model|SerialNumber|BiosVersion|PowerState|ProcessorSummary|MemorySummary)"' || echo "$BODY"
else
    echo "Failed to get system inventory. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildGetSELCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get System Event Log
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "System Event Log:"
    echo "$BODY"
else
    echo "Failed to get SEL. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildClearSELCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Actions/LogService.ClearLog", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Clear System Event Log
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "System Event Log cleared successfully"
else
    echo "Failed to clear SEL. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

// User Management Commands

func buildCreateUserCommand(baseURL string, curlOpts []string, username, password, privilege string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Accounts", baseURL)
	opts := strings.Join(curlOpts, " ")

	// Map privilege levels
	roleID := "Operator"
	switch privilege {
	case "Administrator", "Admin":
		roleID = "Administrator"
	case "Operator":
		roleID = "Operator"
	case "ReadOnly", "Read":
		roleID = "ReadOnly"
	}

	return fmt.Sprintf(`
# Create user: %s
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '{"UserName":"%s","Password":"%s","RoleId":"%s","Enabled":true}' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "User created successfully: %s"
else
    echo "Failed to create user. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, username, opts, username, password, roleID, endpoint, username)
}

func buildDeleteUserCommand(baseURL string, curlOpts []string, username string) string {
	// Note: In production, we'd need to look up the account ID first
	// For now, we'll use a placeholder approach
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Accounts", baseURL)
	opts := strings.Join(curlOpts, " ")
	_ = opts // Will be used when account lookup is implemented
	return fmt.Sprintf(`
# Delete user: %s
# First, find the account ID
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    # Look for the account with matching username
    # This is simplified - in production would parse JSON properly
    echo "Found accounts list, user deletion requires manual account ID lookup"
    echo "$BODY"
else
    echo "Failed to list accounts. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, username, strings.Join(curlOpts, " "), endpoint)
}

func buildChangePasswordCommand(baseURL string, curlOpts []string, username, password string) string {
	// Similar to delete - would need account ID lookup
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Accounts", baseURL)
	_ = curlOpts // Will be used when account lookup is implemented
	_ = password // Will be used when account lookup is implemented
	return fmt.Sprintf(`
# Change password for user: %s
echo "Password change requires account ID lookup - see documentation"
echo "Endpoint: %s"
`, username, endpoint)
}

// Network Configuration Commands

func buildConfigureNetworkCommand(baseURL string, curlOpts []string, ipAddress, netmask, gateway string, dhcpEnabled bool) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces/NIC.1", baseURL)
	opts := strings.Join(curlOpts, " ")

	var payload string
	if dhcpEnabled {
		payload = `{"DHCPv4":{"DHCPEnabled":true}}`
	} else {
		payload = fmt.Sprintf(`{"IPv4Addresses":[{"Address":"%s","SubnetMask":"%s","Gateway":"%s"}]}`, ipAddress, netmask, gateway)
	}

	return fmt.Sprintf(`
# Configure iDRAC network
RESPONSE=$(curl %s -X PATCH \
  -H "Content-Type: application/json" \
  -d '%s' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Network configured successfully"
else
    echo "Failed to configure network. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, payload, endpoint)
}

func buildConfigureDNSCommand(baseURL string, curlOpts []string, dns1, dns2 string, dnsFromDHCP bool) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", baseURL)
	opts := strings.Join(curlOpts, " ")

	var payload string
	if dnsFromDHCP {
		payload = `{"Attributes":{"IPv4.1.DNSFromDHCP":"Enabled"}}`
	} else {
		payload = fmt.Sprintf(`{"Attributes":{"IPv4.1.DNS1":"%s","IPv4.1.DNS2":"%s","IPv4.1.DNSFromDHCP":"Disabled"}}`, dns1, dns2)
	}

	return fmt.Sprintf(`
# Configure DNS
RESPONSE=$(curl %s -X PATCH \
  -H "Content-Type: application/json" \
  -d '%s' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "DNS configured successfully"
else
    echo "Failed to configure DNS. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, payload, endpoint)
}

func buildConfigureNTPCommand(baseURL string, curlOpts []string, ntp1, ntp2 string, ntpEnabled bool) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Attributes", baseURL)
	opts := strings.Join(curlOpts, " ")

	enabledStr := "Disabled"
	if ntpEnabled {
		enabledStr = "Enabled"
	}

	payload := fmt.Sprintf(`{"Attributes":{"NTPEnable":"%s","NTP1":"%s","NTP2":"%s"}}`, enabledStr, ntp1, ntp2)

	return fmt.Sprintf(`
# Configure NTP
RESPONSE=$(curl %s -X PATCH \
  -H "Content-Type: application/json" \
  -d '%s' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "NTP configured successfully"
else
    echo "Failed to configure NTP. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, payload, endpoint)
}

// Lifecycle Controller Commands

func buildGetFirmwareInventoryCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/UpdateService/FirmwareInventory", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get firmware inventory
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Firmware Inventory:"
    echo "$BODY"
else
    echo "Failed to get firmware inventory. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildUpdateFirmwareCommand(baseURL string, curlOpts []string, imageURI string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Update firmware
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '{"ImageURI":"%s","TransferProtocol":"HTTP"}' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "202" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "Firmware update initiated successfully"
    echo "$BODY"
else
    echo "Failed to update firmware. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, imageURI, endpoint)
}

func buildGetJobQueueCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Jobs", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get job queue
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Job Queue:"
    echo "$BODY"
else
    echo "Failed to get job queue. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildDeleteJobCommand(baseURL string, curlOpts []string, jobID string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Jobs/%s", baseURL, jobID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Delete job: %s
RESPONSE=$(curl %s -X DELETE \
  -H "Content-Type: application/json" \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    echo "Job deleted successfully: %s"
else
    echo "Failed to delete job. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, jobID, opts, endpoint, jobID)
}

// RAID Configuration Commands

func buildGetStorageControllersCommand(baseURL string, curlOpts []string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Storage", baseURL)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get storage controllers
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Storage Controllers:"
    echo "$BODY"
else
    echo "Failed to get storage controllers. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, opts, endpoint)
}

func buildGetVirtualDisksCommand(baseURL string, curlOpts []string, controllerID string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Storage/%s/Volumes", baseURL, controllerID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get virtual disks for controller: %s
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Virtual Disks:"
    echo "$BODY"
else
    echo "Failed to get virtual disks. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, controllerID, opts, endpoint)
}

func buildGetPhysicalDisksCommand(baseURL string, curlOpts []string, controllerID string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Storage/%s", baseURL, controllerID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Get physical disks for controller: %s
RESPONSE=$(curl %s -H "Content-Type: application/json" "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ]; then
    echo "Physical Disks:"
    echo "$BODY" | grep -A 10 "Drives" || echo "$BODY"
else
    echo "Failed to get physical disks. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, controllerID, opts, endpoint)
}

func buildCreateVirtualDiskCommand(baseURL string, curlOpts []string, controllerID, raidLevel, disks, vdName string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Storage/%s/Volumes", baseURL, controllerID)
	opts := strings.Join(curlOpts, " ")

	// Build drives array
	diskArray := strings.Split(disks, ",")
	drivesJSON := "["
	for i, disk := range diskArray {
		if i > 0 {
			drivesJSON += ","
		}
		drivesJSON += fmt.Sprintf(`{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Storage/%s/Drives/%s"}`, controllerID, strings.TrimSpace(disk))
	}
	drivesJSON += "]"

	name := vdName
	if name == "" {
		name = fmt.Sprintf("VD-RAID%s", raidLevel)
	}

	payload := fmt.Sprintf(`{"VolumeType":"RAID%s","Name":"%s","Drives":%s}`, raidLevel, name, drivesJSON)

	return fmt.Sprintf(`
# Create virtual disk: RAID %s
RESPONSE=$(curl %s -X POST \
  -H "Content-Type: application/json" \
  -d '%s' \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "202" ] || [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    echo "Virtual disk creation initiated successfully"
    echo "$BODY"
else
    echo "Failed to create virtual disk. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, raidLevel, opts, payload, endpoint)
}

func buildDeleteVirtualDiskCommand(baseURL string, curlOpts []string, vdID string) string {
	endpoint := fmt.Sprintf("%s/redfish/v1/Systems/System.Embedded.1/Storage/Volumes/%s", baseURL, vdID)
	opts := strings.Join(curlOpts, " ")
	return fmt.Sprintf(`
# Delete virtual disk: %s
RESPONSE=$(curl %s -X DELETE \
  -H "Content-Type: application/json" \
  "%s" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed 's/HTTP_CODE:[0-9]*$//')

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "202" ]; then
    echo "Virtual disk deleted successfully: %s"
else
    echo "Failed to delete virtual disk. HTTP Code: $HTTP_CODE"
    echo "$BODY"
    exit 1
fi
`, vdID, opts, endpoint, vdID)
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
