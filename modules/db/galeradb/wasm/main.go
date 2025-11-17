package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Input represents the module input parameters
type Input struct {
	GaleraHost      string `json:"galera_host"`
	GaleraPort      int    `json:"galera_port"`
	GaleraUser      string `json:"galera_user"`
	GaleraPassword  string `json:"galera_password"`
	Command         string `json:"command"`
	ClusterAddress  string `json:"cluster_address"`
	ClusterName     string `json:"cluster_name"`
	NodeAddress     string `json:"node_address"`
	NodeName        string `json:"node_name"`
	SafeToBootstrap int    `json:"safe_to_bootstrap"`
	GrastateFile    string `json:"grastate_file"`
	MysqlService    string `json:"mysql_service"`
	CommandTimeout  int    `json:"command_timeout"`
	QueryLimit      int    `json:"query_limit"`
	ShowAll         bool   `json:"show_all"`
}

// Output represents the module output
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// GaleraStatus represents Galera cluster status
type GaleraStatus struct {
	ClusterSize           string `json:"cluster_size"`
	ClusterStatus         string `json:"cluster_status"`
	LocalStateComment     string `json:"local_state_comment"`
	Ready                 string `json:"ready"`
	Connected             string `json:"connected"`
	LocalState            string `json:"local_state"`
	ClusterUUID           string `json:"cluster_uuid"`
	ProviderVersion       string `json:"provider_version"`
}

func main() {
	if len(os.Args) < 2 {
		output := Output{
			Status:  "failed",
			Message: "No input provided",
			Facts:   map[string]interface{}{},
		}
		json.NewEncoder(os.Stdout).Encode(output)
		os.Exit(1)
	}

	var input Input
	if err := json.Unmarshal([]byte(os.Args[1]), &input); err != nil {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input: %v", err),
			Facts:   map[string]interface{}{},
		}
		json.NewEncoder(os.Stdout).Encode(output)
		os.Exit(1)
	}

	output := executeCommand(&input)
	json.NewEncoder(os.Stdout).Encode(output)

	if output.Status == "failed" {
		os.Exit(1)
	}
}

func executeCommand(input *Input) Output {
	switch input.Command {
	// Cluster Management
	case "bootstrap_cluster":
		return bootstrapCluster(input)
	case "show_cluster_status":
		return showClusterStatus(input)
	case "show_cluster_size":
		return showClusterSize(input)
	case "check_node_status":
		return checkNodeStatus(input)
	case "set_cluster_address":
		return setClusterAddress(input)
	case "show_wsrep_status":
		return showWsrepStatus(input)

	// Node Management
	case "join_cluster":
		return joinCluster(input)
	case "leave_cluster":
		return leaveCluster(input)
	case "restart_node":
		return restartNode(input)
	case "check_node_synced":
		return checkNodeSynced(input)
	case "show_node_state":
		return showNodeState(input)

	// Replication Monitoring
	case "show_replication_status":
		return showReplicationStatus(input)
	case "show_flow_control":
		return showFlowControl(input)
	case "show_apply_lag":
		return showApplyLag(input)
	case "show_conflicts":
		return showConflicts(input)
	case "show_queue_size":
		return showQueueSize(input)

	// Recovery Operations
	case "recover_cluster":
		return recoverCluster(input)
	case "force_bootstrap":
		return forceBootstrap(input)
	case "set_safe_to_bootstrap":
		return setSafeToBootstrap(input)
	case "check_grastate":
		return checkGrastate(input)

	default:
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", input.Command),
			Facts:   map[string]interface{}{},
		}
	}
}

// Helper function to execute mysql command
func executeMysqlCommand(input *Input, query string) (string, error) {
	args := []string{
		"-h", input.GaleraHost,
		"-P", strconv.Itoa(input.GaleraPort),
		"-u", input.GaleraUser,
	}

	if input.GaleraPassword != "" {
		args = append(args, fmt.Sprintf("-p%s", input.GaleraPassword))
	}

	args = append(args, "-e", query, "-N", "-B")

	cmd := exec.Command("mysql", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mysql command failed: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// Helper function to get wsrep status variable
func getWsrepStatus(input *Input, variable string) (string, error) {
	query := fmt.Sprintf("SHOW STATUS LIKE '%s'", variable)
	output, err := executeMysqlCommand(input, query)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(output)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return "", nil
}

// Cluster Management Commands

func bootstrapCluster(input *Input) Output {
	// This requires stopping mysql first
	stopCmd := exec.Command("systemctl", "stop", input.MysqlService)
	if err := stopCmd.Run(); err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to stop mysql service: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	// Set safe_to_bootstrap
	setSafeOutput := setSafeToBootstrap(&Input{
		SafeToBootstrap: 1,
		GrastateFile:    input.GrastateFile,
	})
	if setSafeOutput.Status == "failed" {
		return setSafeOutput
	}

	// Start with new cluster
	startCmd := exec.Command("galera_new_cluster")
	if err := startCmd.Run(); err != nil {
		// Try alternative method
		startCmd = exec.Command("mysqld_safe", "--wsrep-new-cluster", "&")
		if err := startCmd.Start(); err != nil {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to bootstrap cluster: %v", err),
				Facts:   map[string]interface{}{},
			}
		}
	}

	return Output{
		Status:  "changed",
		Message: "Galera cluster bootstrapped successfully",
		Facts: map[string]interface{}{
			"bootstrapped": true,
		},
	}
}

func showClusterStatus(input *Input) Output {
	status := GaleraStatus{}

	variables := map[string]*string{
		"wsrep_cluster_size":         &status.ClusterSize,
		"wsrep_cluster_status":       &status.ClusterStatus,
		"wsrep_local_state_comment":  &status.LocalStateComment,
		"wsrep_ready":                &status.Ready,
		"wsrep_connected":            &status.Connected,
		"wsrep_local_state":          &status.LocalState,
		"wsrep_cluster_uuid":         &status.ClusterUUID,
		"wsrep_provider_version":     &status.ProviderVersion,
	}

	for varName, varPtr := range variables {
		value, err := getWsrepStatus(input, varName)
		if err != nil {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to get %s: %v", varName, err),
				Facts:   map[string]interface{}{},
			}
		}
		*varPtr = value
	}

	facts := map[string]interface{}{
		"cluster_size":          status.ClusterSize,
		"cluster_status":        status.ClusterStatus,
		"local_state_comment":   status.LocalStateComment,
		"ready":                 status.Ready,
		"connected":             status.Connected,
		"local_state":           status.LocalState,
		"cluster_uuid":          status.ClusterUUID,
		"provider_version":      status.ProviderVersion,
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Cluster status: %s, Size: %s, Node state: %s", status.ClusterStatus, status.ClusterSize, status.LocalStateComment),
		Facts:   facts,
	}
}

func showClusterSize(input *Input) Output {
	size, err := getWsrepStatus(input, "wsrep_cluster_size")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get cluster size: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Cluster size: %s nodes", size),
		Facts: map[string]interface{}{
			"cluster_size": size,
		},
	}
}

func checkNodeStatus(input *Input) Output {
	ready, err := getWsrepStatus(input, "wsrep_ready")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to check node status: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	connected, err := getWsrepStatus(input, "wsrep_connected")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to check connection status: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	state, err := getWsrepStatus(input, "wsrep_local_state_comment")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get node state: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	healthy := ready == "ON" && connected == "ON" && state == "Synced"

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Node status - Ready: %s, Connected: %s, State: %s, Healthy: %v", ready, connected, state, healthy),
		Facts: map[string]interface{}{
			"ready":     ready,
			"connected": connected,
			"state":     state,
			"healthy":   healthy,
		},
	}
}

func setClusterAddress(input *Input) Output {
	if input.ClusterAddress == "" {
		return Output{
			Status:  "failed",
			Message: "cluster_address parameter is required",
			Facts:   map[string]interface{}{},
		}
	}

	query := fmt.Sprintf("SET GLOBAL wsrep_cluster_address='%s'", input.ClusterAddress)
	_, err := executeMysqlCommand(input, query)
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to set cluster address: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Cluster address set to: %s", input.ClusterAddress),
		Facts: map[string]interface{}{
			"cluster_address": input.ClusterAddress,
		},
	}
}

func showWsrepStatus(input *Input) Output {
	query := "SHOW STATUS LIKE 'wsrep_%'"
	output, err := executeMysqlCommand(input, query)
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get wsrep status: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	lines := strings.Split(output, "\n")
	facts := make(map[string]interface{})

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			facts[parts[0]] = parts[1]
		}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d wsrep status variables", len(facts)),
		Facts:   facts,
	}
}

// Node Management Commands

func joinCluster(input *Input) Output {
	if input.ClusterAddress == "" {
		return Output{
			Status:  "failed",
			Message: "cluster_address parameter is required",
			Facts:   map[string]interface{}{},
		}
	}

	// Set cluster address
	setOutput := setClusterAddress(input)
	if setOutput.Status == "failed" {
		return setOutput
	}

	// Restart mysql service
	restartCmd := exec.Command("systemctl", "restart", input.MysqlService)
	if err := restartCmd.Run(); err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to restart mysql service: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Node joining cluster at %s", input.ClusterAddress),
		Facts: map[string]interface{}{
			"cluster_address": input.ClusterAddress,
			"action":          "join",
		},
	}
}

func leaveCluster(input *Input) Output {
	stopCmd := exec.Command("systemctl", "stop", input.MysqlService)
	if err := stopCmd.Run(); err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to stop mysql service: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	return Output{
		Status:  "changed",
		Message: "Node left cluster (mysql service stopped)",
		Facts: map[string]interface{}{
			"action": "leave",
		},
	}
}

func restartNode(input *Input) Output {
	restartCmd := exec.Command("systemctl", "restart", input.MysqlService)
	if err := restartCmd.Run(); err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to restart mysql service: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	return Output{
		Status:  "changed",
		Message: "Node restarted successfully",
		Facts: map[string]interface{}{
			"action": "restart",
		},
	}
}

func checkNodeSynced(input *Input) Output {
	state, err := getWsrepStatus(input, "wsrep_local_state_comment")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to check sync status: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	synced := state == "Synced"

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Node sync status: %s (synced: %v)", state, synced),
		Facts: map[string]interface{}{
			"state":  state,
			"synced": synced,
		},
	}
}

func showNodeState(input *Input) Output {
	state, err := getWsrepStatus(input, "wsrep_local_state")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get node state: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	stateComment, err := getWsrepStatus(input, "wsrep_local_state_comment")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get node state comment: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	stateUUID, _ := getWsrepStatus(input, "wsrep_local_state_uuid")

	stateMap := map[string]string{
		"1": "Joining",
		"2": "Donor/Desynced",
		"3": "Joined",
		"4": "Synced",
	}

	stateDesc := stateMap[state]
	if stateDesc == "" {
		stateDesc = "Unknown"
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Node state: %s (%s)", stateComment, stateDesc),
		Facts: map[string]interface{}{
			"state":         state,
			"state_comment": stateComment,
			"state_desc":    stateDesc,
			"state_uuid":    stateUUID,
		},
	}
}

// Replication Monitoring Commands

func showReplicationStatus(input *Input) Output {
	variables := []string{
		"wsrep_last_committed",
		"wsrep_replicated",
		"wsrep_replicated_bytes",
		"wsrep_received",
		"wsrep_received_bytes",
		"wsrep_local_commits",
		"wsrep_local_cert_failures",
		"wsrep_local_bf_aborts",
	}

	facts := make(map[string]interface{})
	for _, varName := range variables {
		value, err := getWsrepStatus(input, varName)
		if err != nil {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to get %s: %v", varName, err),
				Facts:   map[string]interface{}{},
			}
		}
		facts[varName] = value
	}

	return Output{
		Status:  "ok",
		Message: "Replication status retrieved successfully",
		Facts:   facts,
	}
}

func showFlowControl(input *Input) Output {
	variables := []string{
		"wsrep_flow_control_paused",
		"wsrep_flow_control_paused_ns",
		"wsrep_flow_control_sent",
		"wsrep_flow_control_recv",
	}

	facts := make(map[string]interface{})
	for _, varName := range variables {
		value, err := getWsrepStatus(input, varName)
		if err != nil {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to get %s: %v", varName, err),
				Facts:   map[string]interface{}{},
			}
		}
		facts[varName] = value
	}

	return Output{
		Status:  "ok",
		Message: "Flow control status retrieved successfully",
		Facts:   facts,
	}
}

func showApplyLag(input *Input) Output {
	recvQueue, err := getWsrepStatus(input, "wsrep_local_recv_queue")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get receive queue: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	recvQueueAvg, _ := getWsrepStatus(input, "wsrep_local_recv_queue_avg")
	recvQueueMax, _ := getWsrepStatus(input, "wsrep_local_recv_queue_max")
	recvQueueMin, _ := getWsrepStatus(input, "wsrep_local_recv_queue_min")

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Apply lag - Current queue: %s", recvQueue),
		Facts: map[string]interface{}{
			"recv_queue":     recvQueue,
			"recv_queue_avg": recvQueueAvg,
			"recv_queue_max": recvQueueMax,
			"recv_queue_min": recvQueueMin,
		},
	}
}

func showConflicts(input *Input) Output {
	variables := []string{
		"wsrep_local_cert_failures",
		"wsrep_local_bf_aborts",
		"wsrep_cert_deps_distance",
	}

	facts := make(map[string]interface{})
	for _, varName := range variables {
		value, err := getWsrepStatus(input, varName)
		if err != nil {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Failed to get %s: %v", varName, err),
				Facts:   map[string]interface{}{},
			}
		}
		facts[varName] = value
	}

	return Output{
		Status:  "ok",
		Message: "Conflict statistics retrieved successfully",
		Facts:   facts,
	}
}

func showQueueSize(input *Input) Output {
	sendQueue, err := getWsrepStatus(input, "wsrep_local_send_queue")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get send queue: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	recvQueue, err := getWsrepStatus(input, "wsrep_local_recv_queue")
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to get receive queue: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	sendQueueAvg, _ := getWsrepStatus(input, "wsrep_local_send_queue_avg")
	recvQueueAvg, _ := getWsrepStatus(input, "wsrep_local_recv_queue_avg")

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Queue sizes - Send: %s, Recv: %s", sendQueue, recvQueue),
		Facts: map[string]interface{}{
			"send_queue":     sendQueue,
			"recv_queue":     recvQueue,
			"send_queue_avg": sendQueueAvg,
			"recv_queue_avg": recvQueueAvg,
		},
	}
}

// Recovery Operations

func recoverCluster(input *Input) Output {
	// Check grastate.dat first
	grastateOutput := checkGrastate(input)
	if grastateOutput.Status == "failed" {
		return grastateOutput
	}

	// Bootstrap the cluster
	return bootstrapCluster(input)
}

func forceBootstrap(input *Input) Output {
	// Force safe_to_bootstrap to 1
	input.SafeToBootstrap = 1
	setSafeOutput := setSafeToBootstrap(input)
	if setSafeOutput.Status == "failed" {
		return setSafeOutput
	}

	// Bootstrap
	return bootstrapCluster(input)
}

func setSafeToBootstrap(input *Input) Output {
	// Read grastate file
	data, err := os.ReadFile(input.GrastateFile)
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to read grastate file: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	content := string(data)
	newValue := fmt.Sprintf("safe_to_bootstrap: %d", input.SafeToBootstrap)

	// Replace safe_to_bootstrap value
	lines := strings.Split(content, "\n")
	modified := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "safe_to_bootstrap:") {
			lines[i] = newValue
			modified = true
			break
		}
	}

	if !modified {
		return Output{
			Status:  "failed",
			Message: "safe_to_bootstrap entry not found in grastate file",
			Facts:   map[string]interface{}{},
		}
	}

	newContent := strings.Join(lines, "\n")
	err = os.WriteFile(input.GrastateFile, []byte(newContent), 0644)
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to write grastate file: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	status := "ok"
	if modified {
		status = "changed"
	}

	return Output{
		Status:  status,
		Message: fmt.Sprintf("safe_to_bootstrap set to %d", input.SafeToBootstrap),
		Facts: map[string]interface{}{
			"safe_to_bootstrap": input.SafeToBootstrap,
			"grastate_file":     input.GrastateFile,
		},
	}
}

func checkGrastate(input *Input) Output {
	data, err := os.ReadFile(input.GrastateFile)
	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to read grastate file: %v", err),
			Facts:   map[string]interface{}{},
		}
	}

	content := string(data)
	facts := make(map[string]interface{})
	facts["content"] = content

	// Parse grastate.dat
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			facts[key] = value
		}
	}

	return Output{
		Status:  "ok",
		Message: "Grastate file read successfully",
		Facts:   facts,
	}
}
