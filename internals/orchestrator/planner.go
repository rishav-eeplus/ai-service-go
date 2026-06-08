package orchestrator

import (
	"ai-service-go/internals/chats_db"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	// "strings"
	"github.com/sashabaranov/go-openai"
)

var chatbotName = "Anna"
var companyName = "EEHORIZON"

// ReActSystemPrompt encourages the model to follow ReAct (Reasoning + Acting) paradigm
var ReActSystemPrompt = fmt.Sprintf(`You are %s, a concise %s assistant. User's interact with you through a TEXT-ONLY interface. There is no image upload, screenshot, or visual input capability.
		Provide direct, useful answers without unnecessary elaboration.
		Previous conversations may be provided as: [{"role":"user/assistant","content":"conversation"}]
		## Internal Process (Do NOT show to user)
		1. **Thought**: Reason about what to do next
		2. **Action**: Call a tool if needed
		3. **Observation**: Analyze the tool result

		# Tool Usage Guidelines: %s

		## Core Rules
		- Be brief and to the point. Avoid verbose explanations.
		- For greetings, farewells, gratitude (Hi, Hello, Goodbye, Thanks), respond with a short greeting or farewell or acknowledgment.
		- For out-of-scope queries, politely decline in one sentence.
		- NEVER claim to perform actions beyond your capabilities — only those defined in the tools below.
		- If a user request falls outside this list → decline politely in one sentence, do not improvise a workaround.
		- If not having enough information to answer user's question, do not make assumptions. Instead handle it gracefully, and ask for clarification if needed (see Clarification section below). 
		- When providing information about a layer, always explain how to reach it using the "locate a layer" tool.

		## Hard Limits — What Anna CANNOT Do
		These are strictly off-limits. NEVER suggest, imply, or offer these capabilities — not even as follow-up questions:
		- CANNOT accept, view, or analyze images, screenshots, files, or any attachments
		- CANNOT enable, disable, toggle, or interact with layers on the map
		- CANNOT perform any UI actions, clicks, or navigation on behalf of the user
		- CANNOT save, export, download, or bookmark data or layers for the user
		- CANNOT modify user settings, preferences, or account details
		- CANNOT provide real-time data, live prices, or current market conditions beyond what tools return
		- CANNOT execute code, run queries, or process external data sources

		## When You Cannot Answer Confidently
		If a question cannot be answered confidently from available tools or platform knowledge:
		- Do NOT guess or fabricate an answer.
		- Direct the user to contact EEHORIZON customer support or their Key Account Manager.
		- For platform/data issues, suggest they **raise a support ticket**.

		## Clarification (Only when necessary)
		Ask for clarification ONLY when:
		- Multiple similar matches exist (eg. user said substations but multiple available substation layers - "ercot substations", "pjm substations")
		- Query is ambiguous and cannot be answered without user input
		- To narrow down options when there are too many possibilities

		When clarifying:
		- Provide 4-5 options (max 8)
		- Set "needsClarification" to true

		## Response Style (CRITICAL)
		- Be direct and actionable
		- Skip pleasantries and filler phrases
		- Do NOT explain your reasoning process to the user
		- Do NOT mention which tools you used
		- Do NOT add "I hope this helps" or similar closing statements
		- Use clean markdown formatting (eg - bold, italics, lists)
		- AVOID large headings (## H2, # H1) - use **bold text** instead for emphasis
		- Focus on what the user needs to know, nothing more

		# Output Format
		- **result**: Direct answer only. No fluff, no process explanation.
		- **needsClarification**: Boolean - true only when user input is required.
		- **clarificationMessage**: Question for user (only when needsClarification is true).
		- **options**: Array of choices (only when needsClarification is true, 4-5 preferred, max 8).
		- **followUps**: Max 2 relevant questions phrased as if asked BY the user (e.g., "How do I...?" not "You can..."). Empty if needsClarification is true. ONLY suggest follow-ups that are answerable using available tools or platform knowledge — never suggest questions that would require capabilities listed in Hard Limits.
`, chatbotName, companyName, tools.ToolUsageInstructions)

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
			"description": "An array of follow-up questions phrased as if asked BY the user (not suggestions TO the user). These should be natural questions a user might ask next to explore the topic further. Maximum 2 questions.",
			"items": map[string]any{
				"type": "string",
			},
		},
	},
	"required":             []string{"result", "needsClarification", "clarificationMessage", "options", "followUps"},
	"additionalProperties": false,
}

type PlannerOutput struct {
	ID                   string   `json:"id"`
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

	// Create and start progress ticker once for the entire execution
	progressTicker := NewProgressTicker()
	progressTicker.Start(ctx, sendMessage, time.Duration(tickerTimeInterval)*time.Second)
	defer progressTicker.Stop()

	totalInputToken, totalOutputToken := 0, 0
	for count < maxNLoops {
		count++
		// ReAct Step: Thinking/Reasoning phase
		// Send witty thinking message based on the step
		if count == 1 {
			if !sendMessage(types.StreamMessage{
				Type:    "info",
				Message: GetThinkingMessage(),
			}) {
				return nil, fmt.Errorf("failed to send thinking message")
			}
		} else {
			// For subsequent steps, show multi-step progress
			if !sendMessage(types.StreamMessage{
				Type:    "info",
				Message: GetMultiStepMessage(),
			}) {
				return nil, fmt.Errorf("failed to send multi-step message")
			}
		}

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

		if err != nil {
			return nil, err
		}
		// Log token usage
		totalInputToken += resp.Usage.PromptTokens
		totalOutputToken += resp.Usage.CompletionTokens
		msg := resp.Choices[0].Message
		// CASE 1: model wants to call a tool
		if len(msg.ToolCalls) > 0 {
			fmt.Println("Tool call detected:", msg.ToolCalls[0].Function.Name)
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
					logger.Logger.Errorf("Tool Planning and Execution failed - Failed to parse function args: %v", err)
					return nil, fmt.Errorf("failed to parse function args: %w", err)
				}
				// send witty message that tool is being executed
				wittyToolStart := GetToolStartMessage(fn, o.ToolResistory.GetTool(toolCall.Function.Name).InformationMessage().Start)
				if !sendMessage(types.StreamMessage{
					Type:    "info",
					Message: wittyToolStart,
				}) {
					return nil, fmt.Errorf("failed to send tool start message")
				}

				// run the tool
				result, err := o.ToolResistory.Execute(ctx, fn, params, sendMessage)

				if err != nil {
					// Send error recovery message
					if !sendMessage(types.StreamMessage{
						Type:    "info",
						Message: GetErrorRecoveryMessage(),
					}) {
						return nil, fmt.Errorf("failed to send error recovery message")
					}
					return nil, err
				}
				// send witty completion message
				wittyToolEnd := GetToolEndMessage(fn, o.ToolResistory.GetTool(toolCall.Function.Name).InformationMessage().End)
				if !sendMessage(types.StreamMessage{
					Type:    "info",
					Message: wittyToolEnd,
				}) {
					return nil, fmt.Errorf("failed to send tool end message")
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
			// Calculate cost and duration
			chatForDB := &types.AnnaChatType{
				UserName:     input.UserName,
				Query:        input.UserQuery,
				Response:     msg.Content,
				Model:        model,
				Feedback:     0,
				InputTokens:  totalInputToken,
				OutputTokens: totalOutputToken,
			}
			chatID, err := AddChatToDB(ctx, chatForDB, o.ChatDB)
			if err != nil {
				logger.Logger.Errorf("Failed to add chat to DB: %v", err)
			}

			if !sendMessage(types.StreamMessage{
				Type:    "success",
				Message: GetFinalAnswerMessage(),
			}) {
				return nil, fmt.Errorf("failed to send final answer message")
			}
			var output PlannerOutput
			output.ID = chatID
			if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
				logger.Logger.Errorf("Tool Planning and Execution failed  - Failed to parse assistant output: %v", err)
				return nil, fmt.Errorf("failed to parse assistant output: %w", err)
			}
			return &output, nil
		}
	}
	if count == maxNLoops {
		// Tool Planning and Execution failed
		if !sendMessage(types.StreamMessage{
			Type:    "info",
			Message: "🔧 Taking a bit longer than expected... let me wrap this up!",
		}) {
			return nil, fmt.Errorf("failed to send info message")
		}
		// get one last try without tool calls
		msg, err := o.AIManager.OpenAI.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		})
		if err != nil {
			logger.Logger.Errorf("Tool Planning and Execution failed  - Final attempt without tool calls failed: %v", err)
			return nil, err
		}
		// Log token usage for final attempt
		totalInputToken += msg.Usage.PromptTokens
		totalOutputToken += msg.Usage.CompletionTokens

		chatForDB := &types.AnnaChatType{
			UserName:     input.UserName,
			Query:        input.UserQuery,
			Response:     msg.Choices[0].Message.Content,
			Model:        model,
			Feedback:     0,
			InputTokens:  totalInputToken,
			OutputTokens: totalOutputToken,
		}

		chatID, err := AddChatToDB(ctx, chatForDB, o.ChatDB)
		if err != nil {
			logger.Logger.Errorf("Failed to add chat to DB: %v", err)
		}
		if msg.Choices[0].Message.Role == openai.ChatMessageRoleAssistant {
			if !sendMessage(types.StreamMessage{
				Type:    "success",
				Message: "💪 Got there in the end! Here's your answer...",
			}) {
				return nil, fmt.Errorf("failed to send success message")
			}
			var output PlannerOutput
			output.ID = chatID
			if err := json.Unmarshal([]byte(msg.Choices[0].Message.Content), &output); err != nil {
				return nil, fmt.Errorf("failed to parse assistant output: %w", err)
			}
			return &output, nil
		}
	}
	logger.Logger.Errorf("Tool Planning and Execution failed  - Exceeded max loops without final answer")
	return nil, fmt.Errorf("exceeded max loops without final answer")
}

func AddChatToDB(ctx context.Context, dbChat *types.AnnaChatType, db *chats_db.ChatDB) (string, error) {
	return db.AddChat(
		ctx, *dbChat)
}
