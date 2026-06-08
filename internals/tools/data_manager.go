package tools

import (
	"ai-service-go/internals/logger"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const TTL_Duration = 1 * time.Hour

// LayerDataManager manages all cached data for layer-related operations
type LayerDataManager struct {
	// Layer information cache
	availableLayers []AvailableLayersData
	// Layer path navigation cache
	layerPathMappings map[string]string
	// Update cycle information cache
	updateCycleMappings      map[string]map[string]string
	dataAvailabilityMappings map[string]map[string]string
	// Loading state tracking
	layersLoaded     bool
	pathsLoaded      bool
	updateInfoLoaded bool
	// Last loaded timestamps for TTL management
	layersLastLoaded time.Time
	// pathsLastLoaded      time.Time
	// updateInfoLastLoaded time.Time
	// Thread safety
	mutex sync.RWMutex
}

var (
	dataManagerInstance *LayerDataManager
	dataManagerOnce     sync.Once
)

// GetDataManager returns the singleton instance of LayerDataManager
func GetDataManager() *LayerDataManager {
	dataManagerOnce.Do(func() {
		dataManagerInstance = &LayerDataManager{
			availableLayers:          make([]AvailableLayersData, 0),
			layerPathMappings:        make(map[string]string),
			updateCycleMappings:      make(map[string]map[string]string),
			dataAvailabilityMappings: make(map[string]map[string]string),
		}
	})
	return dataManagerInstance
}

// GetAvailableLayers returns the cached available layers, loading them if necessary
func (dm *LayerDataManager) GetAvailableLayers(ctx context.Context) ([]AvailableLayersData, error) {
	dm.mutex.RLock()
	if dm.layersLoaded && len(dm.availableLayers) > 0 && time.Since(dm.layersLastLoaded) < TTL_Duration {
		result := make([]AvailableLayersData, len(dm.availableLayers))
		copy(result, dm.availableLayers)
		dm.mutex.RUnlock()
		return result, nil
	}
	dm.mutex.RUnlock()
	return dm.loadAvailableLayers(ctx)
}

// GetLayerPathMappings returns the cached layer path mappings, loading them if necessary
func (dm *LayerDataManager) GetLayerPathMappings() (map[string]string, error) {
	dm.mutex.RLock()
	if dm.pathsLoaded && len(dm.layerPathMappings) > 0 {
		result := make(map[string]string, len(dm.layerPathMappings))
		maps.Copy(result, dm.layerPathMappings)
		dm.mutex.RUnlock()
		return result, nil
	}
	dm.mutex.RUnlock()
	return dm.loadLayerPaths()
}

// GetUpdateCycleInfo returns the cached update cycle information, loading it if necessary
func (dm *LayerDataManager) GetUpdateCycleInfo() (map[string]map[string]string, map[string]map[string]string, error) {
	dm.mutex.RLock()
	if dm.updateInfoLoaded && len(dm.updateCycleMappings) > 0 {
		updateCycles := make(map[string]map[string]string, len(dm.updateCycleMappings))
		dataAvailability := make(map[string]map[string]string, len(dm.dataAvailabilityMappings))

		for k, v := range dm.updateCycleMappings {
			updateCycles[k] = make(map[string]string)
			maps.Copy(updateCycles[k], v)
		}

		for k, v := range dm.dataAvailabilityMappings {
			dataAvailability[k] = make(map[string]string)
			maps.Copy(dataAvailability[k], v)
		}

		dm.mutex.RUnlock()
		return updateCycles, dataAvailability, nil
	}
	dm.mutex.RUnlock()

	return dm.loadUpdateCycleInfo()
}

// GetLayerPath returns the navigation path for a specific layer
func (dm *LayerDataManager) GetLayerPath(layerName string) (string, error) {
	paths, err := dm.GetLayerPathMappings()
	if err != nil {
		return "", err
	}

	normalizedName := replaceSpaceAndMakeSmallCase(layerName)
	if path, exists := paths[normalizedName]; exists {
		return path, nil
	}

	return "", fmt.Errorf("layer path not found for: %s", layerName)
}

// GetAllLayersUpdateInfo returns all update information for internal use
func (dm *LayerDataManager) GetAllLayersUpdateInfo() (map[string]map[string]string, error) {
	updateCycles, dataAvailability, err := dm.GetUpdateCycleInfo()
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	for layerName := range updateCycles {
		result[layerName] = map[string]string{
			"update_cycle_trial":      updateCycles[layerName]["trial"],
			"data_available_trial":    dataAvailability[layerName]["trial"],
			"update_cycle_standard":   updateCycles[layerName]["standard"],
			"data_available_standard": dataAvailability[layerName]["standard"],
		}
	}

	return result, nil
}

// GetLayerUpdateInfo returns update information for a specific layer and platform
func (dm *LayerDataManager) GetLayerUpdateInfo(layerName, platform string) (string, string, error) {
	updateCycles, dataAvailability, err := dm.GetUpdateCycleInfo()
	if err != nil {
		return "", "", err
	}

	normalizedName := replaceSpaceAndMakeSmallCase(layerName)
	updateCycle := "This layer is maintained with the latest available data, and updates are applied as soon as new information is released"
	dataAvailable := "Latest available data is maintained and updated as new information becomes available"

	if cycleInfo, exists := updateCycles[normalizedName]; exists {
		if cycle, exists := cycleInfo[platform]; exists {
			updateCycle = cycle
		}
	}

	if availInfo, exists := dataAvailability[normalizedName]; exists {
		if avail, exists := availInfo[platform]; exists {
			dataAvailable = avail
		}
	}

	return updateCycle, dataAvailable, nil
}

// loadAvailableLayers loads available layers from API
func (dm *LayerDataManager) loadAvailableLayers(ctx context.Context) ([]AvailableLayersData, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Double-check after acquiring write lock
	if dm.layersLoaded && len(dm.availableLayers) > 0 && time.Since(dm.layersLastLoaded) < TTL_Duration {
		result := make([]AvailableLayersData, len(dm.availableLayers))
		copy(result, dm.availableLayers)
		return result, nil
	}
	logger.Logger.Info("Loading available layers from API...")
	// Call the existing helper function
	layers, err := getAllAvailableLayersFromAPI(ctx)
	if err != nil {
		logger.Logger.Error("Failed to load available layers: " + err.Error())
		if len(dm.availableLayers) > 0 {
			logger.Logger.Warn("Returning stale available layers data due to API failure")
			result := make([]AvailableLayersData, len(dm.availableLayers))
			copy(result, dm.availableLayers)
			return result, nil
		}
		return nil, fmt.Errorf("failed to load available layers from API")
	}
	// Also load paths while we're at it to filter layers
	paths, pathsErr := dm.loadLayerPathsInternal()
	if pathsErr != nil {
		logger.Logger.Warn("Failed to load layer paths for filtering: " + pathsErr.Error())
		dm.availableLayers = layers
	} else {
		// Filter layers that have paths
		filteredLayers := make([]AvailableLayersData, 0, len(layers))
		for _, layer := range layers {
			if _, exists := paths[replaceSpaceAndMakeSmallCase(layer.Name)]; exists {
				filteredLayers = append(filteredLayers, layer)
			}
		}
		dm.availableLayers = filteredLayers
	}
	dm.layersLoaded = true
	dm.layersLastLoaded = time.Now()
	result := make([]AvailableLayersData, len(dm.availableLayers))
	copy(result, dm.availableLayers)
	logger.Logger.Info(fmt.Sprintf("Successfully loaded %d available layers", len(dm.availableLayers)))
	return result, nil
}

// loadLayerPaths loads layer navigation paths from data.json
func (dm *LayerDataManager) loadLayerPaths() (map[string]string, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	return dm.loadLayerPathsInternal()
}

// loadLayerPathsInternal is the internal implementation without mutex (for internal use)
func (dm *LayerDataManager) loadLayerPathsInternal() (map[string]string, error) {
	// Double-check after acquiring write lock
	if dm.pathsLoaded && len(dm.layerPathMappings) > 0 {
		result := make(map[string]string, len(dm.layerPathMappings))
		maps.Copy(result, dm.layerPathMappings)
		return result, nil
	}
	logger.Logger.Info("Loading layer paths from data.json...")
	var datas []dataType
	file, err := os.Open("./data.json")
	if err != nil {
		return nil, fmt.Errorf("error opening data.json: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&datas)
	if err != nil {
		return nil, fmt.Errorf("error decoding data.json: %w", err)
	}

	dm.layerPathMappings = make(map[string]string)
	for i := range datas {
		recursivelyCheckPath(datas[i], 0, []string{}, &dm.layerPathMappings)
	}

	dm.pathsLoaded = true

	result := make(map[string]string, len(dm.layerPathMappings))
	maps.Copy(result, dm.layerPathMappings)
	logger.Logger.Info(fmt.Sprintf("Successfully loaded %d layer paths", len(dm.layerPathMappings)))
	return result, nil
}

// loadUpdateCycleInfo loads update cycle information from updates.json
func (dm *LayerDataManager) loadUpdateCycleInfo() (map[string]map[string]string, map[string]map[string]string, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Double-check after acquiring write lock
	if dm.updateInfoLoaded && len(dm.updateCycleMappings) > 0 {
		updateCycles := make(map[string]map[string]string, len(dm.updateCycleMappings))
		dataAvailability := make(map[string]map[string]string, len(dm.dataAvailabilityMappings))

		for k, v := range dm.updateCycleMappings {
			updateCycles[k] = make(map[string]string)
			maps.Copy(updateCycles[k], v)
		}

		for k, v := range dm.dataAvailabilityMappings {
			dataAvailability[k] = make(map[string]string)
			maps.Copy(dataAvailability[k], v)
		}

		return updateCycles, dataAvailability, nil
	}

	logger.Logger.Info("Loading layer update information from updates.json...")
	loadedData, err := dm.loadUpdatesFileData()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load updates file: %w", err)
	}

	if len(loadedData) == 0 {
		return nil, nil, fmt.Errorf("no data found in updates.json")
	}

	dm.updateCycleMappings = make(map[string]map[string]string)
	dm.dataAvailabilityMappings = make(map[string]map[string]string)

	currentQuarter, currentYearForQuarterlyData, currentYearForAnnualData := dm.checkTime()

	for _, item := range loadedData {
		currentLayer := item.LayerName
		cycle := strings.ToLower(item.Update_Frequency)
		dm.processLayerUpdateInfo(currentLayer, cycle, item.Update_Frequency,
			currentQuarter, currentYearForQuarterlyData, currentYearForAnnualData)
	}

	dm.updateInfoLoaded = true

	// Return deep copies
	updateCycles := make(map[string]map[string]string, len(dm.updateCycleMappings))
	dataAvailability := make(map[string]map[string]string, len(dm.dataAvailabilityMappings))

	for k, v := range dm.updateCycleMappings {
		updateCycles[k] = make(map[string]string)
		maps.Copy(updateCycles[k], v)
	}

	for k, v := range dm.dataAvailabilityMappings {
		dataAvailability[k] = make(map[string]string)
		maps.Copy(dataAvailability[k], v)
	}

	logger.Logger.Info(fmt.Sprintf("Successfully loaded update information for %d layers", len(dm.updateCycleMappings)))
	return updateCycles, dataAvailability, nil
}

// processLayerUpdateInfo processes update information for a single layer
func (dm *LayerDataManager) processLayerUpdateInfo(layerName, cycle, originalFrequency string,
	currentQuarter, currentYearForQuarterlyData, currentYearForAnnualData int) {

	switch cycle {
	case "model_based":
		dm.updateCycleMappings[layerName] = map[string]string{
			"trial":    "The update cycle is typically driven by each ISO's model release schedule. Whenever we complete the update, we will notify you by email.",
			"standard": "The update cycle is typically driven by each ISO's model release schedule. Whenever we complete the update, we will notify you by email.",
		}
		dm.dataAvailabilityMappings[layerName] = map[string]string{
			"trial":    "As per update schedule",
			"standard": "As per update schedule",
		}
	case "iso_driven":
		dm.updateCycleMappings[layerName] = map[string]string{
			"trial":    "The update cycle is based on the timing of releases from the ISO.",
			"standard": "The update cycle is based on the timing of releases from the ISO.",
		}
		dm.dataAvailabilityMappings[layerName] = map[string]string{
			"trial":    "As per ISO releases",
			"standard": "As per ISO releases",
		}
	case "quarterly":
		dm.updateCycleMappings[layerName] = map[string]string{
			"trial":    "annually",
			"standard": "quarterly",
		}
		dm.dataAvailabilityMappings[layerName] = map[string]string{
			"trial":    fmt.Sprintf("up to Q1 - %d", currentYearForQuarterlyData),
			"standard": fmt.Sprintf("up to Q%d - %d", currentQuarter, currentYearForQuarterlyData),
		}
	case "annually":
		dm.updateCycleMappings[layerName] = map[string]string{
			"trial":    "annually",
			"standard": "annually",
		}
		if strings.Contains(layerName, "large_load") {
			dm.dataAvailabilityMappings[layerName] = map[string]string{
				"trial":    fmt.Sprintf("for year %d-%d", currentYearForAnnualData-1, currentYearForAnnualData+3),
				"standard": fmt.Sprintf("for year %d-%d", currentYearForAnnualData, currentYearForAnnualData+6),
			}
		} else {
			dm.dataAvailabilityMappings[layerName] = map[string]string{
				"trial":    fmt.Sprintf("for year %d-%d", currentYearForAnnualData-4, currentYearForAnnualData-2),
				"standard": fmt.Sprintf("for year %d-%d", currentYearForAnnualData-3, currentYearForAnnualData-1),
			}
		}
	default:
		dm.updateCycleMappings[layerName] = map[string]string{
			"trial":    originalFrequency,
			"standard": originalFrequency,
		}
		dm.dataAvailabilityMappings[layerName] = map[string]string{
			"trial":    originalFrequency,
			"standard": originalFrequency,
		}
	}
}

// checkTime calculates current quarter and years for update calculations
func (dm *LayerDataManager) checkTime() (int, int, int) {
	currentTime := time.Now()
	currentYearForQuarterlyData := currentTime.Year()
	currentYearForAnnualData := currentTime.Year()
	currentMonth := currentTime.Month()
	currentQuarter := (int(currentMonth)-1)/3 + 1

	// Check if we are past 3rd friday of current quarter
	thirdFriday := time.Date(currentYearForQuarterlyData, time.Month((currentQuarter-1)*3+1), 15, 0, 0, 0, 0, time.Local)
	for thirdFriday.Weekday() != time.Friday {
		thirdFriday = thirdFriday.AddDate(0, 0, 1)
	}
	if currentTime.Before(thirdFriday) {
		if currentQuarter == 1 {
			currentQuarter = 4
			currentYearForQuarterlyData -= 1
		} else {
			currentQuarter -= 1
		}
	}

	// Check if we are past Jan 31st
	jan31 := time.Date(currentYearForAnnualData, time.January, 31, 0, 0, 0, 0, time.Local)
	if currentTime.Before(jan31) {
		currentYearForAnnualData -= 1
	}

	return currentQuarter, currentYearForQuarterlyData, currentYearForAnnualData
}

// loadUpdatesFileData loads the raw data from updates.json
func (dm *LayerDataManager) loadUpdatesFileData() ([]struct {
	LayerName        string `json:"name"`
	Update_Frequency string `json:"update_frequency"`
}, error) {
	// Try different potential paths for the JSON file
	possiblePaths := []string{
		"./updates.json",
	}

	var filePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			filePath = path
			break
		}
	}

	if filePath == "" {
		// Try to find from current working directory
		wd, err := os.Getwd()
		if err == nil {
			potentialPath := filepath.Join(wd, "data", "updates.json")
			if _, err := os.Stat(potentialPath); err == nil {
				filePath = potentialPath
			}
		}
	}

	if filePath == "" {
		return nil, fmt.Errorf("updates.json file not found in expected locations")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading updates.json file: %w", err)
	}

	loadedData := []struct {
		LayerName        string `json:"name"`
		Update_Frequency string `json:"update_frequency"`
	}{}

	err = json.Unmarshal(content, &loadedData)
	if err != nil {
		return nil, fmt.Errorf("error parsing updates.json file: %w", err)
	}

	for i := 0; i < len(loadedData); i++ {
		loadedData[i].LayerName = replaceSpaceAndMakeSmallCase(loadedData[i].LayerName)
	}

	return loadedData, nil
}

// ClearCache clears all cached data (useful for testing or forced refresh)
func (dm *LayerDataManager) ClearCache() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.availableLayers = make([]AvailableLayersData, 0)
	dm.layerPathMappings = make(map[string]string)
	dm.updateCycleMappings = make(map[string]map[string]string)
	dm.dataAvailabilityMappings = make(map[string]map[string]string)

	dm.layersLoaded = false
	dm.pathsLoaded = false
	dm.updateInfoLoaded = false

	logger.Logger.Info("Data manager cache cleared")
}

// GetCacheStatus returns the current cache loading status
func (dm *LayerDataManager) GetCacheStatus() map[string]bool {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	return map[string]bool{
		"layers_loaded":      dm.layersLoaded,
		"paths_loaded":       dm.pathsLoaded,
		"update_info_loaded": dm.updateInfoLoaded,
	}
}
