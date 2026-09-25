# Agent Loop

最小模型：

```text
User
 ↓
LLM
 ↓
Need Tool?
 ├─ No → Final Answer
 └─ Yes
      ↓
    Tool
      ↓
 Observation
      ↓
     LLM
```

需要重点观察：

1. LLM 如何决定是否调用工具？
2. Tool 参数如何校验？
3. Tool 执行失败如何返回？
4. 如何避免无限循环？
5. Tool 超时如何处理？
6. Agent 崩溃后如何恢复？
7. Tool 是否需要幂等？
8. 哪些工具需要权限门禁？

这些问题逐步会从“Agent 开发”进入“Agent Infra”。
