package main

import (
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

const (
	confirmFlag = "confirm"
)

// buildDNSPurgeCommand creates the dns-purge command.
func buildDNSPurgeCommand() *cli.Command {
	return &cli.Command{
		Name:  "dns-purge",
		Usage: "Deletes all dns records.\nAPI Token Requirements: DNS:Edit",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  confirmFlag,
				Usage: "Auto confirm to delete records",
				Value: false,
			},
		},
		Action: DNSPurge,
	}
}

// DNSPurge is a command to delete all dns records without downloading.
func DNSPurge(c *cli.Context) error {
	logger.Info("Starting DNS Purge")
	zoneID, err := GetZoneID(c)
	if err != nil {
		return err
	}

	zoneResource := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := APIClient.ListDNSRecords(ctx, zoneResource, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		logger.WithError(err).Error("Error getting zone info with ID")
		return err
	}
	if !c.Bool(confirmFlag) {
		var confirmString string
		fmt.Printf("About to remove %d records.\nContinue (y/n): ", len(records))
		if _, err := fmt.Scanln(&confirmString); err != nil {
			return err
		}
		if !strings.EqualFold(confirmString, "y") {
			fmt.Println("Did not get `y` as input. Exiting")
			return nil
		}
	}
	if len(records) == 0 {
		fmt.Println("No records to delete")
		return nil
	}

	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry run: Would have deleted %d records\n", len(records))
		return nil
	}

	errors := RapidDNSDelete(c.Context, zoneResource, records)
	errorCount := len(errors)

	if errorCount == 0 {
		fmt.Printf("Successfully deleted all %d dns records\n", len(records))
	} else {
		fmt.Printf("Error deleting %d dns records.\nPlease review errors and reach out if you believe to be an error with the program\n", errorCount)
	}
	return nil
}
