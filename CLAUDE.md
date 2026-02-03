# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quality Tools Enforcement

CRITICAL: After finishing any code changes, you MUST run these quality checks:

```bash
make quality        # Run tests, formatting, and linting
```
Do NOT skip these checks. Fix all errors before considering the task complete.

---

## Project Overview

agent-exec is a Go CLI tool for automated iterative improvement with Claude Code. It wraps the `claude` CLI to enable tournament-style code evolution and iterative refinement loops.

## Build and Development Commands

### Build
```bash
make compile        # Build to ./app executable
go build -o app ./cmd/agent-exec
```

### Install
```bash
make install                                           # Install to $GOPATH/bin
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

### Quality Checks
```bash
make quality        # Run tests, formatting, and linting
go test ./...       # Run all tests
go fmt ./...        # Format code
golangci-lint run   # Run linter
```

### Run Single Test
```bash
go test -v ./pkg/claude -run TestValidatePrompt
go test -v ./pkg/display -run TestFormatClaudeMessage
```

## Architecture

### Core Flow
The tool executes Claude Code prompts via subprocess (`claude` CLI) and parses streaming JSON output to display formatted results and coordinate multi-step workflows.

**Critical dependency**: This tool requires the `claude` CLI (Claude Code) to be installed and available in PATH.

### Module Structure

**cmd/agent-exec/** - CLI entry point using cobra
- `main.go` - Application entry
- `root.go` - Root command setup
- `evolve.go` - Evolve command configuration
- `loop.go` - Loop command configuration

**pkg/claude/** - Claude CLI interaction
- `prompt.go` - Executes `claude` CLI with system prompt options
- `parser.go` - Parses streaming JSON output (`--output-format stream-json`)
- `validate.go` - Input validation
- `types.go` - Type definitions

**pkg/commands/** - Command implementations
- `evolve/evolve.go` - Tournament evolution: creates competing git branches, runs improvement prompts, uses AI comparison to select winners
- `loop/loop.go` - Simple iteration: runs same prompt N times with optional sleep

**pkg/git/** - Git operations
- `branch.go` - Branch creation, checkout, deletion, squashing

**pkg/events/** - Event-driven architecture
- `emitter.go` - Event emission interface
- `types.go` - Event type definitions and data structures

**pkg/display/** - Output formatting
- `console.go` - Main display coordinator
- `formatter.go` - Event-to-text conversion
- `handlers.go` - Event-specific formatters
- `content.go` - Content filtering logic
- `text.go` - Text wrapping and formatting
- `color.go` - ANSI color helpers

### Evolve Command Architecture

The `evolve` command implements tournament-style code evolution:

1. **Initial Plan**: Creates branch A, runs plan prompt, squashes commits
2. **Evolution Loop** (N iterations):
   - Create challenger branch from current winner
   - Run improvement prompt on challenger
   - Switch to original branch, run comparison prompt
   - Parse comparison result to determine loser
   - Delete loser branch (unless `--debug-keep-branches`)
   - Update current winner
3. **Result**: Final branch contains best implementation

The comparison prompt asks Claude to identify which branch is worse. The parser in `evolve.go:parseBranchFromResponse()` expects the response to contain exactly one branch name.

### Event System

Commands emit events through `events.Emitter` interface. The display package subscribes to these events and formats them for console output. This decouples command logic from presentation.

Event flow: Command → Emitter → Display → Formatter → Console

### Claude CLI Integration

The tool spawns `claude --verbose --output-format stream-json -p "<prompt>"` as a subprocess and parses its stdout line-by-line. Each line is a JSON event (assistant_message, tool_use, tool_result, etc.). The parser extracts and emits these as internal events.

System prompts can be customized per command:
- `--system-prompt`: Replace entire system prompt
- `--append-system-prompt`: Append to default system prompt

### Signal Handling

Both `loop` and `evolve` commands handle SIGINT/SIGTERM gracefully. They check for interrupts before each iteration and during sleep periods, emitting appropriate interrupt events before exiting with code 130.

## Go Version

Requires Go 1.25.3 (as specified in go.mod).

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/term` - Terminal utilities
