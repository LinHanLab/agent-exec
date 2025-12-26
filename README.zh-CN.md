# agent-exec

[English](README.md) | 中文

一个用于无头运行 Claude Code 并提供人类可读格式化输出的命令行工具。支持单次执行和可配置间隔的迭代批处理。

## 安装

```bash
go install github.com/LinHanLab/agent-exec@latest
```

## 命令

### one-shot

执行单个提示词并显示格式化结果。

```bash
agent-exec one-shot <prompt>
```

**示例：**
```bash
agent-exec one-shot "解释 goroutine 的工作原理"
```

### iterations

从文件中读取提示词并重复执行，每次迭代之间有休眠间隔。适用于批处理或监控任务。

```bash
agent-exec iterations <count> <sleep_seconds> <file1> [file2...]
```

| 参数 | 说明 |
|------|------|
| `count` | 迭代次数 |
| `sleep_seconds` | 迭代间隔（秒） |
| `file1...` | 一个或多个包含提示词的文件 |

**示例：**
```bash
agent-exec iterations 5 30 prompts/review.md prompts/summarize.md
```

以上命令会执行两个提示词文件 5 次，每次迭代间隔 30 秒。

## 帮助

```bash
agent-exec --help
```
