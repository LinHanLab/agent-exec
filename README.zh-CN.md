# agent-exec

[English](README.md) | 中文

Claude Code 自动迭代改进工具，用时间和 token 换取最佳结果。

- 专为长时间运行的 AI agent 任务设计
- 以 headless 模式运行 Claude Code CLI，便于自动化
- 终端输出清晰易读
- 跨平台 (macOS, Linux, Windows)

## 安装

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## 工作原理

### Evolve 命令

基于 git 分支的锦标赛式代码进化。创建相互竞争的实现，通过 AI 比较选出优胜者，在淘汰轮次中迭代改进：

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

### Loop 命令

简单的 Claude Code 提示词迭代执行：

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


## 使用

查看详细用法和所有可用参数：

```bash
agent-exec --help
agent-exec loop --help
agent-exec evolve --help
```

## 示例

参见 [examples/run.sh](examples/run.sh)，这是一个使用 evolve 命令创建贪吃蛇游戏的完整示例。
