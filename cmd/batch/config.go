package main

import (
	"errors"
	"fmt"
	"github.com/ardanlabs/conf/v3"
	"github.com/jmontroy90/autoklept/autoklept"
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
	Client  ClientConfig `conf:"help:Config for autoklept client"`
	Source  SourceOpts   `conf:"help:Options for where to get raw input data for parsing"`
	Html    HTMLOpts     `conf:"help:Options for how to parse HTML before LLM analysis"`
	Prompt  PromptOpts   `conf:"help:Options for how to prompt DeepSeek to parse content"`
	Output  OutputConfig `conf:"help:How to output and save parsed content"`
	NumJobs int          `conf:"default:1,flag:jobs,short:j,help:Number of parallel autoklept jobs to run"`
}

func (c Config) ToAutokleptConfig() *autoklept.Config {
	return &autoklept.Config{
		DeepseekAPIKey:  c.Client.DeepseekAPIKey,
		DeepseekTimeout: c.Client.DeepseekTimeout,
	}
}

type ClientConfig struct {
	// Generate and monitor usage at https://platform.deepseek.com/usage.
	DeepseekAPIKey string `conf:"required,help:The Deepseek API Key to use for extracting content"`
	// Insanely high timeout - LLM calls can take awhile!
	DeepseekTimeout time.Duration `conf:"default:300s,help:Request timeout"`
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
	InputContentTag  string `conf:"required,help:The type of content to extract"`
	OutputContentTag string `conf:"required,help:The content output format"`
}

type HTMLOpts struct {
	NodeFinder NodeFinder `conf:"help:How to access a subtree of your input content's HTML for parsing"`
}

type NodeFinder struct {
	Tag     string `conf:"help:The HTML ElementNode tag to look for"`
	AttrKey string `conf:"help:The HTML attribute key for the above Tag type"`
	AttrVal string `conf:"help:The HTML attribute value for the above Tag type and AttrKey"`
}

type SourceOpts struct {
	Urls        []string `conf:"help:URL(s) to fetch"`
	SitemapUrls []string `conf:"help:XML Sitemap(s) to parse for extracting user content"`
}

type OutputConfig struct {
	FilePrefix string `conf:"default:autoklept,help:The file name prefix for autoklept parsed output content"`
}
