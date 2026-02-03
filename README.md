# agent-exec

English | [中文](README.zh-CN.md)

Automated iterative improvement for Claude Code. Prevents premature completion, trading time and tokens for the best results. Minimal controller for iterative refinement and [Ralph loops](https://beuke.org/ralph-wiggum-loop/).

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## Usage

### Evolve Command

Tournament-style code evolution using git branches. Creates competing implementations, uses AI comparison to select winners, and iteratively improves through elimination rounds.

```bash
agent-exec evolve <prompt>
agent-exec evolve <prompt> -i <improve-prompt> -c <compare-prompt> -n <iterations>
agent-exec evolve <prompt> -n <iterations> -s <sleep>
```

#### Basic Options

| Flag | Description |
|------|-------------|
| positional | Prompt string (quoted) |
| `-n, --iterations` | Evolution iterations (default: 3) |
| `-s, --sleep` | Sleep duration between evolution rounds (default: 0, format: 2h30m, 30s, etc.) |

#### Prompt

**Plan Agent:**

| Flag | Description |
|------|-------------|
| `--system-prompt` | Replace system prompt for plan agent (empty = use Claude Code defaults) |
| `--append-system-prompt` | Append to system prompt for plan agent (empty = use Claude Code defaults) |

**Improve Agent:**

| Flag | Description |
|------|-------------|
| `-i, --improve` | Improvement prompt (default: "improve the code quality and fix any issues") |
| `--improve-system-prompt` | Replace system prompt for improve agent (empty = use Claude Code defaults) |
| `--append-improve-system-prompt` | Append to system prompt for improve agent (empty = use Claude Code defaults) |

**Compare Agent:**

| Flag | Description |
|------|-------------|
| `-c, --compare` | Comparison prompt (default: "compare these two implementations and determine which is worse") |
| `--compare-system-prompt` | Replace system prompt for compare agent (empty = use Claude Code defaults) |
| `--append-compare-system-prompt` | Append to system prompt for compare agent (empty = use Claude Code defaults) |
| `--compare-error-retries` | Number of retries when comparison parsing fails (default: 3) |

#### Advanced Options

| Flag | Description |
|------|-------------|
| `--debug-keep-branches` | Debug mode: keep all branches instead of deleting losers |
| `-v, --verbose` | Show full content without truncation |

**How it works:**

```
Base Branch
    │
    ├─> [impl-a3f9c2] ← Initial implementation
    │        │
    │        └─> Winner (Round 0)
    │
    ├─> Round 1: Tournament
    │   ├─> [impl-a3f9c2] (current winner)
    │   ├─> [impl-7b2e1f] ← Improve code
    │   ├─> Compare implementations → Loser: impl-a3f9c2
    │   └─> Winner: [impl-7b2e1f] ✓
    │
    ├─> Round 2: Tournament
    │   ├─> [impl-7b2e1f] (current winner)
    │   ├─> [impl-9d4c8a] ← Improve code
    │   ├─> Compare implementations → Loser: impl-9d4c8a
    │   └─> Winner: [impl-7b2e1f] ✓
    │
    └─> Final: Best implementation on winning branch
```

Each iteration creates a competing branch, improves it, and eliminates the worse implementation through AI comparison. The survivor becomes the champion for the next round.

### Loop Command

Execute prompts in a loop with optional iterations and sleep intervals.

```bash
agent-exec loop <prompt>
agent-exec loop <prompt> -n <iterations>
agent-exec loop <prompt> -n <iterations> -s <sleep>
```

#### Basic Options

| Flag | Description |
|------|-------------|
| positional | Prompt string (quoted) |
| `-n, --iterations` | Number of iterations to run (default: 1) |
| `-s, --sleep` | Sleep duration between iterations (default: 0, format: 2h30m, 30s, etc.) |
| `-v, --verbose` | Show full content without truncation |

#### Prompt

| Flag | Description |
|------|-------------|
| `--system-prompt` | Replace entire system prompt (empty = use Claude Code defaults) |
| `--append-system-prompt` | Append to default system prompt (empty = use Claude Code defaults) |

## Help

```bash
agent-exec --help
agent-exec loop --help
agent-exec evolve --help
```
