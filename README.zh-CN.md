# agent-exec

[English](README.md) | 中文

专为长时间运行的 AI 代理任务设计，防止 LLM 提前停止，以时间和 token 换取最佳结果。极简的 Claude Code 控制器，用于迭代优化和 Ralph 循环。

## 安装

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## 使用方法

### Evolve 命令

通过竞争分支的锦标赛式迭代改进来演化代码。

```bash
agent-exec evolve <prompt>
agent-exec evolve <prompt> -i <improve-prompt> -c <compare-prompt> -n <iterations>
```

| 参数 | 说明 |
|------|------|
| 位置参数 | 提示词字符串（需引号包裹） |
| `-i, --improve` | 改进提示词（默认："improve the code quality and fix any issues"） |
| `-c, --compare` | 比较提示词（默认："compare these two implementations and determine which is worse"） |
| `-n, --iterations` | 演化迭代次数（默认：3） |

**工作原理：**

```
基础分支
    │
    ├─> [impl-a3f9c2] ← 初始实现
    │        │
    │        └─> 胜者（第 0 轮）
    │
    ├─> 第 1 轮：锦标赛
    │   ├─> [impl-a3f9c2]（当前胜者）
    │   ├─> [impl-7b2e1f] ← 改进代码
    │   ├─> 比较实现 → 败者：impl-a3f9c2
    │   └─> 胜者：[impl-7b2e1f] ✓
    │
    ├─> 第 2 轮：锦标赛
    │   ├─> [impl-7b2e1f]（当前胜者）
    │   ├─> [impl-9d4c8a] ← 改进代码
    │   ├─> 比较实现 → 败者：impl-9d4c8a
    │   └─> 胜者：[impl-7b2e1f] ✓
    │
    └─> 最终：胜出分支上的最佳实现
```

每次迭代创建一个竞争分支，改进代码，然后通过 AI 比较淘汰较差的实现。幸存者成为下一轮的冠军。

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

## 帮助

```bash
agent-exec --help
agent-exec run --help
```
