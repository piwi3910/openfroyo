package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TaskInput represents the input from OpenFroyo
type TaskInput struct {
	Vars    map[string]interface{} `json:"vars"`
	Context TaskContext            `json:"context"`
}

// TaskContext provides execution context
type TaskContext struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// TaskOutput represents the output to OpenFroyo
type TaskOutput struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

func main() {
	var input TaskInput

	// Read input from stdin
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Validate required variables
	bmcHost, ok := input.Vars["bmc_host"].(string)
	if !ok || bmcHost == "" {
		output := TaskOutput{
			Status:  "failed",
			Message: "Required variable 'bmc_host' missing or empty",
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	username, ok := input.Vars["username"].(string)
	if !ok || username == "" {
		output := TaskOutput{
			Status:  "failed",
			Message: "Required variable 'username' missing or empty",
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	password, ok := input.Vars["password"].(string)
	if !ok || password == "" {
		output := TaskOutput{
			Status:  "failed",
			Message: "Required variable 'password' missing or empty",
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	command, ok := input.Vars["command"].(string)
	if !ok || command == "" {
		output := TaskOutput{
			Status:  "failed",
			Message: "Required variable 'command' missing or empty",
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Optional parameters with defaults
	interfaceType := getStringVar(input.Vars, "interface", "lanplus")
	privilege := getStringVar(input.Vars, "privilege", "ADMINISTRATOR")

	// Build IPMI script
	script, err := buildIPMIScript(bmcHost, username, password, interfaceType, privilege, command, input.Vars)
	if err != nil {
		output := TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to build IPMI command: %v", err),
			Facts:   make(map[string]interface{}),
		}
		printOutput(output)
		return
	}

	// Return shell_exec fact to execute the IPMI command
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Executing IPMI command: %s", command),
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": fmt.Sprintf("/bin/bash -c %s", shellescape(script)),
				},
			},
		},
	}

	printOutput(result)
}

func buildIPMIScript(bmcHost, username, password, interfaceType, privilege, command string, vars map[string]interface{}) (string, error) {
	var scriptBuilder strings.Builder

	// Script header
	scriptBuilder.WriteString("#!/bin/bash\n")
	scriptBuilder.WriteString("set -e\n\n")

	// Check for ipmitool
	scriptBuilder.WriteString("if ! command -v ipmitool >/dev/null 2>&1; then\n")
	scriptBuilder.WriteString("  echo 'ERROR: ipmitool is not installed'\n")
	scriptBuilder.WriteString("  exit 1\n")
	scriptBuilder.WriteString("fi\n\n")

	// Set variables
	scriptBuilder.WriteString(fmt.Sprintf("BMC_HOST='%s'\n", bmcHost))
	scriptBuilder.WriteString(fmt.Sprintf("USERNAME='%s'\n", username))
	scriptBuilder.WriteString(fmt.Sprintf("PASSWORD='%s'\n", password))
	scriptBuilder.WriteString(fmt.Sprintf("INTERFACE='%s'\n", interfaceType))
	scriptBuilder.WriteString(fmt.Sprintf("PRIVILEGE='%s'\n\n", privilege))

	// Base command
	scriptBuilder.WriteString("IPMI_BASE=\"ipmitool -I $INTERFACE -H $BMC_HOST -U $USERNAME -P $PASSWORD -L $PRIVILEGE\"\n\n")

	// Build command-specific logic
	switch command {
	// Power Management
	case "power_on":
		scriptBuilder.WriteString("$IPMI_BASE chassis power on\n")
	case "power_off":
		scriptBuilder.WriteString("$IPMI_BASE chassis power off\n")
	case "power_cycle":
		scriptBuilder.WriteString("$IPMI_BASE chassis power cycle\n")
	case "power_reset":
		scriptBuilder.WriteString("$IPMI_BASE chassis power reset\n")
	case "power_soft":
		scriptBuilder.WriteString("$IPMI_BASE chassis power soft\n")
	case "power_status":
		scriptBuilder.WriteString("$IPMI_BASE chassis power status\n")

	// Chassis Management
	case "chassis_status":
		scriptBuilder.WriteString("$IPMI_BASE chassis status\n")
	case "chassis_identify_on":
		duration := getIntVar(vars, "duration", 15)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE chassis identify %d\n", duration))
	case "chassis_identify_off":
		scriptBuilder.WriteString("$IPMI_BASE chassis identify 0\n")
	case "chassis_selftest":
		scriptBuilder.WriteString("$IPMI_BASE chassis selftest\n")
	case "set_boot_device":
		bootDevice := getStringVar(vars, "boot_device", "")
		if bootDevice == "" {
			return "", fmt.Errorf("boot_device required for set_boot_device command")
		}
		persistent := getBoolVar(vars, "persistent", false)
		options := "options=efiboot"
		if persistent {
			options += ",persistent"
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE chassis bootdev %s %s\n", bootDevice, options))
	case "get_boot_device":
		scriptBuilder.WriteString("$IPMI_BASE chassis bootparam get 5\n")

	// Sensor Monitoring
	case "get_all_sensors":
		scriptBuilder.WriteString("$IPMI_BASE sensor list\n")
	case "get_sensor":
		sensorName := getStringVar(vars, "sensor_name", "")
		if sensorName == "" {
			return "", fmt.Errorf("sensor_name required for get_sensor command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sensor get '%s'\n", sensorName))
	case "get_sensor_thresholds":
		sensorName := getStringVar(vars, "sensor_name", "")
		if sensorName == "" {
			return "", fmt.Errorf("sensor_name required for get_sensor_thresholds command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sensor thresh '%s'\n", sensorName))

	// SEL (System Event Log)
	case "get_sel":
		scriptBuilder.WriteString("$IPMI_BASE sel list\n")
	case "get_sel_info":
		scriptBuilder.WriteString("$IPMI_BASE sel info\n")
	case "clear_sel":
		scriptBuilder.WriteString("$IPMI_BASE sel clear\n")
	case "save_sel":
		filepath := getStringVar(vars, "filepath", "/tmp/sel.log")
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sel save %s\n", filepath))
	case "get_sel_elist":
		scriptBuilder.WriteString("$IPMI_BASE sel elist\n")
	case "get_sel_time":
		scriptBuilder.WriteString("$IPMI_BASE sel time get\n")

	// FRU (Field Replaceable Unit)
	case "get_fru":
		scriptBuilder.WriteString("$IPMI_BASE fru print\n")
	case "get_fru_id":
		fruID := getIntVar(vars, "fru_id", 0)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE fru print %d\n", fruID))

	// SOL (Serial Over LAN)
	case "sol_activate":
		scriptBuilder.WriteString("$IPMI_BASE sol activate\n")
	case "sol_deactivate":
		scriptBuilder.WriteString("$IPMI_BASE sol deactivate\n")
	case "sol_info":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sol info %d\n", channel))
	case "sol_set":
		channel := getIntVar(vars, "channel", 1)
		param := getStringVar(vars, "param", "")
		value := getStringVar(vars, "value", "")
		if param == "" || value == "" {
			return "", fmt.Errorf("param and value required for sol_set command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sol set %s %s %d\n", param, value, channel))

	// User Management
	case "list_users":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE user list %d\n", channel))
	case "set_user_name":
		userID := getIntVar(vars, "user_id", 0)
		userName := getStringVar(vars, "user_name", "")
		if userName == "" {
			return "", fmt.Errorf("user_name required for set_user_name command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE user set name %d '%s'\n", userID, userName))
	case "set_user_password":
		userID := getIntVar(vars, "user_id", 0)
		userPassword := getStringVar(vars, "user_password", "")
		if userPassword == "" {
			return "", fmt.Errorf("user_password required for set_user_password command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE user set password %d '%s'\n", userID, userPassword))
	case "enable_user":
		userID := getIntVar(vars, "user_id", 0)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE user enable %d\n", userID))
	case "disable_user":
		userID := getIntVar(vars, "user_id", 0)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE user disable %d\n", userID))
	case "set_user_privilege":
		userID := getIntVar(vars, "user_id", 0)
		channel := getIntVar(vars, "channel", 1)
		userPrivilege := getStringVar(vars, "user_privilege", "")
		if userPrivilege == "" {
			return "", fmt.Errorf("user_privilege required for set_user_privilege command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE channel setaccess %d %d privilege=%s\n", channel, userID, userPrivilege))

	// LAN Configuration
	case "get_lan_config":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan print %d\n", channel))
	case "set_lan_ip_static":
		channel := getIntVar(vars, "channel", 1)
		ipAddress := getStringVar(vars, "ip_address", "")
		subnetMask := getStringVar(vars, "subnet_mask", "")
		gateway := getStringVar(vars, "gateway", "")
		if ipAddress == "" || subnetMask == "" || gateway == "" {
			return "", fmt.Errorf("ip_address, subnet_mask, and gateway required for set_lan_ip_static")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d ipsrc static\n", channel))
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d ipaddr %s\n", channel, ipAddress))
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d netmask %s\n", channel, subnetMask))
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d defgw ipaddr %s\n", channel, gateway))
	case "set_lan_ip_dhcp":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d ipsrc dhcp\n", channel))
	case "set_lan_vlan":
		channel := getIntVar(vars, "channel", 1)
		vlanID := getIntVar(vars, "vlan_id", 0)
		if vlanID == 0 {
			scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d vlan id off\n", channel))
		} else {
			scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d vlan id %d\n", channel, vlanID))
		}
	case "set_lan_auth":
		channel := getIntVar(vars, "channel", 1)
		level := getStringVar(vars, "auth_level", "")
		authType := getStringVar(vars, "auth_type", "")
		if level == "" || authType == "" {
			return "", fmt.Errorf("auth_level and auth_type required for set_lan_auth")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE lan set %d auth %s %s\n", channel, level, authType))

	// SDR (Sensor Data Record)
	case "get_sdr_info":
		scriptBuilder.WriteString("$IPMI_BASE sdr info\n")
	case "get_sdr_list":
		scriptBuilder.WriteString("$IPMI_BASE sdr list\n")
	case "get_sdr_type":
		sdrType := getStringVar(vars, "sdr_type", "")
		if sdrType == "" {
			return "", fmt.Errorf("sdr_type required for get_sdr_type command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sdr type '%s'\n", sdrType))
	case "get_sdr_entity":
		entityID := getStringVar(vars, "entity_id", "")
		if entityID == "" {
			return "", fmt.Errorf("entity_id required for get_sdr_entity command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE sdr entity '%s'\n", entityID))

	// Firmware and BIOS
	case "get_bmc_info":
		scriptBuilder.WriteString("$IPMI_BASE mc info\n")
	case "get_device_id":
		scriptBuilder.WriteString("$IPMI_BASE mc info | grep 'Device ID'\n")
	case "get_firmware_version":
		scriptBuilder.WriteString("$IPMI_BASE mc info | grep 'Firmware Revision'\n")
	case "reset_bmc_cold":
		scriptBuilder.WriteString("$IPMI_BASE mc reset cold\n")
	case "reset_bmc_warm":
		scriptBuilder.WriteString("$IPMI_BASE mc reset warm\n")
	case "get_guid":
		scriptBuilder.WriteString("$IPMI_BASE mc guid\n")
	case "get_selftest":
		scriptBuilder.WriteString("$IPMI_BASE mc selftest\n")

	// Watchdog
	case "get_watchdog":
		scriptBuilder.WriteString("$IPMI_BASE mc watchdog get\n")
	case "reset_watchdog":
		scriptBuilder.WriteString("$IPMI_BASE mc watchdog reset\n")
	case "set_watchdog":
		action := getStringVar(vars, "watchdog_action", "reset")
		timeout := getIntVar(vars, "watchdog_timeout", 60)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE mc watchdog set %s %d\n", action, timeout))
	case "off_watchdog":
		scriptBuilder.WriteString("$IPMI_BASE mc watchdog off\n")

	// PEF (Platform Event Filtering)
	case "get_pef":
		scriptBuilder.WriteString("$IPMI_BASE pef\n")

	// Raw Commands
	case "raw":
		rawCommand := getStringVar(vars, "raw_command", "")
		if rawCommand == "" {
			return "", fmt.Errorf("raw_command required for raw command")
		}
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE raw %s\n", rawCommand))

	// Channel Commands
	case "get_channel_info":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE channel info %d\n", channel))
	case "get_channel_authcap":
		channel := getIntVar(vars, "channel", 1)
		scriptBuilder.WriteString(fmt.Sprintf("$IPMI_BASE channel authcap %d 4\n", channel))

	// Session Management
	case "get_session_info":
		scriptBuilder.WriteString("$IPMI_BASE session info all\n")

	default:
		return "", fmt.Errorf("unknown IPMI command: %s", command)
	}

	scriptBuilder.WriteString("\necho \"EXIT_CODE=$?\"\n")

	return scriptBuilder.String(), nil
}

func getStringVar(vars map[string]interface{}, key, defaultValue string) string {
	if val, ok := vars[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntVar(vars map[string]interface{}, key string, defaultValue int) int {
	if val, ok := vars[key].(float64); ok {
		return int(val)
	}
	if val, ok := vars[key].(int); ok {
		return val
	}
	return defaultValue
}

func getBoolVar(vars map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := vars[key].(bool); ok {
		return val
	}
	return defaultValue
}

func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func printOutput(output TaskOutput) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
}
