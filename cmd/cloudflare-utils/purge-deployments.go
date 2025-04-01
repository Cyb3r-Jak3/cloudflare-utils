package main

import (
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
				Usage: "Delete the project as well. Will attempt to delete the project even if there are errors deleting deployments.",
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

// PurgeDeploymentsScreen is the entry point for the purge-deployments command
// It just calls PruneDeploymentsRoot.
func PurgeDeploymentsScreen(c *cli.Context) error {
	logger.Info("Staring purge deployments")
	if err := CheckAPITokenPermission(c.Context, PagesWrite); err != nil {
		return err
	}
	return PruneDeploymentsRoot(c)
}
