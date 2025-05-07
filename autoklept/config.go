package autoklept

import (
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"os"
	"time"
)

const (
	cfgPrefix = "AUTOKLEPT"
)

var (
	ErrSourceRequired = errors.New("at least one data source is required")
)

type Config struct {
	Client ClientConfig `conf:"help:Config for autoklept client"`
	Source SourceOpts   `conf:"help:Options for where to get raw input data for parsing"`
	Html   HTMLOpts     `conf:"help:Options for how to parse HTML before LLM analysis"`
	Prompt PromptOpts   `conf:"help:Options for how to prompt DeepSeek to parse content"`
}

func ParseConfig() (*Config, error) {
	var cfg Config
	if help, err := conf.Parse(cfgPrefix, &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Printf("%s", help)
			os.Exit(0) // early-exit here is fine, for the sake of encapsulation
		}
		return nil, err
	}
	return &cfg, nil
}

type PromptOpts struct {
	InputContentTag  string `conf:"help:The type of content to extract"`
	OutputContentTag string `conf:"help:The content output format"`
}

type HTMLOpts struct {
}

type ClientConfig struct {
	DeepseekAPIKey string `conf:"required,help:The Deepseek API Key to use for extracting content"`
	// Insanely high timeout - LLM calls can take awhile!
	DeepseekTimeout time.Duration `conf:"default:300s,help:Request timeout"`
}

type SourceOpts struct {
	Urls        []string `conf:"help:URL(s) to fetch"`
	SitemapUrls []string `conf:"help:XML Sitemap(s) to parse for extracting user content"`
}
