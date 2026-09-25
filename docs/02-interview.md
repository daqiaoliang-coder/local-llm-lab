# v0.3 面试要点

## Q1：为什么需要 Checkpoint？

因为 Durable State 只记录实体状态，不一定能完整恢复 Agent 的执行上下文。

Checkpoint 记录：

```text
Run
Step
ToolCall
Messages / execution state
```

使 Runtime 可以从最近安全点继续。

## Q2：为什么 RUNNING ToolCall 不能简单全部重试？

因为“执行结果未知”不等于“执行没有发生”。

尤其是有副作用的 Tool：

```text
charge()
```

可能已经扣款成功，只是 Agent 没收到结果。

所以必须结合：

- idempotency
- retry policy
- side-effect semantics

## Q3：idempotency key 放在哪里？

应该成为 ToolCall 的一等属性：

```text
ToolCall
 ├── id
 ├── tool_name
 ├── arguments
 └── idempotency_key
```

下游服务也应该能够识别该 key。

## Q4：SQLite 换 MySQL 难吗？

如果 Runtime 依赖的是：

```text
Store interface
```

而不是 SQLite SQL 细节，则可以替换 Durable Store。

生产环境还需要考虑：

- transaction
- isolation
- locking
- connection pool
- HA
- migration
- concurrent workers

## Q5：v0.3 最大技术难点是什么？

不是“把数据存 SQLite”。

真正难的是：

> **如何证明一次 ToolCall 在 crash 后可以安全地重新执行。**

这就是 Agent Infra 中 Durable Execution 的核心问题。
