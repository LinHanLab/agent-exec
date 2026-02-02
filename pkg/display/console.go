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

// formatBigTitle wraps title with ==== lines and applies reverse video
func (f *JSONFormatter) formatBigTitle(title string, color string) string {
	separator := strings.Repeat("=", len(title))

	// Apply reverse video to separator and title
	highlightedSeparator := f.applyReverseVideo(separator, color)
	highlightedTitle := f.applyReverseVideo(title, color)

	return fmt.Sprintf("\n%s\n%s\n%s\n", highlightedSeparator, highlightedTitle, highlightedSeparator)
}

// formatSmallTitle wraps title with --- lines and applies reverse video
func (f *JSONFormatter) formatSmallTitle(title string, color string) string {
	separator := strings.Repeat("-", len(title))

	// Apply reverse video to separator and title
	highlightedSeparator := f.applyReverseVideo(separator, color)
	highlightedTitle := f.applyReverseVideo(title, color)

	return fmt.Sprintf("%s\n%s\n%s", highlightedSeparator, highlightedTitle, highlightedSeparator)
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

// applyReverseVideo wraps text with reverse video effect
// The color parameter should be the existing color code (e.g., BoldYellow)
func (f *JSONFormatter) applyReverseVideo(text string, color string) string {
	if color == "" {
		return fmt.Sprintf("%s%s%s", ReverseVideo, text, Reset)
	}
	return fmt.Sprintf("%s%s%s%s", color, ReverseVideo, text, Reset)
}

// indentContent adds 4-space prefix to each line of content
func (f *JSONFormatter) indentContent(content string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	indented := make([]string, len(lines))
	for i, line := range lines {
		indented[i] = "    " + line
	}
	return strings.Join(indented, "\n")
}

// Format processes an event and outputs it as human-readable colored text
func (f *JSONFormatter) Format(event events.Event) error {
	var output string
	timeStr := fmt.Sprintf("[%s] ", f.formatTime())

	switch event.Type {
	// Small title event
	case events.EventRunPromptStarted:
		data := event.Data.(events.RunPromptStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🚀 Run Prompt Started"

		// Format title with reverse video
		formattedTitle := f.formatSmallTitle(title, color)

		// Format and indent code block
		codeBlock := f.formatCodeBlock(data.Prompt, "")
		indentedCodeBlock := f.indentContent(codeBlock)

		output = formattedTitle + "\n" + indentedCodeBlock

		// Add optional metadata (indented)
		if data.BaseURL != "" {
			output += "\n" + f.indentContent(fmt.Sprintf("🌐 Base URL: %s", data.BaseURL))
		}
		if data.Cwd != "" {
			output += "\n" + f.indentContent(fmt.Sprintf("📁 Working Directory: %s", data.Cwd))
		}
		if data.FileList != "" {
			output += "\n" + f.indentContent(fmt.Sprintf("📄 File List: %s", data.FileList))
		}

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("💬 %s%s", timeStr, data.Text)
		output = f.applyReverseVideo(message, color)

	case events.EventClaudeToolUse:
		data := event.Data.(events.ToolUseData)
		color := f.getColorForEventType(event.Type)

		// Apply input filtering
		filteredInput := f.applyToolInputFilters(data.Name, data.Input)

		prettyJSON, err := f.formatPrettyJSON(filteredInput)
		if err != nil {
			return fmt.Errorf("failed to format tool input: %w", err)
		}

		// Apply content limiting
		limitedJSON := f.limitCodeBlock(prettyJSON)

		// Format title with reverse video
		title := fmt.Sprintf("🔧 %sTool: %s", timeStr, data.Name)
		highlightedTitle := f.applyReverseVideo(title, color)

		// Format and indent code block
		codeBlock := f.formatCodeBlock(limitedJSON, "json")
		indentedCodeBlock := f.indentContent(codeBlock)

		output = highlightedTitle + "\n" + indentedCodeBlock

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		color := f.getColorForEventType(event.Type)
		limitedContent := f.limitCodeBlock(data.Content)

		// Format title with reverse video
		title := fmt.Sprintf("📋 %sTool Result", timeStr)
		highlightedTitle := f.applyReverseVideo(title, color)

		// Format and indent code block
		codeBlock := f.formatCodeBlock(limitedContent, "")
		indentedCodeBlock := f.indentContent(codeBlock)

		output = highlightedTitle + "\n" + indentedCodeBlock

	case events.EventClaudeExecutionResult:
		data := event.Data.(events.ExecutionResultData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("⏱️ Execution completed in %s", f.formatDuration(data.Duration))
		output = f.applyReverseVideo(message, color)

	// Big title events
	case events.EventLoopStarted:
		data := event.Data.(events.LoopStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🔄 Loop Started"

		// Format title with reverse video
		formattedTitle := f.formatBigTitle(title, color)

		// Indent content
		content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
		indentedContent := f.indentContent(content)

		output = formattedTitle + indentedContent

	case events.EventEvolveStarted:
		data := event.Data.(events.EvolveStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🧬 Evolution Started"

		// Format title with reverse video
		formattedTitle := f.formatBigTitle(title, color)

		// Indent content
		content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
		indentedContent := f.indentContent(content)

		output = formattedTitle + indentedContent

	case events.EventRoundStarted:
		data := event.Data.(events.RoundStartedData)
		color := f.getColorForEventType(event.Type)
		title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)
		output = f.formatBigTitle(title, color)

	// Loop execution events
	case events.EventIterationStarted:
		data := event.Data.(events.IterationStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("▶️ %sIteration %d/%d started", timeStr, data.Current, data.Total)
		output = f.applyReverseVideo(message, color)

	case events.EventIterationCompleted:
		data := event.Data.(events.IterationCompletedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("✅ %sIteration %d/%d completed in %s", timeStr, data.Current, data.Total, f.formatDuration(data.Duration))
		output = f.applyReverseVideo(message, color)

	case events.EventIterationFailed:
		data := event.Data.(events.IterationFailedData)
		color := f.getColorForEventType(event.Type)
		errMsg := "unknown error"
		if data.Error != nil {
			errMsg = data.Error.Error()
		}
		message := fmt.Sprintf("❌ %sIteration %d/%d failed: %s", timeStr, data.Current, data.Total, errMsg)
		output = f.applyReverseVideo(message, color)

	case events.EventLoopCompleted:
		data := event.Data.(events.LoopCompletedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
			data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, f.formatDuration(data.TotalDuration))
		output = f.applyReverseVideo(message, color)

	case events.EventLoopInterrupted:
		data := event.Data.(events.LoopInterruptedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)
		output = f.applyReverseVideo(message, color)

	case events.EventSleepStarted:
		data := event.Data.(events.SleepStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("💤 %sSleeping for %s", timeStr, f.formatDuration(data.Duration))
		output = f.applyReverseVideo(message, color)

	// Evolution events
	case events.EventImprovementStarted:
		data := event.Data.(events.ImprovementStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)
		output = f.applyReverseVideo(message, color)

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)
		output = f.applyReverseVideo(message, color)

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)
		output = f.applyReverseVideo(message, color)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)
		output = f.applyReverseVideo(message, color)

	case events.EventEvolveCompleted:
		data := event.Data.(events.EvolveCompletedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🎉 Evolution completed, final branch: %s (total duration: %s)",
			data.FinalBranch, f.formatDuration(data.TotalDuration))
		output = f.applyReverseVideo(message, color)

	case events.EventEvolveInterrupted:
		data := event.Data.(events.EvolveInterruptedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)
		output = f.applyReverseVideo(message, color)

	// Git operations
	case events.EventGitBranchCreated:
		data := event.Data.(events.BranchCreatedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🌿 %sBranch created: %s", timeStr, data.BranchName)
		if data.Base != "" {
			message += fmt.Sprintf(" (from %s)", data.Base)
		}
		output = f.applyReverseVideo(message, color)

	case events.EventGitBranchCheckedOut:
		data := event.Data.(events.BranchCheckedOutData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)
		output = f.applyReverseVideo(message, color)

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)
		output = f.applyReverseVideo(message, color)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)
		output = f.applyReverseVideo(message, color)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Write output (color already applied within each case)
	_, err := fmt.Fprintf(f.writer, "%s\n", output)
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
