package tools

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type Tool interface {
	Name() string
	Description() string
	Definition() openai.FunctionDefinition
	Execute(ctx context.Context, params map[string]any) (any, error)
}

type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: map[string]Tool{},
	}
}

func (r *ToolRegistry) RegisterTool(t Tool) {
	r.tools[t.Name()] = t
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(ctx, params)
}

func (r *ToolRegistry) GetTool(name string) Tool {
	return r.tools[name]
}

// called by AskModelAndHandleTools
func (r *ToolRegistry) FunctionDefinitions() []openai.FunctionDefinition {
	defs := []openai.FunctionDefinition{}
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}
