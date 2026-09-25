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
	return &OllamaClient{client: openai.NewClientWithConfig(cfg), model: model}
}

func (c *OllamaClient) Chat(ctx context.Context, messages []openai.ChatCompletionMessage, tools []ToolSpec) (ChatResult, error) {
	reqTools := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		reqTools = append(reqTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model, Messages: messages, Tools: reqTools,
	})
	if err != nil {
		return ChatResult{}, err
	}
	if len(resp.Choices) == 0 {
		return ChatResult{}, nil
	}

	return ChatResult{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: resp.Choices[0].Message.ToolCalls,
	}, nil
}
