MODEL ?= qwen3:4b
DB ?= ./data/agent.db

.PHONY: run fmt test

run:
	LLM_MODEL=$(MODEL) DB_PATH=$(DB) go run ./cmd/chat

fmt:
	go fmt ./...

test:
	go test ./...
