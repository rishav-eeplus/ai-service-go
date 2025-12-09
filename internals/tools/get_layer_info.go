package tools

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/types"
	"context"
	"encoding/json"
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
	return `Get detailed information about specific data layers including description and available properties/fields.
	 		Supports multiple comma-separated layers in one call.`
}

func (gl *GetLayerInformation) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	layerNames := strings.Split(params["layers"].(string), ",")
	result := map[string]LayerInformation{}
	var err error
	var x *struct{ Data LayerInformation }
	sendMessage(types.StreamMessage{
		Type: "tool_request",
		Data: func() json.RawMessage {
			data, _ := json.Marshal(struct {
				ToolName string         `json:"tool_name"`
				Params   map[string]any `json:"params"`
			}{
				ToolName: gl.Name(),
				Params:   params,
			})
			return data
		}(),
	})
	for _, layerName := range layerNames {
		layerName = replaceSpaceAndMakeSmallCase(layerName)
		url := fmt.Sprintf("%s?name=%s", config.AppConfig.LayerInfoURL, layerName)
		x, err = makeGetRequest[struct{ Data LayerInformation }](ctx, url)
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

func (gl *GetLayerInformation) InformationMessage() struct{
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Getting detailed information about the layer...`,
		End:   `Detailed information about the layer retrieved.`,
	}
}