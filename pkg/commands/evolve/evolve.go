package evolve

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/LinHanLab/agent-exec/pkg/git"
)

// EvolveConfig holds configuration for the evolution process
type EvolveConfig struct {
	Plan                string        // Initial implementation prompt
	ImprovePrompt       string        // Prompt for improvement step
	ComparePrompt       string        // Prompt for comparison step
	Iterations          int           // Number of evolution iterations
	Sleep               time.Duration // Sleep duration between evolution rounds
	CompareErrorRetries int           // Number of retries when comparison parsing fails

	// System prompts for each step
	PlanSystemPrompt       string
	PlanAppendSystemPrompt string

	ImproveSystemPrompt       string
	ImproveAppendSystemPrompt string

	CompareSystemPrompt       string
	CompareAppendSystemPrompt string
}

// Evolve runs the evolutionary code improvement loop
func Evolve(cfg EvolveConfig, emitter events.Emitter) error {
	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Create git client with emitter
	gitClient := git.NewClient(emitter)

	// Save original branch to return to on error
	originalBranch, err := gitClient.GetCurrentBranch()
	if err != nil {
		return err
	}

	emitter.Emit(events.EventEvolveStarted, events.EvolveStartedData{
		TotalIterations: cfg.Iterations,
	})

	// Check for interrupt
	select {
	case <-sigChan:
		return fmt.Errorf("interrupted")
	default:
	}

	// INITIAL: Create first implementation
	branchA := git.RandomBranchName()

	if err := gitClient.CreateBranch(branchA); err != nil {
		return err
	}

	planOpts := &claude.PromptOptions{
		SystemPrompt:       cfg.PlanSystemPrompt,
		AppendSystemPrompt: cfg.PlanAppendSystemPrompt,
	}
	if _, err := claude.RunPrompt(cfg.Plan, planOpts, emitter); err != nil {
		return err
	}

	if err := gitClient.SquashCommits(originalBranch, "implement: "+truncate(cfg.Plan, 50)); err != nil {
		return err
	}

	winner := branchA

	// EVOLUTION LOOP
	for i := 1; i <= cfg.Iterations; i++ {
		// Check for interrupt
		select {
		case <-sigChan:
			emitter.Emit(events.EventEvolveInterrupted, events.EvolveInterruptedData{
				CompletedRounds: i - 1,
				TotalRounds:     cfg.Iterations,
				Winner:          winner,
			})
			return fmt.Errorf("interrupted")
		default:
		}

		emitter.Emit(events.EventRoundStarted, events.RoundStartedData{
			Round: i,
			Total: cfg.Iterations,
		})

		// Create improvement branch from winner
		branchB := git.RandomBranchName()

		if err := gitClient.CreateBranchFrom(branchB, winner); err != nil {
			return err
		}

		emitter.Emit(events.EventImprovementStarted, events.ImprovementStartedData{
			BranchName: branchB,
		})

		improveOpts := &claude.PromptOptions{
			SystemPrompt:       cfg.ImproveSystemPrompt,
			AppendSystemPrompt: cfg.ImproveAppendSystemPrompt,
		}
		if _, err := claude.RunPrompt(cfg.ImprovePrompt, improveOpts, emitter); err != nil {
			return err
		}

		if err := gitClient.SquashCommits(originalBranch, "improve: round "+fmt.Sprint(i)); err != nil {
			return err
		}

		// Compare branches
		emitter.Emit(events.EventComparisonStarted, events.ComparisonStartedData{
			Branch1: winner,
			Branch2: branchB,
		})

		// Build comparison prompt with branch names
		comparePrompt := fmt.Sprintf("%s\n\nBranch names to compare:\n- %s\n- %s\n\nRespond with ONLY the branch name that should be DELETED (the worse one).",
			cfg.ComparePrompt, winner, branchB)

		// Switch to original branch for comparison (neutral ground)
		if err := gitClient.Checkout(originalBranch); err != nil {
			return err
		}

		compareOpts := &claude.PromptOptions{
			SystemPrompt:       cfg.CompareSystemPrompt,
			AppendSystemPrompt: cfg.CompareAppendSystemPrompt,
		}

		// Retry comparison if parsing fails
		var loser string
		var result string
		var err error
		for attempt := 0; attempt <= cfg.CompareErrorRetries; attempt++ {
			if attempt > 0 {
				emitter.Emit(events.EventComparisonRetry, events.ComparisonRetryData{
					Attempt:     attempt,
					MaxAttempts: cfg.CompareErrorRetries,
				})
			}

			result, err = claude.RunPrompt(comparePrompt, compareOpts, emitter)
			if err != nil {
				return err
			}

			// Try to parse the loser branch from Claude's response
			loser, err = parseBranchFromResponse(result, winner, branchB)
			if err == nil {
				break // Successfully parsed
			}

			if attempt == cfg.CompareErrorRetries {
				return fmt.Errorf("failed to parse comparison result after %d retries: %w", cfg.CompareErrorRetries, err)
			}
		}

		// Update winner
		if loser == winner {
			winner = branchB
		}
		emitter.Emit(events.EventWinnerSelected, events.WinnerSelectedData{
			Winner: winner,
			Loser:  loser,
		})

		// Checkout winner for next iteration
		if err := gitClient.Checkout(winner); err != nil {
			return err
		}

		if err := gitClient.DeleteBranch(loser); err != nil {
			return err
		}

		// Sleep between evolution rounds (skip after last iteration)
		if i < cfg.Iterations && cfg.Sleep > 0 {
			emitter.Emit(events.EventSleepStarted, events.SleepStartedData{
				Duration: cfg.Sleep,
			})

			timer := time.NewTimer(cfg.Sleep)
			select {
			case <-sigChan:
				timer.Stop()
				emitter.Emit(events.EventEvolveInterrupted, events.EvolveInterruptedData{
					CompletedRounds: i,
					TotalRounds:     cfg.Iterations,
					Winner:          winner,
				})
				return fmt.Errorf("interrupted")
			case <-timer.C:
			}
		}
	}

	emitter.Emit(events.EventEvolveCompleted, events.EvolveCompletedData{
		FinalBranch:   winner,
		TotalRounds:   cfg.Iterations,
		TotalDuration: 0, // Not tracking total duration for now
	})

	return nil
}

// parseBranchFromResponse extracts the loser branch name from Claude's response
func parseBranchFromResponse(response, branch1, branch2 string) (string, error) {
	response = strings.TrimSpace(response)

	// Check if response contains either branch name
	contains1 := strings.Contains(response, branch1)
	contains2 := strings.Contains(response, branch2)

	if contains1 && !contains2 {
		return branch1, nil
	}
	if contains2 && !contains1 {
		return branch2, nil
	}

	// If both or neither found, check last line (Claude was asked to respond with just the name)
	lines := strings.Split(response, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])

	if lastLine == branch1 {
		return branch1, nil
	}
	if lastLine == branch2 {
		return branch2, nil
	}

	// Return error if parsing fails
	return "", fmt.Errorf("could not parse loser branch from response")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
