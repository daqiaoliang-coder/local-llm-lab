package llm

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OllamaClient struct {
	client *openai.Client
	model  string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	cfg := openai.DefaultConfig("ollama")
	cfg.BaseURL = baseURL
	return &OllamaClient{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

func (c *OllamaClient) Chat(ctx context.Context, messages []Message) (string, error) {
	reqMessages := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, m := range messages {
		reqMessages = append(reqMessages, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: reqMessages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
