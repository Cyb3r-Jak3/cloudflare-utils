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
	persistRetry       = "persist-retry"
	persistRetryAmount = "persist-retry-amount"
)

var sharedPagesFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  dryRunFlag,
		Usage: "Don't actually delete anything. Just print what would be deleted",
		Value: false,
	},
	&cli.BoolFlag{
		Name:  persistRetry,
		Usage: "Persist retry. If the delete fails, it will retry until it succeeds",
		Value: false,
	},
	&cli.IntFlag{
		Name:  persistRetryAmount,
		Usage: "Number of times to retry the delete if it fails",
		Value: 10,
	},
	&cli.StringFlag{
		Name:     projectNameFlag,
		Aliases:  []string{"p"},
		Usage:    "Pages project to delete the alias from",
		Required: true,
		Sources:  cli.EnvVars("CF_PAGES_PROJECT"),
	},
	&cli.BoolFlag{
		Name:  lotsOfDeploymentsFlag,
		Usage: "If you are getting errors getting all of the deployments, you may need to use this flag.",
		Value: false,
	},
	&cli.BoolFlag{
		Name:  forceFlag,
		Usage: "Force delete deployments",
		Value: false,
	},
}

type pruneDeploymentOptions struct {
	c                   *cli.Command
	ProjectName         string
	SelectedDeployments []cloudflare.PagesProjectDeployment
}

func buildPruneDeploymentsCommand() *cli.Command {
	return &cli.Command{
		Name:   "prune-deployments",
		Usage:  "Prune deployments by either branch of time\nAPI Token Requirements: Pages:Edit",
		Action: PruneDeploymentsScreen,
		Flags: append([]cli.Flag{
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
			//	Usage: "Shortcut for before and after. "+
			//		"Use the format of 1<unit> where the unit is one of "+
			//		"y (year), M (month), w (week), d (day), h (hour), m (minute), s (second)" +
			//		"use a negative number to go back in time. Read the docs for more info",
			//},
		}, sharedPagesFlags...),
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
	projectName := c.String(projectNameFlag)

	allDeployments, err := DeploymentsPaginate(
		PagesDeploymentPaginationOptions{
			CLIContext:  c,
			ctx:         ctx,
			ProjectName: projectName,
		})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	options := pruneDeploymentOptions{
		c:                   c,
		ProjectName:         projectName,
		SelectedDeployments: allDeployments,
	}

	var toDelete []cloudflare.PagesProjectDeployment
	branch := c.String(branchNameFlag)
	before := c.Timestamp(beforeFlag)
	after := c.Timestamp(afterFlag)

	preventPurgeAll := c.Name == "prune-deployments"

	if branch != "" {
		logger.Infof("Pruning by branch: %s", branch)
		toDelete = PruneBranchDeployments(branch, options)
	} else if !before.IsZero() || !after.IsZero() {
		logger.Infoln("Pruning by time")
		toDelete = PruneTimeDeployments(options)
	} else {
		if preventPurgeAll {
			return errors.New("refusing to delete all deployments when a branch or time was specified. This is a safety feature to prevent accidental deletion of all deployments")
		}
		logger.Infoln("Purging all deployments")
		toDelete = options.SelectedDeployments
	}

	if len(toDelete) == 0 {
		fmt.Println("Found no deployments to delete")
		return nil
	}

	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry Run: would delete %d deployments\n", len(toDelete))
		return nil
	}

	failedDeletes := RapidPagesDeploymentDelete(options)
	fmt.Printf("Deleted %d deployments\n", len(toDelete)-len(failedDeletes))
	if len(failedDeletes) > 0 {
		return fmt.Errorf("failed to delete %d deployments", len(failedDeletes))
	}
	if c.Bool(deleteProjectFlag) {
		fmt.Printf("Deleting project: %s\n", projectName)
		if projectDeleteErr := APIClient.DeletePagesProject(ctx, accountRC, projectName); err != nil {
			return fmt.Errorf("error deleting project: %w", projectDeleteErr)
		}
	}
	return nil
}

// PruneBranchDeployments will return a list of deployments to delete based on the branch name.
func PruneBranchDeployments(branch string, options pruneDeploymentOptions) []cloudflare.PagesProjectDeployment {
	var toDelete []cloudflare.PagesProjectDeployment
	logger.Debugf("Got %d deployments to check for branch: %s", len(options.SelectedDeployments), branch)
	for _, deployment := range options.SelectedDeployments {
		if deployment.DeploymentTrigger.Metadata == nil {
			logger.Debugln("No metadata for deployment, skipping")
			continue
		}
		logger.Tracef("Got deployment branch: %s", deployment.DeploymentTrigger.Metadata.Branch)
		if deployment.DeploymentTrigger.Metadata.Branch == branch {
			toDelete = append(toDelete, deployment)
		}
	}
	logger.Debugf("Found %d deployments to delete by branch: %s", len(toDelete), branch)
	return toDelete
}

// PruneTimeDeployments will return a list of deployments to delete based on the time range.
func PruneTimeDeployments(options pruneDeploymentOptions) (toDelete []cloudflare.PagesProjectDeployment) {
	beforeTimestamp := options.c.Timestamp(beforeFlag)
	afterTimestamp := options.c.Timestamp(afterFlag)
	if !beforeTimestamp.IsZero() {
		logger.Debugln("Pruning with before time")
	} else {
		logger.Debugln("Pruning with after time")
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
