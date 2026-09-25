package tools

import (
	"context"
	"encoding/json"
	"sort"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(context.Context, json.RawMessage) (string, error)
}

type Registry struct { tools map[string]Tool }

func NewRegistry() *Registry { return &Registry{tools: map[string]Tool{}} }
func (r *Registry) Register(t Tool) { r.tools[t.Name()] = t }
func (r *Registry) Get(name string) (Tool, bool) { t, ok := r.tools[name]; return t, ok }

func (r *Registry) Specs() []struct {
	Name string
	Description string
	Parameters map[string]any
} {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools { names = append(names, name) }
	sort.Strings(names)

	out := make([]struct {
		Name string
		Description string
		Parameters map[string]any
	}, 0, len(names))

	for _, name := range names {
		t := r.tools[name]
		out = append(out, struct {
			Name string
			Description string
			Parameters map[string]any
		}{t.Name(), t.Description(), t.Parameters()})
	}
	return out
}

func ValidateJSON(raw json.RawMessage) error {
	var v any
	return json.Unmarshal(raw, &v)
}
