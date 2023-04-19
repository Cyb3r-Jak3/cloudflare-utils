package main

import (
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

const (
	deleteProjectFlag     = "delete-project"
	lotsOfDeploymentsFlag = "lots-of-deployments"
)

func BuildPurgeDeploymentsCommand() *cli.Command {
	return &cli.Command{
		Name:   "purge-deployments",
		Usage:  "Delete all deployments for a branch\nAPI Token Requirements: Pages:Edit",
		Action: PurgeDeployments,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     projectNameFlag,
				Aliases:  []string{"p"},
				Usage:    "Pages project to delete the alias from",
				EnvVars:  []string{"CF_PAGES_PROJECT"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:  deleteProjectFlag,
				Usage: "Delete the project as well",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  dryRunFlag,
				Usage: "Don't actually delete anything. Just print what would be deleted",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  lotsOfDeploymentsFlag,
				Usage: "If you are getting errors getting all of the deployments, you may need to use this flag.",
				Value: false,
			},
		},
	}
}

func PurgeDeployments(c *cli.Context) error {
	logger.Debug("Staring purge deployments")
	accountID := c.String(accountIDFlag)
	if accountID == "" {
		return errors.New("`account-id` is required")
	}

	accountResource := cloudflare.AccountIdentifier(accountID)
	projectName := c.String(projectNameFlag)

	allDeployments, err := DeploymentsPaginate(
		PagesDeploymentPaginationOptions{
			CLIContext:      c,
			AccountResource: accountResource,
			ProjectName:     projectName,
		})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	logger.Debugf("Found %d deployments for project %s", len(allDeployments), projectName)

	dryRun := c.Bool(dryRunFlag)
	if dryRun {
		fmt.Printf("Would delete %d deployments for project %s", len(allDeployments), projectName)
		return nil
	}

	deleteErrors := BatchPagesDelete(c.Context, accountResource, projectName, allDeployments)

	errorCount := len(deleteErrors)

	//errorCount := 0
	//for range allDeployments {
	//	deployment := allDeployments[len(allDeployments)-1]
	//	err := APIClient.DeletePagesDeployment(c.Context, accountResource, cloudflare.DeletePagesDeploymentParams{
	//		ProjectName:  projectName,
	//		DeploymentID: deployment.ID,
	//		Force:        forceFlag,
	//	})
	//	if err != nil {
	//		logger.WithField("deployment ID", deployment.ID).Errorf("error deleting deployment: %s", err)
	//		errorCount++
	//	}
	//	allDeployments = allDeployments[:len(allDeployments)-1]
	//	logger.Debugf("Deleted deployment %s", deployment.ID)
	//	logger.Debugf("Remaining deployments: %d", len(allDeployments))
	//}
	if errorCount > 0 {
		return fmt.Errorf("failed to delete %d deployments", errorCount)
	}

	if c.Bool(deleteProjectFlag) {
		err := APIClient.DeletePagesProject(c.Context, accountResource, projectName)
		if err != nil {
			return fmt.Errorf("error deleting project: %w", err)
		}
		fmt.Printf("Deleted project %s", projectName)
	}

	return nil
}
