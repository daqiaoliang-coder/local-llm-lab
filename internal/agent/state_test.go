package agent

import (
	openai "github.com/sashabaranov/go-openai"
	"testing"
)

func TestCheckpointStateRoundTrip(t *testing.T) {
	in := CheckpointState{RunID: "r1", StepNo: 2, Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hello"}}, PendingToolCallIDs: []string{"tc1"}}
	raw, err := EncodeState(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeState(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.RunID != in.RunID || out.StepNo != in.StepNo || len(out.PendingToolCallIDs) != 1 || out.PendingToolCallIDs[0] != "tc1" || len(out.Messages) != 1 || out.Messages[0].Content != "hello" {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}
