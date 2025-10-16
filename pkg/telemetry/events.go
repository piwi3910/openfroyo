package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a telemetry event in the OpenFroyo system.
type Event struct {
	// ID is the unique identifier for this event.
	ID string `json:"id"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Type is the event type.
	Type string `json:"type"`

	// Source identifies where the event originated.
	Source string `json:"source"`

	// RunID is the associated run ID, if applicable.
	RunID string `json:"run_id,omitempty"`

	// PlanUnitID is the associated plan unit ID, if applicable.
	PlanUnitID string `json:"plan_unit_id,omitempty"`

	// ResourceID is the associated resource ID, if applicable.
	ResourceID string `json:"resource_id,omitempty"`

	// Message is a human-readable event message.
	Message string `json:"message"`

	// Level is the event severity level (info, warning, error).
	Level string `json:"level"`

	// Data contains additional event-specific data.
	Data map[string]interface{} `json:"data,omitempty"`
}

// EventType constants for common event types.
const (
	EventTypeRunStarted          = "run.started"
	EventTypeRunCompleted        = "run.completed"
	EventTypeRunFailed           = "run.failed"
	EventTypePlanUnitStarted     = "plan_unit.started"
	EventTypePlanUnitCompleted   = "plan_unit.completed"
	EventTypePlanUnitFailed      = "plan_unit.failed"
	EventTypeResourceStateChanged = "resource.state_changed"
	EventTypeDriftDetected       = "drift.detected"
	EventTypePolicyViolation     = "policy.violation"
	EventTypeProviderInvoked     = "provider.invoked"
	EventTypeError               = "error"
)

// EventLevel constants for event severity.
const (
	EventLevelInfo    = "info"
	EventLevelWarning = "warning"
	EventLevelError   = "error"
)

// EventSubscriber is a function that handles events.
type EventSubscriber func(event Event)

// EventFilter determines if an event should be processed.
type EventFilter func(event Event) bool

// EventPublisher manages event publishing and subscriptions.
type EventPublisher struct {
	config      EventsConfig
	buffer      chan Event
	subscribers []subscriberEntry
	filters     []EventFilter
	wg          sync.WaitGroup
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

type subscriberEntry struct {
	subscriber EventSubscriber
	filter     EventFilter
}

// NewEventPublisher creates a new event publisher with the given configuration.
func NewEventPublisher(cfg EventsConfig) (*EventPublisher, error) {
	if !cfg.Enabled {
		return &EventPublisher{config: cfg}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	ep := &EventPublisher{
		config:      cfg,
		buffer:      make(chan Event, cfg.BufferSize),
		subscribers: make([]subscriberEntry, 0),
		filters:     make([]EventFilter, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start the event processing goroutine
	if cfg.EnableAsync {
		ep.wg.Add(1)
		go ep.processEvents()
	}

	// Start the periodic flush goroutine
	if cfg.FlushInterval > 0 {
		ep.wg.Add(1)
		go ep.periodicFlush()
	}

	return ep, nil
}

// Publish publishes an event to all subscribers.
func (ep *EventPublisher) Publish(event Event) error {
	if !ep.config.Enabled {
		return nil
	}

	// Set ID and timestamp if not already set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Apply global filters
	ep.mu.RLock()
	for _, filter := range ep.filters {
		if !filter(event) {
			ep.mu.RUnlock()
			return nil // Event filtered out
		}
	}
	ep.mu.RUnlock()

	// Send to buffer if async, otherwise process immediately
	if ep.config.EnableAsync {
		select {
		case ep.buffer <- event:
			return nil
		case <-ep.ctx.Done():
			return fmt.Errorf("event publisher stopped")
		default:
			// Buffer full, drop event or log warning
			return fmt.Errorf("event buffer full, event dropped")
		}
	}

	// Synchronous publishing
	ep.deliverEvent(event)
	return nil
}

// PublishRunStarted publishes a run started event.
func (ep *EventPublisher) PublishRunStarted(runID, user string) error {
	return ep.Publish(Event{
		Type:    EventTypeRunStarted,
		Source:  "engine",
		RunID:   runID,
		Message: fmt.Sprintf("Run %s started by %s", runID, user),
		Level:   EventLevelInfo,
		Data: map[string]interface{}{
			"user": user,
		},
	})
}

// PublishRunCompleted publishes a run completed event.
func (ep *EventPublisher) PublishRunCompleted(runID, status string, duration time.Duration) error {
	return ep.Publish(Event{
		Type:    EventTypeRunCompleted,
		Source:  "engine",
		RunID:   runID,
		Message: fmt.Sprintf("Run %s completed with status: %s", runID, status),
		Level:   EventLevelInfo,
		Data: map[string]interface{}{
			"status":   status,
			"duration": duration.Seconds(),
		},
	})
}

// PublishRunFailed publishes a run failed event.
func (ep *EventPublisher) PublishRunFailed(runID, reason string) error {
	return ep.Publish(Event{
		Type:    EventTypeRunFailed,
		Source:  "engine",
		RunID:   runID,
		Message: fmt.Sprintf("Run %s failed: %s", runID, reason),
		Level:   EventLevelError,
		Data: map[string]interface{}{
			"reason": reason,
		},
	})
}

// PublishPlanUnitStarted publishes a plan unit started event.
func (ep *EventPublisher) PublishPlanUnitStarted(runID, planUnitID, resourceID, operation string) error {
	return ep.Publish(Event{
		Type:       EventTypePlanUnitStarted,
		Source:     "engine",
		RunID:      runID,
		PlanUnitID: planUnitID,
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Plan unit %s started: %s on resource %s", planUnitID, operation, resourceID),
		Level:      EventLevelInfo,
		Data: map[string]interface{}{
			"operation": operation,
		},
	})
}

// PublishPlanUnitCompleted publishes a plan unit completed event.
func (ep *EventPublisher) PublishPlanUnitCompleted(runID, planUnitID, resourceID string, duration time.Duration) error {
	return ep.Publish(Event{
		Type:       EventTypePlanUnitCompleted,
		Source:     "engine",
		RunID:      runID,
		PlanUnitID: planUnitID,
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Plan unit %s completed for resource %s", planUnitID, resourceID),
		Level:      EventLevelInfo,
		Data: map[string]interface{}{
			"duration": duration.Seconds(),
		},
	})
}

// PublishPlanUnitFailed publishes a plan unit failed event.
func (ep *EventPublisher) PublishPlanUnitFailed(runID, planUnitID, resourceID, reason string) error {
	return ep.Publish(Event{
		Type:       EventTypePlanUnitFailed,
		Source:     "engine",
		RunID:      runID,
		PlanUnitID: planUnitID,
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Plan unit %s failed for resource %s: %s", planUnitID, resourceID, reason),
		Level:      EventLevelError,
		Data: map[string]interface{}{
			"reason": reason,
		},
	})
}

// PublishResourceStateChanged publishes a resource state change event.
func (ep *EventPublisher) PublishResourceStateChanged(resourceID, oldState, newState string) error {
	return ep.Publish(Event{
		Type:       EventTypeResourceStateChanged,
		Source:     "engine",
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Resource %s state changed from %s to %s", resourceID, oldState, newState),
		Level:      EventLevelInfo,
		Data: map[string]interface{}{
			"old_state": oldState,
			"new_state": newState,
		},
	})
}

// PublishDriftDetected publishes a drift detected event.
func (ep *EventPublisher) PublishDriftDetected(resourceID string, driftCount int) error {
	return ep.Publish(Event{
		Type:       EventTypeDriftDetected,
		Source:     "drift_detector",
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Drift detected on resource %s (%d changes)", resourceID, driftCount),
		Level:      EventLevelWarning,
		Data: map[string]interface{}{
			"drift_count": driftCount,
		},
	})
}

// PublishPolicyViolation publishes a policy violation event.
func (ep *EventPublisher) PublishPolicyViolation(resourceID, policyName, reason string) error {
	return ep.Publish(Event{
		Type:       EventTypePolicyViolation,
		Source:     "policy_engine",
		ResourceID: resourceID,
		Message:    fmt.Sprintf("Policy violation on resource %s: %s - %s", resourceID, policyName, reason),
		Level:      EventLevelError,
		Data: map[string]interface{}{
			"policy": policyName,
			"reason": reason,
		},
	})
}

// Subscribe adds a new event subscriber.
func (ep *EventPublisher) Subscribe(subscriber EventSubscriber, filter EventFilter) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.subscribers = append(ep.subscribers, subscriberEntry{
		subscriber: subscriber,
		filter:     filter,
	})
}

// AddFilter adds a global event filter.
func (ep *EventPublisher) AddFilter(filter EventFilter) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.filters = append(ep.filters, filter)
}

// processEvents processes events from the buffer asynchronously.
func (ep *EventPublisher) processEvents() {
	defer ep.wg.Done()

	batch := make([]Event, 0, ep.config.MaxBatchSize)

	for {
		select {
		case event := <-ep.buffer:
			batch = append(batch, event)

			// Flush batch if it reaches max size
			if len(batch) >= ep.config.MaxBatchSize {
				ep.flushBatch(batch)
				batch = make([]Event, 0, ep.config.MaxBatchSize)
			}

		case <-ep.ctx.Done():
			// Flush remaining events before shutting down
			if len(batch) > 0 {
				ep.flushBatch(batch)
			}
			return
		}
	}
}

// periodicFlush flushes events periodically.
func (ep *EventPublisher) periodicFlush() {
	defer ep.wg.Done()

	ticker := time.NewTicker(ep.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Trigger flush by draining buffer
			// This is handled by the processEvents goroutine
		case <-ep.ctx.Done():
			return
		}
	}
}

// flushBatch delivers a batch of events to subscribers.
func (ep *EventPublisher) flushBatch(events []Event) {
	for _, event := range events {
		ep.deliverEvent(event)
	}
}

// deliverEvent delivers an event to all subscribers.
func (ep *EventPublisher) deliverEvent(event Event) {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	for _, entry := range ep.subscribers {
		// Apply subscriber-specific filter
		if entry.filter != nil && !entry.filter(event) {
			continue
		}

		// Call subscriber in a goroutine to avoid blocking
		go entry.subscriber(event)
	}
}

// Shutdown gracefully shuts down the event publisher.
func (ep *EventPublisher) Shutdown(ctx context.Context) error {
	if !ep.config.Enabled {
		return nil
	}

	// Signal shutdown
	ep.cancel()

	// Wait for processing to complete with timeout
	done := make(chan struct{})
	go func() {
		ep.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("event publisher shutdown timeout")
	}
}

// Common event filters.

// FilterByLevel creates a filter that only allows events of a specific level or higher.
func FilterByLevel(minLevel string) EventFilter {
	levels := map[string]int{
		EventLevelInfo:    0,
		EventLevelWarning: 1,
		EventLevelError:   2,
	}

	minLevelValue := levels[minLevel]

	return func(event Event) bool {
		return levels[event.Level] >= minLevelValue
	}
}

// FilterByType creates a filter that only allows events of specific types.
func FilterByType(types ...string) EventFilter {
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	return func(event Event) bool {
		return typeSet[event.Type]
	}
}

// FilterByRunID creates a filter that only allows events for a specific run.
func FilterByRunID(runID string) EventFilter {
	return func(event Event) bool {
		return event.RunID == runID
	}
}

// FilterByResourceID creates a filter that only allows events for a specific resource.
func FilterByResourceID(resourceID string) EventFilter {
	return func(event Event) bool {
		return event.ResourceID == resourceID
	}
}
