# Crash / Resume Sequence

```text
                    Process A
                       │
                       ▼
                  LLM response
                       │
                       ▼
             persist assistant message
                       │
                       ▼
               ToolCall = RUNNING
                       │
                 execute tool
                       │
                💥 process crash
                       │
═══════════════════════╪════════════════════════════════
                       │ restart
                       ▼
                RecoveryManager
                       │
                       ▼
                 load RUNNING run
                       │
                       ▼
                load latest checkpoint
                       │
                       ▼
               inspect ToolCall state
                       │
              ┌────────┴────────┐
              │                 │
          SUCCEEDED          RUNNING
              │                 │
        reuse result      retry if allowed
              │                 │
              └────────┬────────┘
                       ▼
               append tool result
                       │
                       ▼
                  call LLM
                       │
                       ▼
                 continue Run
```
