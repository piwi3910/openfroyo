# Module Distribution: NATS-based vs Module Registry

Detailed comparison of two approaches for distributing WASM modules to agents.

## Architecture Comparison

### Option A: NATS-based Distribution

```
┌──────────────┐
│ Orchestrator │
│   (froyo)    │
└──────┬───────┘
       │
       │ 1. TaskMessage{
       │      Module: "generic/file"
       │      ModuleURL: "https://github.com/.../generic-file.wasm"
       │      ModuleChecksum: "sha256:abc123..."
       │    }
       │
   ┌───▼────────┐
   │    NATS    │
   └───┬────────┘
       │
       │ 2. Agent receives task
       │
   ┌───▼──────────┐
   │    Agent     │
   │              │
   │ 3. Check cache: /etc/froyo/modules/generic-file.wasm
   │ 4. If not cached:
   │    - Download from ModuleURL (HTTP GET)
   │    - Verify checksum
   │    - Save to cache
   │ 5. Execute module
   │
   └──────────────┘
```

**Flow:**
1. Orchestrator includes module URL + checksum in task message
2. NATS delivers task to agent
3. Agent checks local cache first
4. If not cached, agent downloads via HTTP
5. Agent verifies checksum
6. Agent caches module for future use
7. Agent executes module

---

### Option C: Module Registry

```
┌──────────────┐
│ Orchestrator │
│   (froyo)    │
└──────┬───────┘
       │
       │ 1. TaskMessage{
       │      Module: "generic/file"
       │    }
       │
   ┌───▼────────┐
   │    NATS    │
   └───┬────────┘
       │
       │ 2. Agent receives task
       │
   ┌───▼──────────┐                  ┌─────────────────┐
   │    Agent     │                  │ Module Registry │
   │              │                  │   HTTP Server   │
   │ 3. Check cache                  │                 │
   │ 4. If not cached:               │  - Serves WASM  │
   │    GET /modules/generic/file    │  - Checksums    │
   │    ├──────────────────────────►│  - Metadata     │
   │    │                            │  - Versions     │
   │    │  200 OK + WASM binary      │                 │
   │    │◄───────────────────────────┤                 │
   │ 5. Verify checksum              └─────────────────┘
   │ 6. Cache locally
   │ 7. Execute module
   │
   └──────────────┘
```

**Flow:**
1. Orchestrator sends task with module name only
2. NATS delivers task to agent
3. Agent checks local cache first
4. If not cached, agent queries module registry
5. Registry returns WASM binary + metadata
6. Agent verifies checksum
7. Agent caches module
8. Agent executes module

---

## Detailed Comparison

### 1. Implementation Complexity

#### Option A: NATS-based ✅ Simple

**Orchestrator Changes:**
```go
// internal/orchestrator/executor/agent.go
func (e *AgentExecutor) ExecuteTask(...) (*protocol.TaskResult, error) {
    task := &protocol.TaskMessage{
        TaskID:         taskID,
        Module:         module,
        ModuleURL:      e.getModuleURL(module),      // NEW: Add URL
        ModuleChecksum: e.getModuleChecksum(module), // NEW: Add checksum
        Vars:           vars,
        // ...
    }
    // ... rest unchanged
}

// Simple helper functions
func (e *AgentExecutor) getModuleURL(module string) string {
    return fmt.Sprintf("https://github.com/piwi3910/openfroyo/releases/download/v%s/%s.wasm",
        e.version, module)
}

func (e *AgentExecutor) getModuleChecksum(module string) string {
    // Read from checksums file or embed in binary
    return checksums[module]
}
```

**Agent Changes:**
```go
// internal/agent/executor/executor.go
func (e *Executor) getModulePath(module string, moduleURL, checksum string) (string, error) {
    cachePath := filepath.Join(e.config.Execution.ModuleCache, module+".wasm")

    // Check cache
    if _, err := os.Stat(cachePath); err == nil {
        // Verify cached module checksum
        if e.verifyCachedModule(cachePath, checksum) {
            return cachePath, nil
        }
    }

    // Download from URL
    if err := e.downloadModule(moduleURL, cachePath, checksum); err != nil {
        return "", err
    }

    return cachePath, nil
}

func (e *Executor) downloadModule(url, dest, expectedChecksum string) error {
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("download failed: %w", err)
    }
    defer resp.Body.Close()

    // Download to temp file
    tmpFile, err := os.CreateTemp("", "module-*.wasm")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name())

    hash := sha256.New()
    writer := io.MultiWriter(tmpFile, hash)

    if _, err := io.Copy(writer, resp.Body); err != nil {
        return err
    }
    tmpFile.Close()

    // Verify checksum
    actualChecksum := hex.EncodeToString(hash.Sum(nil))
    if actualChecksum != expectedChecksum {
        return fmt.Errorf("checksum mismatch: expected %s, got %s",
            expectedChecksum, actualChecksum)
    }

    // Move to cache
    return os.Rename(tmpFile.Name(), dest)
}
```

**Total Lines:** ~80 lines of code

**Dependencies:** Standard library only (net/http, crypto/sha256)

**Configuration:** None required

---

#### Option C: Module Registry ⚠️ Complex

**New Registry Server:**
```go
// cmd/froyo-registry/main.go (NEW)
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

type ModuleRegistry struct {
    modulesDir string
    checksums  map[string]string
    metadata   map[string]ModuleMetadata
}

type ModuleMetadata struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Checksum    string   `json:"checksum"`
    Size        int64    `json:"size"`
    Tags        []string `json:"tags"`
}

func (r *ModuleRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Parse module name from path
    // Validate request
    // Serve WASM binary
    // Include checksum in response header
}

func (r *ModuleRegistry) handleModuleList(w http.ResponseWriter, req *http.Request) {
    // List all available modules
}

func (r *ModuleRegistry) handleModuleMetadata(w http.ResponseWriter, req *http.Request) {
    // Return module metadata
}

func (r *ModuleRegistry) handleModuleDownload(w http.ResponseWriter, req *http.Request) {
    // Stream WASM binary
}

func main() {
    registry := &ModuleRegistry{
        modulesDir: "/var/lib/froyo-registry/modules",
        checksums:  loadChecksums(),
        metadata:   loadMetadata(),
    }

    http.Handle("/modules/", registry)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Agent Client:**
```go
// internal/agent/registry/client.go (NEW)
package registry

type Client struct {
    baseURL string
    cache   string
    client  *http.Client
}

func (c *Client) Get(module string) (string, error) {
    // Check cache first
    cachePath := filepath.Join(c.cache, module+".wasm")
    if _, err := os.Stat(cachePath); err == nil {
        return cachePath, nil
    }

    // Get metadata
    metadata, err := c.getMetadata(module)
    if err != nil {
        return "", err
    }

    // Download module
    url := fmt.Sprintf("%s/modules/%s/download", c.baseURL, module)
    if err := c.download(url, cachePath, metadata.Checksum); err != nil {
        return "", err
    }

    return cachePath, nil
}

func (c *Client) getMetadata(module string) (*ModuleMetadata, error) {
    url := fmt.Sprintf("%s/modules/%s/metadata", c.baseURL, module)
    resp, err := c.client.Get(url)
    // Parse metadata JSON
    // Return metadata
}

func (c *Client) download(url, dest, checksum string) error {
    // Download with progress
    // Verify checksum
    // Save to cache
}

func (c *Client) ListModules() ([]ModuleMetadata, error) {
    // Get list of all modules
}
```

**Deployment Stack:**
```yaml
# examples/registry-deployment/deploy-registry.ofy
name: Deploy Module Registry
run:
  - name: Install registry binary
    module: generic/get_url
    vars:
      url: "https://github.com/.../froyo-registry-linux-amd64"
      dest: /usr/local/bin/froyo-registry

  - name: Create systemd service
    module: generic/copy
    # ... systemd configuration
```

**Total Lines:** ~400+ lines of code

**Dependencies:**
- HTTP server framework
- Database (optional, for metadata)
- Authentication library

**Configuration:**
```yaml
# Agent config
execution:
  module_registry:
    url: "https://registry.openfroyo.io:8080"
    timeout: 30s
    retry_count: 3
    verify_tls: true
```

---

### 2. Infrastructure Requirements

#### Option A: NATS-based ✅ Minimal

**What you need:**
- ✅ NATS server (already deployed)
- ✅ HTTP server to host modules (GitHub Releases, S3, nginx, etc.)
- ✅ No additional services

**Example: Use GitHub Releases**
```bash
# Build and release modules
make build-wasm

# Upload to GitHub release
gh release create v1.0.0 \
  bin/modules/generic-file.wasm \
  bin/modules/generic-command.wasm \
  bin/modules/linux-package.wasm \
  --generate-notes

# Modules available at:
# https://github.com/piwi3910/openfroyo/releases/download/v1.0.0/generic-file.wasm
```

**Example: Use S3/MinIO**
```bash
# Upload modules
aws s3 sync bin/modules/ s3://openfroyo-modules/v1.0.0/

# Or use nginx
nginx -c /etc/nginx/nginx.conf
# Serve from /var/www/modules/
```

**Cost:** $0 (using GitHub Releases) or ~$1/month (S3)

---

#### Option C: Module Registry ⚠️ Requires New Service

**What you need:**
- ✅ NATS server (already deployed)
- ❌ Module registry server (NEW)
- ❌ Load balancer (for HA)
- ❌ Database (optional, for metadata)
- ❌ TLS certificates
- ❌ Monitoring/alerting

**Deployment:**
```
┌─────────────────────────────────────┐
│        Load Balancer (HAProxy)      │
│         registry.openfroyo.io       │
└─────────────┬───────────────────────┘
              │
        ┌─────┴─────┐
        │           │
   ┌────▼────┐ ┌────▼────┐
   │Registry │ │Registry │
   │ Node 1  │ │ Node 2  │
   └────┬────┘ └────┬────┘
        │           │
        └─────┬─────┘
              │
      ┌───────▼───────┐
      │   Shared      │
      │   Storage     │
      │   (NFS/S3)    │
      └───────────────┘
```

**Cost:** ~$20-50/month (2 VMs + load balancer)

---

### 3. Scalability

#### Option A: NATS-based ✅ Highly Scalable

**Module serving:**
- CDN-backed (GitHub Releases uses CDN)
- Parallel downloads (agents download independently)
- No bottleneck at orchestrator
- HTTP caching (ETag, Cache-Control)

**Performance:**
- **First task:** Download module (1-5 seconds)
- **Subsequent tasks:** Cache hit (instant)
- **Concurrent agents:** Unlimited (CDN scales)
- **Bandwidth:** CDN handles it

**Example with 1000 agents:**
```
Initial deployment:
- 1000 agents × 10 modules × 1MB = 10GB download
- Spread over 5 minutes = 2GB/min = 34MB/s
- GitHub CDN easily handles this

Ongoing operations:
- Modules cached on all agents
- Zero bandwidth after initial download
- New modules: Only download when needed
```

---

#### Option C: Module Registry ⚠️ Requires Planning

**Module serving:**
- Single point of failure (unless HA)
- Requires load balancing for scale
- Database for metadata (optional)
- Need to plan capacity

**Performance:**
- **First task:** Registry → Agent (network dependent)
- **Subsequent tasks:** Cache hit (instant)
- **Concurrent agents:** Limited by registry capacity
- **Bandwidth:** Must provision appropriately

**Example with 1000 agents:**
```
Initial deployment:
- 1000 agents × 10 modules × 1MB = 10GB download
- Registry must serve 10GB
- Need sufficient bandwidth and connections
- Requires load balancer for HA

Ongoing operations:
- Registry must handle all requests
- Database queries for metadata
- Monitoring required
- Need backup strategy
```

**Capacity planning:**
```
1000 agents, 100 modules, 1MB average:
- Storage: 100MB modules
- Peak bandwidth: 1000 × 10MB = 10GB burst
- Connections: 1000 concurrent
- CPU: Minimal (static file serving)
- RAM: 512MB-1GB (caching)
```

---

### 4. Reliability & Availability

#### Option A: NATS-based ✅ Highly Available

**Single point of failure:** None
- If GitHub is down: Agents use cached modules
- If CDN is down: Agents use cached modules
- If agent can't download: Task fails, agent retries later

**Failure scenarios:**
```
Scenario 1: Module URL unreachable
├─ Agent logs error
├─ Task fails with clear error message
└─ Retry on next execution (module may be cached elsewhere)

Scenario 2: Checksum mismatch
├─ Agent rejects module
├─ Task fails safely
└─ Prevents corrupted module execution

Scenario 3: Network partition
├─ Agent operates from cache
├─ New modules unavailable
└─ Existing modules work fine
```

**Recovery:** Automatic
- Agents retry download
- Cache survives reboots
- No manual intervention needed

---

#### Option C: Module Registry ⚠️ Needs HA Setup

**Single point of failure:** Registry server
- If registry is down: New modules unavailable
- If registry is slow: All agents affected
- If database is down: Metadata unavailable

**Requires:**
- Load balancer (HAProxy, nginx)
- Multiple registry nodes
- Health checks
- Database replication (if used)
- Backup/restore procedures

**Failure scenarios:**
```
Scenario 1: Registry server down
├─ Load balancer fails over to backup
├─ Brief interruption (2-5 seconds)
└─ Requires monitoring to detect

Scenario 2: Database down (if used)
├─ Cannot serve metadata
├─ All downloads fail
└─ Need database HA

Scenario 3: Network partition
├─ Agents in isolated network can't reach registry
├─ Must use cached modules only
└─ New modules unavailable until network restored
```

**Recovery:** Manual + Automated
- Load balancer handles node failures
- Database failover needed
- Monitoring alerts required
- On-call engineer may be needed

---

### 5. Security

#### Option A: NATS-based ✅ Secure by Default

**Threat model:**
```
1. Man-in-the-Middle (MITM)
   ✅ Mitigated: HTTPS for module downloads
   ✅ Mitigated: SHA256 checksum verification

2. Module tampering
   ✅ Mitigated: Checksum mismatch = rejection

3. Compromised module source (GitHub)
   ✅ Mitigated: Use specific release tags
   ✅ Mitigated: Sign releases (optional)

4. Replay attacks
   ✅ Not applicable: Modules are content-addressed
```

**Security features:**
- TLS for downloads (HTTPS)
- Checksum verification (SHA256)
- No credentials needed (public GitHub)
- Immutable releases (v1.0.0 never changes)

**Configuration:**
```yaml
# No credentials in config
# Modules from trusted source (GitHub releases)
execution:
  module_cache: /etc/froyo/modules
  verify_checksums: true  # Enforced
```

---

#### Option C: Module Registry ⚠️ More Attack Surface

**Threat model:**
```
1. Man-in-the-Middle (MITM)
   ⚠️  Requires TLS configuration
   ⚠️  Certificate management

2. Module tampering
   ✅ Mitigated: Checksum verification

3. Compromised registry
   ❌ Single point of trust
   ❌ All agents trust registry

4. Authentication bypass
   ⚠️  Need authentication mechanism
   ⚠️  Credential rotation required

5. DDoS attacks
   ❌ Registry must handle malicious traffic
   ❌ Rate limiting needed
```

**Security requirements:**
- TLS certificates (Let's Encrypt or commercial)
- Authentication (API keys, JWT, mTLS)
- Authorization (which agents can access which modules)
- Rate limiting (prevent abuse)
- Audit logging (track downloads)
- Intrusion detection
- Regular security updates

**Configuration:**
```yaml
execution:
  module_registry:
    url: "https://registry.openfroyo.io"
    auth:
      type: "api_key"
      key: "${REGISTRY_API_KEY}"  # Must manage secrets
    tls:
      verify: true
      ca_file: /etc/froyo/ca.crt
```

---

### 6. Operational Overhead

#### Option A: NATS-based ✅ Zero Ops

**Day 1:**
1. Build modules: `make build-wasm`
2. Create GitHub release: `gh release create v1.0.0 bin/modules/*.wasm`
3. Update orchestrator with release URL
4. Done!

**Ongoing:**
- New module release: Create new GitHub release
- No monitoring needed (GitHub handles it)
- No backups needed (GitHub is backed up)
- No scaling needed (CDN auto-scales)
- No certificates to renew
- No servers to patch

**Incident response:** None (GitHub's problem)

**Cost:** Engineer time = 0 hours/month

---

#### Option C: Module Registry ⚠️ Significant Ops

**Day 1:**
1. Provision servers (2+ for HA)
2. Configure load balancer
3. Deploy registry software
4. Set up TLS certificates
5. Configure monitoring
6. Set up backups
7. Test failover
8. Document runbooks

**Ongoing:**
- Monitor registry health (CPU, memory, disk, network)
- Renew TLS certificates (every 90 days)
- Apply security patches
- Scale based on load
- Backup module storage
- Test disaster recovery
- Rotate credentials
- Review audit logs

**Incident response:**
```
Alert: Registry server down
├─ Page on-call engineer
├─ Investigate issue
├─ Fail over to backup
├─ Root cause analysis
└─ Post-incident review
```

**Cost:** Engineer time = 4-8 hours/month

---

### 7. Versioning & Updates

#### Option A: NATS-based ✅ Simple Versioning

**Module updates:**
```bash
# Version 1.0.0
gh release create v1.0.0 bin/modules/*.wasm

# Version 1.1.0 (with updated modules)
make build-wasm
gh release create v1.1.0 bin/modules/*.wasm

# Orchestrator uses specific version
moduleURL := "https://github.com/.../download/v1.1.0/generic-file.wasm"
```

**Update strategy:**
```
Option 1: Immediate update
- Update orchestrator to use v1.1.0
- Next task downloads new module
- Old cache invalidated by checksum

Option 2: Gradual rollout
- 10% of agents use v1.1.0
- Monitor for issues
- Roll out to 100%

Option 3: Pin versions per datacenter
- DC1: v1.0.0
- DC2: v1.1.0 (canary)
- DC3: v1.0.0
```

**Rollback:**
```bash
# Simply point back to v1.0.0
moduleURL := "https://github.com/.../download/v1.0.0/generic-file.wasm"

# Immediate effect on next task
```

---

#### Option C: Module Registry ⚠️ Complex Versioning

**Module updates:**
```bash
# Upload new version to registry
curl -X POST https://registry/modules/generic-file/v1.1.0 \
  -H "Authorization: Bearer $TOKEN" \
  --data-binary @generic-file.wasm

# Update metadata
curl -X PUT https://registry/modules/generic-file/latest \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"version": "1.1.0"}'
```

**Version management:**
- Track versions in database
- Support version pinning
- Handle version resolution (latest, ~1.0, ^1.0)
- Garbage collect old versions

**Update strategy:**
```
Option 1: Update "latest" pointer
- Risky: All agents get new version immediately
- No gradual rollout

Option 2: Version per agent
- Complex: Track which agent uses which version
- Requires database

Option 3: Canary releases
- Requires feature flags
- Complex routing logic
```

**Rollback:**
```bash
# Update "latest" pointer
curl -X PUT https://registry/modules/generic-file/latest \
  -d '{"version": "1.0.0"}'

# Wait for agents to re-download
# Some agents may have v1.1.0 cached
```

---

### 8. Debugging & Observability

#### Option A: NATS-based ✅ Easy to Debug

**Agent logs:**
```
[INFO] Task received: abc123 (module: generic/file)
[INFO] Checking module cache: /etc/froyo/modules/generic-file.wasm
[INFO] Module not in cache, downloading...
[INFO] Downloading: https://github.com/.../v1.0.0/generic-file.wasm
[INFO] Download complete: 1.2MB in 0.3s
[INFO] Verifying checksum: sha256:abc123...
[INFO] Checksum OK
[INFO] Module cached: /etc/froyo/modules/generic-file.wasm
[INFO] Executing module...
[INFO] Task completed: abc123 (status: ok)
```

**Troubleshooting:**
```bash
# Check if module is cached
ls -lh /etc/froyo/modules/

# Test download manually
curl -L https://github.com/.../generic-file.wasm -o /tmp/test.wasm
sha256sum /tmp/test.wasm

# Check agent logs
journalctl -u froyo-agent | grep "module"

# Verify checksum in task
# (visible in NATS message)
```

**No external systems to debug** ✅

---

#### Option C: Module Registry ⚠️ More Components

**Agent logs:**
```
[INFO] Task received: abc123 (module: generic/file)
[INFO] Checking module cache: /etc/froyo/modules/generic-file.wasm
[INFO] Module not in cache, querying registry...
[INFO] Registry request: GET https://registry/modules/generic-file/metadata
[ERROR] Registry request failed: connection timeout
[WARN] Retrying in 2s...
[INFO] Registry request: GET https://registry/modules/generic-file/metadata
[INFO] Metadata received: v1.0.0, checksum: abc123...
[INFO] Downloading module: GET https://registry/modules/generic-file/download
[INFO] Download complete: 1.2MB in 0.5s
[INFO] Verifying checksum...
[INFO] Checksum OK
[INFO] Module cached
[INFO] Executing module...
```

**Troubleshooting:**
```bash
# Check if module is cached
ls -lh /etc/froyo/modules/

# Test registry connectivity
curl https://registry/health

# Test module download
curl https://registry/modules/generic-file/metadata
curl https://registry/modules/generic-file/download -o /tmp/test.wasm

# Check registry logs (on registry server)
tail -f /var/log/froyo-registry/access.log

# Check database (if used)
psql -c "SELECT * FROM modules WHERE name='generic/file';"

# Check load balancer
curl https://lb/health
```

**Multiple systems to debug** ⚠️

---

### 9. Cost Analysis

#### Option A: NATS-based

**Infrastructure:**
- NATS server: $10-20/month (already deployed for agents)
- Module hosting: $0 (GitHub Releases)
- Total: ~$15/month

**Engineering:**
- Initial implementation: 2-3 hours
- Ongoing maintenance: 0 hours/month
- Cost: $0/month

**Total Cost Year 1:** ~$180

---

#### Option C: Module Registry

**Infrastructure:**
- NATS server: $10-20/month (already deployed)
- Registry servers (2): $40/month
- Load balancer: $10/month
- Database (optional): $15/month
- Monitoring: $10/month
- Total: ~$85/month

**Engineering:**
- Initial implementation: 16-24 hours ($2,000-3,000)
- Ongoing maintenance: 4-8 hours/month ($400-800/month)
- Cost: ~$600/month average

**Total Cost Year 1:** ~$10,000

---

## Decision Matrix

| Criteria | Option A: NATS-based | Option C: Registry | Winner |
|----------|---------------------|-------------------|---------|
| **Implementation Time** | 2-3 hours | 16-24 hours | 🥇 A |
| **Code Complexity** | 80 lines | 400+ lines | 🥇 A |
| **Infrastructure** | Minimal (GitHub) | Multiple servers | 🥇 A |
| **Operational Overhead** | Zero | Significant | 🥇 A |
| **Scalability** | CDN-backed | Must plan capacity | 🥇 A |
| **Reliability** | Highly available | Requires HA setup | 🥇 A |
| **Security** | Secure by default | More attack surface | 🥇 A |
| **Debugging** | Simple | Multiple systems | 🥇 A |
| **Cost (Year 1)** | $180 | $10,000 | 🥇 A |
| **Versioning** | Simple (Git tags) | Complex | 🥇 A |
| **Module Discovery** | Manual | Programmatic API | 🥇 C |
| **Access Control** | None | Fine-grained | 🥇 C |
| **Audit Trail** | GitHub audit log | Detailed logging | 🥇 C |
| **Offline Support** | Cache only | Cache only | 🟰 Tie |

**Score: Option A wins 10/13 criteria**

---

## When to Choose Each Option

### Choose Option A (NATS-based) if:

✅ You want to get agents working ASAP (2-3 hours)
✅ You have < 50 modules
✅ You're okay with public module hosting (GitHub Releases)
✅ You want zero operational overhead
✅ You have limited engineering resources
✅ You value simplicity
✅ You trust GitHub's infrastructure
✅ You don't need fine-grained access control

**Ideal for:**
- Startups / small teams
- MVP / proof of concept
- < 1000 agents
- Public modules (open source)

---

### Choose Option C (Module Registry) if:

✅ You need fine-grained access control (which agents see which modules)
✅ You need detailed audit trails
✅ You have strict security requirements (air-gapped networks)
✅ You have 100+ modules with complex dependencies
✅ You need programmatic module discovery
✅ You have dedicated DevOps team
✅ You can invest 16-24 hours upfront
✅ You can commit to ongoing maintenance

**Ideal for:**
- Large enterprises
- Highly regulated industries (finance, healthcare)
- Air-gapped / on-premises deployments
- 1000+ agents
- Private modules (proprietary)

---

## Hybrid Approach (Best of Both Worlds)

Start with Option A, migrate to Option C later:

### Phase 1: NATS-based (Month 1-6)
- Get agents working quickly
- Validate architecture
- Gather requirements
- Modules on GitHub Releases

### Phase 2: Migration Planning (Month 6-9)
- Evaluate actual needs
- Design registry if needed
- Build registry infrastructure
- Test with subset of modules

### Phase 3: Gradual Migration (Month 9-12)
- Move 10% of modules to registry
- Monitor performance
- Roll out to 100%
- Keep GitHub as backup

**Migration is seamless:**
```go
// Agent supports both!
task := &protocol.TaskMessage{
    Module:         "generic/file",
    ModuleURL:      "https://github.com/.../generic-file.wasm",  // Fallback
    ModuleRegistry: "https://registry/modules/generic-file",     // Preferred
}

// Agent tries registry first, falls back to URL
```

---

## Recommendation

**Start with Option A (NATS-based)**

**Rationale:**
1. ⏱️  Get agents working in 2-3 hours vs 16-24 hours
2. 💰 Save $10k in year 1
3. 🎯 80/20 rule: 80% of the value, 20% of the cost
4. 🔄 Can migrate to registry later if needed
5. ✅ Proven approach (Docker, Kubernetes use similar pattern)
6. 🚀 Focus on core product, not infrastructure

**Implementation Plan:**
1. Add `ModuleURL` and `ModuleChecksum` to `TaskMessage` (30 min)
2. Implement HTTP download in agent (60 min)
3. Add checksum verification (30 min)
4. Test with real tasks (30 min)
5. Document for users (30 min)

**Total: 3 hours to production-ready agents!**

---

## Code Examples

### Option A Implementation

**1. Update Protocol (5 min)**
```go
// internal/protocol/messages.go
type TaskMessage struct {
    TaskID         string
    Type           string
    Module         string
    ModuleURL      string  // NEW: HTTP URL to download module
    ModuleChecksum string  // NEW: SHA256 checksum
    Vars           map[string]interface{}
    Context        TaskContext
    Timeout        int
    Priority       int
    CreatedAt      time.Time
}
```

**2. Update Orchestrator (30 min)**
```go
// internal/orchestrator/executor/agent.go

// Module checksums (generate with: sha256sum *.wasm)
var moduleChecksums = map[string]string{
    "generic/file":    "abc123...",
    "generic/command": "def456...",
    "linux/package":   "ghi789...",
}

func (e *AgentExecutor) ExecuteTask(...) (*protocol.TaskResult, error) {
    task := &protocol.TaskMessage{
        TaskID:         taskID,
        Module:         module,
        ModuleURL:      getModuleURL(module, e.version),
        ModuleChecksum: moduleChecksums[module],
        Vars:           vars,
        // ...
    }
    // ... rest unchanged
}

func getModuleURL(module, version string) string {
    return fmt.Sprintf("https://github.com/piwi3910/openfroyo/releases/download/v%s/%s.wasm",
        version, strings.ReplaceAll(module, "/", "-"))
}
```

**3. Update Agent (90 min)**
```go
// internal/agent/executor/executor.go
func (e *Executor) getModulePath(task *protocol.TaskMessage) (string, error) {
    cachePath := filepath.Join(e.config.Execution.ModuleCache,
        strings.ReplaceAll(task.Module, "/", "-")+".wasm")

    // Check cache and verify checksum
    if e.isValidCached(cachePath, task.ModuleChecksum) {
        e.logger.Printf("[DEBUG] Module cache hit: %s", task.Module)
        return cachePath, nil
    }

    // Download module
    e.logger.Printf("[INFO] Downloading module: %s from %s", task.Module, task.ModuleURL)
    if err := e.downloadModule(task.ModuleURL, cachePath, task.ModuleChecksum); err != nil {
        return "", fmt.Errorf("failed to download module: %w", err)
    }

    return cachePath, nil
}

func (e *Executor) isValidCached(path, expectedChecksum string) bool {
    data, err := os.ReadFile(path)
    if err != nil {
        return false
    }

    hash := sha256.Sum256(data)
    actualChecksum := hex.EncodeToString(hash[:])

    return actualChecksum == expectedChecksum
}

func (e *Executor) downloadModule(url, dest, expectedChecksum string) error {
    // Download to temp file
    tmpFile, err := os.CreateTemp("", "module-*.wasm")
    if err != nil {
        return err
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()

    // HTTP GET with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
    }

    // Download and hash simultaneously
    hash := sha256.New()
    writer := io.MultiWriter(tmpFile, hash)

    written, err := io.Copy(writer, resp.Body)
    if err != nil {
        return err
    }

    // Verify checksum
    actualChecksum := hex.EncodeToString(hash.Sum(nil))
    if actualChecksum != expectedChecksum {
        return fmt.Errorf("checksum mismatch: expected %s, got %s",
            expectedChecksum, actualChecksum)
    }

    e.logger.Printf("[INFO] Module downloaded: %d bytes, checksum OK", written)

    // Move to cache
    return os.Rename(tmpFile.Name(), dest)
}
```

**4. Generate Checksums (5 min)**
```bash
#!/bin/bash
# scripts/generate-checksums.sh

cd bin/modules/
for module in *.wasm; do
    checksum=$(sha256sum $module | awk '{print $1}')
    name=$(basename $module .wasm | tr '-' '/')
    echo "\"$name\": \"$checksum\","
done
```

**Done!** Total implementation: ~2.5 hours

---

## References

- [HTTP Caching Best Practices](https://web.dev/http-cache/)
- [Content Addressable Storage](https://en.wikipedia.org/wiki/Content-addressable_storage)
- [GitHub Releases Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [SHA-256 Hash](https://en.wikipedia.org/wiki/SHA-2)
