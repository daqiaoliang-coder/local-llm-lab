package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
	"local-llm-lab/internal/db"
	"local-llm-lab/internal/llm"
	"local-llm-lab/internal/tools"
)

type Agent struct {
	llm   llm.Client
	tools *tools.Registry
	store *db.Store
}

func New(client llm.Client, registry *tools.Registry, store *db.Store) *Agent {
	return &Agent{llm: client, tools: registry, store: store}
}

func (a *Agent) Run(ctx context.Context, input string) (string, string, error) {
	runID := uuid.NewString()
	if err := a.store.CreateRun(runID, input); err != nil {
		return "", runID, err
	}
	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: "你是一个本地 Agent Infra 学习助手。需要读取工作区文件时使用工具；不要虚构工具结果。"}, {Role: openai.ChatMessageRoleUser, Content: input}}
	return a.execute(ctx, runID, messages, 1)
}

func (a *Agent) Resume(ctx context.Context, runID string) (string, error) {
	r, err := a.store.GetRun(runID)
	if err != nil {
		return "", err
	}
	if r.Status != "RUNNING" {
		return r.Output, nil
	}
	cp, err := a.store.LatestCheckpoint(runID)
	if err != nil {
		return "", fmt.Errorf("load checkpoint: %w", err)
	}
	st, err := DecodeState(cp.State)
	if err != nil {
		return "", fmt.Errorf("decode checkpoint: %w", err)
	}
	// Reconcile pending ToolCalls before asking the LLM again.
	for _, id := range st.PendingToolCallIDs {
		tc, err := a.store.GetToolCall(id)
		if err != nil {
			return "", err
		}
		if tc.Status == "RUNNING" || tc.Status == "RETRYING" {
			if err := a.recoverOne(ctx, tc); err != nil {
				return "", err
			}
			tc, err = a.store.GetToolCall(id)
			if err != nil {
				return "", err
			}
		}
		if tc.Status == "SUCCEEDED" {
			st.Messages = appendToolResultIfMissing(st.Messages, tc.ID, tc.Result)
		} else if tc.Status == "FAILED" {
			st.Messages = appendToolResultIfMissing(st.Messages, tc.ID, "tool error: "+tc.Error)
		}
	}
	st.PendingToolCallIDs = nil
	if state, err := EncodeState(st); err == nil {
		_ = a.store.CreateCheckpoint(uuid.NewString(), runID, cp.StepID, state)
	}
	out, _, err := a.execute(ctx, runID, st.Messages, st.StepNo)
	return out, err
}

func (a *Agent) execute(ctx context.Context, runID string, messages []openai.ChatCompletionMessage, startStep int) (string, string, error) {
	specs := make([]llm.ToolSpec, 0)
	for _, t := range a.tools.Specs() {
		specs = append(specs, llm.ToolSpec{Name: t.Name, Description: t.Description, Parameters: t.Parameters})
	}
	for stepNo := startStep; stepNo <= 8; stepNo++ {
		stepID := uuid.NewString()
		if err := a.store.CreateStep(stepID, runID, stepNo, "LLM", lastUserInput(messages)); err != nil {
			return a.fail(runID, err)
		}
		res, err := a.llm.Chat(ctx, messages, specs)
		if err != nil {
			_ = a.store.CompleteStep(stepID, "FAILED", "", err.Error())
			return a.fail(runID, err)
		}
		if len(res.ToolCalls) == 0 {
			_ = a.store.CompleteStep(stepID, "SUCCEEDED", res.Content, "")
			_ = a.store.CompleteRun(runID, "SUCCEEDED", res.Content, "")
			return res.Content, runID, nil
		}
		// First persist every ToolCall record, then persist the checkpoint. This avoids
		// a recovery window where the checkpoint references a ToolCall that does not exist yet.
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: res.Content, ToolCalls: res.ToolCalls})
		callIDs := make([]string, len(res.ToolCalls))
		toolSteps := make([]string, len(res.ToolCalls))
		for i, call := range res.ToolCalls {
			callID := call.ID
			if callID == "" {
				callID = uuid.NewString()
				res.ToolCalls[i].ID = callID
				messages[len(messages)-1].ToolCalls[i].ID = callID
			}
			callIDs[i] = callID
			toolStepID := uuid.NewString()
			toolSteps[i] = toolStepID
			if err := a.store.CreateStep(toolStepID, runID, stepNo, "TOOL", call.Function.Arguments); err != nil {
				return a.fail(runID, err)
			}
			tool, ok := a.tools.Get(call.Function.Name)
			if !ok {
				_ = a.store.CompleteStep(toolStepID, "FAILED", "", fmt.Sprintf("unknown tool: %s", call.Function.Name))
				continue
			}
			policy := tool.RetryPolicy()
			if policy.MaxAttempts < 1 {
				policy.MaxAttempts = 1
			}
			if err := a.store.CreateToolCall(callID, runID, toolStepID, call.Function.Name, call.Function.Arguments, "tool:"+callID, policy.Retryable, policy.MaxAttempts); err != nil {
				return a.fail(runID, err)
			}
		}
		if state, err := EncodeState(CheckpointState{RunID: runID, StepNo: stepNo, Messages: messages, PendingToolCallIDs: callIDs}); err == nil {
			if err := a.store.CreateCheckpoint(uuid.NewString(), runID, stepID, state); err != nil {
				return a.fail(runID, err)
			}
		} else {
			return a.fail(runID, err)
		}

		for i, call := range res.ToolCalls {
			callID := callIDs[i]
			toolStepID := toolSteps[i]
			tool, ok := a.tools.Get(call.Function.Name)
			if !ok {
				msg := fmt.Sprintf("unknown tool: %s", call.Function.Name)
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: callID, Content: msg})
				remaining := append([]string(nil), callIDs[i+1:]...)
				if state, err := EncodeState(CheckpointState{RunID: runID, StepNo: stepNo, Messages: messages, PendingToolCallIDs: remaining}); err == nil {
					_ = a.store.CreateCheckpoint(uuid.NewString(), runID, toolStepID, state)
				}
				continue
			}
			if err := tools.ValidateJSON([]byte(call.Function.Arguments)); err != nil {
				_ = a.store.CompleteToolCall(callID, "FAILED", "", err.Error())
				_ = a.store.CompleteStep(toolStepID, "FAILED", "", err.Error())
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: callID, Content: err.Error()})
				remaining := append([]string(nil), callIDs[i+1:]...)
				if state, e := EncodeState(CheckpointState{RunID: runID, StepNo: stepNo, Messages: messages, PendingToolCallIDs: remaining}); e == nil {
					_ = a.store.CreateCheckpoint(uuid.NewString(), runID, toolStepID, state)
				}
				continue
			}
			result, toolErr := tool.Execute(ctx, []byte(call.Function.Arguments))
			if toolErr != nil {
				msg := "tool error: " + toolErr.Error()
				_ = a.store.CompleteToolCall(callID, "FAILED", "", msg)
				_ = a.store.CompleteStep(toolStepID, "FAILED", "", msg)
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: callID, Content: msg})
			} else {
				_ = a.store.CompleteToolCall(callID, "SUCCEEDED", result, "")
				_ = a.store.CompleteStep(toolStepID, "SUCCEEDED", result, "")
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: callID, Content: result})
			}
			remaining := append([]string(nil), callIDs[i+1:]...)
			if state, err := EncodeState(CheckpointState{RunID: runID, StepNo: stepNo, Messages: messages, PendingToolCallIDs: remaining}); err == nil {
				_ = a.store.CreateCheckpoint(uuid.NewString(), runID, toolStepID, state)
			}
		}
		_ = a.store.CompleteStep(stepID, "SUCCEEDED", res.Content, "")
	}
	return a.fail(runID, fmt.Errorf("maximum agent steps exceeded"))
}

func (a *Agent) recoverOne(ctx context.Context, tc *db.ToolCallRecord) error {
	tool, ok := a.tools.Get(tc.ToolName)
	if !ok {
		return fmt.Errorf("recovery: unknown tool %s", tc.ToolName)
	}
	if tc.Status == "RUNNING" {
		if !tc.Retryable || tc.Attempt >= tc.MaxAttempts {
			return a.store.MarkToolCallFailed(tc.ID, "crash recovery: retry policy exhausted or non-retryable")
		}
		if err := a.store.MarkToolCallRetrying(tc.ID); err != nil {
			return err
		}
	}
	result, err := tool.Execute(ctx, []byte(tc.Arguments))
	if err != nil {
		msg := "recovery tool error: " + err.Error()
		_ = a.store.MarkToolCallFailed(tc.ID, msg)
		return fmt.Errorf(msg)
	}
	return a.store.CompleteToolCall(tc.ID, "SUCCEEDED", result, "")
}

func appendToolResultIfMissing(ms []openai.ChatCompletionMessage, id, result string) []openai.ChatCompletionMessage {
	for _, m := range ms {
		if m.Role == openai.ChatMessageRoleTool && m.ToolCallID == id {
			return ms
		}
	}
	return append(ms, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: id, Content: result})
}
func (a *Agent) fail(runID string, err error) (string, string, error) {
	_ = a.store.CompleteRun(runID, "FAILED", "", err.Error())
	return "", runID, err
}
func lastUserInput(ms []openai.ChatCompletionMessage) string {
	var b strings.Builder
	for _, m := range ms {
		if m.Role == openai.ChatMessageRoleUser {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(m.Content)
		}
	}
	return b.String()
}
