# agent-exec

[English](README.md) | 中文

一个用于无头运行 Claude Code 并提供人类可读格式化输出的命令行工具。支持单次执行和可配置间隔的迭代批处理。

## 安装

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## 使用方法

### Run 命令

执行提示词，支持可选的迭代次数和间隔时长。

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

**示例：**

```bash
agent-exec run "解释 goroutine 的工作原理"
agent-exec run "审查代码" -n 5
agent-exec run "分析日志" -n 3 -s 2m
```

### Evolve 命令

通过竞争分支的锦标赛式迭代改进来演化代码。

```bash
agent-exec evolve <plan>
agent-exec evolve <plan> -i <improve-prompt> -c <compare-prompt> -n <iterations>
```

| 参数 | 说明 |
|------|------|
| 位置参数 | 实现计划（需引号包裹） |
| `-i, --improve` | 改进提示词（默认："improve the code quality and fix any issues"） |
| `-c, --compare` | 比较提示词（默认："compare these two implementations and determine which is worse"） |
| `-n, --iterations` | 演化迭代次数（默认：3） |

**示例：**

```bash
agent-exec evolve "实现用户 REST API"
agent-exec evolve "构建 CLI 工具" -n 5
agent-exec evolve "创建解析器" -i "优化性能" -c "哪个更快？"
```

## 帮助

```bash
agent-exec --help
agent-exec run --help
```
