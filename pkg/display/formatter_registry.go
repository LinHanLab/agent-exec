package display

import (
	"github.com/LinHanLab/agent-exec/pkg/events"
)

// EventFormatter is a function that formats a specific event type
type EventFormatter func(event events.Event, ctx *FormatContext) (string, error)

// FormatContext provides dependencies to event formatters
type FormatContext struct {
	TextFormatter TextFormatter
	ContentFilter *ContentFilter
	Verbose       bool
}

// eventFormatters maps event types to their formatter functions
var eventFormatters = map[events.EventType]EventFormatter{
	events.EventRunPromptStarted:       formatRunPromptStarted,
	events.EventClaudeAssistantMessage: formatClaudeAssistantMessage,
	events.EventClaudeToolUse:          formatClaudeToolUse,
	events.EventClaudeToolResult:       formatClaudeToolResult,
	events.EventClaudeExecutionResult:  formatClaudeExecutionResult,
	events.EventLoopStarted:            formatLoopStarted,
	events.EventEvolveStarted:          formatEvolveStarted,
	events.EventRoundStarted:           formatRoundStarted,
	events.EventIterationStarted:       formatIterationStarted,
	events.EventIterationCompleted:     formatIterationCompleted,
	events.EventIterationFailed:        formatIterationFailed,
	events.EventLoopCompleted:          formatLoopCompleted,
	events.EventLoopInterrupted:        formatLoopInterrupted,
	events.EventSleepStarted:           formatSleepStarted,
	events.EventImprovementStarted:     formatImprovementStarted,
	events.EventComparisonStarted:      formatComparisonStarted,
	events.EventComparisonRetry:        formatComparisonRetry,
	events.EventWinnerSelected:         formatWinnerSelected,
	events.EventEvolveCompleted:        formatEvolveCompleted,
	events.EventEvolveInterrupted:      formatEvolveInterrupted,
	events.EventGitBranchCreated:       formatGitBranchCreated,
	events.EventGitBranchCheckedOut:    formatGitBranchCheckedOut,
	events.EventGitBranchDeleted:       formatGitBranchDeleted,
	events.EventGitCommitsSquashed:     formatGitCommitsSquashed,
}

// GetColorForEventType returns the ANSI color code for an event type
func GetColorForEventType(eventType events.EventType) string {
	switch eventType {
	case events.EventRunPromptStarted:
		return BoldCyan

	case events.EventLoopStarted,
		events.EventIterationStarted,
		events.EventEvolveStarted,
		events.EventRoundStarted,
		events.EventImprovementStarted,
		events.EventComparisonStarted,
		events.EventSleepStarted:
		return BoldYellow

	case events.EventClaudeExecutionResult,
		events.EventLoopCompleted,
		events.EventEvolveCompleted,
		events.EventIterationCompleted,
		events.EventWinnerSelected:
		return BoldGreen

	case events.EventIterationFailed,
		events.EventLoopInterrupted,
		events.EventEvolveInterrupted:
		return BoldRed

	case events.EventClaudeAssistantMessage,
		events.EventComparisonRetry,
		events.EventGitBranchCreated,
		events.EventGitBranchCheckedOut,
		events.EventGitBranchDeleted,
		events.EventGitCommitsSquashed:
		return Magenta

	case events.EventClaudeToolUse,
		events.EventClaudeToolResult:
		return ""

	default:
		return ""
	}
}
