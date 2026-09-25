MODEL ?= qwen3:4b

.PHONY: run fmt test

run:
	LLM_MODEL=$(MODEL) go run ./cmd/chat

fmt:
	go fmt ./...

test:
	go test ./...
