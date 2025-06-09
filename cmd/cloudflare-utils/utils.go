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
	resultInfo := cloudflare.ResultInfo{}
	if params.CLIContext.Bool(lotsOfDeploymentsFlag) {
		resultInfo.PerPage = 4
	}
	startDeploymentListing := time.Now()
	for {
		res, _, err := APIClient.ListPagesDeployments(params.ctx, params.AccountResource, cloudflare.ListPagesDeploymentsParams{
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
		logger.Tracef("Current result info: %v\n", resultInfo)
		resultInfo = resultInfo.Next()
		if resultInfo.Done() {
			logger.Tracef("Breaking pagination loop after %d deployments. %v\n", len(deployments), resultInfo)
			break
		}
	}
	logger.Debugf("Got %d deployments in %s\n", len(deployments), time.Since(startDeploymentListing))
	return deployments, nil
}

// RapidDNSDelete is a helper function to delete DNS records quickly.
// Uses a pool of goroutines to delete records in parallel.
func RapidDNSDelete(ctx context.Context, rc *cloudflare.ResourceContainer, dnsRecords []cloudflare.DNSRecord) map[string]error {
	p := pool.NewWithResults[pruneResults]().WithMaxGoroutines(maxGoRoutines).WithContext(ctx)
	for _, dnsRecord := range dnsRecords {
		p.Go(func(ctx2 context.Context) (pruneResults, error) {
			err := APIClient.DeleteDNSRecord(ctx2, rc, dnsRecord.ID)
			if err != nil {
				return pruneResults{
					ID:      dnsRecord.ID,
					Success: false,
					Error:   err,
				}, fmt.Errorf("error deleting DNS record %s", dnsRecord.ID)
			}
			return pruneResults{
				ID:      dnsRecord.ID,
				Success: true,
			}, nil
		},
		)
	}
	runResults, err := p.Wait()
	if err != nil {
		logger.WithError(err).Error("Error waiting for DNS record deletion. Some records may not have been deleted")
		fmt.Println("Some DNS records may not have been deleted due to an error. Please try again and report the issue if it persists.")
	}
	results := make(map[string]error)

	for _, result := range runResults {
		if !result.Success {
			logger.WithError(result.Error).Warningf("Failed to delete DNS record: %s", result.ID)
			results[result.ID] = result.Error
		}
	}
	return results
}

type pruneResults struct {
	ID      string
	Success bool
	Error   error
}

// RapidPagesDeploymentDelete is a helper function to delete Pages deployments quickly.
// Uses a pool of goroutines to delete deployments in parallel.
func RapidPagesDeploymentDelete(options pruneDeploymentOptions) map[string]error {
	goRoutines := maxGoRoutines
	if options.c.Bool(lotsOfDeploymentsFlag) {
		goRoutines = 5
	}
	p := pool.NewWithResults[pruneResults]().WithMaxGoroutines(goRoutines).WithContext(options.ctx)
	for _, deployment := range options.SelectedDeployments {
		p.Go(func(ctx2 context.Context) (pruneResults, error) {
			err := APIClient.DeletePagesDeployment(ctx2, options.ResourceContainer, cloudflare.DeletePagesDeploymentParams{
				ProjectName:  options.ProjectName,
				DeploymentID: deployment.ID,
				Force:        true,
			})
			if err != nil {
				return pruneResults{
					ID:      deployment.ID,
					Success: false,
					Error:   err,
				}, fmt.Errorf("error deleting deployment %s: %w", deployment.ID, err)
			}
			return pruneResults{ID: deployment.ID, Success: true, Error: nil}, nil
		},
		)
	}
	runResults, err := p.Wait()
	if err != nil {
		logger.WithError(err).Error("Error waiting for deployment deletion. Some deployments may not have been deleted")
		fmt.Println("Some deployments may not have been deleted due to an error. Please try again and report the issue if it persists.")
	}
	results := make(map[string]error)
	for _, result := range runResults {
		if !result.Success {
			logger.WithError(result.Error).Warningf("Failed to delete deployment: %s", result.ID)
			results[result.ID] = result.Error
		}
	}
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

var APITokenNoPermissionError = errors.New("API Token does not have permission to perform this action")

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
		if errors.Is(err, APITokenNoPermissionError) {
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
			return cloudflare.APIToken{}, APITokenNoPermissionError
		}
		logger.WithError(err).Debug("Error getting API token permissions")
		return cloudflare.APIToken{}, err
	}
	return permissions, nil
}
