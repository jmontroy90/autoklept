package autoklept

import (
	"bytes"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
)

type PromptResponse struct {
	Content          string
	ReasoningContent string
	TokensUsed       int
}

func newPromptResponse(ccr *deepseek.ChatCompletionResponse) *PromptResponse {
	return &PromptResponse{
		Content:          ccr.Choices[0].Message.Content,
		ReasoningContent: ccr.Choices[0].Message.ReasoningContent,
		TokensUsed:       ccr.Usage.TotalTokens,
	}
}

type PromptRequestInput struct {
	InputTag   string
	OutputTag  string
	HTMLFinder *ElementNodeFinder
}

type PromptRequest struct {
	systemRole string
	prompt     string
	nodeFinder *ElementNodeFinder
	ccr        deepseek.ChatCompletionRequest
}

func (pr *PromptRequest) SystemRole() string {
	return pr.systemRole
}

func (pr *PromptRequest) Prompt() string {
	return pr.prompt
}

func (pr *PromptRequest) setPromptWithBytes(bs *bytes.Buffer) {
	p := pr.prompt + "\n" + bs.String()
	pm := deepseek.ChatCompletionMessage{Role: constants.ChatMessageRoleUser, Content: p}
	pr.ccr.Messages = append(pr.ccr.Messages, pm)
}

func buildPromptString(input PromptInputTag, output PromptOutputTag) string {
	return deepseekStdPrompt + "\n" + getInputTagText(input) + "\n" + getOutputTagText(output)
}
