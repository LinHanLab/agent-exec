# agent-exec

[English](README.md) | 中文

让 Claude Code 自动迭代改进代码。通过多轮优化，以时间换质量，避免一次生成就草草结束。支持迭代循环和 [Ralph 循环](https://beuke.org/ralph-wiggum-loop/)。

## 安装

```bash
go install github.com/LinHanLab/agent-exec/cmd/agent-exec@latest
```

## 工作原理

### Evolve 命令

淘汰赛式的代码进化。在 git 分支上生成多个版本，让 AI 两两比较、优胜劣汰，逐轮迭代出最优实现：

```
            开始
              │
              ▼
       ┌────────────┐
       │   分支 A   │◄─────────────┐
       │   (初版)   │              │
       └─────┬──────┘              │
             │                     │
    ┌────────▼────────┐            │
    │   创建分支 B    │            │
    │    (改进版)     │            │
    └────────┬────────┘            │
             │                     │
    ┌────────▼────────┐            │
    │ AI 评判 A vs B  │            │
    │    淘汰败者     │            │
    └────────┬────────┘            │
             │                     │
         胜者 ─────────────────────┘
             │              (循环 N 轮)
             ▼
         最优版本
```

### Loop 命令

循环执行同一提示词：

```
        开始
          │
          ▼
   ┌────────────┐
   │  执行提示  │◄────┐
   └─────┬──────┘     │
         │            │
   ┌─────▼──────┐     │
   │    等待    │     │
   └─────┬──────┘     │
         │            │
         └────────────┘
               (循环 N 次)
         │
         ▼
        结束
```


## 使用

查看帮助：

```bash
agent-exec --help
agent-exec loop --help
agent-exec evolve --help
```

## 示例

参见 [examples/run.sh](examples/run.sh)，演示如何用 evolve 命令生成一个贪吃蛇游戏。
