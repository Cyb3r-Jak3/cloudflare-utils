package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v3"
)

const (
	branchNameFlag = "branch"
	beforeFlag     = "before"
	afterFlag      = "after"
	//timeShortcutFlag = "time" Not implemented yet.
)

type pruneDeploymentOptions struct {
	c                   *cli.Command
	ResourceContainer   *cloudflare.ResourceContainer
	ProjectName         string
	SelectedDeployments []cloudflare.PagesProjectDeployment
}

func buildPruneDeploymentsCommand() *cli.Command {
	return &cli.Command{
		Name:   "prune-deployments",
		Usage:  "Prune deployments by either branch of time\nAPI Token Requirements: Pages:Edit",
		Action: PruneDeploymentsScreen,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     projectNameFlag,
				Aliases:  []string{"p"},
				Usage:    "Pages project to delete the alias from",
				Required: true,
				Sources:  cli.EnvVars("CF_PAGES_PROJECT"),
			},
			&cli.StringFlag{
				Name:    branchNameFlag,
				Aliases: []string{"b"},
				Usage:   "Branch to delete",
				Sources: cli.EnvVars("CF_PAGES_BRANCH"),
			},
			&cli.TimestampFlag{
				Name:  beforeFlag,
				Usage: "Time to delete before",
				Config: cli.TimestampConfig{
					Layouts: []string{"2006-01-02T15:04:05"},
				},
			},
			&cli.TimestampFlag{
				Name:  afterFlag,
				Usage: "Time to delete after",
				Config: cli.TimestampConfig{
					Layouts: []string{"2006-01-02T15:04:05"},
				},
			},
			//&cli.DurationFlag{
			//	Name: timeShortcutFlag,
			//	Usage: "Shortcut for before and after. " +
			//		"Use the format of 1<unit> where unit is one of " +
			//		"y (year), M (month), w (week), d (day), h (hour), m (minute), s (second)" +
			//		"use a negative number to go back in time. Read the docs for more info",
			//},
			&cli.BoolFlag{
				Name:  dryRunFlag,
				Usage: "Don't actually delete anything. Just print what would be deleted",
				Value: false,
			},
		},
	}
}

// PruneDeploymentsScreen is the entry point for the prune-deployments command.
// It handles parsing the CLI arguments and then calls PruneDeploymentsRoot.
func PruneDeploymentsScreen(ctx context.Context, c *cli.Command) error {
	logger.Info("Staring prune deployments")
	if err := CheckAPITokenPermission(ctx, PagesWrite); err != nil {
		return err
	}
	accountID := c.String(accountIDFlag)
	if accountID == "" {
		return errors.New("`account-id` is required for pages commands")
	}

	if c.String(branchNameFlag) != "" && (!c.Timestamp(beforeFlag).IsZero() || !c.Timestamp(afterFlag).IsZero()) {
		return errors.New("cannot specify both a branch and a time range")
	}

	beforeTime := c.Timestamp(beforeFlag)
	afterTime := c.Timestamp(afterFlag)

	if c.String(branchNameFlag) == "" && beforeTime.IsZero() && afterTime.IsZero() {
		return errors.New("need to specify either a branch or a time")
	}
	return PruneDeploymentsRoot(ctx, c)
}

// PruneDeploymentsRoot is the main function for pruning and purging deployments.
func PruneDeploymentsRoot(ctx context.Context, c *cli.Command) error {
	accountResource := cloudflare.AccountIdentifier(c.String(accountIDFlag))
	projectName := c.String(projectNameFlag)

	allDeployments, err := DeploymentsPaginate(
		PagesDeploymentPaginationOptions{
			CLIContext:      c,
			ctx:             ctx,
			AccountResource: accountResource,
			ProjectName:     projectName,
		})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	options := pruneDeploymentOptions{
		c:                   c,
		ResourceContainer:   accountResource,
		ProjectName:         projectName,
		SelectedDeployments: allDeployments,
	}

	var toDelete []cloudflare.PagesProjectDeployment

	if c.String(branchNameFlag) != "" {
		logger.Infoln("Pruning by branch")
		toDelete = PruneBranchDeployments(options)
	} else if !c.Timestamp(beforeFlag).IsZero() || !c.Timestamp(afterFlag).IsZero() {
		logger.Infoln("Pruning by time")
		toDelete = PruneTimeDeployments(options)
	} else {
		logger.Infoln("Purging all deployments")
		toDelete = options.SelectedDeployments
	}

	if len(toDelete) == 0 {
		fmt.Println("Found no deployments to delete")
		return nil
	}

	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry Run: would delete %d deployments", len(toDelete))
		return nil
	}

	failedDeletes := RapidPagesDeploymentDelete(options)
	fmt.Printf("Deleted %d deployments\n", len(toDelete)-len(failedDeletes))
	if len(failedDeletes) > 0 {
		return fmt.Errorf("failed to delete %d deployments", len(failedDeletes))
	}
	return nil
}

// PruneBranchDeployments will return a list of deployments to delete based on the branch name.
func PruneBranchDeployments(options pruneDeploymentOptions) (toDelete []cloudflare.PagesProjectDeployment) {
	selectedBranch := options.c.String(branchNameFlag)

	for _, deployment := range options.SelectedDeployments {
		if deployment.DeploymentTrigger.Metadata == nil {
			continue
		}
		if deployment.DeploymentTrigger.Metadata.Branch == selectedBranch {
			toDelete = append(toDelete, deployment)
		}
	}
	return toDelete
}

// PruneTimeDeployments will return a list of deployments to delete based on the time range.
func PruneTimeDeployments(options pruneDeploymentOptions) (toDelete []cloudflare.PagesProjectDeployment) {
	beforeTimestamp := options.c.Timestamp(beforeFlag)
	afterTimestamp := options.c.Timestamp(afterFlag)
	if !beforeTimestamp.IsZero() {
		logger.Debugln("Pruning with before time")
	} else {
		logger.Debugln("Pruning with  after time")
	}
	for _, deployment := range options.SelectedDeployments {
		if !beforeTimestamp.IsZero() {
			if deployment.CreatedOn.Before(beforeTimestamp) {
				toDelete = append(toDelete, deployment)
			}
		} else {
			if deployment.CreatedOn.After(afterTimestamp) {
				toDelete = append(toDelete, deployment)
			}
		}
	}
	return toDelete
}
