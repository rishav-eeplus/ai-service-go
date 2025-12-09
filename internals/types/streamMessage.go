package types

import "encoding/json"

// StreamMessage represents messages sent via WebSocket
type StreamMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Options []string        `json:"options,omitempty"` // For clarification options
}
