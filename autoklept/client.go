package autoklept

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
)

var (
	ErrNon200ResponseCode = errors.New("non-200 response code when fetching HTML")
)

type Client struct {
	deepseek *deepseek.Client
	cfg      *Config
}

type Config struct {
	DeepseekAPIKey  string // Generate and monitor usage at https://platform.deepseek.com/usage.
	DeepseekTimeout time.Duration
}

func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{deepseek: deepseek.NewClient(apiKey)}
	c.cfg = &Config{DeepseekAPIKey: apiKey}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.cfg.DeepseekTimeout = timeout
	}
}

func (c *Client) BuildURLs(ctx context.Context, sourceURLs, sitemapURLs []string) ([]url.URL, error) {
	var urls []url.URL
	for _, uStr := range sourceURLs {
		u, err := url.Parse(uStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing source URL: %w", err)
		}
		urls = append(urls, *u)
	}
	for _, smUrl := range sitemapURLs {
		found, err := ParseSitemapURLs(ctx, smUrl)
		if err != nil {
			return nil, fmt.Errorf("error parsing sitemap URL: %w", err)
		}
		urls = append(urls, found...)
	}
	return urls, nil
}

func (c *Client) NewPromptRequest(ctx context.Context, reqInput *PromptRequestInput) (*PromptRequest, error) {
	in, err := Validate[PromptInputTag](reqInput.InputTag)
	if err != nil {
		return nil, err
	}
	out, err := Validate[PromptOutputTag](reqInput.OutputTag)
	if err != nil {
		return nil, err
	}
	// This does not have the input HTML attached to it yet
	// The PromptRequest might be the same other than that HTML, so we need only make one.
	// This is meant to capture autoklept's best practices for how to query DeepSeek for best extraction.
	ccr := deepseek.ChatCompletionRequest{
		// This actually perform better than the deepseek-reasoner at clean extraction. Hilarious.
		Model:    deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{{Role: constants.ChatMessageRoleSystem, Content: deepseekSystemRole}},
	}
	var nf *ElementNodeFinder
	if reqInput.HTMLFinder != nil {
		nf = &ElementNodeFinder{
			Tag:     reqInput.HTMLFinder.Tag,
			AttrKey: reqInput.HTMLFinder.AttrKey,
			AttrVal: reqInput.HTMLFinder.AttrVal,
		}
	}
	return &PromptRequest{
		prompt:     buildPromptString(in, out),
		systemRole: deepseekSystemRole,
		ccr:        ccr,
		nodeFinder: nf,
	}, nil
}

// ExecPromptFor executes the given prompt request for the parsed content found at `url`.
func (c *Client) ExecPromptFor(ctx context.Context, pr *PromptRequest, u string) (*PromptResponse, error) {
	uParsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap URL: %w", err)
	}
	// Get HTML from URL and parse as desired
	htmlResp, err := httpGet(ctx, uParsed)
	if err != nil {
		return nil, fmt.Errorf("error fetching HTML from URL: %w", err)
	}
	// TODO: fork depending on pr.nodeFinder existing should happen here, probably
	parsedHtml, err := parseHtmlByTag(htmlResp, pr.nodeFinder)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}
	pr.setPromptWithBytes(parsedHtml)
	resp, err := c.deepseek.CreateChatCompletion(ctx, &pr.ccr)
	if err != nil {
		return nil, fmt.Errorf("error querying DeepSeek: %w", err)
	}
	return newPromptResponse(resp), nil
}

// ParseSitemapURLs doesn't require any LLM
func ParseSitemapURLs(ctx context.Context, sitemapURL string) ([]url.URL, error) {
	u, err := url.Parse(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap URL: %w", err)
	}
	sitemapRaw, err := httpGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error getting sitemap from URL: %w", err)
	}
	return extractUrlSet(sitemapRaw)
}

func httpGet(ctx context.Context, u *url.URL) ([]byte, error) {
	req := &http.Request{URL: u, Method: http.MethodGet}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on HTTP request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrNon200ResponseCode
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
