package tools

import (
	"ai-service-go/internals/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var AllLayerPaths = map[string]string{}

type dataType struct {
	Title    string     `json:"title"`
	Children []dataType `json:"children"`
}
type LocateALayer struct{}

func (gl *LocateALayer) Name() string {
	return "locate_a_layer"
}
func (gl *LocateALayer) Description() string {
	return `Provides step-by-step UI navigation instructions to get or find and enable a specific layer on the platform. Generates instructions based on layer type.`
}

// func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any) (any, error) {
// 	relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
// 	if !ok {
// 		fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
// 		return "Please provide valid layer information.", nil
// 	}

// 	// Extract fields from the map
// 	name, _ := relevantLayerMap["name"].(string)
// 	layerType, _ := relevantLayerMap["layer_type"].(string)
// 	instructions := "The platform has a ISO selector present in top middle section of the screen (in the middle of navigation bar). Use that to select the relevant ISO for the layer.\n"
// 	instructions += "Just to the right of ISO selector, there is layer list, click that to get a popup which lists all layers available for the selected ISO, and provides a button to toggle one.\n"
// 	switch layerType {
// 	case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
// 		instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	case "nationwide":
// 		instructions += fmt.Sprintf("Since the layer %s is a nationwide layer, select nationwide in the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	case "iso-based":
// 		instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	default:
// 		instructions += fmt.Sprintf("Using type %s of the layer %s, select the appropriate ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", layerType, name, name)
// 	}
// 	return instructions, nil
// }

func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
	if !ok {
		fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
		return "Please provide valid layer information.", nil
	}
	openLayerInUI, _ := params["open_layer_in_ui"].(bool)
	fmt.Printf("Open layer in Map: %t", openLayerInUI)

	// Extract fields from the map
	name, _ := (relevantLayerMap["name"].(string))
	name = replaceSpaceAndMakeSmallCase(name)
	layerType, _ := relevantLayerMap["layer_type"].(string)
	allLayerPaths := GetLayerPath()
	layerPath := allLayerPaths[name]
	if openLayerInUI {
		sendMessage(types.StreamMessage{
			Type: "tool_request",
			Data: func() json.RawMessage {
				data, _ := json.Marshal(struct {
					ToolName string         `json:"tool_name"`
					Params   map[string]any `json:"params"`
				}{
					ToolName: gl.Name(),
					Params: map[string]any{
						"relevant_layer": name,
						"path":           layerPath,
					},
				})
				return data
			}(),
		})
	}

	var instructions string
	if openLayerInUI {
		instructions = "The layer has been attempted to open in the UI. If it is not visible, follow these steps to find and enable it manually:\n\n"
	} else {
		instructions = "To find and enable the layer on the platform, follow these steps:\n\n"
	}
	// Step 1: ISO selection
	instructions += "1. **Select the ISO/Region**: The platform has an ISO selector in the top middle section of the screen (in the navigation bar). "
	switch layerType {
	case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
		instructions += fmt.Sprintf("Select '%s' from the ISO selector.\n\n", strings.ToUpper(layerType))
	case "nationwide":
		instructions += "Select 'Nationwide' from the ISO selector.\n\n"
	case "iso-based":
		instructions += "Select the relevant ISO for your region from the ISO selector.\n\n"
	default:
		instructions += "Select the appropriate ISO/region from the ISO selector.\n\n"
	}
	// Step 2: Open layer panel
	instructions += "2. **Open the Layer Panel**: Click on the layer list button (just to the right of the ISO selector) to open a popup that displays all available layers organized in groups.\n\n"
	// Step 3: Navigate through the path
	if layerPath != "" {
		instructions += fmt.Sprintf("3. **Navigate to the Layer**: The layers are organized in groups. Follow this path to find '%s':\n", name)
		instructions += fmt.Sprintf("   **%s**\n\n", layerPath)
		instructions += "   Click through each group/subgroup in order to expand it and reveal the layer.\n\n"
	} else {
		instructions += fmt.Sprintf("3. **Find the Layer**: Search or browse through the grouped layers to find '%s'.\n\n", name)
	}
	// Step 4: Toggle the layer
	instructions += "4. **Enable the Layer**: Once you locate the layer, click the toggle button next to it to enable the layer on the map."

	return instructions, nil

}

func (gl *LocateALayer) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"relevant_layer": map[string]any{
					"type":        "object",
					"description": "The relevant layer information including name, layer information and layer type.",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Name of the layer.",
						},
						"layer_information": map[string]any{
							"type":        "string",
							"description": "Information about the layer.",
						},
						"layer_type": map[string]any{
							"type":        "string",
							"description": "Type of the layer such as iso-based, nationwide, ercot, pjm etc.",
						},
					},
					"required": []string{"name", "layer_information", "layer_type"},
				},
				"open_layer_in_ui": map[string]any{
					"type":        "boolean",
					"description": "If true, attempts to open the layer directly in the UI. If the layer is not visible after attempting, manual instructions will be provided as fallback.",
				},
			},
			"required": []string{"relevant_layer"},
		},
	}
}
func (gl *LocateALayer) InformationMessage() struct {
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Locating the layer on the platform...`,
		End:   `Layer location instructions generated.`,
	}
}

// Helper functions
func GetLayerPath() map[string]string {
	if len(AllLayerPaths) > 0 {
		return AllLayerPaths
	}
	paths := getAllPaths()
	AllLayerPaths = paths
	return AllLayerPaths
}
