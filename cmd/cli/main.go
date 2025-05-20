package main

import (
	"context"
	"fmt"
	"github.com/jmontroy90/autoklept/autoklept"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slog"
	"log"
	"net/url"
	"os"
	"time"
)

const (
	ExtractCmd         = "extract"
	ExtractAPIKeyFlag  = "deepseek-api-key"
	ExtractTimeoutFlag = "deepseek-timeout"
	ExtractURLFlag     = "url"

	SitemapCmd     = "sitemap"
	SitemapURLFlag = "url"
)

type cmdRunner struct {
	// TODO: Dunno if this is necessary as global state, now that we can initialize this Client in multiple ways.
	logger slog.Logger
}

func main() {
	runner := cmdRunner{}
	runner.exec(context.Background())
}

func (r *cmdRunner) exec(ctx context.Context) {
	cmd := &cli.Command{
		Usage: "Autoklept uses AI to extract out meaningful user content from the Web. Using complex AI - to fight complexity!",
		Commands: []*cli.Command{
			{
				Name: SitemapCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     SitemapURLFlag,
						Aliases:  []string{"u"},
						Usage:    "Sitemap URL",
						Required: true,
					},
				},
				Action: r.execSitemapCmd,
			},
			{
				Name: ExtractCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  ExtractAPIKeyFlag,
						Usage: "Deepseek API key",
						// DEEPSEEK_API_KEY is looked for by Deepseek library, so we intercept it here for cleanliness.
						Sources: cli.EnvVars("AUTOKLEPT_DEEPSEEK_API_KEY"),
					},
					&cli.DurationFlag{
						Name:  ExtractTimeoutFlag,
						Usage: "Deepseek API timeout",
						Value: 300 * time.Second,
					},
					&cli.StringFlag{
						Name:     ExtractURLFlag,
						Usage:    "URL from which to extract content",
						Aliases:  []string{"u"},
						Required: true,
					},
				},
				Action: r.execExtractCmd,
			},
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

func (r *cmdRunner) execSitemapCmd(ctx context.Context, cmd *cli.Command) error {
	sm := cmd.String(SitemapURLFlag)
	us, err := autoklept.ParseSitemapURLs(ctx, sm)
	if err != nil {
		return err
	}
	for _, u := range us {
		fmt.Printf("%v\n", u)
	}
	return nil
}

func (r *cmdRunner) execExtractCmd(ctx context.Context, cmd *cli.Command) error {
	key, timeout := cmd.String(ExtractAPIKeyFlag), cmd.Duration(ExtractTimeoutFlag)
	if key == "" {
		return fmt.Errorf("missing required Deepseek API Key")
	}
	// Set client to actually have DeepSeek (TODO: is this silly?)
	c := autoklept.NewClient(key, autoklept.WithTimeout(timeout))
	u := cmd.String(ExtractURLFlag)
	target, err := url.Parse(u)
	if err != nil {
		return err
	}
	// TODO: configurable + HTMLFinders
	// TODO: no elementnodefinder, use CSS selectors that user can specify, you parse, does same shit!
	// .div[id="SITE_CONTAINER"]
	finder := &autoklept.ElementNodeFinder{Tag: "div", AttrKey: "id", AttrVal: "SITE_CONTAINER"}
	pri := autoklept.PromptRequestInput{InputTag: "blog", OutputTag: "markdown", HTMLFinder: finder}
	pr, err := c.NewPromptRequest(ctx, &pri)
	if err != nil {
		return err
	}
	prsp, err := c.ExecPromptFor(ctx, pr, target.String())
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", prsp.Content)
	return nil
}
