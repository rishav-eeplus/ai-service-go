package tools

import (
	"ai-service-go/internals/types"
	"context"
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
	return `Provides step-by-step UI navigation instructions to open/ toggle/ turn on/ get or find and enable a specific layer on the platform. Generates instructions based on layer type.`
}

func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
	if !ok {
		fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
		return "Please provide valid layer information.", nil
	}
	// openLayerInUI, _ := params["open_layer_in_ui"].(bool)
	// Extract fields from the map
	name, _ := (relevantLayerMap["name"].(string))
	name = replaceSpaceAndMakeSmallCase(name)
	layerGroup, _ := relevantLayerMap["layer_group"].(string)
	allLayerPaths := GetLayerPath()
	layerPath := allLayerPaths[name]
	layerGroup = replaceSpaceAndMakeSmallCase(layerGroup)
	instructions := "To find and enable the layer on the platform, follow these steps:\n\n"
	// Step 1: ISO selection
	instructions += "1. **Select the ISO/Region**: The platform has an ISO selector in the top middle section of the screen (in the navigation bar). "
	switch layerGroup {
	case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
		instructions += fmt.Sprintf("Select '%s' from the ISO selector.\n\n", strings.ToUpper(layerGroup))
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
					"description": "The relevant layer information including name, layer information and layer group.",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Name of the layer.",
						},
						"layer_information": map[string]any{
							"type":        "string",
							"description": "Information about the layer.",
						},
						"layer_group": map[string]any{
							"type":        "string",
							"description": "Type of the layer such as iso-based, nationwide, ercot, pjm etc.",
						},
					},
					"required": []string{"name", "layer_information", "layer_group"},
				},
				// "open_layer_in_ui": map[string]any{
				// 	"type":        "boolean",
				// 	"description": "If true, attempts to open/turn-on/toggle the layer directly for the user. If the layer is not visible after attempting, manual instructions will be provided as fallback.",
				// },
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
