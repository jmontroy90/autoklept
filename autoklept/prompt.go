//go:generate stringer -type=PromptOutputTag,PromptInputTag -linecomment -output prompt_tag_string.go
package autoklept

import (
	"bytes"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"strings"
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
	systemRole strings.Builder
	prompt     strings.Builder
	nodeFinder *ElementNodeFinder
	ccr        *deepseek.ChatCompletionRequest
}

func (pr *PromptRequest) SystemRole() string {
	return pr.systemRole.String()
}

func (pr *PromptRequest) Prompt() string {
	return pr.prompt.String()
}

func (pr *PromptRequest) setPromptWithBytes(bs *bytes.Buffer) {
	p := strings.Builder{}
	p.WriteString(pr.prompt.String())
	p.WriteString("\n")
	p.WriteString(bs.String())
	pm := deepseek.ChatCompletionMessage{Role: constants.ChatMessageRoleUser, Content: p.String()}
	pr.ccr.Messages = append(pr.ccr.Messages, pm)
}

func buildPromptString(input PromptInputTag, output PromptOutputTag) strings.Builder {
	var sb strings.Builder
	_, _ = sb.WriteString(deepseekStdPrompt)
	sb.WriteString("\n")
	_, _ = sb.WriteString(getInputTagText(input))
	sb.WriteString("\n")
	_, _ = sb.WriteString(getOutputTagText(output))
	return sb
}
