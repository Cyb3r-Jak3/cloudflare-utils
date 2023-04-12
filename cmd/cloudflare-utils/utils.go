package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SetLogLevel sets the log level based on the CLI flags.
func SetLogLevel(c *cli.Context, logger *logrus.Logger) {
	if c.Bool("debug") {
		logger.SetLevel(logrus.DebugLevel)
	} else if c.Bool("verbose") {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
		case "trace":
			logger.SetLevel(logrus.TraceLevel)
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		default:
			logger.SetLevel(logrus.WarnLevel)
		}
	}
	logger.Debugf("Log Level set to %v", logger.Level)
}

// GetZoneID gets the zone ID from the CLI flags either by name or ID.
func GetZoneID(c *cli.Context, apiClient *cloudflare.API, logger *logrus.Logger) (string, error) {
	zoneName := c.String(zoneNameFlag)
	zoneID := c.String(zoneIDFlag)
	if zoneName == "" && zoneID == "" {
		return "", fmt.Errorf("need `%s` or `%s` set", zoneNameFlag, zoneIDFlag)
	}

	if zoneID == "" {
		id, err := apiClient.ZoneIDByName(zoneName)
		if err != nil {
			logger.WithError(err).Debug("Error getting zone id from name")
			return "", err
		}
		zoneID = id
	}
	return zoneID, nil
}

type PagesDeploymentPaginationOptions struct {
	CLIContext      *cli.Context
	APIClient       *cloudflare.API
	AccountResource *cloudflare.ResourceContainer
	ProjectName     string
}

// DeploymentsPaginate is a helper function to paginate through all deployments.
// This is necessary because we need to be able to adjust the per_page parameter for large projects.
func DeploymentsPaginate(params PagesDeploymentPaginationOptions) ([]cloudflare.PagesProjectDeployment, error) {
	var deployments []cloudflare.PagesProjectDeployment
	resultInfo := cloudflare.ResultInfo{}
	if params.CLIContext.Bool(lotsOfDeploymentsFlag) {
		resultInfo.PerPage = 4
	}
	for {
		res, _, err := params.APIClient.ListPagesDeployments(params.CLIContext.Context, params.AccountResource, cloudflare.ListPagesDeploymentsParams{
			ProjectName: params.ProjectName,
			ResultInfo:  resultInfo,
		})
		if err != nil {
			return []cloudflare.PagesProjectDeployment{}, fmt.Errorf("error listing deployments: %w", err)
		}
		deployments = append(deployments, res...)
		resultInfo = resultInfo.Next()
		if resultInfo.Done() {
			break
		}
	}
	return deployments, nil
}
