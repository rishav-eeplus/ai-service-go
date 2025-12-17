package vector_db

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/utils"
	"context"
	"fmt"
	"os"

	"github.com/qdrant/go-client/qdrant"
)

// VectorStore handles all vector operations
type VectorStore struct {
	Qdrant         *qdrant.Client
	CollectionName string
	VectorSize     uint64
	NChuncks       int
}

var VectorStoreManager VectorStore

// NewVectorStore creates a new VectorStore instance
func NewVectorStore() {
	appConfig := config.AppConfig
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{
		Host: appConfig.QdrantHost,
	})
	if err != nil {
		logger.Logger.Fatalf("Failed to create Qdrant client: %v", err)
	}
	logger.Logger.Info("Vector Store Manager loaded successfully ✅")
	VectorStoreManager = VectorStore{
		Qdrant:         qdrantClient,
		CollectionName: "documents",
		VectorSize:     uint64(appConfig.VectorSize),
		NChuncks:       appConfig.NChunks,
	}
	VectorStoreManager.Initialize()
}

// Initialize initializes the vector store
func (vs *VectorStore) Initialize() error {
	ctx := context.Background()

	// Check if collection exists
	collections, err := vs.Qdrant.ListCollections(ctx)
	if err != nil {
		logger.Logger.Errorf("Failed to list collections: %v", err)
		return err
	}

	exists := false
	for _, col := range collections {
		if col == vs.CollectionName {
			exists = true
			break
		}
	}

	if !exists {
		// Create collection
		err = vs.Qdrant.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: vs.CollectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vs.VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			logger.Logger.Errorf("Failed to create collection: %v", err)
			return err
		}
		logger.Logger.Infof("Created collection: %s", vs.CollectionName)
	}
	logger.Logger.Info("VectorStore Manager initialized successfully ✅")
	return nil
}

// LoadEmbeddings loads embeddings from data.txt file
func (vs *VectorStore) LoadEmbeddings() error {
	ctx := context.Background()

	// Clear collection first
	err := vs.ClearCollection(true)
	if err != nil {
		return err
	}

	// Read data file
	dataPath := "./data.txt"
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "../data.txt"
	}

	content, err := os.ReadFile(dataPath)
	if err != nil {
		logger.Logger.Errorf("Failed to read data file: %v", err)
		return err
	}

	text := string(content)

	// Split text into chunks
	chunks, err := utils.SplitText(text, 2500, 250)
	if err != nil {
		return err
	}

	logger.Logger.Infof("Processing %d chunks", len(chunks))

	// Generate embeddings and create points
	var points []*qdrant.PointStruct
	for i, chunk := range chunks {
		embedding, err := controllers.AiManager.GenerateEmbedding(chunk)
		if err != nil {
			return err
		}

		if len(embedding) != int(vs.VectorSize) {
			return fmt.Errorf("invalid embedding size: expected %d, got %d", vs.VectorSize, len(embedding))
		}

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: qdrant.NewValueMap(map[string]any{
				"text": chunk,
			}),
		}
		points = append(points, point)
	}

	// Upsert points to Qdrant
	_, err = vs.Qdrant.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: vs.CollectionName,
		Points:         points,
	})
	if err != nil {
		logger.Logger.Errorf("Failed to upsert points: %v", err)
		return err
	}

	logger.Logger.Infof("Successfully loaded %d embeddings", len(points))
	return nil
}

// LoadEmbeddingsV2 loads embeddings from data.txt file using section-based splitting
func (vs *VectorStore) LoadEmbeddingsV2() error {
	ctx := context.Background()

	// Clear collection first
	err := vs.ClearCollection(true)
	if err != nil {
		return err
	}

	// Read data file
	dataPath := "./data.txt"
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "../data.txt"
	}

	content, err := os.ReadFile(dataPath)
	if err != nil {
		logger.Logger.Errorf("Failed to read data file: %v", err)
		return err
	}

	text := string(content)

	// Split text into sections
	sections, err := utils.SplitTextBySections(text)
	if err != nil {
		logger.Logger.Errorf("Failed to split text into sections: %v", err)
		return err
	}

	logger.Logger.Infof("Processing %d sections", len(sections))

	// Generate embeddings and create points
	var points []*qdrant.PointStruct
	for i, section := range sections {
		embedding, err := controllers.AiManager.GenerateEmbedding(section.Content)
		if err != nil {
			return err
		}

		if len(embedding) != int(vs.VectorSize) {
			return fmt.Errorf("invalid embedding size: expected %d, got %d", vs.VectorSize, len(embedding))
		}

		// Create payload with section metadata
		payload := map[string]any{
			"text":    section.Content,
			"title":   section.Title,
			"summary": section.Summary,
		}

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)),
			Vectors: qdrant.NewVectors(embedding...),
			Payload: qdrant.NewValueMap(payload),
		}
		points = append(points, point)
	}

	// Upsert points to Qdrant
	_, err = vs.Qdrant.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: vs.CollectionName,
		Points:         points,
	})
	if err != nil {
		logger.Logger.Errorf("Failed to upsert points: %v", err)
		return err
	}

	logger.Logger.Infof("Successfully loaded %d embeddings with section metadata", len(points))
	return nil
}

// estimateTokenCount estimates the number of tokens in a text
// Uses a simple heuristic: ~4 characters per token (common for English text)
func estimateTokenCount(text string) int {
	// Rough estimation: 1 token ≈ 4 characters
	return len(text) / 4
}

// GetAllVectorsWithMetadata retrieves all vectors with their IDs and metadata
func (vs *VectorStore) GetAllVectorsWithMetadata() ([]map[string]interface{}, error) {
	ctx := context.Background()

	// Scroll through all points in the collection
	scrollResult, err := vs.Qdrant.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: vs.CollectionName,
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(false), // Don't include vectors to save bandwidth
		Limit:          qdrant.PtrOf(uint32(10000)),  // Get all points
	})
	if err != nil {
		logger.Logger.Errorf("Failed to scroll points: %v", err)
		return nil, err
	}

	var results []map[string]interface{}
	for _, point := range scrollResult {
		result := map[string]interface{}{
			"id": point.Id.GetNum(),
		}

		// Extract metadata from payload
		if point.Payload != nil {
			metadata := make(map[string]interface{})
			var textContent string

			if summary, ok := point.Payload["summary"]; ok {
				if summaryStr, ok := summary.GetKind().(*qdrant.Value_StringValue); ok {
					metadata["summary"] = summaryStr.StringValue
				}
			}

			if title, ok := point.Payload["title"]; ok {
				if titleStr, ok := title.GetKind().(*qdrant.Value_StringValue); ok {
					metadata["title"] = titleStr.StringValue
				}
			}

			// Get text content for token counting
			if text, ok := point.Payload["text"]; ok {
				if textStr, ok := text.GetKind().(*qdrant.Value_StringValue); ok {
					textContent = textStr.StringValue
				}
			}

			// Calculate token count
			tokenCount := estimateTokenCount(textContent)
			metadata["token_count"] = tokenCount

			result["metadata"] = metadata
		}

		results = append(results, result)
	}
	return results, nil
}

func (vs *VectorStore) GetContentByIDs(ids []uint64) ([]string, error) {
	ctx := context.Background()
	inputIds := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		inputIds[i] = qdrant.NewIDNum(id)
	}
	// Retrieve point by ID
	getResult, err := vs.Qdrant.Get(ctx, &qdrant.GetPoints{
		CollectionName: vs.CollectionName,
		Ids:            inputIds,
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(false),
	})
	if err != nil {
		logger.Logger.Errorf("Failed to get point by ID %v: %v", ids, err)
		return nil, err
	}
	// return content of all points
	var contents []string
	for _, point := range getResult {
		if textValue, ok := point.Payload["text"]; ok {
			if textStr, ok := textValue.GetKind().(*qdrant.Value_StringValue); ok {
				contents = append(contents, textStr.StringValue)
			}
		}
	}
	return contents, nil
}

// ClearCollection clears or drops the collection
func (vs *VectorStore) ClearCollection(dropCollection bool) error {
	ctx := context.Background()

	if dropCollection {
		err := vs.Qdrant.DeleteCollection(ctx, vs.CollectionName)
		if err != nil {
			logger.Logger.Warnf("Failed to delete collection: %v", err)
		}
		return vs.Initialize()
	}

	// Delete all points
	_, err := vs.Qdrant.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: vs.CollectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: &qdrant.Filter{
					Must: []*qdrant.Condition{},
				},
			},
		},
	})

	return err
}

// SearchSimilarChunks searches for similar chunks
func (vs *VectorStore) SearchSimilarChunks(query string, limit int, scoreThreshold float32) (string, error) {
	ctx := context.Background()

	if limit == 0 {
		limit = vs.NChuncks
	}
	if scoreThreshold == 0 {
		scoreThreshold = 0.001
	}
	// Generate query embedding
	queryEmbedding, err := controllers.AiManager.GenerateEmbedding(query)
	if err != nil {
		return "", err
	}

	// Search in Qdrant
	searchResult, err := vs.Qdrant.Query(ctx, &qdrant.QueryPoints{
		CollectionName: vs.CollectionName,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		ScoreThreshold: &scoreThreshold,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		logger.Logger.Errorf("Search error: %v", err)
		return "", fmt.Errorf("failed to search similar chunks")
	}

	// Combine results
	var results []string
	for _, result := range searchResult {
		if textValue, ok := result.Payload["text"]; ok {
			if textStr, ok := textValue.GetKind().(*qdrant.Value_StringValue); ok {
				results = append(results, textStr.StringValue)
			}
		}
	}

	return fmt.Sprintf("%s", results), nil
}
