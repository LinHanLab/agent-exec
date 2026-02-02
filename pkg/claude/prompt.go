package claude

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// PromptOptions holds optional configuration for running prompts
type PromptOptions struct {
	SystemPrompt       string // Replace entire system prompt (empty = use defaults)
	AppendSystemPrompt string // Append to default system prompt (empty = use defaults)
}

// BuildClaudeArgs constructs the claude CLI arguments based on options
func (opts *PromptOptions) BuildClaudeArgs(prompt string) []string {
	args := []string{"--verbose", "--output-format", "stream-json", "-p", prompt}

	if opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	return args
}

// getCwdInfo retrieves current working directory and file list with error handling
func getCwdInfo(emitter events.Emitter) (cwd, fileList string, err error) {
	cwd, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get cwd: %w", err)
	}

	files, err := os.ReadDir(cwd)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cwd files: %w", err)
	}

	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	fileList = " [" + strings.Join(names, ", ") + "]"

	return
}

// RunPrompt executes a single prompt with claude CLI and returns the final result text
func RunPrompt(prompt string, opts *PromptOptions, emitter events.Emitter) (string, error) {
	if err := ValidatePrompt(prompt); err != nil {
		return "", err
	}

	cwd, fileList, err := getCwdInfo(emitter)
	if err != nil {
		return "", err
	}

	emitter.Emit(events.EventPromptStarted, events.PromptStartedData{
		Prompt:   prompt,
		BaseURL:  os.Getenv("ANTHROPIC_BASE_URL"),
		Cwd:      cwd,
		FileList: fileList,
	})

	if opts == nil {
		opts = &PromptOptions{}
	}
	args := opts.BuildClaudeArgs(prompt)
	cmd := exec.Command("claude", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude CLI: %w", err)
	}

	result, parseErr := ParseStreamJSON(stdout, emitter)
	if parseErr != nil {
		_ = cmd.Wait()
		return "", parseErr
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}

	return result, nil
}

// RunPromptLoop executes a prompt in iterations with configurable sleep
func RunPromptLoop(iterations int, sleep time.Duration, prompt string, opts *PromptOptions, emitter events.Emitter) error {
	if err := ValidateLoopArgs(iterations, prompt); err != nil {
		return err
	}

	failedIterations := 0

	if opts == nil {
		opts = &PromptOptions{}
	}

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	emitter.Emit(events.EventLoopStarted, events.LoopStartedData{
		TotalIterations: iterations,
	})

	// Run the iteration loop
	for i := 1; i <= iterations; i++ {
		// Check for interrupt before starting iteration
		select {
		case <-sigChan:
			emitter.Emit(events.EventLoopInterrupted, events.LoopInterruptedData{
				CompletedIterations: i - 1,
				TotalIterations:     iterations,
			})
			return fmt.Errorf("interrupted")
		default:
		}

		emitter.Emit(events.EventIterationStarted, events.IterationStartedData{
			Current: i,
			Total:   iterations,
		})

		// Execute prompt
		startTime := time.Now()
		if _, err := RunPrompt(prompt, opts, emitter); err != nil {
			emitter.Emit(events.EventIterationFailed, events.IterationFailedData{
				Current: i,
				Total:   iterations,
				Error:   err,
			})
			failedIterations++
		} else {
			duration := time.Since(startTime)
			emitter.Emit(events.EventIterationCompleted, events.IterationCompletedData{
				Current:  i,
				Total:    iterations,
				Duration: duration,
			})
		}

		// Sleep between iterations (skip sleep after last iteration)
		if i < iterations && sleep > 0 {
			emitter.Emit(events.EventSleepStarted, events.SleepStartedData{
				Duration: sleep,
			})

			// Interruptible sleep
			timer := time.NewTimer(sleep)
			select {
			case <-sigChan:
				timer.Stop()
				emitter.Emit(events.EventLoopInterrupted, events.LoopInterruptedData{
					CompletedIterations: i,
					TotalIterations:     iterations,
				})
				return fmt.Errorf("interrupted")
			case <-timer.C:
			}
		}
	}

	// Print completion summary
	emitter.Emit(events.EventLoopCompleted, events.LoopCompletedData{
		TotalIterations:      iterations,
		SuccessfulIterations: iterations - failedIterations,
		FailedIterations:     failedIterations,
		TotalDuration:        0, // Not tracking total duration for now
	})

	return nil
}
