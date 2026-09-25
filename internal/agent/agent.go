package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"local-llm-lab/internal/llm"
	"local-llm-lab/internal/tools"
)

type Agent struct {
	llm   llm.Client
	tools *tools.Registry
}

func New(client llm.Client, registry *tools.Registry) *Agent {
	return &Agent{llm: client, tools: registry}
}

func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	system := `你是一个本地 Agent Infra 学习助手。
你可以在需要时使用工具。
工具格式必须严格输出：
TOOL_CALL {"name":"tool_name","args":{...}}
如果不需要工具，直接输出最终答案。
可用工具：
` + a.tools.Describe()

	messages := []llm.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: input},
	}

	for step := 0; step < 5; step++ {
		out, err := a.llm.Chat(ctx, messages)
		if err != nil {
			return "", err
		}

		name, args, ok := parseToolCall(out)
		if !ok {
			return out, nil
		}

		tool, exists := a.tools.Get(name)
		if !exists {
			return "", fmt.Errorf("unknown tool: %s", name)
		}

		result, err := tool.Execute(ctx, args)
		if err != nil {
			result = "tool error: " + err.Error()
		}

		messages = append(messages,
			llm.Message{Role: "assistant", Content: out},
			llm.Message{Role: "user", Content: "TOOL_RESULT " + result},
		)
	}

	return "达到最大 Agent Step 数量。", nil
}

func parseToolCall(s string) (string, json.RawMessage, bool) {
	re := regexp.MustCompile(`(?s)TOOL_CALL\s+(\{.*\})`)
	m := re.FindStringSubmatch(s)
	if len(m) != 2 {
		return "", nil, false
	}

	var call struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal([]byte(m[1]), &call); err != nil {
		return "", nil, false
	}
	if strings.TrimSpace(call.Name) == "" {
		return "", nil, false
	}
	return call.Name, call.Args, true
}
