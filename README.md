# Local LLM Lab v0.1

一个面向 Agent Infra 学习的本地小模型实验室，目标是用一台 32GB Intel Mac 跑通：

- Ollama 本地 LLM
- OpenAI-compatible API
- Go LLM Client
- Tool Calling / Agent Loop
- 本地文件工具
- 可选 RAG 实验
- 后续接入已有 Agent Runtime

## 1. 环境

推荐：

- macOS
- Go 1.22+
- Ollama
- Python 3.10+（后续 RAG / 模型实验使用）

安装 Ollama：

https://ollama.com/

拉取模型：

```bash
ollama pull qwen3:4b
```

启动：

```bash
ollama serve
```

另开终端运行：

```bash
go run ./cmd/chat
```

## 2. 项目结构

```text
local-llm-lab/
├── cmd/chat/              # 最小 CLI
├── internal/llm/          # Ollama/OpenAI-compatible client
├── internal/agent/        # Agent Loop
├── internal/tools/        # Tool Registry + 文件工具
├── docs/                  # 学习与实验记录
├── experiments/           # 实验脚本/记录
└── Makefile
```

## 3. 当前能力

v0.1 已实现：

1. 向本地 Ollama 发送 Chat 请求
2. 基础 Agent Loop
3. Tool Registry
4. `list_files` 工具
5. `read_file` 工具
6. 简单的工具调用协议
7. CLI 交互

> 注意：不同 Ollama/Qwen 版本对原生 tool calling 的行为可能不同。v0.1 保留了明确的 ToolCall 数据结构和执行边界，方便后续替换成原生 OpenAI-compatible tools。

## 4. 运行

```bash
make run
```

或者：

```bash
go run ./cmd/chat
```

默认服务：

```text
http://localhost:11434/v1
```

默认模型：

```text
qwen3:4b
```

可以覆盖：

```bash
LLM_MODEL=qwen3:4b LLM_BASE_URL=http://localhost:11434/v1 go run ./cmd/chat
```

## 5. 建议实验路线

### Experiment 01
单轮 LLM：

```text
User -> LLM -> Answer
```

### Experiment 02
Agent Loop：

```text
User
  ↓
LLM
  ↓
Tool?
  ├── No -> Answer
  └── Yes -> Tool -> Observation -> LLM
```

### Experiment 03
加入 Tool Registry：

```text
Tool
├── name
├── description
├── input schema
└── executor
```

### Experiment 04
加入持久化：

```text
Run
 ├── Step
 ├── ToolCall
 └── Observation
```

### Experiment 05
加入 checkpoint / retry / idempotency。

这一步开始可以逐渐与你的 Agent Runtime 项目融合。

## 6. 与 Agent Runtime 的关系

建议不要把这个实验室做成另一个“大而全 Agent Framework”。

它应该承担：

```text
Local Model
    ↓
LLM Client
    ↓
Agent Loop
    ↓
Tool Calling
    ↓
Runtime Experiments
```

然后把成熟的：

- Run / Step
- Dynamic DAG
- Durable State
- Checkpoint
- Crash Recovery
- Event-driven execution

逐步迁移到你的 `agent-runtime` 项目。

## 7. 后续版本

建议按以下顺序演进：

- v0.2：原生 Tool Calling
- v0.3：SQLite/MySQL Durable State
- v0.4：Checkpoint + Resume
- v0.5：RAG
- v0.6：Sandbox / Remote Tool Provider
- v0.7：Dynamic DAG
- v1.0：Local Agent Runtime Demo

