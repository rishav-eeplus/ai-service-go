package controllers

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/data"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIManager struct {
	OpenAI         *openai.Client
	EmbeddingModel string
	TextGenModel   string
	Instructions   string
}

// AIResponse represents the structured AI response
type AIResponse struct {
	Result                 string   `json:"result"`
	FollowUps              []string `json:"followUps"`
	UpdateCycleQueryLayers []string `json:"updateCycleQueryLayers"`
}

var AiManager OpenAIManager

func InitializeAiManager() {
	appConfig := config.AppConfig
	manager := openai.NewClient(appConfig.OpenaiAPIkey)
	AiManager = OpenAIManager{
		OpenAI:         manager,
		EmbeddingModel: appConfig.EmbeddingModel,
		TextGenModel:   appConfig.TextGenModel,
		Instructions:   data.PromptForUsingTools,
	}
	logger.Logger.Info("AI Manager loaded successfully ✅")
}

// GenerateEmbedding generates an embedding for the given text
func (aiM *OpenAIManager) GenerateEmbedding(text string) ([]float32, error) {
	ctx := context.Background()
	resp, err := AiManager.OpenAI.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(aiM.EmbeddingModel),
	})
	if err != nil {
		logger.Logger.Errorf("Failed to generate embedding: %v", err)
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}

// GenerateResponse generates AI response
func (aiM *OpenAIManager) GetAIResponse(instructions string, userQuery string, previousConversations string, model string, responseSchema any, platform string) (*any, utils.TokenUsage, error) {
	ctx := context.Background()
	if model == "" {
		model = AiManager.TextGenModel
	}
	modelInputText := "\nRequirements: " + userQuery
	if previousConversations != "" {
		modelInputText += "\nPrevious Conversations: " + previousConversations
	}
	schemaJSON, _ := json.Marshal(responseSchema)
	resp, err := aiM.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: instructions,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: modelInputText,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "anna_assistant_schema",
				Schema: json.RawMessage(schemaJSON),
				Strict: true,
			},
		},
		Temperature: 1,
		MaxTokens:   2048,
		TopP:        1,
	})

	if err != nil {
		logger.Logger.Errorf("Failed to generate response: %v", err)
		return nil, utils.TokenUsage{}, fmt.Errorf("something went wrong while generating response")
	}

	// Calculate cost
	usage := utils.TokenUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	cost := utils.CalculatePrice(usage, model)

	logger.Logger.WithFields(map[string]interface{}{
		"level":       "debug",
		"query":       userQuery,
		"tokens-used": resp.Usage.TotalTokens,
		"pricing":     cost,
	}).Info("Generated response")

	// Parse response
	var aiResponse any
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiResponse)
	if err != nil {
		logger.Logger.Errorf("Failed to parse AI response: %v", err)
		return nil, usage, fmt.Errorf("failed to parse response")
	}

	return &aiResponse, usage, nil
}

// GenerateResponse generates AI response
func (aiM *OpenAIManager) GenerateResponseV1(userQuery, previousConversation, retrievedChunks, platform, model string) (*AIResponse, utils.TokenUsage, error) {
	ctx := context.Background()

	if model == "" {
		model = AiManager.TextGenModel
	}

	userContent := fmt.Sprintf("User's query: %s, previous conversation: %s, user's platform: %s, texts to help: %s",
		userQuery, previousConversation, platform, retrievedChunks)

	// Define response schema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"result": map[string]interface{}{
				"type":        "string",
				"description": "The response provided by the assistant.",
			},
			"followUps": map[string]interface{}{
				"type":        "array",
				"description": "An array of follow-up questions for further engagement with the user.",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"updateCycleQueryLayers": map[string]interface{}{
				"type":        "array",
				"description": "An array of layer names if the user query is related to data updates, last updated information, or data freshness for any layer. If the query is not related to these topics, return an empty array.",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required":             []string{"result", "followUps", "updateCycleQueryLayers"},
		"additionalProperties": false,
	}

	schemaJSON, _ := json.Marshal(schema)

	resp, err := aiM.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: aiM.Instructions,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userContent,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "anna_assistant_schema",
				Schema: json.RawMessage(schemaJSON),
				Strict: true,
			},
		},
		Temperature: 1,
		MaxTokens:   2048,
		TopP:        1,
	})

	if err != nil {
		logger.Logger.Errorf("Failed to generate response: %v", err)
		return nil, utils.TokenUsage{}, fmt.Errorf("something went wrong while generating response")
	}

	// Calculate cost
	usage := utils.TokenUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	cost := utils.CalculatePrice(usage, model)

	logger.Logger.WithFields(map[string]interface{}{
		"level":       "debug",
		"query":       userQuery,
		"tokens-used": resp.Usage.TotalTokens,
		"pricing":     cost,
	}).Info("Generated response")

	// Parse response
	var aiResponse AIResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiResponse)
	if err != nil {
		logger.Logger.Errorf("Failed to parse AI response: %v", err)
		return nil, usage, fmt.Errorf("failed to parse response")
	}

	return &aiResponse, usage, nil
}

func (aiM *OpenAIManager) AskModelAndHandleTools(userQuery, previousConversation, platform, model string, registory *tools.ToolRegistry, ctx context.Context) (string, error) {
	userContent := fmt.Sprintf("User's query: %s, previous conversation: %s, user's platform: %s",
		userQuery, previousConversation, platform)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userContent,
		},
	}
	maxNLoops := 8
	count := 1
	for count < maxNLoops {
		count++
		resp, err := aiM.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       aiM.TextGenModel,
			Messages:    messages,
			Functions:   registory.FunctionDefinitions(),
			Temperature: 0,
		})
		if err != nil {
			return "", err
		}
		msg := resp.Choices[0].Message
		// CASE 1: model wants to call a function/tool
		if msg.FunctionCall != nil && msg.FunctionCall.Name != "" {
			fn := msg.FunctionCall.Name
			argsJSON := msg.FunctionCall.Arguments
			// parse args
			params := map[string]any{}
			if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
				return "", fmt.Errorf("failed to parse function args: %w", err)
			}

			// run the tool
			result, err := registory.Execute(ctx, fn, params, nil)
			if err != nil {
				return "", err
			}
			// return tool result back to model as a function response
			resultBytes, _ := json.Marshal(result)
			messages = append(messages,
				msg, // the model's tool call message
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleFunction,
					Name:    fn,
					Content: string(resultBytes),
				},
			)
			continue
		}

		// CASE 2: normal assistant message — final answer
		if msg.Role == openai.ChatMessageRoleAssistant {
			return msg.Content, nil
		}
	}
	if count == maxNLoops {
		// get one last try without tool calls
		msg, err := aiM.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       aiM.TextGenModel,
			Messages:    messages,
			Temperature: 0,
		})
		if err != nil {
			return "", err
		}
		if msg.Choices[0].Message.Role == openai.ChatMessageRoleAssistant {
			return msg.Choices[0].Message.Content, nil
		}
	}

	return "", fmt.Errorf("exceeded max loops without final answer")
}
