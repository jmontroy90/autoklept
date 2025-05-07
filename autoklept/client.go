package autoklept

import "C"
import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
)

var (
	ErrNon200ResponseCode = errors.New("non-200 response code when fetching HTML")
)

type Client struct {
	logger     *slog.Logger
	cfg        *Config
	deepseek   *deepseek.Client
	httpClient *http.Client
}

type Config struct {
	DeepseekAPIKey  string // Generate and monitor usage at https://platform.deepseek.com/usage.
	DeepseekTimeout time.Duration
}

func NewAutoKleptClient(cfg *Config, logger *slog.Logger) (*Client, error) {
	return &Client{
		cfg:        cfg,
		logger:     logger,
		deepseek:   deepseek.NewClient(cfg.DeepseekAPIKey),
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) BuildURLs(ctx context.Context, sourceURLs, sitemapURLs []string) ([]*url.URL, error) {
	var urls []*url.URL
	for _, uStr := range sourceURLs {
		u, err := url.Parse(uStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing source URL: %w", err)
		}
		urls = append(urls, u)
	}
	for _, smUrl := range sitemapURLs {
		found, err := c.parseSitemapURL(ctx, smUrl)
		if err != nil {
			// TODO: I know this is opinionated but I don't to fail the whole thing...
			c.logger.With("sitemapURL", smUrl, "error", err).Error("failed to parse sitemap; continuing processing")
			continue
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
	ccr := &deepseek.ChatCompletionRequest{
		// This actually perform better than the deepseek-reasoner at clean extraction. Hilarious.
		Model:    deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{{Role: constants.ChatMessageRoleSystem, Content: deepseekSystemRole}},
	}
	sys := strings.Builder{}
	sys.WriteString(deepseekSystemRole)
	return &PromptRequest{
		prompt:     buildPromptString(in, out),
		systemRole: sys,
		ccr:        ccr,
		nodeFinder: reqInput.HTMLFinder,
	}, nil
}

// ExecPromptFor executes the given prompt request for the parsed content found at `url`.
func (c *Client) ExecPromptFor(ctx context.Context, pr *PromptRequest, url *url.URL) (*PromptResponse, error) {
	// Get HTML from URL and parse as desired
	htmlResp, err := c.get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error fetching HTML from URL: %w", err)
	}
	// TODO: fork depending on pr.nodeFinder existing should happen here, probably
	parsedHtml, err := parseHtmlByTag(htmlResp, pr.nodeFinder)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}
	pr.setPromptWithBytes(parsedHtml)
	resp, err := c.deepseek.CreateChatCompletion(ctx, pr.ccr)
	if err != nil {
		return nil, fmt.Errorf("error querying DeepSeek: %w", err)
	}
	return newPromptResponse(resp), nil
}

func (c *Client) get(ctx context.Context, u *url.URL) (string, error) {
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

func (c *Client) parseSitemapURL(ctx context.Context, sitemapURL string) ([]*url.URL, error) {
	u, err := url.Parse(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap URL: %w", err)
	}
	sitemapRaw, err := c.get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error getting sitemap from URL: %w", err)
	}
	return extractUrlSet([]byte(sitemapRaw))
}
