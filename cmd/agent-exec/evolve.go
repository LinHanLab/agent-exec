package main

import (
	"fmt"
	"os"

	"github.com/LinHanLab/agent-exec/pkg/evolve"
	"github.com/spf13/cobra"
)

var (
	improvePrompt string
	comparePrompt string
	evolveIters   int
)

var evolveCmd = &cobra.Command{
	Use:   "evolve <plan>",
	Short: "Evolve code through iterative improvement with competing branches",
	Long: `Evolve code through a tournament-style iterative improvement process.

This command:
1. Creates an initial implementation from the plan
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
}
