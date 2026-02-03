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
	Short: "Tournament-style code evolution using competing git branches",
	Long: `Tournament-style code evolution using competing git branches.

Creates an initial implementation, then iteratively improves by:
1. Creating a challenger branch with improvements
2. Using AI comparison to select the better implementation
3. Eliminating the worse branch and continuing with the winner

The final winning branch contains the best evolved implementation.

Example:
  agent-exec evolve "implement user authentication" -n 5 -s 1m`,
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

	evolveCmd.Flags().StringVarP(&improvePrompt, "improve", "i", "improve the code quality and fix any issues", "Prompt for creating improved challenger implementations")
	evolveCmd.Flags().StringVarP(&comparePrompt, "compare", "c", "compare these two implementations and determine which is worse", "Prompt for comparing and selecting worse implementation")
	evolveCmd.Flags().IntVarP(&evolveIters, "iterations", "n", 3, "Number of evolution rounds to run")
	evolveCmd.Flags().DurationVarP(&evolveSleep, "sleep", "s", 0, "Sleep duration between evolution rounds (e.g., 30s, 1m)")
	evolveCmd.Flags().IntVar(&compareErrorRetries, "compare-error-retries", 3, "Retry attempts when comparison branch parsing fails")

	evolveCmd.Flags().StringVar(&planSystemPrompt, "system-prompt", "", "Replace entire system prompt for initial plan step")
	evolveCmd.Flags().StringVar(&planAppendSystemPrompt, "append-system-prompt", "", "Append to default system prompt for initial plan step")

	evolveCmd.Flags().StringVar(&improveSystemPrompt, "improve-system-prompt", "", "Replace entire system prompt for improvement steps")
	evolveCmd.Flags().StringVar(&improveAppendSystemPrompt, "append-improve-system-prompt", "", "Append to default system prompt for improvement steps")

	evolveCmd.Flags().StringVar(&compareSystemPrompt, "compare-system-prompt", "", "Replace entire system prompt for comparison steps")
	evolveCmd.Flags().StringVar(&compareAppendSystemPrompt, "append-compare-system-prompt", "", "Append to default system prompt for comparison steps")

	evolveCmd.Flags().BoolVarP(&evolveVerbose, "verbose", "v", false, "Show verbose output including all Claude events")
	evolveCmd.Flags().BoolVar(&debugKeepBranches, "debug-keep-branches", false, "Keep all branches for debugging instead of deleting losers")
}
