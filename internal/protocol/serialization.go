package protocol

import (
	"encoding/json"
	"fmt"
)

// Marshal serializes a message to JSON bytes
func Marshal(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}
	return data, nil
}

// Unmarshal deserializes JSON bytes into a message
func Unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}

// MarshalTask serializes a TaskMessage to JSON
func MarshalTask(task *TaskMessage) ([]byte, error) {
	return Marshal(task)
}

// UnmarshalTask deserializes a TaskMessage from JSON
func UnmarshalTask(data []byte) (*TaskMessage, error) {
	var task TaskMessage
	if err := Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// MarshalTaskResult serializes a TaskResult to JSON
func MarshalTaskResult(result *TaskResult) ([]byte, error) {
	return Marshal(result)
}

// UnmarshalTaskResult deserializes a TaskResult from JSON
func UnmarshalTaskResult(data []byte) (*TaskResult, error) {
	var result TaskResult
	if err := Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalHealthStatus serializes a HealthStatus to JSON
func MarshalHealthStatus(health *HealthStatus) ([]byte, error) {
	return Marshal(health)
}

// UnmarshalHealthStatus deserializes a HealthStatus from JSON
func UnmarshalHealthStatus(data []byte) (*HealthStatus, error) {
	var health HealthStatus
	if err := Unmarshal(data, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

// MarshalControlMessage serializes a ControlMessage to JSON
func MarshalControlMessage(ctrl *ControlMessage) ([]byte, error) {
	return Marshal(ctrl)
}

// UnmarshalControlMessage deserializes a ControlMessage from JSON
func UnmarshalControlMessage(data []byte) (*ControlMessage, error) {
	var ctrl ControlMessage
	if err := Unmarshal(data, &ctrl); err != nil {
		return nil, err
	}
	return &ctrl, nil
}

// MarshalAgentRegistration serializes an AgentRegistration to JSON
func MarshalAgentRegistration(reg *AgentRegistration) ([]byte, error) {
	return Marshal(reg)
}

// UnmarshalAgentRegistration deserializes an AgentRegistration from JSON
func UnmarshalAgentRegistration(data []byte) (*AgentRegistration, error) {
	var reg AgentRegistration
	if err := Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}
