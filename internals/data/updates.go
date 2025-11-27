package data

import (
	"fmt"
	"strings"
)

var LayerUpdateCycleDictionary = map[string]map[string]string{
	"Substations":                                       {"trial": "annually", "standard": "quarterly"},
	"Operational Resources":                             {"trial": "annually", "standard": "quarterly"},
	"Planned Resources":                                 {"trial": "annually", "standard": "quarterly"},
	"Injection Capacity Contours":                       {"trial": "annually", "standard": "quarterly"},
	"Resource Node LMP Basis Analysis":                  {"trial": "annually", "standard": "annually"},
	"Resource Node LMP Basis Analysis Contours":         {"trial": "annually", "standard": "annually"},
	"Ancillary Services":                                {"trial": "annually", "standard": "annually"},
	"Binding Constraints":                               {"trial": "annually", "standard": "quarterly"},
	"Planned Transmission Upgrades":                     {"trial": "annually", "standard": "quarterly"},
	"Load Forecast Contours":                            {"trial": "annually", "standard": "quarterly"},
	"Large Load Data":                                   {"trial": "annually", "standard": "annually"},
	"Daily Average LMP Chart":                           {"trial": "annually", "standard": "annually"},
	"Transmission lines/ Nationwide Transmission lines": {"trial": "quarterly", "standard": "annually"},
}

var DataAvailableDictionary = map[string]map[string]string{
	"Substations":                               {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Operational Resources":                     {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Planned Resources":                         {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Injection Capacity Contours":               {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Resource Node LMP Basis Analysis":          {"trial": "for years 2021-2023", "standard": "for years 2022-2024"},
	"Resource Node LMP Basis Analysis Contours": {"trial": "for years 2021-2023", "standard": "for years 2022-2024"},
	"Ancillary Services":                        {"trial": "for year 2022-2024", "standard": "for years 2022-2024"},
	"Binding Constraints":                       {"trial": "up to last_year", "standard": "for current_quarter - current_year"},
	"Planned Transmission Upgrades":             {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Load Forecast Contours":                    {"trial": "up to Q4 last_year", "standard": "up to current_quarter - current_year"},
	"Large Load Data":                           {"trial": "for year 2024-2028", "standard": "for year 2025-2031"},
	"Daily Average LMP Chart":                   {"trial": "for year 2022-2024", "standard": "for year 2022-2024"},
	"Transmission lines/ Nationwide Transmission lines":             {"trial": "up to Q4 last_year", "standard": "upto current_year"},
}

// LayerInfo represents update information for a layer
type LayerInfo struct {
	Layer         string
	UpdateCycle   string
	DataAvailable string
	IsAvailable   bool
}

// GetUpdatesOfLayer returns update information for a specific layer
func GetUpdatesOfLayer(layer, platform string, currentYear, currentQuarter int) LayerInfo {
	if layer == "" || platform == "" {
		return LayerInfo{
			Layer:         layer,
			UpdateCycle:   "Data not available",
			DataAvailable: "Data not available",
			IsAvailable:   false,
		}
	}

	matchedLayers := MatchLayerName(layer, LayerUpdateCycleDictionary, 50)

	if len(matchedLayers) == 0 {
		return LayerInfo{
			Layer:         layer,
			UpdateCycle:   "Data not available",
			DataAvailable: "Data not available",
			IsAvailable:   false,
		}
	}

	matchedLayer := matchedLayers[0]
	updateCycle := LayerUpdateCycleDictionary[matchedLayer][platform]
	dataAvailable := DataAvailableDictionary[matchedLayer][platform]

	// Replace placeholders
	dataAvailable = strings.ReplaceAll(dataAvailable, "current_year", fmt.Sprintf("%d", currentYear))
	dataAvailable = strings.ReplaceAll(dataAvailable, "last_year", fmt.Sprintf("%d", currentYear-1))
	dataAvailable = strings.ReplaceAll(dataAvailable, "current_quarter", fmt.Sprintf("Q%d", currentQuarter))

	return LayerInfo{
		Layer:         layer,
		UpdateCycle:   updateCycle,
		DataAvailable: dataAvailable,
		IsAvailable:   updateCycle != "Data not available" && dataAvailable != "Data not available",
	}
}

// GenerateResponseForLayerUpdateQuery generates a formatted response for layer update queries
func GenerateResponseForLayerUpdateQuery(layers []string, platform string, currentYear, currentQuarter int) string {
	if len(layers) == 0 {
		return ""
	}

	var addedInfo strings.Builder
	addedInfo.WriteString("\n")

	var allLayerInfo []LayerInfo
	for _, layer := range layers {
		info := GetUpdatesOfLayer(layer, platform, currentYear, currentQuarter)
		allLayerInfo = append(allLayerInfo, info)
	}

	var availableLayers []LayerInfo
	var unavailableLayers []LayerInfo
	for _, info := range allLayerInfo {
		if info.IsAvailable {
			availableLayers = append(availableLayers, info)
		} else {
			unavailableLayers = append(unavailableLayers, info)
		}
	}

	// Add available layers info
	for _, info := range availableLayers {
		addedInfo.WriteString(fmt.Sprintf("**%s**", info.Layer))

		var updateFreqSentence string
		if info.UpdateCycle == "Most Recent data available" {
			updateFreqSentence = " contains the most recent data available."
		} else {
			updateFreqSentence = fmt.Sprintf(" is updated %s.", strings.ToLower(info.UpdateCycle))
		}

		var dataCoverageSentence string
		if strings.HasPrefix(info.DataAvailable, "Most Recent") {
			dataCoverageSentence = "The most recent data is available for this layer."
		} else {
			dataCoverageSentence = fmt.Sprintf("Data is currently available %s.", strings.ToLower(info.DataAvailable))
		}

		addedInfo.WriteString(fmt.Sprintf("%s %s\n", updateFreqSentence, dataCoverageSentence))
	}

	// Add unavailable layers info
	if len(unavailableLayers) > 0 {
		if len(unavailableLayers) > 1 {
			addedInfo.WriteString("\n⚠️  Unfortunately, I don't have update information available for the following layers: ")
			var layerNames []string
			for _, info := range unavailableLayers {
				layerNames = append(layerNames, info.Layer)
			}
			if len(layerNames) > 1 {
				addedInfo.WriteString(strings.Join(layerNames[:len(layerNames)-1], ", "))
				addedInfo.WriteString(" and ")
				addedInfo.WriteString(layerNames[len(layerNames)-1])
			}
			addedInfo.WriteString(".\n")
		} else {
			addedInfo.WriteString(fmt.Sprintf("\n⚠️  Unfortunately, I don't have update information available for %s.\n", unavailableLayers[0].Layer))
		}
	}

	if platform == "trial" {
		addedInfo.WriteString("\nNote: As you are using the trial version of EEHORIZON, some data layers may have limited update information compared to the standard version.\n")
	}

	return addedInfo.String()
}
