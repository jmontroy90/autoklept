package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/jmontroy90/autoklept/autoklept"
)

func main() {
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}
	ctx := context.Background()
	client := autoklept.NewClient(cfg.Client.DeepseekAPIKey, autoklept.WithTimeout(cfg.Client.DeepseekTimeout))
	urls, err := buildURLs(ctx, cfg.Source.Urls, cfg.Source.SitemapUrls)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// TODO: basic sync.WaitGroup for concurrency
	for i, u := range urls {
		req, err := client.NewPromptRequest(ctx, buildPromptRequestInput(cfg))
		if err != nil {
			log.Fatalf("%v", err)
		}
		resp, err := client.ExecPromptFor(ctx, req, u)
		if err != nil {
			log.Fatalf("%v", err)
		}
		// TODO: better file naming, extract it out?
		if err := os.WriteFile(fmt.Sprintf("out/%s-%d.md", cfg.Output.FilePrefix, i), []byte(resp.Content), 0644); err != nil {
			log.Fatalf("%v", err)
		}
	}

}

func buildPromptRequestInput(cfg *Config) *autoklept.PromptRequestInput {
	var htmlFinder *autoklept.ElementNodeFinder
	if nf := cfg.Html.NodeFinder; nf.Tag != "" && nf.AttrKey != "" && nf.AttrVal != "" {
		htmlFinder = &autoklept.ElementNodeFinder{
			Tag:     cfg.Html.NodeFinder.Tag,
			AttrKey: cfg.Html.NodeFinder.AttrKey,
			AttrVal: cfg.Html.NodeFinder.AttrVal,
		}
	}
	// TODO: translation layers are interesting.
	return &autoklept.PromptRequestInput{
		InputTag:   cfg.Prompt.InputContentTag,
		OutputTag:  cfg.Prompt.OutputContentTag,
		HTMLFinder: htmlFinder,
	}
}

func buildURLs(ctx context.Context, sourceURLs, sitemapURLs []string) ([]*url.URL, error) {
	var urls []*url.URL
	for _, uStr := range sourceURLs {
		u, err := url.Parse(uStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing source URL: %w", err)
		}
		urls = append(urls, u)
	}
	for _, smUrl := range sitemapURLs {
		found, err := autoklept.ParseSitemapURLs(ctx, smUrl)
		if err != nil {
			return nil, fmt.Errorf("")
		}
		urls = append(urls, found...)
	}
	return urls, nil
}
