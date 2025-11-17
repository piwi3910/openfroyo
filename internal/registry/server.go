package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Server is the module registry HTTP server
type Server struct {
	config   *Config
	logger   *log.Logger
	storage  *Storage
	server   *http.Server
	cache    *Cache
	metrics  *Metrics
}

// NewServer creates a new registry server
func NewServer(cfg *Config, logger *log.Logger) (*Server, error) {
	// Create storage
	storage, err := NewStorage(cfg.Storage, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create cache
	cache := NewCache(cfg.Storage.CacheTTL)

	// Create metrics
	metrics := NewMetrics()

	srv := &Server{
		config:  cfg,
		logger:  logger,
		storage: storage,
		cache:   cache,
		metrics: metrics,
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Module endpoints
	mux.HandleFunc("/modules", srv.handleModuleList)
	mux.HandleFunc("/modules/", srv.handleModuleRequest)

	// Health and metrics
	if cfg.Server.EnableHealthz {
		mux.HandleFunc("/healthz", srv.handleHealth)
		mux.HandleFunc("/ready", srv.handleReady)
	}
	if cfg.Server.EnableMetrics {
		mux.HandleFunc("/metrics", srv.handleMetrics)
	}

	// Wrap with middleware
	var handler http.Handler = mux
	if cfg.Logging.AccessLog {
		handler = srv.loggingMiddleware(handler)
	}
	if cfg.Auth.Enabled {
		handler = srv.authMiddleware(handler)
	}
	if cfg.Server.EnableCORS {
		handler = srv.corsMiddleware(handler)
	}

	srv.server = &http.Server{
		Addr:           cfg.Server.Listen,
		Handler:        handler,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return srv, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Printf("Server starting on %s", s.config.Server.Listen)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// handleModuleList lists all available modules
func (s *Server) handleModuleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modules, err := s.storage.ListModules()
	if err != nil {
		s.logger.Printf("[ERROR] Failed to list modules: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.metrics.IncModuleListRequests()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"modules": modules,
		"count":   len(modules),
	})
}

// handleModuleRequest handles module-specific requests
func (s *Server) handleModuleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract module name from path: /modules/{name}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/modules/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	moduleName := parts[0]
	action := "download" // default action
	if len(parts) == 2 {
		action = parts[1]
	}

	switch action {
	case "metadata":
		s.handleModuleMetadata(w, r, moduleName)
	case "download":
		s.handleModuleDownload(w, r, moduleName)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// handleModuleMetadata returns module metadata
func (s *Server) handleModuleMetadata(w http.ResponseWriter, r *http.Request, moduleName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metadata, err := s.storage.GetMetadata(moduleName)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to get metadata for %s: %v", moduleName, err)
		http.Error(w, "Module not found", http.StatusNotFound)
		return
	}

	s.metrics.IncModuleMetadataRequests()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handleModuleDownload serves the WASM module binary
func (s *Server) handleModuleDownload(w http.ResponseWriter, r *http.Request, moduleName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get module path
	modulePath, err := s.storage.GetModulePath(moduleName)
	if err != nil {
		s.logger.Printf("[ERROR] Module not found: %s: %v", moduleName, err)
		http.Error(w, "Module not found", http.StatusNotFound)
		return
	}

	// Get metadata for checksum
	metadata, err := s.storage.GetMetadata(moduleName)
	if err != nil {
		s.logger.Printf("[WARN] No metadata for %s: %v", moduleName, err)
	}

	// Open file
	file, err := os.Open(modulePath)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to open module %s: %v", moduleName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "application/wasm")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.wasm", moduleName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	if metadata != nil && metadata.Checksum != "" {
		w.Header().Set("X-Module-Checksum", metadata.Checksum)
		w.Header().Set("X-Module-Version", metadata.Version)
	}

	// Stream file
	written, err := io.Copy(w, file)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to stream module %s: %v", moduleName, err)
		return
	}

	s.metrics.IncModuleDownloads()
	s.metrics.AddBytesServed(written)

	s.logger.Printf("[INFO] Served module: %s (%d bytes)", moduleName, written)
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// handleReady returns readiness status
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if modules directory is accessible
	if _, err := os.Stat(s.config.Storage.ModulesDir); err != nil {
		http.Error(w, "Not ready", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready",
		"time":   time.Now().Unix(),
	})
}

// handleMetrics returns Prometheus-style metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := s.metrics.GetStats()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP froyo_registry_module_downloads_total Total module downloads\n")
	fmt.Fprintf(w, "# TYPE froyo_registry_module_downloads_total counter\n")
	fmt.Fprintf(w, "froyo_registry_module_downloads_total %d\n", stats.ModuleDownloads)

	fmt.Fprintf(w, "# HELP froyo_registry_bytes_served_total Total bytes served\n")
	fmt.Fprintf(w, "# TYPE froyo_registry_bytes_served_total counter\n")
	fmt.Fprintf(w, "froyo_registry_bytes_served_total %d\n", stats.BytesServed)

	fmt.Fprintf(w, "# HELP froyo_registry_module_list_requests_total Total module list requests\n")
	fmt.Fprintf(w, "# TYPE froyo_registry_module_list_requests_total counter\n")
	fmt.Fprintf(w, "froyo_registry_module_list_requests_total %d\n", stats.ModuleListRequests)

	fmt.Fprintf(w, "# HELP froyo_registry_module_metadata_requests_total Total metadata requests\n")
	fmt.Fprintf(w, "# TYPE froyo_registry_module_metadata_requests_total counter\n")
	fmt.Fprintf(w, "froyo_registry_module_metadata_requests_total %d\n", stats.ModuleMetadataRequests)
}

// Middleware functions

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		s.logger.Printf("[ACCESS] %s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health/metrics endpoints
		if r.URL.Path == "/healthz" || r.URL.Path == "/ready" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		if s.config.Auth.Type == "api_key" {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			valid := false
			for _, key := range s.config.Auth.APIKeys {
				if apiKey == key {
					valid = true
					break
				}
			}

			if !valid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Cache provides in-memory caching
type Cache struct {
	data  map[string]*cacheEntry
	mutex sync.RWMutex
	ttl   int
}

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

func NewCache(ttl int) *Cache {
	return &Cache{
		data: make(map[string]*cacheEntry),
		ttl:  ttl,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(time.Duration(c.ttl) * time.Second),
	}
}

// Metrics tracks registry metrics
type Metrics struct {
	moduleDownloads        int64
	bytesServed            int64
	moduleListRequests     int64
	moduleMetadataRequests int64
	mutex                  sync.RWMutex
}

type MetricsStats struct {
	ModuleDownloads        int64
	BytesServed            int64
	ModuleListRequests     int64
	ModuleMetadataRequests int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncModuleDownloads() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.moduleDownloads++
}

func (m *Metrics) AddBytesServed(bytes int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.bytesServed += bytes
}

func (m *Metrics) IncModuleListRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.moduleListRequests++
}

func (m *Metrics) IncModuleMetadataRequests() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.moduleMetadataRequests++
}

func (m *Metrics) GetStats() MetricsStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return MetricsStats{
		ModuleDownloads:        m.moduleDownloads,
		BytesServed:            m.bytesServed,
		ModuleListRequests:     m.moduleListRequests,
		ModuleMetadataRequests: m.moduleMetadataRequests,
	}
}
