# agent-exec

English | [中文](README.zh-CN.md)

A CLI tool to run Claude Code headlessly with human-readable formatted output. Supports one-shot execution and iterative batch processing with configurable intervals.

## Installation

```bash
go install github.com/LinHanLab/agent-exec@latest
```

## Commands

### one-shot

Execute a single prompt and display the formatted result.

```bash
agent-exec one-shot <prompt>
```

**Example:**
```bash
agent-exec one-shot "explain how goroutines work"
```

### iterations

Run prompts from files repeatedly with sleep intervals between iterations. Useful for batch processing or monitoring tasks.

```bash
agent-exec iterations <count> <sleep_seconds> <file1> [file2...]
```

| Argument | Description |
|----------|-------------|
| `count` | Number of iterations to run |
| `sleep_seconds` | Delay between iterations |
| `file1...` | One or more files containing prompts |

**Example:**
```bash
agent-exec iterations 5 30 prompts/review.md prompts/summarize.md
```

This runs both prompt files 5 times, waiting 30 seconds between each iteration.

## Help

```bash
agent-exec --help
```
