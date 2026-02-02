package loop

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/events"
)

// ValidateLoopArgs validates iteration arguments
func ValidateLoopArgs(iterations int, prompt string) error {
	if iterations < 1 {
		return errors.New("iterations must be a positive number")
	}

	return claude.ValidatePrompt(prompt)
}

// RunPromptLoop executes a prompt in iterations with configurable sleep
func RunPromptLoop(iterations int, sleep time.Duration, prompt string, opts *claude.PromptOptions, emitter events.Emitter) error {
	if err := ValidateLoopArgs(iterations, prompt); err != nil {
		return err
	}

	failedIterations := 0

	if opts == nil {
		opts = &claude.PromptOptions{}
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
		if _, err := claude.RunPrompt(prompt, opts, emitter); err != nil {
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
