package tools

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/types"
	"context"
	"encoding/json"
	"fmt"
	"maps"
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
	return `Get detailed information about a specific data layer including available properties/fields.
			This tool can open a layer information modal for the user if it would be helpful. `
}

func (gl *GetLayerInformation) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	layerNames := strings.Split(params["layers"].(string), ",")
	paramsToSend := make(map[string]any)
	maps.Copy(paramsToSend, params)
	paramsToSend["layers"] = []string{}
	// Check if opening modal is enough
	isOpeningModalEnough := false
	if val, ok := params["isOpeningModalEnough"].(bool); ok {
		isOpeningModalEnough = val
	}
	// Check if opening modal is helpful
	isOpeningModalHelpful := false
	if val, ok := params["isOpeningModalHelpful"].(bool); ok {
		isOpeningModalHelpful = val
	}

	result := map[string]LayerInformation{}
	var err error
	var x *struct{ Data LayerInformation }

	dataManager := GetDataManager()
	availableLayers, err := dataManager.GetAvailableLayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available layers: %w", err)
	}

	for _, layerName := range layerNames {
		for _, availableLayer := range availableLayers {
			if replaceSpaceAndMakeSmallCase(availableLayer.Name) == replaceSpaceAndMakeSmallCase(strings.TrimSpace(layerName)) {
				paramsToSend["layers"] = append(paramsToSend["layers"].([]string), replaceSpaceAndMakeSmallCase(availableLayer.Name))
				break
			}
		}
	}
	if isOpeningModalHelpful && len(paramsToSend["layers"].([]string)) > 0 {
		sendMessage(types.StreamMessage{
			Type: "tool_request",
			Data: func() json.RawMessage {
				data, _ := json.Marshal(struct {
					ToolName string         `json:"tool_name"`
					Params   map[string]any `json:"params"`
				}{
					ToolName: gl.Name(),
					Params:   paramsToSend,
				})
				return data
			}(),
		})
		if isOpeningModalEnough {
			return map[string]any{
				"status":  "modal_opened",
				"message": fmt.Sprintf("Layer information modal has been opened for: %s. The user can view the details in the modal.", strings.Join(layerNames, ", ")),
			}, nil
		}
	}

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
					"description": "The name of the data layer for which information is to be fetched.",
				},
				"isOpeningModalHelpful": map[string]any{
					"type":        "boolean",
					"description": "Set to true if opening the layer information modal would be helpful for the user to visualize the layer details.",
				},
				"isOpeningModalEnough": map[string]any{
					"type": "boolean",
					"description": `Set to true if opening the modal is sufficient and you don't need the raw layer data. 
									Set to false if you need the raw layer data to process, compare, or include in your response.`,
				},
			},
			"required": []string{"layers", "isOpeningModalHelpful", "isOpeningModalEnough"},
		},
	}
}

func (gl *GetLayerInformation) InformationMessage() struct {
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
