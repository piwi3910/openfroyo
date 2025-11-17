package nats

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Config contains NATS connection configuration for orchestrator
type Config struct {
	Servers []string      // NATS server URLs
	Timeout time.Duration // Connection timeout
	Token   string        // Authentication token (optional)
}

// Client wraps a NATS connection for the orchestrator
type Client struct {
	conn   *nats.Conn
	config *Config
	logger *log.Logger
}

// NewClient creates a new orchestrator NATS client
func NewClient(cfg *Config, logger *log.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("at least one NATS server is required")
	}

	// Set default timeout
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &Client{
		config: cfg,
		logger: logger,
	}, nil
}

// Connect establishes a connection to NATS server(s)
func (c *Client) Connect() error {
	opts := []nats.Option{
		nats.Name("froyo-orchestrator"),
		nats.Timeout(c.config.Timeout),
		nats.MaxReconnects(-1), // Infinite reconnects
		nats.ReconnectWait(2 * time.Second),
	}

	// Add token authentication if provided
	if c.config.Token != "" {
		opts = append(opts, nats.Token(c.config.Token))
	}

	// Connect to NATS
	serverURL := strings.Join(c.config.Servers, ",")
	conn, err := nats.Connect(serverURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = conn
	c.logger.Printf("[INFO] Orchestrator connected to NATS: %s", conn.ConnectedUrl())

	return nil
}

// Close closes the NATS connection
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
		c.logger.Println("[INFO] Orchestrator NATS connection closed")
	}
	return nil
}

// PublishTask publishes a task to an agent
func (c *Client) PublishTask(subject string, task *protocol.TaskMessage) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	data, err := protocol.MarshalTask(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := c.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish task: %w", err)
	}

	c.logger.Printf("[DEBUG] Published task %s to %s", task.TaskID, subject)
	return nil
}

// SubscribeToResults subscribes to task results from agents
func (c *Client) SubscribeToResults(subject string, handler func(*protocol.TaskResult)) (*nats.Subscription, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		result, err := protocol.UnmarshalTaskResult(msg.Data)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to unmarshal task result: %v", err)
			return
		}
		handler(result)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to results: %w", err)
	}

	c.logger.Printf("[DEBUG] Subscribed to results: %s", subject)
	return sub, nil
}

// SubscribeToHealth subscribes to agent health updates
func (c *Client) SubscribeToHealth(subject string, handler func(*protocol.HealthStatus)) (*nats.Subscription, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		health, err := protocol.UnmarshalHealthStatus(msg.Data)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to unmarshal health status: %v", err)
			return
		}
		handler(health)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to health: %w", err)
	}

	c.logger.Printf("[DEBUG] Subscribed to health: %s", subject)
	return sub, nil
}

// Request makes a request and waits for a response with timeout
func (c *Client) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}
	return c.conn.Request(subject, data, timeout)
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
