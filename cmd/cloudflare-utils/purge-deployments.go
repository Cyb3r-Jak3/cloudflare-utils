package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

const (
	deleteProjectFlag     = "delete-project"
	lotsOfDeploymentsFlag = "lots-of-deployments"
)

func buildPurgeDeploymentsCommand() *cli.Command {
	return &cli.Command{
		Name:   "purge-deployments",
		Usage:  "Delete all deployments for a branch\nAPI Token Requirements: Pages:Edit",
		Action: PurgeDeploymentsScreen,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  deleteProjectFlag,
				Usage: "Delete the project as well. Will attempt to delete the project even if there are errors deleting deployments.",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  lotsOfDeploymentsFlag,
				Usage: "If you are getting errors getting all of the deployments, you may need to use this flag.",
				Value: false,
			},
		}, sharedPagesFlags...),
	}
}

// PurgeDeploymentsScreen is the entry point for the purge-deployments command
// It just calls PruneDeploymentsRoot.
func PurgeDeploymentsScreen(ctx context.Context, c *cli.Command) error {
	logger.Info("Staring purge deployments")
	if err := CheckAPITokenPermission(ctx, PagesWrite); err != nil {
		return err
	}
	return PruneDeploymentsRoot(ctx, c)
}
