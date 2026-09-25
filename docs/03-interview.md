# v0.4 Interview Notes

### Q1：v0.3 为什么还不算真正的 Durable Execution？
因为只恢复了 ToolCall，而没有恢复 Agent 的 logical execution state。真正 Resume 必须恢复 messages / pending tool calls / completed tool results，然后继续同一个 Run。

### Q2：为什么 checkpoint 要保存 messages？
LLM 是无状态调用。数据库里的 step status 不能告诉下一次 LLM“之前已经说了什么”。messages 是下一次推理的输入，因此属于 durable logical state。

### Q3：为什么不能简单地把 RUNNING ToolCall 再执行一次？
因为执行可能已经成功，只是进程死在“业务副作用完成、结果尚未落库”的窗口。重试可能造成重复副作用。因此生产级实现需要 idempotency key + side-effect contract。

### Q4：checkpoint 与数据库事务是什么关系？
checkpoint 解决跨进程恢复；事务解决单次状态转换的原子性。二者不能互相替代。生产系统还需要 outbox/transactional messaging 处理 DB 与 MQ 的一致性。

### Q5：这个项目和 LangGraph / Durable Execution 平台的差异？
这个实验刻意聚焦 Runtime 最小内核：Run/Step/ToolCall、durable state、checkpoint、recovery、resume、idempotency。成熟框架通常还覆盖图编排、human-in-the-loop、状态存储、可观测性、分布式 worker、租约/fencing、权限和生产运维。
