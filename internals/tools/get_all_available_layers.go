package tools

import (
	"ai-service-go/internals/types"
	"context"

	openai "github.com/sashabaranov/go-openai"
)

var InternalLayers = []string{"hunt_power_(hp-2)"}

type GetAllAvailableLayers struct{}
type AvailableLayersData struct {
	Name             string `json:"name"`
	LayerInformation string `json:"layer_information"`
	LayerType        string `json:"layer_group"`
}

func (gl *GetAllAvailableLayers) Name() string {
	return "get_all_available_layers"
}
func (gl *GetAllAvailableLayers) Description() string {
	return `Returns complete list of all available data layers with their names, descriptions, and types.`
}

func (gl *GetAllAvailableLayers) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	onlyNames := false
	if val, ok := params["onlyNames"].(bool); ok {
		onlyNames = val
	}

	dataManager := GetDataManager()
	layers, err := dataManager.GetAvailableLayers(ctx)
	if err != nil {
		return nil, err
	}

	if onlyNames {
		names := make([]string, len(layers))
		for i, layer := range layers {
			names[i] = makeLayerLikeTitle(layer.Name)
		}
		return names, nil
	}

	formattedLayers := make([]AvailableLayersData, len(layers))
	for i, layer := range layers {
		formattedLayers[i] = layer
		formattedLayers[i].Name = makeLayerLikeTitle(layer.Name)
	}
	return formattedLayers, nil
}

func (gl *GetAllAvailableLayers) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"onlyNames": map[string]any{
					"type":        "boolean",
					"description": "If true, returns only the layer names without descriptions. Defaults to false.",
				},
			},
			"required": []string{},
		},
	}
}

func (gl *GetAllAvailableLayers) InformationMessage() struct {
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Fetching all available layers...`,
		End:   `All layers fetched.`,
	}
}
