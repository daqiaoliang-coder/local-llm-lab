# v0.2 Design

## Native Tool Calling

v0.1 是文本协议：

```text
TOOL_CALL {"name":"...","args":{...}}
```

v0.2 使用 OpenAI-compatible `tools`：

```text
assistant.tool_calls[]
```

因此 ToolCall ID、名称、arguments、tool result 都可以成为明确的 Runtime 数据。

## Durable State

```text
Run
 ├── Step
 │    └── ToolCall
 └── Step
```

- Run：一次完整 Agent 执行
- Step：Runtime 的一次推进
- ToolCall：一次具体工具执行

## 为什么 ToolCall 独立建模？

Tool 是能力定义；ToolCall 是一次执行。

未来可以在 ToolCall 上增加：

- retry_count
- timeout
- lease
- idempotency_key
- worker_id
- token/cost
- started_at / finished_at

## 当前一致性边界

当前已经持久化：

```text
Create Run
Create Step
Create ToolCall
Execute Tool
Complete ToolCall
Complete Step
Complete Run
```

但如果：

```text
ToolCall = RUNNING
process crash
```

仍需要下一版本的 Recovery Manager 判断能否安全重试。

## 下一步

```text
RecoveryManager
 → scan RUNNING state
 → classify
 → retry / resume / manual intervention
```

再加入 Checkpoint、Idempotency、Lease。
