package tools

import (
	"ai-service-go/internals/config"
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type LayerInformation struct {
	Name             string              `json:"name"`
	LayerInformation string              `json:"layer_information"`
	Properties       []map[string]string `json:"properties"`
}

type GetLayerInformation struct{}

func (gl *GetLayerInformation) Name() string {
	return "get_layer_info"
}
func (gl *GetLayerInformation) Description() string {
	return `Fetch detailed information about a specific data layer, including its brief description and key 
			properties available for it on the platform. 
			Use tool get_all_available_layers to get the list of available layers, 
			and match the user input to the correct layer name before calling this tool.
			You can ask for multiple layers at a time by seperating them with comma, instead of calling this tool multiple times.`
}

func (gl *GetLayerInformation) Execute(ctx context.Context, params map[string]any) (any, error) {
	layerNames := strings.Split(params["layers"].(string), ",")
	result := map[string]LayerInformation{}
	var err error
	for _, layerName := range layerNames {
		layerName = formatLayerName(layerName)
		url := fmt.Sprintf("%s?name=%s", config.AppConfig.LayerInfoURL, layerName)
		x, err := makeGetRequest[struct{ Data LayerInformation }](ctx, url)
		if err != nil {
			continue
		}
		result[layerName] = x.Data
	}
	if len(result) == 0 && err != nil {
		return nil, err
	}
	return result, nil
}

func (gl *GetLayerInformation) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"layers": map[string]any{
					"type":        "string",
					"description": "The name of the data layer/layers for which information is to be fetched. If multiple layers, separate by comma(,).",
				},
			},
			"required": []string{"layers"},
		},
	}
}

func formatLayerName(layerName string) string {
	layerName = strings.ReplaceAll(layerName, " ", "_")
	layerName = strings.ToLower(layerName)
	return layerName
}
