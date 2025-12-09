package tools

import (
	"ai-service-go/internals/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func makeGetRequest[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get layer information, status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func getAllAvailableLayers(ctx context.Context) ([]AvailableLayersData, error) {
	type response struct {
		Data []AvailableLayersData `json:"data"`
	}
	url := config.AppConfig.AllLayersURL
	var result *response
	result, err := makeGetRequest[response](ctx, url)
	if err != nil {
		return nil, err
	}
	// remove internal layers
	filteredLayers := []AvailableLayersData{}
	for _, layer := range result.Data {
		isInternal := slices.Contains(InternalLayers, layer.Name)
		if !isInternal {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	AllAvailableLayers = filteredLayers
	return AllAvailableLayers, nil
}

func getAllPaths() map[string]string {
	// remove internal layers
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
		recursivelyCheckPath(datas[i], 0, []string{}, &dictionary)
	}
	return dictionary
}

func replaceSpaceAndMakeSmallCase(str string) string {
	return strings.ToLower(strings.ReplaceAll(str, " ", "_"))
}
func makeLayerLikeTitle(layerName string) string {
	// extract iso if present
	layerName = strings.ReplaceAll(layerName, "_", " ")
	layerName = cases.Title(language.English).String(layerName)
	iso := extractIso(layerName)
	if iso != "" {
		// After title casing, the iso will be title-cased (e.g., "Ercot"), so we need to replace that
		titleCasedIso := cases.Title(language.English).String(iso)
		layerName = strings.ReplaceAll(layerName, titleCasedIso, strings.ToUpper(iso))
	}
	return layerName
}

func recursivelyCheckPath(data dataType, depth int, paths []string, dict *map[string]string) {
	paths = append(paths, data.Title)
	// handle exceptions
	// case 1 - remove children for these layers
	exceptionsCase1 := []string{
		"planned_transmission_upgrades",
		"ercot_765kv_step_phase_1",
		"ercot_765kv_step_phase_2",
		"ercot_765kv_step_phase_3",
	}
	// case 2 - all children of case2 are considered as case2_child
	exceptionsCase2 := []string{
		"ercot_765kv_step",
		"yearly_resource_node_lmp",
		"data_center",
	}
	if len(paths) > 1 {
		parentLayer := replaceSpaceAndMakeSmallCase(paths[len(paths)-2])
		for _, exc := range exceptionsCase2 {
			if strings.Contains(replaceSpaceAndMakeSmallCase(parentLayer), exc) {
				data.Title = fmt.Sprintf(`%s_%s`, parentLayer, data.Title)
				break
			}
		}
		for _, exc := range exceptionsCase1 {
			if strings.Contains(replaceSpaceAndMakeSmallCase(data.Title), exc) {
				data.Children = []dataType{}
				break
			}
		}
	}

	if len(data.Children) == 0 {
		key := replaceSpaceAndMakeSmallCase(data.Title)
		(*dict)[key] = strings.Join(paths, ` ->  `)
		return
	}

	for _, currChild := range data.Children {
		recursivelyCheckPath(currChild, depth+1, paths, dict)
	}
}

func extractIso(layerName string) string {
	isos := []string{"ercot", "miso", "pjm", "caiso", "nyiso", "spp", "iso-ne", "wecc", "serc"}
	for _, iso := range isos {
		if strings.HasPrefix(replaceSpaceAndMakeSmallCase(layerName), iso) {
			return iso
		}
	}
	return ""
}
