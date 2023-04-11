package main

import (
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

const (
	branchNameFlag = "branch"
)

func BuildDeleteBranchCommand() *cli.Command {
	return &cli.Command{
		Name:   "delete-branch-deployments",
		Usage:  "Delete add deployments for a branch\nAPI Token Requirements: Pages:Edit",
		Action: DeleteBranchDeployments,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     projectNameFlag,
				Aliases:  []string{"p"},
				Usage:    "Pages project to delete the alias from",
				Required: true,
				EnvVars:  []string{"CF_PAGES_PROJECT"},
			},
			&cli.StringFlag{
				Name:     branchNameFlag,
				Aliases:  []string{"b"},
				Usage:    "Branch to delete",
				Required: true,
				EnvVars:  []string{"CF_PAGES_BRANCH"},
			},
			&cli.BoolFlag{
				Name:  dryRunFlag,
				Usage: "Don't actually delete anything. Just print what would be deleted",
				Value: false,
			},
		},
	}
}

func DeleteBranchDeployments(c *cli.Context) error {
	accountID := c.String(accountIDFlag)
	if accountID == "" {
		return errors.New("`account-id` is required")
	}
	accountResource := cloudflare.AccountIdentifier(accountID)

	projectName := c.String(projectNameFlag)
	selectedBranch := c.String(branchNameFlag)

	allDeployments, err := DeploymentsPaginate(
		PagesDeploymentPaginationOptions{
			CLIContext:      c,
			APIClient:       APIClient,
			AccountResource: accountResource,
			ProjectName:     projectName,
		})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	var toDelete []cloudflare.PagesProjectDeployment
	for _, deployment := range allDeployments {
		if deployment.DeploymentTrigger.Metadata == nil {
			continue
		}
		if deployment.DeploymentTrigger.Metadata.Branch == selectedBranch {
			toDelete = append(toDelete, deployment)
		}
	}
	if len(toDelete) == 0 {
		fmt.Println("No deployments found with branch", selectedBranch)
		return nil
	}

	errorCount := 0
	for _, deployment := range toDelete {
		if c.Bool(dryRunFlag) {
			fmt.Println("Dry Run: Would delete", deployment.ID)
			continue
		}
		logger.Debugf("Deleting deployment %s", deployment.ID)
		err := APIClient.DeletePagesDeployment(c.Context, accountResource, cloudflare.DeletePagesDeploymentParams{
			ProjectName:  projectName,
			DeploymentID: deployment.ID,
			Force:        true,
		})
		if err != nil {
			logger.WithError(err).WithField("deployment", deployment.ID).Error("error deleting deployment")
			errorCount++
		}
	}
	if c.Bool(dryRunFlag) {
		fmt.Println("Dry run complete")
		return nil
	}
	if errorCount > 0 {
		return fmt.Errorf("error deleting %d deployments out of %d", errorCount, len(toDelete))
	}
	fmt.Println("Deleted", len(toDelete), "deployments")

	return nil
}
