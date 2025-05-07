package autoklept

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
)

var (
	ErrTooManyResponseChoices = errors.New("found more than one choice in the response")
	ErrNon200ResponseCode     = errors.New("non-200 response code when fetching HTML")
)

type AutokleptClient struct {
	cfg        *Config
	deepseek   *deepseek.Client
	httpClient *http.Client
}

func NewAutoKleptClient(cfg *Config) (*AutokleptClient, error) {
	return &AutokleptClient{
		cfg:        cfg,
		deepseek:   deepseek.NewClient(cfg.Client.DeepseekAPIKey),
		httpClient: http.DefaultClient,
	}, nil
}

func (c *AutokleptClient) Get(ctx context.Context, u *url.URL) (string, error) {
	req := &http.Request{Method: http.MethodGet, URL: u}
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error on HTTP request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", ErrNon200ResponseCode
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (c *AutokleptClient) NewPromptRequest(ctx context.Context, renderedHTML io.Reader, pCfg PromptOpts) (*deepseek.ChatCompletionRequest, error) {
	in, err := Validate[PromptInputTag](pCfg.InputContentTag)
	if err != nil {
		return nil, err
	}
	out, err := Validate[PromptOutputTag](pCfg.OutputContentTag)
	if err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(renderedHTML)
	if err != nil {
		return nil, err
	}
	prompt := buildPromptString(in, out, string(bs))
	return &deepseek.ChatCompletionRequest{
		Model: deepseek.DeepSeekChat, // This actually perform better than the deepseek-reasoner at clean extraction. Hilarious.
		Messages: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleSystem, Content: DeepseekSystemRole},
			{Role: constants.ChatMessageRoleUser, Content: prompt},
		},
	}, nil
}

func (c *AutokleptClient) ExecParseRequest(ctx context.Context, req *deepseek.ChatCompletionRequest) (*deepseek.ChatCompletionResponse, error) {
	return c.deepseek.CreateChatCompletion(ctx, req)
}

func (c *AutokleptClient) ParseHtmlAt(ctx context.Context, htmlBody string, tag, attrKey, attrVal string) (io.Reader, error) {
	return parseHtmlByTag(htmlBody, ElementNodeFinder{Tag: tag, AttrKey: attrKey, AttrVal: attrVal})
}

// TODO: Handle multiple choices
func ExtractResponseContent(resp *deepseek.ChatCompletionResponse) string {
	return resp.Choices[0].Message.Content
}
