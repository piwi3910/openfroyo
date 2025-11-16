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

	// Validate required variables
	iloIP, err := getStringVar(input.Vars, "ilo_ip", true)
	if err != nil {
		outputError(err.Error())
		return
	}

	iloUser, err := getStringVar(input.Vars, "ilo_user", true)
	if err != nil {
		outputError(err.Error())
		return
	}

	iloPassword, err := getStringVar(input.Vars, "ilo_password", true)
	if err != nil {
		outputError(err.Error())
		return
	}

	command, err := getStringVar(input.Vars, "command", true)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Optional variables
	timeout := getIntVar(input.Vars, "timeout", 60)
	validateCerts := getBoolVar(input.Vars, "validate_certs", true)

	// Build the appropriate command based on operation
	shellCmd, err := buildILOCommand(iloIP, iloUser, iloPassword, command, timeout, validateCerts, input.Vars)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Return shell_exec fact for the executor to run
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Executing iLO command: %s on %s", command, iloIP),
		Facts: map[string]interface{}{
			"shell_exec": []map[string]interface{}{
				{
					"type":    "shell",
					"command": shellCmd,
				},
			},
		},
	}
	printOutput(result)
}

func buildILOCommand(iloIP, username, password, command string, timeout int, validateCerts bool, vars map[string]interface{}) (string, error) {
	baseURI := fmt.Sprintf("https://%s", iloIP)
	curlOpts := "-s -w '\\nHTTP_STATUS:%{http_code}'"
	if !validateCerts {
		curlOpts += " -k"
	}

	var cmdParts []string
	cmdParts = append(cmdParts, "#!/bin/bash")
	cmdParts = append(cmdParts, "set -e")
	cmdParts = append(cmdParts, "")

	switch strings.ToLower(command) {
	// Power Management Operations
	case "power_on":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"On"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "power_off":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"ForceOff"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "graceful_shutdown":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"GracefulShutdown"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "force_restart":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"ForceRestart"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "power_button":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"PushPowerButton"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "cold_boot":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Reset","ResetType":"ColdBoot"}' "%s/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"`, curlOpts, username, password, baseURI))

	case "get_power_state":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq -r '.PowerState'`, curlOpts, username, password, baseURI))

	// Virtual Media Operations
	case "insert_virtual_cd", "insert_virtual_dvd":
		isoURL, err := getStringVar(vars, "iso_url", true)
		if err != nil {
			return "", fmt.Errorf("iso_url required for insert_virtual_cd: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Image":"%s","Inserted":true}' "%s/redfish/v1/Managers/1/VirtualMedia/2/Actions/VirtualMedia.InsertMedia"`, curlOpts, username, password, isoURL, baseURI))

	case "eject_virtual_cd", "eject_virtual_dvd":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" "%s/redfish/v1/Managers/1/VirtualMedia/2/Actions/VirtualMedia.EjectMedia"`, curlOpts, username, password, baseURI))

	case "insert_virtual_floppy":
		isoURL, err := getStringVar(vars, "iso_url", true)
		if err != nil {
			return "", fmt.Errorf("iso_url required for insert_virtual_floppy: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Image":"%s","Inserted":true}' "%s/redfish/v1/Managers/1/VirtualMedia/1/Actions/VirtualMedia.InsertMedia"`, curlOpts, username, password, isoURL, baseURI))

	case "eject_virtual_floppy":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" "%s/redfish/v1/Managers/1/VirtualMedia/1/Actions/VirtualMedia.EjectMedia"`, curlOpts, username, password, baseURI))

	case "get_virtual_media_status":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1/VirtualMedia"`, curlOpts, username, password, baseURI))

	// System Information Operations
	case "get_system_info":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1"`, curlOpts, username, password, baseURI))

	case "get_health_status":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq -r '.Status'`, curlOpts, username, password, baseURI))

	case "get_hardware_inventory":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq '{ProcessorSummary,MemorySummary,Storage,NetworkInterfaces}'`, curlOpts, username, password, baseURI))

	case "get_firmware_versions":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/UpdateService/FirmwareInventory"`, curlOpts, username, password, baseURI))

	case "get_serial_number":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq -r '.SerialNumber'`, curlOpts, username, password, baseURI))

	case "get_product_info":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq '{Model,Manufacturer,SerialNumber,UUID}'`, curlOpts, username, password, baseURI))

	// User Management Operations
	case "get_users":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/AccountService/Accounts"`, curlOpts, username, password, baseURI))

	case "create_user":
		newUser, err := getStringVar(vars, "new_username", true)
		if err != nil {
			return "", fmt.Errorf("new_username required for create_user: %v", err)
		}
		newPass, err := getStringVar(vars, "new_password", true)
		if err != nil {
			return "", fmt.Errorf("new_password required for create_user: %v", err)
		}
		privilege := getStringVarDefault(vars, "privilege", "Operator")

		// Map privilege levels to iLO privileges
		iloConfigPriv := "false"
		userConfigPriv := "false"
		if privilege == "Administrator" {
			iloConfigPriv = "true"
			userConfigPriv = "true"
		}

		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"UserName":"%s","Password":"%s","Oem":{"Hp":{"Privileges":{"RemoteConsolePriv":true,"iLOConfigPriv":%s,"VirtualMediaPriv":true,"UserConfigPriv":%s,"VirtualPowerAndResetPriv":true}}}}' "%s/redfish/v1/AccountService/Accounts"`, curlOpts, username, password, newUser, newPass, iloConfigPriv, userConfigPriv, baseURI))

	case "delete_user":
		userID, err := getStringVar(vars, "user_id", true)
		if err != nil {
			return "", fmt.Errorf("user_id required for delete_user: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X DELETE -u "%s:%s" "%s/redfish/v1/AccountService/Accounts/%s"`, curlOpts, username, password, baseURI, userID))

	case "modify_user":
		userID, err := getStringVar(vars, "user_id", true)
		if err != nil {
			return "", fmt.Errorf("user_id required for modify_user: %v", err)
		}
		newPass, _ := getStringVar(vars, "new_password", false)

		patchData := "{"
		if newPass != "" {
			patchData += fmt.Sprintf(`"Password":"%s"`, newPass)
		}
		patchData += "}"

		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '%s' "%s/redfish/v1/AccountService/Accounts/%s"`, curlOpts, username, password, patchData, baseURI, userID))

	// Network Configuration Operations
	case "get_network_settings":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1/EthernetInterfaces/1"`, curlOpts, username, password, baseURI))

	case "set_network_static":
		ipAddr, err := getStringVar(vars, "ip_address", true)
		if err != nil {
			return "", fmt.Errorf("ip_address required for set_network_static: %v", err)
		}
		subnetMask, err := getStringVar(vars, "subnet_mask", true)
		if err != nil {
			return "", fmt.Errorf("subnet_mask required for set_network_static: %v", err)
		}
		gateway, err := getStringVar(vars, "gateway", true)
		if err != nil {
			return "", fmt.Errorf("gateway required for set_network_static: %v", err)
		}

		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"IPv4Addresses":[{"Address":"%s","SubnetMask":"%s","Gateway":"%s"}]}' "%s/redfish/v1/Managers/1/EthernetInterfaces/1"`, curlOpts, username, password, ipAddr, subnetMask, gateway, baseURI))

	case "set_network_dhcp":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"DHCPv4":{"DHCPEnabled":true}}' "%s/redfish/v1/Managers/1/EthernetInterfaces/1"`, curlOpts, username, password, baseURI))

	case "set_hostname":
		hostname, err := getStringVar(vars, "hostname", true)
		if err != nil {
			return "", fmt.Errorf("hostname required for set_hostname: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"HostName":"%s"}' "%s/redfish/v1/Managers/1/EthernetInterfaces/1"`, curlOpts, username, password, hostname, baseURI))

	case "set_dns_servers":
		dnsServers, err := getStringVar(vars, "dns_servers", true)
		if err != nil {
			return "", fmt.Errorf("dns_servers required for set_dns_servers: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"Oem":{"Hp":{"IPv4":{"DNSServers":[%s]}}}}' "%s/redfish/v1/Managers/1/EthernetInterfaces/1"`, curlOpts, username, password, dnsServers, baseURI))

	// Firmware Management Operations
	case "get_firmware_inventory":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/UpdateService/FirmwareInventory"`, curlOpts, username, password, baseURI))

	case "update_firmware":
		firmwareURL, err := getStringVar(vars, "firmware_url", true)
		if err != nil {
			return "", fmt.Errorf("firmware_url required for update_firmware: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"ImageURI":"%s"}' "%s/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"`, curlOpts, username, password, firmwareURL, baseURI))

	case "get_update_progress":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/TaskService/Tasks"`, curlOpts, username, password, baseURI))

	// Boot Configuration Operations
	case "set_onetime_boot":
		bootDevice, err := getStringVar(vars, "boot_device", true)
		if err != nil {
			return "", fmt.Errorf("boot_device required for set_onetime_boot (Cd, Hdd, Pxe, BiosSetup): %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"Boot":{"BootSourceOverrideEnabled":"Once","BootSourceOverrideTarget":"%s"}}' "%s/redfish/v1/Systems/1"`, curlOpts, username, password, bootDevice, baseURI))

	case "set_persistent_boot":
		bootDevice, err := getStringVar(vars, "boot_device", true)
		if err != nil {
			return "", fmt.Errorf("boot_device required for set_persistent_boot (Cd, Hdd, Pxe, BiosSetup): %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"Boot":{"BootSourceOverrideEnabled":"Continuous","BootSourceOverrideTarget":"%s"}}' "%s/redfish/v1/Systems/1"`, curlOpts, username, password, bootDevice, baseURI))

	case "get_boot_config":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1" | jq -r '.Boot'`, curlOpts, username, password, baseURI))

	case "set_uefi_boot":
		enabled := getBoolVar(vars, "enabled", true)
		bootMode := "Uefi"
		if !enabled {
			bootMode = "Legacy"
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"Boot":{"BootSourceOverrideMode":"%s"}}' "%s/redfish/v1/Systems/1"`, curlOpts, username, password, bootMode, baseURI))

	// Logs and Events Operations
	case "get_iml_log":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Systems/1/LogServices/IML/Entries"`, curlOpts, username, password, baseURI))

	case "clear_iml_log":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"LogService.ClearLog"}' "%s/redfish/v1/Systems/1/LogServices/IML/Actions/LogService.ClearLog"`, curlOpts, username, password, baseURI))

	case "get_ilo_event_log":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1/LogServices/IEL/Entries"`, curlOpts, username, password, baseURI))

	case "clear_ilo_event_log":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"LogService.ClearLog"}' "%s/redfish/v1/Managers/1/LogServices/IEL/Actions/LogService.ClearLog"`, curlOpts, username, password, baseURI))

	// License Management Operations
	case "get_license_status":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1/LicenseService"`, curlOpts, username, password, baseURI))

	case "install_license":
		licenseKey, err := getStringVar(vars, "license_key", true)
		if err != nil {
			return "", fmt.Errorf("license_key required for install_license: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"LicenseKey":"%s"}' "%s/redfish/v1/Managers/1/LicenseService/Actions/HpeiLOLicense.InstallLicense"`, curlOpts, username, password, licenseKey, baseURI))

	// iLO Settings Operations
	case "reset_ilo":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"Manager.Reset","ResetType":"GracefulRestart"}' "%s/redfish/v1/Managers/1/Actions/Manager.Reset"`, curlOpts, username, password, baseURI))

	case "get_ilo_datetime":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1" | jq -r '.DateTime'`, curlOpts, username, password, baseURI))

	case "set_ilo_datetime":
		datetime, err := getStringVar(vars, "datetime", true)
		if err != nil {
			return "", fmt.Errorf("datetime required for set_ilo_datetime (ISO8601 format): %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X PATCH -u "%s:%s" -H "Content-Type: application/json" -d '{"DateTime":"%s"}' "%s/redfish/v1/Managers/1"`, curlOpts, username, password, datetime, baseURI))

	case "get_ilo_info":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1"`, curlOpts, username, password, baseURI))

	case "ping_from_ilo":
		targetHost, err := getStringVar(vars, "target_host", true)
		if err != nil {
			return "", fmt.Errorf("target_host required for ping_from_ilo: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Target":"%s"}' "%s/redfish/v1/Managers/1/Actions/Oem/Hp/HpiLO.NetworkTest"`, curlOpts, username, password, targetHost, baseURI))

	// Security Operations
	case "get_security_dashboard":
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X GET -u "%s:%s" "%s/redfish/v1/Managers/1/SecurityService"`, curlOpts, username, password, baseURI))

	case "generate_csr":
		commonName, err := getStringVar(vars, "common_name", true)
		if err != nil {
			return "", fmt.Errorf("common_name required for generate_csr: %v", err)
		}
		orgName := getStringVarDefault(vars, "organization", "")
		orgUnit := getStringVarDefault(vars, "organizational_unit", "")
		locality := getStringVarDefault(vars, "locality", "")
		state := getStringVarDefault(vars, "state", "")
		country := getStringVarDefault(vars, "country", "")

		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Action":"GenerateCSR","CommonName":"%s","OrgName":"%s","OrgUnit":"%s","Locality":"%s","State":"%s","Country":"%s"}' "%s/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.GenerateCSR"`, curlOpts, username, password, commonName, orgName, orgUnit, locality, state, country, baseURI))

	case "import_certificate":
		certificate, err := getStringVar(vars, "certificate", true)
		if err != nil {
			return "", fmt.Errorf("certificate required for import_certificate: %v", err)
		}
		cmdParts = append(cmdParts, fmt.Sprintf(`curl %s -X POST -u "%s:%s" -H "Content-Type: application/json" -d '{"Certificate":"%s"}' "%s/redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.ImportCertificate"`, curlOpts, username, password, strings.ReplaceAll(certificate, "\n", "\\n"), baseURI))

	default:
		return "", fmt.Errorf("unsupported command: %s", command)
	}

	return strings.Join(cmdParts, "\n"), nil
}

func getStringVar(vars map[string]interface{}, key string, required bool) (string, error) {
	val, ok := vars[key]
	if !ok {
		if required {
			return "", fmt.Errorf("missing required variable: %s", key)
		}
		return "", nil
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("variable '%s' must be a string", key)
	}

	return strVal, nil
}

func getStringVarDefault(vars map[string]interface{}, key string, defaultVal string) string {
	val, err := getStringVar(vars, key, false)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

func getIntVar(vars map[string]interface{}, key string, defaultVal int) int {
	val, ok := vars[key]
	if !ok {
		return defaultVal
	}

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

	return defaultVal
}

func getBoolVar(vars map[string]interface{}, key string, defaultVal bool) bool {
	val, ok := vars[key]
	if !ok {
		return defaultVal
	}

	boolVal, ok := val.(bool)
	if !ok {
		return defaultVal
	}

	return boolVal
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
