package main

import (
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

func BuildDeleteAliasCommand() *cli.Command {
	return &cli.Command{
		Name:   "delete-alias-deployments",
		Usage:  "Delete an alias",
		Action: DeleteAlias,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "alias",
				Aliases: []string{"a"},
				Usage:   "The alias to delete",
			},
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Usage:   "The project to delete the alias from",
			},
		},
	}
}

func DeleteAlias(c *cli.Context) error {
	projectName := c.String("project")
	if projectName == "" {
		return errors.New("`project` is required")
	}

	selectedAlias := c.String("alias")
	if selectedAlias == "" {
		return errors.New("`alias` is required")
	}

	// Hacky solution to pagination until cloudflare-go supports it
	// https://github.com/cloudflare/cloudflare-go/pull/1264
	var allDeployments []cloudflare.PagesProjectDeployment
	paginate := cloudflare.PaginationOptions{}
	for {
		deployments, res, err := APIClient.ListPagesDeployments(c.Context, cloudflare.AccountIdentifier(c.String(accountIDFlag)), cloudflare.ListPagesDeploymentsParams{
			ProjectName:       projectName,
			PaginationOptions: paginate,
		})
		if err != nil {
			return fmt.Errorf("error listing deployments: %w", err)
		}
		allDeployments = append(allDeployments, deployments...)
		if res.TotalPages == res.Page {
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
		for _, alias := range deployment.Aliases {
			if alias == selectedAlias {
				toDelete = append(toDelete, deployment)
				break
			}
		}
	}
	if len(toDelete) == 0 {
		fmt.Println("No deployments found with alias", selectedAlias)
		return nil
	}

	errorCount := 0
	for _, deployment := range toDelete {
		err := APIClient.DeletePagesDeployment(c.Context, cloudflare.AccountIdentifier(c.String(accountIDFlag)), projectName, deployment.ID)
		if err != nil {
			log.WithError(err).WithField("deployment", deployment.ID).Error("error deleting deployment")
			errorCount++
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("error deleting %d deployments out of %d", errorCount, len(toDelete))
	}
	fmt.Println("Deleted", len(toDelete), "deployments")

	return nil
}
