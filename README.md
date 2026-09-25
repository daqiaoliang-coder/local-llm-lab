# Local LLM Lab v0.3

v0.3 在 v0.2 的 Native Tool Calling + SQLite Durable State 上加入：

- Checkpoint
- Crash Recovery
- Retry Policy
- Idempotency Key
- Recovery Manager
- Resume

目标：把本地 Agent 从“可持久化”推进到“可恢复执行”。

## 核心执行模型

```text
Run
 ├── Step
 │    └── ToolCall
 │
 ├── Checkpoint
 │
 └── Recovery
```

典型故障：

```text
ToolCall = RUNNING
      ↓
process crash
      ↓
restart
      ↓
RecoveryManager
      ↓
判断是否可以安全重试
      ↓
Retry / Fail / Resume
```

## 启动

```bash
ollama pull qwen3:4b
ollama serve
go mod tidy
go run ./cmd/chat
```

默认：

```text
LLM_MODEL=qwen3:4b
DB_PATH=./data/agent.db
LAB_WORKSPACE=.
```

## 演示 Recovery

程序启动时会自动扫描上次异常退出遗留的：

```text
RUNNING runs
RUNNING steps
RUNNING tool_calls
```

并执行恢复策略。

当前文件工具属于只读、确定性较高的工具，因此默认允许自动重试。

## 为什么需要 Idempotency

假设：

```text
ToolCall
   ↓
外部系统已经执行成功
   ↓
Agent 在收到结果前 crash
```

重启后如果再次执行：

```text
ToolCall
   ↓
External Side Effect
```

可能产生重复副作用。

所以：

```text
ToolCall
  + idempotency_key
  + retry policy
```

是 Durable Agent 的重要基础。

本项目 v0.3 对文件读取等只读工具采用 deterministic retry；后续有副作用的工具应增加明确的 retry class。

## v0.3 新增数据

```text
tool_calls
├── attempt
├── max_attempts
├── idempotency_key
├── retryable
└── last_error

checkpoints
├── run_id
├── step_id
├── state
└── created_at
```

## 推荐实验

### 1. 正常执行

```text
请读取 README.md 并总结。
```

### 2. 人工制造 RUNNING ToolCall

可以使用 sqlite3：

```sql
update tool_calls
set status='RUNNING'
where id='...';
```

然后重启程序。

观察启动日志：

```text
[recovery] found RUNNING tool_call ...
[recovery] retrying ...
```

### 3. 思考真正的生产问题

如果 Tool 是：

```text
create_order
charge_payment
send_email
delete_resource
```

是否应该自动 retry？

答案不能由 Runtime 简单决定，而应该由 Tool 的 retry policy / idempotency contract 决定。

## 演进路线

```text
v0.1 Local LLM + Agent Loop
v0.2 Native Tool Calling + Durable State
v0.3 Recovery + Retry + Checkpoint + Idempotency
v0.4 RAG + Memory
v0.5 Event-driven Worker
v0.6 Dynamic DAG
v0.7 Sandbox / Remote Tool Provider
v1.0 Local Agent Runtime
```
