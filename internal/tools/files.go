package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type listFilesTool struct{}

func NewListFilesTool() Tool { return &listFilesTool{} }

func (t *listFilesTool) Name() string { return "list_files" }

func (t *listFilesTool) Description() string {
	return "列出指定目录下的文件和子目录。参数: {\"path\":\".\"}"
}

func (t *listFilesTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", err
	}
	if in.Path == "" {
		in.Path = "."
	}

	entries, err := os.ReadDir(in.Path)
	if err != nil {
		return "", err
	}

	out := ""
	for _, e := range entries {
		kind := "file"
		if e.IsDir() {
			kind = "dir"
		}
		out += fmt.Sprintf("%s\t%s\n", kind, filepath.Join(in.Path, e.Name()))
	}
	return out, nil
}

type readFileTool struct{}

func NewReadFileTool() Tool { return &readFileTool{} }

func (t *readFileTool) Name() string { return "read_file" }

func (t *readFileTool) Description() string {
	return "读取一个文本文件。参数: {\"path\":\"README.md\"}"
}

func (t *readFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", err
	}
	if in.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(in.Path)
	if err != nil {
		return "", err
	}
	const maxBytes = 64 * 1024
	if len(data) > maxBytes {
		data = data[:maxBytes]
	}
	return string(data), nil
}
