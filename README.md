# Local LLM Lab v0.2

v0.2 在 v0.1 的 Ollama + Agent Loop 基础上加入：
- Native Tool Calling
- SQLite Durable State
- Run / Step / ToolCall 持久化

## 启动

```bash
ollama pull qwen3:4b
ollama serve
```

另开终端：

```bash
go mod tidy
go run ./cmd/chat
```

默认环境变量：
- LLM_BASE_URL=http://localhost:11434/v1
- LLM_MODEL=qwen3:4b
- DB_PATH=./data/agent.db
- LAB_WORKSPACE=.

示例：

```text
you> 请查看当前目录有哪些文件
```

执行链路：

```text
User
  ↓
Run
  ↓
LLM Step
  ↓
assistant.tool_calls[]
  ↓
ToolCall
  ↓
Tool Result
  ↓
LLM
  ↓
Final Answer
```

SQLite 核心表：

```text
runs
steps
tool_calls
```

查看：

```bash
sqlite3 ./data/agent.db
```

```sql
select id,status,input,output from runs order by created_at desc;
select id,run_id,step_no,kind,status from steps order by created_at desc;
select id,run_id,step_id,tool_name,status from tool_calls order by created_at desc;
```

## v0.2 的重要边界

已经持久化 Durable State，但还没有真正完成 Crash Recovery。

例如：

```text
ToolCall = RUNNING
      ↓
process crash
```

下一版应增加：

- Recovery Manager
- Retry Policy
- Checkpoint / Resume
- Idempotency Key
- Lease

## 演进路线

```text
v0.1 Local LLM + Agent Loop
v0.2 Native Tool Calling + Durable State
v0.3 Checkpoint + Resume + Retry
v0.4 RAG + Memory
v0.5 Event-driven Worker
v0.6 Dynamic DAG
v0.7 Sandbox / Remote Tool Provider
v1.0 Local Agent Runtime
```
