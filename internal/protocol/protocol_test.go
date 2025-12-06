package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskMessageSerialization(t *testing.T) {
	msg := &TaskMessage{
		TaskID:  "task-123",
		Type:    TaskTypeModule,
		Module:  "generic/command",
		Vars:    map[string]interface{}{"cmd": "echo hello"},
		Context: TaskContext{Host: "web-01", TaskName: "Test task"},
		Timeout: 300,
	}

	// Serialize
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal TaskMessage: %v", err)
	}

	// Deserialize
	var decoded TaskMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TaskMessage: %v", err)
	}

	if decoded.TaskID != msg.TaskID {
		t.Errorf("Expected TaskID '%s', got '%s'", msg.TaskID, decoded.TaskID)
	}
	if decoded.Type != msg.Type {
		t.Errorf("Expected Type '%s', got '%s'", msg.Type, decoded.Type)
	}
	if decoded.Module != msg.Module {
		t.Errorf("Expected Module '%s', got '%s'", msg.Module, decoded.Module)
	}
}

func TestTaskResultSerialization(t *testing.T) {
	now := time.Now()
	result := &TaskResult{
		TaskID:      "task-123",
		AgentID:     "agent-abc",
		Status:      TaskStatusOK,
		Message:     "Success",
		Facts:       map[string]interface{}{"output": "hello"},
		StartedAt:   now,
		CompletedAt: now.Add(5 * time.Second),
		Duration:    5.0,
		Output:      "hello",
	}

	// Serialize
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal TaskResult: %v", err)
	}

	// Deserialize
	var decoded TaskResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TaskResult: %v", err)
	}

	if decoded.TaskID != result.TaskID {
		t.Errorf("Expected TaskID '%s', got '%s'", result.TaskID, decoded.TaskID)
	}
	if decoded.Status != result.Status {
		t.Errorf("Expected Status '%s', got '%s'", result.Status, decoded.Status)
	}
	if decoded.Duration != result.Duration {
		t.Errorf("Expected Duration %f, got %f", result.Duration, decoded.Duration)
	}
}

func TestSubjectBuilder(t *testing.T) {
	sb := NewSubjectBuilder("us-east-1", "agent-123")

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"AgentTasks", sb.AgentTasks, "openfroyo.us-east-1.agents.agent-123.tasks"},
		{"AgentResults", sb.AgentResults, "openfroyo.us-east-1.agents.agent-123.results"},
		{"AgentControl", sb.AgentControl, "openfroyo.us-east-1.agents.agent-123.control"},
		{"AgentHealth", sb.AgentHealth, "openfroyo.us-east-1.agents.agent-123.health"},
		{"AgentRegistration", sb.AgentRegistration, "openfroyo.us-east-1.agents.agent-123.registration"},
		{"BroadcastTasks", sb.BroadcastTasks, "openfroyo.broadcast.tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestHealthStatusSerialization(t *testing.T) {
	status := &HealthStatus{
		AgentID:        "agent-123",
		Datacenter:     "us-east-1",
		Status:         "healthy",
		Uptime:         3600,
		CPUPercent:     25.5,
		MemoryMB:       512,
		TasksExecuted:  100,
		TasksSucceeded: 95,
		TasksFailed:    5,
		Timestamp:      time.Now(),
	}

	// Serialize
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal HealthStatus: %v", err)
	}

	// Deserialize
	var decoded HealthStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal HealthStatus: %v", err)
	}

	if decoded.AgentID != status.AgentID {
		t.Errorf("Expected AgentID '%s', got '%s'", status.AgentID, decoded.AgentID)
	}
	if decoded.TasksExecuted != status.TasksExecuted {
		t.Errorf("Expected TasksExecuted %d, got %d", status.TasksExecuted, decoded.TasksExecuted)
	}
}

func TestControlMessageTypes(t *testing.T) {
	types := []string{
		ControlTypeShutdown,
		ControlTypeReload,
		ControlTypeUpdateSchedule,
	}

	for _, ct := range types {
		msg := &ControlMessage{
			Type:    ct,
			Payload: map[string]interface{}{"key": "value"},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal ControlMessage with type %s: %v", ct, err)
		}

		var decoded ControlMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Failed to unmarshal ControlMessage with type %s: %v", ct, err)
		}

		if decoded.Type != ct {
			t.Errorf("Expected type '%s', got '%s'", ct, decoded.Type)
		}
	}
}

func TestAgentRegistrationSerialization(t *testing.T) {
	reg := &AgentRegistration{
		AgentID:    "agent-123",
		Hostname:   "web-01.example.com",
		Datacenter: "us-east-1",
		Version:    "1.0.0",
		Tags:       map[string]string{"role": "web", "env": "production"},
		Timestamp:  time.Now(),
	}

	// Serialize
	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("Failed to marshal AgentRegistration: %v", err)
	}

	// Deserialize
	var decoded AgentRegistration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AgentRegistration: %v", err)
	}

	if decoded.AgentID != reg.AgentID {
		t.Errorf("Expected AgentID '%s', got '%s'", reg.AgentID, decoded.AgentID)
	}
	if decoded.Tags["role"] != "web" {
		t.Errorf("Expected tag 'role' to be 'web', got '%s'", decoded.Tags["role"])
	}
}

func TestTaskStatusValues(t *testing.T) {
	statuses := []string{
		TaskStatusOK,
		TaskStatusChanged,
		TaskStatusFailed,
	}

	for _, status := range statuses {
		if status == "" {
			t.Errorf("Status constant should not be empty")
		}
	}

	// Verify values
	if TaskStatusOK != "ok" {
		t.Errorf("Expected TaskStatusOK to be 'ok', got '%s'", TaskStatusOK)
	}
	if TaskStatusChanged != "changed" {
		t.Errorf("Expected TaskStatusChanged to be 'changed', got '%s'", TaskStatusChanged)
	}
	if TaskStatusFailed != "failed" {
		t.Errorf("Expected TaskStatusFailed to be 'failed', got '%s'", TaskStatusFailed)
	}
}

func TestTaskTypeValues(t *testing.T) {
	if TaskTypeModule != "module" {
		t.Errorf("Expected TaskTypeModule to be 'module', got '%s'", TaskTypeModule)
	}
	if TaskTypeCommand != "command" {
		t.Errorf("Expected TaskTypeCommand to be 'command', got '%s'", TaskTypeCommand)
	}
}
