package handlers

import (
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/data"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/orchestrator"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"ai-service-go/internals/vector_db"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Handler struct {
	VectorStoreManager *vector_db.VectorStore
	AiManager          *controllers.OpenAIManager
	ToolRegistry       *tools.ToolRegistry
	Orchestrator       *orchestrator.Orchestrator
}

// SuccessResponse struct to hold the success response
type SuccessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Content any    `json:"content"`
}

// ErrorResponse struct to hold the error response
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

func NewHandler(vs *vector_db.VectorStore, aim *controllers.OpenAIManager, tr *tools.ToolRegistry, orch *orchestrator.Orchestrator) *Handler {
	return &Handler{
		VectorStoreManager: vs,
		AiManager:          aim,
		ToolRegistry:       tr,
		Orchestrator:       orch,
	}
}

// SendSuccessResponse sends a success response
func SendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data any) {
	w.WriteHeader(statusCode)
	response := SuccessResponse{
		Status:  "success",
		Message: message,
		Content: data,
		Data:    data}
	encwoder := json.NewEncoder(w)
	encwoder.SetEscapeHTML(false)
	if err := encwoder.Encode(response); err != nil {
		utils.HandleError(w, err, "Error encoding json data", "error", 500)
		return
	}
}

// handleStatus returns the status of the vector store
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	collections, err := h.VectorStoreManager.Qdrant.ListCollections(ctx)
	if err != nil {
		logger.Logger.Error("Error fetching vector store status")
		http.Error(w, "Error fetching vector store status", http.StatusInternalServerError)
		return
	}

	var currentCollection interface{}
	for _, col := range collections {
		if col == h.VectorStoreManager.CollectionName {
			currentCollection = map[string]interface{}{
				"name": col,
			}
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Status Fetched", map[string]interface{}{
		"initialized":           true,
		"status":                "ok",
		"vector_store":          "qdrant",
		"collection_name":       h.VectorStoreManager.CollectionName,
		"vector_size":           h.VectorStoreManager.VectorSize,
		"available_collections": collections,
		"current_collection":    currentCollection,
		"timestamp":             time.Now().UTC(),
	})

}

func (h *Handler) TempHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := h.ToolRegistry.Execute(r.Context(), "get_all_available_layers", nil, nil)
	SendSuccessResponse(w, 200, "Temp Handler", map[string]interface{}{
		"message": "Got the data",
		"data":    data,
	})
}

// handleLoadEmbeddings loads embeddings into the vector store
func (h *Handler) HandleLoadEmbeddings(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret != os.Getenv("SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.VectorStoreManager.LoadEmbeddings()
	if err != nil {
		logger.Logger.Errorf("Error while loading embeddings: %v", err)
		http.Error(w, "Error while loading embeddings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Embeddings loaded successfully", nil)
}

// HandleLoadEmbeddingsV2 loads embeddings into the vector store using section-based splitting
func (h *Handler) HandleLoadEmbeddingsV2(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret != os.Getenv("SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.VectorStoreManager.LoadEmbeddingsV2()
	if err != nil {
		logger.Logger.Errorf("Error while loading embeddings v2: %v", err)
		http.Error(w, "Error while loading embeddings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Embeddings loaded successfully with section metadata", nil)
}

// HandleGetAllVectors retrieves all vectors with their IDs and metadata
func (h *Handler) HandleGetAllVectors(w http.ResponseWriter, r *http.Request) {
	vectors, err := h.VectorStoreManager.GetAllVectorsWithMetadata()
	if err != nil {
		logger.Logger.Errorf("Error while fetching vectors: %v", err)
		http.Error(w, "Error while fetching vectors", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Vectors fetched successfully", map[string]interface{}{
		"count":   len(vectors),
		"vectors": vectors,
	})
}

// QueryRequest represents the request body for handle-query
type QueryRequest struct {
	Query                string `json:"query" binding:"required"`
	PreviousConversation string `json:"previousConversation"`
}

// HandleQuery processes user queries
func (h *Handler) HandleQueryV1(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("Invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = "standard"
	}

	previousConversation := req.PreviousConversation
	if previousConversation == "" {
		previousConversation = "No previous conversation provided"
	}

	// Search for similar chunks
	retrievedChunks, err := h.VectorStoreManager.SearchSimilarChunks(req.Query, 0, 0)
	if err != nil {
		logger.Logger.Error("Could not search similar chunks")
		http.Error(w, "Something bad happened while handling query", http.StatusInternalServerError)
		return
	}

	// Generate response
	response, _, err := h.AiManager.GenerateResponseV1(req.Query, previousConversation, retrievedChunks, platform, "")
	if err != nil {
		logger.Logger.Error("Could not generate response" + err.Error())
		http.Error(w, "Something bad happened while generating response", http.StatusInternalServerError)
		return
	}
	// Handle update cycle query layers
	if len(response.UpdateCycleQueryLayers) > 0 {
		now := time.Now()
		currentYear := now.Year()
		currentQuarter := (int(now.Month()) + 2) / 3

		updateInfo := data.GenerateResponseForLayerUpdateQuery(
			response.UpdateCycleQueryLayers,
			platform,
			currentYear,
			currentQuarter,
		)
		response.Result += updateInfo
	}
	SendSuccessResponse(w, 200, "Query processed successfully", response)
}

// StreamMessage represents different types of messages sent via SSE
type StreamMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// HandleSSEQuery handles streaming query responses via Server-Sent Events
func (h *Handler) HandleSSEQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query                 string `json:"query"`
		PreviousConversation  string `json:"previousConversation"`
		Platform              string `json:"platform"`
		Model                 string `json:"model"`
		ClarificationResponse string `json:"clarificationResponse"`		
		ClarificationCount    int    `json:"clarificationCount"`	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("Error decoding request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Platform == "" {
		req.Platform = "standard"
	}
	if req.PreviousConversation == "" {
		req.PreviousConversation = "No previous conversation provided"
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	userInput := &orchestrator.ClientRequestType{
		UserQuery:             req.Query,
		PreviousConversation:  req.PreviousConversation,
		Platform:              req.Platform,
		ClarificationResponse: req.ClarificationResponse,
		ClarificationCount:    req.ClarificationCount,
	}

	// Call the SSE version
	h.Orchestrator.RunSSE(w, flusher, userInput, req.Model)
}
