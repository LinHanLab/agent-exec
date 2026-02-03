package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "agent-exec",
	Short: "A CLI tool for automated iterative improvement with Claude Code",
	Long: `A CLI tool for automated iterative improvement with Claude Code, trading time and tokens for the best results.

Runs Claude Code CLI in headless mode for automation with human-readable terminal output.

Commands:
  evolve    Tournament-style code evolution using git branches
  loop      Run the same prompt multiple times`,
}
