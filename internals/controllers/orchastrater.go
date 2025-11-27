package controllers

import (
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"context"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

type Intent struct {
	Name        string
	Description string
}

var whatIsALayer = `A layer is a distinct, toggleable data visualization overlay on the EEHorizon map interface that represents a specific category of power infrastructure, grid analysis, or geographic information.
			Each layer displays specific types of information such as substations, transmission lines, operational resources, or analytical data like injection capacity contours or LMP basis analysis.
			Layers can be either:
				ISO-specific: Related to a particular Independent System Operator (e.g., ERCOT Substations, PJM Binding Constraints)
				Nationwide: Available across all regions (e.g., Transmission Lines, Land Parcels, Coal Mines)
			Interactive Elements: Layers contain:
				Geographic elements that can be clicked for detailed information
				Visual representations using colors, symbols, and legends
				Data points that can be filtered, searched, and exported
				Composable: Multiple layers can be overlaid simultaneously on the map to provide a holistic view of power infrastructure and grid dynamics.
			Conditional Visibility: Some layers (especially contour/heat maps) require filters to be applied before they appear on the map.
			Examples of Layers:
				Substations: Shows power infrastructure locations with injection capacity data
				Operational & Planned Resources: Displays generation facilities by fuel type
				Injection Capacity Contour: Heat map showing available transfer capacity
				SSR Contour: Sub-synchronous resonance risk visualization
				Binding Constraints: Top 50 transmission constraints by shadow price
				Ancillary Services Markets: Market pricing and service data
				Planned Transmission Upgrades: Future grid expansion projects
				Land Parcels: Property boundary information (Nationwide)
				Wind Turbines & Buffer: Renewable energy infrastructure locations
				Each layer represents a unique data dimension that helps users make informed decisions about energy infrastructure, grid operations, and investment planning.`

var intents = []Intent{
	{
		Name:        "get_information_about_a_layer",
		Description: whatIsALayer + ` User wants to get information about a layer present in the platform, it can be used to understand the layer's purpose, data it contains, and how it can be utilized.`,
	}, {
		Name:        "get_updates_about_data_layers",
		Description: whatIsALayer + `User is inquiring about update cycle of data layers, including how frequently data is refreshed and what years/timeperiod data is available for a layer.`,
	}, {
		Name:        "find_layer_in_platform",
		Description: whatIsALayer + `User is looking to identify or locate a specific data layer within the platform, possibly to understand its relevance or to access its data.`,
	}, {
		Name: "available_layers_in_platform",
		Description: whatIsALayer + ` User wants to know the list of available data layers in the platform, which can help in understanding the breadth of data coverage and selecting relevant layers for analysis or queries.
						Also user wants to know about helpful layers based on their use-cases.`,
	}, {
		Name:        "how_to_use_a_tool",
		Description: `User is seeking guidance on whether is it possible to do something in the platform, and if so, how to effectively utilize the available tools or features to achieve their goals.`,
	}, {
		Name:        "general_platform_query",
		Description: `This could include questions about platform features, capabilities, or other topics.`,
	},
}

func GetIntentOfUserQuery(userQuery string, previousConversations string) ([]Intent, error) {
	intentSchema := map[string]interface{}{
		"type":        "object",
		"description": "Response containing list of identified intents from the predefined list that match the user's needs.",
		"properties": map[string]interface{}{
			"intents": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Array of intent names that match the user query",
			},
		},
		"required":             []string{"intents"},
		"additionalProperties": false,
	}
	intentsSummary := "Predefined Intents:\n"
	for i, intent := range intents {
		intentsSummary += fmt.Sprintf("%d. Name : %s, Description: %s\n", i+1, intent.Name, intent.Description)
	}
	instructions := fmt.Sprintf(`Determine the most relevant intents from the predefined list %s  based on the user's query and previous conversations.  
							Provide the identified intents as an array of intent names.`, intentsSummary)
	var err error
	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversations, "", intentSchema, "")
	if err != nil {
		return nil, err
	}

	// Parse the response which should be a map with "intents" key
	responseMap, ok := (*rawResponse).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", rawResponse)
	}

	intentsArray, ok := responseMap["intents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected intents type: %T", responseMap["intents"])
	}

	var response []string
	for _, item := range intentsArray {
		if str, ok := item.(string); ok {
			response = append(response, str)
		}
	}

	var matchedIntents []Intent
	for _, respIntent := range response {
		for _, intent := range intents {
			if intent.Name == respIntent {
				matchedIntents = append(matchedIntents, intent)
				break
			}
		}
	}
	return matchedIntents, nil
}

// checkRelevantVectors examines vector metadata to determine if any additional context is needed
func checkRelevantVectors(userQuery string, previousConversation string, identifiedIntents []Intent, vectorManager interface {
	GetAllVectorsWithMetadata() ([]map[string]interface{}, error)
}) ([]uint64, error) {
	// Get all vectors with metadata
	vectors, err := vectorManager.GetAllVectorsWithMetadata()
	if err != nil {
		return nil, err
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
	for _, intent := range identifiedIntents {
		intentNames = append(intentNames, intent.Name)
	}

	instructions := fmt.Sprintf(`Based on the user's query and the identified intents %v, examine the following vector summaries and determine which vectors contain relevant information that would be helpful to answer the query.
						Return the IDs of relevant vectors. If no vectors are particularly relevant beyond what the intents already cover, return an empty array.
						Vector Summaries:
						%s`, intentNames, strings.Join(vectorSummaries, "\n"))

	rawResponse, _, err := AiManager.GetAIResponse(instructions, userQuery, previousConversation, "", vectorSelectionSchema, "")
	if err != nil {
		return nil, err
	}

	// Parse response
	responseMap, ok := (*rawResponse).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", rawResponse)
	}

	idsArray, ok := responseMap["relevant_vector_ids"].([]interface{})
	if !ok {
		return []uint64{}, nil // Return empty if parsing fails
	}

	var relevantVectorIDs []uint64
	for _, item := range idsArray {
		// Handle both float64 (JSON number) and int
		switch v := item.(type) {
		case float64:
			relevantVectorIDs = append(relevantVectorIDs, uint64(v))
		case int:
			relevantVectorIDs = append(relevantVectorIDs, uint64(v))
		case int64:
			relevantVectorIDs = append(relevantVectorIDs, uint64(v))
		}
	}

	return relevantVectorIDs, nil
}

// enhanceResponseWithVectors fetches specific vectors and enhances the response
func enhanceResponseWithVectors(originalResponse any, vectorIDs []uint64, userQuery string, previousConversation string, toolRegistry *tools.ToolRegistry) (any, error) {
	// Fetch the specific vectors by their IDs
	if userGuideInfoTool := toolRegistry.GetTool("get_user_guide_info"); userGuideInfoTool != nil {
		if guideTool, ok := userGuideInfoTool.(*tools.GetUserGuideInformation); ok {
			// Get vector content by searching (as a workaround since we don't have direct ID fetch)
			vectorContext, err := guideTool.VectorManager.SearchSimilarChunks(userQuery, len(vectorIDs), 0)
			if err != nil {
				return originalResponse, err
			}

			// Enhance the original response with vector context
			responseMap, ok := originalResponse.(map[string]interface{})
			if !ok {
				return originalResponse, fmt.Errorf("unexpected response type")
			}

			result, _ := responseMap["result"].(string)

			// Ask AI to enhance the response with additional context
			enhanceInstructions := fmt.Sprintf(`The original response was: %s
				
				Additional relevant context from user guide:
				%s
				
				Enhance the original response by incorporating relevant information from the additional context if it adds value. 
				Keep the same tone and structure. If the additional context doesn't add value, return the original response as-is.`, result, vectorContext)

			enhancedResponse, _, err := AiManager.GetAIResponse(enhanceInstructions, userQuery, previousConversation, "", finalResponseSchema, "")
			if err != nil {
				return originalResponse, err
			}

			return enhancedResponse, nil
		}
	}
	return originalResponse, nil
}

func HandleResponseBasedOnIntent(intent Intent, userQuery string, previousConversations string, platform string, toolRegistry *tools.ToolRegistry) (any, error) {
	if intent.Name == "" {
		return nil, fmt.Errorf("no intents identified")
	}
	switch intent.Name {
	case "get_information_about_a_layer":
		return handleGetInformationAboutALayer(userQuery, previousConversations, toolRegistry)
	case "get_updates_about_data_layers":
		return handleGetUpdatesAboutDataLayers(userQuery, previousConversations, toolRegistry, platform)
	case "available_layers_in_platform":
		return handleAvailableLayersInPlatform(userQuery, previousConversations, toolRegistry)
	case "find_layer_in_platform":
		return handleFindLayerInPlatform(userQuery, previousConversations, toolRegistry)
	case "how_to_use_a_tool":
		return handleHowToUseATool(userQuery, previousConversations, toolRegistry)
	case "general_platform_query":
		return handleGeneralPlatformQuery(userQuery, previousConversations, toolRegistry)
	default:
		return nil, fmt.Errorf("unknown intent: %s", intent.Name)
	}
}

func (aiM *OpenAIManager) GenerateResponseV2(userQuery, previousConversation, platform, model string, registory *tools.ToolRegistry, ctx context.Context) (any, utils.TokenUsage, error) {
	identifiedIntents, err := GetIntentOfUserQuery(userQuery, previousConversation)
	fmt.Println("Identified intents:", identifiedIntents)
	if err != nil {
		return nil, utils.TokenUsage{}, err
	}

	// Check if any vectors contain relevant information
	var relevantVectorIDs []uint64
	if userGuideInfoTool := registory.GetTool("get_user_guide_info"); userGuideInfoTool != nil {
		if guideTool, ok := userGuideInfoTool.(*tools.GetUserGuideInformation); ok {
			if vectorManager, ok := guideTool.VectorManager.(interface {
				GetAllVectorsWithMetadata() ([]map[string]interface{}, error)
			}); ok {
				relevantVectorIDs, err = checkRelevantVectors(userQuery, previousConversation, identifiedIntents, vectorManager)
				if err != nil {
					utils.Logger.Warnf("Failed to check relevant vectors: %v", err)
				} else {
					fmt.Println("Relevant vector IDs:", relevantVectorIDs)
				}
			}
		}
	}

	var finalResponse any
	var totalUsage utils.TokenUsage
	// for now only handle the first intent
	finalResponse, err = HandleResponseBasedOnIntent(identifiedIntents[0], userQuery, previousConversation, platform, registory)
	if err != nil {
		return nil, utils.TokenUsage{}, err
	}

	// If relevant vectors were found, fetch and include them in the response
	if len(relevantVectorIDs) > 0 {
		finalResponse, err = enhanceResponseWithVectors(finalResponse, relevantVectorIDs, userQuery, previousConversation, registory)
		if err != nil {
			utils.Logger.Warnf("Failed to enhance response with vectors: %v", err)
		}
	}

	aiResponse := finalResponse
	return &aiResponse, totalUsage, nil
}

// StreamMessage represents messages sent via WebSocket
type StreamMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}




// GenerateResponseV2Stream is a streaming version that sends updates via WebSocket
func (aiM *OpenAIManager) GenerateResponseV2Stream(conn *websocket.Conn, userQuery, previousConversation, platform, model string, registory *tools.ToolRegistry, ctx context.Context) {
	// Helper function to send messages with error handling
	sendMessage := func(msg StreamMessage) bool {
		if err := conn.WriteJSON(msg); err != nil {
			utils.Logger.Errorf("Error writing to WebSocket: %v", err)
			return false
		}
		return true
	}

	// Step 1: Identify intents
	if !sendMessage(StreamMessage{
		Type:    "status",
		Message: "Identifying query intent...",
	}) {
		return
	}

	identifiedIntents, err := GetIntentOfUserQuery(userQuery, previousConversation)
	if err != nil {
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Error identifying intents: %v", err),
		})
		return
	}

	if !sendMessage(StreamMessage{
		Type:    "intents",
		Data:    identifiedIntents,
		Message: fmt.Sprintf("Identified %d intent(s)", len(identifiedIntents)),
	}) {
		return
	}

	// Step 2: Check for relevant vectors
	if !sendMessage(StreamMessage{
		Type:    "status",
		Message: "Checking for relevant vectors...",
	}) {
		return
	}

	var relevantVectorIDs []uint64
	if userGuideInfoTool := registory.GetTool("get_user_guide_info"); userGuideInfoTool != nil {
		if guideTool, ok := userGuideInfoTool.(*tools.GetUserGuideInformation); ok {
			if vectorManager, ok := guideTool.VectorManager.(interface {
				GetAllVectorsWithMetadata() ([]map[string]interface{}, error)
			}); ok {
				relevantVectorIDs, err = checkRelevantVectors(userQuery, previousConversation, identifiedIntents, vectorManager)
				if err != nil {
					utils.Logger.Warnf("Failed to check relevant vectors: %v", err)
					sendMessage(StreamMessage{
						Type:    "warning",
						Message: "Could not check vectors, continuing without them",
					})
				} else {
					if !sendMessage(StreamMessage{
						Type:    "vectors",
						Data:    relevantVectorIDs,
						Message: fmt.Sprintf("Found %d relevant vector(s)", len(relevantVectorIDs)),
					}) {
						return
					}
				}
			}
		}
	}

	// Step 3: Generate response based on intent
	if !sendMessage(StreamMessage{
		Type:    "status",
		Message: "Generating response...",
	}) {
		return
	}

	var finalResponse any
	if len(identifiedIntents) > 0 {
		finalResponse, err = HandleResponseBasedOnIntent(identifiedIntents[0], userQuery, previousConversation, platform, registory)
		if err != nil {
			sendMessage(StreamMessage{
				Type:    "error",
				Message: fmt.Sprintf("Error generating response: %v", err),
			})
			return
		}
	} else {
		sendMessage(StreamMessage{
			Type:    "error",
			Message: "No intents identified",
		})
		return
	}

	// Step 4: Enhance with vectors if available
	if len(relevantVectorIDs) > 0 {
		if !sendMessage(StreamMessage{
			Type:    "status",
			Message: "Enhancing response with vector data...",
		}) {
			return
		}

		finalResponse, err = enhanceResponseWithVectors(finalResponse, relevantVectorIDs, userQuery, previousConversation, registory)
		if err != nil {
			utils.Logger.Warnf("Failed to enhance response with vectors: %v", err)
			sendMessage(StreamMessage{
				Type:    "warning",
				Message: "Could not enhance with vectors, using original response",
			})
		}
	}

	// Step 5: Send final response
	if !sendMessage(StreamMessage{
		Type:    "response",
		Data:    finalResponse,
		Message: "Response generated successfully",
	}) {
		return
	}

	// Send completion
	sendMessage(StreamMessage{
		Type:    "complete",
		Message: "Processing complete",
	})
}
