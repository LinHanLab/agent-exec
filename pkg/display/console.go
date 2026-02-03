package display

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
	"golang.org/x/term"
)

const (
	// Content limits for non-verbose mode
	MaxCodeBlockLines = 10
	MaxCodeBlockChars = 5000
	// Default terminal width if detection fails
	DefaultTerminalWidth = 80
	// Indentation for content blocks
	ContentIndent = "    "
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
	writer        io.Writer
	verbose       bool
	terminalWidth int
}

// NewConsoleFormatter creates a new JSONFormatter
func NewConsoleFormatter(writer io.Writer, verbose bool) *JSONFormatter {
	return &JSONFormatter{
		writer:        writer,
		verbose:       verbose,
		terminalWidth: getTerminalWidth(),
	}
}

// getTerminalWidth returns the current terminal width, or default if detection fails
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// getSpacingBefore returns a single blank line before each event
func (f *JSONFormatter) getSpacingBefore(eventType events.EventType) string {
	// Always use single blank line between events
	return "\n"
}

// wrapText wraps text to fit within terminal width, respecting indentation
// Only wraps at natural break points (spaces, commas, hyphens)
func (f *JSONFormatter) wrapText(text string, indent string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(lines))

	// Calculate available width for content (terminal width - indent length)
	availableWidth := f.terminalWidth - len(indent)
	if availableWidth <= 0 {
		availableWidth = DefaultTerminalWidth - len(indent)
	}

	for _, line := range lines {
		// If line fits within available width, keep it as is
		if len(line) <= availableWidth {
			wrapped = append(wrapped, indent+line)
			continue
		}

		// For long lines, try to wrap at natural break points
		remaining := line
		for len(remaining) > 0 {
			if len(remaining) <= availableWidth {
				wrapped = append(wrapped, indent+remaining)
				break
			}

			// Find a good break point (space, comma, etc.)
			breakPoint := -1
			for i := availableWidth - 1; i > availableWidth/2 && i < len(remaining); i-- {
				if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == '-' {
					breakPoint = i + 1
					break
				}
			}

			// If no break point found, don't wrap (keep line as is to preserve content)
			if breakPoint == -1 {
				wrapped = append(wrapped, indent+remaining)
				break
			}

			wrapped = append(wrapped, indent+remaining[:breakPoint])
			remaining = strings.TrimLeft(remaining[breakPoint:], " ")
		}
	}

	return strings.Join(wrapped, "\n")
}

// formatContent formats content with indentation (no wrapping for JSON)
func (f *JSONFormatter) formatContent(content string) string {
	if content == "" {
		return ""
	}
	return "\n" + f.indentContent(content) + "\n"
}

// formatWrappedContent formats content with indentation and wrapping for plain text
func (f *JSONFormatter) formatWrappedContent(content string) string {
	if content == "" {
		return ""
	}
	return "\n" + f.wrapText(content, ContentIndent) + "\n"
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

// indentContent adds ContentIndent prefix to each line of content
func (f *JSONFormatter) indentContent(content string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	indented := make([]string, len(lines))
	for i, line := range lines {
		indented[i] = ContentIndent + line
	}
	return strings.Join(indented, "\n")
}

// Format processes an event and outputs it as human-readable colored text
func (f *JSONFormatter) Format(event events.Event) error {
	var output string
	timeStr := fmt.Sprintf("[%s] ", f.formatTime())

	// Add spacing before event
	spacing := f.getSpacingBefore(event.Type)

	switch event.Type {
	case events.EventRunPromptStarted:
		data := event.Data.(events.RunPromptStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🚀 Run Prompt Started"

		// Format title
		formattedTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

		// Format and wrap prompt content (plain text, so wrap it)
		promptContent := f.formatWrappedContent(data.Prompt)

		output = formattedTitle + promptContent

		// Add optional metadata (indented)
		if data.BaseURL != "" {
			output += f.indentContent(fmt.Sprintf("🌐 Base URL: %s", data.BaseURL)) + "\n"
		}
		if data.Cwd != "" {
			output += f.indentContent(fmt.Sprintf("📁 Working Directory: %s", data.Cwd)) + "\n"
		}
		if data.FileList != "" {
			output += f.indentContent(fmt.Sprintf("📄 File List: %s", data.FileList)) + "\n"
		}

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("💬 %s%s", timeStr, data.Text)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

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

		// Format title
		title := fmt.Sprintf("🔧 %sTool: %s", timeStr, data.Name)
		coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

		// Format and wrap content
		jsonContent := f.formatContent(limitedJSON)

		output = coloredTitle + jsonContent

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		color := f.getColorForEventType(event.Type)
		limitedContent := f.limitCodeBlock(data.Content)

		// Format title
		title := fmt.Sprintf("📋 %sTool Result", timeStr)
		coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

		// Format and wrap content
		resultContent := f.formatContent(limitedContent)

		output = coloredTitle + resultContent

	case events.EventClaudeExecutionResult:
		data := event.Data.(events.ExecutionResultData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("⏱️ Execution completed in %s", f.formatDuration(data.Duration))
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	// Big title events
	case events.EventLoopStarted:
		data := event.Data.(events.LoopStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🔄 Loop Started"

		// Format title with reverse video
		formattedTitle := f.applyReverseVideo(title, color)

		// Indent content
		content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
		indentedContent := f.indentContent(content)

		output = formattedTitle + "\n" + indentedContent

	case events.EventEvolveStarted:
		data := event.Data.(events.EvolveStartedData)
		color := f.getColorForEventType(event.Type)
		title := "🧬 Evolution Started"

		// Format title with reverse video
		formattedTitle := f.applyReverseVideo(title, color)

		// Indent content
		content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
		indentedContent := f.indentContent(content)

		output = formattedTitle + "\n" + indentedContent

	case events.EventRoundStarted:
		data := event.Data.(events.RoundStartedData)
		color := f.getColorForEventType(event.Type)
		title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)

		// Format title with reverse video
		output = f.applyReverseVideo(title, color)

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
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	// Evolution events
	case events.EventImprovementStarted:
		data := event.Data.(events.ImprovementStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

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
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventGitBranchCheckedOut:
		data := event.Data.(events.BranchCheckedOutData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		color := f.getColorForEventType(event.Type)
		message := fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)
		output = fmt.Sprintf("%s%s%s", color, message, Reset)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Write spacing first, then output
	_, err := fmt.Fprint(f.writer, spacing)
	if err != nil {
		return fmt.Errorf("failed to write spacing to console: %w", err)
	}

	// Write output (color already applied within each case)
	_, err = fmt.Fprintf(f.writer, "%s\n", output)
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
