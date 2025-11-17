package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Storage handles module storage and metadata
type Storage struct {
	config    StorageConfig
	logger    *log.Logger
	metadata  map[string]*ModuleMetadata
	checksums map[string]string
	mutex     sync.RWMutex
}

// ModuleMetadata contains module information
type ModuleMetadata struct {
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Description string   `json:"description" yaml:"description"`
	Checksum    string   `json:"checksum" yaml:"checksum"`
	Size        int64    `json:"size" yaml:"size"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Author      string   `json:"author,omitempty" yaml:"author,omitempty"`
	License     string   `json:"license,omitempty" yaml:"license,omitempty"`
}

// NewStorage creates a new storage instance
func NewStorage(cfg StorageConfig, logger *log.Logger) (*Storage, error) {
	s := &Storage{
		config:    cfg,
		logger:    logger,
		metadata:  make(map[string]*ModuleMetadata),
		checksums: make(map[string]string),
	}

	// Create modules directory if it doesn't exist
	if err := os.MkdirAll(cfg.ModulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create modules directory: %w", err)
	}

	// Load metadata
	if err := s.loadMetadata(); err != nil {
		logger.Printf("[WARN] Failed to load metadata: %v", err)
	}

	// Load checksums
	if err := s.loadChecksums(); err != nil {
		logger.Printf("[WARN] Failed to load checksums: %v", err)
	}

	// Scan modules directory
	if err := s.scanModules(); err != nil {
		return nil, fmt.Errorf("failed to scan modules: %w", err)
	}

	return s, nil
}

// loadMetadata loads module metadata from file
func (s *Storage) loadMetadata() error {
	if s.config.MetadataFile == "" {
		return nil
	}

	data, err := os.ReadFile(s.config.MetadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}

	var metadataList []ModuleMetadata
	if err := yaml.Unmarshal(data, &metadataList); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range metadataList {
		meta := &metadataList[i]
		s.metadata[meta.Name] = meta
	}

	s.logger.Printf("[INFO] Loaded metadata for %d modules", len(metadataList))
	return nil
}

// loadChecksums loads checksums from file
func (s *Storage) loadChecksums() error {
	if s.config.ChecksumsFile == "" {
		return nil
	}

	data, err := os.ReadFile(s.config.ChecksumsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return err
	}

	var checksums map[string]string
	if err := yaml.Unmarshal(data, &checksums); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.checksums = checksums

	s.logger.Printf("[INFO] Loaded checksums for %d modules", len(checksums))
	return nil
}

// scanModules scans the modules directory and computes checksums
func (s *Storage) scanModules() error {
	entries, err := os.ReadDir(s.config.ModulesDir)
	if err != nil {
		return err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".wasm") {
			continue
		}

		moduleName := strings.TrimSuffix(entry.Name(), ".wasm")
		modulePath := filepath.Join(s.config.ModulesDir, entry.Name())

		// Get file info
		info, err := entry.Info()
		if err != nil {
			s.logger.Printf("[WARN] Failed to stat %s: %v", entry.Name(), err)
			continue
		}

		// Compute checksum if not already loaded
		s.mutex.RLock()
		_, hasChecksum := s.checksums[moduleName]
		s.mutex.RUnlock()

		if !hasChecksum {
			checksum, err := s.computeChecksum(modulePath)
			if err != nil {
				s.logger.Printf("[WARN] Failed to compute checksum for %s: %v", moduleName, err)
			} else {
				s.mutex.Lock()
				s.checksums[moduleName] = checksum
				s.mutex.Unlock()
			}
		}

		// Create metadata if not already loaded
		s.mutex.RLock()
		_, hasMeta := s.metadata[moduleName]
		s.mutex.RUnlock()

		if !hasMeta {
			s.mutex.Lock()
			s.metadata[moduleName] = &ModuleMetadata{
				Name:     moduleName,
				Version:  "unknown",
				Size:     info.Size(),
				Checksum: s.checksums[moduleName],
			}
			s.mutex.Unlock()
		}

		count++
	}

	s.logger.Printf("[INFO] Scanned %d modules in %s", count, s.config.ModulesDir)
	return nil
}

// computeChecksum computes SHA256 checksum of a file
func (s *Storage) computeChecksum(path string) (string, error) {
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

// ListModules returns a list of all available modules
func (s *Storage) ListModules() ([]*ModuleMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	modules := make([]*ModuleMetadata, 0, len(s.metadata))
	for _, meta := range s.metadata {
		modules = append(modules, meta)
	}

	return modules, nil
}

// GetMetadata returns metadata for a specific module
func (s *Storage) GetMetadata(name string) (*ModuleMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	meta, exists := s.metadata[name]
	if !exists {
		return nil, fmt.Errorf("module not found: %s", name)
	}

	return meta, nil
}

// GetModulePath returns the filesystem path to a module
func (s *Storage) GetModulePath(name string) (string, error) {
	// Convert module name (e.g., "generic/file") to filename (e.g., "generic-file.wasm")
	filename := strings.ReplaceAll(name, "/", "-") + ".wasm"
	path := filepath.Join(s.config.ModulesDir, filename)

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("module not found: %s", name)
	}

	return path, nil
}

// GetChecksum returns the checksum for a module
func (s *Storage) GetChecksum(name string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	checksum, exists := s.checksums[name]
	if !exists {
		return "", fmt.Errorf("checksum not found for module: %s", name)
	}

	return checksum, nil
}

// SaveMetadata saves all metadata to file
func (s *Storage) SaveMetadata() error {
	if s.config.MetadataFile == "" {
		return nil
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	metadataList := make([]ModuleMetadata, 0, len(s.metadata))
	for _, meta := range s.metadata {
		metadataList = append(metadataList, *meta)
	}

	data, err := yaml.Marshal(metadataList)
	if err != nil {
		return err
	}

	return os.WriteFile(s.config.MetadataFile, data, 0644)
}

// SaveChecksums saves all checksums to file
func (s *Storage) SaveChecksums() error {
	if s.config.ChecksumsFile == "" {
		return nil
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := yaml.Marshal(s.checksums)
	if err != nil {
		return err
	}

	return os.WriteFile(s.config.ChecksumsFile, data, 0644)
}
