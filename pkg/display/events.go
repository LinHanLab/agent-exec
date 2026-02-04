package display

import (
	"encoding/json"
	"fmt"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// mustGetEventData safely extracts typed data from an event.
// Panics with a clear message if type assertion fails.
func mustGetEventData[T any](event events.Event, expectedType string) T {
	data, ok := event.Data.(T)
	if !ok {
		panic(fmt.Sprintf("event data for %s must be %T, got %T",
			expectedType, data, event.Data))
	}
	return data
}

// formatPrettyJSON marshals data to indented JSON.
// Returns an error only if marshaling fails (e.g., unsupported type).
func formatPrettyJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// =============================================================================
// Prompt Execution Formatters
// =============================================================================

// formatRunPromptStarted formats prompt execution start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatRunPromptStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.RunPromptStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🚀 Run Prompt Started"

	// Format title
	formattedTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Format prompt content with box border (only prompt gets border)
	promptContent := ctx.TextFormatter.FormatContentWithFrame(data.Prompt, true)

	output := formattedTitle + promptContent

	// Add optional metadata (indented, no frame)
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

// =============================================================================
// Claude Streaming Formatters
// =============================================================================

// formatClaudeAssistantMessage formats assistant message events.
// Returns nil error (error return satisfies EventFormatter type).
func formatClaudeAssistantMessage(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.AssistantMessageData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	title := fmt.Sprintf("💬 %sAssistant", timeStr)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Pass color to FormatContentWithFrameAndColor so it applies to all wrapped lines
	content := ctx.TextFormatter.FormatContentWithFrameAndColor(data.Text, color)

	return coloredTitle + content, nil
}

// formatClaudeToolUse formats tool use events.
// Returns an error if JSON marshaling fails in formatPrettyJSON.
func formatClaudeToolUse(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ToolUseData](event, string(event.Type))
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

	// Format JSON content with frame
	jsonContent := ctx.TextFormatter.FormatContentWithFrame(limitedJSON)

	return coloredTitle + jsonContent, nil
}

// formatClaudeToolResult formats tool result events.
// Returns nil error (error return satisfies EventFormatter type).
func formatClaudeToolResult(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ToolResultData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	limitedContent := ctx.ContentFilter.LimitCodeBlock(data.Content)

	// Format title
	title := fmt.Sprintf("📋 %sTool Result", timeStr)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	// Format result content with frame
	resultContent := ctx.TextFormatter.FormatContentWithFrame(limitedContent)

	return coloredTitle + resultContent, nil
}

// formatClaudeExecutionResult formats execution result events.
// Returns nil error (error return satisfies EventFormatter type).
func formatClaudeExecutionResult(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ExecutionResultData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⏱️ Execution completed in %s", ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// =============================================================================
// Loop Execution Formatters
// =============================================================================

// formatLoopStarted formats loop execution start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatLoopStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🔄 Loop Started"

	// Format title with reverse video
	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	// Indent content (no frame for short metadata)
	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

// formatIterationStarted formats iteration start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatIterationStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.IterationStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("▶️ %sIteration %d/%d started", timeStr, data.Current, data.Total)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatIterationCompleted formats iteration completion events.
// Returns nil error (error return satisfies EventFormatter type).
func formatIterationCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.IterationCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("✅ %sIteration %d/%d completed in %s", timeStr, data.Current, data.Total, ctx.TextFormatter.FormatDuration(data.Duration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatIterationFailed formats iteration failure events.
// Returns nil error (error return satisfies EventFormatter type).
func formatIterationFailed(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.IterationFailedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	errMsg := "unknown error"
	if data.Error != nil {
		errMsg = data.Error.Error()
	}
	message := fmt.Sprintf("❌ %sIteration %d/%d failed: %s", timeStr, data.Current, data.Total, errMsg)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatLoopCompleted formats loop completion events.
// Returns nil error (error return satisfies EventFormatter type).
func formatLoopCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
		data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatLoopInterrupted formats loop interruption events.
// Returns nil error (error return satisfies EventFormatter type).
func formatLoopInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopInterruptedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatSleepStarted formats sleep start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatSleepStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.SleepStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("💤 %sSleeping for %s", timeStr, ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// =============================================================================
// Evolution Workflow Formatters
// =============================================================================

// formatEvolveStarted formats evolution start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatEvolveStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🧬 Evolution Started"

	// Format title with reverse video
	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	// Indent content (no frame for short metadata)
	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

// formatRoundStarted formats round start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatRoundStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.RoundStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)

	// Format title with reverse video
	return ctx.TextFormatter.ApplyReverseVideo(title, color), nil
}

// formatImprovementStarted formats improvement start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatImprovementStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ImprovementStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatComparisonStarted formats comparison start events.
// Returns nil error (error return satisfies EventFormatter type).
func formatComparisonStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ComparisonStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatComparisonRetry formats comparison retry events.
// Returns nil error (error return satisfies EventFormatter type).
func formatComparisonRetry(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ComparisonRetryData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatWinnerSelected formats winner selection events.
// Returns nil error (error return satisfies EventFormatter type).
func formatWinnerSelected(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.WinnerSelectedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatEvolveCompleted formats evolution completion events.
// Returns nil error (error return satisfies EventFormatter type).
func formatEvolveCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🎉 Evolution completed, final branch: %s (total duration: %s)",
		data.FinalBranch, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// formatEvolveInterrupted formats evolution interruption events.
// Returns nil error (error return satisfies EventFormatter type).
func formatEvolveInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveInterruptedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

// =============================================================================
// Git Operation Formatters
// =============================================================================

// formatGitBranchCreated formats git branch creation events.
// Returns nil error (error return satisfies EventFormatter type).
func formatGitBranchCreated(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.BranchCreatedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🌿 %sBranch created: %s", timeStr, data.BranchName)
	if data.Base != "" {
		message += fmt.Sprintf(" (from %s)", data.Base)
	}
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatGitBranchCheckedOut formats git branch checkout events.
// Returns nil error (error return satisfies EventFormatter type).
func formatGitBranchCheckedOut(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.BranchCheckedOutData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatGitBranchDeleted formats git branch deletion events.
// Returns nil error (error return satisfies EventFormatter type).
func formatGitBranchDeleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.BranchDeletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

// formatGitCommitsSquashed formats git commit squash events.
// Returns nil error (error return satisfies EventFormatter type).
func formatGitCommitsSquashed(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.CommitsSquashedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}
