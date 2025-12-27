package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent-exec",
	Short: "Run Claude CLI prompts with formatted output",
	Long:  `agent-exec wraps the Claude CLI to run prompts with formatted output, supporting one-shot execution and iterative batch processing.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
