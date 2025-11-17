package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/piwi3910/openfroyo/internal/agent/config"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Executor manages task execution
type Executor struct {
	config      *config.Config
	logger      *log.Logger
	semaphore   chan struct{}
	activeTasks sync.Map
	stats       ExecutionStats
	statsMutex  sync.RWMutex
}

// ExecutionStats tracks execution statistics
type ExecutionStats struct {
	TotalExecuted int64
	Succeeded     int64
	Failed        int64
	LastTaskAt    *time.Time
}

// NewExecutor creates a new executor
func NewExecutor(cfg *config.Config, logger *log.Logger) (*Executor, error) {
	// Create work directory if it doesn't exist
	if err := os.MkdirAll(cfg.Execution.WorkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Create module cache directory if it doesn't exist
	if err := os.MkdirAll(cfg.Execution.ModuleCache, 0755); err != nil {
		return nil, fmt.Errorf("failed to create module cache directory: %w", err)
	}

	return &Executor{
		config:    cfg,
		logger:    logger,
		semaphore: make(chan struct{}, cfg.Execution.MaxConcurrent),
	}, nil
}

// Execute executes a task and returns the result
func (e *Executor) Execute(task *protocol.TaskMessage) (*protocol.TaskResult, error) {
	// Acquire semaphore slot (limit concurrency)
	e.semaphore <- struct{}{}
	defer func() { <-e.semaphore }()

	startTime := time.Now()

	// Track active task
	e.activeTasks.Store(task.TaskID, task)
	defer e.activeTasks.Delete(task.TaskID)

	e.logger.Printf("[INFO] Executing task: %s (module: %s)", task.TaskID, task.Module)

	// Execute based on task type
	var result *protocol.TaskResult
	var err error

	switch task.Type {
	case protocol.TaskTypeModule:
		result, err = e.executeModule(task)
	case protocol.TaskTypeCommand:
		result, err = e.executeCommand(task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	if err != nil {
		e.logger.Printf("[ERROR] Task execution failed: %s: %v", task.TaskID, err)
		result = &protocol.TaskResult{
			TaskID:      task.TaskID,
			AgentID:     e.config.Agent.ID,
			Status:      protocol.TaskStatusFailed,
			Message:     "Task execution failed",
			Error:       err.Error(),
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			Duration:    time.Since(startTime).Seconds(),
			Facts:       make(map[string]interface{}),
		}
	}

	// Update statistics
	e.updateStats(result.Status)

	e.logger.Printf("[INFO] Task completed: %s (status: %s, duration: %.2fs)",
		task.TaskID, result.Status, result.Duration)

	return result, nil
}

// executeModule executes a WASM module task
func (e *Executor) executeModule(task *protocol.TaskMessage) (*protocol.TaskResult, error) {
	startTime := time.Now()

	// Get module path (from cache or download)
	modulePath, err := e.getModulePath(task.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	// Prepare input JSON
	inputData := map[string]interface{}{
		"vars":    task.Vars,
		"context": task.Context,
	}

	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Encode input as base64
	inputBase64 := base64.StdEncoding.EncodeToString(inputJSON)

	// Execute froyo-runner
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = e.config.Execution.TaskTimeout
	}

	output, err := e.runWithTimeout(
		e.config.Execution.RunnerPath,
		[]string{"--module", modulePath, "--input-base64", inputBase64},
		timeout,
	)

	if err != nil {
		return nil, fmt.Errorf("runner execution failed: %w", err)
	}

	// Parse output JSON
	var moduleOutput struct {
		Status  string                 `json:"status"`
		Message string                 `json:"message"`
		Facts   map[string]interface{} `json:"facts"`
	}

	if err := json.Unmarshal(output, &moduleOutput); err != nil {
		return nil, fmt.Errorf("failed to parse module output: %w", err)
	}

	// Build result
	result := &protocol.TaskResult{
		TaskID:      task.TaskID,
		AgentID:     e.config.Agent.ID,
		Status:      moduleOutput.Status,
		Message:     moduleOutput.Message,
		Facts:       moduleOutput.Facts,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Duration:    time.Since(startTime).Seconds(),
		Output:      string(output),
	}

	return result, nil
}

// executeCommand executes a shell command task
func (e *Executor) executeCommand(task *protocol.TaskMessage) (*protocol.TaskResult, error) {
	startTime := time.Now()

	// Get command from vars
	cmdStr, ok := task.Vars["cmd"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'cmd' in task vars")
	}

	// Execute command
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = e.config.Execution.TaskTimeout
	}

	output, err := e.runWithTimeout("sh", []string{"-c", cmdStr}, timeout)

	status := protocol.TaskStatusOK
	errMsg := ""
	if err != nil {
		status = protocol.TaskStatusFailed
		errMsg = err.Error()
	}

	result := &protocol.TaskResult{
		TaskID:      task.TaskID,
		AgentID:     e.config.Agent.ID,
		Status:      status,
		Message:     "Command executed",
		Facts:       map[string]interface{}{"output": string(output)},
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Duration:    time.Since(startTime).Seconds(),
		Output:      string(output),
		Error:       errMsg,
	}

	return result, nil
}

// getModulePath returns the path to a WASM module (from cache or downloads it)
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

// runWithTimeout runs a command with a timeout
func (e *Executor) runWithTimeout(name string, args []string, timeout time.Duration) ([]byte, error) {
	cmd := exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wait with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		// Timeout - kill process
		if err := cmd.Process.Kill(); err != nil {
			e.logger.Printf("[WARN] Failed to kill process: %v", err)
		}
		return nil, fmt.Errorf("command timed out after %v", timeout)
	case err := <-done:
		if err != nil {
			// Command failed - return stderr
			if stderr.Len() > 0 {
				return stderr.Bytes(), err
			}
			return stdout.Bytes(), err
		}
		return stdout.Bytes(), nil
	}
}

// GetStats returns current execution statistics
func (e *Executor) GetStats() ExecutionStats {
	e.statsMutex.RLock()
	defer e.statsMutex.RUnlock()
	return e.stats
}

// updateStats updates execution statistics
func (e *Executor) updateStats(status string) {
	e.statsMutex.Lock()
	defer e.statsMutex.Unlock()

	e.stats.TotalExecuted++
	now := time.Now()
	e.stats.LastTaskAt = &now

	switch status {
	case protocol.TaskStatusOK, protocol.TaskStatusChanged:
		e.stats.Succeeded++
	case protocol.TaskStatusFailed:
		e.stats.Failed++
	}
}

// GetActiveTasks returns the number of currently active tasks
func (e *Executor) GetActiveTasks() int {
	count := 0
	e.activeTasks.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
