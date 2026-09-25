# Experiment 02 - Tool Calling

测试：

```text
请查看当前目录有哪些文件。
```

预期：

```text
LLM
 ↓
TOOL_CALL list_files
 ↓
Tool Result
 ↓
LLM
 ↓
Answer
```

观察工具调用格式、参数错误和失败处理。
