package orchestrator

import (
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var FinalOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"result": map[string]interface{}{
			"type":        "string",
			"description": "The response provided by the assistant.",
		},
		"followUps": map[string]interface{}{
			"type":        "array",
			"description": "An array of follow-up questions for further engagement with the user. The question should be such that it can be future questions asked by the user to get more clarity on their requirements.",
			"items": map[string]interface{}{
				"type": "string",
			},
		},
	},
	"required":             []string{"result", "followUps"},
	"additionalProperties": false,
}

type PlannerOutput struct {
	Result    string   `json:"result"`
	FollowUps []string `json:"followUps"`
}

func (o *Orchestrator) PlannerAndToolExecuter(sendMessage func(msg StreamMessage) bool, input *ClientRequestType, useful_tools []tools.Tool, model string, ctx context.Context) (*PlannerOutput, error) {
	// Use context.Background() if ctx is nil
	if ctx == nil {
		ctx = context.Background()
	}
	if model == "" {
		model = o.AIManager.TextGenModel
	}

	userContent := fmt.Sprintf("User's query: %s, previous conversation: %s, user's platform: %s",
		input.UserQuery, input.PreviousConversation, input.Platform)
	userContent += fmt.Sprintf("\nYou must use this tool: %s", strings.Join(extractIntentNames(useful_tools), ", "))
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userContent,
		},
	}
	jsonSchema, _ := json.Marshal(FinalOutputSchema)
	if model == "" {
		model = o.AIManager.TextGenModel
	}
	maxNLoops := 8
	count := 1
	sendMessage(StreamMessage{
		Type:    "info",
		Message: "Starting planning and tool execution...",
	})
	for count < maxNLoops {
		count++
		resp, err := o.AIManager.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Tools:       o.ToolResistory.ToolDefinitions(),
			Temperature: 0,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "anna_assistant_schema",
					Schema: json.RawMessage(jsonSchema),
					Strict: true,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		msg := resp.Choices[0].Message
		// CASE 1: model wants to call a tool
		if len(msg.ToolCalls) > 0 {
			sendMessage(StreamMessage{
				Type:    "info",
				Message: fmt.Sprintf("Model requested to call tool: %s", msg.ToolCalls[0].Function.Name),
			})

			// Append the assistant's message with tool calls
			messages = append(messages, msg)
			// Process each tool call
			for _, toolCall := range msg.ToolCalls {
				fn := toolCall.Function.Name
				argsJSON := toolCall.Function.Arguments

				// parse args
				params := map[string]any{}
				if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
					utils.Logger.Errorf("Tool Planning and Execution failed - Failed to parse function args: %v", err)
					return nil, fmt.Errorf("failed to parse function args: %w", err)
				}

				// run the tool
				result, err := o.ToolResistory.Execute(ctx, fn, params)
				if err != nil {
					return nil, err
				}

				// return tool result back to model as a tool response
				resultBytes, _ := json.Marshal(result)
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(resultBytes),
					ToolCallID: toolCall.ID,
				})
			}
			continue
		}
		// CASE 2: normal assistant message — final answer
		if msg.Role == openai.ChatMessageRoleAssistant {
			sendMessage(StreamMessage{
				Type:    "info",
				Message: "Model provided final answer.",
			})
			var output PlannerOutput
			if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
				utils.Logger.Errorf("Tool Planning and Execution failed  - Failed to parse assistant output: %v", err)
				return nil, fmt.Errorf("failed to parse assistant output: %w", err)
			}
			return &output, nil
		}
	}
	if count == maxNLoops {
		// Tool Planning and Execution failed
		sendMessage(StreamMessage{
			Type:    "error",
			Message: "Tool Planning and Execution exceeded max loops, making final attempt without tool calls...",
		})
		// get one last try without tool calls
		msg, err := o.AIManager.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0,
		})
		if err != nil {
			utils.Logger.Errorf("Tool Planning and Execution failed  - Final attempt without tool calls failed: %v", err)
			return nil, err
		}
		if msg.Choices[0].Message.Role == openai.ChatMessageRoleAssistant {
			sendMessage(StreamMessage{
				Type:    "info",
				Message: "Model provided final answer on last attempt without tool calls.",
			})
			var output PlannerOutput
			if err := json.Unmarshal([]byte(msg.Choices[0].Message.Content), &output); err != nil {
				return nil, fmt.Errorf("failed to parse assistant output: %w", err)
			}
			return &output, nil
		}
	}
	return nil, fmt.Errorf("exceeded max loops without final answer")
}
