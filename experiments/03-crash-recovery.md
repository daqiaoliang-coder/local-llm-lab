# Experiment 03 - Crash Recovery

v0.2 暂不实现自动恢复。

思考：

1. 哪些 ToolCall 可以重试？
2. Tool 是否幂等？
3. 是否需要 idempotency key？
4. 外部副作用是否已经发生？
5. 如何区分执行失败和结果丢失？
6. Step 继续还是重建？
7. Run 如何恢复？

这些问题是 v0.3 的核心。
