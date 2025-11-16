package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

// SONiC connection configuration
type SONiCConfig struct {
	Host          string
	Username      string
	Password      string
	Port          int
	UseREST       bool
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

	// Parse SONiC configuration
	config, err := parseConfig(input.Vars)
	if err != nil {
		outputError(fmt.Sprintf("Configuration error: %v", err))
		return
	}

	// Get command
	command, ok := getStringVar(input.Vars, "command")
	if !ok {
		outputError("Missing required variable: command")
		return
	}

	// Execute the appropriate SONiC operation
	result, err := executeCommand(command, config, input.Vars)
	if err != nil {
		outputError(fmt.Sprintf("Command execution failed: %v", err))
		return
	}

	printOutput(result)
}

func parseConfig(vars map[string]interface{}) (*SONiCConfig, error) {
	config := &SONiCConfig{
		Port:          443,
		UseREST:       true,
		ValidateCerts: false,
	}

	var ok bool
	config.Host, ok = getStringVar(vars, "host")
	if !ok {
		return nil, fmt.Errorf("missing required variable: host")
	}

	config.Username, ok = getStringVar(vars, "username")
	if !ok {
		return nil, fmt.Errorf("missing required variable: username")
	}

	config.Password, ok = getStringVar(vars, "password")
	if !ok {
		return nil, fmt.Errorf("missing required variable: password")
	}

	if port, ok := getIntVar(vars, "port"); ok {
		config.Port = port
	}

	if useREST, ok := getBoolVar(vars, "use_rest"); ok {
		config.UseREST = useREST
	}

	if validateCerts, ok := getBoolVar(vars, "validate_certs"); ok {
		config.ValidateCerts = validateCerts
	}

	return config, nil
}

func executeCommand(command string, config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	switch command {
	// VLAN Management
	case "create_vlan":
		return createVLAN(config, vars)
	case "delete_vlan":
		return deleteVLAN(config, vars)
	case "add_vlan_member":
		return addVLANMember(config, vars)
	case "remove_vlan_member":
		return removeVLANMember(config, vars)
	case "show_vlan":
		return showVLAN(config, vars)

	// Interface Management
	case "configure_interface":
		return configureInterface(config, vars)
	case "set_interface_ip":
		return setInterfaceIP(config, vars)
	case "set_interface_mtu":
		return setInterfaceMTU(config, vars)
	case "set_interface_speed":
		return setInterfaceSpeed(config, vars)
	case "shutdown_interface":
		return shutdownInterface(config, vars)
	case "startup_interface":
		return startupInterface(config, vars)
	case "show_interface":
		return showInterface(config, vars)
	case "show_interface_status":
		return showInterfaceStatus(config, vars)

	// Port Channel/LAG
	case "create_portchannel":
		return createPortChannel(config, vars)
	case "delete_portchannel":
		return deletePortChannel(config, vars)
	case "add_portchannel_member":
		return addPortChannelMember(config, vars)
	case "remove_portchannel_member":
		return removePortChannelMember(config, vars)
	case "show_portchannel":
		return showPortChannel(config, vars)

	// BGP Configuration
	case "configure_bgp":
		return configureBGP(config, vars)
	case "add_bgp_neighbor":
		return addBGPNeighbor(config, vars)
	case "remove_bgp_neighbor":
		return removeBGPNeighbor(config, vars)
	case "configure_bgp_network":
		return configureBGPNetwork(config, vars)
	case "show_bgp_summary":
		return showBGPSummary(config, vars)
	case "show_bgp_neighbors":
		return showBGPNeighbors(config, vars)

	// ACL Configuration
	case "create_acl_table":
		return createACLTable(config, vars)
	case "delete_acl_table":
		return deleteACLTable(config, vars)
	case "add_acl_rule":
		return addACLRule(config, vars)
	case "show_acl":
		return showACL(config, vars)

	// Route Management
	case "add_static_route":
		return addStaticRoute(config, vars)
	case "delete_static_route":
		return deleteStaticRoute(config, vars)
	case "show_ip_route":
		return showIPRoute(config, vars)

	// System Configuration
	case "set_hostname":
		return setHostname(config, vars)
	case "configure_ntp":
		return configureNTP(config, vars)
	case "save_config":
		return saveConfig(config, vars)
	case "show_running_config":
		return showRunningConfig(config, vars)

	default:
		return TaskOutput{}, fmt.Errorf("unknown command: %s", command)
	}
}

// VLAN Management Functions

func createVLAN(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	vlanID, ok := getIntVar(vars, "vlan_id")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: vlan_id")
	}

	if vlanID < 1 || vlanID > 4094 {
		return TaskOutput{}, fmt.Errorf("vlan_id must be between 1 and 4094")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		// Use REST API
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-vlan:sonic-vlan/VLAN/VLAN_LIST", config.Host, config.Port)
		payload := fmt.Sprintf(`{"sonic-vlan:VLAN_LIST":[{"name":"Vlan%d"}]}`, vlanID)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		// Use CLI
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config vlan add %d", vlanID)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Creating VLAN %d", vlanID),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func deleteVLAN(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	vlanID, ok := getIntVar(vars, "vlan_id")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: vlan_id")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-vlan:sonic-vlan/VLAN/VLAN_LIST=Vlan%d", config.Host, config.Port, vlanID)
		curlCmd := buildCurlCommand(config, "DELETE", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config vlan del %d", vlanID)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Deleting VLAN %d", vlanID),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func addVLANMember(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	vlanID, ok := getIntVar(vars, "vlan_id")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: vlan_id")
	}

	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	tagging := "untagged"
	if tagged, ok := getBoolVar(vars, "tagged"); ok && tagged {
		tagging = "tagged"
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-vlan:sonic-vlan/VLAN_MEMBER/VLAN_MEMBER_LIST", config.Host, config.Port)
		payload := fmt.Sprintf(`{"sonic-vlan:VLAN_MEMBER_LIST":[{"name":"Vlan%d","port":"%s","tagging_mode":"%s"}]}`, vlanID, interfaceName, tagging)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		cmdStr := fmt.Sprintf("config vlan member add %d %s", vlanID, interfaceName)
		if tagging == "tagged" {
			cmdStr += " -t"
		}
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, cmdStr),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Adding interface %s to VLAN %d (%s)", interfaceName, vlanID, tagging),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func removeVLANMember(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	vlanID, ok := getIntVar(vars, "vlan_id")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: vlan_id")
	}

	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-vlan:sonic-vlan/VLAN_MEMBER/VLAN_MEMBER_LIST=Vlan%d,%s", config.Host, config.Port, vlanID, interfaceName)
		curlCmd := buildCurlCommand(config, "DELETE", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config vlan member del %d %s", vlanID, interfaceName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Removing interface %s from VLAN %d", interfaceName, vlanID),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showVLAN(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-vlan:sonic-vlan", config.Host, config.Port)
		curlCmd := buildCurlCommand(config, "GET", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, "show vlan brief"),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: "Showing VLAN configuration",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// Interface Management Functions

func configureInterface(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}
	var settings []string

	// Collect all interface settings
	if mtu, ok := getIntVar(vars, "mtu"); ok {
		settings = append(settings, fmt.Sprintf("mtu %d", mtu))
	}
	if speed, ok := getStringVar(vars, "speed"); ok {
		settings = append(settings, fmt.Sprintf("speed %s", speed))
	}
	if adminStatus, ok := getStringVar(vars, "admin_status"); ok {
		if adminStatus == "up" {
			settings = append(settings, "startup")
		} else {
			settings = append(settings, "shutdown")
		}
	}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s/config", config.Host, config.Port, interfaceName)

		configMap := make(map[string]interface{})
		configMap["name"] = interfaceName

		if mtu, ok := getIntVar(vars, "mtu"); ok {
			configMap["mtu"] = mtu
		}
		if adminStatus, ok := getStringVar(vars, "admin_status"); ok {
			if adminStatus == "up" {
				configMap["enabled"] = true
			} else {
				configMap["enabled"] = false
			}
		}

		payloadMap := map[string]interface{}{
			"openconfig-interfaces:config": configMap,
		}
		payload, _ := json.Marshal(payloadMap)

		curlCmd := buildCurlCommand(config, "PATCH", apiURL, string(payload))
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		for _, setting := range settings {
			commands = append(commands, map[string]interface{}{
				"type":    "shell",
				"command": buildSSHCommand(config, fmt.Sprintf("config interface %s %s", setting, interfaceName)),
			})
		}
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Configuring interface %s", interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func setInterfaceIP(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	ipAddress, ok := getStringVar(vars, "ip_address")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: ip_address")
	}

	prefixLength, ok := getIntVar(vars, "prefix_length")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: prefix_length")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s/subinterfaces/subinterface=0/openconfig-if-ip:ipv4/addresses", config.Host, config.Port, interfaceName)
		payload := fmt.Sprintf(`{"openconfig-if-ip:addresses":{"address":[{"ip":"%s","config":{"ip":"%s","prefix-length":%d}}]}}`, ipAddress, ipAddress, prefixLength)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config interface ip add %s %s/%d", interfaceName, ipAddress, prefixLength)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Setting IP address %s/%d on interface %s", ipAddress, prefixLength, interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func setInterfaceMTU(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	mtu, ok := getIntVar(vars, "mtu")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: mtu")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s/config/mtu", config.Host, config.Port, interfaceName)
		payload := fmt.Sprintf(`{"openconfig-interfaces:mtu":%d}`, mtu)

		curlCmd := buildCurlCommand(config, "PUT", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config interface mtu %s %d", interfaceName, mtu)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Setting MTU %d on interface %s", mtu, interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func setInterfaceSpeed(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	speed, ok := getStringVar(vars, "speed")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: speed")
	}

	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("config interface speed %s %s", interfaceName, speed)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Setting speed %s on interface %s", speed, interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func shutdownInterface(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s/config/enabled", config.Host, config.Port, interfaceName)
		payload := `{"openconfig-interfaces:enabled":false}`

		curlCmd := buildCurlCommand(config, "PUT", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config interface shutdown %s", interfaceName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Shutting down interface %s", interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func startupInterface(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s/config/enabled", config.Host, config.Port, interfaceName)
		payload := `{"openconfig-interfaces:enabled":true}`

		curlCmd := buildCurlCommand(config, "PUT", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config interface startup %s", interfaceName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Starting up interface %s", interfaceName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showInterface(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		interfaceName = ""
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces", config.Host, config.Port)
		if interfaceName != "" {
			apiURL = fmt.Sprintf("https://%s:%d/restconf/data/openconfig-interfaces:interfaces/interface=%s", config.Host, config.Port, interfaceName)
		}
		curlCmd := buildCurlCommand(config, "GET", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		cmd := "show interfaces status"
		if interfaceName != "" {
			cmd = fmt.Sprintf("show interfaces status %s", interfaceName)
		}
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, cmd),
		})
	}

	msg := "Showing all interfaces"
	if interfaceName != "" {
		msg = fmt.Sprintf("Showing interface %s", interfaceName)
	}

	return TaskOutput{
		Status:  "ok",
		Message: msg,
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showInterfaceStatus(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, "show interfaces status"),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "Showing interface status",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// Port Channel Functions

func createPortChannel(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	portchannel, ok := getStringVar(vars, "portchannel")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: portchannel")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-portchannel:sonic-portchannel/PORTCHANNEL/PORTCHANNEL_LIST", config.Host, config.Port)
		payload := fmt.Sprintf(`{"sonic-portchannel:PORTCHANNEL_LIST":[{"name":"%s"}]}`, portchannel)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config portchannel add %s", portchannel)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Creating port channel %s", portchannel),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func deletePortChannel(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	portchannel, ok := getStringVar(vars, "portchannel")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: portchannel")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-portchannel:sonic-portchannel/PORTCHANNEL/PORTCHANNEL_LIST=%s", config.Host, config.Port, portchannel)
		curlCmd := buildCurlCommand(config, "DELETE", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config portchannel del %s", portchannel)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Deleting port channel %s", portchannel),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func addPortChannelMember(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	portchannel, ok := getStringVar(vars, "portchannel")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: portchannel")
	}

	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-portchannel:sonic-portchannel/PORTCHANNEL_MEMBER/PORTCHANNEL_MEMBER_LIST", config.Host, config.Port)
		payload := fmt.Sprintf(`{"sonic-portchannel:PORTCHANNEL_MEMBER_LIST":[{"name":"%s","port":"%s"}]}`, portchannel, interfaceName)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config portchannel member add %s %s", portchannel, interfaceName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Adding interface %s to port channel %s", interfaceName, portchannel),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func removePortChannelMember(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	portchannel, ok := getStringVar(vars, "portchannel")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: portchannel")
	}

	interfaceName, ok := getStringVar(vars, "interface")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: interface")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-portchannel:sonic-portchannel/PORTCHANNEL_MEMBER/PORTCHANNEL_MEMBER_LIST=%s,%s", config.Host, config.Port, portchannel, interfaceName)
		curlCmd := buildCurlCommand(config, "DELETE", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config portchannel member del %s %s", portchannel, interfaceName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Removing interface %s from port channel %s", interfaceName, portchannel),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showPortChannel(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-portchannel:sonic-portchannel", config.Host, config.Port)
		curlCmd := buildCurlCommand(config, "GET", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, "show interfaces portchannel"),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: "Showing port channel configuration",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// BGP Functions

func configureBGP(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	bgpASN, ok := getIntVar(vars, "bgp_asn")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: bgp_asn")
	}

	routerID, ok := getStringVar(vars, "router_id")
	if !ok {
		routerID = ""
	}

	var commands []map[string]interface{}

	// BGP configuration typically requires CLI access
	vtyshCmd := fmt.Sprintf("configure terminal; router bgp %d", bgpASN)
	if routerID != "" {
		vtyshCmd += fmt.Sprintf("; bgp router-id %s", routerID)
	}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("vtysh -c \"%s\"", vtyshCmd)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Configuring BGP AS %d", bgpASN),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func addBGPNeighbor(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	bgpASN, ok := getIntVar(vars, "bgp_asn")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: bgp_asn")
	}

	neighborIP, ok := getStringVar(vars, "neighbor_ip")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: neighbor_ip")
	}

	neighborASN, ok := getIntVar(vars, "neighbor_asn")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: neighbor_asn")
	}

	var commands []map[string]interface{}

	vtyshCmd := fmt.Sprintf("configure terminal; router bgp %d; neighbor %s remote-as %d", bgpASN, neighborIP, neighborASN)

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("vtysh -c \"%s\"", vtyshCmd)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Adding BGP neighbor %s (AS %d)", neighborIP, neighborASN),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func removeBGPNeighbor(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	bgpASN, ok := getIntVar(vars, "bgp_asn")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: bgp_asn")
	}

	neighborIP, ok := getStringVar(vars, "neighbor_ip")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: neighbor_ip")
	}

	var commands []map[string]interface{}

	vtyshCmd := fmt.Sprintf("configure terminal; router bgp %d; no neighbor %s", bgpASN, neighborIP)

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("vtysh -c \"%s\"", vtyshCmd)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Removing BGP neighbor %s", neighborIP),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func configureBGPNetwork(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	bgpASN, ok := getIntVar(vars, "bgp_asn")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: bgp_asn")
	}

	network, ok := getStringVar(vars, "network")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: network")
	}

	var commands []map[string]interface{}

	vtyshCmd := fmt.Sprintf("configure terminal; router bgp %d; network %s", bgpASN, network)

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("vtysh -c \"%s\"", vtyshCmd)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Configuring BGP network %s", network),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showBGPSummary(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/openconfig-bgp:bgp", config.Host, config.Port)
		curlCmd := buildCurlCommand(config, "GET", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, "show ip bgp summary"),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: "Showing BGP summary",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showBGPNeighbors(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, "show ip bgp neighbors"),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "Showing BGP neighbors",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// ACL Functions

func createACLTable(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	tableName, ok := getStringVar(vars, "table_name")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: table_name")
	}

	tableType, ok := getStringVar(vars, "table_type")
	if !ok {
		tableType = "L3"
	}

	stage, ok := getStringVar(vars, "stage")
	if !ok {
		stage = "INGRESS"
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-acl:sonic-acl/ACL_TABLE/ACL_TABLE_LIST", config.Host, config.Port)
		payload := fmt.Sprintf(`{"sonic-acl:ACL_TABLE_LIST":[{"aclname":"%s","type":"%s","stage":"%s"}]}`, tableName, tableType, stage)

		curlCmd := buildCurlCommand(config, "POST", apiURL, payload)
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config acl add table %s %s -s %s", tableName, tableType, stage)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Creating ACL table %s", tableName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func deleteACLTable(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	tableName, ok := getStringVar(vars, "table_name")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: table_name")
	}

	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-acl:sonic-acl/ACL_TABLE/ACL_TABLE_LIST=%s", config.Host, config.Port, tableName)
		curlCmd := buildCurlCommand(config, "DELETE", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, fmt.Sprintf("config acl remove table %s", tableName)),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Deleting ACL table %s", tableName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func addACLRule(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	tableName, ok := getStringVar(vars, "table_name")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: table_name")
	}

	ruleName, ok := getStringVar(vars, "rule_name")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: rule_name")
	}

	priority, ok := getIntVar(vars, "priority")
	if !ok {
		priority = 100
	}

	action, ok := getStringVar(vars, "action")
	if !ok {
		action = "FORWARD"
	}

	var commands []map[string]interface{}

	// Build rule configuration
	ruleConfig := fmt.Sprintf("config acl add rule %s %s --priority %d --action %s", tableName, ruleName, priority, action)

	// Add optional match criteria
	if srcIP, ok := getStringVar(vars, "src_ip"); ok {
		ruleConfig += fmt.Sprintf(" --src-ip %s", srcIP)
	}
	if dstIP, ok := getStringVar(vars, "dst_ip"); ok {
		ruleConfig += fmt.Sprintf(" --dst-ip %s", dstIP)
	}
	if protocol, ok := getStringVar(vars, "protocol"); ok {
		ruleConfig += fmt.Sprintf(" --protocol %s", protocol)
	}
	if srcPort, ok := getIntVar(vars, "src_port"); ok {
		ruleConfig += fmt.Sprintf(" --src-port %d", srcPort)
	}
	if dstPort, ok := getIntVar(vars, "dst_port"); ok {
		ruleConfig += fmt.Sprintf(" --dst-port %d", dstPort)
	}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, ruleConfig),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Adding ACL rule %s to table %s", ruleName, tableName),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showACL(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	if config.UseREST {
		apiURL := fmt.Sprintf("https://%s:%d/restconf/data/sonic-acl:sonic-acl", config.Host, config.Port)
		curlCmd := buildCurlCommand(config, "GET", apiURL, "")
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": curlCmd,
		})
	} else {
		commands = append(commands, map[string]interface{}{
			"type":    "shell",
			"command": buildSSHCommand(config, "show acl table"),
		})
	}

	return TaskOutput{
		Status:  "ok",
		Message: "Showing ACL configuration",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// Route Management Functions

func addStaticRoute(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	prefix, ok := getStringVar(vars, "prefix")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: prefix")
	}

	nexthop, ok := getStringVar(vars, "nexthop")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: nexthop")
	}

	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("config route add prefix %s nexthop %s", prefix, nexthop)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Adding static route %s via %s", prefix, nexthop),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func deleteStaticRoute(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	prefix, ok := getStringVar(vars, "prefix")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: prefix")
	}

	nexthop, ok := getStringVar(vars, "nexthop")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: nexthop")
	}

	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("config route del prefix %s nexthop %s", prefix, nexthop)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Deleting static route %s via %s", prefix, nexthop),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showIPRoute(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, "show ip route"),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "Showing IP routing table",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// System Configuration Functions

func setHostname(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	hostname, ok := getStringVar(vars, "hostname")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: hostname")
	}

	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("config hostname %s", hostname)),
	})

	return TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Setting hostname to %s", hostname),
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func configureNTP(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	ntpServer, ok := getStringVar(vars, "ntp_server")
	if !ok {
		return TaskOutput{}, fmt.Errorf("missing required variable: ntp_server")
	}

	action := "add"
	if remove, ok := getBoolVar(vars, "remove"); ok && remove {
		action = "del"
	}

	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, fmt.Sprintf("config ntp %s %s", action, ntpServer)),
	})

	msg := fmt.Sprintf("Adding NTP server %s", ntpServer)
	if action == "del" {
		msg = fmt.Sprintf("Removing NTP server %s", ntpServer)
	}

	return TaskOutput{
		Status:  "ok",
		Message: msg,
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func saveConfig(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, "config save -y"),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "Saving configuration",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

func showRunningConfig(config *SONiCConfig, vars map[string]interface{}) (TaskOutput, error) {
	var commands []map[string]interface{}

	commands = append(commands, map[string]interface{}{
		"type":    "shell",
		"command": buildSSHCommand(config, "show runningconfiguration all"),
	})

	return TaskOutput{
		Status:  "ok",
		Message: "Showing running configuration",
		Facts: map[string]interface{}{
			"shell_exec": commands,
		},
	}, nil
}

// Helper Functions

func buildCurlCommand(config *SONiCConfig, method, url, payload string) string {
	insecure := ""
	if !config.ValidateCerts {
		insecure = " -k"
	}

	auth := fmt.Sprintf("-u %s:%s", config.Username, config.Password)
	headers := "-H 'Content-Type: application/yang-data+json' -H 'Accept: application/yang-data+json'"

	if payload != "" {
		return fmt.Sprintf("curl%s -X %s %s %s -d '%s' %s", insecure, method, auth, headers, payload, url)
	}
	return fmt.Sprintf("curl%s -X %s %s %s %s", insecure, method, auth, headers, url)
}

func buildSSHCommand(config *SONiCConfig, command string) string {
	return fmt.Sprintf("sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s '%s'",
		config.Password, config.Username, config.Host, command)
}

func getStringVar(vars map[string]interface{}, key string) (string, bool) {
	val, ok := vars[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func getIntVar(vars map[string]interface{}, key string) (int, bool) {
	val, ok := vars[key]
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func getBoolVar(vars map[string]interface{}, key string) (bool, bool) {
	val, ok := vars[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
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
