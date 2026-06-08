package utils

import "fmt"

// TokenUsage represents OpenAI token usage
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelPricing stores input and output prices per 1M tokens
type ModelPricing struct {
	Input  float64
	Output float64
}

// Prices per 1M tokens
var modelPrices = map[string]ModelPricing{
	// GPT-5 Series
	"gpt-5.2":             {Input: 1.75, Output: 14.00},
	"gpt-5.1":             {Input: 1.25, Output: 10.00},
	"gpt-5":               {Input: 1.25, Output: 10.00},
	"gpt-5-mini":          {Input: 0.25, Output: 2.00},
	"gpt-5-nano":          {Input: 0.05, Output: 0.40},
	"gpt-5.2-chat-latest": {Input: 1.75, Output: 14.00},
	"gpt-5.1-chat-latest": {Input: 1.25, Output: 10.00},
	"gpt-5-chat-latest":   {Input: 1.25, Output: 10.00},
	"gpt-5.1-codex-max":   {Input: 1.25, Output: 10.00},
	"gpt-5.1-codex":       {Input: 1.25, Output: 10.00},
	"gpt-5-codex":         {Input: 1.25, Output: 10.00},
	"gpt-5.2-pro":         {Input: 21.00, Output: 168.00},
	"gpt-5-pro":           {Input: 15.00, Output: 120.00},
	"gpt-5.1-codex-mini":  {Input: 0.25, Output: 2.00},
	"gpt-5-search-api":    {Input: 1.25, Output: 10.00},
	// GPT-4 Series
	"gpt-4.1":                    {Input: 2.00, Output: 8.00},
	"gpt-4.1-mini":               {Input: 0.40, Output: 1.60},
	"gpt-4.1-nano":               {Input: 0.10, Output: 0.40},
	"gpt-4o":                     {Input: 2.50, Output: 10.00},
	"gpt-4o-2024-05-13":          {Input: 5.00, Output: 15.00},
	"gpt-4o-mini":                {Input: 0.15, Output: 0.60},
	"gpt-4o-mini-search-preview": {Input: 0.15, Output: 0.60},
	"gpt-4o-search-preview":      {Input: 2.50, Output: 10.00},
	// Realtime Models
	"gpt-realtime":                 {Input: 4.00, Output: 16.00},
	"gpt-realtime-mini":            {Input: 0.60, Output: 2.40},
	"gpt-4o-realtime-preview":      {Input: 5.00, Output: 20.00},
	"gpt-4o-mini-realtime-preview": {Input: 0.60, Output: 2.40},
	// Audio Models
	"gpt-audio":                 {Input: 2.50, Output: 10.00},
	"gpt-audio-mini":            {Input: 0.60, Output: 2.40},
	"gpt-4o-audio-preview":      {Input: 2.50, Output: 10.00},
	"gpt-4o-mini-audio-preview": {Input: 0.15, Output: 0.60},
	// O-Series Models
	"o1":                    {Input: 15.00, Output: 60.00},
	"o1-pro":                {Input: 150.00, Output: 600.00},
	"o3-pro":                {Input: 20.00, Output: 80.00},
	"o3":                    {Input: 2.00, Output: 8.00},
	"o3-deep-research":      {Input: 10.00, Output: 40.00},
	"o4-mini":               {Input: 1.10, Output: 4.40},
	"o4-mini-deep-research": {Input: 2.00, Output: 8.00},
	"o3-mini":               {Input: 1.10, Output: 4.40},
	"o1-mini":               {Input: 1.10, Output: 4.40},
	// Other Models
	"codex-mini-latest":    {Input: 1.50, Output: 6.00},
	"computer-use-preview": {Input: 3.00, Output: 12.00},
	"gpt-image-1.5":        {Input: 5.00, Output: 10.00},
	"chatgpt-image-latest": {Input: 5.00, Output: 10.00},
	"gpt-image-1":          {Input: 5.00, Output: 10.00},
	"gpt-image-1-mini":     {Input: 2.00, Output: 0.00},
}

// CalculateCost calculates the cost in dollars based on token usage and model
func CalculateCost(inputTokens, outputTokens int, model string) (float64, error) {
	if inputTokens == 0 && outputTokens == 0 {
		return 0, fmt.Errorf("invalid token counts")
	}

	pricing, exists := modelPrices[model]
	if !exists {
		return 0, fmt.Errorf("model pricing not defined for: %s", model)
	}

	// Calculate cost: (tokens * price per 1M tokens) / 1,000,000
	cost := (float64(inputTokens)*pricing.Input + float64(outputTokens)*pricing.Output) / 1000000
	return cost, nil
}

// CalculatePrice (deprecated - use CalculateCost instead)
func CalculatePrice(usage TokenUsage, model string) interface{} {
	if usage.CompletionTokens == 0 || usage.PromptTokens == 0 {
		return "invalid inputs"
	}

	cost, err := CalculateCost(usage.PromptTokens, usage.CompletionTokens, model)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("$%.6f", cost)
}
