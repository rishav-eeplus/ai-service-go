package orchestrator

import (
	"ai-service-go/internals/tools"
	"encoding/json"
	"fmt"
	"strings"
)

func (o *Orchestrator) Router(sendMessage func(msg StreamMessage) bool, input *ClientRequestType) ([]tools.Tool, error) {
	sendMessage(StreamMessage{
		Type:    "info",
		Message: "Identifying useful_tools...",
	})
	intentSchema := map[string]interface{}{
		"type":        "object",
		"description": "Response containing list of useful_tools from the predefined list of their name with name and descriptions that match the user's needs.",
		"properties": map[string]interface{}{
			"useful_tools": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Array of useful_tools that may help user for their query",
			},
		},
		"required":             []string{"useful_tools"},
		"additionalProperties": false,
	}
	intentsSummary := "Predefined Intents:\n"
	for i, tool := range o.ToolResistory.AllTools() {
		intentsSummary += fmt.Sprintf("%d. Name : %s, Description: %s\n", i+1, tool.Name(), tool.Description())
	}
	instructions := fmt.Sprintf(`Determine the most relevant tools from the predefined list %s  based on the user's query and previous conversations.  
							Provide the identified tools as an array of tool names.`, intentsSummary)
	var err error
	rawResponse, _, err := o.AIManager.GetAIResponse(instructions, input.UserQuery, input.PreviousConversation, "", intentSchema, input.Platform)
	if err != nil {
		return nil, err
	}

	// Parse the response which should be a map with "useful_tools" key
	responseMap, ok := (*rawResponse).(map[string]interface{})
	if !ok {
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("unexpected response type: %T", rawResponse),
		})
		return nil, fmt.Errorf("unexpected response type: %T", rawResponse)
	}

	usefulToolsArray, ok := responseMap["useful_tools"].([]interface{})
	if !ok {
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("unexpected useful_tools type: %T", responseMap["useful_tools"]),
		})
		return nil, fmt.Errorf("unexpected useful_tools type: %T", responseMap["useful_tools"])
	}

	var response []string
	for _, item := range usefulToolsArray {
		if str, ok := item.(string); ok {
			response = append(response, str)
		}
	}

	var matchedTools []tools.Tool
	for _, respTool := range response {
		for _, tool := range usefulToolsArray {
			if tool == respTool {
				matchedTools = append(matchedTools, o.ToolResistory.GetTool(respTool))
				break
			}
		}
	}
	matchedToolsJSON, err := json.Marshal(matchedTools)
	if err != nil {
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Errorf("failed to marshal matched useful_tools: %v", err).Error(),
		})
		return nil, fmt.Errorf("failed to marshal matched useful_tools: %v", err)
	}
	sendMessage(StreamMessage{
		Type:    "info",
		Data:    matchedToolsJSON,
		Message: fmt.Sprintf("Identified tool(s) %s", strings.Join(response, ", ")),
	})
	return matchedTools, nil
}
