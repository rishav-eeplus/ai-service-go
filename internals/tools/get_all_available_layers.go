package tools

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

var AllAvailableLayers []AvailableLayersData
var InternalLayers = []string{"hunt_power_(hp-2)"}

type GetAllAvailableLayers struct{}
type AvailableLayersData struct {
	Name             string `json:"name"`
	LayerInformation string `json:"layer_information"`
	LayerType        string `json:"layer_type"`
}

func (gl *GetAllAvailableLayers) Name() string {
	return "get_all_available_layers"
}
func (gl *GetAllAvailableLayers) Description() string {
	return `Returns complete list of all available data layers with their names, descriptions, and types.`
}

func (gl *GetAllAvailableLayers) Execute(ctx context.Context, params map[string]any) (any, error) {
	if len(AllAvailableLayers) > 0 {
		return AllAvailableLayers, nil
	}
	layers , err := getAllAvailableLayers(ctx)
	if err != nil {
		return nil, err
	}
	paths := getAllPaths()
	AllLayerPaths = paths
	filteredLayers := []AvailableLayersData{}
	for _, layer := range layers {
		if _, exists := paths[replaceSpaceAndMakeSmallCase(layer.Name)]; exists {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	AllAvailableLayers = filteredLayers
	return AllAvailableLayers, nil
}

func (gl *GetAllAvailableLayers) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters:  nil,
	}
}
