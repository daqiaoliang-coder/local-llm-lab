# v0.4 — Durable Agent Execution / Resume

## 目标

v0.3 能发现并重试 RUNNING ToolCall，但还不能真正恢复 Agent 的逻辑执行。v0.4 把恢复边界提升到 **Run Resume**：

```text
LLM → assistant(tool calls) → Tool → checkpoint
                         ↓ crash
                    restart process
                         ↓
              restore checkpoint
                         ↓
              reconcile ToolCall
                         ↓
                  continue LLM
                         ↓
                    complete Run
```

## 核心设计

1. **Checkpoint 持久化完整 message state**：不仅保存 tool 的结果，还保存下一次 LLM 调用需要的完整消息序列。
2. **ToolCall 是独立 durable state machine**：RUNNING / RETRYING / SUCCEEDED / FAILED。
3. **Resume 先 reconcile，再调用 LLM**：如果上次崩溃发生在 Tool 执行期间，先按 Tool 的 RetryPolicy 恢复；如果 ToolCall 已成功，则复用持久化结果，不重复执行。
4. **稳定 idempotency key**：`tool:<tool_call_id>` 跨进程保持不变。
5. **Tool 显式声明 RetryPolicy**：只读 deterministic 工具可以重试；副作用工具必须自己定义幂等协议。

## 最关键的工程认识

> execution state ≠ conversation state。

只保存数据库中的“任务状态”无法恢复 Agent；必须保存能让 LLM 从断点继续工作的 logical state，也就是 messages、pending ToolCall 和对应结果。

## 崩溃窗口

| 崩溃点 | 重启行为 |
|---|---|
| LLM 返回后、checkpoint 前 | 该版本通过先持久化 assistant tool-call message，避免进入不可恢复窗口 |
| Tool 执行中 | ToolCall=RUNNING，Resume 根据 RetryPolicy 恢复 |
| Tool 执行完成、DB commit 后 | ToolCall=SUCCEEDED，Resume 直接复用 result |
| Tool result checkpoint 后 | 直接恢复完整 messages，继续 LLM |

真正的生产系统还需要 transaction/outbox、lease/fencing token、worker heartbeat、并发 claim，以及副作用系统提供幂等键。
