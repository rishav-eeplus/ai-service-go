package orchestrator

import (
	"ai-service-go/internals/tools"
	"ai-service-go/internals/types"
	"ai-service-go/internals/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	// "strings"

	"github.com/sashabaranov/go-openai"
)

// ReActSystemPrompt encourages the model to follow ReAct (Reasoning + Acting) paradigm
var ReActSystemPrompt = fmt.Sprintf(`You are Anna, a female assistant for EEHORIZON who uses tools and her wit to assist users to understand 
        and use platform features effectively. Engage in witty conversations in respectful manner for added context and assistance.
        You may also be provided with previous conversations in this format: previous conversation: [{"role":"user/assistant","content":"conversation"}],
		## Internal Reasoning Process
		For each step, internally follow this cycle:
		1. **Thought**: Reason about what to do next
		2. **Action**: Call a tool if needed
		3. **Observation**: Analyze the tool result
		Continue this cycle until you have enough information.		
		# Tool Usage Guidelines: %s
		
		## RESTRICTIONS (IMPORTANT - READ CAREFULLY)
		You CANNOT and must NEVER claim to do actions on behalf of the user or the platform, or any actions outside 
		 the scope of your capabilities.
		
		## Constraints
		- Be interactive and engaging while assisting users.
		- If a user greets you like Hi Anna or Hello Anna, simply respond with a polite greeting.
        - If a query falls outside the scope of EEHORIZON, politely apologize, acknowledging the impossibility of helping in a creative way.

		## Clarification Rules (IMPORTANT)
		When you encounter ambiguous situations that require user input, ask for clarification:
		- If a search returns multiple matching items (e.g., multiple layers with similar names like "substations", "transmission_substations", "distribution_substations"), set "needsClarification" to true and list the options.
		- If the user's query is vague or could mean multiple things, ask for clarification.
		- If you need more specific information to proceed, ask for it.
		- When asking for clarification, provide clear options for the user to choose from.
		- 4-5 options are preffered, but if required not more than 8 options should be presented to the user.

		## Final Output Rules
		- Your final response must ONLY contain the direct answer to the user's question.
		- Do NOT include your internal reasoning process (Thought/Action/Observation) in the final output.
		- Do NOT explain what tools you used or how you arrived at the answer.
		- Write as if you're speaking directly to the user—concise, helpful, and to the point.
		- Write clean markdown content. 
		- The "result" field should contain ONLY the answer the user needs, nothing else.
		# Output Format
        - Return all responses in JSON format.
        - **result**: Your response.
        - **needsClarification**: Boolean - Set to true ONLY when you need the user to choose from options or provide more information.
        - **clarificationMessage**: String - The question you want to ask the user (only when needsClarification is true).
        - **options**: Array of strings - The options for the user to choose from (only when needsClarification is true, usual 4-5, max 8 options).
        - **followUps**: An array of questions(maximum 2) formatted as if the user is asking them to the assistant,  
		   also answered using the available tools and user guide data. Leave empty if needsClarification is true.
`, tools.ToolUsageInstructions)

var FinalOutputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"result": map[string]any{
			"type":        "string",
			"description": "The response provided by the assistant.",
		},
		"needsClarification": map[string]any{
			"type":        "boolean",
			"description": "Set to true when the assistant needs the user to choose from options or provide more information.",
		},
		"clarificationMessage": map[string]any{
			"type":        "string",
			"description": "The question to ask the user when clarification is needed.",
		},
		"options": map[string]any{
			"type":        "array",
			"description": "Options for the user to choose from when clarification is needed. Maximum 8 options.",
			"items": map[string]any{
				"type": "string",
			},
		},
		"followUps": map[string]any{
			"type":        "array",
			"description": "An array of follow-up questions for further engagement with the user. The question should be such that it can be future questions asked by the user to get more clarity on their requirements.",
			"items": map[string]any{
				"type": "string",
			},
		},
	},
	"required":             []string{"result", "needsClarification", "clarificationMessage", "options", "followUps"},
	"additionalProperties": false,
}

type PlannerOutput struct {
	Result               string   `json:"result"`
	NeedsClarification   bool     `json:"needsClarification"`
	ClarificationMessage string   `json:"clarificationMessage"`
	Options              []string `json:"options"`
	FollowUps            []string `json:"followUps"`
}

func (o *Orchestrator) PlannerAndToolExecuter(sendMessage func(msg types.StreamMessage) bool, input *ClientRequestType, model string, ctx context.Context, loopsRemaining int) (*PlannerOutput, error) {
	// Use context.Background() if ctx is nil
	if ctx == nil {
		ctx = context.Background()
	}
	if model == "" {
		model = o.AIManager.TextGenModel
	}

	userContent := fmt.Sprintf("User's query: %s, previous conversation: %s, user's platform: %s",
		input.UserQuery, input.PreviousConversation, input.Platform)

	// If this is the last clarification loop, instruct the AI to provide a final response
	if loopsRemaining <= 1 {
		userContent += "\n\n**IMPORTANT: This is your final opportunity to respond. Do NOT ask for clarification. You MUST provide a complete response based on all the context you have gathered so far. If there are multiple options, provide information about ALL of them or make a reasonable choice and explain your reasoning.**"
	}

	// userContent += fmt.Sprintf("\nYou must use these tools: %s. Reasoning for using these tools: %s", strings.Join(extractIntentNames(useful_tools), ", "), reasoningForUsingTool)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: ReActSystemPrompt,
		},
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
	count := 0

	// Create progress ticker for dynamic messages during long waits
	progressTicker := NewProgressTicker()

	for count < maxNLoops {
		count++
		// ReAct Step: Thinking/Reasoning phase
		// Send witty thinking message based on the step
		if count == 1 {
			sendMessage(types.StreamMessage{
				Type:    "info",
				Message: GetThinkingMessage(),
			})
		} else {
			// For subsequent steps, show multi-step progress
			sendMessage(types.StreamMessage{
				Type:    "info",
				Message: GetMultiStepMessage(),
			})
		}

		// Start progress ticker for AI completion (sends messages every 4 seconds)
		progressTicker.Start(ctx, sendMessage, 4*time.Second)
		resp, err := o.AIManager.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Tools:    o.ToolResistory.ToolDefinitions(),
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "anna_assistant_schema",
					Schema: json.RawMessage(jsonSchema),
					Strict: true,
				},
			},
		})

		// Stop the progress ticker after AI completion
		progressTicker.Stop()

		if err != nil {
			return nil, err
		}
		// Log token usage
		utils.Logger.Infof("Token usage - Prompt: %d, Completion: %d, Total: %d",
			resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
		msg := resp.Choices[0].Message
		// CASE 1: model wants to call a tool
		if len(msg.ToolCalls) > 0 {
			// ReAct Step: Action phase

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
				// send witty message that tool is being executed
				wittyToolStart := GetToolStartMessage(fn, o.ToolResistory.GetTool(toolCall.Function.Name).InformationMessage().Start)
				sendMessage(types.StreamMessage{
					Type:    "info",
					Message: wittyToolStart,
				})

				// Start progress ticker for tool execution
				progressTicker.Start(ctx, sendMessage, 3*time.Second)

				// run the tool
				result, err := o.ToolResistory.Execute(ctx, fn, params, sendMessage)

				// Stop the progress ticker after tool execution
				progressTicker.Stop()

				if err != nil {
					// Send error recovery message
					sendMessage(types.StreamMessage{
						Type:    "info",
						Message: GetErrorRecoveryMessage(),
					})
					return nil, err
				}
				// send witty completion message
				wittyToolEnd := GetToolEndMessage(fn, o.ToolResistory.GetTool(toolCall.Function.Name).InformationMessage().End)
				sendMessage(types.StreamMessage{
					Type:    "info",
					Message: wittyToolEnd,
				})
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
			sendMessage(types.StreamMessage{
				Type:    "success",
				Message: GetFinalAnswerMessage(),
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
		sendMessage(types.StreamMessage{
			Type:    "info",
			Message: "🔧 Taking a bit longer than expected... let me wrap this up!",
		})
		// get one last try without tool calls
		msg, err := o.AIManager.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		})
		if err != nil {
			utils.Logger.Errorf("Tool Planning and Execution failed  - Final attempt without tool calls failed: %v", err)
			return nil, err
		}
		// Log token usage for final attempt
		utils.Logger.Infof("Token usage (final attempt) - Prompt: %d, Completion: %d, Total: %d",
			msg.Usage.PromptTokens, msg.Usage.CompletionTokens, msg.Usage.TotalTokens)
		if msg.Choices[0].Message.Role == openai.ChatMessageRoleAssistant {
			sendMessage(types.StreamMessage{
				Type:    "success",
				Message: "💪 Got there in the end! Here's your answer...",
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
