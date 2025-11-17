package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client is a module registry client
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
	cacheDir   string
	apiKey     string
}

// ModuleMetadata represents module metadata from registry
type ModuleMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Checksum    string   `json:"checksum"`
	Size        int64    `json:"size"`
	Tags        []string `json:"tags,omitempty"`
}

// Config contains registry client configuration
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	RetryCount int
	CacheDir   string
}

// NewClient creates a new registry client
func NewClient(cfg Config, logger *log.Logger) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("registry base URL is required")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}

	// Create cache directory
	if cfg.CacheDir != "" {
		if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	return &Client{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger:   logger,
		cacheDir: cfg.CacheDir,
		apiKey:   cfg.APIKey,
	}, nil
}

// GetModule downloads a module from the registry (with caching)
func (c *Client) GetModule(moduleName string) (string, error) {
	// Check cache first
	if c.cacheDir != "" {
		cachePath := c.getCachePath(moduleName)
		if c.isValidCached(moduleName, cachePath) {
			c.logger.Printf("[DEBUG] Module cache hit: %s", moduleName)
			return cachePath, nil
		}
	}

	// Get metadata
	metadata, err := c.GetMetadata(moduleName)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata: %w", err)
	}

	// Download module
	cachePath := c.getCachePath(moduleName)
	if err := c.downloadModule(moduleName, cachePath, metadata.Checksum); err != nil {
		return "", fmt.Errorf("failed to download module: %w", err)
	}

	return cachePath, nil
}

// GetMetadata retrieves module metadata from the registry
func (c *Client) GetMetadata(moduleName string) (*ModuleMetadata, error) {
	url := fmt.Sprintf("%s/modules/%s/metadata", c.baseURL, moduleName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var metadata ModuleMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

// downloadModule downloads a module and verifies its checksum
func (c *Client) downloadModule(moduleName, destPath, expectedChecksum string) error {
	c.logger.Printf("[INFO] Downloading module: %s from %s", moduleName, c.baseURL)

	// Create temp file
	tmpFile, err := os.CreateTemp("", "module-*.wasm")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	// Download with retries
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			c.logger.Printf("[INFO] Retry attempt %d for %s", attempt, moduleName)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		err = c.downloadWithTimeout(moduleName, tmpFile)
		if err == nil {
			break
		}
		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("download failed after retries: %w", lastErr)
	}

	tmpFile.Close()

	// Verify checksum
	if expectedChecksum != "" {
		actualChecksum, err := computeChecksum(tmpPath)
		if err != nil {
			return fmt.Errorf("failed to compute checksum: %w", err)
		}

		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s",
				expectedChecksum, actualChecksum)
		}

		c.logger.Printf("[INFO] Checksum verified: %s", moduleName)
	}

	// Move to cache
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move to cache: %w", err)
	}

	c.logger.Printf("[INFO] Module cached: %s", destPath)
	return nil
}

// downloadWithTimeout downloads a module with timeout
func (c *Client) downloadWithTimeout(moduleName string, dest io.Writer) error {
	url := fmt.Sprintf("%s/modules/%s/download", c.baseURL, moduleName)

	ctx, cancel := context.WithTimeout(context.Background(), c.httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	written, err := io.Copy(dest, resp.Body)
	if err != nil {
		return err
	}

	c.logger.Printf("[DEBUG] Downloaded %d bytes", written)
	return nil
}

// ListModules lists all available modules
func (c *Client) ListModules() ([]*ModuleMetadata, error) {
	url := fmt.Sprintf("%s/modules", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var result struct {
		Modules []*ModuleMetadata `json:"modules"`
		Count   int               `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Modules, nil
}

// Health checks if the registry is reachable
func (c *Client) Health() error {
	url := fmt.Sprintf("%s/healthz", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// getCachePath returns the cache path for a module
func (c *Client) getCachePath(moduleName string) string {
	filename := strings.ReplaceAll(moduleName, "/", "-") + ".wasm"
	return filepath.Join(c.cacheDir, filename)
}

// isValidCached checks if a cached module is valid
func (c *Client) isValidCached(moduleName, cachePath string) bool {
	// Check if file exists
	if _, err := os.Stat(cachePath); err != nil {
		return false
	}

	// Get metadata to verify checksum
	metadata, err := c.GetMetadata(moduleName)
	if err != nil {
		c.logger.Printf("[WARN] Failed to get metadata for cached module %s: %v", moduleName, err)
		return false
	}

	// Verify checksum
	if metadata.Checksum != "" {
		actualChecksum, err := computeChecksum(cachePath)
		if err != nil {
			c.logger.Printf("[WARN] Failed to compute checksum for %s: %v", cachePath, err)
			return false
		}

		if actualChecksum != metadata.Checksum {
			c.logger.Printf("[WARN] Checksum mismatch for cached %s", moduleName)
			return false
		}
	}

	return true
}

// computeChecksum computes SHA256 checksum of a file
func computeChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
