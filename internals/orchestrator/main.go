package orchestrator

import (
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"ai-service-go/internals/vector_db"
	"context"
	"encoding/json"
	"fmt"

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
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message,omitempty"`
}

type ClientRequestType struct {
	UserQuery            string
	PreviousConversation string
	Platform             string
}

func (o *Orchestrator) Run(conn *websocket.Conn, input *ClientRequestType, model string) {

	sendMessage := func(msg StreamMessage) bool {
		if err := conn.WriteJSON(msg); err != nil {
			utils.Logger.Errorf("Error writing to WebSocket: %v", err)
			return false
		}
		return true
	}

	useful_tools, reasoning, err := o.Router(sendMessage, input)
	if err != nil {
		utils.Logger.Errorf("Router error: %v", err)
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Failed to identify intents: %v", err),
		})
		return
	}

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

	// Step 3: Planning and Tool Execution
	planOutput, err := o.PlannerAndToolExecuter(sendMessage, input, useful_tools, reasoning, model, context.Background())
	if err != nil {
		utils.Logger.Errorf("Planner error: %v", err)
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Failed to generate response: %v", err),
		})
		return
	}
	//
	planOutputJSON, err := json.Marshal(planOutput)
	if err != nil {
		utils.Logger.Errorf("Error marshalling plan output: %v", err)
		sendMessage(StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Failed to process response: %v", err),
		})
		return
	}
	// Final Output// Step 5: Send final response
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
	conn.Close()
}

func extractIntentNames(intents []tools.Tool) []string {
	var names []string
	for _, tool := range intents {
		names = append(names, tool.Name())
	}
	return names
}
