package main

import (
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
	"strings"
)

const (
	confirmFlag = "confirm"
)

// BuildDNSPurgeCommand creates the dns-purge command
func BuildDNSPurgeCommand() *cli.Command {
	return &cli.Command{
		Name:  "dns-purge",
		Usage: "Deletes all DNS records",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  confirmFlag,
				Usage: "Auto confirm to delete records. API Token Requirements: DNS:Edit",
				Value: false,
			},
		},
		Action: DNSPurge,
	}
}

// DNSPurge is a command to delete all DNS records without downloading
func DNSPurge(c *cli.Context) error {
	//Always setup
	if err := setup(c); err != nil {
		return err
	}

	zoneName := c.String(zoneNameFlag)
	if zoneName == "" {
		return errors.New("need zone-name set when downloading DNS")
	}
	// Get the Zone ID which is what the API calls are made with
	// Easier to ask for this than having users get the ID
	// ToDo: Allow for Zone ID input
	id, err := APIClient.ZoneIDByName(zoneName)
	if err != nil {
		log.WithError(err).Debug("Error getting zone id from name")
		return err
	}

	// Get all DNS records
	records, err := APIClient.DNSRecords(ctx, id, cloudflare.DNSRecord{})
	if err != nil {
		log.WithError(err).Error("Error getting zone info with ID")
		return err
	}
	if !c.Bool(confirmFlag) {
		var confirmString string
		fmt.Printf("About to remove %d records.\n Continue: y/n", len(records))
		if _, err := fmt.Scanln(&confirmString); err != nil {
			return err
		}
		if !strings.HasPrefix("y", confirmString) {
			fmt.Println("Did not get `y` as input. Exiting")
			return nil
		}

	}
	errorCount := 0
	for _, record := range records {
		if err := APIClient.DeleteDNSRecord(ctx, id, record.ID); err != nil {
			log.WithError(err).Errorf("Error deleting record: %s ID %s", record.Name, record.ID)
			errorCount++
		}
	}
	if errorCount == 0 {
		fmt.Printf("Successfully deleted all %d DNS records\n", len(records))
	} else {
		fmt.Printf("Error deleting %d DNS records.\nPlease review errors and reach out if you belive to be an error with the program", errorCount)
	}
	return nil
}
