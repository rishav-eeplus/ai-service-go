package data

import (
	"regexp"
	"strings"
)

// MatchLayerName matches user input to layer names with a threshold
func MatchLayerName(userInput string, layerDict map[string]map[string]string, threshold float64) []string {
	matchScores := make(map[string]float64)

	for layer := range layerDict {
		score := 0.0
		matched := false

		// Exact phrase match
		pattern := `\b` + regexp.QuoteMeta(layer) + `\b`
		re := regexp.MustCompile(`(?i)` + pattern)
		if re.MatchString(userInput) {
			score = 100.0
			matched = true
		}

		// Partial phrase match
		if !matched {
			similarity := calculateSimilarity(strings.ToLower(userInput), strings.ToLower(layer))
			if similarity > 0.6 {
				score = similarity * 50
				matched = true
			}
		}

		if matched {
			matchScores[layer] = score
		}
	}

	// Filter and sort results
	var results []struct {
		layer string
		score float64
	}

	for layer, score := range matchScores {
		if score >= threshold {
			results = append(results, struct {
				layer string
				score float64
			}{layer, score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Extract layer names
	var matched []string
	for _, r := range results {
		matched = append(matched, r.layer)
	}

	return matched
}

// calculateSimilarity calculates Jaccard similarity between two strings
func calculateSimilarity(str1, str2 string) float64 {
	words1 := strings.Fields(strings.ToLower(str1))
	words2 := strings.Fields(strings.ToLower(str2))

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
