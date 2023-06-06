package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/conc/pool"

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
	logger.Debugf("cloudflare-utils: %s", versionString)
}

// GetZoneID gets the zone ID from the CLI flags either by name or ID.
func GetZoneID(c *cli.Context) (string, error) {
	zoneName := c.String(zoneNameFlag)
	zoneID := c.String(zoneIDFlag)
	if zoneName == "" && zoneID == "" {
		return "", fmt.Errorf("need `%s` or `%s` set", zoneNameFlag, zoneIDFlag)
	}

	if zoneID == "" {
		id, err := APIClient.ZoneIDByName(zoneName)
		if err != nil {
			logger.WithError(err).Error("Error getting zone id from name")
			return "", err
		}
		zoneID = id
	}
	return zoneID, nil
}

type PagesDeploymentPaginationOptions struct {
	CLIContext      *cli.Context
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
	startDeploymentListing := time.Now()
	for {
		res, _, err := APIClient.ListPagesDeployments(params.CLIContext.Context, params.AccountResource, cloudflare.ListPagesDeploymentsParams{
			ProjectName: params.ProjectName,
			ResultInfo:  resultInfo,
		})
		if err != nil {
			if len(deployments) != 0 {
				logrus.WithError(err).Error("Unable to get all deployments")
				return deployments, fmt.Errorf("error listing deployments: %w", err)
			}
			return []cloudflare.PagesProjectDeployment{}, fmt.Errorf("error listing deployments: %w", err)
		}
		deployments = append(deployments, res...)
		resultInfo = resultInfo.Next()
		if resultInfo.DoneCount() {
			break
		}
	}
	logrus.Debugf("Got %d deployments in %s", len(deployments), time.Since(startDeploymentListing))
	return deployments, nil
}

func BatchPagesDelete(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string, deployments []cloudflare.PagesProjectDeployment) []error {
	p := pool.NewWithResults[error]()
	p.WithMaxGoroutines(50)
	for _, deployment := range deployments {
		p.Go(func() error {
			err := APIClient.DeletePagesDeployment(ctx, rc, cloudflare.DeletePagesDeploymentParams{
				ProjectName:  projectName,
				DeploymentID: deployment.ID,
				Force:        true,
			})
			if err != nil {
				logrus.WithError(err).Warningf("Error deletinging deployment: %s", deployment.ID)
			}
			return err
		},
		)
	}
	return p.Wait()
}
