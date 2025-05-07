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
	cfg, err := autoklept.ParseConfig()
	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}
	//ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DeepseekTimeout)
	//defer cancel()
	ctx := context.Background()
	client, err := autoklept.NewAutoKleptClient(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	var urls []*url.URL
	for _, uStr := range cfg.Source.Urls {
		u, err := url.Parse(uStr)
		if err != nil {
			log.Fatalf("%v", err)
		}
		urls = append(urls, u)
	}
	// TODO: refactor into functions
	for _, smUrl := range cfg.Source.SitemapUrls {
		u, err := url.Parse(smUrl)
		if err != nil {
			log.Fatalf("%v", err)
		}
		sitemapRaw, err := client.Get(ctx, u)
		if err != nil {
			log.Fatalf("%v", err)
		}
		us, err := autoklept.ExtractUrlSet([]byte(sitemapRaw))
		if err != nil {
			log.Fatalf("%v", err)
		}
		urls = append(urls, us...)
	}
	// TODO: basic sync.WaitGroup for concurrency
	for i, u := range urls {
		htmlResp, err := client.Get(ctx, u)
		if err != nil {
			log.Fatalf("%v", err)
		}
		// TODO: configurable html pathing
		parsedHtml, err := client.ParseHtmlAt(ctx, htmlResp, "div", "id", "SITE_CONTAINER")
		if err != nil {
			log.Fatalf("%v", err)
		}
		req, err := client.NewPromptRequest(ctx, parsedHtml, cfg.Prompt)
		if err != nil {
			log.Fatalf("%v", err)
		}
		extractResp, err := client.ExecParseRequest(ctx, req)
		if err != nil {
			log.Fatalf("%v", err)
		}
		output := autoklept.ExtractResponseContent(extractResp)
		// TODO: better file naming, extract it out?
		if err := os.WriteFile(fmt.Sprintf("out/%d.md", i), []byte(output), 0644); err != nil {
			log.Fatalf("%v", err)
		}
	}

}
