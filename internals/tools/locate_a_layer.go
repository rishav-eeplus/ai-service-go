package tools

import (
	"ai-service-go/internals/config"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type dataType struct {
	Title    string     `json:"title"`
	Children []dataType `json:"children"`
}
type LocateALayer struct{}

func (gl *LocateALayer) Name() string {
	return "locate_a_layer"
}
func (gl *LocateALayer) Description() string {
	return `Provides step-by-step UI navigation instructions to get or find and enable a specific layer on the platform. Generates instructions based on layer type.`
}

// func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any) (any, error) {
// 	relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
// 	if !ok {
// 		fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
// 		return "Please provide valid layer information.", nil
// 	}

// 	// Extract fields from the map
// 	name, _ := relevantLayerMap["name"].(string)
// 	layerType, _ := relevantLayerMap["layer_type"].(string)
// 	instructions := "The platform has a ISO selector present in top middle section of the screen (in the middle of navigation bar). Use that to select the relevant ISO for the layer.\n"
// 	instructions += "Just to the right of ISO selector, there is layer list, click that to get a popup which lists all layers available for the selected ISO, and provides a button to toggle one.\n"
// 	switch layerType {
// 	case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
// 		instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	case "nationwide":
// 		instructions += fmt.Sprintf("Since the layer %s is a nationwide layer, select nationwide in the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	case "iso-based":
// 		instructions += fmt.Sprintf("Since the layer %s is an ISO based layer, first select the relevant ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", name, name)
// 	default:
// 		instructions += fmt.Sprintf("Using type %s of the layer %s, select the appropriate ISO from the ISO selector, then open the layer list and toggle the layer %s from there.", layerType, name, name)
// 	}
// 	return instructions, nil
// }

func (gl *LocateALayer) Execute(ctx context.Context, params map[string]any) (any, error) {
	relevantLayerMap, ok := params["relevant_layer"].(map[string]any)
	if !ok {
		fmt.Printf("Type assertion failed for relevant_layer: %v\n", params["relevant_layer"])
		return "Please provide valid layer information.", nil
	}
	// Extract fields from the map
	name, _ := relevantLayerMap["name"].(string)
	layerType, _ := relevantLayerMap["layer_type"].(string)
	allLayerPaths := GetLayerPath()
	layerPath := allLayerPaths[name]
	instructions := "To find and enable the layer on the platform, follow these steps:\n\n"
	// Step 1: ISO selection
	instructions += "1. **Select the ISO/Region**: The platform has an ISO selector in the top middle section of the screen (in the navigation bar). "
	switch layerType {
	case "ercot", "pjm", "miso", "caiso", "nyiso", "iso-ne", "spp", "serc", "wecc":
		instructions += fmt.Sprintf("Select '%s' from the ISO selector.\n\n", strings.ToUpper(layerType))
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

// Helper functions
func GetLayerPath() map[string]string {
	// load the data.json file
	var datas []dataType
	file, err := os.Open("./data.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&datas)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}
	dictionary := map[string]string{}
	for i := range datas {
		recursivelyCheckPath(datas[i], "", 0, []string{}, &dictionary)
	}
	url := config.AppConfig.AllLayersURL
	type response struct {
		Data []AvailableLayersData `json:"data"`
	}
	xxx := map[string]string{}
	result, _ := makeGetRequest[response](context.Background(), url)
	for i := range len(result.Data) {
		x := result.Data[i]
		xxx[x.Name] = dictionary[x.Name]
	}
	return xxx
}
func replaceSpaceAndMakeSmallCase(str string) string {
	return strings.ToLower(strings.ReplaceAll(str, " ", "_"))
}
func recursivelyCheckPath(data dataType, layerName string, depth int, paths []string, dict *map[string]string) {
	paths = append(paths, data.Title)
	var extractedIso string
	if depth == 0 {
		(*dict)[replaceSpaceAndMakeSmallCase(data.Title)] = data.Title
		extractedIso = extractIso(data.Title)
		if extractedIso != "" {
			layerName = extractedIso + "_"
		}
	} else {
		extractedIso = extractIso(layerName)
		if extractedIso != "" {
			layerName += data.Title + "_"

		} else {
			layerName = data.Title
		}
	}
	if (len(data.Children) == 0) || strings.Contains(replaceSpaceAndMakeSmallCase(data.Title), "planned_transmission_upgrades") {
		if string(layerName[len(layerName)-1]) == "_" {
			layerName = layerName[:len(layerName)-1]
		}
		requiredLayers := []string{
			"substations",
			"operational_resources",
			"planned_resources",
			"resource_node_lmp_basis_analysis_contour",
			"yearly_ancillary_services_pricing",
			"basis_analysis",
			"load_forecast_contours",
			"planned_transmission_upgrades",
			"top_50_binding_constraints",
			"rtp_load_forecast_contours",
		}
		var key string
		isIsoRequired := slices.Contains(requiredLayers, replaceSpaceAndMakeSmallCase(data.Title))
		if isIsoRequired {
			key = replaceSpaceAndMakeSmallCase(layerName)
		} else {
			key = replaceSpaceAndMakeSmallCase(data.Title)
		}
		(*dict)[key] = strings.Join(paths, ` ->  `)
		return
	}

	for i := 0; i < len(data.Children); i++ {
		currChild := data.Children[i]
		recursivelyCheckPath(currChild, layerName, depth+1, paths, dict)
	}
}
func extractIso(layerName string) string {
	isos := []string{"ercot", "miso", "pjm", "caiso", "nyiso", "spp", "iso-ne", "wecc", "serc"}
	for _, iso := range isos {
		if strings.HasPrefix(strings.ToLower(strings.ReplaceAll(layerName, " ", "_")), iso) {
			return iso
		}
	}
	return ""
}
