package agent

import (
	"encoding/json"
	openai "github.com/sashabaranov/go-openai"
)

type CheckpointState struct {
	RunID              string                         `json:"run_id"`
	StepNo             int                            `json:"step_no"`
	Messages           []openai.ChatCompletionMessage `json:"messages"`
	PendingToolCallIDs []string                       `json:"pending_tool_call_ids,omitempty"`
}

func EncodeState(s CheckpointState) (string, error) { b, err := json.Marshal(s); return string(b), err }
func DecodeState(raw string) (CheckpointState, error) {
	var s CheckpointState
	err := json.Unmarshal([]byte(raw), &s)
	return s, err
}
