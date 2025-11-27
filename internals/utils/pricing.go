package utils

import "fmt"

// TokenUsage represents OpenAI token usage
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CalculatePrice calculates the price based on token usage and model
func CalculatePrice(usage TokenUsage, model string) interface{} {
	if usage.CompletionTokens == 0 || usage.PromptTokens == 0 {
		return "invalid inputs"
	}

	switch model {
	case "gpt-4o":
		price := (float64(usage.PromptTokens)*2.5 + float64(usage.CompletionTokens)*10) / 1000000
		return fmt.Sprintf("%.6f", price)
	case "gpt-4o-mini":
		price := (float64(usage.PromptTokens)*0.15 + float64(usage.CompletionTokens)*0.6) / 1000000
		return fmt.Sprintf("%.6f", price)
	default:
		return "model not defined"
	}
}
