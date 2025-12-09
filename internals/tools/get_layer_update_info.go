package tools

import (
	"ai-service-go/internals/types"
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type GetUpdateInformation struct{}

func (gu *GetUpdateInformation) Name() string {
	return "get_layer_update_info"
}

func (gu *GetUpdateInformation) Description() string {
	return `Get update cycle frequency (annually/quarterly) and data availability timeline 
			for specific data layers on trial or standard platforms. 
			Supports multiple comma-separated layers.`
}

func (gu *GetUpdateInformation) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	layers := strings.Split(params["layer"].(string), ",")
	for i := range layers {
		layers[i] = strings.TrimSpace(layers[i])
		layers[i] = replaceSpaceAndMakeSmallCase(layers[i])
	}
	platform := params["platform"].(string)
	if len(layers) == 0 || platform == "" {
		return "Update cycle : Not available, Available data till : Not available ", nil
	}
	matchedLayers := []string{}
	for _, layer := range layers {
		matches := MatchLayerName(layer, LayerUpdateCycleDictionary, 0.6)
		if len(matches) > 0 {
			matchedLayers = append(matchedLayers, matches...)
		}
	}
	if len(matchedLayers) == 0 {
		return "Update cycle : Not available, Available data till : Not available. Layer might not be updated frequently.", nil

	}
	currentYear := time.Now().Year()
	currentMonth := time.Now().Month()
	currentQuarter := (int(currentMonth)-1)/3 + 1
	matchedLayer := matchedLayers[0]
	updateCycle := LayerUpdateCycleDictionary[matchedLayer][platform]
	dataAvailable := DataAvailableDictionary[matchedLayer][platform]

	// Replace placeholders
	dataAvailable = strings.ReplaceAll(dataAvailable, "current_year", fmt.Sprintf("%d", currentYear))
	dataAvailable = strings.ReplaceAll(dataAvailable, "last_year", fmt.Sprintf("%d", currentYear-1))
	dataAvailable = strings.ReplaceAll(dataAvailable, "current_quarter", fmt.Sprintf("Q%d", currentQuarter))
	return fmt.Sprintf("Matched Layer name %s ,Update cycle : %s, Available data till : %s", matchedLayer, updateCycle, dataAvailable), nil
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

func (gl *GetUpdateInformation) InformationMessage() struct{
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

var LayerUpdateCycleDictionary = map[string]map[string]string{
	"Substation":                               {"trial": "annually", "standard": "quarterly"},
	"Operational Resource":                     {"trial": "annually", "standard": "quarterly"},
	"Planned Resource":                         {"trial": "annually", "standard": "quarterly"},
	"Injection Capacity Contour":               {"trial": "annually", "standard": "quarterly"},
	"Resource Node LMP Basis Analysis":         {"trial": "annually", "standard": "annually"},
	"Resource Node LMP Basis Analysis Contour": {"trial": "annually", "standard": "annually"},
	"Ancillary Service":                        {"trial": "annually", "standard": "annually"},
	"Binding Constraint":                       {"trial": "annually", "standard": "quarterly"},
	"Planned Transmission Upgrade":             {"trial": "annually", "standard": "quarterly"},
	"Load Forecast Contour":                    {"trial": "annually", "standard": "quarterly"},
	"Large Load Data":                          {"trial": "annually", "standard": "annually"},
	"Daily Average LMP Chart":                  {"trial": "annually", "standard": "annually"},
	"Transmission line":                        {"trial": "quarterly", "standard": "annually"},
}

var DataAvailableDictionary = map[string]map[string]string{
	"Substation":                               {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Operational Resource":                     {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Planned Resource":                         {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Injection Capacity Contour":               {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Resource Node LMP Basis Analysis":         {"trial": "for years 2021-2023", "standard": "for years 2022-2024"},
	"Resource Node LMP Basis Analysis Contour": {"trial": "for years 2021-2023", "standard": "for years 2022-2024"},
	"Ancillary Service":                        {"trial": "for year 2022-2024", "standard": "for years 2022-2024"},
	"Binding Constraint":                       {"trial": "up to last_year", "standard": "for current_quarter - current_year"},
	"Planned Transmission Upgrade":             {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Load Forecast Contour":                    {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Large Load Data":                          {"trial": "for year 2024-2028", "standard": "for year 2025-2031"},
	"Daily Average LMP Chart":                  {"trial": "for year 2022-2024", "standard": "for year 2022-2024"},
	"Transmission line":                        {"trial": "up to Q4 last_year", "standard": "upto current_year"},
}

// MatchLayerName matches user input to layer names with a threshold
func MatchLayerName(userInput string, layerDict map[string]map[string]string, threshold float64) []string {
	matched := []string{}
	for layer := range layerDict {
		x := strings.ReplaceAll(layer, " ", "_")
		x = strings.ToLower(x)
		if strings.Contains(userInput, x) {
			matched = append(matched, layer)
			continue
		}
	}
	return matched
}
