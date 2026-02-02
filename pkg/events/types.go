package events

import "time"

// EventType represents the type of event being emitted
type EventType string

// Event type constants
const (
	// Prompt execution events
	EventPromptStarted EventType = "prompt_started"

	// Claude streaming events
	EventClaudeAssistantMessage EventType = "claude_assistant_message"
	EventClaudeToolUse          EventType = "claude_tool_use"
	EventClaudeToolResult       EventType = "claude_tool_result"
	EventClaudeExecutionResult  EventType = "claude_execution_result"

	// Git operation events
	EventGitBranchCreated    EventType = "git_branch_created"
	EventGitBranchCheckedOut EventType = "git_branch_checked_out"
	EventGitBranchDeleted    EventType = "git_branch_deleted"
	EventGitCommitsSquashed  EventType = "git_commits_squashed"

	// Loop execution events
	EventLoopStarted        EventType = "loop_started"
	EventIterationStarted   EventType = "iteration_started"
	EventIterationCompleted EventType = "iteration_completed"
	EventIterationFailed    EventType = "iteration_failed"
	EventLoopCompleted      EventType = "loop_completed"
	EventLoopInterrupted    EventType = "loop_interrupted"

	// Evolution workflow events
	EventEvolveStarted      EventType = "evolve_started"
	EventRoundStarted       EventType = "round_started"
	EventImprovementStarted EventType = "improvement_started"
	EventComparisonStarted  EventType = "comparison_started"
	EventComparisonRetry    EventType = "comparison_retry"
	EventWinnerSelected     EventType = "winner_selected"
	EventEvolveCompleted    EventType = "evolve_completed"
	EventEvolveInterrupted  EventType = "evolve_interrupted"

	EventSleepStarted EventType = "sleep_started"
)

// Event represents a single event in the system
type Event struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// PromptStartedData contains data for EventPromptStarted
type PromptStartedData struct {
	Prompt   string
	BaseURL  string
	Cwd      string
	FileList string
}

// AssistantMessageData contains data for EventAssistantMessage
type AssistantMessageData struct {
	Text string
}

// ToolUseData contains data for EventToolUse
type ToolUseData struct {
	Name  string
	Input map[string]interface{}
}

// ToolResultData contains data for EventToolResult
type ToolResultData struct {
	Content string
}

// ExecutionResultData contains data for EventExecutionResult
type ExecutionResultData struct {
	Duration time.Duration
}

// LoopStartedData contains data for EventLoopStarted
type LoopStartedData struct {
	TotalIterations int
}

// IterationStartedData contains data for EventIterationStarted
type IterationStartedData struct {
	Current int
	Total   int
}

// IterationCompletedData contains data for EventIterationCompleted
type IterationCompletedData struct {
	Current  int
	Total    int
	Duration time.Duration
}

// IterationFailedData contains data for EventIterationFailed
type IterationFailedData struct {
	Current int
	Total   int
	Error   error
}

// SleepStartedData contains data for EventSleepStarted
type SleepStartedData struct {
	Duration time.Duration
}

// LoopCompletedData contains data for EventLoopCompleted
type LoopCompletedData struct {
	TotalIterations      int
	SuccessfulIterations int
	FailedIterations     int
	TotalDuration        time.Duration
}

// LoopInterruptedData contains data for EventLoopInterrupted
type LoopInterruptedData struct {
	CompletedIterations int
	TotalIterations     int
}

// EvolveStartedData contains data for EventEvolveStarted
type EvolveStartedData struct {
	Prompt     string
	Iterations int
}

// BranchCreatedData contains data for EventBranchCreated
type BranchCreatedData struct {
	BranchName string
	Base       string // Optional: base branch for CreateBranchFrom
}

// BranchCheckedOutData contains data for EventBranchCheckedOut
type BranchCheckedOutData struct {
	BranchName string
}

// BranchDeletedData contains data for EventBranchDeleted
type BranchDeletedData struct {
	BranchName string
}

// CommitsSquashedData contains data for EventCommitsSquashed
type CommitsSquashedData struct {
	BranchName string
}

// RoundStartedData contains data for EventRoundStarted
type RoundStartedData struct {
	Round int
	Total int
}

// ImprovementStartedData contains data for EventImprovementStarted
type ImprovementStartedData struct {
	BranchName string
}

// ComparisonStartedData contains data for EventComparisonStarted
type ComparisonStartedData struct {
	Branch1 string
	Branch2 string
}

// ComparisonRetryData contains data for EventComparisonRetry
type ComparisonRetryData struct {
	Attempt     int
	MaxAttempts int
}

// WinnerSelectedData contains data for EventWinnerSelected
type WinnerSelectedData struct {
	Winner string
	Loser  string
}

// EvolveCompletedData contains data for EventEvolveCompleted
type EvolveCompletedData struct {
	FinalBranch   string
	TotalRounds   int
	TotalDuration time.Duration
}

// EvolveInterruptedData contains data for EventEvolveInterrupted
type EvolveInterruptedData struct {
	CompletedRounds int
	TotalRounds     int
	Winner          string
}
