package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) Describe() string {
	out := ""
	for _, t := range r.tools {
		out += fmt.Sprintf("- %s: %s\n", t.Name(), t.Description())
	}
	return out
}
