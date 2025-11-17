package agent

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piwi3910/openfroyo/internal/agent/config"
	"github.com/piwi3910/openfroyo/internal/agent/executor"
	"github.com/piwi3910/openfroyo/internal/agent/health"
	agentNats "github.com/piwi3910/openfroyo/internal/agent/nats"
	"github.com/piwi3910/openfroyo/internal/agent/scheduler"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// Agent represents the OpenFroyo agent
type Agent struct {
	config       *config.Config
	logger       *log.Logger
	natsClient   *agentNats.Client
	subscriber   *agentNats.Subscriber
	publisher    *agentNats.Publisher
	executor     *executor.Executor
	scheduler    *scheduler.Scheduler
	healthMon    *health.Monitor
	healthRep    *health.Reporter
	subjectBuilder *protocol.SubjectBuilder
	stopChan     chan struct{}
}

// New creates a new agent
func New(cfg *config.Config, logger *log.Logger) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Create NATS client
	natsClient, err := agentNats.NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	// Create subscriber and publisher
	subscriber := agentNats.NewSubscriber(natsClient, logger)
	publisher := agentNats.NewPublisher(natsClient, logger)

	// Create executor
	exec, err := executor.NewExecutor(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create health monitor
	healthMon := health.NewMonitor(cfg, logger)

	// Create subject builder
	subjectBuilder := protocol.NewSubjectBuilder(cfg.Agent.Datacenter, cfg.Agent.ID)

	// Create scheduler (executor function will be set later)
	sched := scheduler.NewScheduler(cfg, logger, nil)

	// Create health reporter
	healthRep := health.NewReporter(
		healthMon,
		publisher,
		subjectBuilder.AgentHealth(),
		30*time.Second, // Report every 30 seconds
		logger,
	)

	return &Agent{
		config:         cfg,
		logger:         logger,
		natsClient:     natsClient,
		subscriber:     subscriber,
		publisher:      publisher,
		executor:       exec,
		scheduler:      sched,
		healthMon:      healthMon,
		healthRep:      healthRep,
		subjectBuilder: subjectBuilder,
		stopChan:       make(chan struct{}),
	}, nil
}

// Run starts the agent and blocks until shutdown
func (a *Agent) Run() error {
	a.logger.Println("[INFO] Starting OpenFroyo Agent")
	a.logger.Printf("[INFO] Agent ID: %s, Datacenter: %s", a.config.Agent.ID, a.config.Agent.Datacenter)

	// Connect to NATS
	if err := a.natsClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer a.natsClient.Close()

	// Publish agent registration
	if err := a.publishRegistration(); err != nil {
		a.logger.Printf("[WARN] Failed to publish registration: %v", err)
	}

	// Subscribe to task messages
	taskSubject := a.subjectBuilder.AgentTasks()
	if err := a.subscriber.SubscribeToTasks(taskSubject, a.handleTask); err != nil {
		return fmt.Errorf("failed to subscribe to tasks: %w", err)
	}

	// Subscribe to broadcast tasks
	broadcastSubject := a.subjectBuilder.BroadcastTasks()
	if err := a.subscriber.SubscribeToBroadcast(broadcastSubject, a.handleTask); err != nil {
		a.logger.Printf("[WARN] Failed to subscribe to broadcasts: %v", err)
	}

	// Subscribe to control messages
	controlSubject := a.subjectBuilder.AgentControl()
	if err := a.subscriber.SubscribeToControl(controlSubject, a.handleControl); err != nil {
		a.logger.Printf("[WARN] Failed to subscribe to control: %v", err)
	}

	// Start health reporter
	a.healthRep.Start()
	defer a.healthRep.Stop()

	// Start scheduler
	if err := a.scheduler.Start(); err != nil {
		a.logger.Printf("[WARN] Failed to start scheduler: %v", err)
	}
	defer a.scheduler.Stop()

	a.logger.Println("[INFO] Agent is running")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		a.logger.Printf("[INFO] Received signal: %v, shutting down...", sig)
	case <-a.stopChan:
		a.logger.Println("[INFO] Shutdown requested")
	}

	a.logger.Println("[INFO] Agent stopped")
	return nil
}

// Stop stops the agent
func (a *Agent) Stop() {
	close(a.stopChan)
}

// handleTask handles an incoming task
func (a *Agent) handleTask(task *protocol.TaskMessage) error {
	a.logger.Printf("[INFO] Handling task: %s", task.TaskID)

	// Execute task
	result, err := a.executor.Execute(task)
	if err != nil {
		a.logger.Printf("[ERROR] Task execution failed: %v", err)
		// Create error result
		result = &protocol.TaskResult{
			TaskID:    task.TaskID,
			AgentID:   a.config.Agent.ID,
			Status:    protocol.TaskStatusFailed,
			Message:   "Execution failed",
			Error:     err.Error(),
			Facts:     make(map[string]interface{}),
			StartedAt: time.Now(),
			CompletedAt: time.Now(),
		}
	}

	// Update health stats
	a.healthMon.UpdateTaskStats(result.Status)

	// Publish result
	resultSubject := a.subjectBuilder.AgentResults()
	if err := a.publisher.PublishTaskResult(resultSubject, result); err != nil {
		a.logger.Printf("[ERROR] Failed to publish task result: %v", err)
		return err
	}

	return nil
}

// handleControl handles control messages
func (a *Agent) handleControl(ctrl *protocol.ControlMessage) error {
	a.logger.Printf("[INFO] Received control message: %s", ctrl.Type)

	switch ctrl.Type {
	case protocol.ControlTypeShutdown:
		a.logger.Println("[INFO] Shutdown requested via control message")
		go func() {
			time.Sleep(1 * time.Second) // Give time to finish current tasks
			a.Stop()
		}()

	case protocol.ControlTypeReload:
		a.logger.Println("[INFO] Reload requested (not yet implemented)")
		// TODO: Reload configuration

	case protocol.ControlTypeUpdateSchedule:
		a.logger.Println("[INFO] Schedule update requested (not yet implemented)")
		// TODO: Update schedules

	default:
		a.logger.Printf("[WARN] Unknown control message type: %s", ctrl.Type)
	}

	return nil
}

// publishRegistration publishes agent registration/announcement
func (a *Agent) publishRegistration() error {
	hostname, _ := os.Hostname()

	tags := make(map[string]string)
	for _, tag := range a.config.Agent.Tags {
		tags[tag] = "true"
	}

	reg := &protocol.AgentRegistration{
		AgentID:    a.config.Agent.ID,
		Hostname:   hostname,
		Datacenter: a.config.Agent.Datacenter,
		Version:    a.config.Agent.Version,
		Tags:       tags,
		Timestamp:  time.Now(),
	}

	subject := a.subjectBuilder.AgentRegistration()
	return a.publisher.PublishAgentRegistration(subject, reg)
}
