package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

const (
	// Content limits for non-verbose mode
	MaxCodeBlockLines = 10
	MaxCodeBlockChars = 5000
)

// ToolInputFilter defines which fields to hide for specific tools
type ToolInputFilter struct {
	ToolName string
	Fields   []string
}

// Default filters for non-verbose mode
var defaultToolInputFilters = []ToolInputFilter{
	{
		ToolName: "Write",
		Fields:   []string{"content"},
	},
	{
		ToolName: "Edit",
		Fields:   []string{"new_string", "old_string"},
	},
}

// JSONFormatter formats events as human-readable text with color highlighting
type JSONFormatter struct {
	writer  io.Writer
	verbose bool
}

// NewConsoleFormatter creates a new JSONFormatter
func NewConsoleFormatter(writer io.Writer, verbose bool) *JSONFormatter {
	return &JSONFormatter{
		writer:  writer,
		verbose: verbose,
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

// applyToolInputFilters replaces filtered field values with placeholder text
func (f *JSONFormatter) applyToolInputFilters(toolName string, input map[string]interface{}) map[string]interface{} {
	if f.verbose {
		return input
	}

	// Create a copy to avoid modifying original
	filtered := make(map[string]interface{})
	for k, v := range input {
		filtered[k] = v
	}

	// Apply filters by replacing values with placeholder
	for _, filter := range defaultToolInputFilters {
		if filter.ToolName == toolName {
			for _, field := range filter.Fields {
				if _, exists := filtered[field]; exists {
					filtered[field] = "<hidden, use --verbose to see>"
				}
			}
		}
	}

	return filtered
}

// limitCodeBlock truncates content to max lines and chars
func (f *JSONFormatter) limitCodeBlock(content string) string {
	if f.verbose {
		return content
	}

	lines := strings.Split(content, "\n")

	// Limit by line count
	if len(lines) > MaxCodeBlockLines {
		hiddenLines := len(lines) - MaxCodeBlockLines
		lines = lines[:MaxCodeBlockLines]
		lines = append(lines, fmt.Sprintf("... (%d more lines hidden, use --verbose to see all)", hiddenLines))
	}

	result := strings.Join(lines, "\n")

	// Limit by character count
	if len(result) > MaxCodeBlockChars {
		hiddenChars := len(content) - MaxCodeBlockChars
		result = result[:MaxCodeBlockChars] + fmt.Sprintf("\n... (%d more characters hidden, use --verbose to see all)", hiddenChars)
	}

	return result
}

// Format processes an event and outputs it as human-readable colored text
func (f *JSONFormatter) Format(event events.Event) error {
	var output string
	timeStr := fmt.Sprintf("[%s] ", f.formatTime())

	switch event.Type {
	// Small title event
	case events.EventRunPromptStarted:
		data := event.Data.(events.RunPromptStartedData)
		title := "🚀 Run Prompt Started"
		output = f.formatSmallTitle(title) + "\n" + f.formatCodeBlock(data.Prompt, "")

		// Add optional metadata
		if data.BaseURL != "" {
			output += fmt.Sprintf("\n🌐 Base URL: %s", data.BaseURL)
		}
		if data.Cwd != "" {
			output += fmt.Sprintf("\n📁 Working Directory: %s", data.Cwd)
		}
		if data.FileList != "" {
			output += fmt.Sprintf("\n📄 File List: %s", data.FileList)
		}

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		output = fmt.Sprintf("💬 %s%s", timeStr, data.Text)

	case events.EventClaudeToolUse:
		data := event.Data.(events.ToolUseData)

		// Apply input filtering
		filteredInput := f.applyToolInputFilters(data.Name, data.Input)

		prettyJSON, err := f.formatPrettyJSON(filteredInput)
		if err != nil {
			return fmt.Errorf("failed to format tool input: %w", err)
		}

		// Apply content limiting
		limitedJSON := f.limitCodeBlock(prettyJSON)

		output = fmt.Sprintf("🔧 %sTool: %s\n%s", timeStr, data.Name, f.formatCodeBlock(limitedJSON, "json"))

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		limitedContent := f.limitCodeBlock(data.Content)
		output = fmt.Sprintf("📋 %sTool Result\n%s", timeStr, f.formatCodeBlock(limitedContent, ""))

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
		output = fmt.Sprintf("▶️ %sIteration %d/%d started", timeStr, data.Current, data.Total)

	case events.EventIterationCompleted:
		data := event.Data.(events.IterationCompletedData)
		output = fmt.Sprintf("✅ %sIteration %d/%d completed in %s", timeStr, data.Current, data.Total, f.formatDuration(data.Duration))

	case events.EventIterationFailed:
		data := event.Data.(events.IterationFailedData)
		errMsg := "unknown error"
		if data.Error != nil {
			errMsg = data.Error.Error()
		}
		output = fmt.Sprintf("❌ %sIteration %d/%d failed: %s", timeStr, data.Current, data.Total, errMsg)

	case events.EventLoopCompleted:
		data := event.Data.(events.LoopCompletedData)
		output = fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
			data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, f.formatDuration(data.TotalDuration))

	case events.EventLoopInterrupted:
		data := event.Data.(events.LoopInterruptedData)
		output = fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)

	case events.EventSleepStarted:
		data := event.Data.(events.SleepStartedData)
		output = fmt.Sprintf("💤 %sSleeping for %s", timeStr, f.formatDuration(data.Duration))

	// Evolution events
	case events.EventImprovementStarted:
		data := event.Data.(events.ImprovementStartedData)
		output = fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		output = fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		output = fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		output = fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)

	case events.EventEvolveCompleted:
		data := event.Data.(events.EvolveCompletedData)
		output = fmt.Sprintf("🎉 Evolution completed: %s (Total: %s)", data.FinalBranch, f.formatDuration(data.TotalDuration))

	case events.EventEvolveInterrupted:
		data := event.Data.(events.EvolveInterruptedData)
		output = fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)

	// Git operations
	case events.EventGitBranchCreated:
		data := event.Data.(events.BranchCreatedData)
		output = fmt.Sprintf("🌿 %sBranch created: %s", timeStr, data.BranchName)
		if data.Base != "" {
			output += fmt.Sprintf(" (from %s)", data.Base)
		}

	case events.EventGitBranchCheckedOut:
		data := event.Data.(events.BranchCheckedOutData)
		output = fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		output = fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		output = fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)

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
