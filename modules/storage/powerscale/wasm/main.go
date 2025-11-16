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

// PowerScaleConfig holds connection configuration
type PowerScaleConfig struct {
	Host          string
	User          string
	Password      string
	Port          string
	ValidateCerts bool
}

// HTTPRequest represents an HTTP request to be executed by the host
type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
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

	// Extract and validate configuration
	config, err := extractConfig(input.Vars)
	if err != nil {
		outputError(err.Error())
		return
	}

	// Extract command
	command := getString(input.Vars, "command", "")
	if command == "" {
		outputError("Missing required parameter: command")
		return
	}

	// Route to appropriate handler
	var result TaskOutput
	switch command {
	// SMB Share Management
	case "create_smb_share":
		result = createSMBShare(config, input.Vars)
	case "delete_smb_share":
		result = deleteSMBShare(config, input.Vars)
	case "modify_smb_share":
		result = modifySMBShare(config, input.Vars)
	case "show_smb_share":
		result = showSMBShare(config, input.Vars)
	case "show_smb_shares":
		result = showSMBShares(config, input.Vars)
	case "add_smb_permission":
		result = addSMBPermission(config, input.Vars)
	case "remove_smb_permission":
		result = removeSMBPermission(config, input.Vars)
	case "show_smb_sessions":
		result = showSMBSessions(config, input.Vars)

	// NFS Export Management
	case "create_nfs_export":
		result = createNFSExport(config, input.Vars)
	case "delete_nfs_export":
		result = deleteNFSExport(config, input.Vars)
	case "modify_nfs_export":
		result = modifyNFSExport(config, input.Vars)
	case "show_nfs_export":
		result = showNFSExport(config, input.Vars)
	case "show_nfs_exports":
		result = showNFSExports(config, input.Vars)
	case "reload_nfs_exports":
		result = reloadNFSExports(config, input.Vars)

	// Access Zone Management
	case "create_access_zone":
		result = createAccessZone(config, input.Vars)
	case "delete_access_zone":
		result = deleteAccessZone(config, input.Vars)
	case "modify_access_zone":
		result = modifyAccessZone(config, input.Vars)
	case "show_access_zone":
		result = showAccessZone(config, input.Vars)
	case "show_access_zones":
		result = showAccessZones(config, input.Vars)

	// Quota Management
	case "create_quota":
		result = createQuota(config, input.Vars)
	case "delete_quota":
		result = deleteQuota(config, input.Vars)
	case "modify_quota":
		result = modifyQuota(config, input.Vars)
	case "show_quota":
		result = showQuota(config, input.Vars)
	case "show_quotas":
		result = showQuotas(config, input.Vars)
	case "show_quota_reports":
		result = showQuotaReports(config, input.Vars)

	// Snapshot Management
	case "create_snapshot":
		result = createSnapshot(config, input.Vars)
	case "delete_snapshot":
		result = deleteSnapshot(config, input.Vars)
	case "restore_snapshot":
		result = restoreSnapshot(config, input.Vars)
	case "create_snapshot_schedule":
		result = createSnapshotSchedule(config, input.Vars)
	case "delete_snapshot_schedule":
		result = deleteSnapshotSchedule(config, input.Vars)
	case "show_snapshot":
		result = showSnapshot(config, input.Vars)
	case "show_snapshots":
		result = showSnapshots(config, input.Vars)

	// Cluster Management
	case "show_cluster_config":
		result = showClusterConfig(config, input.Vars)
	case "show_cluster_nodes":
		result = showClusterNodes(config, input.Vars)
	case "show_cluster_version":
		result = showClusterVersion(config, input.Vars)
	case "show_cluster_capacity":
		result = showClusterCapacity(config, input.Vars)
	case "show_cluster_performance":
		result = showClusterPerformance(config, input.Vars)
	case "show_cluster_health":
		result = showClusterHealth(config, input.Vars)

	// User/Group Management
	case "create_user":
		result = createUser(config, input.Vars)
	case "delete_user":
		result = deleteUser(config, input.Vars)
	case "create_group":
		result = createGroup(config, input.Vars)
	case "delete_group":
		result = deleteGroup(config, input.Vars)

	// SyncIQ Replication
	case "create_synciq_policy":
		result = createSyncIQPolicy(config, input.Vars)
	case "delete_synciq_policy":
		result = deleteSyncIQPolicy(config, input.Vars)
	case "show_synciq_policies":
		result = showSyncIQPolicies(config, input.Vars)

	default:
		result = TaskOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", command),
			Facts:   make(map[string]interface{}),
		}
	}

	printOutput(result)
}

// extractConfig extracts and validates PowerScale configuration
func extractConfig(vars map[string]interface{}) (*PowerScaleConfig, error) {
	host := getString(vars, "powerscale_host", "")
	if host == "" {
		return nil, fmt.Errorf("missing required parameter: powerscale_host")
	}

	password := getString(vars, "powerscale_password", "")
	if password == "" {
		return nil, fmt.Errorf("missing required parameter: powerscale_password")
	}

	validateCerts := getBool(vars, "validate_certs", false)

	return &PowerScaleConfig{
		Host:          host,
		User:          getString(vars, "powerscale_user", "admin"),
		Password:      password,
		Port:          getString(vars, "powerscale_port", "8080"),
		ValidateCerts: validateCerts,
	}, nil
}

// SMB Share Management Functions

func createSMBShare(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	path := getString(vars, "path", "")
	if path == "" {
		return failedOutput("Missing required parameter: path")
	}

	zone := getString(vars, "access_zone", "System")
	description := getString(vars, "description", "")
	browsable := getBool(vars, "browsable", true)
	ntfsACL := getBool(vars, "ntfs_acl_support", true)

	// Build request body
	body := map[string]interface{}{
		"name":              shareName,
		"path":              path,
		"browsable":         browsable,
		"ntfs_acl_support":  ntfsACL,
		"access_based_enumeration": false,
	}

	if description != "" {
		body["description"] = description
	}

	// Handle permissions if provided
	if perms := getInterface(vars, "permissions"); perms != nil {
		body["permissions"] = perms
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("SMB share '%s' created", shareName), true)
}

func deleteSMBShare(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares/%s", config.Host, config.Port, shareName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("SMB share '%s' deleted", shareName), true)
}

func modifySMBShare(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	zone := getString(vars, "access_zone", "System")

	// Build request body with only provided parameters
	body := make(map[string]interface{})

	if desc := getString(vars, "description", ""); desc != "" {
		body["description"] = desc
	}
	if browsable := getInterface(vars, "browsable"); browsable != nil {
		body["browsable"] = getBool(vars, "browsable", true)
	}
	if perms := getInterface(vars, "permissions"); perms != nil {
		body["permissions"] = perms
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares/%s", config.Host, config.Port, shareName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("SMB share '%s' modified", shareName), true)
}

func showSMBShare(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares/%s", config.Host, config.Port, shareName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Retrieved SMB share '%s'", shareName), false)
}

func showSMBShares(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all SMB shares", false)
}

func addSMBPermission(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	permissions := getInterface(vars, "permissions")
	if permissions == nil {
		return failedOutput("Missing required parameter: permissions")
	}

	zone := getString(vars, "access_zone", "System")

	body := map[string]interface{}{
		"permissions": permissions,
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares/%s", config.Host, config.Port, shareName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Added permissions to SMB share '%s'", shareName), true)
}

func removeSMBPermission(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	shareName := getString(vars, "share_name", "")
	if shareName == "" {
		return failedOutput("Missing required parameter: share_name")
	}

	permissions := getInterface(vars, "permissions")
	if permissions == nil {
		return failedOutput("Missing required parameter: permissions")
	}

	zone := getString(vars, "access_zone", "System")

	body := map[string]interface{}{
		"permissions": permissions,
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/7/protocols/smb/shares/%s", config.Host, config.Port, shareName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Removed permissions from SMB share '%s'", shareName), true)
}

func showSMBSessions(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/protocols/smb/sessions", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved active SMB sessions", false)
}

// NFS Export Management Functions

func createNFSExport(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	paths := getStringSlice(vars, "export_paths")
	if len(paths) == 0 {
		// Try single path
		if path := getString(vars, "export_path", ""); path != "" {
			paths = []string{path}
		} else {
			return failedOutput("Missing required parameter: export_path or export_paths")
		}
	}

	zone := getString(vars, "access_zone", "System")

	body := map[string]interface{}{
		"paths": paths,
	}

	// Add optional parameters
	if clients := getStringSlice(vars, "clients"); len(clients) > 0 {
		body["clients"] = clients
	}
	if readOnly := getBool(vars, "read_only", false); readOnly {
		body["read_only"] = true
	}
	if rootClients := getStringSlice(vars, "root_clients"); len(rootClients) > 0 {
		body["root_clients"] = rootClients
	}
	if secFlavors := getStringSlice(vars, "security_flavors"); len(secFlavors) > 0 {
		body["security_flavors"] = secFlavors
	}
	if mapRoot := getString(vars, "map_root", ""); mapRoot != "" {
		body["map_root"] = map[string]string{"user": mapRoot}
	}
	if mapAll := getString(vars, "map_all", ""); mapAll != "" {
		body["map_all"] = map[string]string{"user": mapAll}
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/exports", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("NFS export created for paths: %v", paths), true)
}

func deleteNFSExport(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	exportID := getString(vars, "export_id", "")
	if exportID == "" {
		return failedOutput("Missing required parameter: export_id")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/exports/%s", config.Host, config.Port, exportID)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("NFS export %s deleted", exportID), true)
}

func modifyNFSExport(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	exportID := getString(vars, "export_id", "")
	if exportID == "" {
		return failedOutput("Missing required parameter: export_id")
	}

	zone := getString(vars, "access_zone", "System")

	body := make(map[string]interface{})

	if clients := getStringSlice(vars, "clients"); len(clients) > 0 {
		body["clients"] = clients
	}
	if readOnly := getInterface(vars, "read_only"); readOnly != nil {
		body["read_only"] = getBool(vars, "read_only", false)
	}
	if rootClients := getStringSlice(vars, "root_clients"); len(rootClients) > 0 {
		body["root_clients"] = rootClients
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/exports/%s", config.Host, config.Port, exportID)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("NFS export %s modified", exportID), true)
}

func showNFSExport(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	exportID := getString(vars, "export_id", "")
	if exportID == "" {
		return failedOutput("Missing required parameter: export_id")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/exports/%s", config.Host, config.Port, exportID)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Retrieved NFS export %s", exportID), false)
}

func showNFSExports(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/exports", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all NFS exports", false)
}

func reloadNFSExports(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/4/protocols/nfs/reload", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "NFS exports reloaded", true)
}

// Access Zone Management Functions

func createAccessZone(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zoneName := getString(vars, "zone_name", "")
	if zoneName == "" {
		return failedOutput("Missing required parameter: zone_name")
	}

	path := getString(vars, "path", "")
	if path == "" {
		return failedOutput("Missing required parameter: path")
	}

	groupnet := getString(vars, "groupnet", "groupnet0")

	body := map[string]interface{}{
		"name":     zoneName,
		"path":     path,
		"groupnet": groupnet,
	}

	if authProviders := getStringSlice(vars, "auth_providers"); len(authProviders) > 0 {
		body["auth_providers"] = authProviders
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/3/zones", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Access zone '%s' created", zoneName), true)
}

func deleteAccessZone(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zoneName := getString(vars, "zone_name", "")
	if zoneName == "" {
		return failedOutput("Missing required parameter: zone_name")
	}

	url := fmt.Sprintf("https://%s:%s/platform/3/zones/%s", config.Host, config.Port, zoneName)

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Access zone '%s' deleted", zoneName), true)
}

func modifyAccessZone(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zoneName := getString(vars, "zone_name", "")
	if zoneName == "" {
		return failedOutput("Missing required parameter: zone_name")
	}

	body := make(map[string]interface{})

	if authProviders := getStringSlice(vars, "auth_providers"); len(authProviders) > 0 {
		body["auth_providers"] = authProviders
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/3/zones/%s", config.Host, config.Port, zoneName)

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Access zone '%s' modified", zoneName), true)
}

func showAccessZone(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	zoneName := getString(vars, "zone_name", "")
	if zoneName == "" {
		return failedOutput("Missing required parameter: zone_name")
	}

	url := fmt.Sprintf("https://%s:%s/platform/3/zones/%s", config.Host, config.Port, zoneName)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Retrieved access zone '%s'", zoneName), false)
}

func showAccessZones(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/zones", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all access zones", false)
}

// Quota Management Functions

func createQuota(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	quotaPath := getString(vars, "quota_path", "")
	if quotaPath == "" {
		return failedOutput("Missing required parameter: quota_path")
	}

	quotaType := getString(vars, "quota_type", "directory")

	body := map[string]interface{}{
		"path":     quotaPath,
		"type":     quotaType,
		"enforced": getBool(vars, "enforce", true),
		"thresholds": map[string]interface{}{},
	}

	thresholds := body["thresholds"].(map[string]interface{})

	if hard := getInt64(vars, "hard_threshold", 0); hard > 0 {
		thresholds["hard"] = hard
	}
	if soft := getInt64(vars, "soft_threshold", 0); soft > 0 {
		thresholds["soft"] = soft
	}
	if advisory := getInt64(vars, "advisory_threshold", 0); advisory > 0 {
		thresholds["advisory"] = advisory
	}

	// Add user/group specific parameters
	if quotaType == "user" {
		if user := getString(vars, "quota_user", ""); user != "" {
			body["persona"] = map[string]string{"name": user}
		}
	} else if quotaType == "group" {
		if group := getString(vars, "quota_group", ""); group != "" {
			body["persona"] = map[string]string{"name": group}
		}
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/quota/quotas", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Quota created for path '%s'", quotaPath), true)
}

func deleteQuota(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	quotaID := getString(vars, "quota_id", "")
	if quotaID == "" {
		return failedOutput("Missing required parameter: quota_id")
	}

	url := fmt.Sprintf("https://%s:%s/platform/1/quota/quotas/%s", config.Host, config.Port, quotaID)

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Quota %s deleted", quotaID), true)
}

func modifyQuota(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	quotaID := getString(vars, "quota_id", "")
	if quotaID == "" {
		return failedOutput("Missing required parameter: quota_id")
	}

	body := map[string]interface{}{
		"thresholds": map[string]interface{}{},
	}

	thresholds := body["thresholds"].(map[string]interface{})

	if hard := getInt64(vars, "hard_threshold", 0); hard > 0 {
		thresholds["hard"] = hard
	}
	if soft := getInt64(vars, "soft_threshold", 0); soft > 0 {
		thresholds["soft"] = soft
	}
	if advisory := getInt64(vars, "advisory_threshold", 0); advisory > 0 {
		thresholds["advisory"] = advisory
	}

	if enforce := getInterface(vars, "enforce"); enforce != nil {
		body["enforced"] = getBool(vars, "enforce", true)
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/quota/quotas/%s", config.Host, config.Port, quotaID)

	req := HTTPRequest{
		Method: "PUT",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Quota %s modified", quotaID), true)
}

func showQuota(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	quotaID := getString(vars, "quota_id", "")
	if quotaID == "" {
		return failedOutput("Missing required parameter: quota_id")
	}

	url := fmt.Sprintf("https://%s:%s/platform/1/quota/quotas/%s", config.Host, config.Port, quotaID)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Retrieved quota %s", quotaID), false)
}

func showQuotas(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/1/quota/quotas", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all quotas", false)
}

func showQuotaReports(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/1/quota/reports", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved quota reports", false)
}

// Snapshot Management Functions

func createSnapshot(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	path := getString(vars, "snapshot_path", "")
	if path == "" {
		return failedOutput("Missing required parameter: snapshot_path")
	}

	name := getString(vars, "snapshot_name", "")

	body := map[string]interface{}{
		"path": path,
	}

	if name != "" {
		body["name"] = name
	}
	if alias := getString(vars, "snapshot_alias", ""); alias != "" {
		body["alias"] = alias
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/snapshots", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Snapshot created for path '%s'", path), true)
}

func deleteSnapshot(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	snapshotID := getString(vars, "snapshot_id", "")
	if snapshotID == "" {
		// Try snapshot name
		if name := getString(vars, "snapshot_name", ""); name != "" {
			snapshotID = name
		} else {
			return failedOutput("Missing required parameter: snapshot_id or snapshot_name")
		}
	}

	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/snapshots/%s", config.Host, config.Port, snapshotID)

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Snapshot %s deleted", snapshotID), true)
}

func restoreSnapshot(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	snapshotID := getString(vars, "snapshot_id", "")
	if snapshotID == "" {
		return failedOutput("Missing required parameter: snapshot_id")
	}

	// Note: PowerScale snapshot restore is typically done via file copy from .snapshot directory
	// This is a placeholder for the restore operation
	return failedOutput("Snapshot restore should be done by copying files from /ifs/.snapshot/<snapshot_name>/ to target location")
}

func createSnapshotSchedule(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	scheduleName := getString(vars, "schedule_name", "")
	if scheduleName == "" {
		return failedOutput("Missing required parameter: schedule_name")
	}

	path := getString(vars, "snapshot_path", "")
	if path == "" {
		return failedOutput("Missing required parameter: snapshot_path")
	}

	schedule := getString(vars, "schedule", "")
	if schedule == "" {
		return failedOutput("Missing required parameter: schedule")
	}

	body := map[string]interface{}{
		"name":     scheduleName,
		"path":     path,
		"schedule": schedule,
	}

	if pattern := getString(vars, "pattern", ""); pattern != "" {
		body["pattern"] = pattern
	}
	if duration := getString(vars, "duration", ""); duration != "" {
		body["duration"] = duration
	}
	if retention := getString(vars, "retention", ""); retention != "" {
		body["retention"] = retention
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/schedules", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Snapshot schedule '%s' created", scheduleName), true)
}

func deleteSnapshotSchedule(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	scheduleName := getString(vars, "schedule_name", "")
	if scheduleName == "" {
		return failedOutput("Missing required parameter: schedule_name")
	}

	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/schedules/%s", config.Host, config.Port, scheduleName)

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Snapshot schedule '%s' deleted", scheduleName), true)
}

func showSnapshot(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	snapshotID := getString(vars, "snapshot_id", "")
	if snapshotID == "" {
		return failedOutput("Missing required parameter: snapshot_id")
	}

	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/snapshots/%s", config.Host, config.Port, snapshotID)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Retrieved snapshot %s", snapshotID), false)
}

func showSnapshots(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/1/snapshot/snapshots", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all snapshots", false)
}

// Cluster Management Functions

func showClusterConfig(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/cluster/config", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster configuration", false)
}

func showClusterNodes(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/cluster/nodes", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster nodes", false)
}

func showClusterVersion(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/cluster/version", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster version", false)
}

func showClusterCapacity(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/1/statistics/current?key=cluster.disk.bytes.used&key=cluster.disk.bytes.total&key=cluster.disk.bytes.avail", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster capacity", false)
}

func showClusterPerformance(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/1/statistics/current?key=cluster.cpu.sys.avg&key=cluster.disk.bytes.in.rate&key=cluster.disk.bytes.out.rate", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster performance metrics", false)
}

func showClusterHealth(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/cluster/diagnostics", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved cluster health status", false)
}

// User/Group Management Functions

func createUser(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	userName := getString(vars, "user_name", "")
	if userName == "" {
		return failedOutput("Missing required parameter: user_name")
	}

	zone := getString(vars, "access_zone", "System")

	body := map[string]interface{}{
		"name": userName,
	}

	if uid := getString(vars, "user_id", ""); uid != "" {
		if uidInt, err := strconv.Atoi(uid); err == nil {
			body["uid"] = uidInt
		}
	}
	if primaryGroup := getString(vars, "primary_group", ""); primaryGroup != "" {
		body["primary_group"] = primaryGroup
	}
	if homeDir := getString(vars, "home_directory", ""); homeDir != "" {
		body["home_directory"] = homeDir
	}
	if shell := getString(vars, "shell", ""); shell != "" {
		body["shell"] = shell
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/auth/users", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("User '%s' created", userName), true)
}

func deleteUser(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	userName := getString(vars, "user_name", "")
	if userName == "" {
		return failedOutput("Missing required parameter: user_name")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/1/auth/users/%s", config.Host, config.Port, userName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("User '%s' deleted", userName), true)
}

func createGroup(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	groupName := getString(vars, "group_name", "")
	if groupName == "" {
		return failedOutput("Missing required parameter: group_name")
	}

	zone := getString(vars, "access_zone", "System")

	body := map[string]interface{}{
		"name": groupName,
	}

	if gid := getString(vars, "group_id", ""); gid != "" {
		if gidInt, err := strconv.Atoi(gid); err == nil {
			body["gid"] = gidInt
		}
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/1/auth/groups", config.Host, config.Port)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("Group '%s' created", groupName), true)
}

func deleteGroup(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	groupName := getString(vars, "group_name", "")
	if groupName == "" {
		return failedOutput("Missing required parameter: group_name")
	}

	zone := getString(vars, "access_zone", "System")

	url := fmt.Sprintf("https://%s:%s/platform/1/auth/groups/%s", config.Host, config.Port, groupName)
	if zone != "System" {
		url += fmt.Sprintf("?zone=%s", zone)
	}

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("Group '%s' deleted", groupName), true)
}

// SyncIQ Replication Functions

func createSyncIQPolicy(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	policyName := getString(vars, "policy_name", "")
	if policyName == "" {
		return failedOutput("Missing required parameter: policy_name")
	}

	sourcePath := getString(vars, "source_root_path", "")
	if sourcePath == "" {
		return failedOutput("Missing required parameter: source_root_path")
	}

	targetHost := getString(vars, "target_host", "")
	if targetHost == "" {
		return failedOutput("Missing required parameter: target_host")
	}

	targetPath := getString(vars, "target_path", "")
	if targetPath == "" {
		return failedOutput("Missing required parameter: target_path")
	}

	body := map[string]interface{}{
		"name":             policyName,
		"source_root_path": sourcePath,
		"target_host":      targetHost,
		"target_path":      targetPath,
		"action":           getString(vars, "action", "sync"),
		"enabled":          getBool(vars, "enabled", true),
	}

	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://%s:%s/platform/3/sync/policies", config.Host, config.Port)

	req := HTTPRequest{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": basicAuth(config.User, config.Password),
		},
		Body: string(bodyJSON),
	}

	return httpRequestOutput(req, fmt.Sprintf("SyncIQ policy '%s' created", policyName), true)
}

func deleteSyncIQPolicy(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	policyName := getString(vars, "policy_name", "")
	if policyName == "" {
		return failedOutput("Missing required parameter: policy_name")
	}

	url := fmt.Sprintf("https://%s:%s/platform/3/sync/policies/%s", config.Host, config.Port, policyName)

	req := HTTPRequest{
		Method: "DELETE",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, fmt.Sprintf("SyncIQ policy '%s' deleted", policyName), true)
}

func showSyncIQPolicies(config *PowerScaleConfig, vars map[string]interface{}) TaskOutput {
	url := fmt.Sprintf("https://%s:%s/platform/3/sync/policies", config.Host, config.Port)

	req := HTTPRequest{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Authorization": basicAuth(config.User, config.Password),
		},
	}

	return httpRequestOutput(req, "Retrieved all SyncIQ policies", false)
}

// Utility Functions

func getString(vars map[string]interface{}, key string, defaultValue string) string {
	if val, ok := vars[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getBool(vars map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := vars[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
		if strVal, ok := val.(string); ok {
			return strings.ToLower(strVal) == "true"
		}
	}
	return defaultValue
}

func getInt64(vars map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := vars[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case string:
			if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

func getInterface(vars map[string]interface{}, key string) interface{} {
	if val, ok := vars[key]; ok {
		return val
	}
	return nil
}

func getStringSlice(vars map[string]interface{}, key string) []string {
	if val, ok := vars[key]; ok {
		if sliceVal, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(sliceVal))
			for _, item := range sliceVal {
				if strItem, ok := item.(string); ok {
					result = append(result, strItem)
				}
			}
			return result
		}
		if strVal, ok := val.(string); ok && strVal != "" {
			// Handle comma-separated string
			return strings.Split(strVal, ",")
		}
	}
	return []string{}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64Encode(auth)
}

func base64Encode(data string) string {
	// Simple base64 encoding (simplified for WASM)
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	bytes := []byte(data)
	result := ""

	for i := 0; i < len(bytes); i += 3 {
		b1, b2, b3 := bytes[i], byte(0), byte(0)
		if i+1 < len(bytes) {
			b2 = bytes[i+1]
		}
		if i+2 < len(bytes) {
			b3 = bytes[i+2]
		}

		result += string(base64Chars[(b1>>2)&0x3F])
		result += string(base64Chars[((b1&0x03)<<4)|((b2>>4)&0x0F)])

		if i+1 < len(bytes) {
			result += string(base64Chars[((b2&0x0F)<<2)|((b3>>6)&0x03)])
		} else {
			result += "="
		}

		if i+2 < len(bytes) {
			result += string(base64Chars[b3&0x3F])
		} else {
			result += "="
		}
	}

	return result
}

func httpRequestOutput(req HTTPRequest, message string, changed bool) TaskOutput {
	status := "ok"
	if changed {
		status = "changed"
	}

	return TaskOutput{
		Status:  status,
		Message: message,
		Facts: map[string]interface{}{
			"http_request": req,
		},
	}
}

func failedOutput(message string) TaskOutput {
	return TaskOutput{
		Status:  "failed",
		Message: message,
		Facts:   make(map[string]interface{}),
	}
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
