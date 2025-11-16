package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

// RedfishConfig holds the Redfish connection configuration
type RedfishConfig struct {
	BaseURI       string
	Username      string
	Password      string
	Command       string
	Timeout       int
	ValidateCerts bool
	ResourceID    string
	Attributes    map[string]interface{}
	BootSource    string
	BootMode      string
	FirmwareURI   string
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

	// Parse configuration
	config, err := parseConfig(input.Vars)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Build the shell command based on the Redfish operation
	shellCmd, err := buildRedfishCommand(config)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Return shell_exec fact for the generic handler
	result := TaskOutput{
		Status:  "ok",
		Message: fmt.Sprintf("Executing Redfish %s operation", config.Command),
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

func parseConfig(vars map[string]interface{}) (*RedfishConfig, error) {
	config := &RedfishConfig{
		Timeout:       30,
		ValidateCerts: true,
		ResourceID:    "System.Embedded.1",
	}

	// Required: baseuri
	if baseuri, ok := vars["baseuri"].(string); ok {
		config.BaseURI = strings.TrimRight(baseuri, "/")
	} else {
		return nil, fmt.Errorf("missing required variable: baseuri")
	}

	// Required: username
	if username, ok := vars["username"].(string); ok {
		config.Username = username
	} else {
		return nil, fmt.Errorf("missing required variable: username")
	}

	// Required: password
	if password, ok := vars["password"].(string); ok {
		config.Password = password
	} else {
		return nil, fmt.Errorf("missing required variable: password")
	}

	// Required: command
	if command, ok := vars["command"].(string); ok {
		config.Command = command
	} else {
		return nil, fmt.Errorf("missing required variable: command")
	}

	// Optional: timeout
	if timeout, ok := vars["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	}

	// Optional: validate_certs
	if validateCerts, ok := vars["validate_certs"].(bool); ok {
		config.ValidateCerts = validateCerts
	}

	// Optional: resource_id
	if resourceID, ok := vars["resource_id"].(string); ok {
		config.ResourceID = resourceID
	}

	// Optional: attributes (for BIOS configuration)
	if attributes, ok := vars["attributes"].(map[string]interface{}); ok {
		config.Attributes = attributes
	}

	// Optional: boot_source
	if bootSource, ok := vars["boot_source"].(string); ok {
		config.BootSource = bootSource
	}

	// Optional: boot_mode
	if bootMode, ok := vars["boot_mode"].(string); ok {
		config.BootMode = bootMode
	}

	// Optional: firmware_uri
	if firmwareURI, ok := vars["firmware_uri"].(string); ok {
		config.FirmwareURI = firmwareURI
	}

	return config, nil
}

func buildRedfishCommand(config *RedfishConfig) (string, error) {
	// Build curl options
	curlOpts := []string{"-s", "-w", `'\nHTTP_STATUS:%{http_code}'`}

	if !config.ValidateCerts {
		curlOpts = append(curlOpts, "-k")
	}

	curlOptsStr := strings.Join(curlOpts, " ")

	// Escape password for shell
	password := strings.ReplaceAll(config.Password, "'", "'\\''")

	switch config.Command {
	case "PowerOn":
		return buildPowerCommand(config, "On", curlOptsStr, password), nil
	case "PowerOff":
		return buildPowerCommand(config, "ForceOff", curlOptsStr, password), nil
	case "GracefulShutdown":
		return buildPowerCommand(config, "GracefulShutdown", curlOptsStr, password), nil
	case "ForceRestart":
		return buildPowerCommand(config, "ForceRestart", curlOptsStr, password), nil
	case "PowerCycle":
		return buildPowerCommand(config, "PowerCycle", curlOptsStr, password), nil
	case "Nmi":
		return buildPowerCommand(config, "Nmi", curlOptsStr, password), nil
	case "GetPowerState":
		return buildGetPowerStateCommand(config, curlOptsStr, password), nil
	case "GetSystemInfo":
		return buildGetSystemInfoCommand(config, curlOptsStr, password), nil
	case "GetProcessorInfo":
		return buildGetProcessorInfoCommand(config, curlOptsStr, password), nil
	case "GetMemoryInfo":
		return buildGetMemoryInfoCommand(config, curlOptsStr, password), nil
	case "GetStorageInfo":
		return buildGetStorageInfoCommand(config, curlOptsStr, password), nil
	case "GetNetworkInfo":
		return buildGetNetworkInfoCommand(config, curlOptsStr, password), nil
	case "GetBiosAttributes":
		return buildGetBiosAttributesCommand(config, curlOptsStr, password), nil
	case "SetBiosAttributes":
		return buildSetBiosAttributesCommand(config, curlOptsStr, password)
	case "ResetBiosToDefaults":
		return buildResetBiosCommand(config, curlOptsStr, password), nil
	case "GetFirmwareInventory":
		return buildGetFirmwareInventoryCommand(config, curlOptsStr, password), nil
	case "UpdateFirmware":
		return buildUpdateFirmwareCommand(config, curlOptsStr, password)
	case "SetBootSource":
		return buildSetBootSourceCommand(config, curlOptsStr, password)
	case "SetBootMode":
		return buildSetBootModeCommand(config, curlOptsStr, password)
	case "GetBootConfiguration":
		return buildGetBootConfigurationCommand(config, curlOptsStr, password), nil
	default:
		return "", fmt.Errorf("unknown command: %s", config.Command)
	}
}

func buildPowerCommand(config *RedfishConfig, resetType, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Power operation: %s
curl %s -X POST \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  -d '{"ResetType":"%s"}' \
  "%s/redfish/v1/Systems/%s/Actions/ComputerSystem.Reset"
`, resetType, curlOpts, config.Username, password, resetType, config.BaseURI, config.ResourceID)
}

func buildGetPowerStateCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get power state
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s" | jq -r '.PowerState'
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetSystemInfoCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get system information
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetProcessorInfoCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get processor information
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s/Processors"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetMemoryInfoCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get memory information
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s/Memory"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetStorageInfoCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get storage information
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s/Storage"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetNetworkInfoCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get network interface information
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s/EthernetInterfaces"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetBiosAttributesCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get BIOS attributes
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s/Bios"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildSetBiosAttributesCommand(config *RedfishConfig, curlOpts, password string) (string, error) {
	if config.Attributes == nil || len(config.Attributes) == 0 {
		return "", fmt.Errorf("attributes required for SetBiosAttributes command")
	}

	attributesJSON, err := json.Marshal(config.Attributes)
	if err != nil {
		return "", fmt.Errorf("failed to marshal attributes: %v", err)
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

# Set BIOS attributes
curl %s -X PATCH \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  -d '{"Attributes":%s}' \
  "%s/redfish/v1/Systems/%s/Bios/Settings"
`, curlOpts, config.Username, password, string(attributesJSON), config.BaseURI, config.ResourceID), nil
}

func buildResetBiosCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Reset BIOS to defaults
curl %s -X POST \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  "%s/redfish/v1/Systems/%s/Bios/Actions/Bios.ResetBios"
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
}

func buildGetFirmwareInventoryCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get firmware inventory
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/UpdateService/FirmwareInventory"
`, curlOpts, config.Username, password, config.BaseURI)
}

func buildUpdateFirmwareCommand(config *RedfishConfig, curlOpts, password string) (string, error) {
	if config.FirmwareURI == "" {
		return "", fmt.Errorf("firmware_uri required for UpdateFirmware command")
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

# Update firmware
curl %s -X POST \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  -d '{"ImageURI":"%s","TransferProtocol":"HTTP"}' \
  "%s/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"
`, curlOpts, config.Username, password, config.FirmwareURI, config.BaseURI), nil
}

func buildSetBootSourceCommand(config *RedfishConfig, curlOpts, password string) (string, error) {
	if config.BootSource == "" {
		return "", fmt.Errorf("boot_source required for SetBootSource command")
	}

	// Validate boot source
	validSources := map[string]bool{
		"None": true, "Pxe": true, "Hdd": true, "Cd": true,
		"BiosSetup": true, "UefiShell": true, "Usb": true,
		"UefiTarget": true, "SDCard": true, "UefiHttp": true,
	}
	if !validSources[config.BootSource] {
		return "", fmt.Errorf("invalid boot_source: %s", config.BootSource)
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

# Set boot source
curl %s -X PATCH \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  -d '{"Boot":{"BootSourceOverrideTarget":"%s","BootSourceOverrideEnabled":"Once"}}' \
  "%s/redfish/v1/Systems/%s"
`, curlOpts, config.Username, password, config.BootSource, config.BaseURI, config.ResourceID), nil
}

func buildSetBootModeCommand(config *RedfishConfig, curlOpts, password string) (string, error) {
	if config.BootMode == "" {
		return "", fmt.Errorf("boot_mode required for SetBootMode command")
	}

	// Validate boot mode
	if config.BootMode != "UEFI" && config.BootMode != "Legacy" {
		return "", fmt.Errorf("invalid boot_mode: %s (must be UEFI or Legacy)", config.BootMode)
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

# Set boot mode
curl %s -X PATCH \
  -u '%s:%s' \
  -H "Content-Type: application/json" \
  -d '{"Boot":{"BootSourceOverrideMode":"%s"}}' \
  "%s/redfish/v1/Systems/%s"
`, curlOpts, config.Username, password, config.BootMode, config.BaseURI, config.ResourceID), nil
}

func buildGetBootConfigurationCommand(config *RedfishConfig, curlOpts, password string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Get boot configuration
curl %s -X GET \
  -u '%s:%s' \
  -H "Accept: application/json" \
  "%s/redfish/v1/Systems/%s" | jq '.Boot'
`, curlOpts, config.Username, password, config.BaseURI, config.ResourceID)
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
