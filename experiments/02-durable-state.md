# Experiment 02 - Durable State

输入：

```text
请读取 README.md 并总结。
```

查看：

```sql
select * from runs;
select * from steps;
select * from tool_calls;
```

目标：

```text
Run
 ├── LLM Step
 ├── TOOL Step
 │    └── ToolCall
 └── LLM Step
```
