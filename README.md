# agent-exec

English | [中文](README.zh-CN.md)

Automated iterative improvement for Claude Code. Prevents premature completion, trading time and tokens for the best results. Minimal controller for iterative refinement and [Ralph loops](https://beuke.org/ralph-wiggum-loop/).

## Installation

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## How It Works

### Evolve Command

Tournament-style code evolution using git branches. Creates competing implementations, uses AI comparison to select winners, and iteratively improves through elimination rounds:

```
┌─────────────────────────────────────────────────────┐
│  Create Branch A: Run Initial Plan                  │
└─────────────────┬───────────────────────────────────┘
                  │
                  ▼
          ┌───────────────┐
    ┌────►│  Round Start  │
    │     └───────┬───────┘
    │             │
    │             ▼
    │     ┌───────────────────────────────┐
    │     │  Create Challenger Branch B   │
    │     │  From Current Winner          │
    │     └───────┬───────────────────────┘
    │             │
    │             ▼
    │     ┌───────────────────────────────┐
    │     │  Run Improvement Prompt on B  │
    │     └───────┬───────────────────────┘
    │             │
    │             ▼
    │     ┌───────────────────────────────┐
    │     │  Switch to Original Branch    │
    │     │  Run Comparison Prompt        │
    │     │  (A vs B - which is worse?)   │
    │     └───────┬───────────────────────┘
    │             │
    │             ▼
    │     ┌────────────────┐
    │     │  Parse Result  │
    │     │  Identify Loser│
    │     └───────┬────────┘
    │             │
    │             ▼
    │     ┌───────────────────────────────┐
    │     │  Delete Loser Branch          │
    │     │  Winner Becomes Current       │
    │     └───────┬───────────────────────┘
    │             │
    │             ▼
    │     ┌───────────────┐
    │     │  Next Round?  ├──Yes──┐
    │     └───────┬───────┘       │
    │             │ No            │
    └─────────────┘               │
                  │               │
                  ▼               │
          ┌───────────────┐       │
          │  Final Winner │       │
          └───────────────┘       │
                                  │
                  ┌───────────────┘
                  │
                  ▼
          ┌───────────────┐
          │  Sleep (opt)  │
          └───────────────┘
```

### Loop Command

Simple iterative execution of Claude Code prompts:

```
┌──────────────────────────────────────────┐
│  Start Loop (N iterations)               │
└──────────────┬───────────────────────────┘
               │
               ▼
       ┌───────────────┐
       │  Run Prompt   │◄─────┐
       └───────┬───────┘      │
               │              │
               ▼              │
       ┌───────────────┐      │
       │  Sleep (opt)  │      │
       └───────┬───────┘      │
               │              │
               ▼              │
       ┌───────────────┐      │
       │  Next Round?  ├──Yes─┘
       └───────┬───────┘
               │ No
               ▼
       ┌───────────────┐
       │   Complete    │
       └───────────────┘
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

