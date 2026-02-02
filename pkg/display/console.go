package display

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// JSONFormatter formats events as JSON lines with color highlighting
type JSONFormatter struct {
	writer io.Writer
}

// NewConsoleFormatter creates a new JSONFormatter
func NewConsoleFormatter(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
	}
}

// Format processes an event and outputs it as a colored JSON line
func (f *JSONFormatter) Format(event events.Event) error {
	var jsonBytes []byte
	var err error

	switch event.Type {
	// Prompt execution events
	case events.EventPromptStarted:
		data := event.Data.(events.PromptStartedData)
		output := struct {
			Type     string `json:"type"`
			Time     string `json:"time"`
			Prompt   string `json:"prompt"`
			Cwd      string `json:"cwd,omitempty"`
			BaseURL  string `json:"base_url,omitempty"`
			FileList string `json:"file_list,omitempty"`
		}{
			Type:     string(event.Type),
			Time:     f.formatTime(),
			Prompt:   data.Prompt,
			Cwd:      data.Cwd,
			BaseURL:  data.BaseURL,
			FileList: data.FileList,
		}
		jsonBytes, err = json.Marshal(output)

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		output := struct {
			Type string `json:"type"`
			Time string `json:"time"`
			Text string `json:"text"`
		}{
			Type: string(event.Type),
			Time: f.formatTime(),
			Text: data.Text,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventClaudeToolUse:
		data := event.Data.(events.ToolUseData)
		output := struct {
			Type  string                 `json:"type"`
			Time  string                 `json:"time"`
			Name  string                 `json:"name"`
			Input map[string]interface{} `json:"input"`
		}{
			Type:  string(event.Type),
			Time:  f.formatTime(),
			Name:  data.Name,
			Input: data.Input,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		output := struct {
			Type    string `json:"type"`
			Time    string `json:"time"`
			Content string `json:"content"`
		}{
			Type:    string(event.Type),
			Time:    f.formatTime(),
			Content: data.Content,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventClaudeExecutionResult:
		data := event.Data.(events.ExecutionResultData)
		output := struct {
			Type        string  `json:"type"`
			Time        string  `json:"time"`
			DurationSec float64 `json:"duration_sec"`
		}{
			Type:        string(event.Type),
			Time:        f.formatTime(),
			DurationSec: data.Duration.Seconds(),
		}
		jsonBytes, err = json.Marshal(output)

	// Loop execution events
	case events.EventLoopStarted:
		data := event.Data.(events.LoopStartedData)
		output := struct {
			Type            string `json:"type"`
			Time            string `json:"time"`
			TotalIterations int    `json:"total_iterations"`
		}{
			Type:            string(event.Type),
			Time:            f.formatTime(),
			TotalIterations: data.TotalIterations,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventIterationStarted:
		data := event.Data.(events.IterationStartedData)
		output := struct {
			Type    string `json:"type"`
			Time    string `json:"time"`
			Current int    `json:"current"`
			Total   int    `json:"total"`
		}{
			Type:    string(event.Type),
			Time:    f.formatTime(),
			Current: data.Current,
			Total:   data.Total,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventIterationCompleted:
		data := event.Data.(events.IterationCompletedData)
		output := struct {
			Type        string  `json:"type"`
			Time        string  `json:"time"`
			Current     int     `json:"current"`
			Total       int     `json:"total"`
			DurationSec float64 `json:"duration_sec"`
		}{
			Type:        string(event.Type),
			Time:        f.formatTime(),
			Current:     data.Current,
			Total:       data.Total,
			DurationSec: data.Duration.Seconds(),
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventIterationFailed:
		data := event.Data.(events.IterationFailedData)
		output := struct {
			Type    string `json:"type"`
			Time    string `json:"time"`
			Current int    `json:"current"`
			Total   int    `json:"total"`
			Error   string `json:"error"`
		}{
			Type:    string(event.Type),
			Time:    f.formatTime(),
			Current: data.Current,
			Total:   data.Total,
			Error:   data.Error.Error(),
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventSleepStarted:
		data := event.Data.(events.SleepStartedData)
		output := struct {
			Type     string `json:"type"`
			Time     string `json:"time"`
			Duration string `json:"duration"`
		}{
			Type:     string(event.Type),
			Time:     f.formatTime(),
			Duration: data.Duration.String(),
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventLoopCompleted:
		data := event.Data.(events.LoopCompletedData)
		output := struct {
			Type            string  `json:"type"`
			Time            string  `json:"time"`
			TotalIterations int     `json:"total_iterations"`
			Successful      int     `json:"successful"`
			Failed          int     `json:"failed"`
			DurationSec     float64 `json:"duration_sec"`
		}{
			Type:            string(event.Type),
			Time:            f.formatTime(),
			TotalIterations: data.TotalIterations,
			Successful:      data.SuccessfulIterations,
			Failed:          data.FailedIterations,
			DurationSec:     data.TotalDuration.Seconds(),
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventLoopInterrupted:
		data := event.Data.(events.LoopInterruptedData)
		output := struct {
			Type      string `json:"type"`
			Time      string `json:"time"`
			Completed int    `json:"completed"`
			Total     int    `json:"total"`
		}{
			Type:      string(event.Type),
			Time:      f.formatTime(),
			Completed: data.CompletedIterations,
			Total:     data.TotalIterations,
		}
		jsonBytes, err = json.Marshal(output)

	// Evolution events
	case events.EventEvolveStarted:
		data := event.Data.(events.EvolveStartedData)
		output := struct {
			Type       string `json:"type"`
			Time       string `json:"time"`
			Prompt     string `json:"prompt"`
			Iterations int    `json:"iterations"`
		}{
			Type:       string(event.Type),
			Time:       f.formatTime(),
			Prompt:     data.Prompt,
			Iterations: data.Iterations,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventRoundStarted:
		data := event.Data.(events.RoundStartedData)
		output := struct {
			Type  string `json:"type"`
			Time  string `json:"time"`
			Round int    `json:"round"`
			Total int    `json:"total"`
		}{
			Type:  string(event.Type),
			Time:  f.formatTime(),
			Round: data.Round,
			Total: data.Total,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventImprovementStarted:
		data := event.Data.(events.ImprovementStartedData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Branch string `json:"branch"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Branch: data.BranchName,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		output := struct {
			Type    string `json:"type"`
			Time    string `json:"time"`
			Branch1 string `json:"branch1"`
			Branch2 string `json:"branch2"`
		}{
			Type:    string(event.Type),
			Time:    f.formatTime(),
			Branch1: data.Branch1,
			Branch2: data.Branch2,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		output := struct {
			Type        string `json:"type"`
			Time        string `json:"time"`
			Attempt     int    `json:"attempt"`
			MaxAttempts int    `json:"max_attempts"`
		}{
			Type:        string(event.Type),
			Time:        f.formatTime(),
			Attempt:     data.Attempt,
			MaxAttempts: data.MaxAttempts,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Winner string `json:"winner"`
			Loser  string `json:"loser"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Winner: data.Winner,
			Loser:  data.Loser,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventEvolveCompleted:
		data := event.Data.(events.EvolveCompletedData)
		output := struct {
			Type        string  `json:"type"`
			Time        string  `json:"time"`
			FinalBranch string  `json:"final_branch"`
			TotalRounds int     `json:"total_rounds"`
			DurationSec float64 `json:"duration_sec"`
		}{
			Type:        string(event.Type),
			Time:        f.formatTime(),
			FinalBranch: data.FinalBranch,
			TotalRounds: data.TotalRounds,
			DurationSec: data.TotalDuration.Seconds(),
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventEvolveInterrupted:
		data := event.Data.(events.EvolveInterruptedData)
		output := struct {
			Type            string `json:"type"`
			Time            string `json:"time"`
			CompletedRounds int    `json:"completed_rounds"`
			TotalRounds     int    `json:"total_rounds"`
			Winner          string `json:"winner"`
		}{
			Type:            string(event.Type),
			Time:            f.formatTime(),
			CompletedRounds: data.CompletedRounds,
			TotalRounds:     data.TotalRounds,
			Winner:          data.Winner,
		}
		jsonBytes, err = json.Marshal(output)

	// Git operations
	case events.EventGitBranchCreated:
		data := event.Data.(events.BranchCreatedData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Branch string `json:"branch"`
			Base   string `json:"base,omitempty"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Branch: data.BranchName,
			Base:   data.Base,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventGitBranchCheckedOut:
		data := event.Data.(events.BranchCheckedOutData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Branch string `json:"branch"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Branch: data.BranchName,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Branch string `json:"branch"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Branch: data.BranchName,
		}
		jsonBytes, err = json.Marshal(output)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		output := struct {
			Type   string `json:"type"`
			Time   string `json:"time"`
			Branch string `json:"branch"`
		}{
			Type:   string(event.Type),
			Time:   f.formatTime(),
			Branch: data.BranchName,
		}
		jsonBytes, err = json.Marshal(output)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Get color for event type and write colored JSON line
	color := f.getColorForEventType(event.Type)
	_, err = fmt.Fprintf(f.writer, "%s%s%s\n", color, string(jsonBytes), Reset)
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
	// Priority 1 - Bold Cyan (ANTHROPIC_BASE_URL)
	case events.EventPromptStarted:
		return BoldCyan

	// Priority 2 - Bold Yellow (Phase Markers)
	case events.EventLoopStarted,
		events.EventIterationStarted,
		events.EventEvolveStarted,
		events.EventRoundStarted,
		events.EventImprovementStarted,
		events.EventComparisonStarted,
		events.EventSleepStarted:
		return BoldYellow

	// Priority 3 - Bold Green (Success Results)
	case events.EventClaudeExecutionResult,
		events.EventLoopCompleted,
		events.EventEvolveCompleted,
		events.EventIterationCompleted,
		events.EventWinnerSelected:
		return BoldGreen

	// Priority 4 - Bold Red (Errors/Interruptions)
	case events.EventIterationFailed,
		events.EventLoopInterrupted,
		events.EventEvolveInterrupted:
		return BoldRed

	// Priority 5 - Magenta (Tool Operations)
	case events.EventClaudeToolUse,
		events.EventClaudeToolResult:
		return Magenta

	// Priority 6 - Cyan (Git Operations)
	case events.EventGitBranchCreated,
		events.EventGitBranchCheckedOut,
		events.EventGitBranchDeleted,
		events.EventGitCommitsSquashed:
		return Cyan

	// Priority 7 - White/No Color (Regular Messages)
	case events.EventClaudeAssistantMessage,
		events.EventComparisonRetry:
		return ""

	default:
		return ""
	}
}
