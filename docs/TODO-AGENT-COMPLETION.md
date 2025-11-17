# Agent Architecture - Missing Components

This document tracks missing components that need to be implemented for a production-ready agent architecture.

## ✅ Completed Components

- [x] **Agent Binary** (`cmd/froyo-agent/`) - Compiles successfully
- [x] **NATS Client** with auto-reconnect and health monitoring
- [x] **Task Executor** with concurrency control
- [x] **Pull Scheduler** for continuous compliance
- [x] **Health Monitoring** and reporting
- [x] **Protocol Definitions** (messages, subjects, serialization)
- [x] **Orchestrator NATS Integration**
- [x] **Hybrid Mode Support** (SSH + Agent)
- [x] **CLI Flags** for NATS connection
- [x] **Configuration System** (YAML loader with validation)
- [x] **Makefile** builds agent for 5 platforms
- [x] **NATS Deployment** stacks (Docker + systemd)
- [x] **Agent Onboarding** stack
- [x] **Comprehensive Documentation**

## 🔴 Critical Missing Components

### 1. Module Distribution Mechanism

**Status:** ⚠️ **CRITICAL - Agent cannot execute tasks without this**

**Location:** `internal/agent/executor/executor.go:216-228`

**Current Code:**
```go
func (e *Executor) getModulePath(module string) (string, error) {
    // For now, assume modules are in a standard location
    // In a real implementation, this would download modules if not cached
    modulePath := filepath.Join(e.config.Execution.ModuleCache, module+".wasm")

    // Check if module exists
    if _, err := os.Stat(modulePath); err != nil {
        // Module doesn't exist - would need to download it
        // For now, return an error
        return "", fmt.Errorf("module not found in cache: %s (would need to implement download)", module)
    }

    return modulePath, nil
}
```

**Problem:**
- Agents have no way to obtain WASM modules
- Modules must exist in cache or tasks fail
- No download mechanism implemented

**Proposed Solutions:**

#### Option A: Module Registry (Recommended for Production)
Create a module registry service that agents can download from:

```go
// ModuleRegistry provides WASM module downloads
type ModuleRegistry struct {
    baseURL string
    cache   string
    client  *http.Client
}

func (r *ModuleRegistry) Get(module string) (string, error) {
    // Check cache first
    cachePath := filepath.Join(r.cache, module+".wasm")
    if _, err := os.Stat(cachePath); err == nil {
        return cachePath, nil
    }

    // Download from registry
    url := fmt.Sprintf("%s/modules/%s.wasm", r.baseURL, module)
    resp, err := r.client.Get(url)
    if err != nil {
        return "", fmt.Errorf("failed to download module: %w", err)
    }
    defer resp.Body.Close()

    // Verify checksum
    if err := r.verifyChecksum(resp.Body, module); err != nil {
        return "", err
    }

    // Save to cache
    if err := r.saveToCache(resp.Body, cachePath); err != nil {
        return "", err
    }

    return cachePath, nil
}
```

**Configuration:**
```yaml
execution:
  module_registry:
    url: "https://modules.openfroyo.io"
    verify_checksums: true
    cache_ttl: 3600
```

**Implementation Steps:**
1. Create module registry HTTP server
2. Add checksum verification
3. Implement caching with TTL
4. Add module registry configuration to agent
5. Update `getModulePath` to download modules

**Estimated effort:** 4-6 hours

#### Option B: NATS-based Module Distribution
Send WASM modules over NATS with task messages:

```go
type TaskMessage struct {
    TaskID    string
    Type      string
    Module    string
    ModuleURL string                 // Download URL
    ModuleChecksum string             // SHA256 checksum
    ModuleData []byte                 // Optional: inline module
    Vars      map[string]interface{}
    // ...
}
```

**Pros:** No separate infrastructure needed
**Cons:** Large messages, potential NATS payload limits

**Implementation Steps:**
1. Add module URL/checksum to TaskMessage
2. Implement HTTP download in agent
3. Verify checksums before execution

**Estimated effort:** 2-3 hours

#### Option C: Pre-populate Modules During Onboarding
Copy all modules to agents during onboarding:

**Update `onboard-agent.ofy`:**
```yaml
- name: Download all WASM modules
  module: generic/get_url
  hosts:
    - web-01
    - web-02
  vars:
    urls:
      - https://github.com/.../generic-file.wasm
      - https://github.com/.../generic-command.wasm
      - https://github.com/.../linux-package.wasm
    dest: /etc/froyo/modules/
```

**Pros:** Simple, no dynamic downloads needed
**Cons:**
- Modules must be updated separately
- Large initial download
- Version management complexity

**Implementation Steps:**
1. Generate module list script
2. Update onboarding stack to download all modules
3. Document module update procedure

**Estimated effort:** 1-2 hours

---

### 2. GitHub Actions CI/CD

**Status:** ❌ **Missing**

**Location:** `.github/workflows/`

**What's Needed:**

#### Workflow 1: Build and Test
```yaml
# .github/workflows/build.yml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build CLI
        run: make build-cli

      - name: Build Agent
        run: make build-agent

      - name: Build Runner
        run: make build-runner

      - name: Run Tests
        run: make test

      - name: Upload Binaries
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: bin/
```

#### Workflow 2: Release
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4

      - name: Build All Platforms
        run: make build

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: bin/*
          generate_release_notes: true
```

**Implementation Steps:**
1. Create `.github/workflows/build.yml`
2. Create `.github/workflows/release.yml`
3. Test workflows on feature branch
4. Document release process

**Estimated effort:** 1-2 hours

---

### 3. Testing Suite

**Status:** ❌ **Missing**

**What's Needed:**

#### Unit Tests
```go
// internal/agent/executor/executor_test.go
package executor_test

func TestExecuteModule(t *testing.T) {
    // Test module execution
}

func TestExecuteCommand(t *testing.T) {
    // Test command execution
}

func TestConcurrencyLimit(t *testing.T) {
    // Test max_concurrent enforcement
}

func TestTimeout(t *testing.T) {
    // Test task timeout handling
}
```

#### Integration Tests
```go
// internal/agent/integration_test.go
func TestAgentOrchestatorCommunication(t *testing.T) {
    // Start test NATS server
    // Start agent
    // Send task via orchestrator
    // Verify result received
}

func TestModuleDownload(t *testing.T) {
    // Test module registry download
}

func TestHealthMonitoring(t *testing.T) {
    // Verify health heartbeats
}
```

**Implementation Steps:**
1. Create test structure
2. Write unit tests for executor
3. Write unit tests for NATS client
4. Create integration test harness
5. Add test coverage reporting

**Estimated effort:** 8-12 hours

---

## 🟡 Important Missing Components

### 4. Agent Auto-Discovery

**Status:** ⚠️ **Nice to have**

**Feature:** Agents register themselves with orchestrator on startup

**Implementation:**
```go
// Agent publishes discovery message on startup
type AgentDiscovery struct {
    AgentID    string
    Datacenter string
    Version    string
    Hostname   string
    IP         string
    Tags       []string
    Timestamp  time.Time
}

// Publish to: openfroyo.{dc}.discovery
```

**Benefits:**
- Automatic inventory management
- Dynamic host detection
- Health status tracking

**Estimated effort:** 3-4 hours

---

### 5. Metrics Collection

**Status:** ⚠️ **Important for production**

**Feature:** Expose Prometheus metrics for monitoring

**Implementation:**
```go
// internal/agent/metrics/collector.go
type Collector struct {
    tasksExecuted   prometheus.Counter
    taskDuration    prometheus.Histogram
    activeT asks     prometheus.Gauge
    natsConnected   prometheus.Gauge
}

// Expose HTTP endpoint for Prometheus
http.Handle("/metrics", promhttp.Handler())
```

**Metrics to track:**
- Tasks executed (by status)
- Task duration histogram
- Active tasks gauge
- NATS connection status
- Module cache hit rate
- Health check status

**Estimated effort:** 4-6 hours

---

### 6. Stack Validation Tooling

**Status:** ⚠️ **Developer productivity**

**Feature:** Validate stack files before execution

```bash
# Validate stack syntax
froyo validate stacks/web-stack.ofy

# Check for:
# - YAML syntax errors
# - Invalid module references
# - Missing required vars
# - Inventory host existence
# - Module availability
```

**Implementation Steps:**
1. Create validation package
2. Add schema validation
3. Check module references
4. Verify inventory hosts
5. Add to CLI as `froyo validate`

**Estimated effort:** 3-4 hours

---

## 🟢 Nice-to-Have Components

### 7. Agent Web UI

**Status:** ⭕ **Future enhancement**

**Feature:** Web-based agent dashboard

**Features:**
- View connected agents
- Monitor task execution
- View logs in real-time
- Trigger ad-hoc tasks
- View metrics/graphs

**Estimated effort:** 20-40 hours

---

### 8. Drift Detection

**Status:** ⭕ **Future enhancement**

**Feature:** Detect configuration drift

```yaml
pull:
  enabled: true
  drift_detection:
    enabled: true
    report_subject: openfroyo.dc1.drift
  stacks:
    - name: security-baseline
      interval: 300
      check_drift: true
```

**Estimated effort:** 6-8 hours

---

### 9. Rollback Mechanism

**Status:** ⭕ **Future enhancement**

**Feature:** Automatic rollback on task failure

```yaml
run:
  - name: Deploy application
    module: generic/copy
    rollback_on_failure: true
    rollback_steps:
      - name: Restore previous version
        module: generic/copy
```

**Estimated effort:** 8-12 hours

---

## 📊 Priority Matrix

### Immediate (P0) - Blocks Production Use
1. **Module Distribution** - Without this, agents cannot execute ANY tasks

### High Priority (P1) - Needed for Production
2. **GitHub Actions CI/CD** - For reliable builds and releases
3. **Testing Suite** - For code quality and regression prevention

### Medium Priority (P2) - Important for Operations
4. **Agent Auto-Discovery** - Reduces manual inventory management
5. **Metrics Collection** - Essential for monitoring
6. **Stack Validation** - Improves developer experience

### Low Priority (P3) - Nice to Have
7. **Agent Web UI** - Improves usability but not essential
8. **Drift Detection** - Valuable for compliance
9. **Rollback Mechanism** - Safety feature

---

## 🎯 Recommended Implementation Order

### Phase 1: Core Functionality (Critical)
**Timeline:** 1-2 days

1. ✅ **Module Distribution** (Option B: NATS-based)
   - Add ModuleURL to TaskMessage
   - Implement HTTP download in agent
   - Add checksum verification

2. ✅ **Basic Testing**
   - Unit tests for executor
   - Integration test for agent-orchestrator
   - Verify end-to-end task execution

### Phase 2: Production Readiness
**Timeline:** 2-3 days

3. ✅ **GitHub Actions**
   - Build workflow
   - Release workflow
   - Test on sample PR

4. ✅ **Comprehensive Testing**
   - Full unit test coverage
   - Integration test suite
   - Load testing (100+ concurrent tasks)

5. ✅ **Metrics Collection**
   - Prometheus metrics
   - Grafana dashboard examples
   - Alerting rules

### Phase 3: Operational Excellence
**Timeline:** 3-5 days

6. ✅ **Agent Auto-Discovery**
7. ✅ **Stack Validation**
8. ✅ **Documentation Updates**

### Phase 4: Future Enhancements
**Timeline:** TBD

9. ⭕ **Agent Web UI**
10. ⭕ **Drift Detection**
11. ⭕ **Rollback Mechanism**

---

## 🚀 Quick Win: Minimum Viable Agent

To get agents working with minimal effort:

### Option 1: Pre-populate Modules (1-2 hours)
Update agent onboarding to copy all modules during installation:

```yaml
- name: Download all WASM modules
  module: generic/get_url
  loop:
    - generic/file
    - generic/command
    - generic/copy
    - linux/package
  vars:
    url: "https://github.com/piwi3910/openfroyo/releases/download/v1.0.0/{{ item }}.wasm"
    dest: "/etc/froyo/modules/{{ item }}.wasm"
```

### Option 2: Module URL in Task (2-3 hours)
Orchestrator sends module download URL with each task:

```go
task := &protocol.TaskMessage{
    TaskID:         taskID,
    Module:         "generic/file",
    ModuleURL:      "https://github.com/.../generic-file.wasm",
    ModuleChecksum: "sha256:abc123...",
    // ...
}
```

Agent downloads module on first use and caches it.

---

## 📝 Next Actions

1. **Choose module distribution approach** (Option B recommended for quick win)
2. **Implement module downloading** in agent executor
3. **Add basic tests** to verify functionality
4. **Create GitHub Actions** for CI/CD
5. **Test end-to-end** with real tasks
6. **Document module management** for users

---

## 📚 References

- [Agent Architecture](architecture/agent-architecture.md)
- [Pull Mode Quickstart](QUICKSTART-PULL-MODE.md)
- [Module Development](../modules/README.md)
- [NATS Documentation](https://docs.nats.io/)
