package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
)

var (
	log       = logrus.New()
	ctx       = context.Background()
	Version   = "DEV"
	BuildTime = "unknown"
	APIClient *cloudflare.API
)

const (
	apiTokenFlag = "api-token"
	apiEmailFlag = "api-email"
	apiKeyFlag   = "api-key"
	zoneNameFlag = "zone-name"
)

func main() {
	app := &cli.App{
		Name:    "cloudflare-utils",
		Usage:   "Program for quick cloudflare utils",
		Version: fmt.Sprintf("%s (built %s)", Version, BuildTime),
		Suggest: true,
		Authors: []*cli.Author{
			{
				Name:  "Cyb3r-Jak3",
				Email: "git@cyberjake.xyz",
			},
		},
		Commands: []*cli.Command{
			BuildDNSCleanerCommand(),
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
	// Set up log level
	setLogLevel(c)

	// Create Cloudflare API Client
	if c.String(apiTokenFlag) != "" {
		// Create new API Client using an API Token
		APIClient, err = cloudflare.NewWithAPIToken(c.String(apiTokenFlag), cloudflare.UserAgent("cloudflare-utils"))
		if err != nil {
			log.WithError(err).Error("Error creating new API instance with token")
			return err
		}
	} else if c.String(apiKeyFlag) != "" || c.String(apiEmailFlag) != "" {
		// Create new API Client using legacy API Key and API Email
		if c.String(apiKeyFlag) == "" || c.String(apiEmailFlag) == "" {
			log.Error("Need to have both API Key and Email set for legacy method")
		}
		log.Warning("Using legacy method. Using API tokens is recommended")
		APIClient, err = cloudflare.New(c.String(apiKeyFlag), c.String(apiTokenFlag), cloudflare.UserAgent("cloudflare-utils"))
		if err != nil {
			log.WithError(err).Error("Error creating new API instance with legacy method")
		}
	} else {
		return errors.New("no authentication method detected")
	}
	return err
}
