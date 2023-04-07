package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"runtime/debug"
	"sort"
)

var (
	log       = logrus.New()
	ctx       = context.Background()
	version   = "DEV"
	date      = "unknown"
	goVersion = "unknown"
	APIClient *cloudflare.API
)

const (
	apiTokenFlag  = "api-token"
	apiEmailFlag  = "api-email"
	apiKeyFlag    = "api-key"
	zoneNameFlag  = "zone-name"
	accountIDFlag = "account-id"
	zoneIDFlag    = "zone-id"
)

func main() {
	if buildInfo, available := debug.ReadBuildInfo(); available {
		goVersion = buildInfo.GoVersion
	}
	app := &cli.App{
		Name:    "cloudflare-utils",
		Usage:   "Program for quick cloudflare utils",
		Version: fmt.Sprintf("%s (built %s with %s)", version, date, goVersion),
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
			BuildDeleteAliasCommand(),
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
				Name:    accountIDFlag,
				Usage:   "Account ID",
				EnvVars: []string{"CLOUDFLARE_ACCOUNT_ID"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				EnvVars: []string{"LOG_LEVEL_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				EnvVars: []string{"LOG_LEVEL_DEBUG"},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("Error running app")
	}
}

func setup(c *cli.Context) (err error) {
	setLogLevel(c)

	if c.String(apiTokenFlag) != "" {
		APIClient, err = cloudflare.NewWithAPIToken(c.String(apiTokenFlag), cloudflare.UserAgent(fmt.Sprintf("cloudflare-utils/%s", version)))
		if err != nil {
			log.WithError(err).Error("Error creating new API instance with token")
		}
		return err
	}
	if c.String(apiKeyFlag) != "" || c.String(apiEmailFlag) != "" {
		if c.String(apiKeyFlag) == "" || c.String(apiEmailFlag) == "" {
			return errors.New("need to have both API Key and Email set for legacy method")
		}
		log.Warning("Using legacy method. Using API tokens is recommended")
		APIClient, err = cloudflare.New(c.String(apiKeyFlag), c.String(apiTokenFlag), cloudflare.UserAgent(fmt.Sprintf("cloudflare-utils/%s", version)))
		if err != nil {
			log.WithError(err).Error("Error creating new API instance with legacy method")
		}
		return err
	}

	return errors.New("no authentication method detected")
}
