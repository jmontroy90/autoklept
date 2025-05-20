//go:generate stringer -type=PromptOutputTag,PromptInputTag -linecomment -output prompt_tag_string.go
package autoklept

import "errors"

// This file contains different prompt fragments that can be composed to tell autoklept how you want to parse.
// These prompt options will be tested for quality and iterated on over time, to provide users with the best possible results.
const (
	deepseekSystemRole = "- You are extremely good at parsing HTML.\n" +
		"- You are aware of the pitfalls of tricky characters like backticks.\n" +
		"- It is your top priority not to change the actual user content you extract from any parsed HTML. You must not reword, summarize, or remove anything - you must reproduce the original user language exactly and completely.\n" +
		"- You must preserve the structure of the original content as much as possible.\n" +
		"- If you drop any content from the original blog, my life will be ruined, so please try your best.\n" +
		"- When you output, do not write anything before or after the raw output you formatted from the original content. DO NOT output any backtick open / close blocks, like ```markdown\n<content here...>\n``` or ```toml\n<content here...>\n```."

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
	PromptOutputHugo                            // Hugo
)

type PromptTagText string

const (
	AllInputText       = "The input HTML contains text, images, and video."
	BlogInputText      = "The input HTML contains a blog."
	TextOutputText     = "Output the content as raw text, without any formatting except line breaks."
	MarkdownOutputText = "Output the content as markdown, preserving as much of the original formatting as possible."
	HugoOutputText     = "Output the content as a Hugo-compatible blog markdown document, preserving as much of the original formatting as possible.\n" +
		"Create the markdown document with a TOML-formatted front matter section with the following fields:\n" +
		"- \"date\", which is populated with the input blog's creation or posting date\n" +
		"- \"draft\", which is hard-coded to true\n" +
		"- \"tags\", which is populated as a TOML list of tags on the original blog\n" +
		"- \"title\", which is populated from the blog's title\n" +
		"\nBe sure to strip out the original title and tags from the text of the input blog, since they are captured by the front matter metadata.\n"

	SimpleOutputText = "Output the content as simplified HTML, where as much site-specific HTML slop has been stripped out, while still preserving as much of the original structure and rendering."
)

// TODO: I feel like I'm just reproducing a relational database right now.
var (
	outputTag2Text = map[PromptOutputTag]string{
		PromptOutputText:     TextOutputText,
		PromptOutputMarkdown: MarkdownOutputText,
		PromptOutputSimple:   SimpleOutputText,
		PromptOutputHugo:     HugoOutputText,
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
	panic("coding error, unknown PromptInputTag")
}
