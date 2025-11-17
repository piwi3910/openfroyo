package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/piwi3910/openfroyo/internal/agent/config"
	"github.com/piwi3910/openfroyo/internal/protocol"
)

// StackExecutor is a function that executes a stack
type StackExecutor func(stackName string, vars map[string]interface{}) error

// Scheduler manages periodic stack execution
type Scheduler struct {
	config    *config.Config
	logger    *log.Logger
	executor  StackExecutor
	schedules map[string]*ScheduleState
	mutex     sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// ScheduleState tracks the state of a scheduled stack
type ScheduleState struct {
	Schedule  *protocol.PullSchedule
	LastRun   *time.Time
	NextRun   *time.Time
	RunCount  int64
	Failures  int64
	timer     *time.Timer
	stopChan  chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config, logger *log.Logger, executor StackExecutor) *Scheduler {
	return &Scheduler{
		config:    cfg,
		logger:    logger,
		executor:  executor,
		schedules: make(map[string]*ScheduleState),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	if !s.config.Pull.Enabled {
		s.logger.Println("[INFO] Pull mode disabled, scheduler not started")
		return nil
	}

	s.logger.Println("[INFO] Starting pull mode scheduler")

	// Load schedules from config
	for _, stackCfg := range s.config.Pull.Stacks {
		if !stackCfg.Enabled {
			continue
		}

		schedule := &protocol.PullSchedule{
			StackName: stackCfg.Name,
			Interval:  stackCfg.Interval,
			Enabled:   stackCfg.Enabled,
			Vars:      make(map[string]interface{}),
		}

		if err := s.AddSchedule(schedule); err != nil {
			s.logger.Printf("[ERROR] Failed to add schedule for %s: %v", stackCfg.Name, err)
		}
	}

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Println("[INFO] Stopping scheduler")

	close(s.stopChan)

	// Stop all schedules
	s.mutex.Lock()
	for _, state := range s.schedules {
		if state.timer != nil {
			state.timer.Stop()
		}
		close(state.stopChan)
	}
	s.mutex.Unlock()

	s.wg.Wait()
	s.logger.Println("[INFO] Scheduler stopped")
}

// AddSchedule adds a new schedule
func (s *Scheduler) AddSchedule(schedule *protocol.PullSchedule) error {
	if schedule.Interval <= 0 {
		return fmt.Errorf("interval must be > 0")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.schedules[schedule.StackName]; exists {
		return fmt.Errorf("schedule already exists for stack: %s", schedule.StackName)
	}

	now := time.Now()
	nextRun := now.Add(time.Duration(schedule.Interval) * time.Second)

	state := &ScheduleState{
		Schedule: schedule,
		NextRun:  &nextRun,
		stopChan: make(chan struct{}),
	}

	s.schedules[schedule.StackName] = state

	// Start schedule timer
	s.wg.Add(1)
	go s.runSchedule(state)

	s.logger.Printf("[INFO] Added schedule: %s (interval: %ds, next run: %s)",
		schedule.StackName, schedule.Interval, nextRun.Format(time.RFC3339))

	return nil
}

// RemoveSchedule removes a schedule
func (s *Scheduler) RemoveSchedule(stackName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, exists := s.schedules[stackName]
	if !exists {
		return fmt.Errorf("schedule not found: %s", stackName)
	}

	if state.timer != nil {
		state.timer.Stop()
	}
	close(state.stopChan)

	delete(s.schedules, stackName)

	s.logger.Printf("[INFO] Removed schedule: %s", stackName)
	return nil
}

// GetSchedules returns all current schedules
func (s *Scheduler) GetSchedules() []*ScheduleState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	schedules := make([]*ScheduleState, 0, len(s.schedules))
	for _, state := range s.schedules {
		schedules = append(schedules, state)
	}
	return schedules
}

// runSchedule runs a schedule in a loop
func (s *Scheduler) runSchedule(state *ScheduleState) {
	defer s.wg.Done()

	interval := time.Duration(state.Schedule.Interval) * time.Second
	state.timer = time.NewTimer(interval)

	for {
		select {
		case <-state.timer.C:
			// Execute stack
			s.executeStack(state)

			// Schedule next run
			state.timer.Reset(interval)

		case <-state.stopChan:
			return

		case <-s.stopChan:
			return
		}
	}
}

// executeStack executes a scheduled stack
func (s *Scheduler) executeStack(state *ScheduleState) {
	now := time.Now()
	state.LastRun = &now
	state.RunCount++

	s.logger.Printf("[INFO] Executing scheduled stack: %s (run #%d)",
		state.Schedule.StackName, state.RunCount)

	// Execute stack
	if s.executor != nil {
		if err := s.executor(state.Schedule.StackName, state.Schedule.Vars); err != nil {
			s.logger.Printf("[ERROR] Failed to execute stack %s: %v",
				state.Schedule.StackName, err)
			state.Failures++
		} else {
			s.logger.Printf("[INFO] Stack execution completed: %s", state.Schedule.StackName)
		}
	} else {
		s.logger.Println("[WARN] No executor configured")
	}

	// Update next run
	nextRun := now.Add(time.Duration(state.Schedule.Interval) * time.Second)
	state.NextRun = &nextRun
}
