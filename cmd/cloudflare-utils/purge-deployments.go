package main

import (
	"errors"
	"fmt"

	"github.com/Cyb3r-Jak3/cloudflare-utils/internal/consts"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

func BuildPurgeDeploymentsCommand() *cli.Command {
	return &cli.Command{
		Name:   "purge-deployments",
		Usage:  "Delete all deployments for a branch\nAPI Token Requirements: Pages:Edit",
		Action: PurgeDeployments,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     consts.ProjectNameFlag,
				Aliases:  []string{"p"},
				Usage:    "Pages project to delete the alias from",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "delete-project",
				Usage: "Delete the project as well",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  consts.DryRunFlag,
				Usage: "Don't actually delete anything. Just print what would be deleted",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Force delete deployments",
				Value: false,
			},
		},
	}
}

func PurgeDeployments(c *cli.Context) error {
	accountID := c.String(consts.AccountIDFlag)
	if accountID == "" {
		return errors.New("`account-id` is required")
	}

	accountResource := cloudflare.AccountIdentifier(accountID)
	projectName := c.String(consts.ProjectNameFlag)

	allDeployments, _, err := APIClient.ListPagesDeployments(c.Context, accountResource, cloudflare.ListPagesDeploymentsParams{
		ProjectName: projectName,
	})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	dryRun := c.Bool(consts.DryRunFlag)
	if dryRun {
		fmt.Printf("Would delete %d deployments for project %s", len(allDeployments), projectName)
		return nil
	}
	forceFlag := c.Bool(consts.ForceFlag)
	errorCount := 0
	for _, deployment := range allDeployments {
		err := APIClient.DeletePagesDeployment(c.Context, accountResource, projectName, deployment.ID, cloudflare.DeletePagesDeploymentParams{Force: forceFlag})
		if err != nil {
			logger.WithField("deployment ID", deployment.ID).Errorf("error deleting deployment: %s", err)
			errorCount++
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("failed to delete %d deployments", errorCount)
	}
	return nil
}
