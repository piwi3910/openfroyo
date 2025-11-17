package protocol

import "fmt"

const (
	// Subject prefix for all OpenFroyo messages
	SubjectPrefix = "openfroyo"

	// Subject components
	SubjectAgents    = "agents"
	SubjectBroadcast = "broadcast"
	SubjectControl   = "control"

	// Subject suffixes
	SuffixTasks        = "tasks"
	SuffixResults      = "results"
	SuffixHealth       = "health"
	SuffixRegistration = "registration"
)

// SubjectBuilder helps construct NATS subjects
type SubjectBuilder struct {
	datacenter string
	agentID    string
}

// NewSubjectBuilder creates a new subject builder
func NewSubjectBuilder(datacenter, agentID string) *SubjectBuilder {
	return &SubjectBuilder{
		datacenter: datacenter,
		agentID:    agentID,
	}
}

// AgentTasks returns the subject for tasks directed to a specific agent
// Format: openfroyo.{dc}.agents.{id}.tasks
func (sb *SubjectBuilder) AgentTasks() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, sb.agentID, SuffixTasks)
}

// AgentResults returns the subject for results from a specific agent
// Format: openfroyo.{dc}.agents.{id}.results
func (sb *SubjectBuilder) AgentResults() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, sb.agentID, SuffixResults)
}

// AgentHealth returns the subject for health status from a specific agent
// Format: openfroyo.{dc}.agents.{id}.health
func (sb *SubjectBuilder) AgentHealth() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, sb.agentID, SuffixHealth)
}

// AgentRegistration returns the subject for agent registration
// Format: openfroyo.{dc}.agents.{id}.registration
func (sb *SubjectBuilder) AgentRegistration() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, sb.agentID, SuffixRegistration)
}

// DatacenterAgentTasks returns the wildcard subject for all agents in a datacenter
// Format: openfroyo.{dc}.agents.*.tasks
func (sb *SubjectBuilder) DatacenterAgentTasks() string {
	return fmt.Sprintf("%s.%s.%s.*.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, SuffixTasks)
}

// DatacenterAgentResults returns the wildcard subject for all agent results in a datacenter
// Format: openfroyo.{dc}.agents.*.results
func (sb *SubjectBuilder) DatacenterAgentResults() string {
	return fmt.Sprintf("%s.%s.%s.*.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, SuffixResults)
}

// DatacenterAgentHealth returns the wildcard subject for all agent health in a datacenter
// Format: openfroyo.{dc}.agents.*.health
func (sb *SubjectBuilder) DatacenterAgentHealth() string {
	return fmt.Sprintf("%s.%s.%s.*.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, SuffixHealth)
}

// BroadcastTasks returns the broadcast subject for tasks to all agents
// Format: openfroyo.broadcast.tasks
func (sb *SubjectBuilder) BroadcastTasks() string {
	return fmt.Sprintf("%s.%s.%s", SubjectPrefix, SubjectBroadcast, SuffixTasks)
}

// AgentControl returns the subject for control messages to a specific agent
// Format: openfroyo.{dc}.agents.{id}.control
func (sb *SubjectBuilder) AgentControl() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, sb.agentID, SubjectControl)
}

// DatacenterControl returns the subject for control messages to all agents in a datacenter
// Format: openfroyo.{dc}.agents.*.control
func (sb *SubjectBuilder) DatacenterControl() string {
	return fmt.Sprintf("%s.%s.%s.*.%s",
		SubjectPrefix, sb.datacenter, SubjectAgents, SubjectControl)
}

// Helper functions for common patterns

// AgentTasksSubject returns the task subject for a specific agent
func AgentTasksSubject(datacenter, agentID string) string {
	return NewSubjectBuilder(datacenter, agentID).AgentTasks()
}

// AgentResultsSubject returns the results subject for a specific agent
func AgentResultsSubject(datacenter, agentID string) string {
	return NewSubjectBuilder(datacenter, agentID).AgentResults()
}

// AgentHealthSubject returns the health subject for a specific agent
func AgentHealthSubject(datacenter, agentID string) string {
	return NewSubjectBuilder(datacenter, agentID).AgentHealth()
}

// DatacenterTasksSubject returns the wildcard subject for all agents in a datacenter
func DatacenterTasksSubject(datacenter string) string {
	return NewSubjectBuilder(datacenter, "*").DatacenterAgentTasks()
}
