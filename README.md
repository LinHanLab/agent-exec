# agent-exec

English | [中文](README.zh-CN.md)

Designed for long-running AI agent tasks, preventing LLM early stops, trading time and tokens for the best results. Minimal Claude Code controller for iterative refinement and Ralph loops.

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## Usage

### Evolve Command

Evolve code through tournament-style iterative improvement with competing branches.

```bash
agent-exec evolve <prompt>
agent-exec evolve <prompt> -i <improve-prompt> -c <compare-prompt> -n <iterations>
agent-exec evolve <prompt> -n <iterations> -s <sleep>
```

| Flag | Description |
|------|-------------|
| positional | Prompt string (quoted) |
| `-i, --improve` | Improvement prompt (default: "improve the code quality and fix any issues") |
| `-c, --compare` | Comparison prompt (default: "compare these two implementations and determine which is worse") |
| `-n, --iterations` | Evolution iterations (default: 3) |
| `-s, --sleep` | Sleep duration between evolution rounds (default: 0, format: 2h30m, 30s, etc.) |
| `--compare-error-retries` | Number of retries when comparison parsing fails (default: 3) |
| `--system-prompt` | Replace system prompt for plan agent (empty = use Claude Code defaults) |
| `--append-system-prompt` | Append to system prompt for plan agent (empty = use Claude Code defaults) |
| `--improve-system-prompt` | Replace system prompt for improve agent (empty = use Claude Code defaults) |
| `--append-improve-system-prompt` | Append to system prompt for improve agent (empty = use Claude Code defaults) |
| `--compare-system-prompt` | Replace system prompt for compare agent (empty = use Claude Code defaults) |
| `--append-compare-system-prompt` | Append to system prompt for compare agent (empty = use Claude Code defaults) |

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
| `--system-prompt` | Replace entire system prompt (empty = use Claude Code defaults) |
| `--append-system-prompt` | Append to default system prompt (empty = use Claude Code defaults) |

## Help

```bash
agent-exec --help
agent-exec run --help
agent-exec evolve --help
```
