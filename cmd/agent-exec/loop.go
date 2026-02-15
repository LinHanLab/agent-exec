package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/commands/loop"
	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/LinHanLab/agent-exec/pkg/git"
	"github.com/spf13/cobra"
)

var (
	iterations         int
	sleep              time.Duration
	systemPrompt       string
	appendSystemPrompt string
	verbose            bool
	statusLine         bool
)

var loopCmd = &cobra.Command{
	Use:   "loop <prompt>",
	Short: "Run the same prompt multiple times",
	Long: `Run the same prompt multiple times for iterative refinement.

Each iteration runs Claude Code with the given prompt. Use -n to set the
number of iterations and -s to add sleep between runs.

Example:
  agent-exec loop "improve code quality" -n 5 -s 30s`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]

		opts := &claude.PromptOptions{
			SystemPrompt:       systemPrompt,
			AppendSystemPrompt: appendSystemPrompt,
		}

		// Create emitter and display
		emitter := events.NewChannelEmitter(100)
		baseFormatter := display.NewConsoleFormatter(os.Stdout, verbose)
		gitClient := git.NewClient(emitter)

		var formatter display.Formatter
		if statusLine {
			formatter = display.NewStatusLineFormatter(baseFormatter, os.Stdout, true, gitClient)
		} else {
			formatter = baseFormatter
		}

		disp := display.NewDisplay(formatter, emitter)
		disp.Start()

		var err error
		if iterations == 1 {
			_, err = claude.RunPrompt(prompt, opts, emitter)
		} else {
			err = loop.RunPromptLoop(iterations, sleep, prompt, opts, emitter)
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
	rootCmd.AddCommand(loopCmd)

	loopCmd.Flags().IntVarP(&iterations, "iterations", "n", 1, "Number of times to execute the prompt")
	loopCmd.Flags().DurationVarP(&sleep, "sleep", "s", 0, "Sleep duration between iterations (e.g., 30s, 1m)")
	loopCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "Replace entire system prompt sent to Claude")
	loopCmd.Flags().StringVar(&appendSystemPrompt, "append-system-prompt", "", "Append additional instructions to default system prompt")
	loopCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including all Claude events")
	loopCmd.Flags().BoolVar(&statusLine, "status-line", true, "Show updating status line")
}
