package host

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openfroyo/openfroyo/pkg/engine"
)

// CapabilityEnforcer enforces capability restrictions for WASM providers.
type CapabilityEnforcer struct {
	// grantedCapabilities is the set of capabilities granted to this provider.
	grantedCapabilities map[string]bool

	// httpClient is the HTTP client for net:outbound capability.
	httpClient *http.Client

	// tempDir is the temporary directory for fs:temp capability.
	tempDir string

	// microRunnerPath is the path to the micro-runner executable.
	microRunnerPath string

	// secretsDecryptor is the function to decrypt secrets.
	secretsDecryptor func(encrypted string) (string, error)
}

// NewCapabilityEnforcer creates a new capability enforcer.
func NewCapabilityEnforcer(capabilities []string, tempDir, microRunnerPath string) *CapabilityEnforcer {
	enforcer := &CapabilityEnforcer{
		grantedCapabilities: make(map[string]bool),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tempDir:         tempDir,
		microRunnerPath: microRunnerPath,
	}

	// Build capability set
	for _, cap := range capabilities {
		enforcer.grantedCapabilities[cap] = true
	}

	return enforcer
}

// SetSecretsDecryptor sets the secrets decryption function.
func (e *CapabilityEnforcer) SetSecretsDecryptor(fn func(string) (string, error)) {
	e.secretsDecryptor = fn
}

// HasCapability checks if a capability is granted.
func (e *CapabilityEnforcer) HasCapability(capability engine.ProviderCapability) bool {
	return e.grantedCapabilities[string(capability)]
}

// ValidateCapabilities validates that all requested capabilities are allowed.
func (e *CapabilityEnforcer) ValidateCapabilities(requested []string) error {
	var missing []string
	for _, cap := range requested {
		if !e.grantedCapabilities[cap] {
			missing = append(missing, cap)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required capabilities: %v", missing)
	}

	return nil
}

// HTTPRequest performs an HTTP request if net:outbound capability is granted.
func (e *CapabilityEnforcer) HTTPRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	if !e.HasCapability(engine.CapabilityNetOutbound) {
		return nil, fmt.Errorf("capability net:outbound not granted")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}

// CreateTempFile creates a temporary file if fs:temp capability is granted.
func (e *CapabilityEnforcer) CreateTempFile(name string) (*os.File, error) {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return nil, fmt.Errorf("capability fs:temp not granted")
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(e.tempDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temp file
	filePath := filepath.Join(e.tempDir, name)

	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(filePath), e.tempDir) {
		return nil, fmt.Errorf("invalid file path: path traversal detected")
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	return file, nil
}

// ReadTempFile reads a temporary file if fs:temp capability is granted.
func (e *CapabilityEnforcer) ReadTempFile(name string) ([]byte, error) {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return nil, fmt.Errorf("capability fs:temp not granted")
	}

	filePath := filepath.Join(e.tempDir, name)

	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(filePath), e.tempDir) {
		return nil, fmt.Errorf("invalid file path: path traversal detected")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	return data, nil
}

// WriteTempFile writes data to a temporary file if fs:temp capability is granted.
func (e *CapabilityEnforcer) WriteTempFile(name string, data []byte) error {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return fmt.Errorf("capability fs:temp not granted")
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(e.tempDir, 0750); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	filePath := filepath.Join(e.tempDir, name)

	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(filePath), e.tempDir) {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	if err := os.WriteFile(filePath, data, 0640); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	return nil
}

// DeleteTempFile deletes a temporary file if fs:temp capability is granted.
func (e *CapabilityEnforcer) DeleteTempFile(name string) error {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return fmt.Errorf("capability fs:temp not granted")
	}

	filePath := filepath.Join(e.tempDir, name)

	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(filePath), e.tempDir) {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete temp file: %w", err)
	}

	return nil
}

// ListTempFiles lists all temporary files if fs:temp capability is granted.
func (e *CapabilityEnforcer) ListTempFiles() ([]string, error) {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return nil, fmt.Errorf("capability fs:temp not granted")
	}

	entries, err := os.ReadDir(e.tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list temp files: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ReadFile reads a file if fs:read capability is granted.
func (e *CapabilityEnforcer) ReadFile(path string) ([]byte, error) {
	if !e.HasCapability(engine.CapabilityFSRead) {
		return nil, fmt.Errorf("capability fs:read not granted")
	}

	// Prevent reading sensitive files
	if e.isSensitiveFile(path) {
		return nil, fmt.Errorf("access to sensitive file denied: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile writes a file if fs:write capability is granted.
func (e *CapabilityEnforcer) WriteFile(path string, data []byte, perm os.FileMode) error {
	if !e.HasCapability(engine.CapabilityFSWrite) {
		return fmt.Errorf("capability fs:write not granted")
	}

	// Prevent writing to sensitive locations
	if e.isSensitivePath(path) {
		return fmt.Errorf("access to sensitive path denied: %s", path)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadEnv reads an environment variable if env:read capability is granted.
func (e *CapabilityEnforcer) ReadEnv(key string) (string, error) {
	if !e.HasCapability(engine.CapabilityEnvRead) {
		return "", fmt.Errorf("capability env:read not granted")
	}

	// Filter sensitive environment variables
	if e.isSensitiveEnvVar(key) {
		return "", fmt.Errorf("access to sensitive environment variable denied: %s", key)
	}

	return os.Getenv(key), nil
}

// DecryptSecret decrypts a secret if secrets:read capability is granted.
func (e *CapabilityEnforcer) DecryptSecret(encrypted string) (string, error) {
	if !e.HasCapability(engine.CapabilitySecretsRead) {
		return "", fmt.Errorf("capability secrets:read not granted")
	}

	if e.secretsDecryptor == nil {
		return "", fmt.Errorf("secrets decryptor not configured")
	}

	decrypted, err := e.secretsDecryptor(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return decrypted, nil
}

// isSensitiveFile checks if a file path is sensitive and should be restricted.
func (e *CapabilityEnforcer) isSensitiveFile(path string) bool {
	sensitivePaths := []string{
		"/etc/shadow",
		"/etc/passwd",
		"/root/.ssh",
		"/.aws/credentials",
		"/.kube/config",
	}

	cleanPath := filepath.Clean(path)
	for _, sensitive := range sensitivePaths {
		if strings.Contains(cleanPath, sensitive) {
			return true
		}
	}

	return false
}

// isSensitivePath checks if a directory path is sensitive and should be restricted.
func (e *CapabilityEnforcer) isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/etc",
		"/root",
		"/sys",
		"/proc",
		"/dev",
	}

	cleanPath := filepath.Clean(path)
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(cleanPath, sensitive) {
			return true
		}
	}

	return false
}

// isSensitiveEnvVar checks if an environment variable is sensitive.
func (e *CapabilityEnforcer) isSensitiveEnvVar(key string) bool {
	sensitiveVars := []string{
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"SSH_PRIVATE_KEY",
		"DATABASE_PASSWORD",
		"API_KEY",
		"SECRET",
		"TOKEN",
		"PASSWORD",
	}

	upperKey := strings.ToUpper(key)
	for _, sensitive := range sensitiveVars {
		if strings.Contains(upperKey, sensitive) {
			return true
		}
	}

	return false
}

// Cleanup cleans up temporary files.
func (e *CapabilityEnforcer) Cleanup() error {
	if !e.HasCapability(engine.CapabilityFSTemp) {
		return nil // Nothing to clean up
	}

	if err := os.RemoveAll(e.tempDir); err != nil {
		return fmt.Errorf("failed to clean up temp directory: %w", err)
	}

	return nil
}
