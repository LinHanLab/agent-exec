# agent-exec

English | [中文](README.zh-CN.md)

Automated iterative improvement for Claude Code, trading time and tokens for the best results.

- Designed for long-running AI agent tasks
- Runs Claude Code CLI in headless mode for automation
- Human-readable terminal output
- Cross-platform (macOS, Linux, Windows)

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## How It Works

### Evolve Command

Tournament-style code evolution using git branches. Creates competing implementations, uses AI comparison to select winners, and iteratively improves through elimination rounds:

```
              Start
                │
                ▼
         ┌─────────────┐
         │  Branch A   │◄──────────────┐
         │  (initial)  │               │
         └──────┬──────┘               │
                │                      │
     ┌──────────▼──────────┐           │
     │   Create Branch B   │           │
     │   (improve A)       │           │
     └──────────┬──────────┘           │
                │                      │
     ┌──────────▼──────────┐           │
     │   AI: Compare A,B   │           │
     │   Delete loser      │           │
     └──────────┬──────────┘           │
                │                      │
         Winner ───────────────────────┘
                │                (repeat N times)
                ▼
          Best Result
```

### Loop Command

Simple iterative execution of Claude Code prompts:

```
         Start
           │
           ▼
    ┌─────────────┐
    │ Run Prompt  │◄─────┐
    └──────┬──────┘      │
           │             │
    ┌──────▼──────┐      │
    │   Sleep     │      │
    └──────┬──────┘      │
           │             │
           └─────────────┘
                  (repeat N times)
           │
           ▼
        Complete
```


## Usage

For detailed usage and all available flags:

```bash
agent-exec --help
agent-exec loop --help
agent-exec evolve --help
```

## Examples

See [examples/run.sh](examples/run.sh) for a complete example of creating a snake game using the evolve command.

