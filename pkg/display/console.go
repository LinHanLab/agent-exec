package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// JSONFormatter formats events as human-readable text with color highlighting
type JSONFormatter struct {
	writer io.Writer
}

// NewConsoleFormatter creates a new JSONFormatter
func NewConsoleFormatter(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
	}
}

// formatBigTitle wraps title with ==== lines and empty lines
func (f *JSONFormatter) formatBigTitle(title string) string {
	separator := strings.Repeat("=", len(title))
	return fmt.Sprintf("\n%s\n%s\n%s\n", separator, title, separator)
}

// formatSmallTitle wraps title with --- lines
func (f *JSONFormatter) formatSmallTitle(title string) string {
	separator := strings.Repeat("-", len(title))
	return fmt.Sprintf("%s\n%s\n%s", separator, title, separator)
}

// formatCodeBlock wraps content in ``` with optional language
func (f *JSONFormatter) formatCodeBlock(content string, language string) string {
	if language != "" {
		return fmt.Sprintf("\n```%s\n%s\n```\n", language, content)
	}
	return fmt.Sprintf("\n```\n%s\n```\n", content)
}

// formatPrettyJSON marshals data to indented JSON
func (f *JSONFormatter) formatPrettyJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// formatDuration formats duration in human-readable format
func (f *JSONFormatter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// Format processes an event and outputs it as human-readable colored text
func (f *JSONFormatter) Format(event events.Event) error {
	var output string
	timeStr := fmt.Sprintf("[%s]", f.formatTime())

	switch event.Type {
	// Small title event
	case events.EventRunPromptStarted:
		data := event.Data.(events.RunPromptStartedData)
		title := "🚀 Run Prompt Started"
		output = f.formatSmallTitle(title) + "\n" + f.formatCodeBlock(data.Prompt, "")

		// Add optional metadata
		if data.Cwd != "" {
			output += fmt.Sprintf("\n📁 Working Directory: %s", data.Cwd)
		}
		if data.BaseURL != "" {
			output += fmt.Sprintf("\n🌐 Base URL: %s", data.BaseURL)
		}
		if data.FileList != "" {
			output += fmt.Sprintf("\n📄 File List: %s", data.FileList)
		}

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		output = fmt.Sprintf("💬 %s %s", timeStr, data.Text)

	case events.EventClaudeToolUse:
		data := event.Data.(events.ToolUseData)
		prettyJSON, err := f.formatPrettyJSON(data.Input)
		if err != nil {
			return fmt.Errorf("failed to format tool input: %w", err)
		}
		output = fmt.Sprintf("🔧 %s Tool: %s\n%s", timeStr, data.Name, f.formatCodeBlock(prettyJSON, "json"))

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		output = fmt.Sprintf("📋 %s Tool Result\n%s", timeStr, f.formatCodeBlock(data.Content, ""))

	case events.EventClaudeExecutionResult:
		data := event.Data.(events.ExecutionResultData)
		output = fmt.Sprintf("⏱️ Execution completed in %s", f.formatDuration(data.Duration))

	// Big title events
	case events.EventLoopStarted:
		data := event.Data.(events.LoopStartedData)
		title := "🔄 Loop Started"
		output = f.formatBigTitle(title) + fmt.Sprintf("\n🔢 Iterations: %d", data.TotalIterations)

	case events.EventEvolveStarted:
		data := event.Data.(events.EvolveStartedData)
		title := "🧬 Evolution Started"
		output = f.formatBigTitle(title) + fmt.Sprintf("\n🔢 Iterations: %d", data.TotalIterations)

	case events.EventRoundStarted:
		data := event.Data.(events.RoundStartedData)
		title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)
		output = f.formatBigTitle(title)

	// Loop execution events
	case events.EventIterationStarted:
		data := event.Data.(events.IterationStartedData)
		output = fmt.Sprintf("▶️ %s Iteration %d/%d started", timeStr, data.Current, data.Total)

	case events.EventIterationCompleted:
		data := event.Data.(events.IterationCompletedData)
		output = fmt.Sprintf("✅ %s Iteration %d/%d completed in %s", timeStr, data.Current, data.Total, f.formatDuration(data.Duration))

	case events.EventIterationFailed:
		data := event.Data.(events.IterationFailedData)
		errMsg := "unknown error"
		if data.Error != nil {
			errMsg = data.Error.Error()
		}
		output = fmt.Sprintf("❌ %s Iteration %d/%d failed: %s", timeStr, data.Current, data.Total, errMsg)

	case events.EventLoopCompleted:
		data := event.Data.(events.LoopCompletedData)
		output = fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
			data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, f.formatDuration(data.TotalDuration))

	case events.EventLoopInterrupted:
		data := event.Data.(events.LoopInterruptedData)
		output = fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)

	case events.EventSleepStarted:
		data := event.Data.(events.SleepStartedData)
		output = fmt.Sprintf("💤 %s Sleeping for %s", timeStr, f.formatDuration(data.Duration))

	// Evolution events
	case events.EventImprovementStarted:
		data := event.Data.(events.ImprovementStartedData)
		output = fmt.Sprintf("🔨 %s Improving branch: %s", timeStr, data.BranchName)

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		output = fmt.Sprintf("⚖️ %s Comparing: %s vs %s", timeStr, data.Branch1, data.Branch2)

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		output = fmt.Sprintf("🔁 %s Comparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		output = fmt.Sprintf("🏆 %s Winner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)

	case events.EventEvolveCompleted:
		data := event.Data.(events.EvolveCompletedData)
		output = fmt.Sprintf("🎉 Evolution completed: %s (Total: %s)", data.FinalBranch, f.formatDuration(data.TotalDuration))

	case events.EventEvolveInterrupted:
		data := event.Data.(events.EvolveInterruptedData)
		output = fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)

	// Git operations
	case events.EventGitBranchCreated:
		data := event.Data.(events.BranchCreatedData)
		output = fmt.Sprintf("🌿 %s Branch created: %s", timeStr, data.BranchName)
		if data.Base != "" {
			output += fmt.Sprintf(" (from %s)", data.Base)
		}

	case events.EventGitBranchCheckedOut:
		data := event.Data.(events.BranchCheckedOutData)
		output = fmt.Sprintf("🔀 %s Checked out branch: %s", timeStr, data.BranchName)

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		output = fmt.Sprintf("🗑️ %s Branch deleted: %s", timeStr, data.BranchName)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		output = fmt.Sprintf("📦 %s Commits squashed on branch: %s", timeStr, data.BranchName)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Get color for event type and write colored output
	color := f.getColorForEventType(event.Type)
	_, err := fmt.Fprintf(f.writer, "%s%s%s\n\n", color, output, Reset)
	if err != nil {
		return fmt.Errorf("failed to write to console: %w", err)
	}

	return nil
}

// Flush ensures all buffered output is written
func (f *JSONFormatter) Flush() error {
	return nil
}

// formatTime returns current time in HH:MM:SS format
func (f *JSONFormatter) formatTime() string {
	return time.Now().Format("15:04:05")
}

// getColorForEventType returns the ANSI color code for an event type
func (f *JSONFormatter) getColorForEventType(eventType events.EventType) string {
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
