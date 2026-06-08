package tools

import (
	"ai-service-go/internals/types"
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	// "ai-service-go/internals/vector_db"
)

type GetUserGuideInformation struct {
	VectorManager interface {
		SearchSimilarChunks(query string, nChunks int, threshold float32) (string, error)
		GetAllVectorsWithMetadata() ([]map[string]interface{}, error)
	}
}

func (gl *GetUserGuideInformation) Name() string {
	return "get_user_guide_info"
}
func (gl *GetUserGuideInformation) Description() string {
	description := "Searches platform documentation and user guides for general platform usage, features, workflows, and how-to information. Use specifically for non-layer-specific questions about platform functionality."
	vectorsWithMetaData, err := gl.VectorManager.GetAllVectorsWithMetadata()
	if err != nil {
		return description
	}
	description += "It contains following titles:\n"
	description += strings.Join(getAllVectorTitles(vectorsWithMetaData), "\n")
	return description
}

func (gl *GetUserGuideInformation) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	query := params["query"].(string)
	n_chunks := params["n_chunks"]
	n_chunksInt, ok := n_chunks.(int)
	if !ok && n_chunksInt <= 0 {
		n_chunksInt = 2
	}
	return gl.VectorManager.SearchSimilarChunks(query, n_chunksInt, 0)
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
				"n_chunks": map[string]any{
					"type":        "integer",
					"description": "The number of relevant chunks to retrieve from the user guide. Default is 2.",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (gl *GetUserGuideInformation) InformationMessage() struct {
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Reading from user guide...`,
		End:   `User guide information retrieved.`,
	}
}

func getAllVectorTitles(vectors []map[string]interface{}) []string {
	titles := []string{}
	for _, vector := range vectors {
		metadata, ok := vector["metadata"].(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := metadata["title"].(string)
		titles = append(titles, title)
	}
	return titles
}
