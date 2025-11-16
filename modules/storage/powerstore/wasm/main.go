package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Input represents the JSON input from froyo-runner
type Input struct {
	Vars    map[string]interface{} `json:"vars"`
	Context struct {
		Host     string `json:"host"`
		TaskName string `json:"task_name"`
	} `json:"context"`
}

// Output represents the JSON output to froyo-runner
type Output struct {
	Status  string                 `json:"status"` // ok, changed, failed
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts,omitempty"`
}

// PowerStoreClient handles API interactions
type PowerStoreClient struct {
	BaseURL      string
	Username     string
	Password     string
	HTTPClient   *http.Client
	ValidateCerts bool
}

// NewPowerStoreClient creates a new PowerStore API client
func NewPowerStoreClient(host, username, password string, validateCerts bool, port int) *PowerStoreClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !validateCerts},
	}

	return &PowerStoreClient{
		BaseURL:      fmt.Sprintf("https://%s:%d/api/rest", host, port),
		Username:     username,
		Password:     password,
		HTTPClient:   &http.Client{Transport: tr, Timeout: 30 * time.Second},
		ValidateCerts: validateCerts,
	}
}

// doRequest performs an HTTP request to PowerStore API
func (c *PowerStoreClient) doRequest(method, endpoint string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %v", err)
	}

	return respBody, resp.StatusCode, nil
}

// VOLUME MANAGEMENT OPERATIONS

func (c *PowerStoreClient) createVolume(vars map[string]interface{}) Output {
	volumeName, ok := vars["volume_name"].(string)
	if !ok || volumeName == "" {
		return Output{Status: "failed", Message: "volume_name is required"}
	}

	volumeSize, ok := vars["volume_size"]
	if !ok {
		return Output{Status: "failed", Message: "volume_size is required"}
	}

	// Convert size to bytes (input is in GB)
	var sizeBytes int64
	switch v := volumeSize.(type) {
	case float64:
		sizeBytes = int64(v * 1024 * 1024 * 1024)
	case int:
		sizeBytes = int64(v) * 1024 * 1024 * 1024
	case string:
		size, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("invalid volume_size: %v", err)}
		}
		sizeBytes = size * 1024 * 1024 * 1024
	default:
		return Output{Status: "failed", Message: "volume_size must be a number"}
	}

	reqBody := map[string]interface{}{
		"name": volumeName,
		"size": sizeBytes,
	}

	if description, ok := vars["description"].(string); ok && description != "" {
		reqBody["description"] = description
	}

	respBody, statusCode, err := c.doRequest("POST", "/volume", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' created successfully", volumeName),
		Facts: map[string]interface{}{
			"volume_id":   result["id"],
			"volume_name": volumeName,
			"volume_size": sizeBytes,
		},
	}
}

func (c *PowerStoreClient) deleteVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/volume/"+volumeID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Volume already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' deleted successfully", volumeID),
	}
}

func (c *PowerStoreClient) modifyVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	reqBody := make(map[string]interface{})

	if name, ok := vars["volume_name"].(string); ok && name != "" {
		reqBody["name"] = name
	}
	if desc, ok := vars["description"].(string); ok {
		reqBody["description"] = desc
	}

	if len(reqBody) == 0 {
		return Output{Status: "ok", Message: "No modifications specified"}
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/volume/"+volumeID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' modified successfully", volumeID),
	}
}

func (c *PowerStoreClient) cloneVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	cloneName, ok := vars["clone_name"].(string)
	if !ok || cloneName == "" {
		return Output{Status: "failed", Message: "clone_name is required"}
	}

	reqBody := map[string]interface{}{
		"name": cloneName,
	}

	if description, ok := vars["description"].(string); ok && description != "" {
		reqBody["description"] = description
	}

	respBody, statusCode, err := c.doRequest("POST", "/volume/"+volumeID+"/clone", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume cloned successfully as '%s'", cloneName),
		Facts: map[string]interface{}{
			"clone_id":   result["id"],
			"clone_name": cloneName,
		},
	}
}

func (c *PowerStoreClient) resizeVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	newSize, ok := vars["new_size"]
	if !ok {
		return Output{Status: "failed", Message: "new_size is required"}
	}

	var sizeBytes int64
	switch v := newSize.(type) {
	case float64:
		sizeBytes = int64(v * 1024 * 1024 * 1024)
	case int:
		sizeBytes = int64(v) * 1024 * 1024 * 1024
	case string:
		size, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("invalid new_size: %v", err)}
		}
		sizeBytes = size * 1024 * 1024 * 1024
	default:
		return Output{Status: "failed", Message: "new_size must be a number"}
	}

	reqBody := map[string]interface{}{
		"size": sizeBytes,
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/volume/"+volumeID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' resized to %d GB", volumeID, sizeBytes/(1024*1024*1024)),
	}
}

func (c *PowerStoreClient) mapVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	reqBody := map[string]interface{}{
		"host_id": hostID,
	}

	respBody, statusCode, err := c.doRequest("POST", "/volume/"+volumeID+"/attach", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' mapped to host '%s'", volumeID, hostID),
	}
}

func (c *PowerStoreClient) unmapVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	reqBody := map[string]interface{}{
		"host_id": hostID,
	}

	respBody, statusCode, err := c.doRequest("POST", "/volume/"+volumeID+"/detach", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Volume '%s' unmapped from host '%s'", volumeID, hostID),
	}
}

func (c *PowerStoreClient) showVolume(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	respBody, statusCode, err := c.doRequest("GET", "/volume/"+volumeID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var volume map[string]interface{}
	if err := json.Unmarshal(respBody, &volume); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved volume '%s'", volumeID),
		Facts:   map[string]interface{}{"volume": volume},
	}
}

func (c *PowerStoreClient) showVolumes(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/volume", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var volumes []map[string]interface{}
	if err := json.Unmarshal(respBody, &volumes); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d volumes", len(volumes)),
		Facts:   map[string]interface{}{"volumes": volumes, "count": len(volumes)},
	}
}

func (c *PowerStoreClient) showVolumeDetails(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	respBody, statusCode, err := c.doRequest("GET", "/volume/"+volumeID+"?select=*", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var volumeDetails map[string]interface{}
	if err := json.Unmarshal(respBody, &volumeDetails); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved detailed info for volume '%s'", volumeID),
		Facts:   map[string]interface{}{"volume_details": volumeDetails},
	}
}

// HOST MANAGEMENT OPERATIONS

func (c *PowerStoreClient) createHost(vars map[string]interface{}) Output {
	hostName, ok := vars["host_name"].(string)
	if !ok || hostName == "" {
		return Output{Status: "failed", Message: "host_name is required"}
	}

	reqBody := map[string]interface{}{
		"name": hostName,
	}

	if osType, ok := vars["os_type"].(string); ok && osType != "" {
		reqBody["os_type"] = osType
	}

	if description, ok := vars["description"].(string); ok && description != "" {
		reqBody["description"] = description
	}

	if initiators, ok := vars["initiators"].([]interface{}); ok && len(initiators) > 0 {
		reqBody["initiators"] = initiators
	}

	respBody, statusCode, err := c.doRequest("POST", "/host", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host '%s' created successfully", hostName),
		Facts: map[string]interface{}{
			"host_id":   result["id"],
			"host_name": hostName,
		},
	}
}

func (c *PowerStoreClient) deleteHost(vars map[string]interface{}) Output {
	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/host/"+hostID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Host already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host '%s' deleted successfully", hostID),
	}
}

func (c *PowerStoreClient) modifyHost(vars map[string]interface{}) Output {
	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	reqBody := make(map[string]interface{})

	if name, ok := vars["host_name"].(string); ok && name != "" {
		reqBody["name"] = name
	}
	if desc, ok := vars["description"].(string); ok {
		reqBody["description"] = desc
	}
	if osType, ok := vars["os_type"].(string); ok && osType != "" {
		reqBody["os_type"] = osType
	}

	if len(reqBody) == 0 {
		return Output{Status: "ok", Message: "No modifications specified"}
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/host/"+hostID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host '%s' modified successfully", hostID),
	}
}

func (c *PowerStoreClient) addInitiator(vars map[string]interface{}) Output {
	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	initiator, ok := vars["initiator"].(string)
	if !ok || initiator == "" {
		return Output{Status: "failed", Message: "initiator (WWN or IQN) is required"}
	}

	reqBody := map[string]interface{}{
		"add": []map[string]interface{}{
			{
				"port_name": initiator,
			},
		},
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/host/"+hostID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Initiator '%s' added to host '%s'", initiator, hostID),
	}
}

func (c *PowerStoreClient) removeInitiator(vars map[string]interface{}) Output {
	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	initiator, ok := vars["initiator"].(string)
	if !ok || initiator == "" {
		return Output{Status: "failed", Message: "initiator (WWN or IQN) is required"}
	}

	reqBody := map[string]interface{}{
		"remove": []map[string]interface{}{
			{
				"port_name": initiator,
			},
		},
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/host/"+hostID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Initiator '%s' removed from host '%s'", initiator, hostID),
	}
}

func (c *PowerStoreClient) showHosts(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/host", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(respBody, &hosts); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d hosts", len(hosts)),
		Facts:   map[string]interface{}{"hosts": hosts, "count": len(hosts)},
	}
}

// HOST GROUP MANAGEMENT OPERATIONS

func (c *PowerStoreClient) createHostGroup(vars map[string]interface{}) Output {
	groupName, ok := vars["host_group_name"].(string)
	if !ok || groupName == "" {
		return Output{Status: "failed", Message: "host_group_name is required"}
	}

	reqBody := map[string]interface{}{
		"name": groupName,
	}

	if description, ok := vars["description"].(string); ok && description != "" {
		reqBody["description"] = description
	}

	respBody, statusCode, err := c.doRequest("POST", "/host_group", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host group '%s' created successfully", groupName),
		Facts: map[string]interface{}{
			"host_group_id":   result["id"],
			"host_group_name": groupName,
		},
	}
}

func (c *PowerStoreClient) deleteHostGroup(vars map[string]interface{}) Output {
	groupID, ok := vars["host_group_id"].(string)
	if !ok || groupID == "" {
		return Output{Status: "failed", Message: "host_group_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/host_group/"+groupID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Host group already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host group '%s' deleted successfully", groupID),
	}
}

func (c *PowerStoreClient) addHostToGroup(vars map[string]interface{}) Output {
	groupID, ok := vars["host_group_id"].(string)
	if !ok || groupID == "" {
		return Output{Status: "failed", Message: "host_group_id is required"}
	}

	hostID, ok := vars["host_id"].(string)
	if !ok || hostID == "" {
		return Output{Status: "failed", Message: "host_id is required"}
	}

	reqBody := map[string]interface{}{
		"add": []string{hostID},
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/host_group/"+groupID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Host '%s' added to group '%s'", hostID, groupID),
	}
}

func (c *PowerStoreClient) showHostGroups(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/host_group", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var groups []map[string]interface{}
	if err := json.Unmarshal(respBody, &groups); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d host groups", len(groups)),
		Facts:   map[string]interface{}{"host_groups": groups, "count": len(groups)},
	}
}

// SNAPSHOT MANAGEMENT OPERATIONS

func (c *PowerStoreClient) createSnapshot(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	snapshotName, ok := vars["snapshot_name"].(string)
	if !ok || snapshotName == "" {
		return Output{Status: "failed", Message: "snapshot_name is required"}
	}

	reqBody := map[string]interface{}{
		"name":      snapshotName,
		"volume_id": volumeID,
	}

	if description, ok := vars["description"].(string); ok && description != "" {
		reqBody["description"] = description
	}

	respBody, statusCode, err := c.doRequest("POST", "/volume_snapshot", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Snapshot '%s' created successfully", snapshotName),
		Facts: map[string]interface{}{
			"snapshot_id":   result["id"],
			"snapshot_name": snapshotName,
		},
	}
}

func (c *PowerStoreClient) deleteSnapshot(vars map[string]interface{}) Output {
	snapshotID, ok := vars["snapshot_id"].(string)
	if !ok || snapshotID == "" {
		return Output{Status: "failed", Message: "snapshot_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/volume_snapshot/"+snapshotID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Snapshot already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Snapshot '%s' deleted successfully", snapshotID),
	}
}

func (c *PowerStoreClient) restoreSnapshot(vars map[string]interface{}) Output {
	snapshotID, ok := vars["snapshot_id"].(string)
	if !ok || snapshotID == "" {
		return Output{Status: "failed", Message: "snapshot_id is required"}
	}

	reqBody := map[string]interface{}{}

	respBody, statusCode, err := c.doRequest("POST", "/volume_snapshot/"+snapshotID+"/restore", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Snapshot '%s' restored successfully", snapshotID),
	}
}

func (c *PowerStoreClient) createSnapshotRule(vars map[string]interface{}) Output {
	ruleName, ok := vars["snapshot_rule_name"].(string)
	if !ok || ruleName == "" {
		return Output{Status: "failed", Message: "snapshot_rule_name is required"}
	}

	reqBody := map[string]interface{}{
		"name": ruleName,
	}

	if retention, ok := vars["desired_retention"]; ok {
		reqBody["desired_retention"] = retention
	}

	if interval, ok := vars["interval"]; ok {
		reqBody["interval"] = interval
	}

	respBody, statusCode, err := c.doRequest("POST", "/snapshot_rule", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Snapshot rule '%s' created successfully", ruleName),
		Facts: map[string]interface{}{
			"snapshot_rule_id": result["id"],
		},
	}
}

func (c *PowerStoreClient) deleteSnapshotRule(vars map[string]interface{}) Output {
	ruleID, ok := vars["snapshot_rule_id"].(string)
	if !ok || ruleID == "" {
		return Output{Status: "failed", Message: "snapshot_rule_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/snapshot_rule/"+ruleID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Snapshot rule already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Snapshot rule '%s' deleted successfully", ruleID),
	}
}

func (c *PowerStoreClient) showSnapshots(vars map[string]interface{}) Output {
	endpoint := "/volume_snapshot"

	if volumeID, ok := vars["volume_id"].(string); ok && volumeID != "" {
		endpoint += "?volume_id=" + volumeID
	}

	respBody, statusCode, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var snapshots []map[string]interface{}
	if err := json.Unmarshal(respBody, &snapshots); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d snapshots", len(snapshots)),
		Facts:   map[string]interface{}{"snapshots": snapshots, "count": len(snapshots)},
	}
}

// PERFORMANCE METRICS OPERATIONS

func (c *PowerStoreClient) showVolumeMetrics(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	endpoint := fmt.Sprintf("/metrics/generate?entity=volume&entity_id=%s", volumeID)

	if interval, ok := vars["interval"].(string); ok && interval != "" {
		endpoint += "&interval=" + interval
	}

	respBody, statusCode, err := c.doRequest("POST", endpoint, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(respBody, &metrics); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved metrics for volume '%s'", volumeID),
		Facts:   map[string]interface{}{"metrics": metrics},
	}
}

func (c *PowerStoreClient) showApplianceMetrics(vars map[string]interface{}) Output {
	applianceID, ok := vars["appliance_id"].(string)
	if !ok || applianceID == "" {
		return Output{Status: "failed", Message: "appliance_id is required"}
	}

	endpoint := fmt.Sprintf("/metrics/generate?entity=appliance&entity_id=%s", applianceID)

	if interval, ok := vars["interval"].(string); ok && interval != "" {
		endpoint += "&interval=" + interval
	}

	respBody, statusCode, err := c.doRequest("POST", endpoint, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(respBody, &metrics); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved metrics for appliance '%s'", applianceID),
		Facts:   map[string]interface{}{"metrics": metrics},
	}
}

func (c *PowerStoreClient) showNodeMetrics(vars map[string]interface{}) Output {
	nodeID, ok := vars["node_id"].(string)
	if !ok || nodeID == "" {
		return Output{Status: "failed", Message: "node_id is required"}
	}

	endpoint := fmt.Sprintf("/metrics/generate?entity=node&entity_id=%s", nodeID)

	if interval, ok := vars["interval"].(string); ok && interval != "" {
		endpoint += "&interval=" + interval
	}

	respBody, statusCode, err := c.doRequest("POST", endpoint, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(respBody, &metrics); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved metrics for node '%s'", nodeID),
		Facts:   map[string]interface{}{"metrics": metrics},
	}
}

func (c *PowerStoreClient) showClusterMetrics(vars map[string]interface{}) Output {
	clusterID, ok := vars["cluster_id"].(string)
	if !ok || clusterID == "" {
		return Output{Status: "failed", Message: "cluster_id is required"}
	}

	endpoint := fmt.Sprintf("/metrics/generate?entity=cluster&entity_id=%s", clusterID)

	if interval, ok := vars["interval"].(string); ok && interval != "" {
		endpoint += "&interval=" + interval
	}

	respBody, statusCode, err := c.doRequest("POST", endpoint, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(respBody, &metrics); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved metrics for cluster '%s'", clusterID),
		Facts:   map[string]interface{}{"metrics": metrics},
	}
}

// REPLICATION OPERATIONS

func (c *PowerStoreClient) createReplicationSession(vars map[string]interface{}) Output {
	volumeID, ok := vars["volume_id"].(string)
	if !ok || volumeID == "" {
		return Output{Status: "failed", Message: "volume_id is required"}
	}

	remoteSystemID, ok := vars["remote_system_id"].(string)
	if !ok || remoteSystemID == "" {
		return Output{Status: "failed", Message: "remote_system_id is required"}
	}

	reqBody := map[string]interface{}{
		"local_resource_id":  volumeID,
		"remote_system_id":   remoteSystemID,
	}

	if ruleName, ok := vars["replication_rule"].(string); ok && ruleName != "" {
		reqBody["replication_rule"] = ruleName
	}

	respBody, statusCode, err := c.doRequest("POST", "/replication_session", reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 201 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Replication session created for volume '%s'", volumeID),
		Facts: map[string]interface{}{
			"replication_session_id": result["id"],
		},
	}
}

func (c *PowerStoreClient) deleteReplicationSession(vars map[string]interface{}) Output {
	sessionID, ok := vars["replication_session_id"].(string)
	if !ok || sessionID == "" {
		return Output{Status: "failed", Message: "replication_session_id is required"}
	}

	respBody, statusCode, err := c.doRequest("DELETE", "/replication_session/"+sessionID, nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode == 404 {
		return Output{Status: "ok", Message: "Replication session already deleted or does not exist"}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Replication session '%s' deleted successfully", sessionID),
	}
}

func (c *PowerStoreClient) pauseReplication(vars map[string]interface{}) Output {
	sessionID, ok := vars["replication_session_id"].(string)
	if !ok || sessionID == "" {
		return Output{Status: "failed", Message: "replication_session_id is required"}
	}

	reqBody := map[string]interface{}{
		"pause": true,
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/replication_session/"+sessionID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Replication session '%s' paused", sessionID),
	}
}

func (c *PowerStoreClient) resumeReplication(vars map[string]interface{}) Output {
	sessionID, ok := vars["replication_session_id"].(string)
	if !ok || sessionID == "" {
		return Output{Status: "failed", Message: "replication_session_id is required"}
	}

	reqBody := map[string]interface{}{
		"pause": false,
	}

	respBody, statusCode, err := c.doRequest("PATCH", "/replication_session/"+sessionID, reqBody)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 && statusCode != 204 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Replication session '%s' resumed", sessionID),
	}
}

func (c *PowerStoreClient) showReplicationSessions(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/replication_session", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var sessions []map[string]interface{}
	if err := json.Unmarshal(respBody, &sessions); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d replication sessions", len(sessions)),
		Facts:   map[string]interface{}{"replication_sessions": sessions, "count": len(sessions)},
	}
}

// SYSTEM INFORMATION OPERATIONS

func (c *PowerStoreClient) showClusterInfo(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/cluster", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var clusters []map[string]interface{}
	if err := json.Unmarshal(respBody, &clusters); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: "Retrieved cluster information",
		Facts:   map[string]interface{}{"clusters": clusters},
	}
}

func (c *PowerStoreClient) showAppliances(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/appliance", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var appliances []map[string]interface{}
	if err := json.Unmarshal(respBody, &appliances); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d appliances", len(appliances)),
		Facts:   map[string]interface{}{"appliances": appliances, "count": len(appliances)},
	}
}

func (c *PowerStoreClient) showNodes(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/node", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(respBody, &nodes); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d nodes", len(nodes)),
		Facts:   map[string]interface{}{"nodes": nodes, "count": len(nodes)},
	}
}

func (c *PowerStoreClient) showCapacity(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/capacity", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var capacity map[string]interface{}
	if err := json.Unmarshal(respBody, &capacity); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: "Retrieved capacity information",
		Facts:   map[string]interface{}{"capacity": capacity},
	}
}

func (c *PowerStoreClient) showAlerts(vars map[string]interface{}) Output {
	respBody, statusCode, err := c.doRequest("GET", "/alert", nil)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("API request failed: %v", err)}
	}

	if statusCode != 200 {
		return Output{Status: "failed", Message: fmt.Sprintf("API returned status %d: %s", statusCode, string(respBody))}
	}

	var alerts []map[string]interface{}
	if err := json.Unmarshal(respBody, &alerts); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("failed to parse response: %v", err)}
	}

	return Output{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved %d alerts", len(alerts)),
		Facts:   map[string]interface{}{"alerts": alerts, "count": len(alerts)},
	}
}

// executeCommand routes to the appropriate command handler
func executeCommand(client *PowerStoreClient, command string, vars map[string]interface{}) Output {
	switch strings.ToLower(command) {
	// Volume Management
	case "create_volume":
		return client.createVolume(vars)
	case "delete_volume":
		return client.deleteVolume(vars)
	case "modify_volume":
		return client.modifyVolume(vars)
	case "clone_volume":
		return client.cloneVolume(vars)
	case "resize_volume":
		return client.resizeVolume(vars)
	case "map_volume":
		return client.mapVolume(vars)
	case "unmap_volume":
		return client.unmapVolume(vars)
	case "show_volume":
		return client.showVolume(vars)
	case "show_volumes":
		return client.showVolumes(vars)
	case "show_volume_details":
		return client.showVolumeDetails(vars)

	// Host Management
	case "create_host":
		return client.createHost(vars)
	case "delete_host":
		return client.deleteHost(vars)
	case "modify_host":
		return client.modifyHost(vars)
	case "add_initiator":
		return client.addInitiator(vars)
	case "remove_initiator":
		return client.removeInitiator(vars)
	case "show_hosts":
		return client.showHosts(vars)

	// Host Group Management
	case "create_host_group":
		return client.createHostGroup(vars)
	case "delete_host_group":
		return client.deleteHostGroup(vars)
	case "add_host_to_group":
		return client.addHostToGroup(vars)
	case "show_host_groups":
		return client.showHostGroups(vars)

	// Snapshot Management
	case "create_snapshot":
		return client.createSnapshot(vars)
	case "delete_snapshot":
		return client.deleteSnapshot(vars)
	case "restore_snapshot":
		return client.restoreSnapshot(vars)
	case "create_snapshot_rule":
		return client.createSnapshotRule(vars)
	case "delete_snapshot_rule":
		return client.deleteSnapshotRule(vars)
	case "show_snapshots":
		return client.showSnapshots(vars)

	// Performance Metrics
	case "show_volume_metrics":
		return client.showVolumeMetrics(vars)
	case "show_appliance_metrics":
		return client.showApplianceMetrics(vars)
	case "show_node_metrics":
		return client.showNodeMetrics(vars)
	case "show_cluster_metrics":
		return client.showClusterMetrics(vars)

	// Replication
	case "create_replication_session":
		return client.createReplicationSession(vars)
	case "delete_replication_session":
		return client.deleteReplicationSession(vars)
	case "pause_replication":
		return client.pauseReplication(vars)
	case "resume_replication":
		return client.resumeReplication(vars)
	case "show_replication_sessions":
		return client.showReplicationSessions(vars)

	// System Information
	case "show_cluster_info":
		return client.showClusterInfo(vars)
	case "show_appliances":
		return client.showAppliances(vars)
	case "show_nodes":
		return client.showNodes(vars)
	case "show_capacity":
		return client.showCapacity(vars)
	case "show_alerts":
		return client.showAlerts(vars)

	default:
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", command),
		}
	}
}

func main() {
	// Read input from stdin (passed by froyo-runner)
	var input Input
	decoder := json.NewDecoder(bytes.NewReader([]byte{}))
	if err := decoder.Decode(&input); err != nil {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to parse input: %v", err),
		}
		json.NewEncoder(&bytes.Buffer{}).Encode(output)
		return
	}

	// Extract PowerStore connection parameters
	host, _ := input.Vars["powerstore_host"].(string)
	username, _ := input.Vars["powerstore_user"].(string)
	password, _ := input.Vars["powerstore_password"].(string)
	command, _ := input.Vars["command"].(string)

	validateCerts := true
	if val, ok := input.Vars["validate_certs"].(bool); ok {
		validateCerts = val
	}

	port := 443
	if val, ok := input.Vars["port"]; ok {
		switch v := val.(type) {
		case float64:
			port = int(v)
		case int:
			port = v
		}
	}

	// Validate required parameters
	if host == "" || username == "" || password == "" || command == "" {
		output := Output{
			Status:  "failed",
			Message: "Missing required parameters: powerstore_host, powerstore_user, powerstore_password, and command are required",
		}
		json.NewEncoder(&bytes.Buffer{}).Encode(output)
		return
	}

	// Create PowerStore client
	client := NewPowerStoreClient(host, username, password, validateCerts, port)

	// Execute the command
	output := executeCommand(client, command, input.Vars)

	// Output result as JSON
	json.NewEncoder(&bytes.Buffer{}).Encode(output)
}
