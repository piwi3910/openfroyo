package main

import (
	"encoding/base64"
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

// ShellExec represents a shell command to execute
type ShellExec struct {
	Type    string `json:"type"`
	Command string `json:"command"`
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

	// Extract and validate required parameters
	host := getStringVar(input.Vars, "host")
	username := getStringVar(input.Vars, "username")
	password := getStringVar(input.Vars, "password")
	command := getStringVar(input.Vars, "command")

	if host == "" {
		outputError("Missing required variable: host")
		return
	}
	if username == "" {
		outputError("Missing required variable: username")
		return
	}
	if password == "" {
		outputError("Missing required variable: password")
		return
	}
	if command == "" {
		outputError("Missing required variable: command")
		return
	}

	// Extract optional parameters
	useRest := getBoolVar(input.Vars, "use_rest", true)
	validateCerts := getBoolVar(input.Vars, "validate_certs", false)
	port := getStringVar(input.Vars, "port")
	if port == "" {
		port = "443"
	}
	sshPort := getStringVar(input.Vars, "ssh_port")
	if sshPort == "" {
		sshPort = "22"
	}

	// Process the command
	var shellCommands []ShellExec
	var err2 error
	var message string

	switch command {
	// VLAN Management (5 commands)
	case "create_vlan":
		shellCommands, message, err2 = createVLAN(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "delete_vlan":
		shellCommands, message, err2 = deleteVLAN(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "configure_vlan":
		shellCommands, message, err2 = configureVLAN(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_vlan":
		shellCommands, message, err2 = showVLAN(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_vlan_brief":
		shellCommands, message, err2 = showVLANBrief(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// Interface Management (8 commands)
	case "configure_interface":
		shellCommands, message, err2 = configureInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "shutdown_interface":
		shellCommands, message, err2 = shutdownInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "no_shutdown_interface":
		shellCommands, message, err2 = noShutdownInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "set_description":
		shellCommands, message, err2 = setDescription(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "set_mtu":
		shellCommands, message, err2 = setMTU(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "set_speed":
		shellCommands, message, err2 = setSpeed(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_interface":
		shellCommands, message, err2 = showInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_interface_status":
		shellCommands, message, err2 = showInterfaceStatus(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// Port Channel/LAG (4 commands)
	case "create_port_channel":
		shellCommands, message, err2 = createPortChannel(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "delete_port_channel":
		shellCommands, message, err2 = deletePortChannel(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "add_interface_to_lag":
		shellCommands, message, err2 = addInterfaceToLAG(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "remove_interface_from_lag":
		shellCommands, message, err2 = removeInterfaceFromLAG(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// Layer 3 Configuration (5 commands)
	case "configure_ip_interface":
		shellCommands, message, err2 = configureIPInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "create_vlan_interface":
		shellCommands, message, err2 = createVLANInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "configure_static_route":
		shellCommands, message, err2 = configureStaticRoute(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_ip_interface":
		shellCommands, message, err2 = showIPInterface(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_ip_route":
		shellCommands, message, err2 = showIPRoute(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// Spanning Tree (3 commands)
	case "configure_spanning_tree":
		shellCommands, message, err2 = configureSpanningTree(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "set_stp_priority":
		shellCommands, message, err2 = setSTPPriority(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_spanning_tree":
		shellCommands, message, err2 = showSpanningTree(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// System Configuration (5 commands)
	case "set_hostname":
		shellCommands, message, err2 = setHostname(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "configure_ntp":
		shellCommands, message, err2 = configureNTP(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "configure_dns":
		shellCommands, message, err2 = configureDNS(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "save_config":
		shellCommands, message, err2 = saveConfig(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_running_config":
		shellCommands, message, err2 = showRunningConfig(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// User Management (3 commands)
	case "create_user":
		shellCommands, message, err2 = createUser(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "delete_user":
		shellCommands, message, err2 = deleteUser(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "change_password":
		shellCommands, message, err2 = changePassword(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	// ACL Configuration (3 commands)
	case "create_acl":
		shellCommands, message, err2 = createACL(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "delete_acl":
		shellCommands, message, err2 = deleteACL(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)
	case "show_acl":
		shellCommands, message, err2 = showACL(input.Vars, host, username, password, useRest, validateCerts, port, sshPort)

	default:
		outputError(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	if err2 != nil {
		outputError(err2.Error())
		return
	}

	// Convert shell commands to the format expected by the executor
	shellExecList := make([]map[string]interface{}, len(shellCommands))
	for i, cmd := range shellCommands {
		shellExecList[i] = map[string]interface{}{
			"type":    cmd.Type,
			"command": cmd.Command,
		}
	}

	result := TaskOutput{
		Status:  "changed",
		Message: message,
		Facts: map[string]interface{}{
			"shell_exec": shellExecList,
		},
	}
	printOutput(result)
}

// ============================================================================
// VLAN Management Commands
// ============================================================================

func createVLAN(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")
	vlanName := getStringVar(vars, "vlan_name")

	if vlanID == "" {
		return nil, "", fmt.Errorf("Missing required variable: vlan_id")
	}

	if useRest {
		// REST API approach
		jsonData := fmt.Sprintf(`{"dell-vlan:vlan":[{"vlan-id":%s,"name":"%s"}]}`, vlanID, vlanName)
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"POST", "/restconf/data/dell-vlan:vlans", jsonData)
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			fmt.Sprintf("Created VLAN %s (%s) via REST API", vlanID, vlanName), nil
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface vlan %s", vlanID),
	}
	if vlanName != "" {
		commands = append(commands, fmt.Sprintf("name %s", vlanName))
	}
	commands = append(commands, "exit", "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Created VLAN %s (%s) via SSH CLI", vlanID, vlanName), nil
}

func deleteVLAN(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")

	if vlanID == "" {
		return nil, "", fmt.Errorf("Missing required variable: vlan_id")
	}

	if useRest {
		// REST API approach
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"DELETE", fmt.Sprintf("/restconf/data/dell-vlan:vlans/vlan=%s", vlanID), "")
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			fmt.Sprintf("Deleted VLAN %s via REST API", vlanID), nil
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("no interface vlan %s", vlanID),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Deleted VLAN %s via SSH CLI", vlanID), nil
}

func configureVLAN(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")
	vlanName := getStringVar(vars, "vlan_name")
	description := getStringVar(vars, "vlan_description")

	if vlanID == "" {
		return nil, "", fmt.Errorf("Missing required variable: vlan_id")
	}

	if useRest {
		// REST API approach
		jsonData := fmt.Sprintf(`{"dell-vlan:vlan":[{"vlan-id":%s,"name":"%s","description":"%s"}]}`,
			vlanID, vlanName, description)
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"PUT", fmt.Sprintf("/restconf/data/dell-vlan:vlans/vlan=%s", vlanID), jsonData)
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			fmt.Sprintf("Configured VLAN %s via REST API", vlanID), nil
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface vlan %s", vlanID),
	}
	if vlanName != "" {
		commands = append(commands, fmt.Sprintf("name %s", vlanName))
	}
	if description != "" {
		commands = append(commands, fmt.Sprintf("description %s", description))
	}
	commands = append(commands, "exit", "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured VLAN %s via SSH CLI", vlanID), nil
}

func showVLAN(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")

	if useRest {
		path := "/restconf/data/dell-vlan:vlans"
		if vlanID != "" {
			path = fmt.Sprintf("/restconf/data/dell-vlan:vlans/vlan=%s", vlanID)
		}
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts, "GET", path, "")
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			"Retrieved VLAN information via REST API", nil
	}

	// CLI approach via SSH
	cmd := "show vlan"
	if vlanID != "" {
		cmd = fmt.Sprintf("show vlan id %s", vlanID)
	}
	commands := []string{cmd}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved VLAN information via SSH CLI", nil
}

func showVLANBrief(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	if useRest {
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"GET", "/restconf/data/dell-vlan:vlans", "")
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			"Retrieved VLAN brief information via REST API", nil
	}

	// CLI approach via SSH
	commands := []string{"show vlan brief"}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved VLAN brief information via SSH CLI", nil
}

// ============================================================================
// Interface Management Commands
// ============================================================================

func configureInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	description := getStringVar(vars, "description")
	mtu := getStringVar(vars, "mtu")
	speed := getStringVar(vars, "speed")
	shutdown := getBoolVar(vars, "shutdown", false)
	switchportMode := getStringVar(vars, "switchport_mode")
	accessVLAN := getStringVar(vars, "access_vlan")
	trunkVLANs := getStringVar(vars, "trunk_allowed_vlans")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}

	if useRest {
		// REST API approach
		ifaceName := strings.ReplaceAll(iface, " ", "%2F")
		jsonData := buildInterfaceJSON(description, mtu, speed, !shutdown)
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"PUT", fmt.Sprintf("/restconf/data/ietf-interfaces:interfaces/interface=%s", ifaceName), jsonData)
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			fmt.Sprintf("Configured interface %s via REST API", iface), nil
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
	}
	if description != "" {
		commands = append(commands, fmt.Sprintf("description %s", description))
	}
	if mtu != "" {
		commands = append(commands, fmt.Sprintf("mtu %s", mtu))
	}
	if speed != "" {
		commands = append(commands, fmt.Sprintf("speed %s", speed))
	}
	if shutdown {
		commands = append(commands, "shutdown")
	} else {
		commands = append(commands, "no shutdown")
	}
	if switchportMode != "" {
		commands = append(commands, fmt.Sprintf("switchport mode %s", switchportMode))
		if switchportMode == "access" && accessVLAN != "" {
			commands = append(commands, fmt.Sprintf("switchport access vlan %s", accessVLAN))
		}
		if switchportMode == "trunk" && trunkVLANs != "" {
			commands = append(commands, fmt.Sprintf("switchport trunk allowed vlan %s", trunkVLANs))
		}
	}
	commands = append(commands, "exit", "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured interface %s via SSH CLI", iface), nil
}

func shutdownInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}

	// CLI approach via SSH (simpler for shutdown operation)
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		"shutdown",
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Shutdown interface %s", iface), nil
}

func noShutdownInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		"no shutdown",
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Enabled interface %s", iface), nil
}

func setDescription(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	description := getStringVar(vars, "description")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if description == "" {
		return nil, "", fmt.Errorf("Missing required variable: description")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		fmt.Sprintf("description %s", description),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Set description for interface %s", iface), nil
}

func setMTU(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	mtu := getStringVar(vars, "mtu")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if mtu == "" {
		return nil, "", fmt.Errorf("Missing required variable: mtu")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		fmt.Sprintf("mtu %s", mtu),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Set MTU %s for interface %s", mtu, iface), nil
}

func setSpeed(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	speed := getStringVar(vars, "speed")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if speed == "" {
		return nil, "", fmt.Errorf("Missing required variable: speed")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		fmt.Sprintf("speed %s", speed),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Set speed %s for interface %s", speed, iface), nil
}

func showInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")

	if useRest {
		path := "/restconf/data/ietf-interfaces:interfaces"
		if iface != "" {
			ifaceName := strings.ReplaceAll(iface, " ", "%2F")
			path = fmt.Sprintf("/restconf/data/ietf-interfaces:interfaces/interface=%s", ifaceName)
		}
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts, "GET", path, "")
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			"Retrieved interface information via REST API", nil
	}

	// CLI approach via SSH
	cmd := "show interfaces"
	if iface != "" {
		cmd = fmt.Sprintf("show interface %s", iface)
	}
	commands := []string{cmd}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved interface information via SSH CLI", nil
}

func showInterfaceStatus(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	// CLI approach via SSH
	commands := []string{"show interface status"}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved interface status via SSH CLI", nil
}

// ============================================================================
// Port Channel/LAG Commands
// ============================================================================

func createPortChannel(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	portChannelID := getStringVar(vars, "port_channel_id")
	lagMode := getStringVar(vars, "lag_mode")

	if portChannelID == "" {
		return nil, "", fmt.Errorf("Missing required variable: port_channel_id")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface port-channel %s", portChannelID),
	}
	if lagMode != "" {
		commands = append(commands, fmt.Sprintf("channel-group mode %s", lagMode))
	}
	commands = append(commands, "exit", "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Created port-channel %s", portChannelID), nil
}

func deletePortChannel(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	portChannelID := getStringVar(vars, "port_channel_id")

	if portChannelID == "" {
		return nil, "", fmt.Errorf("Missing required variable: port_channel_id")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("no interface port-channel %s", portChannelID),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Deleted port-channel %s", portChannelID), nil
}

func addInterfaceToLAG(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	portChannelID := getStringVar(vars, "port_channel_id")
	lagMode := getStringVar(vars, "lag_mode")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if portChannelID == "" {
		return nil, "", fmt.Errorf("Missing required variable: port_channel_id")
	}
	if lagMode == "" {
		lagMode = "active"
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		fmt.Sprintf("channel-group %s mode %s", portChannelID, lagMode),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Added interface %s to port-channel %s", iface, portChannelID), nil
}

func removeInterfaceFromLAG(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	portChannelID := getStringVar(vars, "port_channel_id")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if portChannelID == "" {
		return nil, "", fmt.Errorf("Missing required variable: port_channel_id")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		fmt.Sprintf("no channel-group %s", portChannelID),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Removed interface %s from port-channel %s", iface, portChannelID), nil
}

// ============================================================================
// Layer 3 Configuration Commands
// ============================================================================

func configureIPInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")
	ipAddress := getStringVar(vars, "ip_address")
	netmask := getStringVar(vars, "netmask")

	if iface == "" {
		return nil, "", fmt.Errorf("Missing required variable: interface")
	}
	if ipAddress == "" {
		return nil, "", fmt.Errorf("Missing required variable: ip_address")
	}
	if netmask == "" {
		return nil, "", fmt.Errorf("Missing required variable: netmask")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface %s", iface),
		"no switchport",
		fmt.Sprintf("ip address %s %s", ipAddress, netmask),
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured IP address %s/%s on interface %s", ipAddress, netmask, iface), nil
}

func createVLANInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")
	ipAddress := getStringVar(vars, "ip_address")
	netmask := getStringVar(vars, "netmask")

	if vlanID == "" {
		return nil, "", fmt.Errorf("Missing required variable: vlan_id")
	}
	if ipAddress == "" {
		return nil, "", fmt.Errorf("Missing required variable: ip_address")
	}
	if netmask == "" {
		return nil, "", fmt.Errorf("Missing required variable: netmask")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("interface vlan %s", vlanID),
		fmt.Sprintf("ip address %s %s", ipAddress, netmask),
		"no shutdown",
		"exit",
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Created VLAN interface %s with IP %s/%s", vlanID, ipAddress, netmask), nil
}

func configureStaticRoute(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	destination := getStringVar(vars, "destination")
	netmask := getStringVar(vars, "netmask")
	nextHop := getStringVar(vars, "next_hop")

	if destination == "" {
		return nil, "", fmt.Errorf("Missing required variable: destination")
	}
	if netmask == "" {
		return nil, "", fmt.Errorf("Missing required variable: netmask")
	}
	if nextHop == "" {
		return nil, "", fmt.Errorf("Missing required variable: next_hop")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("ip route %s %s %s", destination, netmask, nextHop),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured static route to %s/%s via %s", destination, netmask, nextHop), nil
}

func showIPInterface(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	iface := getStringVar(vars, "interface")

	// CLI approach via SSH
	cmd := "show ip interface"
	if iface != "" {
		cmd = fmt.Sprintf("show ip interface %s", iface)
	}
	commands := []string{cmd}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved IP interface information via SSH CLI", nil
}

func showIPRoute(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	// CLI approach via SSH
	commands := []string{"show ip route"}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved IP routing table via SSH CLI", nil
}

// ============================================================================
// Spanning Tree Commands
// ============================================================================

func configureSpanningTree(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	stpMode := getStringVar(vars, "stp_mode")

	if stpMode == "" {
		stpMode = "rstp"
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("spanning-tree mode %s", stpMode),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured spanning tree mode %s", stpMode), nil
}

func setSTPPriority(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	stpPriority := getStringVar(vars, "stp_priority")
	vlanID := getStringVar(vars, "vlan_id")

	if stpPriority == "" {
		return nil, "", fmt.Errorf("Missing required variable: stp_priority")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
	}
	if vlanID != "" {
		commands = append(commands, fmt.Sprintf("spanning-tree vlan %s priority %s", vlanID, stpPriority))
	} else {
		commands = append(commands, fmt.Sprintf("spanning-tree priority %s", stpPriority))
	}
	commands = append(commands, "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Set spanning tree priority to %s", stpPriority), nil
}

func showSpanningTree(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	vlanID := getStringVar(vars, "vlan_id")

	// CLI approach via SSH
	cmd := "show spanning-tree"
	if vlanID != "" {
		cmd = fmt.Sprintf("show spanning-tree vlan %s", vlanID)
	}
	commands := []string{cmd}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved spanning tree information via SSH CLI", nil
}

// ============================================================================
// System Configuration Commands
// ============================================================================

func setHostname(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	hostname := getStringVar(vars, "hostname")

	if hostname == "" {
		return nil, "", fmt.Errorf("Missing required variable: hostname")
	}

	if useRest {
		jsonData := fmt.Sprintf(`{"dell-system:hostname":"%s"}`, hostname)
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"PUT", "/restconf/data/dell-system:system/hostname", jsonData)
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			fmt.Sprintf("Set hostname to %s via REST API", hostname), nil
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("hostname %s", hostname),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Set hostname to %s via SSH CLI", hostname), nil
}

func configureNTP(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	ntpServer := getStringVar(vars, "ntp_server")

	if ntpServer == "" {
		return nil, "", fmt.Errorf("Missing required variable: ntp_server")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("ntp server %s", ntpServer),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured NTP server %s", ntpServer), nil
}

func configureDNS(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	dnsServer := getStringVar(vars, "dns_server")

	if dnsServer == "" {
		return nil, "", fmt.Errorf("Missing required variable: dns_server")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("ip name-server %s", dnsServer),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Configured DNS server %s", dnsServer), nil
}

func saveConfig(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	// CLI approach via SSH
	commands := []string{"write memory"}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Saved running configuration", nil
}

func showRunningConfig(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	if useRest {
		curlCmd := buildCurlCommand(host, port, username, password, validateCerts,
			"GET", "/restconf/data/dell-system:running-config", "")
		return []ShellExec{{Type: "shell", Command: curlCmd}},
			"Retrieved running configuration via REST API", nil
	}

	// CLI approach via SSH
	commands := []string{"show running-config"}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved running configuration via SSH CLI", nil
}

// ============================================================================
// User Management Commands
// ============================================================================

func createUser(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	newUser := getStringVar(vars, "user")
	newPassword := getStringVar(vars, "new_password")
	role := getStringVar(vars, "role")

	if newUser == "" {
		return nil, "", fmt.Errorf("Missing required variable: user")
	}
	if newPassword == "" {
		return nil, "", fmt.Errorf("Missing required variable: new_password")
	}
	if role == "" {
		role = "sysadmin"
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("username %s password %s role %s", newUser, newPassword, role),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Created user %s with role %s", newUser, role), nil
}

func deleteUser(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	user := getStringVar(vars, "user")

	if user == "" {
		return nil, "", fmt.Errorf("Missing required variable: user")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("no username %s", user),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Deleted user %s", user), nil
}

func changePassword(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	user := getStringVar(vars, "user")
	newPassword := getStringVar(vars, "new_password")

	if user == "" {
		return nil, "", fmt.Errorf("Missing required variable: user")
	}
	if newPassword == "" {
		return nil, "", fmt.Errorf("Missing required variable: new_password")
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("username %s password %s", user, newPassword),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Changed password for user %s", user), nil
}

// ============================================================================
// ACL Configuration Commands
// ============================================================================

func createACL(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	aclName := getStringVar(vars, "acl_name")
	aclType := getStringVar(vars, "acl_type")
	sequence := getStringVar(vars, "sequence")
	action := getStringVar(vars, "action")
	protocol := getStringVar(vars, "protocol")
	source := getStringVar(vars, "source")
	dest := getStringVar(vars, "destination_acl")

	if aclName == "" {
		return nil, "", fmt.Errorf("Missing required variable: acl_name")
	}
	if aclType == "" {
		aclType = "standard"
	}
	if action == "" {
		action = "permit"
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("ip access-list %s %s", aclType, aclName),
	}

	// Build ACL rule
	rule := action
	if sequence != "" {
		rule = fmt.Sprintf("seq %s %s", sequence, action)
	}
	if protocol != "" {
		rule += fmt.Sprintf(" %s", protocol)
	}
	if source != "" {
		rule += fmt.Sprintf(" %s", source)
	}
	if dest != "" {
		rule += fmt.Sprintf(" %s", dest)
	}

	commands = append(commands, rule, "exit", "write memory")

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Created ACL %s", aclName), nil
}

func deleteACL(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	aclName := getStringVar(vars, "acl_name")
	aclType := getStringVar(vars, "acl_type")

	if aclName == "" {
		return nil, "", fmt.Errorf("Missing required variable: acl_name")
	}
	if aclType == "" {
		aclType = "standard"
	}

	// CLI approach via SSH
	commands := []string{
		"configure terminal",
		fmt.Sprintf("no ip access-list %s %s", aclType, aclName),
		"write memory",
	}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		fmt.Sprintf("Deleted ACL %s", aclName), nil
}

func showACL(vars map[string]interface{}, host, username, password string, useRest, validateCerts bool, port, sshPort string) ([]ShellExec, string, error) {
	aclName := getStringVar(vars, "acl_name")

	// CLI approach via SSH
	cmd := "show ip access-lists"
	if aclName != "" {
		cmd = fmt.Sprintf("show ip access-lists %s", aclName)
	}
	commands := []string{cmd}

	sshCmd := buildSSHCommand(host, sshPort, username, password, commands)
	return []ShellExec{{Type: "shell", Command: sshCmd}},
		"Retrieved ACL information via SSH CLI", nil
}

// ============================================================================
// Utility Functions
// ============================================================================

func buildCurlCommand(host, port, username, password string, validateCerts bool, method, path, jsonData string) string {
	insecure := ""
	if !validateCerts {
		insecure = "-k "
	}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	if method == "GET" || method == "DELETE" {
		return fmt.Sprintf(`curl -X %s %s-H "Authorization: Basic %s" -H "Accept: application/json" https://%s:%s%s`,
			method, insecure, auth, host, port, path)
	}

	// POST or PUT with data
	return fmt.Sprintf(`curl -X %s %s-H "Authorization: Basic %s" -H "Content-Type: application/json" -H "Accept: application/json" -d '%s' https://%s:%s%s`,
		method, insecure, auth, jsonData, host, port, path)
}

func buildSSHCommand(host, port, username, password string, commands []string) string {
	// Join commands with semicolon
	cmdString := strings.Join(commands, "; ")

	// Build sshpass command for non-interactive SSH
	return fmt.Sprintf(`sshpass -p '%s' ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p %s %s@%s '%s'`,
		password, port, username, host, cmdString)
}

func buildInterfaceJSON(description, mtu, speed string, enabled bool) string {
	// Build a basic interface configuration JSON
	parts := []string{}

	if description != "" {
		parts = append(parts, fmt.Sprintf(`"description":"%s"`, description))
	}
	if mtu != "" {
		parts = append(parts, fmt.Sprintf(`"mtu":%s`, mtu))
	}
	if speed != "" {
		parts = append(parts, fmt.Sprintf(`"speed":"%s"`, speed))
	}
	parts = append(parts, fmt.Sprintf(`"enabled":%t`, enabled))

	return fmt.Sprintf(`{"ietf-interfaces:interface":{%s}}`, strings.Join(parts, ","))
}

func getStringVar(vars map[string]interface{}, key string) string {
	if val, ok := vars[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

func getBoolVar(vars map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := vars[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "yes" || v == "1"
		case float64:
			return v != 0
		}
	}
	return defaultVal
}

func getIntVar(vars map[string]interface{}, key string, defaultVal int) int {
	if val, ok := vars[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultVal
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
