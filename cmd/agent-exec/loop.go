package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/claude"
	"github.com/LinHanLab/agent-exec/pkg/commands/loop"
	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
	"github.com/spf13/cobra"
)

var (
	iterations         int
	sleep              time.Duration
	systemPrompt       string
	appendSystemPrompt string
	verbose            bool
)

var loopCmd = &cobra.Command{
	Use:   "loop <prompt>",
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := args[0]

		opts := &claude.PromptOptions{
			SystemPrompt:       systemPrompt,
			AppendSystemPrompt: appendSystemPrompt,
		}

		// Create emitter and display
		emitter := events.NewChannelEmitter(100)
		formatter := display.NewConsoleFormatter(os.Stdout, verbose)
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

	loopCmd.Flags().IntVarP(&iterations, "iterations", "n", 1, "")
	loopCmd.Flags().DurationVarP(&sleep, "sleep", "s", 0, "")
	loopCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "")
	loopCmd.Flags().StringVar(&appendSystemPrompt, "append-system-prompt", "", "")
	loopCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "")
}
