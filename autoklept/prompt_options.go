package autoklept

import "errors"

// This file contains different prompt fragments that can be composed to tell autoklept how you want to parse.
// These prompt options will be tested for quality and iterated on over time, to provide users with the best possible results.
const (
	deepseekSystemRole = `- You are extremely good at parsing HTML.
- You are aware of the pitfalls of tricky characters like backticks.
- It is your top priority not to change the actual user content you extract from any parsed HTML. You must not reword, summarize, or remove anything - you must reproduce the original user language exactly and completely.
- You must preserve the structure of the original content as much as possible.
- If you drop any content from the original blog, my life will be ruined, so please try your best.
- When you output, do not write anything before or after the raw output you formatted from the original content. Omit any backticks markdown open / close blocks.`

	deepseekStdPrompt = "Extract out all actual user content from the following HTML. "
)

var (
	ErrInvalidEnumMember = errors.New("invalid enum member")
)

type PromptInputTag int

const (
	PromptInputAll  PromptInputTag = iota // All
	PromptInputBlog                       // Blog
)

type PromptOutputTag int

const (
	PromptOutputText     PromptOutputTag = iota // Text
	PromptOutputMarkdown                        // Markdown
	PromptOutputSimple                          // Simple
)

type PromptTagText string

const (
	AllInputText       = "The input HTML contains text, images, and video."
	BlogInputText      = "The input HTML contains a blog."
	TextOutputText     = "Output the content as raw text, without any formatting except line breaks."
	MarkdownOutputText = "Output the content as markdown, preserving as much of the original formatting as possible."
	SimpleOutputText   = "Output the content as simplified HTML, where as much site-specific HTML slop has been stripped out, while still preserving as much of the original structure and rendering."
)

// TODO: I feel like I'm just reproducing a relational database right now.
var (
	outputTag2Text = map[PromptOutputTag]string{
		PromptOutputText:     TextOutputText,
		PromptOutputMarkdown: MarkdownOutputText,
		PromptOutputSimple:   SimpleOutputText,
	}
	inputTag2Text = map[PromptInputTag]string{
		PromptInputAll:  AllInputText,
		PromptInputBlog: BlogInputText,
	}
)

func getOutputTagText(tag PromptOutputTag) string {
	if text, ok := outputTag2Text[tag]; ok {
		return text
	}
	panic("coding error, unknown PromptOutputTag") // I mean, it is cause to panic.
}

func getInputTagText(tag PromptInputTag) string {
	if text, ok := inputTag2Text[tag]; ok {
		return text
	}
	panic("coding error, unknown PromptInputTag") // I mean, it is cause to panic.
}
