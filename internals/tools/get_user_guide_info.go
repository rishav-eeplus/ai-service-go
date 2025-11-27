package tools

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	// "ai-service-go/internals/vector_db"
)

type GetUserGuideInformation struct {
	VectorManager interface {
		SearchSimilarChunks(query string, nChunks int, threshold float32) (string, error)
	}
}

func (gl *GetUserGuideInformation) Name() string {
	return "get_user_guide_info"
}
func (gl *GetUserGuideInformation) Description() string {
	return "Fetches general information and acts as a manual for users on how to effectively utilize data platform. This is go to tool when no other tool can be used to answer the user query."
}

func (gl *GetUserGuideInformation) Execute(ctx context.Context, params map[string]any) (any, error) {
	query := params["query"].(string)
	n_chunks := 3
	return gl.VectorManager.SearchSimilarChunks(query, n_chunks, 0)
}

func (gl *GetUserGuideInformation) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The user's query for which information is to be fetched from the user guide.",
				},
			},
			"required": []string{"query"},
		},
	}
}
