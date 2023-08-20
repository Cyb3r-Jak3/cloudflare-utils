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

const maxGoRoutines = 10

// SetLogLevel sets the log level based on the CLI flags.
func SetLogLevel(c *cli.Context, logger *logrus.Logger) {
	if c.Bool("debug") {
		logger.SetLevel(logrus.DebugLevel)
	} else if c.Bool("verbose") {
		logger.SetLevel(logrus.InfoLevel)
	} else if c.Bool("trace") {
		logger.SetLevel(logrus.TraceLevel)
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
	logger.Debugf("Log Level set to %v\n", logger.Level)
	logger.Debugf("cloudflare-utils: %s\n", versionString)
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
			logger.WithError(err).Errorln("Error getting zone id from name")
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
				logger.WithError(err).Errorln("Unable to get any deployments")
				return deployments, fmt.Errorf("error listing deployments: %w", err)
			}
			return []cloudflare.PagesProjectDeployment{}, fmt.Errorf("error listing deployments: %w", err)
		}
		deployments = append(deployments, res...)
		resultInfo = resultInfo.Next()
		if resultInfo.Done() {
			break
		}
	}
	logger.Debugf("Got %d deployments in %s\n", len(deployments), time.Since(startDeploymentListing))
	return deployments, nil
}

// RapidDNSDelete is a helper function to delete DNS records quickly.
// Uses a pool of goroutines to delete records in parallel.
func RapidDNSDelete(ctx context.Context, rc *cloudflare.ResourceContainer, dnsRecords []cloudflare.DNSRecord) map[string]error {
	p := pool.NewWithResults[bool]()
	results := make(map[string]error)
	p.WithMaxGoroutines(maxGoRoutines)
	for _, dnsRecord := range dnsRecords {
		p.Go(func() bool {
			err := APIClient.DeleteDNSRecord(ctx, rc, dnsRecord.ID)
			if err != nil {
				logger.WithError(err).Warningf("Error deleting DNS record: %s\n", dnsRecord.ID)
				results[dnsRecord.ID] = err
				return false
			}
			return true
		},
		)
	}
	p.Wait()
	return results
}

// RapidPagesDeploymentDelete is a helper function to delete Pages deployments quickly.
// Uses a pool of goroutines to delete deployments in parallel.
func RapidPagesDeploymentDelete(options pruneDeploymentOptions) map[string]error {
	p := pool.NewWithResults[bool]()
	goRoutines := maxGoRoutines
	if options.c.Bool(lotsOfDeploymentsFlag) {
		goRoutines = 5
	}
	results := make(map[string]error)
	p.WithMaxGoroutines(goRoutines)
	for _, deployment := range options.SelectedDeployments {
		p.Go(func() bool {
			err := APIClient.DeletePagesDeployment(options.c.Context, options.ResourceContainer, cloudflare.DeletePagesDeploymentParams{
				ProjectName:  options.ProjectName,
				DeploymentID: deployment.ID,
				Force:        true,
			})
			if err != nil {
				logger.WithError(err).Warningf("Error deleting deployment: %s\n", deployment.ID)
				results[deployment.ID] = err
				return false
			}
			return true
		},
		)
	}
	p.Wait()
	return results
}
