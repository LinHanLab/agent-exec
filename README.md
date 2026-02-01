# agent-exec

English | [中文](README.zh-CN.md)

Designed for long-running AI agent tasks, preventing LLM early stops, trading time and tokens for the best results. Minimal Claude Code controller for iterative refinement and Ralph loops.

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

**How it works:**

```
Base Branch
    │
    ├─> [impl-a3f9c2] ← Agent A: Initial implementation
    │        │
    │        └─> Winner (Round 0)
    │
    ├─> Round 1: Tournament
    │   ├─> [impl-a3f9c2] (current winner)
    │   ├─> [impl-7b2e1f] ← Agent B: Improve code
    │   ├─> Agent C: Compare both → Loser: impl-a3f9c2
    │   └─> Winner: [impl-7b2e1f] ✓
    │
    ├─> Round 2: Tournament
    │   ├─> [impl-7b2e1f] (current winner)
    │   ├─> [impl-9d4c8a] ← Agent B: Improve code
    │   ├─> Agent C: Compare both → Loser: impl-9d4c8a
    │   └─> Winner: [impl-7b2e1f] ✓
    │
    └─> Final: Best implementation on winning branch
```

Each iteration creates a competing branch, improves it, and eliminates the worse implementation through AI comparison. The survivor becomes the champion for the next round.

## Help

```bash
agent-exec --help
agent-exec run --help
```
