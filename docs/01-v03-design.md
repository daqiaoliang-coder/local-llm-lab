# v0.3 Design

## 1. Durable Execution

v0.2 解决：

```text
状态可以落盘
```

v0.3 进一步解决：

```text
进程崩溃后知道“当时做到哪里”
```

因此增加：

```text
Checkpoint
RecoveryManager
Retry Policy
Idempotency Key
```

## 2. Recovery

启动：

```text
Process Start
    ↓
RecoveryManager
    ↓
scan RUNNING ToolCall
    ↓
Retry Policy
    ├── retry
    └── fail
```

当前默认只对只读文件 Tool 自动重试。

## 3. 为什么不能所有 Tool 都 retry？

因为：

```text
read_file
```

通常没有副作用。

但：

```text
create_order
charge_payment
send_email
delete_resource
```

可能有副作用。

如果：

```text
external side effect succeeded
      ↓
process crashed
      ↓
result lost
```

再次执行可能产生重复副作用。

所以真正生产级 Runtime 必须要求 Tool 提供：

```text
Retry Policy
Idempotency Contract
Side-effect Class
```

## 4. Checkpoint

Checkpoint 的意义：

```text
Tool succeeded
      ↓
persist checkpoint
      ↓
continue
```

Crash 后不一定要从头开始。

下一版可以把：

```text
messages
current step
completed tool calls
```

完整恢复。

## 5. v0.3 仍然没有解决的问题

当前 RecoveryManager 能恢复 ToolCall，但还没有做到完整：

```text
Run Resume
```

即：

```text
Run
 ↓
LLM state
 ↓
ToolCall
 ↓
Crash
 ↓
Restart
 ↓
恢复原 Agent message state
 ↓
继续 LLM
```

这是后续真正的 Resume 机制。
