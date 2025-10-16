package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

// Loader handles loading policies from various sources.
type Loader struct {
	logger  zerolog.Logger
	cache   map[string]*Policy
	mu      sync.RWMutex
	watcher *fsnotify.Watcher
}

// NewLoader creates a new policy loader.
func NewLoader(logger zerolog.Logger) *Loader {
	return &Loader{
		logger: logger.With().Str("component", "policy-loader").Logger(),
		cache:  make(map[string]*Policy),
	}
}

// LoadFromPaths loads policies from a list of file or directory paths.
func (l *Loader) LoadFromPaths(ctx context.Context, paths []string) ([]Policy, error) {
	var allPolicies []Policy

	for _, path := range paths {
		policies, err := l.loadFromPath(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to load from path %s: %w", path, err)
		}
		allPolicies = append(allPolicies, policies...)
	}

	l.logger.Info().
		Int("total", len(allPolicies)).
		Int("sources", len(paths)).
		Msg("Policies loaded from paths")

	return allPolicies, nil
}

// loadFromPath loads policies from a single path (file or directory).
func (l *Loader) loadFromPath(ctx context.Context, path string) ([]Policy, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return l.loadFromDirectory(ctx, path)
	}

	policy, err := l.loadFromFile(ctx, path)
	if err != nil {
		return nil, err
	}

	return []Policy{*policy}, nil
}

// loadFromDirectory loads all .rego files from a directory recursively.
func (l *Loader) loadFromDirectory(ctx context.Context, dirPath string) ([]Policy, error) {
	var policies []Policy

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.rego files
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".rego") && !strings.HasSuffix(path, ".json") {
			return nil
		}

		policy, err := l.loadFromFile(ctx, path)
		if err != nil {
			l.logger.Warn().Err(err).Str("path", path).Msg("Failed to load policy file")
			return nil // Continue processing other files
		}

		policies = append(policies, *policy)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return policies, nil
}

// loadFromFile loads a policy from a single file.
func (l *Loader) loadFromFile(ctx context.Context, filePath string) (*Policy, error) {
	// Check cache first
	l.mu.RLock()
	if cached, exists := l.cache[filePath]; exists {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var policy *Policy

	// Determine file type and parse accordingly
	switch {
	case strings.HasSuffix(filePath, ".rego"):
		policy = l.parseRegoFile(filePath, data)
	case strings.HasSuffix(filePath, ".json"):
		policy, err = l.parseJSONFile(data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	// Cache the policy
	l.mu.Lock()
	l.cache[filePath] = policy
	l.mu.Unlock()

	l.logger.Debug().
		Str("path", filePath).
		Str("policy", policy.Name).
		Msg("Policy loaded from file")

	return policy, nil
}

// parseRegoFile parses a .rego file into a Policy.
func (l *Loader) parseRegoFile(filePath string, data []byte) *Policy {
	// Extract policy name from file path
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, ".rego")

	// Try to extract metadata from comments
	description := l.extractDescription(string(data))

	return &Policy{
		Name:        name,
		Description: description,
		Rego:        string(data),
		Severity:    SeverityWarning, // Default severity
		Enabled:     true,
		Tags:        []string{},
		Metadata: map[string]interface{}{
			"source": filePath,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// parseJSONFile parses a JSON policy definition.
func (l *Loader) parseJSONFile(data []byte) (*Policy, error) {
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse JSON policy: %w", err)
	}

	// Set defaults if not specified
	if policy.Severity == "" {
		policy.Severity = SeverityWarning
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	if policy.UpdatedAt.IsZero() {
		policy.UpdatedAt = time.Now()
	}

	return &policy, nil
}

// extractDescription extracts description from Rego comments.
func (l *Loader) extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var description strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if comment != "" && !strings.HasPrefix(comment, "package") {
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(comment)
			}
		} else if trimmed != "" && description.Len() > 0 {
			// Stop at first non-comment, non-empty line
			break
		}
	}

	return description.String()
}

// LoadBundle loads a policy bundle.
func (l *Loader) LoadBundle(ctx context.Context, bundlePath string) (*PolicyBundle, error) {
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle: %w", err)
	}

	var bundle PolicyBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}

	l.logger.Info().
		Str("bundle", bundle.Name).
		Str("version", bundle.Version).
		Int("policies", len(bundle.Policies)).
		Msg("Policy bundle loaded")

	return &bundle, nil
}

// Watch starts watching paths for policy changes and triggers reload on change.
func (l *Loader) Watch(ctx context.Context, paths []string, reloadFn func([]Policy) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	l.watcher = watcher

	// Add paths to watcher
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			l.logger.Warn().Err(err).Str("path", path).Msg("Failed to stat path for watching")
			continue
		}

		if info.IsDir() {
			// Watch directory recursively
			if err := l.watchDirectory(path); err != nil {
				l.logger.Warn().Err(err).Str("path", path).Msg("Failed to watch directory")
			}
		} else {
			if err := watcher.Add(path); err != nil {
				l.logger.Warn().Err(err).Str("path", path).Msg("Failed to watch file")
			}
		}
	}

	// Start watching in background
	go l.processEvents(ctx, paths, reloadFn)

	l.logger.Info().
		Int("paths", len(paths)).
		Msg("Started watching policy paths")

	return nil
}

// watchDirectory adds all files in a directory to the watcher.
func (l *Loader) watchDirectory(dirPath string) error {
	return filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return l.watcher.Add(path)
		}

		return nil
	})
}

// processEvents processes file system events and triggers reloads.
func (l *Loader) processEvents(ctx context.Context, paths []string, reloadFn func([]Policy) error) {
	// Debounce reload events
	var reloadTimer *time.Timer
	reloadDelay := 500 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			if l.watcher != nil {
				_ = l.watcher.Close()
			}
			return

		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}

			// Only reload on write or create events for .rego or .json files
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if strings.HasSuffix(event.Name, ".rego") || strings.HasSuffix(event.Name, ".json") {
					l.logger.Debug().
						Str("file", event.Name).
						Str("op", event.Op.String()).
						Msg("Policy file changed")

					// Clear cache for this file
					l.mu.Lock()
					delete(l.cache, event.Name)
					l.mu.Unlock()

					// Debounce reload
					if reloadTimer != nil {
						reloadTimer.Stop()
					}
					reloadTimer = time.AfterFunc(reloadDelay, func() {
						if err := l.triggerReload(ctx, paths, reloadFn); err != nil {
							l.logger.Error().Err(err).Msg("Failed to reload policies")
						}
					})
				}
			}

		case err, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
			l.logger.Error().Err(err).Msg("Watcher error")
		}
	}
}

// triggerReload reloads all policies from watched paths.
func (l *Loader) triggerReload(ctx context.Context, paths []string, reloadFn func([]Policy) error) error {
	l.logger.Info().Msg("Reloading policies...")

	policies, err := l.LoadFromPaths(ctx, paths)
	if err != nil {
		return fmt.Errorf("failed to reload policies: %w", err)
	}

	if err := reloadFn(policies); err != nil {
		return fmt.Errorf("failed to apply reloaded policies: %w", err)
	}

	l.logger.Info().
		Int("count", len(policies)).
		Msg("Policies reloaded successfully")

	return nil
}

// StopWatching stops watching for file changes.
func (l *Loader) StopWatching() error {
	if l.watcher != nil {
		return l.watcher.Close()
	}
	return nil
}

// ClearCache clears the policy cache.
func (l *Loader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cache = make(map[string]*Policy)
	l.logger.Debug().Msg("Policy cache cleared")
}
