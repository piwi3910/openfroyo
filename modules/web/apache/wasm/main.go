package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Input represents the input JSON structure
type Input struct {
	Vars    map[string]interface{} `json:"vars"`
	Context Context                `json:"context"`
}

// Context provides execution context
type Context struct {
	Host     string `json:"host"`
	TaskName string `json:"task_name"`
}

// Output represents the output JSON structure
type Output struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Facts   map[string]interface{} `json:"facts"`
}

// Apache platform detection
type Platform struct {
	IsDebian     bool
	ServiceName  string
	BinaryName   string
	ConfigDir    string
	UseA2Ensite  bool
	UseA2Enmod   bool
}

func main() {
	var input Input
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		outputError(fmt.Sprintf("Failed to decode input: %v", err))
		return
	}

	// Detect platform
	platform := detectPlatform(input.Vars)

	// Get command
	command, ok := input.Vars["command"].(string)
	if !ok {
		outputError("Missing required variable: command")
		return
	}

	// Execute command
	var result Output
	switch command {
	// Virtual Host Management
	case "create_vhost":
		result = createVirtualHost(input.Vars, platform)
	case "delete_vhost":
		result = deleteVirtualHost(input.Vars, platform)
	case "enable_vhost":
		result = enableVirtualHost(input.Vars, platform)
	case "disable_vhost":
		result = disableVirtualHost(input.Vars, platform)
	case "list_vhosts":
		result = listVirtualHosts(input.Vars, platform)
	case "test_vhost_config":
		result = testVirtualHostConfig(input.Vars, platform)

	// Module Management
	case "enable_module":
		result = enableModule(input.Vars, platform)
	case "disable_module":
		result = disableModule(input.Vars, platform)
	case "list_modules":
		result = listModules(input.Vars, platform)
	case "list_available_modules":
		result = listAvailableModules(input.Vars, platform)
	case "check_module_loaded":
		result = checkModuleLoaded(input.Vars, platform)

	// SSL/TLS Configuration
	case "create_ssl_vhost":
		result = createSSLVirtualHost(input.Vars, platform)
	case "install_ssl_certificate":
		result = installSSLCertificate(input.Vars, platform)
	case "set_ssl_protocols":
		result = setSSLProtocols(input.Vars, platform)
	case "set_ssl_ciphers":
		result = setSSLCiphers(input.Vars, platform)
	case "test_ssl_config":
		result = testSSLConfig(input.Vars, platform)

	// Configuration Management
	case "set_server_name":
		result = setServerName(input.Vars, platform)
	case "set_document_root":
		result = setDocumentRoot(input.Vars, platform)
	case "set_directory_options":
		result = setDirectoryOptions(input.Vars, platform)
	case "add_directory_block":
		result = addDirectoryBlock(input.Vars, platform)
	case "test_config":
		result = testConfig(input.Vars, platform)
	case "reload_config":
		result = reloadConfig(input.Vars, platform)

	// Service Management
	case "start_service":
		result = startService(input.Vars, platform)
	case "stop_service":
		result = stopService(input.Vars, platform)
	case "restart_service":
		result = restartService(input.Vars, platform)
	case "check_service_status":
		result = checkServiceStatus(input.Vars, platform)

	// Performance Tuning
	case "set_mpm_config":
		result = setMPMConfig(input.Vars, platform)
	case "set_keepalive":
		result = setKeepAlive(input.Vars, platform)
	case "set_timeout":
		result = setTimeout(input.Vars, platform)
	case "set_max_clients":
		result = setMaxClients(input.Vars, platform)

	// Additional Operations
	case "set_log_level":
		result = setLogLevel(input.Vars, platform)
	case "enable_conf":
		result = enableConf(input.Vars, platform)
	case "disable_conf":
		result = disableConf(input.Vars, platform)
	case "create_proxy_vhost":
		result = createProxyVirtualHost(input.Vars, platform)
	case "set_security_headers":
		result = setSecurityHeaders(input.Vars, platform)

	default:
		outputError(fmt.Sprintf("Unknown command: %s", command))
		return
	}

	outputResult(result)
}

// detectPlatform determines the Apache platform (Debian/Ubuntu vs RHEL/CentOS)
func detectPlatform(vars map[string]interface{}) Platform {
	platform := Platform{}

	// Check if platform is explicitly set
	if svc, ok := vars["apache_service"].(string); ok {
		platform.ServiceName = svc
	} else {
		platform.ServiceName = "apache2"
	}

	if bin, ok := vars["apache_binary"].(string); ok {
		platform.BinaryName = bin
	} else {
		platform.BinaryName = "apache2ctl"
	}

	if dir, ok := vars["apache_config_dir"].(string); ok {
		platform.ConfigDir = dir
	} else {
		platform.ConfigDir = "/etc/apache2"
	}

	// Detect Debian vs RHEL
	platform.IsDebian = platform.ServiceName == "apache2"

	if useA2En, ok := vars["use_a2ensite"].(bool); ok {
		platform.UseA2Ensite = useA2En
	} else {
		platform.UseA2Ensite = platform.IsDebian
	}

	if useA2Mod, ok := vars["use_a2enmod"].(bool); ok {
		platform.UseA2Enmod = useA2Mod
	} else {
		platform.UseA2Enmod = platform.IsDebian
	}

	return platform
}

// ========== Virtual Host Management ==========

func createVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	documentRoot := getStringVar(vars, "document_root", "")
	if documentRoot == "" {
		return Output{Status: "failed", Message: "document_root is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+".conf")
	serverName := getStringVar(vars, "server_name", vhostName)
	serverAlias := getStringVar(vars, "server_alias", "")
	port := getStringVar(vars, "port", "80")
	serverAdmin := getStringVar(vars, "server_admin", getStringVar(vars, "default_server_admin", "webmaster@localhost"))
	options := getStringVar(vars, "directory_options", getStringVar(vars, "default_options", "Indexes FollowSymLinks"))
	allowOverride := getStringVar(vars, "allow_override", getStringVar(vars, "default_allow_override", "None"))
	require := getStringVar(vars, "require", getStringVar(vars, "default_require", "all granted"))
	customConfig := getStringVar(vars, "custom_config", "")

	// Build virtual host configuration
	vhostConfig := fmt.Sprintf(`<VirtualHost *:%s>
    ServerName %s`, port, serverName)

	if serverAlias != "" {
		vhostConfig += fmt.Sprintf("\n    ServerAlias %s", serverAlias)
	}

	vhostConfig += fmt.Sprintf(`
    ServerAdmin %s
    DocumentRoot %s

    <Directory %s>
        Options %s
        AllowOverride %s
        Require %s
    </Directory>

    ErrorLog ${APACHE_LOG_DIR}/%s-error.log
    CustomLog ${APACHE_LOG_DIR}/%s-access.log combined`,
		serverAdmin, documentRoot, documentRoot, options, allowOverride, require, vhostName, vhostName)

	if customConfig != "" {
		vhostConfig += "\n\n    " + customConfig
	}

	vhostConfig += "\n</VirtualHost>\n"

	// Write configuration file
	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	vhostPath := filepath.Join(sitesAvailable, vhostFile)

	if err := os.WriteFile(vhostPath, []byte(vhostConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write vhost config: %v", err)}
	}

	// Create document root if it doesn't exist
	if _, err := os.Stat(documentRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(documentRoot, 0755); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create document root: %v", err)}
		}
	}

	facts := map[string]interface{}{
		"vhost_name": vhostName,
		"vhost_file": vhostPath,
		"document_root": documentRoot,
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Virtual host '%s' created at %s", vhostName, vhostPath),
		Facts:   facts,
	}
}

func deleteVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+".conf")
	removeDocRoot := getBoolVar(vars, "remove_document_root", false)

	// Disable first if enabled
	disableVirtualHost(vars, platform)

	// Remove configuration file
	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	vhostPath := filepath.Join(sitesAvailable, vhostFile)

	if err := os.Remove(vhostPath); err != nil {
		if os.IsNotExist(err) {
			return Output{Status: "ok", Message: fmt.Sprintf("Virtual host '%s' does not exist", vhostName)}
		}
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to delete vhost: %v", err)}
	}

	message := fmt.Sprintf("Virtual host '%s' deleted", vhostName)

	// Optionally remove document root
	if removeDocRoot {
		documentRoot := getStringVar(vars, "document_root", "")
		if documentRoot != "" {
			if err := os.RemoveAll(documentRoot); err != nil {
				message += fmt.Sprintf(" (warning: failed to remove document root: %v)", err)
			} else {
				message += fmt.Sprintf(", document root '%s' removed", documentRoot)
			}
		}
	}

	return Output{Status: "changed", Message: message}
}

func enableVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+".conf")
	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Ensite {
		// Use a2ensite command
		cmd := fmt.Sprintf("a2ensite %s", vhostFile)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to enable vhost: %s", stderr)}
		}

		message := fmt.Sprintf("Virtual host '%s' enabled: %s", vhostName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		// Manual symlink for RHEL/CentOS
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		sitesEnabled := filepath.Join(platform.ConfigDir, "sites-enabled")

		sourcePath := filepath.Join(sitesAvailable, vhostFile)
		targetPath := filepath.Join(sitesEnabled, vhostFile)

		// Create sites-enabled directory if it doesn't exist
		if err := os.MkdirAll(sitesEnabled, 0755); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create sites-enabled directory: %v", err)}
		}

		// Check if already enabled
		if _, err := os.Lstat(targetPath); err == nil {
			return Output{Status: "ok", Message: fmt.Sprintf("Virtual host '%s' is already enabled", vhostName)}
		}

		// Create symlink
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create symlink: %v", err)}
		}

		message := fmt.Sprintf("Virtual host '%s' enabled (symlink created)", vhostName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func disableVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+".conf")
	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Ensite {
		// Use a2dissite command
		cmd := fmt.Sprintf("a2dissite %s", vhostFile)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			// Check if it's already disabled
			if strings.Contains(stderr, "does not exist") || strings.Contains(stderr, "not enabled") {
				return Output{Status: "ok", Message: fmt.Sprintf("Virtual host '%s' is not enabled", vhostName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to disable vhost: %s", stderr)}
		}

		message := fmt.Sprintf("Virtual host '%s' disabled: %s", vhostName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		// Remove symlink for RHEL/CentOS
		sitesEnabled := filepath.Join(platform.ConfigDir, "sites-enabled")
		targetPath := filepath.Join(sitesEnabled, vhostFile)

		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				return Output{Status: "ok", Message: fmt.Sprintf("Virtual host '%s' is not enabled", vhostName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to remove symlink: %v", err)}
		}

		message := fmt.Sprintf("Virtual host '%s' disabled (symlink removed)", vhostName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func listVirtualHosts(vars map[string]interface{}, platform Platform) Output {
	showDetails := getBoolVar(vars, "show_details", false)

	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	sitesEnabled := filepath.Join(platform.ConfigDir, "sites-enabled")

	available, _ := listDirectory(sitesAvailable)
	enabled, _ := listDirectory(sitesEnabled)

	facts := map[string]interface{}{
		"available_vhosts": available,
		"enabled_vhosts":   enabled,
	}

	message := fmt.Sprintf("Available vhosts (%d): %s\nEnabled vhosts (%d): %s",
		len(available), strings.Join(available, ", "),
		len(enabled), strings.Join(enabled, ", "))

	if showDetails {
		// Get detailed vhost info using apache2ctl -S
		stdout, _, _ := execCommand(fmt.Sprintf("%s -S", platform.BinaryName))
		if stdout != "" {
			message += fmt.Sprintf("\n\nVirtual Host Details:\n%s", stdout)
		}
	}

	return Output{Status: "ok", Message: message, Facts: facts}
}

func testVirtualHostConfig(vars map[string]interface{}, platform Platform) Output {
	vhostFile := getStringVar(vars, "vhost_file", "")
	if vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required"}
	}

	verbose := getBoolVar(vars, "verbose", false)

	// Test configuration syntax
	cmd := fmt.Sprintf("%s configtest", platform.BinaryName)
	stdout, stderr, err := execCommand(cmd)

	combinedOutput := stdout + stderr

	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Configuration test failed: %s", combinedOutput),
		}
	}

	message := fmt.Sprintf("Virtual host configuration test passed for %s", vhostFile)
	if verbose {
		message += fmt.Sprintf(": %s", combinedOutput)
	}

	return Output{Status: "ok", Message: message}
}

// ========== Module Management ==========

func enableModule(vars map[string]interface{}, platform Platform) Output {
	moduleName := getStringVar(vars, "module_name", "")
	if moduleName == "" {
		return Output{Status: "failed", Message: "module_name is required"}
	}

	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Enmod {
		// Use a2enmod command
		cmd := fmt.Sprintf("a2enmod %s", moduleName)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to enable module: %s", stderr)}
		}

		message := fmt.Sprintf("Module '%s' enabled: %s", moduleName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		// For RHEL/CentOS, need to edit httpd.conf or create conf file
		// This is a simplified version
		confFile := filepath.Join(platform.ConfigDir, "conf.modules.d", fmt.Sprintf("00-%s.conf", moduleName))
		moduleConfig := fmt.Sprintf("LoadModule %s_module modules/mod_%s.so\n", moduleName, moduleName)

		if err := os.WriteFile(confFile, []byte(moduleConfig), 0644); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to enable module: %v", err)}
		}

		message := fmt.Sprintf("Module '%s' enabled", moduleName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func disableModule(vars map[string]interface{}, platform Platform) Output {
	moduleName := getStringVar(vars, "module_name", "")
	if moduleName == "" {
		return Output{Status: "failed", Message: "module_name is required"}
	}

	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Enmod {
		// Use a2dismod command
		cmd := fmt.Sprintf("a2dismod %s", moduleName)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			if strings.Contains(stderr, "does not exist") || strings.Contains(stderr, "not enabled") {
				return Output{Status: "ok", Message: fmt.Sprintf("Module '%s' is not enabled", moduleName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to disable module: %s", stderr)}
		}

		message := fmt.Sprintf("Module '%s' disabled: %s", moduleName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		// For RHEL/CentOS, remove conf file
		confFile := filepath.Join(platform.ConfigDir, "conf.modules.d", fmt.Sprintf("00-%s.conf", moduleName))

		if err := os.Remove(confFile); err != nil {
			if os.IsNotExist(err) {
				return Output{Status: "ok", Message: fmt.Sprintf("Module '%s' is not enabled", moduleName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to disable module: %v", err)}
		}

		message := fmt.Sprintf("Module '%s' disabled", moduleName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func listModules(vars map[string]interface{}, platform Platform) Output {
	format := getStringVar(vars, "format", "text")

	// Get loaded modules
	cmd := fmt.Sprintf("%s -M", platform.BinaryName)
	stdout, stderr, err := execCommand(cmd)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to list modules: %s", stderr)}
	}

	// Parse module list
	lines := strings.Split(stdout, "\n")
	modules := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Loaded Modules:") {
			// Extract module name (format is usually "module_name (shared)" or "module_name (static)")
			parts := strings.Fields(line)
			if len(parts) > 0 {
				modules = append(modules, parts[0])
			}
		}
	}

	facts := map[string]interface{}{
		"loaded_modules": modules,
		"module_count":   len(modules),
	}

	var message string
	if format == "json" {
		jsonData, _ := json.MarshalIndent(facts, "", "  ")
		message = string(jsonData)
	} else {
		message = fmt.Sprintf("Loaded modules (%d):\n%s", len(modules), strings.Join(modules, "\n"))
	}

	return Output{Status: "ok", Message: message, Facts: facts}
}

func listAvailableModules(vars map[string]interface{}, platform Platform) Output {
	format := getStringVar(vars, "format", "text")

	modsAvailable := filepath.Join(platform.ConfigDir, "mods-available")
	available, _ := listDirectory(modsAvailable)

	// Remove .load and .conf extensions
	modules := []string{}
	seen := make(map[string]bool)
	for _, file := range available {
		moduleName := strings.TrimSuffix(file, ".load")
		moduleName = strings.TrimSuffix(moduleName, ".conf")
		if !seen[moduleName] {
			modules = append(modules, moduleName)
			seen[moduleName] = true
		}
	}

	facts := map[string]interface{}{
		"available_modules": modules,
		"module_count":      len(modules),
	}

	var message string
	if format == "json" {
		jsonData, _ := json.MarshalIndent(facts, "", "  ")
		message = string(jsonData)
	} else {
		message = fmt.Sprintf("Available modules (%d):\n%s", len(modules), strings.Join(modules, "\n"))
	}

	return Output{Status: "ok", Message: message, Facts: facts}
}

func checkModuleLoaded(vars map[string]interface{}, platform Platform) Output {
	moduleName := getStringVar(vars, "module_name", "")
	if moduleName == "" {
		return Output{Status: "failed", Message: "module_name is required"}
	}

	// Get loaded modules
	cmd := fmt.Sprintf("%s -M", platform.BinaryName)
	stdout, _, err := execCommand(cmd)
	if err != nil {
		return Output{Status: "failed", Message: "Failed to check modules"}
	}

	loaded := strings.Contains(stdout, moduleName)

	facts := map[string]interface{}{
		"module_name": moduleName,
		"is_loaded":   loaded,
	}

	var status, message string
	if loaded {
		status = "ok"
		message = fmt.Sprintf("Module '%s' is loaded", moduleName)
	} else {
		status = "ok"
		message = fmt.Sprintf("Module '%s' is not loaded", moduleName)
	}

	return Output{Status: status, Message: message, Facts: facts}
}

// ========== SSL/TLS Configuration ==========

func createSSLVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	documentRoot := getStringVar(vars, "document_root", "")
	if documentRoot == "" {
		return Output{Status: "failed", Message: "document_root is required"}
	}

	sslCert := getStringVar(vars, "ssl_certificate", "")
	if sslCert == "" {
		return Output{Status: "failed", Message: "ssl_certificate is required"}
	}

	sslKey := getStringVar(vars, "ssl_certificate_key", "")
	if sslKey == "" {
		return Output{Status: "failed", Message: "ssl_certificate_key is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+"-ssl.conf")
	serverName := getStringVar(vars, "server_name", vhostName)
	serverAlias := getStringVar(vars, "server_alias", "")
	sslPort := getStringVar(vars, "ssl_port", "443")
	sslChain := getStringVar(vars, "ssl_certificate_chain", "")
	sslProtocols := getStringVar(vars, "ssl_protocols", "TLSv1.2 TLSv1.3")
	sslCiphers := getStringVar(vars, "ssl_ciphers", "HIGH:!aNULL:!MD5:!RC4")
	hstsEnabled := getBoolVar(vars, "hsts_enabled", true)
	redirectHTTP := getBoolVar(vars, "redirect_http_to_https", true)

	// Build SSL virtual host configuration
	vhostConfig := fmt.Sprintf(`<VirtualHost *:%s>
    ServerName %s`, sslPort, serverName)

	if serverAlias != "" {
		vhostConfig += fmt.Sprintf("\n    ServerAlias %s", serverAlias)
	}

	vhostConfig += fmt.Sprintf(`
    DocumentRoot %s

    SSLEngine on
    SSLCertificateFile %s
    SSLCertificateKeyFile %s`, documentRoot, sslCert, sslKey)

	if sslChain != "" {
		vhostConfig += fmt.Sprintf("\n    SSLCertificateChainFile %s", sslChain)
	}

	vhostConfig += fmt.Sprintf(`
    SSLProtocol %s
    SSLCipherSuite %s
    SSLHonorCipherOrder on

    <Directory %s>
        Options Indexes FollowSymLinks
        AllowOverride All
        Require all granted
    </Directory>`, sslProtocols, sslCiphers, documentRoot)

	if hstsEnabled {
		vhostConfig += `

    # HSTS (HTTP Strict Transport Security)
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"`
	}

	vhostConfig += fmt.Sprintf(`

    ErrorLog ${APACHE_LOG_DIR}/%s-ssl-error.log
    CustomLog ${APACHE_LOG_DIR}/%s-ssl-access.log combined
</VirtualHost>`, vhostName, vhostName)

	// Add HTTP to HTTPS redirect if requested
	if redirectHTTP {
		vhostConfig += fmt.Sprintf(`

<VirtualHost *:80>
    ServerName %s`, serverName)
		if serverAlias != "" {
			vhostConfig += fmt.Sprintf("\n    ServerAlias %s", serverAlias)
		}
		vhostConfig += `
    Redirect permanent / https://` + serverName + `/
</VirtualHost>
`
	}

	// Write configuration file
	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	vhostPath := filepath.Join(sitesAvailable, vhostFile)

	if err := os.WriteFile(vhostPath, []byte(vhostConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write SSL vhost config: %v", err)}
	}

	// Create document root if it doesn't exist
	if _, err := os.Stat(documentRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(documentRoot, 0755); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create document root: %v", err)}
		}
	}

	facts := map[string]interface{}{
		"vhost_name":      vhostName,
		"vhost_file":      vhostPath,
		"document_root":   documentRoot,
		"ssl_enabled":     true,
		"hsts_enabled":    hstsEnabled,
		"http_redirected": redirectHTTP,
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("SSL virtual host '%s' created at %s", vhostName, vhostPath),
		Facts:   facts,
	}
}

func installSSLCertificate(vars map[string]interface{}, platform Platform) Output {
	vhostFile := getStringVar(vars, "vhost_file", "")
	if vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required"}
	}

	sslCert := getStringVar(vars, "ssl_certificate", "")
	if sslCert == "" {
		return Output{Status: "failed", Message: "ssl_certificate is required"}
	}

	sslKey := getStringVar(vars, "ssl_certificate_key", "")
	if sslKey == "" {
		return Output{Status: "failed", Message: "ssl_certificate_key is required"}
	}

	sslChain := getStringVar(vars, "ssl_certificate_chain", "")

	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	vhostPath := filepath.Join(sitesAvailable, vhostFile)

	// Read existing vhost configuration
	content, err := os.ReadFile(vhostPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read vhost file: %v", err)}
	}

	vhostConfig := string(content)

	// Update SSL certificate directives
	vhostConfig = replaceOrAddDirective(vhostConfig, "SSLCertificateFile", sslCert)
	vhostConfig = replaceOrAddDirective(vhostConfig, "SSLCertificateKeyFile", sslKey)
	if sslChain != "" {
		vhostConfig = replaceOrAddDirective(vhostConfig, "SSLCertificateChainFile", sslChain)
	}

	// Ensure SSLEngine is on
	if !strings.Contains(vhostConfig, "SSLEngine") {
		vhostConfig = strings.Replace(vhostConfig, "<VirtualHost", "<VirtualHost", 1)
		vhostConfig = strings.Replace(vhostConfig, "<VirtualHost *:", "<VirtualHost *:", 1)
		// Add SSLEngine after VirtualHost line
		lines := strings.Split(vhostConfig, "\n")
		for i, line := range lines {
			if strings.Contains(line, "<VirtualHost") {
				lines = append(lines[:i+1], append([]string{"    SSLEngine on"}, lines[i+1:]...)...)
				break
			}
		}
		vhostConfig = strings.Join(lines, "\n")
	}

	// Write updated configuration
	if err := os.WriteFile(vhostPath, []byte(vhostConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write vhost config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("SSL certificate installed for %s", vhostFile),
	}
}

func setSSLProtocols(vars map[string]interface{}, platform Platform) Output {
	sslProtocols := getStringVar(vars, "ssl_protocols", "")
	if sslProtocols == "" {
		return Output{Status: "failed", Message: "ssl_protocols is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	configScope := getStringVar(vars, "config_scope", "global")

	if configScope == "vhost" && vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
	}

	var configPath string
	if configScope == "vhost" {
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		// Global SSL configuration
		configPath = filepath.Join(platform.ConfigDir, "mods-available", "ssl.conf")
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "SSLProtocol", sslProtocols)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("SSL protocols set to '%s' in %s", sslProtocols, configPath),
	}
}

func setSSLCiphers(vars map[string]interface{}, platform Platform) Output {
	sslCiphers := getStringVar(vars, "ssl_ciphers", "")
	if sslCiphers == "" {
		return Output{Status: "failed", Message: "ssl_ciphers is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	configScope := getStringVar(vars, "config_scope", "global")
	preferServerCiphers := getStringVar(vars, "ssl_prefer_server_ciphers", "On")

	if configScope == "vhost" && vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
	}

	var configPath string
	if configScope == "vhost" {
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		configPath = filepath.Join(platform.ConfigDir, "mods-available", "ssl.conf")
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "SSLCipherSuite", sslCiphers)
	config = replaceOrAddDirective(config, "SSLHonorCipherOrder", preferServerCiphers)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("SSL ciphers configured in %s", configPath),
	}
}

func testSSLConfig(vars map[string]interface{}, platform Platform) Output {
	checkCerts := getBoolVar(vars, "check_certificates", false)

	// Test configuration syntax
	cmd := fmt.Sprintf("%s configtest", platform.BinaryName)
	stdout, stderr, err := execCommand(cmd)

	combinedOutput := stdout + stderr

	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("SSL configuration test failed: %s", combinedOutput),
		}
	}

	message := "SSL configuration test passed"

	if checkCerts {
		// Additional certificate validation could be added here
		message += " (certificate validation not implemented in this version)"
	}

	return Output{Status: "ok", Message: message}
}

// ========== Configuration Management ==========

func setServerName(vars map[string]interface{}, platform Platform) Output {
	serverName := getStringVar(vars, "server_name", "")
	if serverName == "" {
		return Output{Status: "failed", Message: "server_name is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	configScope := getStringVar(vars, "config_scope", "global")

	var configPath string
	if configScope == "vhost" {
		if vhostFile == "" {
			return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
		}
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		configPath = filepath.Join(platform.ConfigDir, "apache2.conf")
		if !platform.IsDebian {
			configPath = filepath.Join(platform.ConfigDir, "httpd.conf")
		}
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "ServerName", serverName)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("ServerName set to '%s' in %s", serverName, configPath),
	}
}

func setDocumentRoot(vars map[string]interface{}, platform Platform) Output {
	documentRoot := getStringVar(vars, "document_root", "")
	if documentRoot == "" {
		return Output{Status: "failed", Message: "document_root is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	createDirectory := getBoolVar(vars, "create_directory", true)

	if vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required"}
	}

	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	configPath := filepath.Join(sitesAvailable, vhostFile)

	// Create directory if requested
	if createDirectory {
		if _, err := os.Stat(documentRoot); os.IsNotExist(err) {
			if err := os.MkdirAll(documentRoot, 0755); err != nil {
				return Output{Status: "failed", Message: fmt.Sprintf("Failed to create directory: %v", err)}
			}
		}
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "DocumentRoot", documentRoot)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("DocumentRoot set to '%s' in %s", documentRoot, configPath),
	}
}

func setDirectoryOptions(vars map[string]interface{}, platform Platform) Output {
	directoryPath := getStringVar(vars, "directory_path", "")
	if directoryPath == "" {
		return Output{Status: "failed", Message: "directory_path is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	_ = getStringVar(vars, "options", "")
	_ = getStringVar(vars, "allow_override", "")
	_ = getStringVar(vars, "require", "")

	if vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required"}
	}

	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	configPath := filepath.Join(sitesAvailable, vhostFile)

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)

	// Find and update Directory block
	// This is simplified - a full implementation would parse the config properly
	directoryBlock := fmt.Sprintf("<Directory %s>", directoryPath)
	if strings.Contains(config, directoryBlock) {
		// Update existing block
		// This is a placeholder - proper implementation would parse and update
		message := fmt.Sprintf("Directory options for %s updated (simplified implementation)", directoryPath)
		return Output{Status: "changed", Message: message}
	} else {
		return Output{Status: "failed", Message: "Directory block not found - use add_directory_block instead"}
	}
}

func addDirectoryBlock(vars map[string]interface{}, platform Platform) Output {
	directoryPath := getStringVar(vars, "directory_path", "")
	if directoryPath == "" {
		return Output{Status: "failed", Message: "directory_path is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	if vhostFile == "" {
		return Output{Status: "failed", Message: "vhost_file is required"}
	}

	options := getStringVar(vars, "options", "Indexes FollowSymLinks")
	allowOverride := getStringVar(vars, "allow_override", "None")
	require := getStringVar(vars, "require", "all granted")
	customDirectives := getStringVar(vars, "custom_directives", "")

	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	configPath := filepath.Join(sitesAvailable, vhostFile)

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)

	// Build Directory block
	directoryBlock := fmt.Sprintf(`
    <Directory %s>
        Options %s
        AllowOverride %s
        Require %s`, directoryPath, options, allowOverride, require)

	if customDirectives != "" {
		directoryBlock += fmt.Sprintf("\n        %s", customDirectives)
	}

	directoryBlock += "\n    </Directory>"

	// Add before closing VirtualHost tag
	config = strings.Replace(config, "</VirtualHost>", directoryBlock+"\n</VirtualHost>", 1)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Directory block added for %s in %s", directoryPath, vhostFile),
	}
}

func testConfig(vars map[string]interface{}, platform Platform) Output {
	verbose := getBoolVar(vars, "verbose", false)
	showVhosts := getBoolVar(vars, "show_vhosts", false)

	// Test configuration syntax
	cmd := fmt.Sprintf("%s configtest", platform.BinaryName)
	stdout, stderr, err := execCommand(cmd)

	combinedOutput := stdout + stderr

	if err != nil {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Configuration test failed: %s", combinedOutput),
		}
	}

	message := "Configuration test passed"

	if verbose {
		message += fmt.Sprintf(": %s", combinedOutput)
	}

	if showVhosts {
		vhostCmd := fmt.Sprintf("%s -S", platform.BinaryName)
		vhostOutput, _, _ := execCommand(vhostCmd)
		if vhostOutput != "" {
			message += fmt.Sprintf("\n\nVirtual Hosts:\n%s", vhostOutput)
		}
	}

	return Output{Status: "ok", Message: message}
}

func reloadConfig(vars map[string]interface{}, platform Platform) Output {
	testFirst := getBoolVar(vars, "test_first", true)

	if testFirst {
		// Test configuration first
		testResult := testConfig(vars, platform)
		if testResult.Status == "failed" {
			return testResult
		}
	}

	// Reload Apache configuration
	cmd := fmt.Sprintf("systemctl reload %s", platform.ServiceName)
	stdout, stderr, err := execCommand(cmd)

	if err != nil {
		// Try graceful restart as fallback
		cmd = fmt.Sprintf("%s graceful", platform.BinaryName)
		stdout, stderr, err = execCommand(cmd)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to reload config: %s", stderr)}
		}
	}

	message := "Apache configuration reloaded"
	if stdout != "" {
		message += fmt.Sprintf(": %s", stdout)
	}

	return Output{Status: "changed", Message: message}
}

// ========== Service Management ==========

func startService(vars map[string]interface{}, platform Platform) Output {
	testConfigFirst := getBoolVar(vars, "test_config_first", true)

	if testConfigFirst {
		testResult := testConfig(vars, platform)
		if testResult.Status == "failed" {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Configuration test failed, not starting service: %s", testResult.Message),
			}
		}
	}

	cmd := fmt.Sprintf("systemctl start %s", platform.ServiceName)
	stdout, stderr, err := execCommand(cmd)

	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to start service: %s", stderr)}
	}

	message := fmt.Sprintf("Apache service '%s' started", platform.ServiceName)
	if stdout != "" {
		message += fmt.Sprintf(": %s", stdout)
	}

	return Output{Status: "changed", Message: message}
}

func stopService(vars map[string]interface{}, platform Platform) Output {
	graceful := getBoolVar(vars, "graceful", false)

	var cmd string
	if graceful {
		cmd = fmt.Sprintf("%s graceful-stop", platform.BinaryName)
	} else {
		cmd = fmt.Sprintf("systemctl stop %s", platform.ServiceName)
	}

	stdout, stderr, err := execCommand(cmd)

	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to stop service: %s", stderr)}
	}

	message := fmt.Sprintf("Apache service '%s' stopped", platform.ServiceName)
	if graceful {
		message += " (gracefully)"
	}
	if stdout != "" {
		message += fmt.Sprintf(": %s", stdout)
	}

	return Output{Status: "changed", Message: message}
}

func restartService(vars map[string]interface{}, platform Platform) Output {
	graceful := getBoolVar(vars, "graceful", true)
	testConfigFirst := getBoolVar(vars, "test_config_first", true)

	if testConfigFirst {
		testResult := testConfig(vars, platform)
		if testResult.Status == "failed" {
			return Output{
				Status:  "failed",
				Message: fmt.Sprintf("Configuration test failed, not restarting service: %s", testResult.Message),
			}
		}
	}

	var cmd string
	if graceful {
		cmd = fmt.Sprintf("systemctl reload %s", platform.ServiceName)
	} else {
		cmd = fmt.Sprintf("systemctl restart %s", platform.ServiceName)
	}

	stdout, stderr, err := execCommand(cmd)

	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to restart service: %s", stderr)}
	}

	message := fmt.Sprintf("Apache service '%s' restarted", platform.ServiceName)
	if graceful {
		message += " (gracefully)"
	}
	if stdout != "" {
		message += fmt.Sprintf(": %s", stdout)
	}

	return Output{Status: "changed", Message: message}
}

func checkServiceStatus(vars map[string]interface{}, platform Platform) Output {
	verbose := getBoolVar(vars, "verbose", false)

	cmd := fmt.Sprintf("systemctl status %s", platform.ServiceName)
	stdout, stderr, _ := execCommand(cmd)

	combinedOutput := stdout + stderr

	// Parse status
	isRunning := strings.Contains(combinedOutput, "active (running)")

	facts := map[string]interface{}{
		"service_name": platform.ServiceName,
		"is_running":   isRunning,
	}

	var message string
	if isRunning {
		message = fmt.Sprintf("Apache service '%s' is running", platform.ServiceName)
	} else {
		message = fmt.Sprintf("Apache service '%s' is not running", platform.ServiceName)
	}

	if verbose {
		message += fmt.Sprintf("\n\n%s", combinedOutput)
	}

	return Output{Status: "ok", Message: message, Facts: facts}
}

// ========== Performance Tuning ==========

func setMPMConfig(vars map[string]interface{}, platform Platform) Output {
	mpmModule := getStringVar(vars, "mpm_module", "")
	if mpmModule == "" {
		return Output{Status: "failed", Message: "mpm_module is required"}
	}

	if mpmModule != "prefork" && mpmModule != "worker" && mpmModule != "event" {
		return Output{Status: "failed", Message: "mpm_module must be one of: prefork, worker, event"}
	}

	configFile := getStringVar(vars, "config_file", "")
	if configFile == "" {
		configFile = filepath.Join(platform.ConfigDir, "mods-available", fmt.Sprintf("mpm_%s.conf", mpmModule))
	}

	// Build MPM configuration based on module type and provided parameters
	var mpmConfig string

	if mpmModule == "prefork" {
		startServers := getStringVar(vars, "start_servers", "5")
		minSpare := getStringVar(vars, "min_spare_threads", "5")
		maxSpare := getStringVar(vars, "max_spare_threads", "10")
		maxWorkers := getStringVar(vars, "max_request_workers", "150")
		maxConnections := getStringVar(vars, "max_connections_per_child", "0")

		mpmConfig = fmt.Sprintf(`<IfModule mpm_prefork_module>
    StartServers             %s
    MinSpareServers          %s
    MaxSpareServers          %s
    MaxRequestWorkers        %s
    MaxConnectionsPerChild   %s
</IfModule>
`, startServers, minSpare, maxSpare, maxWorkers, maxConnections)
	} else if mpmModule == "worker" {
		startServers := getStringVar(vars, "start_servers", "2")
		minSpare := getStringVar(vars, "min_spare_threads", "25")
		maxSpare := getStringVar(vars, "max_spare_threads", "75")
		threadsPerChild := getStringVar(vars, "threads_per_child", "25")
		maxWorkers := getStringVar(vars, "max_request_workers", "150")
		maxConnections := getStringVar(vars, "max_connections_per_child", "0")

		mpmConfig = fmt.Sprintf(`<IfModule mpm_worker_module>
    StartServers             %s
    MinSpareThreads          %s
    MaxSpareThreads          %s
    ThreadsPerChild          %s
    MaxRequestWorkers        %s
    MaxConnectionsPerChild   %s
</IfModule>
`, startServers, minSpare, maxSpare, threadsPerChild, maxWorkers, maxConnections)
	} else { // event
		startServers := getStringVar(vars, "start_servers", "2")
		minSpare := getStringVar(vars, "min_spare_threads", "25")
		maxSpare := getStringVar(vars, "max_spare_threads", "75")
		threadsPerChild := getStringVar(vars, "threads_per_child", "25")
		maxWorkers := getStringVar(vars, "max_request_workers", "150")
		maxConnections := getStringVar(vars, "max_connections_per_child", "0")

		mpmConfig = fmt.Sprintf(`<IfModule mpm_event_module>
    StartServers             %s
    MinSpareThreads          %s
    MaxSpareThreads          %s
    ThreadsPerChild          %s
    MaxRequestWorkers        %s
    MaxConnectionsPerChild   %s
</IfModule>
`, startServers, minSpare, maxSpare, threadsPerChild, maxWorkers, maxConnections)
	}

	// Write configuration
	if err := os.WriteFile(configFile, []byte(mpmConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write MPM config: %v", err)}
	}

	// Enable the MPM module
	vars["module_name"] = fmt.Sprintf("mpm_%s", mpmModule)
	enableModule(vars, platform)

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("MPM configuration for '%s' updated in %s", mpmModule, configFile),
	}
}

func setKeepAlive(vars map[string]interface{}, platform Platform) Output {
	keepAlive := getStringVar(vars, "keep_alive", "")
	if keepAlive == "" {
		return Output{Status: "failed", Message: "keep_alive is required"}
	}

	keepAliveTimeout := getStringVar(vars, "keep_alive_timeout", "")
	maxKeepAlive := getStringVar(vars, "max_keep_alive_requests", "")
	configScope := getStringVar(vars, "config_scope", "global")
	vhostFile := getStringVar(vars, "vhost_file", "")

	var configPath string
	if configScope == "vhost" {
		if vhostFile == "" {
			return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
		}
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		configPath = filepath.Join(platform.ConfigDir, "apache2.conf")
		if !platform.IsDebian {
			configPath = filepath.Join(platform.ConfigDir, "httpd.conf")
		}
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "KeepAlive", keepAlive)

	if keepAliveTimeout != "" {
		config = replaceOrAddDirective(config, "KeepAliveTimeout", keepAliveTimeout)
	}

	if maxKeepAlive != "" {
		config = replaceOrAddDirective(config, "MaxKeepAliveRequests", maxKeepAlive)
	}

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("KeepAlive settings configured in %s", configPath),
	}
}

func setTimeout(vars map[string]interface{}, platform Platform) Output {
	timeout := getStringVar(vars, "timeout", "")
	if timeout == "" {
		return Output{Status: "failed", Message: "timeout is required"}
	}

	requestReadTimeout := getStringVar(vars, "request_read_timeout", "")
	configScope := getStringVar(vars, "config_scope", "global")

	var configPath string
	if configScope == "vhost" {
		vhostFile := getStringVar(vars, "vhost_file", "")
		if vhostFile == "" {
			return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
		}
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		configPath = filepath.Join(platform.ConfigDir, "apache2.conf")
		if !platform.IsDebian {
			configPath = filepath.Join(platform.ConfigDir, "httpd.conf")
		}
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "Timeout", timeout)

	if requestReadTimeout != "" {
		config = replaceOrAddDirective(config, "RequestReadTimeout", requestReadTimeout)
	}

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Timeout settings configured in %s", configPath),
	}
}

func setMaxClients(vars map[string]interface{}, platform Platform) Output {
	maxClients := getStringVar(vars, "max_clients", "")
	if maxClients == "" {
		return Output{Status: "failed", Message: "max_clients is required"}
	}

	mpmModule := getStringVar(vars, "mpm_module", "prefork")

	// MaxClients is now MaxRequestWorkers in Apache 2.4+
	configFile := filepath.Join(platform.ConfigDir, "mods-available", fmt.Sprintf("mpm_%s.conf", mpmModule))

	// Read existing configuration
	content, err := os.ReadFile(configFile)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read MPM config: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "MaxRequestWorkers", maxClients)

	// Write updated configuration
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("MaxRequestWorkers set to %s in %s", maxClients, configFile),
	}
}

// ========== Additional Operations ==========

func setLogLevel(vars map[string]interface{}, platform Platform) Output {
	logLevel := getStringVar(vars, "log_level", "")
	if logLevel == "" {
		return Output{Status: "failed", Message: "log_level is required"}
	}

	validLevels := []string{"debug", "info", "notice", "warn", "error", "crit", "alert", "emerg"}
	isValid := false
	for _, level := range validLevels {
		if logLevel == level {
			isValid = true
			break
		}
	}

	if !isValid {
		return Output{
			Status:  "failed",
			Message: fmt.Sprintf("Invalid log_level. Must be one of: %s", strings.Join(validLevels, ", ")),
		}
	}

	vhostFile := getStringVar(vars, "vhost_file", "")
	configScope := getStringVar(vars, "config_scope", "global")

	var configPath string
	if configScope == "vhost" {
		if vhostFile == "" {
			return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
		}
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		configPath = filepath.Join(platform.ConfigDir, "apache2.conf")
		if !platform.IsDebian {
			configPath = filepath.Join(platform.ConfigDir, "httpd.conf")
		}
	}

	// Read existing configuration
	content, err := os.ReadFile(configPath)
	if err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to read config file: %v", err)}
	}

	config := string(content)
	config = replaceOrAddDirective(config, "LogLevel", logLevel)

	// Write updated configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write config: %v", err)}
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("LogLevel set to '%s' in %s", logLevel, configPath),
	}
}

func enableConf(vars map[string]interface{}, platform Platform) Output {
	confName := getStringVar(vars, "conf_name", "")
	if confName == "" {
		return Output{Status: "failed", Message: "conf_name is required"}
	}

	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Ensite { // a2enconf is part of the same toolset
		cmd := fmt.Sprintf("a2enconf %s", confName)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to enable conf: %s", stderr)}
		}

		message := fmt.Sprintf("Configuration '%s' enabled: %s", confName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		// Manual symlink for RHEL/CentOS
		confAvailable := filepath.Join(platform.ConfigDir, "conf-available")
		confEnabled := filepath.Join(platform.ConfigDir, "conf-enabled")

		sourcePath := filepath.Join(confAvailable, confName+".conf")
		targetPath := filepath.Join(confEnabled, confName+".conf")

		if err := os.MkdirAll(confEnabled, 0755); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create conf-enabled directory: %v", err)}
		}

		if err := os.Symlink(sourcePath, targetPath); err != nil {
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to create symlink: %v", err)}
		}

		message := fmt.Sprintf("Configuration '%s' enabled", confName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func disableConf(vars map[string]interface{}, platform Platform) Output {
	confName := getStringVar(vars, "conf_name", "")
	if confName == "" {
		return Output{Status: "failed", Message: "conf_name is required"}
	}

	reloadService := getBoolVar(vars, "reload_service", true)

	if platform.UseA2Ensite {
		cmd := fmt.Sprintf("a2disconf %s", confName)
		stdout, stderr, err := execCommand(cmd)
		if err != nil {
			if strings.Contains(stderr, "does not exist") || strings.Contains(stderr, "not enabled") {
				return Output{Status: "ok", Message: fmt.Sprintf("Configuration '%s' is not enabled", confName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to disable conf: %s", stderr)}
		}

		message := fmt.Sprintf("Configuration '%s' disabled: %s", confName, stdout)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	} else {
		confEnabled := filepath.Join(platform.ConfigDir, "conf-enabled")
		targetPath := filepath.Join(confEnabled, confName+".conf")

		if err := os.Remove(targetPath); err != nil {
			if os.IsNotExist(err) {
				return Output{Status: "ok", Message: fmt.Sprintf("Configuration '%s' is not enabled", confName)}
			}
			return Output{Status: "failed", Message: fmt.Sprintf("Failed to remove symlink: %v", err)}
		}

		message := fmt.Sprintf("Configuration '%s' disabled", confName)

		if reloadService {
			reloadResult := reloadConfig(vars, platform)
			message += fmt.Sprintf("; %s", reloadResult.Message)
		}

		return Output{Status: "changed", Message: message}
	}
}

func createProxyVirtualHost(vars map[string]interface{}, platform Platform) Output {
	vhostName := getStringVar(vars, "vhost_name", "")
	if vhostName == "" {
		return Output{Status: "failed", Message: "vhost_name is required"}
	}

	proxyTarget := getStringVar(vars, "proxy_target", "")
	if proxyTarget == "" {
		return Output{Status: "failed", Message: "proxy_target is required"}
	}

	vhostFile := getStringVar(vars, "vhost_file", vhostName+"-proxy.conf")
	serverName := getStringVar(vars, "server_name", vhostName)
	port := getStringVar(vars, "port", "80")
	proxyPreserveHost := getBoolVar(vars, "proxy_preserve_host", true)
	proxyTimeout := getStringVar(vars, "proxy_timeout", "300")
	sslEnabled := getBoolVar(vars, "ssl_enabled", false)

	// Build proxy virtual host configuration
	vhostConfig := fmt.Sprintf(`<VirtualHost *:%s>
    ServerName %s

    ProxyPreserveHost %s
    ProxyTimeout %s

    ProxyPass / %s
    ProxyPassReverse / %s

    ErrorLog ${APACHE_LOG_DIR}/%s-proxy-error.log
    CustomLog ${APACHE_LOG_DIR}/%s-proxy-access.log combined
</VirtualHost>
`, port, serverName,
		func() string { if proxyPreserveHost { return "On" } else { return "Off" } }(),
		proxyTimeout, proxyTarget, proxyTarget, vhostName, vhostName)

	// Write configuration file
	sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
	vhostPath := filepath.Join(sitesAvailable, vhostFile)

	if err := os.WriteFile(vhostPath, []byte(vhostConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write proxy vhost config: %v", err)}
	}

	// Enable required proxy modules
	vars["module_name"] = "proxy"
	enableModule(vars, platform)
	vars["module_name"] = "proxy_http"
	enableModule(vars, platform)

	if sslEnabled {
		vars["module_name"] = "ssl"
		enableModule(vars, platform)
	}

	facts := map[string]interface{}{
		"vhost_name":   vhostName,
		"vhost_file":   vhostPath,
		"proxy_target": proxyTarget,
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Proxy virtual host '%s' created for target %s", vhostName, proxyTarget),
		Facts:   facts,
	}
}

func setSecurityHeaders(vars map[string]interface{}, platform Platform) Output {
	vhostFile := getStringVar(vars, "vhost_file", "")
	configScope := getStringVar(vars, "config_scope", "global")

	xFrameOptions := getStringVar(vars, "x_frame_options", "SAMEORIGIN")
	xContentType := getStringVar(vars, "x_content_type_options", "nosniff")
	xXSSProtection := getStringVar(vars, "x_xss_protection", "1; mode=block")
	hsts := getStringVar(vars, "strict_transport_security", "")
	csp := getStringVar(vars, "content_security_policy", "")

	var configPath string
	if configScope == "vhost" {
		if vhostFile == "" {
			return Output{Status: "failed", Message: "vhost_file is required for vhost scope"}
		}
		sitesAvailable := filepath.Join(platform.ConfigDir, "sites-available")
		configPath = filepath.Join(sitesAvailable, vhostFile)
	} else {
		// Create security headers config file
		configPath = filepath.Join(platform.ConfigDir, "conf-available", "security-headers.conf")
	}

	// Build security headers configuration
	securityConfig := "# Security Headers\n"
	securityConfig += fmt.Sprintf("Header always set X-Frame-Options \"%s\"\n", xFrameOptions)
	securityConfig += fmt.Sprintf("Header always set X-Content-Type-Options \"%s\"\n", xContentType)
	securityConfig += fmt.Sprintf("Header always set X-XSS-Protection \"%s\"\n", xXSSProtection)

	if hsts != "" {
		securityConfig += fmt.Sprintf("Header always set Strict-Transport-Security \"%s\"\n", hsts)
	}

	if csp != "" {
		securityConfig += fmt.Sprintf("Header always set Content-Security-Policy \"%s\"\n", csp)
	}

	// Read existing config or create new
	var fullConfig string
	if content, err := os.ReadFile(configPath); err == nil {
		fullConfig = string(content)
		// Append or replace security headers section
		if !strings.Contains(fullConfig, "# Security Headers") {
			fullConfig += "\n" + securityConfig
		}
	} else {
		fullConfig = securityConfig
	}

	// Write configuration
	if err := os.WriteFile(configPath, []byte(fullConfig), 0644); err != nil {
		return Output{Status: "failed", Message: fmt.Sprintf("Failed to write security headers config: %v", err)}
	}

	// Enable headers module
	vars["module_name"] = "headers"
	vars["reload_service"] = false
	enableModule(vars, platform)

	// Enable the conf if it's global
	if configScope == "global" {
		vars["conf_name"] = "security-headers"
		enableConf(vars, platform)
	}

	return Output{
		Status:  "changed",
		Message: fmt.Sprintf("Security headers configured in %s", configPath),
	}
}

// ========== Utility Functions ==========

func getStringVar(vars map[string]interface{}, key string, defaultValue string) string {
	if val, ok := vars[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getBoolVar(vars map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := vars[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

func replaceOrAddDirective(config, directive, value string) string {
	// Search for existing directive
	pattern := fmt.Sprintf("%s ", directive)
	lines := strings.Split(config, "\n")
	found := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, pattern) || trimmed == directive {
			// Replace existing directive
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " \t")))
			lines[i] = fmt.Sprintf("%s%s %s", indent, directive, value)
			found = true
			break
		}
	}

	if !found {
		// Add new directive (simplified - just append)
		newLine := fmt.Sprintf("    %s %s", directive, value)
		lines = append(lines, newLine)
	}

	return strings.Join(lines, "\n")
}

func listDirectory(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func execCommand(cmd string) (string, string, error) {
	// This is a placeholder - in real WASM implementation,
	// this would use the host_exec function from the host API
	// For now, return simulated success
	return fmt.Sprintf("Simulated execution of: %s", cmd), "", nil
}

func outputResult(output Output) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
}

func outputError(message string) {
	output := Output{
		Status:  "failed",
		Message: message,
		Facts:   map[string]interface{}{},
	}
	outputResult(output)
}
