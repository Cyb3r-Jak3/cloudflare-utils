package main

import (
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-github/v68/github"
	"github.com/urfave/cli/v2"
)

const (
	allTunnelsFlag     = "all-tunnels"
	includeDeletedFlag = "include-deleted"
	activeOnlyFlag     = "healthy-only"
)

func buildTunnelVersionCommand() *cli.Command {
	return &cli.Command{
		Name:   "tunnel-versions",
		Usage:  "Get version of tunnel connectors\nAPI Token Requirements: Cloudflare Tunnel:Read",
		Action: TunnelVersionAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    allTunnelsFlag,
				Aliases: []string{"a"},
				Usage:   "Reports versions of all connectors not just outdated ones",
				EnvVars: []string{"ALL_TUNNELS"},
			},
			&cli.BoolFlag{
				Name:    includeDeletedFlag,
				Aliases: []string{"d"},
				Usage:   "Include deleted tunnels in the report",
				EnvVars: []string{"INCLUDE_DELETED_TUNNELS"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    activeOnlyFlag,
				Aliases: []string{"o"},
				Usage:   "Only report on healthy tunnels",
				EnvVars: []string{"ACTIVE_TUNNELS_ONLY"},
				Value:   false,
			},
		},
	}
}

func GetLatestVersion() (string, error) {
	gClient := github.NewClient(nil)
	release, _, err := gClient.Repositories.GetLatestRelease(ctx, "cloudflare", "cloudflared")
	if err != nil {
		return "", err
	}
	return *release.TagName, nil
}

func TunnelVersionAction(c *cli.Context) error {
	accountRC := cloudflare.AccountIdentifier(c.String(accountIDFlag))
	tunnels, _, err := APIClient.ListTunnels(c.Context, accountRC, cloudflare.TunnelListParams{
		IsDeleted: cloudflare.BoolPtr(c.Bool(includeDeletedFlag)),
	})
	if err != nil {
		logger.WithError(err).Error("Error getting tunnels from API")
		return err
	}
	if c.Bool(activeOnlyFlag) {
		screenedTunnels := make([]cloudflare.Tunnel, 0)
		for _, tunnel := range tunnels {
			if tunnel.Status == "healthy" {
				screenedTunnels = append(screenedTunnels, tunnel)
			}
		}
		tunnels = screenedTunnels
	}

	latestVersion, err := GetLatestVersion()
	if err != nil {
		logger.WithError(err).Error("Error getting latest release of cloudflared from github")
		return err
	}
	logger.WithField("latestVersion", latestVersion).Debug("Cloudflared latest version")
	logger.Debugf("There are %d tunnels", len(tunnels))
	countedMap := make(map[string]map[string]int)
	allTunnels := c.Bool(allTunnelsFlag)
	for _, tunnel := range tunnels {
		connectorVersionMap := make(map[string][]string)
		for _, connector := range tunnel.Connections {
			if allTunnels || connector.ClientVersion != latestVersion {
				connectorVersionMap[tunnel.Name] = append(connectorVersionMap[tunnel.Name], connector.ClientVersion)
			}
		}

		if connectorVersionMap[tunnel.Name] != nil {
			logger.Debugf("Connector version count for %s: %#v", tunnel.Name, getUniqueVersions(connectorVersionMap[tunnel.Name]))
			countedMap[tunnel.Name] = getUniqueVersions(connectorVersionMap[tunnel.Name])
		} else {
			logger.Debugf("No outdated connectors for tunnel: %s", tunnel.Name)
		}
	}

	logger.Tracef("Connector version map: %#v", countedMap)
	if len(countedMap) == 0 {
		fmt.Println("All connectors are up to date")
		return nil
	}
	fmt.Printf("There are %d outdated connectors\n", len(countedMap))
	for tunnelName, connectorVersions := range countedMap {
		fmt.Printf("Tunnel: %s\n", tunnelName)
		for connectorVersion, count := range connectorVersions {
			fmt.Printf("\tVersion: %s, Count: %d\n", connectorVersion, count)
		}
	}
	return nil
}

func getUniqueVersions(connectorVersions []string) (uniqueVersions map[string]int) {
	uniqueVersions = make(map[string]int)
	for _, connectorVersion := range connectorVersions {
		uniqueVersions[connectorVersion] = uniqueVersions[connectorVersion] + 1
	}
	return uniqueVersions
}
