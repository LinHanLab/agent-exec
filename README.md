# agent-exec

English | [中文](README.zh-CN.md)

A CLI tool to run Claude Code headlessly with human-readable formatted output. Supports one-shot execution and iterative batch processing with configurable intervals.

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## Usage

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

### Examples

Run a single prompt:
```bash
agent-exec run "explain how goroutines work"
```

Run 5 iterations:
```bash
agent-exec run "review the code" -n 5
```

Run 3 iterations with 2 minute sleep between each:
```bash
agent-exec run "analyze logs" -n 3 -s 2m
```

## Help

```bash
agent-exec --help
agent-exec run --help
```
