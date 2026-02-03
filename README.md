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

### Loop Command

Execute prompts in a loop with optional iterations and sleep intervals.

