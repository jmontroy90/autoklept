package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jmontroy90/autoklept/autoklept"
)

func main() {
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	client, err := autoklept.NewAutoKleptClient(cfg.ToAutokleptConfig(), logger)
	if err != nil {
		log.Fatalf("%v", err)
	}
	urls, err := client.BuildURLs(ctx, cfg.Source.Urls, cfg.Source.SitemapUrls)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// TODO: basic sync.WaitGroup for concurrency
	for i, url := range urls {
		req, err := client.NewPromptRequest(ctx, buildPromptRequestInput(cfg))
		if err != nil {
			log.Fatalf("%v", err)
		}
		resp, err := client.ExecPromptFor(ctx, req, url)
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
