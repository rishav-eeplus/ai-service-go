package orchestrator

import (
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"ai-service-go/internals/vector_db"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type Orchestrator struct {
	AIManager     controllers.OpenAIManager
	ToolResistory *tools.ToolRegistry
	VectorManager *vector_db.VectorStore
}

func NewOrchestrator(aiManager controllers.OpenAIManager, toolRegistry *tools.ToolRegistry, vectorManager *vector_db.VectorStore) *Orchestrator {
	return &Orchestrator{
		AIManager:     aiManager,
		ToolResistory: toolRegistry,
		VectorManager: vectorManager,
	}
}

// StreamMessage represents messages sent via WebSocket
type StreamMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Options []string        `json:"options,omitempty"` // For clarification options
}

type ClientRequestType struct {
	UserQuery             string
	PreviousConversation  string
	Platform              string
	ClarificationResponse string // User's response to a clarification request
}

func (o *Orchestrator) Run(conn *websocket.Conn, input *ClientRequestType, model string) {

	sendMessage := func(msg StreamMessage) bool {
		if err := conn.WriteJSON(msg); err != nil {
			utils.Logger.Errorf("Error writing to WebSocket: %v", err)
			return false
		}
		return true
	}

	// Helper function to wait for user response with extended timeout
	waitForUserResponse := func() (string, error) {
		// Extend the read deadline to give user time to respond (5 minutes)
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		var response struct {
			Type     string `json:"type"`
			Response string `json:"response"`
		}
		if err := conn.ReadJSON(&response); err != nil {
			utils.Logger.Errorf("Error reading clarification response: %v", err)
			return "", err
		}

		// Reset deadline back to normal
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		return response.Response, nil
	}

	// useful_tools, reasoning, err := o.Router(sendMessage, input)
	// if err != nil {
	// 	utils.Logger.Errorf("Router error: %v", err)
	// 	sendMessage(StreamMessage{
	// 		Type:    "error",
	// 		Message: fmt.Sprintf("Failed to identify intents: %v", err),
	// 	})
	// 	return
	// }

	// Step 2: Retrieval
	// retrievedContext, err := o.Retreiver(sendMessage, context.Background(), input, extractIntentNames(identifiedIntents))
	// if err != nil {
	// 	utils.Logger.Errorf("Retriever error: %v", err)
	// 	sendMessage(StreamMessage{
	// 		Type:    "error",
	// 		Message: fmt.Sprintf("Failed to retrieve context: %v", err),
	// 	})
	// 	return
	// }

	// Loop to handle clarification requests
	maxClarificationLoops := 3
	for i := 0; i < maxClarificationLoops; i++ {
		// Step 3: Planning and Tool Execution
		planOutput, err := o.PlannerAndToolExecuter(sendMessage, input, model, context.Background(), maxClarificationLoops-i)
		if err != nil {
			utils.Logger.Errorf("Planner error: %v", err)
			sendMessage(StreamMessage{
				Type:    "error",
				Message: fmt.Sprintf("Failed to generate response: %v", err),
			})
			return
		}

		// Check if clarification is needed
		if planOutput.NeedsClarification && len(planOutput.Options) > 0 {
			// Send clarification request to client
			if !sendMessage(StreamMessage{
				Type:    "clarification",
				Message: planOutput.ClarificationMessage,
				Options: planOutput.Options,
			}) {
				return
			}

			// Wait for user response
			userResponse, err := waitForUserResponse()
			if err != nil {
				utils.Logger.Errorf("Error waiting for user clarification: %v", err)
				sendMessage(StreamMessage{
					Type:    "error",
					Message: "Failed to receive your selection. Please try again.",
				})
				return
			}

			// Update input with user's clarification response and continue the loop
			input.UserQuery = fmt.Sprintf("%s. User selected: %s", input.UserQuery, userResponse)
			input.ClarificationResponse = userResponse
			sendMessage(StreamMessage{
				Type:    "info",
				Message: fmt.Sprintf("Processing your selection: %s", userResponse),
			})
			continue
		}

		// No clarification needed, send the final response
		planOutputJSON, err := json.Marshal(planOutput)
		if err != nil {
			utils.Logger.Errorf("Error marshalling plan output: %v", err)
			sendMessage(StreamMessage{
				Type:    "error",
				Message: fmt.Sprintf("Failed to process response: %v", err),
			})
			return
		}

		// Final Output
		if !sendMessage(StreamMessage{
			Type:    "response",
			Data:    planOutputJSON,
			Message: "Response generated successfully",
		}) {
			return
		}

		// Send completion
		sendMessage(StreamMessage{
			Type:    "complete",
			Message: "Processing complete",
		})
		return
	}

	// If we've exceeded max clarification loops (this should rarely happen now since last loop forces a response)
	sendMessage(StreamMessage{
		Type:    "warning",
		Message: "Maximum clarification attempts reached. The response was generated with the available context.",
	})
}