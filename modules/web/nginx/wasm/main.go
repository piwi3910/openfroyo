package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Input represents the JSON input from the orchestrator
type Input struct {
	Vars    map[string]interface{} `json:"vars"`
	Context struct {
		Host     string `json:"host"`
		TaskName string `json:"task_name"`
	} `json:"context"`

	// Command to execute
	Command string `json:"command"`

	// Service configuration
	NginxConfigDir      string `json:"nginx_config_dir"`
	NginxSitesAvailable string `json:"nginx_sites_available"`
	NginxSitesEnabled   string `json:"nginx_sites_enabled"`
	NginxConfD          string `json:"nginx_conf_d"`

	// Server block parameters
	ServerName     string `json:"server_name"`
	ServerFile     string `json:"server_file"`
	ListenPort     string `json:"listen_port"`
	ListenSSLPort  string `json:"listen_ssl_port"`
	Root           string `json:"root"`
	Index          string `json:"index"`
	AccessLog      string `json:"access_log"`
	ErrorLog       string `json:"error_log"`

	// SSL/TLS parameters
	SSLCertificate            string `json:"ssl_certificate"`
	SSLCertificateKey         string `json:"ssl_certificate_key"`
	SSLProtocols              string `json:"ssl_protocols"`
	SSLCiphers                string `json:"ssl_ciphers"`
	SSLPreferServerCiphers    string `json:"ssl_prefer_server_ciphers"`
	SSLSessionCache           string `json:"ssl_session_cache"`
	SSLSessionTimeout         string `json:"ssl_session_timeout"`
	HTTP2                     string `json:"http2"`
	HSTSEnabled               string `json:"hsts_enabled"`
	HSTSMaxAge                string `json:"hsts_max_age"`

	// Upstream parameters
	UpstreamName         string   `json:"upstream_name"`
	UpstreamServers      []string `json:"upstream_servers"`
	UpstreamServer       string   `json:"upstream_server"`
	LoadBalanceMethod    string   `json:"load_balance_method"`
	HashKey              string   `json:"hash_key"`
	UpstreamKeepalive    string   `json:"upstream_keepalive"`

	// Location parameters
	LocationPath     string `json:"location_path"`
	LocationModifier string `json:"location_modifier"`
	ProxyPass        string `json:"proxy_pass"`
	FastCGIPass      string `json:"fastcgi_pass"`
	TryFiles         string `json:"try_files"`
	Alias            string `json:"alias"`

	// Rewrite parameters
	RewriteRegex       string `json:"rewrite_regex"`
	RewriteReplacement string `json:"rewrite_replacement"`
	RewriteFlag        string `json:"rewrite_flag"`

	// Header parameters
	HeaderName   string `json:"header_name"`
	HeaderValue  string `json:"header_value"`
	HeaderAction string `json:"header_action"`

	// Performance parameters
	WorkerProcesses   string `json:"worker_processes"`
	WorkerConnections string `json:"worker_connections"`

	// Additional options
	AdditionalConfig string `json:"additional_config"`
	Force            string `json:"force"`
}

// Output represents the JSON output to return
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// Command handler function type
type CommandHandler func(*Input) *Output

// Command registry
var commands = map[string]CommandHandler{
	// Server block management
	"create_server_block":  createServerBlock,
	"delete_server_block":  deleteServerBlock,
	"enable_server_block":  enableServerBlock,
	"disable_server_block": disableServerBlock,
	"list_server_blocks":   listServerBlocks,
	"test_server_config":   testServerConfig,

	// Upstream and load balancing
	"create_upstream":          createUpstream,
	"delete_upstream":          deleteUpstream,
	"add_upstream_server":      addUpstreamServer,
	"remove_upstream_server":   removeUpstreamServer,
	"set_load_balance_method":  setLoadBalanceMethod,
	"show_upstreams":           showUpstreams,

	// SSL/TLS configuration
	"create_ssl_server":       createSSLServer,
	"install_ssl_certificate": installSSLCertificate,
	"set_ssl_protocols":       setSSLProtocols,
	"set_ssl_ciphers":         setSSLCiphers,
	"enable_http2":            enableHTTP2,
	"test_ssl_config":         testSSLConfig,

	// Location blocks
	"add_location_block":  addLocationBlock,
	"add_proxy_pass":      addProxyPass,
	"add_fastcgi_pass":    addFastCGIPass,
	"add_static_location": addStaticLocation,
	"add_rewrite_rule":    addRewriteRule,

	// Configuration management
	"set_server_name": setServerName,
	"set_root":        setRoot,
	"set_index":       setIndex,
	"add_header":      addHeader,
	"test_config":     testConfig,
	"reload_config":   reloadConfig,

	// Service management
	"start_service":         startService,
	"stop_service":          stopService,
	"restart_service":       restartService,
	"check_service_status":  checkServiceStatus,

	// Performance tuning
	"set_worker_processes":   setWorkerProcesses,
	"set_worker_connections": setWorkerConnections,
}

func main() {
	// Read input from stdin
	var input Input
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Failed to decode input: %v", err),
			Facts:   make(map[string]interface{}),
		}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	// Set defaults if not provided
	if input.NginxConfigDir == "" {
		input.NginxConfigDir = "/etc/nginx"
	}
	if input.NginxSitesAvailable == "" {
		input.NginxSitesAvailable = filepath.Join(input.NginxConfigDir, "sites-available")
	}
	if input.NginxSitesEnabled == "" {
		input.NginxSitesEnabled = filepath.Join(input.NginxConfigDir, "sites-enabled")
	}
	if input.NginxConfD == "" {
		input.NginxConfD = filepath.Join(input.NginxConfigDir, "conf.d")
	}

	// Find and execute the command handler
	handler, exists := commands[input.Command]
	if !exists {
		output := Output{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown command: %s", input.Command),
			Facts:   make(map[string]interface{}),
		}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	// Execute the command
	output := handler(&input)

	// Return the output
	json.NewEncoder(os.Stdout).Encode(output)
}

// ============================================================================
// Server Block Management
// ============================================================================

func createServerBlock(input *Input) *Output {
	if input.ServerName == "" {
		return &Output{Status: "failed", Message: "server_name is required", Facts: make(map[string]interface{})}
	}
	if input.ServerFile == "" {
		input.ServerFile = input.ServerName + ".conf"
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)

	// Check if file already exists
	if fileExists(configPath) && input.Force != "true" {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server block %s already exists", input.ServerFile),
			Facts:   map[string]interface{}{"config_path": configPath, "exists": true},
		}
	}

	// Generate server block configuration
	config := generateServerBlock(input)

	// Write configuration file
	if err := writeFile(configPath, config, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	// Test configuration
	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Created server block: %s", input.ServerFile),
		Facts:   map[string]interface{}{"config_path": configPath, "server_name": input.ServerName},
	}
}

func deleteServerBlock(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	enabledPath := filepath.Join(input.NginxSitesEnabled, input.ServerFile)

	// Check if file exists
	if !fileExists(configPath) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server block %s does not exist", input.ServerFile),
			Facts:   make(map[string]interface{}),
		}
	}

	// Disable first if enabled
	if fileExists(enabledPath) || isSymlink(enabledPath) {
		if err := removeFile(enabledPath); err != nil {
			return &Output{Status: "failed", Message: fmt.Sprintf("Failed to remove symlink: %v", err), Facts: make(map[string]interface{})}
		}
	}

	// Delete the configuration file
	if err := removeFile(configPath); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to delete config: %v", err), Facts: make(map[string]interface{})}
	}

	// Test and reload
	if testNginxConfig() {
		reloadNginx()
	}

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Deleted server block: %s", input.ServerFile),
		Facts:   make(map[string]interface{}),
	}
}

func enableServerBlock(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	availablePath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	enabledPath := filepath.Join(input.NginxSitesEnabled, input.ServerFile)

	// Check if available file exists
	if !fileExists(availablePath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found in sites-available", input.ServerFile), Facts: make(map[string]interface{})}
	}

	// Check if already enabled
	if fileExists(enabledPath) || isSymlink(enabledPath) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server block %s already enabled", input.ServerFile),
			Facts:   make(map[string]interface{}),
		}
	}

	// Create symlink
	if err := createSymlink(availablePath, enabledPath); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to create symlink: %v", err), Facts: make(map[string]interface{})}
	}

	// Test and reload
	if !testNginxConfig() {
		removeFile(enabledPath)
		return &Output{Status: "failed", Message: "Nginx configuration test failed, symlink removed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Enabled server block: %s", input.ServerFile),
		Facts:   map[string]interface{}{"enabled_path": enabledPath},
	}
}

func disableServerBlock(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	enabledPath := filepath.Join(input.NginxSitesEnabled, input.ServerFile)

	// Check if symlink exists
	if !fileExists(enabledPath) && !isSymlink(enabledPath) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server block %s already disabled", input.ServerFile),
			Facts:   make(map[string]interface{}),
		}
	}

	// Remove symlink
	if err := removeFile(enabledPath); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to remove symlink: %v", err), Facts: make(map[string]interface{})}
	}

	// Reload nginx
	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Disabled server block: %s", input.ServerFile),
		Facts:   make(map[string]interface{}),
	}
}

func listServerBlocks(input *Input) *Output {
	available := listFiles(input.NginxSitesAvailable)
	enabled := listFiles(input.NginxSitesEnabled)

	return &Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d available, %d enabled server blocks", len(available), len(enabled)),
		Facts: map[string]interface{}{
			"available": available,
			"enabled":   enabled,
		},
	}
}

func testServerConfig(input *Input) *Output {
	if testNginxConfig() {
		return &Output{
			Status:  "ok",
			Message: "Nginx configuration test passed",
			Facts:   map[string]interface{}{"valid": true},
		}
	}
	return &Output{
		Status:  "failed",
		Message: "Nginx configuration test failed",
		Facts:   map[string]interface{}{"valid": false},
	}
}

// ============================================================================
// Upstream and Load Balancing
// ============================================================================

func createUpstream(input *Input) *Output {
	if input.UpstreamName == "" {
		return &Output{Status: "failed", Message: "upstream_name is required", Facts: make(map[string]interface{})}
	}
	if len(input.UpstreamServers) == 0 {
		return &Output{Status: "failed", Message: "upstream_servers is required", Facts: make(map[string]interface{})}
	}

	upstreamFile := filepath.Join(input.NginxConfD, fmt.Sprintf("upstream-%s.conf", input.UpstreamName))

	// Check if already exists
	if fileExists(upstreamFile) && input.Force != "true" {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Upstream %s already exists", input.UpstreamName),
			Facts:   map[string]interface{}{"upstream_file": upstreamFile},
		}
	}

	// Generate upstream configuration
	config := generateUpstreamBlock(input)

	// Write configuration
	if err := writeFile(upstreamFile, config, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	// Test and reload
	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Created upstream: %s", input.UpstreamName),
		Facts: map[string]interface{}{
			"upstream_file": upstreamFile,
			"servers":       input.UpstreamServers,
		},
	}
}

func deleteUpstream(input *Input) *Output {
	if input.UpstreamName == "" {
		return &Output{Status: "failed", Message: "upstream_name is required", Facts: make(map[string]interface{})}
	}

	upstreamFile := filepath.Join(input.NginxConfD, fmt.Sprintf("upstream-%s.conf", input.UpstreamName))

	if !fileExists(upstreamFile) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Upstream %s does not exist", input.UpstreamName),
			Facts:   make(map[string]interface{}),
		}
	}

	if err := removeFile(upstreamFile); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to delete upstream: %v", err), Facts: make(map[string]interface{})}
	}

	if testNginxConfig() {
		reloadNginx()
	}

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Deleted upstream: %s", input.UpstreamName),
		Facts:   make(map[string]interface{}),
	}
}

func addUpstreamServer(input *Input) *Output {
	if input.UpstreamName == "" {
		return &Output{Status: "failed", Message: "upstream_name is required", Facts: make(map[string]interface{})}
	}
	if input.UpstreamServer == "" {
		return &Output{Status: "failed", Message: "upstream_server is required", Facts: make(map[string]interface{})}
	}

	upstreamFile := filepath.Join(input.NginxConfD, fmt.Sprintf("upstream-%s.conf", input.UpstreamName))

	if !fileExists(upstreamFile) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Upstream %s not found", input.UpstreamName), Facts: make(map[string]interface{})}
	}

	// Read current config
	content, err := readFile(upstreamFile)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Check if server already exists
	if strings.Contains(content, "server "+input.UpstreamServer) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server %s already in upstream %s", input.UpstreamServer, input.UpstreamName),
			Facts:   make(map[string]interface{}),
		}
	}

	// Add server line before closing brace
	newContent := strings.Replace(content, "}", fmt.Sprintf("    server %s;\n}", input.UpstreamServer), 1)

	if err := writeFile(upstreamFile, newContent, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		writeFile(upstreamFile, content, 0644) // Rollback
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Added server %s to upstream %s", input.UpstreamServer, input.UpstreamName),
		Facts:   make(map[string]interface{}),
	}
}

func removeUpstreamServer(input *Input) *Output {
	if input.UpstreamName == "" {
		return &Output{Status: "failed", Message: "upstream_name is required", Facts: make(map[string]interface{})}
	}
	if input.UpstreamServer == "" {
		return &Output{Status: "failed", Message: "upstream_server is required", Facts: make(map[string]interface{})}
	}

	upstreamFile := filepath.Join(input.NginxConfD, fmt.Sprintf("upstream-%s.conf", input.UpstreamName))

	if !fileExists(upstreamFile) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Upstream %s not found", input.UpstreamName), Facts: make(map[string]interface{})}
	}

	content, err := readFile(upstreamFile)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	if !strings.Contains(content, "server "+input.UpstreamServer) {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("Server %s not found in upstream %s", input.UpstreamServer, input.UpstreamName),
			Facts:   make(map[string]interface{}),
		}
	}

	// Remove server line
	re := regexp.MustCompile(`(?m)^\s*server\s+` + regexp.QuoteMeta(input.UpstreamServer) + `\s*;?\s*$\n?`)
	newContent := re.ReplaceAllString(content, "")

	if err := writeFile(upstreamFile, newContent, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		writeFile(upstreamFile, content, 0644) // Rollback
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Removed server %s from upstream %s", input.UpstreamServer, input.UpstreamName),
		Facts:   make(map[string]interface{}),
	}
}

func setLoadBalanceMethod(input *Input) *Output {
	if input.UpstreamName == "" {
		return &Output{Status: "failed", Message: "upstream_name is required", Facts: make(map[string]interface{})}
	}

	upstreamFile := filepath.Join(input.NginxConfD, fmt.Sprintf("upstream-%s.conf", input.UpstreamName))

	if !fileExists(upstreamFile) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Upstream %s not found", input.UpstreamName), Facts: make(map[string]interface{})}
	}

	content, err := readFile(upstreamFile)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Remove existing load balance method
	re := regexp.MustCompile(`(?m)^\s*(least_conn|ip_hash|hash|random)\s*.*$\n?`)
	newContent := re.ReplaceAllString(content, "")

	// Add new method if not round_robin (default)
	if input.LoadBalanceMethod != "round_robin" && input.LoadBalanceMethod != "" {
		var methodLine string
		switch input.LoadBalanceMethod {
		case "least_conn", "ip_hash", "random":
			methodLine = fmt.Sprintf("    %s;\n", input.LoadBalanceMethod)
		case "hash":
			key := input.HashKey
			if key == "" {
				key = "$request_uri"
			}
			methodLine = fmt.Sprintf("    hash %s;\n", key)
		default:
			return &Output{Status: "failed", Message: fmt.Sprintf("Unknown load balance method: %s", input.LoadBalanceMethod), Facts: make(map[string]interface{})}
		}

		// Insert after upstream declaration
		newContent = regexp.MustCompile(`(upstream\s+\w+\s*\{)`).ReplaceAllString(newContent, "$1\n"+methodLine)
	}

	if err := writeFile(upstreamFile, newContent, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		writeFile(upstreamFile, content, 0644)
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Set load balance method to %s for upstream %s", input.LoadBalanceMethod, input.UpstreamName),
		Facts:   make(map[string]interface{}),
	}
}

func showUpstreams(input *Input) *Output {
	files := listFiles(input.NginxConfD)
	upstreams := []string{}

	for _, file := range files {
		if strings.HasPrefix(file, "upstream-") && strings.HasSuffix(file, ".conf") {
			upstreamName := strings.TrimSuffix(strings.TrimPrefix(file, "upstream-"), ".conf")
			upstreams = append(upstreams, upstreamName)
		}
	}

	return &Output{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d upstreams", len(upstreams)),
		Facts:   map[string]interface{}{"upstreams": upstreams},
	}
}

// ============================================================================
// SSL/TLS Configuration
// ============================================================================

func createSSLServer(input *Input) *Output {
	if input.ServerName == "" {
		return &Output{Status: "failed", Message: "server_name is required", Facts: make(map[string]interface{})}
	}
	if input.SSLCertificate == "" || input.SSLCertificateKey == "" {
		return &Output{Status: "failed", Message: "ssl_certificate and ssl_certificate_key are required", Facts: make(map[string]interface{})}
	}
	if input.ServerFile == "" {
		input.ServerFile = input.ServerName + "-ssl.conf"
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)

	if fileExists(configPath) && input.Force != "true" {
		return &Output{
			Status:  "ok",
			Message: fmt.Sprintf("SSL server block %s already exists", input.ServerFile),
			Facts:   map[string]interface{}{"config_path": configPath},
		}
	}

	config := generateSSLServerBlock(input)

	if err := writeFile(configPath, config, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Created SSL server block: %s", input.ServerFile),
		Facts:   map[string]interface{}{"config_path": configPath, "server_name": input.ServerName},
	}
}

func installSSLCertificate(input *Input) *Output {
	// This is a placeholder - in real implementation, you'd copy certificate files
	if input.SSLCertificate == "" || input.SSLCertificateKey == "" {
		return &Output{Status: "failed", Message: "ssl_certificate and ssl_certificate_key are required", Facts: make(map[string]interface{})}
	}

	// Verify certificate files exist
	if !fileExists(input.SSLCertificate) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Certificate file not found: %s", input.SSLCertificate), Facts: make(map[string]interface{})}
	}
	if !fileExists(input.SSLCertificateKey) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Certificate key file not found: %s", input.SSLCertificateKey), Facts: make(map[string]interface{})}
	}

	return &Output{
		Status:  "ok",
		Message: "SSL certificate verified",
		Facts: map[string]interface{}{
			"certificate":     input.SSLCertificate,
			"certificate_key": input.SSLCertificateKey,
		},
	}
}

func setSSLProtocols(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Update or add ssl_protocols directive
	re := regexp.MustCompile(`(?m)^\s*ssl_protocols\s+.*$`)
	newLine := fmt.Sprintf("    ssl_protocols %s;", input.SSLProtocols)

	if re.MatchString(content) {
		content = re.ReplaceAllString(content, newLine)
	} else {
		// Add after ssl_certificate_key
		content = regexp.MustCompile(`(ssl_certificate_key\s+[^;]+;)`).ReplaceAllString(content, "$1\n"+newLine)
	}

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Set SSL protocols to: %s", input.SSLProtocols),
		Facts:   make(map[string]interface{}),
	}
}

func setSSLCiphers(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	re := regexp.MustCompile(`(?m)^\s*ssl_ciphers\s+.*$`)
	newLine := fmt.Sprintf("    ssl_ciphers %s;", input.SSLCiphers)

	if re.MatchString(content) {
		content = re.ReplaceAllString(content, newLine)
	} else {
		content = regexp.MustCompile(`(ssl_protocols\s+[^;]+;)`).ReplaceAllString(content, "$1\n"+newLine)
	}

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Set SSL ciphers to: %s", input.SSLCiphers),
		Facts:   make(map[string]interface{}),
	}
}

func enableHTTP2(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Check if http2 already enabled
	if strings.Contains(content, "http2") {
		return &Output{Status: "ok", Message: "HTTP/2 already enabled", Facts: make(map[string]interface{})}
	}

	// Add http2 to listen directive
	content = regexp.MustCompile(`(listen\s+443\s+ssl)`).ReplaceAllString(content, "$1 http2")

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{Status: "changed", Message: "Enabled HTTP/2", Facts: make(map[string]interface{})}
}

func testSSLConfig(input *Input) *Output {
	// Run nginx -t specifically for SSL
	if testNginxConfig() {
		return &Output{
			Status:  "ok",
			Message: "SSL configuration test passed",
			Facts:   map[string]interface{}{"valid": true},
		}
	}
	return &Output{
		Status:  "failed",
		Message: "SSL configuration test failed",
		Facts:   map[string]interface{}{"valid": false},
	}
}

// ============================================================================
// Location Blocks
// ============================================================================

func addLocationBlock(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}
	if input.LocationPath == "" {
		return &Output{Status: "failed", Message: "location_path is required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Generate location block
	locationBlock := generateLocationBlock(input)

	// Insert before closing server brace
	content = strings.Replace(content, "\n}", "\n"+locationBlock+"\n}", 1)

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Added location block: %s", input.LocationPath),
		Facts:   make(map[string]interface{}),
	}
}

func addProxyPass(input *Input) *Output {
	if input.ProxyPass == "" {
		return &Output{Status: "failed", Message: "proxy_pass is required", Facts: make(map[string]interface{})}
	}
	return addLocationBlock(input)
}

func addFastCGIPass(input *Input) *Output {
	if input.FastCGIPass == "" {
		return &Output{Status: "failed", Message: "fastcgi_pass is required", Facts: make(map[string]interface{})}
	}
	return addLocationBlock(input)
}

func addStaticLocation(input *Input) *Output {
	if input.TryFiles == "" {
		input.TryFiles = "$uri $uri/ =404"
	}
	return addLocationBlock(input)
}

func addRewriteRule(input *Input) *Output {
	if input.ServerFile == "" {
		return &Output{Status: "failed", Message: "server_file is required", Facts: make(map[string]interface{})}
	}
	if input.RewriteRegex == "" || input.RewriteReplacement == "" {
		return &Output{Status: "failed", Message: "rewrite_regex and rewrite_replacement are required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	// Generate rewrite rule
	flag := input.RewriteFlag
	if flag == "" {
		flag = "last"
	}
	rewriteLine := fmt.Sprintf("    rewrite %s %s %s;\n", input.RewriteRegex, input.RewriteReplacement, flag)

	// Add inside location block if specified, otherwise in server block
	if input.LocationPath != "" {
		// Find location block and add inside
		locationPattern := fmt.Sprintf(`(location\s+%s\s*\{)`, regexp.QuoteMeta(input.LocationPath))
		content = regexp.MustCompile(locationPattern).ReplaceAllString(content, "$1\n"+rewriteLine)
	} else {
		// Add to server block
		content = regexp.MustCompile(`(server\s*\{)`).ReplaceAllString(content, "$1\n"+rewriteLine)
	}

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Added rewrite rule: %s -> %s", input.RewriteRegex, input.RewriteReplacement),
		Facts:   make(map[string]interface{}),
	}
}

// ============================================================================
// Configuration Management
// ============================================================================

func setServerName(input *Input) *Output {
	if input.ServerFile == "" || input.ServerName == "" {
		return &Output{Status: "failed", Message: "server_file and server_name are required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	content = regexp.MustCompile(`(?m)^\s*server_name\s+.*$`).ReplaceAllString(content, fmt.Sprintf("    server_name %s;", input.ServerName))

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{Status: "changed", Message: fmt.Sprintf("Set server_name to: %s", input.ServerName), Facts: make(map[string]interface{})}
}

func setRoot(input *Input) *Output {
	if input.ServerFile == "" || input.Root == "" {
		return &Output{Status: "failed", Message: "server_file and root are required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	content = regexp.MustCompile(`(?m)^\s*root\s+.*$`).ReplaceAllString(content, fmt.Sprintf("    root %s;", input.Root))

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{Status: "changed", Message: fmt.Sprintf("Set root to: %s", input.Root), Facts: make(map[string]interface{})}
}

func setIndex(input *Input) *Output {
	if input.ServerFile == "" || input.Index == "" {
		return &Output{Status: "failed", Message: "server_file and index are required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	content = regexp.MustCompile(`(?m)^\s*index\s+.*$`).ReplaceAllString(content, fmt.Sprintf("    index %s;", input.Index))

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{Status: "changed", Message: fmt.Sprintf("Set index to: %s", input.Index), Facts: make(map[string]interface{})}
}

func addHeader(input *Input) *Output {
	if input.ServerFile == "" || input.HeaderName == "" || input.HeaderValue == "" {
		return &Output{Status: "failed", Message: "server_file, header_name, and header_value are required", Facts: make(map[string]interface{})}
	}

	configPath := filepath.Join(input.NginxSitesAvailable, input.ServerFile)
	if !fileExists(configPath) {
		return &Output{Status: "failed", Message: fmt.Sprintf("Server block %s not found", input.ServerFile), Facts: make(map[string]interface{})}
	}

	content, err := readFile(configPath)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	action := input.HeaderAction
	if action == "" {
		action = "add"
	}

	headerLine := fmt.Sprintf("    add_header %s \"%s\" always;\n", input.HeaderName, input.HeaderValue)
	if action != "add" {
		headerLine = fmt.Sprintf("    %s_header %s \"%s\";\n", action, input.HeaderName, input.HeaderValue)
	}

	// Add inside server block
	content = regexp.MustCompile(`(server\s*\{)`).ReplaceAllString(content, "$1\n"+headerLine)

	if err := writeFile(configPath, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Added header: %s: %s", input.HeaderName, input.HeaderValue),
		Facts:   make(map[string]interface{}),
	}
}

func testConfig(input *Input) *Output {
	return testServerConfig(input)
}

func reloadConfig(input *Input) *Output {
	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Configuration test failed, not reloading", Facts: make(map[string]interface{})}
	}

	if reloadNginx() {
		return &Output{Status: "changed", Message: "Nginx configuration reloaded", Facts: make(map[string]interface{})}
	}
	return &Output{Status: "failed", Message: "Failed to reload Nginx", Facts: make(map[string]interface{})}
}

// ============================================================================
// Service Management
// ============================================================================

func startService(input *Input) *Output {
	stdout, stderr, err := execCommand("systemctl", []string{"start", "nginx"}, 30)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to start nginx: %v\n%s", err, stderr), Facts: make(map[string]interface{})}
	}
	return &Output{Status: "changed", Message: "Nginx service started", Facts: map[string]interface{}{"output": stdout}}
}

func stopService(input *Input) *Output {
	stdout, stderr, err := execCommand("systemctl", []string{"stop", "nginx"}, 30)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to stop nginx: %v\n%s", err, stderr), Facts: make(map[string]interface{})}
	}
	return &Output{Status: "changed", Message: "Nginx service stopped", Facts: map[string]interface{}{"output": stdout}}
}

func restartService(input *Input) *Output {
	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Configuration test failed, not restarting", Facts: make(map[string]interface{})}
	}

	stdout, stderr, err := execCommand("systemctl", []string{"restart", "nginx"}, 30)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to restart nginx: %v\n%s", err, stderr), Facts: make(map[string]interface{})}
	}
	return &Output{Status: "changed", Message: "Nginx service restarted", Facts: map[string]interface{}{"output": stdout}}
}

func checkServiceStatus(input *Input) *Output {
	stdout, stderr, err := execCommand("systemctl", []string{"is-active", "nginx"}, 10)
	if err != nil {
		return &Output{
			Status:  "ok",
			Message: "Nginx service is not running",
			Facts:   map[string]interface{}{"running": false, "output": stderr},
		}
	}

	running := strings.TrimSpace(stdout) == "active"
	return &Output{
		Status:  "ok",
		Message: fmt.Sprintf("Nginx service is %s", stdout),
		Facts:   map[string]interface{}{"running": running, "status": stdout},
	}
}

// ============================================================================
// Performance Tuning
// ============================================================================

func setWorkerProcesses(input *Input) *Output {
	if input.WorkerProcesses == "" {
		return &Output{Status: "failed", Message: "worker_processes is required", Facts: make(map[string]interface{})}
	}

	mainConfig := filepath.Join(input.NginxConfigDir, "nginx.conf")
	content, err := readFile(mainConfig)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	content = regexp.MustCompile(`(?m)^\s*worker_processes\s+.*$`).ReplaceAllString(content, fmt.Sprintf("worker_processes %s;", input.WorkerProcesses))

	if err := writeFile(mainConfig, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Set worker_processes to: %s", input.WorkerProcesses),
		Facts:   make(map[string]interface{}),
	}
}

func setWorkerConnections(input *Input) *Output {
	if input.WorkerConnections == "" {
		return &Output{Status: "failed", Message: "worker_connections is required", Facts: make(map[string]interface{})}
	}

	mainConfig := filepath.Join(input.NginxConfigDir, "nginx.conf")
	content, err := readFile(mainConfig)
	if err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to read config: %v", err), Facts: make(map[string]interface{})}
	}

	content = regexp.MustCompile(`(?m)^\s*worker_connections\s+.*$`).ReplaceAllString(content, fmt.Sprintf("    worker_connections %s;", input.WorkerConnections))

	if err := writeFile(mainConfig, content, 0644); err != nil {
		return &Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err), Facts: make(map[string]interface{})}
	}

	if !testNginxConfig() {
		return &Output{Status: "failed", Message: "Nginx configuration test failed", Facts: make(map[string]interface{})}
	}

	reloadNginx()

	return &Output{
		Status:  "changed",
		Message: fmt.Sprintf("Set worker_connections to: %s", input.WorkerConnections),
		Facts:   make(map[string]interface{}),
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func generateServerBlock(input *Input) string {
	var sb strings.Builder

	sb.WriteString("server {\n")
	sb.WriteString(fmt.Sprintf("    listen %s;\n", input.ListenPort))
	sb.WriteString(fmt.Sprintf("    server_name %s;\n", input.ServerName))

	if input.Root != "" {
		sb.WriteString(fmt.Sprintf("    root %s;\n", input.Root))
	}
	if input.Index != "" {
		sb.WriteString(fmt.Sprintf("    index %s;\n", input.Index))
	}

	sb.WriteString("\n    location / {\n")
	sb.WriteString("        try_files $uri $uri/ =404;\n")
	sb.WriteString("    }\n")

	if input.AccessLog != "" {
		sb.WriteString(fmt.Sprintf("\n    access_log %s;\n", input.AccessLog))
	}
	if input.ErrorLog != "" {
		sb.WriteString(fmt.Sprintf("    error_log %s;\n", input.ErrorLog))
	}

	if input.AdditionalConfig != "" {
		sb.WriteString("\n")
		sb.WriteString(input.AdditionalConfig)
		sb.WriteString("\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

func generateSSLServerBlock(input *Input) string {
	var sb strings.Builder

	// HTTP redirect
	sb.WriteString("server {\n")
	sb.WriteString(fmt.Sprintf("    listen %s;\n", input.ListenPort))
	sb.WriteString(fmt.Sprintf("    server_name %s;\n", input.ServerName))
	sb.WriteString("    return 301 https://$server_name$request_uri;\n")
	sb.WriteString("}\n\n")

	// HTTPS server
	sb.WriteString("server {\n")
	if input.HTTP2 == "true" {
		sb.WriteString(fmt.Sprintf("    listen %s ssl http2;\n", input.ListenSSLPort))
	} else {
		sb.WriteString(fmt.Sprintf("    listen %s ssl;\n", input.ListenSSLPort))
	}
	sb.WriteString(fmt.Sprintf("    server_name %s;\n", input.ServerName))

	sb.WriteString(fmt.Sprintf("\n    ssl_certificate %s;\n", input.SSLCertificate))
	sb.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", input.SSLCertificateKey))

	if input.SSLProtocols != "" {
		sb.WriteString(fmt.Sprintf("    ssl_protocols %s;\n", input.SSLProtocols))
	}
	if input.SSLCiphers != "" {
		sb.WriteString(fmt.Sprintf("    ssl_ciphers %s;\n", input.SSLCiphers))
	}
	if input.SSLPreferServerCiphers != "" {
		sb.WriteString(fmt.Sprintf("    ssl_prefer_server_ciphers %s;\n", input.SSLPreferServerCiphers))
	}

	if input.HSTSEnabled == "true" {
		maxAge := input.HSTSMaxAge
		if maxAge == "" {
			maxAge = "31536000"
		}
		sb.WriteString(fmt.Sprintf("    add_header Strict-Transport-Security \"max-age=%s\" always;\n", maxAge))
	}

	if input.Root != "" {
		sb.WriteString(fmt.Sprintf("\n    root %s;\n", input.Root))
	}
	if input.Index != "" {
		sb.WriteString(fmt.Sprintf("    index %s;\n", input.Index))
	}

	sb.WriteString("\n    location / {\n")
	sb.WriteString("        try_files $uri $uri/ =404;\n")
	sb.WriteString("    }\n")

	if input.AdditionalConfig != "" {
		sb.WriteString("\n")
		sb.WriteString(input.AdditionalConfig)
		sb.WriteString("\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

func generateUpstreamBlock(input *Input) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("upstream %s {\n", input.UpstreamName))

	// Add load balancing method
	if input.LoadBalanceMethod != "" && input.LoadBalanceMethod != "round_robin" {
		switch input.LoadBalanceMethod {
		case "least_conn", "ip_hash", "random":
			sb.WriteString(fmt.Sprintf("    %s;\n", input.LoadBalanceMethod))
		case "hash":
			key := input.HashKey
			if key == "" {
				key = "$request_uri"
			}
			sb.WriteString(fmt.Sprintf("    hash %s;\n", key))
		}
	}

	// Add servers
	for _, server := range input.UpstreamServers {
		sb.WriteString(fmt.Sprintf("    server %s;\n", server))
	}

	// Add keepalive if specified
	if input.UpstreamKeepalive != "" {
		sb.WriteString(fmt.Sprintf("\n    keepalive %s;\n", input.UpstreamKeepalive))
	}

	sb.WriteString("}\n")

	return sb.String()
}

func generateLocationBlock(input *Input) string {
	var sb strings.Builder

	modifier := input.LocationModifier
	if modifier != "" {
		sb.WriteString(fmt.Sprintf("    location %s %s {\n", modifier, input.LocationPath))
	} else {
		sb.WriteString(fmt.Sprintf("    location %s {\n", input.LocationPath))
	}

	if input.ProxyPass != "" {
		sb.WriteString(fmt.Sprintf("        proxy_pass %s;\n", input.ProxyPass))
		sb.WriteString("        proxy_set_header Host $host;\n")
		sb.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		sb.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		sb.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
	} else if input.FastCGIPass != "" {
		sb.WriteString(fmt.Sprintf("        fastcgi_pass %s;\n", input.FastCGIPass))
		sb.WriteString("        fastcgi_index index.php;\n")
		sb.WriteString("        include fastcgi_params;\n")
		sb.WriteString("        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n")
	} else if input.Alias != "" {
		sb.WriteString(fmt.Sprintf("        alias %s;\n", input.Alias))
	}

	if input.TryFiles != "" {
		sb.WriteString(fmt.Sprintf("        try_files %s;\n", input.TryFiles))
	}

	if input.AdditionalConfig != "" {
		sb.WriteString("        ")
		sb.WriteString(strings.ReplaceAll(input.AdditionalConfig, "\n", "\n        "))
		sb.WriteString("\n")
	}

	sb.WriteString("    }\n")

	return sb.String()
}

func testNginxConfig() bool {
	_, _, err := execCommand("nginx", []string{"-t"}, 10)
	return err == nil
}

func reloadNginx() bool {
	_, _, err := execCommand("systemctl", []string{"reload", "nginx"}, 30)
	if err != nil {
		// Fallback to nginx -s reload
		_, _, err = execCommand("nginx", []string{"-s", "reload"}, 30)
	}
	return err == nil
}

// File operation helpers (using host API)
func fileExists(path string) bool {
	stat, err := hostFileStat(path)
	return err == nil && stat != ""
}

func isSymlink(path string) bool {
	// This is a simplified check - in real implementation would use host API
	return true
}

func readFile(path string) (string, error) {
	content, err := hostFileRead(path)
	if err != nil {
		return "", err
	}
	return content, nil
}

func writeFile(path string, content string, mode int) error {
	return hostFileWrite(path, content, mode)
}

func removeFile(path string) error {
	_, _, err := execCommand("rm", []string{"-f", path}, 10)
	return err
}

func createSymlink(target, link string) error {
	_, _, err := execCommand("ln", []string{"-s", target, link}, 10)
	return err
}

func listFiles(dir string) []string {
	stdout, _, err := execCommand("ls", []string{"-1", dir}, 10)
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	result := []string{}
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func execCommand(cmd string, args []string, timeout int) (string, string, error) {
	stdout, stderr, exitCode, err := hostExec(cmd, args, timeout)
	if err != nil {
		return stdout, stderr, err
	}
	if exitCode != 0 {
		return stdout, stderr, fmt.Errorf("command failed with exit code %d", exitCode)
	}
	return stdout, stderr, nil
}

// ============================================================================
// Host API Stubs (to be implemented by WASM runtime)
// ============================================================================

func hostExec(cmd string, args []string, timeout int) (stdout string, stderr string, exitCode int, err error) {
	// This is implemented by the WASM host
	return "", "", 0, fmt.Errorf("not implemented in stub")
}

func hostFileRead(path string) (string, error) {
	// This is implemented by the WASM host
	return "", fmt.Errorf("not implemented in stub")
}

func hostFileWrite(path string, content string, mode int) error {
	// This is implemented by the WASM host
	return fmt.Errorf("not implemented in stub")
}

func hostFileStat(path string) (string, error) {
	// This is implemented by the WASM host
	return "", fmt.Errorf("not implemented in stub")
}
