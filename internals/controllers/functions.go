package controllers

import (
	// "ai-service-go/internals/tools"
	"ai-service-go/internals/tools"
	"context"
	"fmt"
	"log"
	"strings"
)

var finalResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"result": map[string]any{
			"type":        "string",
			"description": "The response provided by the assistant.",
		},
		"followUps": map[string]interface{}{
			"type":        "array",
			"description": "An array of follow-up questions for further engagement with the user.",
			"items": map[string]interface{}{
				"type": "string",
			},
		},
	},
	"required":             []string{"result", "followUps"},
	"additionalProperties": false,
}

func extractGetRelevantLayers(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) ([]tools.AvailableLayersData, error) {
	all_layers, err := toolRegistry.Execute(context.Background(), "get_all_available_layers", map[string]any{}, nil)
	if err != nil {
		return nil, err
	}
	all_layersSlice, ok := all_layers.([]tools.AvailableLayersData)
	if !ok {
		log.Printf("unexpected type for all_layers: %T", all_layers)
	}
	var input any
	input = all_layersSlice
	if len(all_layersSlice) == 0 {
		input = all_layers
	}
	// EXTRACT RELEVANT LAYERS FROM all_layers BASED ON userQuery
	instructions := fmt.Sprintf(`Using the list of layers having their name and descriptions provided, return the relevant layers that match the user's query based. 
	                 	Be accurate in matching layers name. Return an array of layer names only that are relevant to the user's query. 
						If no layers match, return an empty array. Here is the list of layers and their descriptions: %v`, input)
	layerSchema := map[string]interface{}{
		"type":        "object",
		"description": "Response containing list of relevant layer names that match the user's query.",
		"properties": map[string]interface{}{
			"layers": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Array of layer names relevant to the user query",
			},
		},
		"required":             []string{"layers"},
		"additionalProperties": false,
	}
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", layerSchema, "")
	if err != nil {
		return nil, err
	}
	fmt.Println(rawResponse)
	// Parse the response which should be a map with "layers" key
	responseMap, ok := (*rawResponse).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", rawResponse)
	}

	layersArray, ok := responseMap["layers"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected layers type: %T", responseMap["layers"])
	}

	var response []string
	for _, item := range layersArray {
		if str, ok := item.(string); ok {
			response = append(response, str)
		}
	}
	var extractedLayers []tools.AvailableLayersData
	for _, layerName := range response {
		for _, layer := range all_layersSlice {
			if layer.Name == layerName {
				extractedLayers = append(extractedLayers, layer)
				break
			}
		}
	}
	return extractedLayers, nil
}

func handleGetInformationAboutALayer(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) (any, error) {
	relevantLayers, err := extractGetRelevantLayers(userQuery, previousConversations, toolRegistry)
	relevantLayersNames := []string{}
	for _, layer := range relevantLayers {
		relevantLayersNames = append(relevantLayersNames, layer.Name)
	}
	if err != nil {
		return nil, err
	}
	layersInformation, err := toolRegistry.Execute(context.Background(), "get_layer_info", map[string]any{
		"layers": strings.Join(relevantLayersNames, ","),
	}, nil)
	instructions := AiManager.Instructions + fmt.Sprintf(`Using the information provided about the layers: %v, generate a comprehensive response to address the user's query.`, layersInformation)
	if len(relevantLayers) > 1 {
		instructions += fmt.Sprintf(`Inform users that these layers were also extracted based on query %v, but only %s was used`, relevantLayers[1:], relevantLayers[0])
	}
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil
}

func handleGetUpdatesAboutDataLayers(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry, platform string) (any, error) {
	relevantLayers, err := extractGetRelevantLayers(userQuery, previousConversations, toolRegistry)
	relevantLayersNames := []string{}
	for _, layer := range relevantLayers {
		relevantLayersNames = append(relevantLayersNames, layer.Name)
	}
	if err != nil {
		return nil, err
	}
	layerUpdatesInformation, err := toolRegistry.Execute(context.Background(), "get_layer_update_info", map[string]any{
		"layer":    strings.Join(relevantLayersNames, ","),
		"platform": platform,
	},nil)
	instructions := AiManager.Instructions + fmt.Sprintf(`Using the information provided about the layers: %v, generate a comprehensive response to address the user's query.`, layerUpdatesInformation)
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil
}

func handleAvailableLayersInPlatform(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) (any, error) {
	// Implementation for handling "available_layers_in_platform" intent
	all_layers, err := toolRegistry.Execute(context.Background(), "get_all_available_layers", map[string]any{}, nil)
	if err != nil {
		return nil, err
	}
	instructions := AiManager.Instructions + fmt.Sprintf(`Using the list of layers having their name and descriptions provided, generate a comprehensive response to address the user's query. Here is the list of layers and their descriptions: %v If list is too long, summarize appropriately.`, all_layers)
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil
}

func handleFindLayerInPlatform(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) (any, error) {
	// Implementation for handling "find_layer_in_platform" intent
	relevantLayers, err := extractGetRelevantLayers(userQuery, previousConversations, toolRegistry)
	if err != nil {
		return nil, err
	}
	// extracted layers
	instructions := AiManager.Instructions + fmt.Sprintf(`The tools useful for this will be iso-selector and layer list. 
									the extracted layer's layer_type is nationwide, then it will be available after selecting nationwide which is usually the last option available.
									If layer_type is an ISO like ERCOT, MISO etc. the ISO should be selected first using iso-selector tool,
									The platform has a Layer list which is located in right section of top navigation bar. This is first option available right of ISO Selector. 
							layer_names		It contains all the layers available in the platform after selecting an ISO or nationwide. Using this information user can locate the layer 
									in the platform. Using the list of layers having their name and layer_type provided, generate a comprehensive response to address the user's query.
									Here is the list of relevant layers:%v`, relevantLayers)
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil

}

func handleHowToUseATool(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) (any, error) {
	// Implementation for handling "how_to_use_a_tool" intent
	extractedChunks, err := toolRegistry.Execute(context.Background(), "get_user_guide_info", map[string]any{
		"query": userQuery,
	}, nil)
	if err != nil {
		return nil, err
	}
	instructions := AiManager.Instructions + fmt.Sprintf(`Using the information provided about how to use tools: %v, generate a comprehensive response to address the user's query.`, extractedChunks)
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil
}

func handleGeneralPlatformQuery(userQuery string, previousConversations string, toolRegistry *tools.ToolRegistry) (any, error) {
	// Implementation for handling "general_platform_query" intent
	extractedChunks, err := toolRegistry.Execute(context.Background(), "get_user_guide_info", map[string]any{
		"query": userQuery,
	}, nil)
	if err != nil {
		return nil, err
	}
	instructions := fmt.Sprintf(`Using the extracted tools %v, generate a comprehensive response to address the user's query.`, extractedChunks)
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", finalResponseSchema, "")
	return rawResponse, nil
}
