package tools

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/types"
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type GetHelpSupport struct{}

func (gl *GetHelpSupport) Name() string {
	return "get_help_support"
}
func (gl *GetHelpSupport) Description() string {
	return `This tool is used when unable to answer to user query or user needs any human assistance.`
}

func (gl *GetHelpSupport) Execute(ctx context.Context, params map[string]any, sendMessage func(msg types.StreamMessage) bool) (any, error) {
	supportEmail := config.AppConfig.SupportEmail
	var response string
	response += "You can use 'Support' option available on the platform, which is available at the top right corner of the navigation bar, to raise tickets. Your queries will be addressed promptly."
	response += "For further assistance, please reach out to our support team at " + supportEmail + "."
	return response, nil
}

func (gl *GetHelpSupport) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        gl.Name(),
		Description: gl.Description(),
		Parameters:  nil,
	}
}
func (gl *GetHelpSupport) InformationMessage() struct {
	Start string
	End   string
} {
	return struct {
		Start string
		End   string
	}{
		Start: `Getting support options...`,
		End:   `Support options retrieved.`,
	}
}