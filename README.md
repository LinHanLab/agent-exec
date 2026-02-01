# agent-exec

English | [中文](README.zh-CN.md)

A CLI tool to run Claude Code headlessly with human-readable formatted output. Supports one-shot execution and iterative batch processing with configurable intervals.

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## Usage

### Run Command

Execute prompts with optional iterations and sleep intervals.

```bash
agent-exec run <prompt>
agent-exec run <prompt> -n <iterations>
agent-exec run <prompt> -n <iterations> -s <sleep>
```

| Flag | Description |
|------|-------------|
| positional | Prompt string (quoted) |
| `-n, --iterations` | Run count (default: 1) |
| `-s, --sleep` | Sleep duration between iterations (default: 0, format: 2h30m, 30s, etc.) |

**Examples:**

```bash
agent-exec run "explain how goroutines work"
agent-exec run "review the code" -n 5
agent-exec run "analyze logs" -n 3 -s 2m
```

### Evolve Command

Evolve code through tournament-style iterative improvement with competing branches.

```bash
agent-exec evolve <plan>
agent-exec evolve <plan> -i <improve-prompt> -c <compare-prompt> -n <iterations>
```

| Flag | Description |
|------|-------------|
| positional | Implementation plan (quoted) |
| `-i, --improve` | Improvement prompt (default: "improve the code quality and fix any issues") |
| `-c, --compare` | Comparison prompt (default: "compare these two implementations and determine which is worse") |
| `-n, --iterations` | Evolution iterations (default: 3) |

**Examples:**

```bash
agent-exec evolve "implement a REST API for users"
agent-exec evolve "build a CLI tool" -n 5
agent-exec evolve "create a parser" -i "optimize performance" -c "which is faster?"
```

## Help

```bash
agent-exec --help
agent-exec run --help
```
