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

// ToolDefinitions converts function definitions to tool definitions for the new API
func (r *ToolRegistry) ToolDefinitions() []openai.Tool {
	tools := []openai.Tool{}
	for _, t := range r.tools {
		def := t.Definition()
		tools = append(tools, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &def,
		})
	}
	return tools
}

func (r *ToolRegistry) AllTools() []Tool {
	tools := []Tool{}
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}
