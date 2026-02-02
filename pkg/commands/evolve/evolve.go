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

// EvolutionRunner holds state for the evolution process
type EvolutionRunner struct {
	config         EvolveConfig
	gitClient      *git.Client
	emitter        events.Emitter
	originalBranch string
	currentWinner  string
	sigChan        chan os.Signal
}

// Evolve runs the evolutionary code improvement loop
func Evolve(cfg EvolveConfig, emitter events.Emitter) error {
	runner := &EvolutionRunner{
		config:  cfg,
		emitter: emitter,
	}
	return runner.run()
}

// run orchestrates the entire evolution process
func (r *EvolutionRunner) run() error {
	r.setupSignals()
	defer signal.Stop(r.sigChan)

	r.gitClient = git.NewClient(r.emitter)

	var err error
	r.originalBranch, err = r.gitClient.GetCurrentBranch()
	if err != nil {
		return err
	}

	r.emitter.Emit(events.EventEvolveStarted, events.EvolveStartedData{
		TotalIterations: r.config.Iterations,
	})

	if err := r.checkInterrupted(); err != nil {
		return err
	}

	if err := r.executeInitialPlan(); err != nil {
		return err
	}

	// EVOLUTION LOOP
	for i := 1; i <= r.config.Iterations; i++ {
		if err := r.checkInterrupted(); err != nil {
			r.emitter.Emit(events.EventEvolveInterrupted, events.EvolveInterruptedData{
				CompletedRounds: i - 1,
				TotalRounds:     r.config.Iterations,
				Winner:          r.currentWinner,
			})
			return err
		}

		r.emitter.Emit(events.EventRoundStarted, events.RoundStartedData{
			Round: i,
			Total: r.config.Iterations,
		})

		challenger, err := r.improveWinner(i)
		if err != nil {
			return err
		}

		if err := r.compareAndUpdate(challenger); err != nil {
			return err
		}

		if i < r.config.Iterations && r.config.Sleep > 0 {
			if err := r.waitBetweenRounds(i); err != nil {
				return err
			}
		}
	}

	r.emitter.Emit(events.EventEvolveCompleted, events.EvolveCompletedData{
		FinalBranch:   r.currentWinner,
		TotalRounds:   r.config.Iterations,
		TotalDuration: 0,
	})

	return nil
}

// setupSignals configures signal handling for graceful shutdown
func (r *EvolutionRunner) setupSignals() {
	r.sigChan = make(chan os.Signal, 1)
	signal.Notify(r.sigChan, syscall.SIGINT, syscall.SIGTERM)
}

// checkInterrupted checks if an interrupt signal was received
func (r *EvolutionRunner) checkInterrupted() error {
	select {
	case <-r.sigChan:
		return fmt.Errorf("interrupted")
	default:
		return nil
	}
}

// executeInitialPlan creates and runs the initial implementation
func (r *EvolutionRunner) executeInitialPlan() error {
	branchA := git.RandomBranchName()

	if err := r.gitClient.CreateBranch(branchA); err != nil {
		return err
	}

	planOpts := &claude.PromptOptions{
		SystemPrompt:       r.config.PlanSystemPrompt,
		AppendSystemPrompt: r.config.PlanAppendSystemPrompt,
	}
	if _, err := claude.RunPrompt(r.config.Plan, planOpts, r.emitter); err != nil {
		return err
	}

	if err := r.gitClient.SquashCommits(r.originalBranch, "implement: "+truncate(r.config.Plan, 50)); err != nil {
		return err
	}

	r.currentWinner = branchA
	return nil
}

// improveWinner creates an improvement branch and runs the improvement prompt
func (r *EvolutionRunner) improveWinner(roundNum int) (string, error) {
	challenger := git.RandomBranchName()

	if err := r.gitClient.CreateBranchFrom(challenger, r.currentWinner); err != nil {
		return "", err
	}

	r.emitter.Emit(events.EventImprovementStarted, events.ImprovementStartedData{
		BranchName: challenger,
	})

	improveOpts := &claude.PromptOptions{
		SystemPrompt:       r.config.ImproveSystemPrompt,
		AppendSystemPrompt: r.config.ImproveAppendSystemPrompt,
	}
	if _, err := claude.RunPrompt(r.config.ImprovePrompt, improveOpts, r.emitter); err != nil {
		return "", err
	}

	if err := r.gitClient.SquashCommits(r.originalBranch, "improve: round "+fmt.Sprint(roundNum)); err != nil {
		return "", err
	}

	return challenger, nil
}

// compareAndUpdate compares branches and updates the winner
func (r *EvolutionRunner) compareAndUpdate(challenger string) error {
	r.emitter.Emit(events.EventComparisonStarted, events.ComparisonStartedData{
		Branch1: r.currentWinner,
		Branch2: challenger,
	})

	comparePrompt := fmt.Sprintf("%s\n\nBranch names to compare:\n- %s\n- %s\n\nRespond with ONLY the branch name that should be DELETED (the worse one).",
		r.config.ComparePrompt, r.currentWinner, challenger)

	if err := r.gitClient.Checkout(r.originalBranch); err != nil {
		return err
	}

	compareOpts := &claude.PromptOptions{
		SystemPrompt:       r.config.CompareSystemPrompt,
		AppendSystemPrompt: r.config.CompareAppendSystemPrompt,
	}

	var loser string
	var err error
	for attempt := 0; attempt <= r.config.CompareErrorRetries; attempt++ {
		if attempt > 0 {
			r.emitter.Emit(events.EventComparisonRetry, events.ComparisonRetryData{
				Attempt:     attempt,
				MaxAttempts: r.config.CompareErrorRetries,
			})
		}

		result, runErr := claude.RunPrompt(comparePrompt, compareOpts, r.emitter)
		if runErr != nil {
			return runErr
		}

		loser, err = parseBranchFromResponse(result, r.currentWinner, challenger)
		if err == nil {
			break
		}

		if attempt == r.config.CompareErrorRetries {
			return fmt.Errorf("failed to parse comparison result after %d retries: %w", r.config.CompareErrorRetries, err)
		}
	}

	if loser == r.currentWinner {
		r.currentWinner = challenger
	}

	r.emitter.Emit(events.EventWinnerSelected, events.WinnerSelectedData{
		Winner: r.currentWinner,
		Loser:  loser,
	})

	if err := r.gitClient.Checkout(r.currentWinner); err != nil {
		return err
	}

	if err := r.gitClient.DeleteBranch(loser); err != nil {
		return err
	}

	return nil
}

// waitBetweenRounds implements interruptible sleep between evolution rounds
func (r *EvolutionRunner) waitBetweenRounds(completedRound int) error {
	r.emitter.Emit(events.EventSleepStarted, events.SleepStartedData{
		Duration: r.config.Sleep,
	})

	timer := time.NewTimer(r.config.Sleep)
	select {
	case <-r.sigChan:
		timer.Stop()
		r.emitter.Emit(events.EventEvolveInterrupted, events.EvolveInterruptedData{
			CompletedRounds: completedRound,
			TotalRounds:     r.config.Iterations,
			Winner:          r.currentWinner,
		})
		return fmt.Errorf("interrupted")
	case <-timer.C:
		return nil
	}
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
