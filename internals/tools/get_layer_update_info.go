package tools

import (
	"ai-service-go/internals/types"
	"context"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type GetUpdateInformation struct{}

func (gu *GetUpdateInformation) Name() string {
	return "get_layer_update_info"
}

func (gu *GetUpdateInformation) Description() string {
	return `Get update cycle frequency (annually/quarterly) and data availability timeline 
			for specific data layers. 
			Supports multiple comma-separated layers.`
}

func (gu *GetUpdateInformation) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	rawLayerNames, ok := params["layer"].(string)
	if !ok {
		return "Invalid layer parameter", nil
	}
	layers := strings.Split(rawLayerNames, ",")
	forInternalUse := params["for_internal_use"]

	for i := range layers {
		layers[i] = strings.TrimSpace(layers[i])
		layers[i] = replaceSpaceAndMakeSmallCase(layers[i])
	}

	dataManager := GetDataManager()

	if forInternalUse != nil && forInternalUse.(bool) {
		// if internal user, return all layers update info
		allInfo, err := dataManager.GetAllLayersUpdateInfo()
		if err != nil {
			log.Printf("Error loading layer data: %v", err)
			return "", err
		}
		return allInfo, nil
	}

	platform := params["platform"].(string)
	if len(layers) == 0 || platform == "" {
		return "Update cycle : Not available, Available data till : Not available ", nil
	}

	layerToCheck := layers[0]
	updateCycle, dataAvailable, err := dataManager.GetLayerUpdateInfo(layerToCheck, platform)
	if err != nil {
		log.Printf("Error getting layer update info: %v", err)
		return fmt.Sprintf("Matched Layer name %s, Update cycle : Not available, Available data till : Not available", layers[0]), nil
	}

	return fmt.Sprintf("Matched Layer name %s ,Update cycle : %s, Available data till : %s", layers[0], updateCycle, dataAvailable), nil
}

func (gu *GetUpdateInformation) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gu.Name(),
		Description: gu.Description(),
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"layer": map[string]any{
					"type":        "string",
					"description": "The name of the data layers for which update information is to be fetched. If there are multiple layers, separate them by comma.",
				},
				"platform": map[string]any{
					"type":        "string",
					"description": "The platform type, either 'trial' or 'standard'.",
				},
			},
			"required": []string{"layer", "platform"},
		},
	}
}

func (gl *GetUpdateInformation) InformationMessage() struct {
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Fetching layer update information...`,
		End:   `Layer update information fetched.`,
	}
}
