package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/commands/evolve"
	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/LinHanLab/agent-exec/pkg/git"
	"github.com/spf13/cobra"
)

var (
	improvePrompt       string
	comparePrompt       string
	evolveIters         int
	evolveSleep         time.Duration
	compareErrorRetries int

	evolveSystemPrompt       string
	evolveAppendSystemPrompt string

	improveSystemPrompt       string
	improveAppendSystemPrompt string

	compareSystemPrompt       string
	compareAppendSystemPrompt string

	evolveVerbose     bool
	debugKeepBranches bool
	evolveStatusLine  bool
)

var evolveCmd = &cobra.Command{
	Use:   "evolve <prompt>",
	Short: "Tournament-style code evolution using git branches",
	Long: `Tournament-style code evolution using git branches.

Creates an initial implementation, then iteratively improves through rounds:
  1. Create a challenger branch from the current winner
  2. Run improvement prompt on the challenger
  3. AI compares both branches and eliminates the loser
  4. Repeat with the winner

Example:
  agent-exec evolve "implement a snake game" -n 3`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]

		cfg := evolve.EvolveConfig{
			Prompt:              prompt,
			ImprovePrompt:       improvePrompt,
			ComparePrompt:       comparePrompt,
			Iterations:          evolveIters,
			Sleep:               evolveSleep,
			CompareErrorRetries: compareErrorRetries,
			DebugKeepBranches:   debugKeepBranches,

			SystemPrompt:       evolveSystemPrompt,
			AppendSystemPrompt: evolveAppendSystemPrompt,

			ImproveSystemPrompt:       improveSystemPrompt,
			ImproveAppendSystemPrompt: improveAppendSystemPrompt,

			CompareSystemPrompt:       compareSystemPrompt,
			CompareAppendSystemPrompt: compareAppendSystemPrompt,
		}

		// Create emitter and display
		emitter := events.NewChannelEmitter(100)
		baseFormatter := display.NewConsoleFormatter(os.Stdout, evolveVerbose)
		gitClient := git.NewClient(emitter)

		var formatter display.Formatter
		if evolveStatusLine {
			formatter = display.NewStatusLineFormatter(baseFormatter, os.Stdout, true, gitClient)
		} else {
			formatter = baseFormatter
		}

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

	evolveCmd.Flags().StringVar(&evolveSystemPrompt, "system-prompt", "", "Replace entire system prompt for initial prompt")
	evolveCmd.Flags().StringVar(&evolveAppendSystemPrompt, "append-system-prompt", "", "Append to default system prompt for initial prompt")

	evolveCmd.Flags().StringVar(&improveSystemPrompt, "improve-system-prompt", "", "Replace entire system prompt for improvement steps")
	evolveCmd.Flags().StringVar(&improveAppendSystemPrompt, "append-improve-system-prompt", "", "Append to default system prompt for improvement steps")

	evolveCmd.Flags().StringVar(&compareSystemPrompt, "compare-system-prompt", "", "Replace entire system prompt for comparison steps")
	evolveCmd.Flags().StringVar(&compareAppendSystemPrompt, "append-compare-system-prompt", "", "Append to default system prompt for comparison steps")

	evolveCmd.Flags().BoolVarP(&evolveVerbose, "verbose", "v", false, "Show verbose output including all Claude events")
	evolveCmd.Flags().BoolVar(&debugKeepBranches, "debug-keep-branches", false, "Keep all branches for debugging instead of deleting losers")
	evolveCmd.Flags().BoolVar(&evolveStatusLine, "status-line", true, "Show updating status line")
}
