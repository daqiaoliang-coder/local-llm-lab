package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"local-llm-lab/internal/agent"
	"local-llm-lab/internal/db"
	"local-llm-lab/internal/llm"
	"local-llm-lab/internal/tools"
)

func main() {
	ctx := context.Background()
	baseURL := getenv("LLM_BASE_URL", "http://localhost:11434/v1")
	model := getenv("LLM_MODEL", "qwen3:4b")
	dbPath := getenv("DB_PATH", "./data/agent.db")
	workspace := getenv("LAB_WORKSPACE", ".")

	store, err := db.Open(dbPath)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	client := llm.NewOllamaClient(baseURL, model)
	registry := tools.NewRegistry()
	registry.Register(tools.NewListFilesTool(workspace))
	registry.Register(tools.NewReadFileTool(workspace))

	ag := agent.New(client, registry, store)
	recovery := agent.NewRecoveryManager(store, registry, ag)
	if err := recovery.Recover(ctx); err != nil {
		fmt.Printf("[recovery] error=%v\n", err)
	}

	fmt.Printf("Local LLM Lab v0.4 | model=%s | db=%s\n", model, dbPath)
	fmt.Println("输入 exit 退出。")

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

		answer, runID, err := ag.Run(ctx, input)
		if err != nil {
			fmt.Printf("run=%s error=%v\n\n", runID, err)
			continue
		}
		fmt.Printf("run=%s\nagent> %s\n\n", runID, answer)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
