package tools

import (
	"context"

	"ai-service-go/internals/config"
	openai "github.com/sashabaranov/go-openai"
)

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
	return `Fetch a list of all available data layers which helps in matching user queries to relevant layers on the platform. 
	This is useful for getting layer input parameters, as they are name sensitive and need to be matched correctly.`
}

func (gl *GetAllAvailableLayers) Execute(ctx context.Context, params map[string]any) (any, error) {

	type response struct {
		Data []AvailableLayersData `json:"data"`
	}
	url := config.AppConfig.AllLayersURL
	var result *response
	result, err := makeGetRequest[response](ctx, url)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (gl *GetAllAvailableLayers) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters:  nil,
	}
}
