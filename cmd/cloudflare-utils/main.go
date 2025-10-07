package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

var (
	version       = "DEV"
	date          = "unknown"
	APIClient     *cloudflare.API
	logger        = logrus.New()
	ctx           = context.Background()
	startTime     = time.Now()
	versionString = fmt.Sprintf("%s (built %s)", version, date)
	accountRC     *cloudflare.ResourceContainer
	zoneRC        *cloudflare.ResourceContainer
)

func buildApp() *cli.Command {
	if buildInfo, available := debug.ReadBuildInfo(); available {
		versionString = fmt.Sprintf("%s (built %s with %s)", version, date, buildInfo.GoVersion)
	}
	app := &cli.Command{
		Name:    "cloudflare-utils",
		Usage:   "Program for quick cloudflare utils",
		Version: versionString,
		Suggest: true,
		Authors: []any{
			&mail.Address{
				Name:    "Cyb3r-Jak3",
				Address: "git@cyberjake.xyz",
			},
		},
		Before: setup,
		Commands: []*cli.Command{
			buildDNSCleanerCommand(),
			buildDNSPurgeCommand(),
			buildPruneDeploymentsCommand(),
			buildPurgeDeploymentsCommand(),
			buildGenerateDocsCommand(),
			buildTunnelVersionCommand(),
			buildListSyncCommand(),
			buildCacheCleanerCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    apiTokenFlag,
				Usage:   "A scoped API token (preferred)",
				Sources: cli.EnvVars("CLOUDFLARE_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    apiEmailFlag,
				Usage:   "Cloudflare API email (legacy)",
				Sources: cli.EnvVars("CLOUDFLARE_API_EMAIL"),
			},
			&cli.StringFlag{
				Name:    apiKeyFlag,
				Usage:   "Cloudflare Global API key (legacy)",
				Sources: cli.EnvVars("CLOUDFLARE_API_KEY"),
			},
			&cli.StringFlag{
				Name:    zoneNameFlag,
				Usage:   "Domain name of your zone",
				Sources: cli.EnvVars("CLOUDFLARE_ZONE_NAME"),
			},
			&cli.StringFlag{
				Name:    zoneIDFlag,
				Usage:   "Zone ID of your zone",
				Sources: cli.EnvVars("CLOUDFLARE_ZONE_ID"),
			},
			&cli.StringFlag{
				Name:    accountIDFlag,
				Usage:   "Account ID",
				Sources: cli.EnvVars("CLOUDFLARE_ACCOUNT_ID"),
			},
			&cli.FloatFlag{
				Name:   rateLimitFlag,
				Usage:  "Rate limit for API calls.\nDefault is 4 which matches the Cloudflare API limit of 1200 calls per 5 minutes",
				Value:  4,
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose logging",
				Aliases: []string{"V"},
				Sources: cli.EnvVars("LOG_LEVEL_VERBOSE"),
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug logging",
				Aliases: []string{"d"},
				Sources: cli.EnvVars("LOG_LEVEL_DEBUG"),
			},
			&cli.BoolFlag{
				Name:    "trace",
				Sources: cli.EnvVars("LOG_LEVEL_TRACE"),
			},
			&cli.StringFlag{
				Name:    extraUserAgentFlag,
				Usage:   "Extra string to append to the user agent. Can be used for tracking purposes in Cloudflare logs.",
				Sources: cli.EnvVars("CLOUDFLARE_EXTRA_USER_AGENT"),
			},
			&cli.StringFlag{
				Name:    "with-base-url",
				Usage:   "Use base URL for API requests. Useful for testing with a local Cloudflare API mock",
				Hidden:  true,
				Sources: cli.EnvVars("CLOUDFLARE_BASE_URL"),
			},
		},
		EnableShellCompletion: true,
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	return app
}

func main() {
	app := buildApp()
	err := app.Run(context.Background(), os.Args)
	logger.Debugf("Running took: %v", time.Since(startTime))
	if err != nil {
		fmt.Printf("Error running app: %s\n", err)
		os.Exit(1)
	}
}

func setup(ctx context.Context, c *cli.Command) (context context.Context, err error) {
	SetLogLevel(c, logger)
	if c.Args().First() == "help" || common.StringSearch("help", c.Args().Slice()) || common.StringSearch("help", c.FlagNames()) || c.Args().First() == "generate-doc" || len(c.Args().Slice()) == 0 {
		return ctx, nil
	}

	apiToken := strings.TrimSpace(c.String(apiTokenFlag))
	apiEmail := strings.TrimSpace(c.String(apiEmailFlag))
	apiKey := strings.TrimSpace(c.String(apiKeyFlag))

	if apiToken == "" && apiEmail == "" && apiKey == "" {
		return ctx, errors.New("no authentication method detected")
	}

	rateLimit := c.Float(rateLimitFlag)
	if c.Bool(lotsOfDeploymentsFlag) && rateLimit == 4 {
		rateLimit = 3
	}
	userAgent := fmt.Sprintf("cloudflare-utils/%s", version)
	if c.String(extraUserAgentFlag) != "" {
		userAgent = fmt.Sprintf("%s (%s)", userAgent, c.String(extraUserAgentFlag))
	}
	cfClientOptions := []cloudflare.Option{
		cloudflare.UsingRateLimit(rateLimit),
		cloudflare.UserAgent(userAgent),
		cloudflare.Debug(logger.Level == logrus.TraceLevel),
		cloudflare.UsingLogger(logger),
	}
	if c.String("with-base-url") != "" {
		cfClientOptions = append(cfClientOptions, cloudflare.BaseURL(c.String("with-base-url")))
	}

	if apiToken != "" {
		APIClient, err = cloudflare.NewWithAPIToken(apiToken, cfClientOptions...)
		if err != nil {
			logger.WithError(err).Error("Error creating new API instance with token")
		}
	}
	if apiEmail != "" || apiKey != "" {
		if apiEmail == "" || apiKey == "" {
			return ctx, errors.New("need to have both API Key and Email set for legacy method")
		}
		logger.Warning("Using legacy method. Using API tokens is recommended")
		APIClient, err = cloudflare.New(apiKey, apiEmail, cfClientOptions...)
		if err != nil {
			logger.WithError(err).Error("Error creating new API instance with legacy method")
		}
	}
	if c.String("account-id") != "" {
		accountRC = cloudflare.AccountIdentifier(c.String("account-id"))
	}
	if c.String("zone-id") != "" {
		zoneRC = cloudflare.ZoneIdentifier(c.String("zone-id"))
	}

	return ctx, err
}

func buildGenerateDocsCommand() *cli.Command {
	return &cli.Command{
		Name:   "generate-doc",
		Hidden: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format",
				Value:   "man",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			logger.Trace("Generating docs")
			formatString := c.String("format")
			if !common.StringSearch(formatString, []string{"man", "markdown"}) {
				return errors.New("invalid format")
			}

			var output string
			var err error
			if formatString == "man" {
				output, err = docs.ToMan(c)
			} else {
				output, err = docs.ToMarkdown(c)
			}
			if err != nil {
				return err
			}
			if c.String("output") != "" {
				err = os.WriteFile(c.String("output"), []byte(output), 0600)
				if err != nil {
					return fmt.Errorf("error writing to output file: %s", err)
				}
			} else {
				_, err = fmt.Fprintln(os.Stdout, output)
				if err != nil {
					return fmt.Errorf("error writing to stdout: %s", err)
				}
			}
			return nil
		},
	}
}
