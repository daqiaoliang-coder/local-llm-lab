package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"local-llm-lab/internal/db"
	"local-llm-lab/internal/tools"
)

type RecoveryManager struct {
	store *db.Store
	tools *tools.Registry
}

func NewRecoveryManager(store *db.Store, registry *tools.Registry) *RecoveryManager {
	return &RecoveryManager{store: store, tools: registry}
}

func (r *RecoveryManager) Recover(ctx context.Context) error {
	items, err := r.store.FindRunningToolCalls()
	if err != nil {
		return err
	}

	for _, item := range items {
		log.Printf("[recovery] found RUNNING tool_call id=%s tool=%s attempt=%d/%d",
			item.ID, item.ToolName, item.Attempt, item.MaxAttempts)

		if !item.Retryable {
			msg := "recovery refused: tool is not retryable"
			_ = r.store.MarkToolCallFailed(item.ID, msg)
			log.Printf("[recovery] failed id=%s: %s", item.ID, msg)
			continue
		}

		if item.Attempt >= item.MaxAttempts {
			msg := "recovery refused: max attempts reached"
			_ = r.store.MarkToolCallFailed(item.ID, msg)
			log.Printf("[recovery] failed id=%s: %s", item.ID, msg)
			continue
		}

		tool, ok := r.tools.Get(item.ToolName)
		if !ok {
			msg := fmt.Sprintf("recovery failed: unknown tool %s", item.ToolName)
			_ = r.store.MarkToolCallFailed(item.ID, msg)
			continue
		}

		if err := tools.ValidateJSON([]byte(item.Arguments)); err != nil {
			_ = r.store.MarkToolCallFailed(item.ID, err.Error())
			continue
		}

		if err := r.store.MarkToolCallRetrying(item.ID); err != nil {
			return err
		}

		result, toolErr := tool.Execute(ctx, []byte(item.Arguments))
		if toolErr != nil {
			msg := toolErr.Error()
			_ = r.store.CompleteToolCall(item.ID, "FAILED", "", msg)
			log.Printf("[recovery] retry failed id=%s err=%s", item.ID, msg)
			continue
		}

		if err := r.store.CompleteToolCall(item.ID, "SUCCEEDED", result, ""); err != nil {
			return err
		}

		log.Printf("[recovery] retry succeeded id=%s", item.ID)
	}

	return nil
}

func EncodeCheckpointState(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
