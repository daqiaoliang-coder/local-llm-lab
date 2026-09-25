# local-llm-lab v0.4

一个面向后端 / Agent Infra 学习的本地 LLM 实验室。重点不是“跑一个聊天机器人”，而是把一个本地模型逐步演进成一个可恢复的 Agent Runtime。

## v0.4 新增

- **Durable Agent Execution**：Run 不再只是一次函数调用，而是可持久化、可恢复的执行单元。
- **Checkpoint = logical state**：持久化完整 `openai.ChatCompletionMessage` 序列，让重启后的 Agent 能继续下一次 LLM 调用。
- **Run Resume**：启动时扫描 `RUNNING` Run，恢复 checkpoint，reconcile pending ToolCall，再继续原 Run。
- **Crash Recovery**：ToolCall 卡在 RUNNING/RETRYING 时，根据 RetryPolicy 恢复。
- **Idempotency Key**：`tool:<tool_call_id>` 在进程重启后保持不变。
- **Tool Retry Contract**：Tool 自己声明是否可重试以及最大次数。
- **多 ToolCall pending state**：支持一次 LLM 返回多个工具调用时逐个恢复。

## 快速开始

```bash
ollama pull qwen3:4b
ollama serve

go mod tidy
go run ./cmd/chat
```

如果你已经有 v0.3 的数据库，可以直接复用 `./data/agent.db`；迁移逻辑会保留已有数据。

## 代码结构

```text
cmd/chat/                  CLI
internal/llm/              Ollama / OpenAI-compatible client
internal/tools/            Tool + RetryPolicy + file tools
internal/db/               durable state / checkpoint / ToolCall repository
internal/agent/            Run / Resume / Recovery

docs/01-v04-design.md      v0.4 设计
docs/02-crash-sequence.md  Crash / Resume 时序
docs/03-interview.md       面试问答
```

## 关键状态机

```text
Run:      RUNNING ───────────────→ SUCCEEDED / FAILED

ToolCall: RUNNING → RETRYING → SUCCEEDED
              │                   └→ FAILED
              └────────────────────→ FAILED

Checkpoint:
  assistant tool-call message
          ↓
  pending_tool_call_ids
          ↓
  tool result appended
          ↓
  pending_tool_call_ids = []
          ↓
  next LLM call
```

## 如何理解这个项目的核心价值

v0.2 解决“状态落盘”；v0.3 解决“发现失败并重试 Tool”；v0.4 才真正解决“恢复 Agent 的逻辑执行”。

最值得在面试中讲的不是 SQLite，而是：

> **execution state ≠ conversation state。**
>
> Durable Agent Runtime 不仅要知道“哪个 Step 在运行”，还必须保存能让 LLM 从断点继续工作的 logical state。

## 生产级差距

这是本地实验室，不声称是生产 Runtime。距离生产级还需要：

- DB transaction / Outbox
- Worker lease / heartbeat / fencing token
- 分布式任务 claim 与并发控制
- 副作用工具的真正幂等协议
- Tool sandbox / capability / permission
- tracing / metrics / replay / audit
- 多 Worker 调度与水平扩展
- 更完整的 cancel / timeout / human-in-the-loop

这些恰好也是下一阶段可以继续做成 Agent Infra 项目的方向。
