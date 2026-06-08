package types

import (
	"encoding/json"
	"time"
)

// StreamMessage represents messages sent via WebSocket
type StreamMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Options []string        `json:"options,omitempty"` // For clarification options
}

type AnnaChatType struct {
	UserName     string    `json:"user_name"`
	Query        string    `json:"query"`
	Response     string    `json:"response"`
	Feedback     int       `json:"feedback"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
}
