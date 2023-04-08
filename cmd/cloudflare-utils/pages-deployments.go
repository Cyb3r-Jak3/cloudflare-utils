package main

import (
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

const (
	projectNameFlag = "project"
	branchNameFlag  = "branch"
	dryRunFlag      = "dry-run"
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
			},
			&cli.StringFlag{
				Name:     branchNameFlag,
				Aliases:  []string{"b"},
				Usage:    "Branch to delete",
				Required: true,
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

	// Hacky solution to pagination until cloudflare-go supports it
	// https://github.com/cloudflare/cloudflare-go/pull/1264
	var allDeployments []cloudflare.PagesProjectDeployment
	paginate := cloudflare.PaginationOptions{}
	for {
		//log.Debugf("Getting page %d", paginate.Page)
		deployments, res, err := APIClient.ListPagesDeployments(c.Context, accountResource, cloudflare.ListPagesDeploymentsParams{
			ProjectName:       projectName,
			PaginationOptions: paginate,
		})
		if err != nil {
			return fmt.Errorf("error listing deployments: %w", err)
		}
		allDeployments = append(allDeployments, deployments...)
		if len(deployments) == 0 || res.Page == res.TotalPages {
			break
		}
		paginate.Page = res.Page + 1
	}

	//deployments, _, err := APIClient.ListPagesDeployments(c.Context, cloudflare.AccountIdentifier(c.String(accountIDFlag)), cloudflare.ListPagesDeploymentsParams{
	//	ProjectName: projectName,
	//})
	//if err != nil {
	//	return fmt.Errorf("error listing deployments: %w", err)
	//}
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
			fmt.Println("Would delete", deployment.ID)
			continue
		}
		fmt.Println("Deleting", deployment.ID)
		err := APIClient.DeletePagesDeployment(c.Context, accountResource, projectName, deployment.ID)
		if err != nil {
			log.WithError(err).WithField("deployment", deployment.ID).Error("error deleting deployment")
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
