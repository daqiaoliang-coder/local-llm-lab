package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"local-llm-lab/internal/agent"
	"local-llm-lab/internal/llm"
	"local-llm-lab/internal/tools"
)

func main() {
	baseURL := getenv("LLM_BASE_URL", "http://localhost:11434/v1")
	model := getenv("LLM_MODEL", "qwen3:4b")

	client := llm.NewOllamaClient(baseURL, model)
	registry := tools.NewRegistry()
	registry.Register(tools.NewListFilesTool())
	registry.Register(tools.NewReadFileTool())

	ag := agent.New(client, registry)

	fmt.Printf("Local LLM Lab v0.1 | model=%s | endpoint=%s\n", model, baseURL)
	fmt.Println("输入 exit 退出。")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("you> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		answer, err := ag.Run(context.Background(), input)
		if err != nil {
			fmt.Printf("error: %v\n\n", err)
			continue
		}
		fmt.Printf("agent> %s\n\n", answer)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
