package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Run multiple prompts in iterations with configurable sleep between
func runIterations(iterations int, sleepSeconds int, promptFiles []string) error {
	if err := validateIterationArgs(iterations, sleepSeconds, promptFiles); err != nil {
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

		// Execute prompts with fail-fast within iteration
		iterationFailed := false
		for _, promptFile := range promptFiles {
			// Check for interrupt before each prompt
			select {
			case <-sigChan:
				fmt.Println("\n\n⚠️  Stopping all iterations...")
				return fmt.Errorf("interrupted")
			default:
			}

			prompt, err := readPromptFile(promptFile)
			if err != nil {
				fmt.Printf("❌ Failed to read prompt file %s: %v\n", promptFile, err)
				iterationFailed = true
				break
			}

			if err := runOneShot(prompt); err != nil {
				fmt.Printf("❌ Prompt failed: %v\n", err)
				iterationFailed = true
				break
			}
		}

		if iterationFailed {
			fmt.Printf("❌ Iteration %d failed (one or more prompts failed)\n", i)
			failedIterations++
		} else {
			fmt.Printf("✅ Iteration %d completed successfully\n", i)
		}

		// Sleep between iterations (skip sleep after last iteration)
		if i < iterations && sleepSeconds > 0 {
			fmt.Printf("💤 Sleeping for %d seconds...\n", sleepSeconds)

			// Interruptible sleep
			timer := time.NewTimer(time.Duration(sleepSeconds) * time.Second)
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

// Read and validate prompt file contents
func readPromptFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	prompt := string(content)
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt file is empty or whitespace-only")
	}

	return prompt, nil
}
