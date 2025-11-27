package tools

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type LocateALayer struct{}

func (gl *LocateALayer) Name() string {
	return "locate_a_layer"
}
func (gl *LocateALayer) Description() string {
	return `Helps users to find/locate a layer in the data platform.`
}

func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any) (any, error) {
    relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
    if !ok {
        fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
        return "Please provide valid layer information.", nil
    }

    // Extract fields from the map
    name, _ := relevantLayerMap["name"].(string)
    layerType, _ := relevantLayerMap["layer_type"].(string)
    instructions := "The platform has a ISO selector present in top middle section of the screen (in the middle of navigation bar). Use that to select the relevant ISO for the layer.\n"
    instructions += "Just to the right of ISO selector, there is layer list, click that to get a popup which lists all layers available for the selected ISO, and provides a button to toggle one.\n"
    switch layerType {
    case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
        instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
    case "nationwide":
        instructions += fmt.Sprintf("Since the layer %s is a nationwide layer, select nationwide in the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
    case "iso-based":
        instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
    default:
        instructions += fmt.Sprintf("Using type %s of the layer %s, select the appropriate ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", layerType, name, name)
    }
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
			},
			"required": []string{"relevant_layer"},
		},
	}
}
