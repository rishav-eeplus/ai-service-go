package orchestrator

import (
	"ai-service-go/internals/types"
	"ai-service-go/internals/logger"
	"context"
	"fmt"
	"strings"
)

// Retreiver is responsible for checking any relevant chunks available in our data source to make some context available for the LLM.

func (o *Orchestrator) Retreiver(sendMessage func(msg types.StreamMessage) bool, ctx context.Context, input *ClientRequestType, intents []string) (string, error) {
	sendMessage(types.StreamMessage{
		Type:    "info",
		Message: "Retrieving relevant context from vector store...",
	})
	vectors, err := o.VectorManager.GetAllVectorsWithMetadata()
	if err != nil {
		logger.Logger.Errorf("Error while getting vector store context for query: %s", input.UserQuery)
		return "", err
	}
	// Build a summary of vectors for AI to evaluate
	var vectorSummaries []string
	var vectorIDMap = make(map[int]uint64)
	for i, vector := range vectors {
		metadata, ok := vector["metadata"].(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := metadata["title"].(string)
		summary, _ := metadata["summary"].(string)
		id, _ := vector["id"].(uint64)
		vectorIDMap[i] = id
		vectorSummaries = append(vectorSummaries, fmt.Sprintf("Vector ID: %d, Title: %s, Summary: %s", id, title, summary))
	}
	// Ask AI to identify relevant vectors
	vectorSelectionSchema := map[string]interface{}{
		"type":        "object",
		"description": "Response containing list of vector IDs that are relevant to answer the user query.",
		"properties": map[string]interface{}{
			"relevant_vector_ids": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "integer",
				},
				"description": "Array of vector IDs that contain relevant information to answer the user query",
			},
		},
		"required":             []string{"relevant_vector_ids"},
		"additionalProperties": false,
	}

	intentNames := []string{}
	for _, intent := range intents {
		intentNames = append(intentNames, intent)
	}

	instructions := fmt.Sprintf(`Based on the user's query %s, past conversations %s, and the identified intents %v, examine the following vector summaries and determine which vectors contain relevant information that would be helpful to answer the query.
						Return the IDs of relevant vectors. If no vectors are particularly relevant beyond what the intents already cover, return an empty array.
						Vector Summaries:
						%s`, input.UserQuery, input.PreviousConversation, intentNames, strings.Join(vectorSummaries, "\n"))
	rawResponse, _, err := o.AIManager.GetAIResponse(instructions, input.UserQuery, input.PreviousConversation, "", vectorSelectionSchema, input.Platform)
	if err != nil {
		logger.Logger.Errorf("Error getting AI response for vector selection: %v", err)
		return "", err
	}
	responseMap, ok := (*rawResponse).(map[string]interface{})
	if !ok {
		logger.Logger.Errorf("Invalid response format from AI for query: %s", input.UserQuery)
		return "", fmt.Errorf("invalid response format")
	}

	// Parse the vector IDs array properly
	idsArray, ok := responseMap["relevant_vector_ids"].([]interface{})
	if !ok {
		logger.Logger.Warn("No relevant vector IDs found, continuing without them")
		idsArray = []interface{}{}
	}

	var selectedVectors []uint64
	for _, item := range idsArray {
		// Handle both float64 (JSON number) and int
		switch v := item.(type) {
		case float64:
			selectedVectors = append(selectedVectors, uint64(v))
		case int:
			selectedVectors = append(selectedVectors, uint64(v))
		case int64:
			selectedVectors = append(selectedVectors, uint64(v))
		case uint64:
			selectedVectors = append(selectedVectors, v)
		}
	}

	relevantVectorContents, err := o.VectorManager.GetContentByIDs(selectedVectors)
	if err != nil {
		logger.Logger.Errorf("Error while getting relevant vectors from AI for query: %s", input.UserQuery)
		return "", err
	}
	contextForPlanning := "User Query: " + input.UserQuery + "\n"
	contextForPlanning += "Previous Conversations: " + input.PreviousConversation + "\n"
	// contextForPlanning += "Relevant information from vector store:"
	sendMessage(types.StreamMessage{
		Type:    "info",
		Message: "Retrieved relevant context from vector store. Got " + fmt.Sprintf("%d", len(relevantVectorContents)) + " relevant chunks.",
	})
	return contextForPlanning, nil
}
