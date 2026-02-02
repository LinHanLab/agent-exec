package display

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

const (
	displayWidth     = 76
	promptMaxLen     = 270
	truncateSuffix   = "[...Truncated]"
	maxResultDisplay = 200
)

// ConsoleFormatter formats events for console output with emoji and colors
type ConsoleFormatter struct {
	writer io.Writer
}

// NewConsoleFormatter creates a new ConsoleFormatter
func NewConsoleFormatter(writer io.Writer) *ConsoleFormatter {
	return &ConsoleFormatter{
		writer: writer,
	}
}

// write is a helper that checks write errors and returns early
func (f *ConsoleFormatter) write(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(f.writer, format, args...)
	if err != nil {
		return fmt.Errorf("failed to write to console: %w", err)
	}
	return nil
}

// writeln is a helper that checks write errors and returns early
func (f *ConsoleFormatter) writeln(args ...interface{}) error {
	_, err := fmt.Fprintln(f.writer, args...)
	if err != nil {
		return fmt.Errorf("failed to write line to console: %w", err)
	}
	return nil
}

// Format processes an event and outputs it to the console
func (f *ConsoleFormatter) Format(event events.Event) error {
	var err error
	switch event.Type {
	// Prompt execution events
	case events.EventPromptStarted:
		data := event.Data.(events.PromptStartedData)
		if err = f.writeln("▐ 🪄PROMPT"); err != nil {
			return err
		}
		if err = f.writeln("▐ ─────────"); err != nil {
			return err
		}
		displayPrompt := Truncate(data.Prompt, promptMaxLen, truncateSuffix)
		if err = f.printPrefixed(displayPrompt, "▐ ", displayWidth); err != nil {
			return err
		}
		if err = f.writeln(); err != nil {
			return err
		}

		// Display BaseURL if set
		if data.BaseURL != "" {
			if err = f.write("🌐 ANTHROPIC_BASE_URL: %s\n", data.BaseURL); err != nil {
				return err
			}
		}

		// Display cwd info
		if data.Cwd != "" {
			err = f.write("🚀 Starting(cwd: %s%s)\n", data.Cwd, data.FileList)
		}

	// Claude streaming events
	case events.EventClaudeAssistantMessage:
		data := event.Data.(events.AssistantMessageData)
		err = f.write("💬 %s\n", data.Text)

	case events.EventClaudeToolUse:
		data := event.Data.(events.ToolUseData)
		if err = f.write("🔧 %s%s%s", Bold, data.Name, Reset); err != nil {
			return err
		}
		if len(data.Input) > 0 {
			if err = f.writeln(); err != nil {
				return err
			}
			err = f.formatToolInputs(data.Input)
		} else {
			err = f.writeln()
		}

	case events.EventClaudeToolResult:
		data := event.Data.(events.ToolResultData)
		if data.Content != "" {
			result := f.truncateResult(data.Content, maxResultDisplay)
			err = f.write("✅ %sResult%s: %s\n", Green, Reset, result)
		}

	case events.EventClaudeExecutionResult:
		data := event.Data.(events.ExecutionResultData)
		durationSec := data.Duration.Seconds()
		err = f.write("⏱️ Duration: %.2fs\n", durationSec)

	// Loop execution events
	case events.EventLoopStarted:
		// No output for this event

	case events.EventIterationStarted:
		data := event.Data.(events.IterationStartedData)
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		if err = f.write("Starting iteration %d of %d\n", data.Current, data.Total); err != nil {
			return err
		}
		err = f.writeln("=========================================")

	case events.EventIterationCompleted:
		data := event.Data.(events.IterationCompletedData)
		err = f.write("✅ Iteration %d completed successfully\n", data.Current)

	case events.EventIterationFailed:
		data := event.Data.(events.IterationFailedData)
		if err = f.write("❌ Prompt failed: %v\n", data.Error); err != nil {
			return err
		}
		err = f.write("❌ Iteration %d failed\n", data.Current)

	case events.EventSleepStarted:
		data := event.Data.(events.SleepStartedData)
		err = f.write("💤 Sleeping for %s...\n", data.Duration)

	case events.EventLoopCompleted:
		data := event.Data.(events.LoopCompletedData)
		if data.FailedIterations == 0 {
			err = f.write("🎉 All %d iterations succeeded.\n", data.TotalIterations)
		} else {
			err = f.write("⚠️ %d of %d iterations failed.\n", data.FailedIterations, data.TotalIterations)
		}

	case events.EventLoopInterrupted:
		err = f.writeln("\n\n⚠️ Stopping all iterations...")

	// Evolution events
	case events.EventEvolveStarted:
		data := event.Data.(events.EvolveStartedData)
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		if err = f.writeln("🧬 Starting Evolution"); err != nil {
			return err
		}
		if err = f.write("Iterations: %d\n", data.Iterations); err != nil {
			return err
		}
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		err = f.writeln()

	case events.EventGitBranchCreated:
		data := event.Data.(events.BranchCreatedData)
		if data.Base == "" {
			// Initial branch
			err = f.write("🌱 Creating initial branch: %s\n", data.BranchName)
		} else {
			// Improvement branch
			err = f.write("🌿 Creating improvement branch: %s (from %s)\n", data.BranchName, data.Base)
		}

	case events.EventGitBranchCheckedOut:
		// No output for this event

	case events.EventGitBranchDeleted:
		data := event.Data.(events.BranchDeletedData)
		err = f.write("🗑️ Deleting loser branch: %s\n", data.BranchName)

	case events.EventGitCommitsSquashed:
		data := event.Data.(events.CommitsSquashedData)
		err = f.write("📦 Squashing commits on %s\n", data.BranchName)

	case events.EventRoundStarted:
		data := event.Data.(events.RoundStartedData)
		if err = f.writeln(); err != nil {
			return err
		}
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		if err = f.write("🔄 Evolution Round %d of %d\n", data.Round, data.Total); err != nil {
			return err
		}
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		err = f.writeln()

	case events.EventImprovementStarted:
		if err = f.writeln(); err != nil {
			return err
		}
		if err = f.writeln("✨ Improving code..."); err != nil {
			return err
		}
		err = f.writeln()

	case events.EventComparisonStarted:
		data := event.Data.(events.ComparisonStartedData)
		if err = f.writeln(); err != nil {
			return err
		}
		if err = f.writeln("⚖️ Comparing implementations..."); err != nil {
			return err
		}
		if err = f.write("Branch 1: %s\n", data.Branch1); err != nil {
			return err
		}
		if err = f.write("Branch 2: %s\n", data.Branch2); err != nil {
			return err
		}
		err = f.writeln()

	case events.EventComparisonRetry:
		data := event.Data.(events.ComparisonRetryData)
		err = f.write("⚠️ Retry attempt %d/%d for comparison...\n", data.Attempt, data.MaxAttempts)

	case events.EventWinnerSelected:
		data := event.Data.(events.WinnerSelectedData)
		err = f.write("🏆 Winner: %s\n", data.Winner)

	case events.EventEvolveCompleted:
		data := event.Data.(events.EvolveCompletedData)
		if err = f.writeln(); err != nil {
			return err
		}
		if err = f.writeln("========================================="); err != nil {
			return err
		}
		if err = f.write("🎉 Evolution complete! Final winner: %s\n", data.FinalBranch); err != nil {
			return err
		}
		err = f.writeln("=========================================")

	case events.EventEvolveInterrupted:
		data := event.Data.(events.EvolveInterruptedData)
		err = f.write("\n\n⚠️ Interrupted. Completed %d of %d rounds. Current winner: %s\n",
			data.CompletedRounds, data.TotalRounds, data.Winner)
	}

	return err
}

// Flush ensures all buffered output is written
func (f *ConsoleFormatter) Flush() error {
	return nil
}

// formatToolInputs formats and prints tool inputs with proper alignment
func (f *ConsoleFormatter) formatToolInputs(inputs map[string]interface{}) error {
	keys := make([]string, 0, len(inputs))
	for k := range inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := inputs[key]
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		if err := f.write("     %s%s%s: %s\n", Cyan, key, Reset, valueStr); err != nil {
			return err
		}
	}
	return nil
}

// truncateResult truncates a result string to maxLen if needed
func (f *ConsoleFormatter) truncateResult(result string, maxLen int) string {
	if len(result) > maxLen {
		return result[:maxLen] + "..."
	}
	return result
}

// printPrefixed prints text with left bar prefix, handles line wrapping
func (f *ConsoleFormatter) printPrefixed(text string, prefix string, totalWidth int) error {
	prefixLen := len(prefix)
	contentWidth := totalWidth - prefixLen
	if contentWidth < 1 {
		contentWidth = 1
	}

	wrapped := Wrap(text, contentWidth)
	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		if err := f.writeln(prefix + line); err != nil {
			return err
		}
	}
	return nil
}
