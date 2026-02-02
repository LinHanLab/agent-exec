package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/spf13/cobra"
)

var (
	iterations         int
	sleep              time.Duration
	systemPrompt       string
	appendSystemPrompt string
)

var runCmd = &cobra.Command{
	Use:   "run <prompt>",
	Short: "Run a prompt with Claude CLI",
	Long: `Run a prompt with Claude CLI, optionally repeating for multiple iterations.

Examples:
  agent-exec run "explain this code"
  agent-exec run "review this file" -n 5
  agent-exec run "analyze logs" -n 3 -s 2h30m`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]

		opts := &claude.PromptOptions{
			SystemPrompt:       systemPrompt,
			AppendSystemPrompt: appendSystemPrompt,
		}

		// Create emitter and display
		emitter := events.NewChannelEmitter(100)
		formatter := display.NewConsoleFormatter(os.Stdout)
		disp := display.NewDisplay(formatter, emitter)
		disp.Start()

		var err error
		if iterations == 1 {
			_, err = claude.RunPrompt(prompt, opts, emitter)
		} else {
			err = claude.RunPromptLoop(iterations, sleep, prompt, opts, emitter)
		}

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
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(&iterations, "iterations", "n", 1, "number of iterations to run")
	runCmd.Flags().DurationVarP(&sleep, "sleep", "s", 0, "sleep duration between iterations (e.g., 2h30m, 10s)")
	runCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "replace entire system prompt (empty = use Claude Code defaults)")
	runCmd.Flags().StringVar(&appendSystemPrompt, "append-system-prompt", "", "append to default system prompt (empty = use Claude Code defaults)")
}
