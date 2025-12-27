# agent-exec

[English](README.md) | 中文

一个用于无头运行 Claude Code 并提供人类可读格式化输出的命令行工具。支持单次执行和可配置间隔的迭代批处理。

## 安装

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## 使用方法

```bash
agent-exec run <prompt>
agent-exec run <prompt> -n <iterations>
agent-exec run <prompt> -n <iterations> -s <sleep>
```

| 参数 | 说明 |
|------|------|
| 位置参数 | 提示词字符串（需引号包裹） |
| `-n, --iterations` | 运行次数（默认：1） |
| `-s, --sleep` | 迭代间隔时长（默认：0，格式：2h30m、30s 等） |

### 示例

执行单个提示词：
```bash
agent-exec run "解释 goroutine 的工作原理"
```

运行 5 次迭代：
```bash
agent-exec run "审查代码" -n 5
```

运行 3 次迭代，每次间隔 2 分钟：
```bash
agent-exec run "分析日志" -n 3 -s 2m
```

## 帮助

```bash
agent-exec --help
agent-exec run --help
```
