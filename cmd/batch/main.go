package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/jmontroy90/autoklept/autoklept"
	"github.com/pelletier/go-toml"
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
	if cfg.NumJobs == 1 {
		if err := processSequential(ctx, client, *cfg, urls); err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		if err := processParallel(ctx, client, *cfg, urls); err != nil {
			log.Fatalf("%v", err)
		}
	}
}

func processSequential(ctx context.Context, client *autoklept.Client, cfg Config, urls []string) error {
	for _, u := range urls {
		if err := processURL(ctx, client, cfg, u); err != nil {
			return fmt.Errorf("error processing url '%s': %w", u, err)
		}
	}
	return nil
}

func processParallel(ctx context.Context, client *autoklept.Client, cfg Config, urls []string) error {
	var wg sync.WaitGroup
	uChan := make(chan string)
	for i := 0; i < cfg.NumJobs; i++ {
		wg.Add(1)
		go func(cu <-chan string, w *sync.WaitGroup) {
			defer w.Done()
			for u := range cu {
				if err := processURL(ctx, client, cfg, u); err != nil {
					// TODO: errgroup?
					log.Printf("FAILED PROCESSING URL: '%s': %v\n", u, err)
				}
			}
		}(uChan, &wg)
	}
	for _, u := range urls {
		uChan <- u
	}
	close(uChan)
	wg.Wait()
	return nil
}

func processURL(ctx context.Context, client *autoklept.Client, cfg Config, u string) error {
	req, err := client.NewPromptRequest(ctx, buildPromptRequestInput(cfg))
	if err != nil {
		return err
	}
	resp, err := client.ExecPromptFor(ctx, req, u)
	if err != nil {
		return err
	}
	// TODO: there's like a whole "parsers" thingy implied by this lol
	// TODO: Probably need to try to strip out bad output formatting if the LLM decides to go rogue over time,
	// but there's only so much to really try to do here.
	outFile := fmt.Sprintf("%s-%s.md", cfg.Output.FilePrefix, hashPrefix(u, 5))
	if strings.ToLower(cfg.Prompt.OutputContentTag) == strings.ToLower(autoklept.PromptOutputHugo.String()) {
		fm, err := parseTOMLFrontMatter(resp.Content)
		if err != nil {
			return err
		}
		outFile = fmt.Sprintf("%s.md", cleanTitle(fm.Title))
	}
	if err := os.WriteFile(fmt.Sprintf("out/%s", outFile), []byte(resp.Content), 0644); err != nil {
		return err
	}
	return nil
}

func cleanTitle(t string) string {
	t = strings.ToLower(t)
	t = strings.ReplaceAll(t, " ", "-")
	t = strings.ReplaceAll(t, ":", "")
	return t
}

func buildPromptRequestInput(cfg Config) *autoklept.PromptRequestInput {
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

func buildURLs(ctx context.Context, sourceURLs, sitemapURLs []string) ([]string, error) {
	var urls []string
	for _, uStr := range sourceURLs {
		u, err := url.Parse(uStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing source URL: %w", err)
		}
		urls = append(urls, u.String())
	}
	for _, smUrl := range sitemapURLs {
		found, err := autoklept.ParseSitemapURLs(ctx, smUrl)
		if err != nil {
			return nil, fmt.Errorf("error parsing sitemap URLs: %w", err)
		}
		for _, f := range found {
			urls = append(urls, f.String())
		}
	}
	return urls, nil
}

func hashPrefix(s string, prefixLen int) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])[:prefixLen]
}

type FrontMatter struct {
	Title  string
	Author string
	Date   string
	Tags   []string
}

func parseTOMLFrontMatter(content string) (*FrontMatter, error) {
	if !strings.HasPrefix(content, "+++\n") {
		return nil, fmt.Errorf("no TOML front matter found")
	}
	parts := strings.SplitN(content, "+++\n", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid TOML front matter format")
	}
	var fm FrontMatter
	if err := toml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}
