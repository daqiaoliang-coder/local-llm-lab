package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}
