package claude

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/format"
)

const (
	DisplayWidth   = 76
	PromptMaxLen   = 270
	TruncateSuffix = "[...Truncated]"
)

// RunPrompt executes a single prompt with claude CLI
func RunPrompt(prompt string) error {
	if err := ValidatePrompt(prompt); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	fmt.Println("▐ 🪄PROMPT")
	fmt.Println("▐ ─────────")

	displayPrompt := format.Truncate(prompt, PromptMaxLen, TruncateSuffix)
	format.PrintPrefixed(displayPrompt, "▐ ", DisplayWidth)

	fmt.Println()

	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		fmt.Printf("🌐 ANTHROPIC_BASE_URL: %s\n", baseURL)
		fmt.Println()
	}

	fmt.Println("🚀 Starting...")
	fmt.Println()

	cmd := exec.Command("claude", "--verbose", "--output-format", "stream-json", "-p", prompt)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude CLI: %w", err)
	}

	if err := ParseStreamJSON(stdout); err != nil {
		_ = cmd.Wait()
		return fmt.Errorf("failed to parse output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude CLI failed: %w", err)
	}

	return nil
}

// RunPromptLoop executes a prompt in iterations with configurable sleep
func RunPromptLoop(iterations int, sleep time.Duration, prompt string) error {
	if err := ValidateLoopArgs(iterations, prompt); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	failedIterations := 0

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Run the iteration loop
	for i := 1; i <= iterations; i++ {
		// Check for interrupt before starting iteration
		select {
		case <-sigChan:
			fmt.Println("\n\n⚠️  Stopping all iterations...")
			return fmt.Errorf("interrupted")
		default:
		}

		fmt.Println("=========================================")
		fmt.Printf("Starting iteration %d of %d\n", i, iterations)
		fmt.Println("=========================================")

		// Execute prompt
		if err := RunPrompt(prompt); err != nil {
			fmt.Printf("❌ Prompt failed: %v\n", err)
			fmt.Printf("❌ Iteration %d failed\n", i)
			failedIterations++
		} else {
			fmt.Printf("✅ Iteration %d completed successfully\n", i)
		}

		// Sleep between iterations (skip sleep after last iteration)
		if i < iterations && sleep > 0 {
			fmt.Printf("💤 Sleeping for %s...\n", sleep)

			// Interruptible sleep
			timer := time.NewTimer(sleep)
			select {
			case <-sigChan:
				timer.Stop()
				fmt.Println("\n\n⚠️  Stopping all iterations...")
				return fmt.Errorf("interrupted")
			case <-timer.C:
			}
		}

		fmt.Println()
	}

	// Print completion summary
	if failedIterations == 0 {
		fmt.Printf("🎉 All %d iterations succeeded.\n", iterations)
	} else {
		fmt.Printf("⚠️  %d of %d iterations failed.\n", failedIterations, iterations)
	}

	return nil
}
