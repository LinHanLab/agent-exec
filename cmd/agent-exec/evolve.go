package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/commands/evolve"
	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/spf13/cobra"
)

var (
	improvePrompt       string
	comparePrompt       string
	evolveIters         int
	evolveSleep         time.Duration
	compareErrorRetries int

	planSystemPrompt       string
	planAppendSystemPrompt string

	improveSystemPrompt       string
	improveAppendSystemPrompt string

	compareSystemPrompt       string
	compareAppendSystemPrompt string

	evolveVerbose     bool
	debugKeepBranches bool
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
			Plan:                plan,
			ImprovePrompt:       improvePrompt,
			ComparePrompt:       comparePrompt,
			Iterations:          evolveIters,
			Sleep:               evolveSleep,
			CompareErrorRetries: compareErrorRetries,
			DebugKeepBranches:   debugKeepBranches,

			PlanSystemPrompt:       planSystemPrompt,
			PlanAppendSystemPrompt: planAppendSystemPrompt,

			ImproveSystemPrompt:       improveSystemPrompt,
			ImproveAppendSystemPrompt: improveAppendSystemPrompt,

			CompareSystemPrompt:       compareSystemPrompt,
			CompareAppendSystemPrompt: compareAppendSystemPrompt,
		}

		// Create emitter and display
		emitter := events.NewChannelEmitter(100)
		formatter := display.NewConsoleFormatter(os.Stdout, evolveVerbose)
		disp := display.NewDisplay(formatter, emitter)
		disp.Start()

		err := evolve.Evolve(cfg, emitter)

		// Close emitter and wait for display to finish
		emitter.Close()
		disp.Wait()

		if err != nil {
			if err.Error() == "interrupted" {
				os.Exit(130)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(evolveCmd)

	evolveCmd.Flags().StringVarP(&improvePrompt, "improve", "i", "improve the code quality and fix any issues", "prompt for improvement agent")
	evolveCmd.Flags().StringVarP(&comparePrompt, "compare", "c", "compare these two implementations and determine which is worse", "prompt for comparison agent")
	evolveCmd.Flags().IntVarP(&evolveIters, "iterations", "n", 3, "number of evolution iterations")
	evolveCmd.Flags().DurationVarP(&evolveSleep, "sleep", "s", 0, "sleep duration between evolution rounds (e.g., 2h30m, 10s)")
	evolveCmd.Flags().IntVar(&compareErrorRetries, "compare-error-retries", 3, "number of retries when comparison parsing fails")

	evolveCmd.Flags().StringVar(&planSystemPrompt, "system-prompt", "", "replace system prompt for plan agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&planAppendSystemPrompt, "append-system-prompt", "", "append to system prompt for plan agent (empty = use Claude Code defaults)")

	evolveCmd.Flags().StringVar(&improveSystemPrompt, "improve-system-prompt", "", "replace system prompt for improve agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&improveAppendSystemPrompt, "append-improve-system-prompt", "", "append to system prompt for improve agent (empty = use Claude Code defaults)")

	evolveCmd.Flags().StringVar(&compareSystemPrompt, "compare-system-prompt", "", "replace system prompt for compare agent (empty = use Claude Code defaults)")
	evolveCmd.Flags().StringVar(&compareAppendSystemPrompt, "append-compare-system-prompt", "", "append to system prompt for compare agent (empty = use Claude Code defaults)")

	evolveCmd.Flags().BoolVarP(&evolveVerbose, "verbose", "v", false, "show full content without truncation")
	evolveCmd.Flags().BoolVar(&debugKeepBranches, "debug-keep-branches", false, "debug mode: keep all branches instead of deleting losers")
}
