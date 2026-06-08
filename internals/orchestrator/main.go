package orchestrator

import (
	"ai-service-go/internals/chats_db"
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/types"
	"ai-service-go/internals/vector_db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Orchestrator struct {
	AIManager     controllers.OpenAIManager
	ToolResistory *tools.ToolRegistry
	VectorManager *vector_db.VectorStore
	ChatDB        *chats_db.ChatDB
}

func NewOrchestrator(aiManager controllers.OpenAIManager, toolRegistry *tools.ToolRegistry, vectorManager *vector_db.VectorStore, chatDB *chats_db.ChatDB) *Orchestrator {
	return &Orchestrator{
		AIManager:     aiManager,
		ToolResistory: toolRegistry,
		VectorManager: vectorManager,
		ChatDB:        chatDB,
	}
}

type ClientRequestType struct {
	UserName              string
	UserQuery             string
	PreviousConversation  string
	Platform              string
	ClarificationResponse string // User's response to a clarification request
	ClarificationCount    int    // Number of clarifications already done
	ClientToolRun         struct {
		ToolName string
		Params   map[string]any
	}
}

func (o *Orchestrator) RunSSE(w http.ResponseWriter, flusher http.Flusher, input *ClientRequestType, model string) {
	sendMessage := func(msg types.StreamMessage) bool {
		data, err := json.Marshal(msg)
		if err != nil {
			logger.Logger.Errorf("Error marshalling message: %v", err)
			return false
		}

		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return true
	}

	// Send started message
	sendMessage(types.StreamMessage{
		Type:    "started",
		Message: "Processing your query...",
	})

	// Define max clarifications allowed
	const maxClarifications = 3
	// If user provided a clarification response, append it to the query
	if input.ClarificationResponse != "" {
		input.UserQuery = fmt.Sprintf("%s. User selected: %s", input.UserQuery, input.ClarificationResponse)
	}
	// Planning and Tool Execution
	planOutput, err := o.PlannerAndToolExecuter(sendMessage, input, model, context.Background(), maxClarifications-input.ClarificationCount)
	if err != nil {
		logger.Logger.Errorf("Planner error: %v", err)
		sendMessage(types.StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Failed to generate response: %v", err),
		})
		return
	}

	// Check if clarification is needed and if we have remaining clarification attempts
	if planOutput.NeedsClarification && len(planOutput.Options) > 0 {
		// Send clarification request with updated count and end stream
		type ClarificationData struct {
			Message            string   `json:"message"`
			Options            []string `json:"options"`
			ClarificationCount int      `json:"clarificationCount"`
		}
		clarificationJSON, _ := json.Marshal(ClarificationData{
			Message:            planOutput.ClarificationMessage,
			Options:            planOutput.Options,
			ClarificationCount: input.ClarificationCount + 1,
		})
		sendMessage(types.StreamMessage{
			Type:    "clarification",
			Message: planOutput.ClarificationMessage,
			Options: planOutput.Options,
			Data:    clarificationJSON,
		})
		return
	}

	// No clarification needed, send the final response
	planOutputJSON, err := json.Marshal(planOutput)
	if err != nil {
		logger.Logger.Errorf("Error marshalling plan output: %v", err)
		sendMessage(types.StreamMessage{
			Type:    "error",
			Message: fmt.Sprintf("Failed to process response: %v", err),
		})
		return
	}

	// Send response
	sendMessage(types.StreamMessage{
		Type:    "response",
		Data:    planOutputJSON,
		Message: "Response generated successfully",
	})

	// Send completion
	sendMessage(types.StreamMessage{
		Type:    "complete",
		Message: "Processing complete",
	})
}
