package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "agent-exec",
	Short: "Automated iterative improvement for Claude Code",
	Long: `Automated iterative improvement for Claude Code, trading time and tokens for the best results.

Runs Claude Code CLI in headless mode for automation with human-readable terminal output.

Commands:
  evolve    Tournament-style code evolution using git branches
  loop      Run the same prompt multiple times`,
}
