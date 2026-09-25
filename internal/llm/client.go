package llm

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
)

type ToolSpec struct {
	Name string
	Description string
	Parameters map[string]any
}

type ChatResult struct {
	Content string
	ToolCalls []openai.ToolCall
}

type Client interface {
	Chat(context.Context, []openai.ChatCompletionMessage, []ToolSpec) (ChatResult, error)
}
