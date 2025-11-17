package nats

import (
	"fmt"
	"log"

	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Publisher manages publishing messages to NATS
type Publisher struct {
	client *Client
	logger *log.Logger
}

// NewPublisher creates a new publisher
func NewPublisher(client *Client, logger *log.Logger) *Publisher {
	return &Publisher{
		client: client,
		logger: logger,
	}
}

// PublishTaskResult publishes a task execution result
func (p *Publisher) PublishTaskResult(subject string, result *protocol.TaskResult) error {
	data, err := protocol.MarshalTaskResult(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result: %w", err)
	}

	if err := p.client.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish task result: %w", err)
	}

	p.logger.Printf("[DEBUG] Published task result: %s (status: %s)", result.TaskID, result.Status)
	return nil
}

// PublishHealthStatus publishes agent health status
func (p *Publisher) PublishHealthStatus(subject string, health *protocol.HealthStatus) error {
	data, err := protocol.MarshalHealthStatus(health)
	if err != nil {
		return fmt.Errorf("failed to marshal health status: %w", err)
	}

	if err := p.client.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish health status: %w", err)
	}

	p.logger.Printf("[DEBUG] Published health status: %s (status: %s)", health.AgentID, health.Status)
	return nil
}

// PublishAgentRegistration publishes agent registration/announcement
func (p *Publisher) PublishAgentRegistration(subject string, reg *protocol.AgentRegistration) error {
	data, err := protocol.MarshalAgentRegistration(reg)
	if err != nil {
		return fmt.Errorf("failed to marshal agent registration: %w", err)
	}

	if err := p.client.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish agent registration: %w", err)
	}

	p.logger.Printf("[INFO] Published agent registration: %s", reg.AgentID)
	return nil
}
