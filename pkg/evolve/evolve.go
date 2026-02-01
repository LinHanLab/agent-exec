package evolve

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/git"
)

// EvolveConfig holds configuration for the evolution process
type EvolveConfig struct {
	Plan          string        // Initial implementation prompt
	ImprovePrompt string        // Prompt for improvement step
	ComparePrompt string        // Prompt for comparison step
	Iterations    int           // Number of evolution iterations
	Sleep         time.Duration // Sleep duration between evolution rounds

	// System prompts for each step
	PlanSystemPrompt       string
	PlanAppendSystemPrompt string

	ImproveSystemPrompt       string
	ImproveAppendSystemPrompt string

	CompareSystemPrompt       string
	CompareAppendSystemPrompt string
}

// Evolve runs the evolutionary code improvement loop
func Evolve(cfg EvolveConfig) (string, error) {
	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Save original branch to return to on error
	originalBranch, err := git.GetCurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	fmt.Println("=========================================")
	fmt.Println("🧬 Starting Evolution")
	fmt.Printf("   Iterations: %d\n", cfg.Iterations)
	fmt.Printf("   Base branch: %s\n", originalBranch)
	fmt.Println("=========================================")
	fmt.Println()

	// Check for interrupt
	select {
	case <-sigChan:
		return "", fmt.Errorf("interrupted")
	default:
	}

	// INITIAL: Create first implementation
	branchA := git.RandomBranchName()
	fmt.Printf("🌱 Creating initial branch: %s\n", branchA)

	if err := git.CreateBranch(branchA); err != nil {
		return "", fmt.Errorf("failed to create initial branch: %w", err)
	}

	fmt.Println()
	fmt.Println("🔨 Implementing plan...")
	fmt.Println()

	planOpts := &claude.PromptOptions{
		SystemPrompt:       cfg.PlanSystemPrompt,
		AppendSystemPrompt: cfg.PlanAppendSystemPrompt,
	}
	if _, err := claude.RunPrompt(cfg.Plan, planOpts); err != nil {
		return "", fmt.Errorf("initial implementation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("📦 Squashing commits on %s\n", branchA)
	if err := git.SquashCommits(originalBranch, "implement: "+truncate(cfg.Plan, 50)); err != nil {
		return "", fmt.Errorf("failed to squash: %w", err)
	}

	winner := branchA

	// EVOLUTION LOOP
	for i := 1; i <= cfg.Iterations; i++ {
		// Check for interrupt
		select {
		case <-sigChan:
			fmt.Println("\n\n⚠️  Interrupted. Current winner:", winner)
			return winner, fmt.Errorf("interrupted")
		default:
		}

		fmt.Println()
		fmt.Println("=========================================")
		fmt.Printf("🔄 Evolution Round %d of %d\n", i, cfg.Iterations)
		fmt.Println("=========================================")
		fmt.Println()

		// Create improvement branch from winner
		branchB := git.RandomBranchName()
		fmt.Printf("🌿 Creating improvement branch: %s (from %s)\n", branchB, winner)

		if err := git.CreateBranchFrom(branchB, winner); err != nil {
			return winner, fmt.Errorf("failed to create improvement branch: %w", err)
		}

		fmt.Println()
		fmt.Println("✨ Improving code...")
		fmt.Println()

		improveOpts := &claude.PromptOptions{
			SystemPrompt:       cfg.ImproveSystemPrompt,
			AppendSystemPrompt: cfg.ImproveAppendSystemPrompt,
		}
		if _, err := claude.RunPrompt(cfg.ImprovePrompt, improveOpts); err != nil {
			return winner, fmt.Errorf("improvement failed: %w", err)
		}

		fmt.Println()
		fmt.Printf("📦 Squashing commits on %s\n", branchB)
		if err := git.SquashCommits(originalBranch, "improve: round "+fmt.Sprint(i)); err != nil {
			return winner, fmt.Errorf("failed to squash improvement: %w", err)
		}

		// Compare branches
		fmt.Println()
		fmt.Println("⚖️ Comparing implementations...")
		fmt.Printf("   Branch 1: %s\n", winner)
		fmt.Printf("   Branch 2: %s\n", branchB)
		fmt.Println()

		// Build comparison prompt with branch names
		comparePrompt := fmt.Sprintf("%s\n\nBranch names to compare:\n- %s\n- %s\n\nRespond with ONLY the branch name that should be DELETED (the worse one).",
			cfg.ComparePrompt, winner, branchB)

		// Switch to original branch for comparison (neutral ground)
		if err := git.Checkout(originalBranch); err != nil {
			return winner, fmt.Errorf("failed to checkout base for comparison: %w", err)
		}

		compareOpts := &claude.PromptOptions{
			SystemPrompt:       cfg.CompareSystemPrompt,
			AppendSystemPrompt: cfg.CompareAppendSystemPrompt,
		}
		result, err := claude.RunPrompt(comparePrompt, compareOpts)
		if err != nil {
			return winner, fmt.Errorf("comparison failed: %w", err)
		}

		// Parse the loser branch from Claude's response
		loser := parseBranchFromResponse(result, winner, branchB)

		// Update winner
		if loser == winner {
			winner = branchB
		}
		fmt.Printf("🏆 Winner: %s\n", winner)

		// Checkout winner for next iteration
		if err := git.Checkout(winner); err != nil {
			return winner, fmt.Errorf("failed to checkout winner: %w", err)
		}

		fmt.Printf("\n🗑️  Deleting loser branch: %s\n", loser)
		if err := git.DeleteBranch(loser); err != nil {
			return winner, fmt.Errorf("failed to delete loser: %w", err)
		}

		// Sleep between evolution rounds (skip after last iteration)
		if i < cfg.Iterations && cfg.Sleep > 0 {
			fmt.Printf("💤 Sleeping for %s before next evolution round...\n", cfg.Sleep)

			timer := time.NewTimer(cfg.Sleep)
			select {
			case <-sigChan:
				timer.Stop()
				fmt.Println("\n\n⚠️  Interrupted. Current winner:", winner)
				return winner, fmt.Errorf("interrupted")
			case <-timer.C:
			}
		}
	}

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Printf("🎉 Evolution complete! Final winner: %s\n", winner)
	fmt.Println("=========================================")

	return winner, nil
}

// parseBranchFromResponse extracts the loser branch name from Claude's response
func parseBranchFromResponse(response, branch1, branch2 string) string {
	response = strings.TrimSpace(response)

	// Check if response contains either branch name
	contains1 := strings.Contains(response, branch1)
	contains2 := strings.Contains(response, branch2)

	if contains1 && !contains2 {
		return branch1
	}
	if contains2 && !contains1 {
		return branch2
	}

	// If both or neither found, check last line (Claude was asked to respond with just the name)
	lines := strings.Split(response, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])

	if lastLine == branch1 {
		return branch1
	}
	if lastLine == branch2 {
		return branch2
	}

	// Default to branch2 (the newer one) if parsing fails
	fmt.Printf("⚠️  Could not parse loser from response, defaulting to: %s\n", branch2)
	return branch2
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
