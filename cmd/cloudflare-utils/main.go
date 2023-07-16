package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"time"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	version       = "DEV"
	date          = "unknown"
	APIClient     *cloudflare.API
	logger        = logrus.New()
	ctx           = context.Background()
	startTime     = time.Now()
	versionString = fmt.Sprintf("%s (built %s)", version, date)
)

func main() {
	if buildInfo, available := debug.ReadBuildInfo(); available {
		goVersion := buildInfo.GoVersion
		versionString = fmt.Sprintf("%s (built %s with %s)", version, date, goVersion)
	}
	app := &cli.App{
		Name:    "cloudflare-utils",
		Usage:   "Program for quick cloudflare utils",
		Version: versionString,
		Suggest: true,
		Authors: []*cli.Author{
			{
				Name:  "Cyb3r-Jak3",
				Email: "git@cyberjake.xyz",
			},
		},
		Before: setup,
		Commands: []*cli.Command{
			BuildDNSCleanerCommand(),
			BuildDNSPurgeCommand(),
			BuildPruneDeploymentsCommand(),
			BuildPurgeDeploymentsCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    apiTokenFlag,
				Usage:   "A scoped API token (preferred)",
				EnvVars: []string{"CLOUDFLARE_API_TOKEN"},
			},
			&cli.StringFlag{
				Name:    apiEmailFlag,
				Usage:   "Cloudflare API email (legacy)",
				EnvVars: []string{"CLOUDFLARE_API_EMAIL"},
			},
			&cli.StringFlag{
				Name:    apiKeyFlag,
				Usage:   "Cloudflare Global API key (legacy)",
				EnvVars: []string{"CLOUDFLARE_API_KEY"},
			},
			&cli.StringFlag{
				Name:    zoneNameFlag,
				Usage:   "Domain name of your zone",
				EnvVars: []string{"CLOUDFLARE_ZONE_NAME"},
			},
			&cli.StringFlag{
				Name:    zoneIDFlag,
				Usage:   "Zone ID of your zone",
				EnvVars: []string{"CLOUDFLARE_ZONE_ID"},
			},
			&cli.StringFlag{
				Name:    accountIDFlag,
				Usage:   "Account ID",
				EnvVars: []string{"CLOUDFLARE_ACCOUNT_ID"},
			},
			&cli.Float64Flag{
				Name:   "rate-limit",
				Usage:  "Rate limit for API calls.\nDefault is 4 which matches the Cloudflare API limit of 1200 calls per 5 minutes",
				Value:  4,
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Enable verbose logging",
				Aliases: []string{"V"},
				EnvVars: []string{"LOG_LEVEL_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug logging",
				Aliases: []string{"d"},
				EnvVars: []string{"LOG_LEVEL_DEBUG"},
			},
			&cli.BoolFlag{
				Name:    "trace",
				EnvVars: []string{"LOG_LEVEL_TRACE"},
				Hidden:  true,
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	err := app.Run(os.Args)
	logger.Debugf("Running took: %v", time.Since(startTime))
	if err != nil {
		fmt.Printf("Error running app: %s\n", err)
		os.Exit(1)
	}
}

func setup(c *cli.Context) (err error) {
	SetLogLevel(c, logger)
	if c.Args().First() == "help" || common.StringSearch("help", c.Args().Slice()) || common.StringSearch("help", c.FlagNames()) {
		return nil
	}

	apiToken := c.String(apiTokenFlag)
	apiEmail := c.String(apiEmailFlag)
	apiKey := c.String(apiKeyFlag)

	if apiToken == "" && apiEmail == "" && apiKey == "" {
		return errors.New("no authentication method detected")
	}

	rateLimit := c.Float64("rate-limit")
	if c.Bool(lotsOfDeploymentsFlag) {
		rateLimit = 3
	}

	cfClientOptions := []cloudflare.Option{
		cloudflare.UsingRateLimit(rateLimit),
		cloudflare.UserAgent(fmt.Sprintf("cloudflare-utils/%s", version)),
		cloudflare.Debug(c.Bool("trace")),
		cloudflare.UsingLogger(logger),
	}

	if apiToken != "" {
		APIClient, err = cloudflare.NewWithAPIToken(apiToken, cfClientOptions...)
		if err != nil {
			logger.WithError(err).Error("Error creating new API instance with token")
		}
	}
	if apiEmail != "" || apiKey != "" {
		if apiEmail == "" || apiKey == "" {
			return errors.New("need to have both API Key and Email set for legacy method")
		}
		logger.Warning("Using legacy method. Using API tokens is recommended")
		APIClient, err = cloudflare.New(apiKey, apiEmail, cfClientOptions...)
		if err != nil {
			logger.WithError(err).Error("Error creating new API instance with legacy method")
		}
	}

	return err
}
