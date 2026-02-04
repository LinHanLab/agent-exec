package display

import (
	"encoding/json"
	"fmt"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// mustGetEventData safely extracts typed data from an event.
func mustGetEventData[T any](event events.Event, expectedType string) T {
	data, ok := event.Data.(T)
	if !ok {
		panic(fmt.Sprintf("event data for %s must be %T, got %T",
			expectedType, data, event.Data))
	}
	return data
}

// formatPrettyJSON marshals data to indented JSON.
func formatPrettyJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func formatRunPromptStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.RunPromptStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🚀 Run Prompt Started"

	formattedTitle := fmt.Sprintf("%s%s%s", color, title, Reset)
	promptContent := ctx.TextFormatter.FormatContentWithFrame(data.Prompt, true)

	output := formattedTitle + promptContent

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
	data := mustGetEventData[events.AssistantMessageData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	title := fmt.Sprintf("💬 %sAssistant", timeStr)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	content := ctx.TextFormatter.FormatContentWithFrameAndColor(data.Text, color)

	return coloredTitle + content, nil
}

func formatClaudeToolUse(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ToolUseData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())

	filteredInput := ctx.ContentFilter.ApplyToolInputFilters(data.Name, data.Input)

	prettyJSON, err := formatPrettyJSON(filteredInput)
	if err != nil {
		return "", fmt.Errorf("failed to format tool input: %w", err)
	}

	limitedJSON := ctx.ContentFilter.LimitCodeBlock(prettyJSON)

	title := fmt.Sprintf("🔧 %sTool: %s", timeStr, data.Name)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	jsonContent := ctx.TextFormatter.FormatContentWithFrame(limitedJSON)

	return coloredTitle + jsonContent, nil
}

func formatClaudeToolResult(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ToolResultData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	limitedContent := ctx.ContentFilter.LimitCodeBlock(data.Content)

	title := fmt.Sprintf("📋 %sTool Result", timeStr)
	coloredTitle := fmt.Sprintf("%s%s%s", color, title, Reset)

	resultContent := ctx.TextFormatter.FormatContentWithFrame(limitedContent)

	return coloredTitle + resultContent, nil
}

func formatClaudeExecutionResult(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ExecutionResultData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⏱️ Execution completed in %s", ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatLoopStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🔄 Loop Started"

	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

func formatIterationStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.IterationStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("▶️ %sIteration %d/%d started", timeStr, data.Current, data.Total)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatIterationCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.IterationCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("✅ %sIteration %d/%d completed in %s", timeStr, data.Current, data.Total, ctx.TextFormatter.FormatDuration(data.Duration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

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

func formatLoopCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🏁 Loop completed: %d/%d successful, %d failed (Total: %s)",
		data.SuccessfulIterations, data.TotalIterations, data.FailedIterations, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatLoopInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.LoopInterruptedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("⚠️ Loop interrupted: %d/%d iterations completed", data.CompletedIterations, data.TotalIterations)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatSleepStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.SleepStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("💤 %sSleeping for %s", timeStr, ctx.TextFormatter.FormatDuration(data.Duration))
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatEvolveStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := "🧬 Evolution Started"

	formattedTitle := ctx.TextFormatter.ApplyReverseVideo(title, color)

	content := fmt.Sprintf("🔢 Iterations: %d", data.TotalIterations)
	indentedContent := ctx.TextFormatter.IndentContent(content)

	return formattedTitle + "\n" + indentedContent, nil
}

func formatRoundStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.RoundStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	title := fmt.Sprintf("🎯 Round %d/%d", data.Round, data.Total)

	return ctx.TextFormatter.ApplyReverseVideo(title, color), nil
}

func formatImprovementStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ImprovementStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔨 %sImproving branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatComparisonStarted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ComparisonStartedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("⚖️ %sComparing: %s vs %s", timeStr, data.Branch1, data.Branch2)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatComparisonRetry(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.ComparisonRetryData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔁 %sComparison retry %d/%d", timeStr, data.Attempt, data.MaxAttempts)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatWinnerSelected(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.WinnerSelectedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🏆 %sWinner: %s (eliminated: %s)", timeStr, data.Winner, data.Loser)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatEvolveCompleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveCompletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🎉 Evolution completed, final branch: %s (total duration: %s)",
		data.FinalBranch, ctx.TextFormatter.FormatDuration(data.TotalDuration))
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

func formatEvolveInterrupted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.EvolveInterruptedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	message := fmt.Sprintf("🛑 Evolution interrupted: %d/%d rounds completed", data.CompletedRounds, data.TotalRounds)
	return ctx.TextFormatter.ApplyReverseVideo(message, color), nil
}

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

func formatGitBranchCheckedOut(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.BranchCheckedOutData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🔀 %sChecked out branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatGitBranchDeleted(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.BranchDeletedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("🗑️ %sBranch deleted: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}

func formatGitCommitsSquashed(event events.Event, ctx *FormatContext) (string, error) {
	data := mustGetEventData[events.CommitsSquashedData](event, string(event.Type))
	color := GetColorForEventType(event.Type)
	timeStr := fmt.Sprintf("[%s] ", ctx.TextFormatter.FormatTime())
	message := fmt.Sprintf("📦 %sCommits squashed on branch: %s", timeStr, data.BranchName)
	return fmt.Sprintf("%s%s%s", color, message, Reset), nil
}
