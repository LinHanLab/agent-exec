package display

import (
	"encoding/json"
	"fmt"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// EventFormatter is a function that formats a specific event type
type EventFormatter func(event events.Event, ctx *FormatContext) (string, error)

// FormatContext provides dependencies to event formatters
type FormatContext struct {
	TextFormatter *TextFormatter
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

// formatPrettyJSON marshals data to indented JSON
func formatPrettyJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Individual event formatters

func formatRunPromptStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.RunPromptStartedData)
	color := GetColorForEventType(event.Type)
	title := "🚀 Run Prompt Started"

	// Format title
	formattedTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Format and wrap prompt content (plain text, so wrap it)
	promptContent := ctx.TextFormatter.FormatWrappedContent(data.Prompt)

	output := formattedTitle + promptContent

	// Add optional metadata (indented)
	if data.BaseURL != "" {
		output += ctx.TextFormatter.IndentContent(fmt.Sprintf("🌐 Base URL: %s%s%s", BoldUnderline, data.BaseURL, Reset)) + "\n"
	}
	if data.Cwd != "" {
		output += ctx.TextFormatter.IndentContent(fmt.Sprintf("📁 Working Directory: %s", data.Cwd)) + "\n"
	}
	if data.FileList != "" {
		output += ctx.TextFormatter.IndentContent(fmt.Sprintf("📄 File List: %s", data.FileList)) + "\n"
	}

	return output, nil
}

func formatClaudeAssistantMessage(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.AssistantMessageData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("💬 %s%s", timeStr, data.Text)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatClaudeToolUse(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ToolUseData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())

	// Apply input filtering
	filteredInput := ctx.ContentFilter.ApplyToolInputFilters(data.Name, data.Input)

	prettyJSON, err := formatPrettyJSON(filteredInput)
	if err != nil {
		return "", fmt.Errorf("failed to format tool input: %w", err)
	}

	// Apply content limiting
	limitedJSON := ctx.ContentFilter.LimitCodeBlock(prettyJSON)

	// Format title
	title := fmt.Sprintf("🔧 %sTool: %s", timeStr, data.Name)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Format and wrap content
	jsonContent := ctx.TextFormatter.FormatContent(limitedJSON)

	return coloredTitle + jsonContent, nil
}

func formatClaudeToolResult(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ToolResultData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	limitedContent := ctx.ContentFilter.LimitCodeBlock(data.Content)

	// Format title
	title := fmt.Sprintf("📋 %sTool Result", timeStr)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Format and wrap content
	resultContent := ctx.TextFormatter.FormatContent(limitedContent)

	return coloredTitle + resultContent, nil
}

func formatClaudeExecutionResult(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ExecutionResultData)
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⏱️ Execution completed in %s", ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatLoopStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.LoopStartedData)
	color := GetColorForEventType(event.Type)
	title := "🔄 Loop Started"

	// Format title with reverse video
	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	// Indent content
	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

func formatEvolveStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.EvolveStartedData)
	color := GetColorForEventType(event.Type)
	title := "🧬 Evolution Started"

	// Format title with reverse video
	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	// Indent content
	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

func formatRoundStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.RoundStartedData)
	color := GetColorForEventType(event.Type)
	title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)

	// Format title with reverse video
	return ctx.TextFormatter.ApplyReverseVideo(title, color), nil
}

func formatIterationStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.IterationStartedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("▶️ %sIteration %d/%d started", timeStr, data.Current, data.Total)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatIterationCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.IterationCompletedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("✅ %sIteration %d/%d completed in %s", timeStr, data.Current, data.Total, ctx.TextFormatter.FormatDuration(data.Duration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatIterationFailed(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.IterationFailedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	errMsg := "unknown error"
	if data.Error != nil {
		errMsg = data.Error.Error()
	}
	message := fmt.Sprintf("❌ %sIteration %d/%d failed: %s", timeStr, data.Current, data.Total, errMsg)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatLoopCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.LoopCompletedData)
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
		data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatLoopInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.LoopInterruptedData)
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatSleepStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.SleepStartedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("💤 %sSleeping for %s", timeStr, ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatImprovementStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ImprovementStartedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatComparisonStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ComparisonStartedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatComparisonRetry(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.ComparisonRetryData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatWinnerSelected(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.WinnerSelectedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatEvolveCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.EvolveCompletedData)
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🎉 Evolution completed, final branch: %s (total duration: %s)",
		data.FinalBranch, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatEvolveInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.EvolveInterruptedData)
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatGitBranchCreated(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.BranchCreatedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🌿 %sBranch created: %s", timeStr, data.BranchName)
	if data.Base != "" {
		message += fmt.Sprintf(" (from %s)", data.Base)
	}
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatGitBranchCheckedOut(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.BranchCheckedOutData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatGitBranchDeleted(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.BranchDeletedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatGitCommitsSquashed(event events.Event, ctx *FormatContext) (string, error) {
	data := event.Data.(events.CommitsSquashedData)
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}
