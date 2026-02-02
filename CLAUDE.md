# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

agent-exec is a CLI tool that wraps the Claude CLI to run prompts headlessly with human-readable formatted output. It supports one-shot execution and iterative batch processing with configurable intervals.

## Development Commands

### Build
```bash
make compile          # Build binary to ./app
go build -o app ./cmd/agent-exec
```

### Install
```bash
make install          # Install to $GOPATH/bin
go install ./cmd/agent-exec
```

### Quality Checks
```bash
make quality          # Run all quality checks (tests, fmt, lint)
go test ./...         # Run all tests
go fmt ./...          # Format code
golangci-lint run     # Run linter
```

### Run Tests
```bash
go test ./...                           # All tests
go test ./pkg/claude                    # Package-specific tests
go test -v ./pkg/format/text_test.go    # Single test file
go test -run TestFunctionName ./pkg/... # Specific test function
```

## Architecture

### Package Structure

- **cmd/agent-exec**: CLI entry point using cobra
  - `main.go`: Entry point
  - `root.go`: Root command definition
  - `loop.go`: Loop command implementation with iteration/sleep flags
  - `evolve.go`: Evolve command implementation

- **pkg/claude**: Core Claude CLI interaction utilities
  - `prompt.go`: Executes claude CLI commands, handles basic prompt execution
  - `parser.go`: Parses streaming JSON output from claude CLI
  - `types.go`: JSON structure definitions for claude CLI output
  - `validate.go`: Input validation for prompts

- **pkg/commands**: Command-specific business logic
  - `loop/`: Loop command logic with iteration and sleep handling
  - `evolve/`: Evolve command logic for tournament-style code improvement

- **pkg/format**: Text formatting utilities
  - `text.go`: Text truncation, wrapping, and prefixed printing
  - `color.go`: ANSI color code constants

### Key Design Patterns

**Streaming JSON Parser**: The tool parses claude CLI's `--output-format stream-json` output line-by-line. Each line is a JSON object with type "assistant", "user", or "result". The parser extracts and formats:
- Assistant text messages (💬)
- Tool use calls with inputs (🔧)
- Tool results (✅)
- Final result with duration (⏱️)

**Graceful Interruption**: The loop command (pkg/commands/loop) uses signal channels to handle SIGINT/SIGTERM during both execution and sleep periods, allowing clean shutdown of multi-iteration runs.

**Display Formatting**: Prompts are truncated to 270 chars and wrapped at 76 chars width with a "▐ " prefix for visual clarity.

## Testing

Tests use table-driven patterns. See `pkg/format/text_test.go` and `pkg/claude/validate_test.go` for examples.

## Dependencies

- github.com/spf13/cobra: CLI framework
- Go 1.25.3+
- Requires `claude` CLI to be installed and in PATH
