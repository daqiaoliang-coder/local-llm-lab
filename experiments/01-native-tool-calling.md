# Experiment 01 - Native Tool Calling

输入：

```text
请查看当前目录有哪些文件。
```

观察：

```text
LLM
 ↓
assistant.tool_calls[]
 ↓
list_files
 ↓
tool message
 ↓
LLM
 ↓
final answer
```

重点：
- Tool schema
- ToolCall ID
- JSON arguments
- Tool result
- 多 ToolCall
