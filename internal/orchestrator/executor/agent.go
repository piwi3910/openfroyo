package executor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	orchestratorNats "github.com/piwi3910/openfroyo/internal/orchestrator/nats"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// AgentExecutor executes tasks on remote agents via NATS
type AgentExecutor struct {
	natsClient    *orchestratorNats.Client
	logger        *log.Logger
	resultChans   map[string]chan *protocol.TaskResult
	resultMutex   sync.RWMutex
	resultTimeout time.Duration
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(natsClient *orchestratorNats.Client, logger *log.Logger) *AgentExecutor {
	return &AgentExecutor{
		natsClient:    natsClient,
		logger:        logger,
		resultChans:   make(map[string]chan *protocol.TaskResult),
		resultTimeout: 300 * time.Second, // 5 minutes default
	}
}

// Initialize sets up result subscriptions
func (e *AgentExecutor) Initialize(datacenter string) error {
	// Subscribe to all agent results in this datacenter
	resultSubject := fmt.Sprintf("%s.%s.%s.*.%s",
		protocol.SubjectPrefix, datacenter, protocol.SubjectAgents, protocol.SuffixResults)

	_, err := e.natsClient.SubscribeToResults(resultSubject, e.handleResult)
	if err != nil {
		return fmt.Errorf("failed to subscribe to agent results: %w", err)
	}

	e.logger.Printf("[INFO] Agent executor initialized for datacenter: %s", datacenter)
	return nil
}

// ExecuteTask executes a task on an agent and waits for the result
func (e *AgentExecutor) ExecuteTask(datacenter, agentID, taskName, module string, vars map[string]interface{}, timeout int) (*protocol.TaskResult, error) {
	// Generate unique task ID
	taskID := uuid.New().String()

	// Create task message
	task := &protocol.TaskMessage{
		TaskID:  taskID,
		Type:    protocol.TaskTypeModule,
		Module:  module,
		Vars:    vars,
		Context: protocol.TaskContext{
			TaskName:   taskName,
			Host:       agentID,
			Datacenter: datacenter,
		},
		Timeout:   timeout,
		Priority:  5,
		CreatedAt: time.Now(),
	}

	// Create result channel for this task
	resultChan := make(chan *protocol.TaskResult, 1)
	e.resultMutex.Lock()
	e.resultChans[taskID] = resultChan
	e.resultMutex.Unlock()

	// Clean up result channel when done
	defer func() {
		e.resultMutex.Lock()
		delete(e.resultChans, taskID)
		e.resultMutex.Unlock()
		close(resultChan)
	}()

	// Publish task to agent
	subject := protocol.AgentTasksSubject(datacenter, agentID)
	if err := e.natsClient.PublishTask(subject, task); err != nil {
		return nil, fmt.Errorf("failed to publish task: %w", err)
	}

	e.logger.Printf("[INFO] Task %s published to agent %s", taskID, agentID)

	// Wait for result with timeout
	resultTimeout := time.Duration(timeout) * time.Second
	if timeout == 0 {
		resultTimeout = e.resultTimeout
	}

	select {
	case result := <-resultChan:
		e.logger.Printf("[INFO] Task %s completed: %s", taskID, result.Status)
		return result, nil

	case <-time.After(resultTimeout):
		return nil, fmt.Errorf("task %s timed out after %v", taskID, resultTimeout)
	}
}

// handleResult handles incoming task results
func (e *AgentExecutor) handleResult(result *protocol.TaskResult) {
	e.logger.Printf("[DEBUG] Received result for task: %s (status: %s)", result.TaskID, result.Status)

	e.resultMutex.RLock()
	resultChan, exists := e.resultChans[result.TaskID]
	e.resultMutex.RUnlock()

	if exists {
		// Send result to waiting channel (non-blocking)
		select {
		case resultChan <- result:
			// Result delivered
		default:
			e.logger.Printf("[WARN] Result channel full for task: %s", result.TaskID)
		}
	} else {
		e.logger.Printf("[DEBUG] No waiting channel for task: %s (may have timed out)", result.TaskID)
	}
}

// SetResultTimeout sets the default result timeout
func (e *AgentExecutor) SetResultTimeout(timeout time.Duration) {
	e.resultTimeout = timeout
}
