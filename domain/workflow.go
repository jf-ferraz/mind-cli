package domain

import "time"

// WorkflowState represents the persisted state of an in-progress workflow.
type WorkflowState struct {
	Type           RequestType         `json:"type"`
	Descriptor     string              `json:"descriptor"`
	IterationPath  string              `json:"iteration_path"`
	Branch         string              `json:"branch"`
	LastAgent      string              `json:"last_agent"`
	RemainingChain []string            `json:"remaining_chain"`
	Session        int                 `json:"session"`
	TotalSessions  int                 `json:"total_sessions"`
	Artifacts      []CompletedArtifact `json:"artifacts,omitempty"`
	DispatchLog    []DispatchEntry     `json:"dispatch_log,omitempty"`
	Decisions      []string            `json:"decisions,omitempty"`
	HandoffContext string              `json:"handoff_context,omitempty"`
}

// IsIdle returns true if no workflow is in progress.
func (s *WorkflowState) IsIdle() bool {
	return s == nil || s.Type == ""
}

// CompletedArtifact records an output from a completed agent.
type CompletedArtifact struct {
	Agent    string `json:"agent"`
	Output   string `json:"output"`
	Location string `json:"location"`
}

// DispatchEntry records a single agent dispatch.
type DispatchEntry struct {
	Agent     string        `json:"agent"`
	File      string        `json:"file"`
	Model     string        `json:"model"`
	Status    string        `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration"`
}
