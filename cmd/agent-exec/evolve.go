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
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(1),
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

	evolveCmd.Flags().StringVarP(&improvePrompt, "improve", "i", "improve the code quality and fix any issues", "")
	evolveCmd.Flags().StringVarP(&comparePrompt, "compare", "c", "compare these two implementations and determine which is worse", "")
	evolveCmd.Flags().IntVarP(&evolveIters, "iterations", "n", 3, "")
	evolveCmd.Flags().DurationVarP(&evolveSleep, "sleep", "s", 0, "")
	evolveCmd.Flags().IntVar(&compareErrorRetries, "compare-error-retries", 3, "")

	evolveCmd.Flags().StringVar(&planSystemPrompt, "system-prompt", "", "")
	evolveCmd.Flags().StringVar(&planAppendSystemPrompt, "append-system-prompt", "", "")

	evolveCmd.Flags().StringVar(&improveSystemPrompt, "improve-system-prompt", "", "")
	evolveCmd.Flags().StringVar(&improveAppendSystemPrompt, "append-improve-system-prompt", "", "")

	evolveCmd.Flags().StringVar(&compareSystemPrompt, "compare-system-prompt", "", "")
	evolveCmd.Flags().StringVar(&compareAppendSystemPrompt, "append-compare-system-prompt", "", "")

	evolveCmd.Flags().BoolVarP(&evolveVerbose, "verbose", "v", false, "")
	evolveCmd.Flags().BoolVar(&debugKeepBranches, "debug-keep-branches", false, "")
}
