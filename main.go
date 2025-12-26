package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/LinHanLab/agent-exec/pkg/claude"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "one-shot":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: one-shot requires exactly one prompt argument")
			fmt.Fprintln(os.Stderr, "Usage: agent-exec one-shot <prompt>")
			os.Exit(1)
		}
		if err := claude.RunPrompt(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "iterations":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Error: iterations requires at least 3 arguments")
			fmt.Fprintln(os.Stderr, "Usage: agent-exec iterations <count> <sleep_seconds> <file1> [file2...]")
			os.Exit(1)
		}

		iterations, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid iteration count: %s\n", os.Args[2])
			os.Exit(1)
		}

		sleepSeconds, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid sleep seconds: %s\n", os.Args[3])
			os.Exit(1)
		}

		promptFiles := os.Args[4:]

		if err := claude.RunPromptLoop(iterations, sleepSeconds, promptFiles); err != nil {
			if err.Error() == "interrupted" {
				os.Exit(130)
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `agent-exec - Run Claude CLI prompts with formatted output

Usage:
  agent-exec one-shot <prompt>
  agent-exec iterations <count> <sleep_seconds> <file1> [file2...]

Commands:
  one-shot     Run a single prompt
               Example: agent-exec one-shot "explain this code"

  iterations   Run multiple prompts in iterations
               Example: agent-exec iterations 5 30 task1.md task2.md

               Arguments:
                 count          Number of iterations to run
                 sleep_seconds  Seconds to sleep between iterations
                 file1...       One or more prompt files to execute

Options:
  -h, --help   Show this help message
`
	fmt.Print(usage)
}
