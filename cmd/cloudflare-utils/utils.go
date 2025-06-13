package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/sourcegraph/conc/pool"

	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

const maxGoRoutines = 10

// SetLogLevel sets the log level based on the CLI flags.
func SetLogLevel(c *cli.Command, logger *logrus.Logger) {
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
	logger.Debugf("Log Level set to %v", logger.Level)
	logger.Debugf("cloudflare-utils: %s", versionString)
}

// GetZoneID gets the zone ID from the CLI flags either by name or ID.
func GetZoneID(ctx context.Context, c *cli.Command) (string, error) {
	zoneName := c.String(zoneNameFlag)
	zoneID := c.String(zoneIDFlag)
	if zoneName == "" && zoneID == "" {
		return "", fmt.Errorf("need `%s` or `%s` set", zoneNameFlag, zoneIDFlag)
	}

	if zoneID == "" {
		id, err := APIClient.ZoneIDByName(zoneName)
		if err != nil {
			if logrus.DebugLevel >= logger.Level {
				zones, lErr := APIClient.ListZones(ctx)
				if lErr != nil {
					logger.WithError(err).Debugln("Error listing zones")
				}
				logger.Debugf("Got %d zones", len(zones))
				for _, zone := range zones {
					logger.Debugf("Zone: %s", zone.Name)
				}
			}
			logger.WithError(err).Errorln("Error getting zone id from name")
			return "", err
		}
		zoneID = id
	}
	return zoneID, nil
}

type PagesDeploymentPaginationOptions struct {
	CLIContext      *cli.Command
	ctx             context.Context
	AccountResource *cloudflare.ResourceContainer
	ProjectName     string
}

// DeploymentsPaginate is a helper function to paginate through all deployments.
// This is necessary because we need to be able to adjust the per_page parameter for large projects.
func DeploymentsPaginate(params PagesDeploymentPaginationOptions) ([]cloudflare.PagesProjectDeployment, error) {
	var deployments []cloudflare.PagesProjectDeployment
	resultInfo := &cloudflare.ResultInfo{}
	if params.CLIContext.Bool(lotsOfDeploymentsFlag) {
		resultInfo.PerPage = 4
	}
	startDeploymentListing := time.Now()
	for {
		res, innerResultInfo, err := APIClient.ListPagesDeployments(params.ctx, params.AccountResource, cloudflare.ListPagesDeploymentsParams{
			ProjectName: params.ProjectName,
			ResultInfo:  *resultInfo,
		})
		if err != nil {
			if len(deployments) != 0 {
				logger.WithError(err).Errorln("Unable to get any deployments")
				return deployments, fmt.Errorf("error listing deployments: %w", err)
			}
			return []cloudflare.PagesProjectDeployment{}, fmt.Errorf("error listing deployments: %w", err)
		}
		deployments = append(deployments, res...)
		if innerResultInfo.Page == innerResultInfo.TotalPages {
			logger.Tracef("Breaking pagination loop after %d deployments.\n", len(deployments))
			break
		}
		resultInfo = innerResultInfo
	}
	duration := time.Since(startDeploymentListing)
	minutes := int(duration.Minutes())
	seconds := duration.Seconds() - float64(minutes*60)
	logger.Debugf("Got %d deployments in %dm %.2fs\n", len(deployments), minutes, seconds)
	return deployments, nil
}

// RapidDNSDelete is a helper function to delete DNS records quickly.
// Uses a pool of goroutines to delete records in parallel.
func RapidDNSDelete(rc *cloudflare.ResourceContainer, dnsRecords []cloudflare.DNSRecord) map[string]error {
	p := pool.NewWithResults[bool]()
	results := make(map[string]error)
	p.WithMaxGoroutines(maxGoRoutines)
	for _, dnsRecord := range dnsRecords {
		p.Go(func() bool {
			err := APIClient.DeleteDNSRecord(context.Background(), rc, dnsRecord.ID)
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
			err := APIClient.DeletePagesDeployment(context.Background(), options.ResourceContainer, cloudflare.DeletePagesDeploymentParams{
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

type APIPermissionName string

const (
	DNSWrite    APIPermissionName = "DNSWrite"
	PagesWrite  APIPermissionName = "PagesWrite"
	TunnelRead  APIPermissionName = "TunnelRead"
	TunnelWrite APIPermissionName = "TunnelWrite"
)

var apiPermissionMap = map[APIPermissionName]string{
	DNSWrite:    "4755a26eedb94da69e1066d98aa820be",
	PagesWrite:  "8d28297797f24fb8a0c332fe0866ec89",
	TunnelRead:  "efea2ab8357b47888938f101ae5e053f",
	TunnelWrite: "c07321b023e944ff818fec44d8203567",
}

func CheckAPITokenPermission(ctx context.Context, permission ...APIPermissionName) error {
	if APIClient.APIToken == "" {
		logger.Debug("No API Token set. Skipping permission check")
		return nil
	}
	if logger.Level >= logrus.DebugLevel {
		logger.Debugf("Checking API Token permission: %s", permission)
	}
	token, err := VerifyAPIToken(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "API token is not authorized to check if it has the correct permissions") {
			return nil
		}
		return err
	}
	permissionIDMap := make([]string, len(permission))
	for _, p := range permission {
		permissionIDMap = append(permissionIDMap, apiPermissionMap[p])
	}
	logger.Debugf("There are %d policies", len(token.Policies))
	for _, policy := range token.Policies {
		logger.Debugf("Policy ID: %s, N", policy.ID)
		for _, p := range policy.PermissionGroups {
			if slices.Contains(permissionIDMap, p.ID) {
				return nil
			}
		}
	}
	return fmt.Errorf("API Token does not have permission %s", permission)
}

func VerifyAPIToken(ctx context.Context) (cloudflare.APIToken, error) {
	verified, err := APIClient.VerifyAPIToken(ctx)
	if err != nil {
		logger.WithError(err).Error("Error verifying API token")
		return cloudflare.APIToken{}, err
	}
	permissions, err := APIClient.GetAPIToken(ctx, verified.ID)
	if err != nil {
		if strings.Contains(err.Error(), "Unauthorized to access requested resource") {
			logger.Debug("API token is not authorized to check if it has the correct permissions")
			return cloudflare.APIToken{}, errors.New("API token is not authorized to check if it has the correct permissions")
		}
		logger.WithError(err).Debug("Error getting API token permissions")
		return cloudflare.APIToken{}, err
	}
	return permissions, nil
}
