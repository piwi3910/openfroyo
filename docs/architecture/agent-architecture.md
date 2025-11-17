# OpenFroyo Agent Architecture

This document describes the architecture for OpenFroyo's agent-based (pull mode) execution system.

## Overview

OpenFroyo supports two execution modes:

1. **Push Mode (SSH-based)**: Orchestrator connects via SSH and executes tasks directly
2. **Pull Mode (Agent-based)**: Agents connect to NATS message queue and pull tasks

Both modes can coexist in a hybrid deployment, with some hosts using SSH and others using agents.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        OpenFroyo CLI                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ Stack Parser │  │   Inventory  │  │  Execution Engine    │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
│         │                 │                      │              │
│         └─────────────────┴──────────────────────┘              │
│                            │                                     │
│         ┌──────────────────┴──────────────────┐                │
│         │                                      │                │
│    ┌────▼──────┐                     ┌────────▼────────┐       │
│    │ SSH Mode  │                     │   Agent Mode    │       │
│    │ Executor  │                     │  NATS Publisher │       │
│    └─────┬─────┘                     └────────┬────────┘       │
└──────────┼──────────────────────────────────────┼──────────────┘
           │                                      │
           │                                      │
      SSH  │                            ┌─────────▼─────────┐
           │                            │   NATS Server     │
           │                            │  Message Queue    │
           │                            └─────────┬─────────┘
           │                                      │
           │                            Subject:  │
           │                       openfroyo.{dc}.agents.{id}
           │                                      │
           │                                      │
     ┌─────▼─────────────┐            ┌──────────▼──────────┐
     │  Target Host      │            │  Agent Host         │
     │  (SSH Mode)       │            │  (Pull Mode)        │
     │                   │            │                     │
     │ ┌──────────────┐  │            │ ┌────────────────┐ │
     │ │ froyo-runner │  │            │ │  froyo-agent   │ │
     │ └──────────────┘  │            │ │                │ │
     └───────────────────┘            │ │ ┌────────────┐ │ │
                                      │ │ │NATS Client │ │ │
                                      │ │ └────────────┘ │ │
                                      │ │ ┌────────────┐ │ │
                                      │ │ │  Executor  │ │ │
                                      │ │ └────────────┘ │ │
                                      │ │ ┌────────────┐ │ │
                                      │ │ │ Scheduler  │ │ │
                                      │ │ └────────────┘ │ │
                                      │ └────────────────┘ │
                                      └─────────────────────┘
```

## Components

### 1. froyo-agent (New)

The agent binary that runs on target hosts in pull mode.

**Responsibilities:**
- Connect to NATS message queue
- Subscribe to datacenter-specific subjects
- Pull and execute tasks from queue
- Execute periodic stacks (pull mode scheduling)
- Report task status and results
- Maintain health status

**Package Structure:**
```
cmd/froyo-agent/
  └── main.go                    # Agent entry point

internal/agent/
  ├── config/
  │   ├── config.go              # Configuration structures
  │   └── loader.go              # Load and validate config
  ├── nats/
  │   ├── client.go              # NATS client wrapper
  │   ├── subscriber.go          # Message subscription
  │   └── publisher.go           # Status/result publishing
  ├── executor/
  │   ├── executor.go            # Task execution engine
  │   ├── wasm.go                # WASM module execution
  │   └── runner.go              # froyo-runner integration
  ├── scheduler/
  │   ├── scheduler.go           # Pull mode scheduler
  │   ├── stack.go               # Stack execution
  │   └── cron.go                # Cron-like scheduling
  ├── health/
  │   ├── monitor.go             # Health monitoring
  │   └── reporter.go            # Health status reporting
  └── agent.go                   # Main agent orchestration
```

### 2. Orchestrator NATS Integration (Modified)

Extend existing orchestrator to support agent mode execution.

**Package Structure:**
```
internal/orchestrator/
  ├── nats/
  │   ├── publisher.go           # Publish tasks to agents
  │   ├── subscriber.go          # Subscribe to agent responses
  │   ├── router.go              # Subject-based routing
  │   └── tracker.go             # Task status tracking
  ├── executor/
  │   ├── ssh.go                 # Existing SSH executor
  │   └── agent.go               # New agent-based executor
  └── hybrid.go                  # Hybrid mode orchestration
```

### 3. Shared Types and Protocols

Common message formats and types used by both agent and orchestrator.

**Package Structure:**
```
internal/protocol/
  ├── messages.go                # Message type definitions
  ├── subjects.go                # NATS subject patterns
  └── serialization.go           # JSON serialization
```

## Message Protocol

### NATS Subject Patterns

```
openfroyo.{datacenter}.agents.{agent_id}.tasks       # Direct task to specific agent
openfroyo.{datacenter}.agents.*.tasks                # Broadcast to DC agents
openfroyo.{datacenter}.agents.{agent_id}.status      # Agent status updates
openfroyo.{datacenter}.agents.{agent_id}.results     # Task execution results
openfroyo.{datacenter}.agents.{agent_id}.health      # Health heartbeats
openfroyo.broadcast.tasks                            # Global broadcast
```

**Examples:**
```
openfroyo.dc1.agents.web-01.tasks                    # Task for web-01 in DC1
openfroyo.dc2.agents.*.tasks                         # All agents in DC2
openfroyo.dc1.agents.web-01.results                  # Results from web-01
openfroyo.dc1.agents.web-01.health                   # Health from web-01
```

### Message Types

#### 1. Task Message (Orchestrator → Agent)

```go
type TaskMessage struct {
    TaskID      string                 `json:"task_id"`      // Unique task ID
    Type        string                 `json:"type"`         // "module" or "command"
    Module      string                 `json:"module"`       // Module path (e.g., "generic/file")
    Vars        map[string]interface{} `json:"vars"`         // Input variables
    Context     TaskContext            `json:"context"`      // Execution context
    Timeout     int                    `json:"timeout"`      // Timeout in seconds
    Priority    int                    `json:"priority"`     // Task priority (1-10)
    CreatedAt   time.Time              `json:"created_at"`   // Task creation time
}

type TaskContext struct {
    StackName   string            `json:"stack_name"`    // Source stack name
    TaskName    string            `json:"task_name"`     // Task name
    Host        string            `json:"host"`          // Target hostname
    Datacenter  string            `json:"datacenter"`    // Datacenter
    Tags        map[string]string `json:"tags"`          // Additional tags
}
```

#### 2. Task Result Message (Agent → Orchestrator)

```go
type TaskResult struct {
    TaskID      string                 `json:"task_id"`      // Matching task ID
    AgentID     string                 `json:"agent_id"`     // Agent identifier
    Status      string                 `json:"status"`       // "ok", "changed", "failed"
    Message     string                 `json:"message"`      // Result message
    Facts       map[string]interface{} `json:"facts"`        // Discovered facts
    StartedAt   time.Time              `json:"started_at"`   // Execution start time
    CompletedAt time.Time              `json:"completed_at"` // Execution end time
    Duration    float64                `json:"duration"`     // Duration in seconds
    Error       string                 `json:"error"`        // Error message if failed
}
```

#### 3. Health Status Message (Agent → Orchestrator)

```go
type HealthStatus struct {
    AgentID         string            `json:"agent_id"`         // Agent identifier
    Hostname        string            `json:"hostname"`         // System hostname
    Datacenter      string            `json:"datacenter"`       // Datacenter location
    Version         string            `json:"version"`          // Agent version
    Status          string            `json:"status"`           // "healthy", "degraded", "unhealthy"
    Uptime          int64             `json:"uptime"`           // Uptime in seconds
    TasksExecuted   int64             `json:"tasks_executed"`   // Total tasks executed
    TasksSucceeded  int64             `json:"tasks_succeeded"`  // Successful tasks
    TasksFailed     int64             `json:"tasks_failed"`     // Failed tasks
    LastTaskAt      time.Time         `json:"last_task_at"`     // Last task execution
    CPUPercent      float64           `json:"cpu_percent"`      // CPU usage %
    MemoryMB        int64             `json:"memory_mb"`        // Memory usage MB
    Tags            map[string]string `json:"tags"`             // Agent tags
    Timestamp       time.Time         `json:"timestamp"`        // Status timestamp
}
```

#### 4. Pull Schedule Message (Configuration)

```go
type PullSchedule struct {
    StackName   string            `json:"stack_name"`   // Stack to execute
    Interval    int               `json:"interval"`     // Interval in seconds
    Enabled     bool              `json:"enabled"`      // Enable/disable
    Vars        map[string]interface{} `json:"vars"`    // Stack variables
    Tags        []string          `json:"tags"`         // Required agent tags
}
```

## Execution Flow

### Push Mode (SSH-based)

```
1. CLI parses stack file
2. Executor connects to host via SSH
3. Executor uploads froyo-runner if needed
4. Executor uploads WASM modules
5. Executor executes: froyo-runner --module x.wasm --input <json>
6. Executor receives result
7. Executor moves to next task
```

### Pull Mode (Agent-based)

```
1. Agent connects to NATS on startup
2. Agent subscribes to: openfroyo.{dc}.agents.{id}.tasks
3. Agent waits for task messages

When task arrives:
4. Agent receives TaskMessage from NATS
5. Agent downloads WASM module if not cached
6. Agent executes module with froyo-runner
7. Agent publishes TaskResult to NATS
8. Agent continues listening

Periodic execution (pull scheduling):
9. Scheduler triggers based on interval
10. Agent executes configured stack
11. Agent publishes results
12. Agent schedules next execution
```

### Orchestrator with Agents

```
1. CLI parses stack file
2. Executor detects agent mode for host
3. Executor publishes TaskMessage to NATS subject
4. Executor subscribes to result subject
5. Executor waits for TaskResult (with timeout)
6. Executor processes result
7. Executor moves to next task
```

## Agent Lifecycle

### Startup

```
1. Load configuration from /etc/froyo/agent.yml
2. Validate configuration
3. Initialize NATS client
4. Connect to NATS server(s)
5. Subscribe to task subjects
6. Subscribe to control subjects
7. Initialize task executor
8. Initialize pull scheduler
9. Start health monitoring
10. Publish initial health status
11. Enter main event loop
```

### Main Event Loop

```
while running:
    - Process incoming task messages
    - Execute scheduled stacks
    - Send health heartbeats (every 30s)
    - Clean up cached modules (if needed)
    - Handle reconnection if NATS disconnected
```

### Shutdown

```
1. Receive SIGTERM/SIGINT
2. Stop accepting new tasks
3. Wait for active tasks to complete (grace period)
4. Publish final health status (status: "shutdown")
5. Close NATS connection
6. Exit cleanly
```

## NATS Connection Management

### Connection Options

```go
nats.Connect(
    servers,
    nats.Name(agentID),
    nats.MaxReconnects(-1),           // Infinite reconnects
    nats.ReconnectWait(2*time.Second),
    nats.Timeout(5*time.Second),
    nats.PingInterval(20*time.Second),
    nats.MaxPingsOutstanding(2),
    nats.ReconnectHandler(onReconnect),
    nats.DisconnectHandler(onDisconnect),
    nats.ClosedHandler(onClosed),
)
```

### Reconnection Strategy

```
1. On disconnect: Log warning, continue operation
2. Reconnect automatically with exponential backoff
3. On reconnect: Re-subscribe to all subjects
4. Publish health status update
5. Resume normal operation
```

## Security Considerations

### NATS Authentication

```yaml
# Option 1: Token-based
nats:
  auth:
    token: "secret-token"

# Option 2: Credentials file
nats:
  auth:
    credentials_file: /etc/froyo/nats.creds

# Option 3: Username/Password
nats:
  auth:
    username: froyo-agent
    password: secret
```

### TLS Encryption

```yaml
nats:
  tls:
    enabled: true
    cert_file: /etc/froyo/tls/agent.crt
    key_file: /etc/froyo/tls/agent.key
    ca_file: /etc/froyo/tls/ca.crt
    verify: true
```

### Agent Permissions

- Agent runs as `froyo` user (non-root)
- Limited filesystem access via systemd security
- WASM modules execute in sandboxed environment
- Shell commands execute with agent user permissions

## Multi-Datacenter Deployment

### NATS Super Cluster

```
DC1 NATS Cluster ◄─────Gateway─────► DC2 NATS Cluster
       │                                     │
       │                                     │
  Agents (DC1)                          Agents (DC2)
```

### Subject Routing

- Agents only subscribe to their datacenter subjects
- Gateway routes messages between clusters
- Orchestrator publishes to specific DC subjects
- Broadcast subject reaches all DCs

### Example

```go
// Publish to DC1 agents only
nc.Publish("openfroyo.dc1.agents.web-01.tasks", taskMsg)

// Publish to all DC1 agents
nc.Publish("openfroyo.dc1.agents.*.tasks", taskMsg)

// Publish to all agents globally
nc.Publish("openfroyo.broadcast.tasks", taskMsg)
```

## Performance Considerations

### Agent Performance

- **Concurrency**: Max 5 concurrent tasks (configurable)
- **Module Caching**: Cache WASM modules locally
- **Message Batching**: Batch health status updates
- **Resource Limits**: systemd memory/CPU limits

### NATS Performance

- **Queue Groups**: Use queue groups for load balancing
- **Message Size**: Limit messages to 1MB (NATS default)
- **Persistence**: Use JetStream for reliable delivery (optional)
- **Monitoring**: Track message latency and throughput

## Monitoring and Observability

### Health Metrics

- Agent uptime
- Task execution counts (success/failure)
- Task execution duration
- CPU and memory usage
- NATS connection status
- Last heartbeat timestamp

### Logging

- Structured JSON logging
- Log levels: debug, info, warn, error
- Log rotation and compression
- Centralized log aggregation (syslog, journald)

### Alerting

- Agent disconnected > 5 minutes
- Task failure rate > 10%
- High CPU/memory usage
- NATS connection failures

## Configuration Reference

See `examples/agent-onboarding/templates/agent-config.yml.j2` for complete configuration example.

## Development Roadmap

### Phase 1: Core Agent (Current)
- [ ] Agent configuration loading
- [ ] NATS client connection
- [ ] Basic task execution
- [ ] Health monitoring

### Phase 2: Pull Mode
- [ ] Pull scheduler implementation
- [ ] Periodic stack execution
- [ ] Drift detection

### Phase 3: Orchestrator Integration
- [ ] NATS publisher in orchestrator
- [ ] Agent mode executor
- [ ] Hybrid mode support

### Phase 4: Advanced Features
- [ ] JetStream for reliable delivery
- [ ] Task prioritization
- [ ] Resource scheduling
- [ ] Agent clustering

## References

- [NATS Documentation](https://docs.nats.io/)
- [NATS Go Client](https://github.com/nats-io/nats.go)
- [Agent Onboarding Example](../../examples/agent-onboarding/)
