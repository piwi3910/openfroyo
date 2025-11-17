package nats

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// TaskHandler is a function that handles incoming task messages
type TaskHandler func(*protocol.TaskMessage) error

// Subscriber manages subscriptions to NATS subjects
type Subscriber struct {
	client       *Client
	logger       *log.Logger
	taskHandler  TaskHandler
	subscription *nats.Subscription
}

// NewSubscriber creates a new subscriber
func NewSubscriber(client *Client, logger *log.Logger) *Subscriber {
	return &Subscriber{
		client: client,
		logger: logger,
	}
}

// SubscribeToTasks subscribes to task messages for this agent
func (s *Subscriber) SubscribeToTasks(subject string, handler TaskHandler) error {
	s.taskHandler = handler

	sub, err := s.client.Subscribe(subject, s.handleTaskMessage)
	if err != nil {
		return err
	}

	s.subscription = sub
	s.logger.Printf("[INFO] Listening for tasks on: %s", subject)
	return nil
}

// SubscribeToBroadcast subscribes to broadcast task messages
func (s *Subscriber) SubscribeToBroadcast(subject string, handler TaskHandler) error {
	_, err := s.client.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Printf("[DEBUG] Received broadcast task on: %s", subject)
		s.handleTaskMessage(msg)
	})
	return err
}

// SubscribeToControl subscribes to control messages
func (s *Subscriber) SubscribeToControl(subject string, handler func(*protocol.ControlMessage) error) error {
	_, err := s.client.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Printf("[DEBUG] Received control message on: %s", subject)

		// Unmarshal control message
		ctrl, err := protocol.UnmarshalControlMessage(msg.Data)
		if err != nil {
			s.logger.Printf("[ERROR] Failed to unmarshal control message: %v", err)
			return
		}

		// Handle control message
		if err := handler(ctrl); err != nil {
			s.logger.Printf("[ERROR] Failed to handle control message: %v", err)
		}
	})
	return err
}

// Unsubscribe unsubscribes from all subjects
func (s *Subscriber) Unsubscribe() error {
	if s.subscription != nil {
		return s.subscription.Unsubscribe()
	}
	return nil
}

// handleTaskMessage handles an incoming task message
func (s *Subscriber) handleTaskMessage(msg *nats.Msg) {
	s.logger.Printf("[DEBUG] Received task message (%d bytes)", len(msg.Data))

	// Unmarshal task
	task, err := protocol.UnmarshalTask(msg.Data)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to unmarshal task: %v", err)
		return
	}

	s.logger.Printf("[INFO] Received task: %s (type: %s, module: %s)",
		task.TaskID, task.Type, task.Module)

	// Handle task
	if s.taskHandler != nil {
		if err := s.taskHandler(task); err != nil {
			s.logger.Printf("[ERROR] Failed to handle task %s: %v", task.TaskID, err)
		}
	} else {
		s.logger.Println("[WARN] No task handler registered")
	}
}
