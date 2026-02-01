package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/evolve"
	"github.com/spf13/cobra"
)

var (
	improvePrompt string
	comparePrompt string
	evolveIters   int
	evolveSleep   time.Duration

	planSystemPrompt       string
	planAppendSystemPrompt string

	improveSystemPrompt       string
	improveAppendSystemPrompt string

	compareSystemPrompt       string
	compareAppendSystemPrompt string
)

var evolveCmd = &cobra.Command{
	Use:   "evolve <prompt>",
	Short: "Evolve code through iterative improvement with competing branches",
	Long: `Evolve code through a tournament-style iterative improvement process.

This command:
1. Creates an initial implementation from the prompt
2. Generates improved versions on competing branches
3. Compares branches and eliminates the weaker one
4. Repeats for N iterations

Each branch is squashed to a single commit relative to the base branch.
Branch names are randomized to prevent bias in comparison.

Examples:
  agent-exec evolve "implement a REST API for users" --improve "improve code quality" --compare "which implementation is better?"
  agent-exec evolve "build a CLI tool" -i "refactor for clarity" -c "compare implementations" -n 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		plan := args[0]

		cfg := evolve.EvolveConfig{
			Plan:          plan,
			ImprovePrompt: improvePrompt,
			ComparePrompt: comparePrompt,
			Iterations:    evolveIters,
			Sleep:         evolveSleep,

			PlanSystemPrompt:       planSystemPrompt,
			PlanAppendSystemPrompt: planAppendSystemPrompt,

			ImproveSystemPrompt:       improveSystemPrompt,
			ImproveAppendSystemPrompt: improveAppendSystemPrompt,

			CompareSystemPrompt:       compareSystemPrompt,
			CompareAppendSystemPrompt: compareAppendSystemPrompt,
		}

		winner, err := evolve.Evolve(cfg)
		if err != nil {
			if err.Error() == "interrupted" {
				fmt.Printf("\nInterrupted. Winner so far: %s\n", winner)
				os.Exit(130)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nFinal branch: %s\n", winner)
	},
}

func init() {
	rootCmd.AddCommand(evolveCmd)

	evolveCmd.Flags().StringVarP(&improvePrompt, "improve", "i", "improve the code quality and fix any issues", "prompt for improvement agent")
	evolveCmd.Flags().StringVarP(&comparePrompt, "compare", "c", "compare these two implementations and determine which is worse", "prompt for comparison agent")
	evolveCmd.Flags().IntVarP(&evolveIters, "iterations", "n", 3, "number of evolution iterations")
	evolveCmd.Flags().DurationVarP(&evolveSleep, "sleep", "s", 0, "sleep duration between evolution rounds (e.g., 2h30m, 10s)")

	evolveCmd.Flags().StringVar(&planSystemPrompt, "system-prompt", "", "replace system prompt for plan agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&planAppendSystemPrompt, "append-system-prompt", "", "append to system prompt for plan agent (empty = use Claude Code defaults)")

	evolveCmd.Flags().StringVar(&improveSystemPrompt, "improve-system-prompt", "", "replace system prompt for improve agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&improveAppendSystemPrompt, "append-improve-system-prompt", "", "append to system prompt for improve agent (empty = use Claude Code defaults)")

	evolveCmd.Flags().StringVar(&compareSystemPrompt, "compare-system-prompt", "", "replace system prompt for compare agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&compareAppendSystemPrompt, "append-compare-system-prompt", "", "append to system prompt for compare agent (empty = use Claude Code defaults)")
}
