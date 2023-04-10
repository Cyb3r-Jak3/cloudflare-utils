package dns

import (
	"fmt"
	"github.com/Cyb3r-Jak3/cloudflare-utils/internal"
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
		Usage: "Deletes all DNS records.\nAPI Token Requirements: DNS:Edit",
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

// DNSPurge is a command to delete all DNS records without downloading
func DNSPurge(c *cli.Context) error {
	zoneID, err := internal.GetZoneID(c)
	if err != nil {
		return err
	}

	// Get all DNS records
	records, _, err := main.APIClient.ListDNSRecords(main.ctx, cloudflare.ResourceIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		log.WithError(err).Error("Error getting zone info with ID")
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
	errorCount := 0
	for _, record := range records {
		if err := main.APIClient.DeleteDNSRecord(main.ctx, cloudflare.ZoneIdentifier(zoneID), record.ID); err != nil {
			main.log.WithError(err).Errorf("Error deleting record: %s ID %s", record.Name, record.ID)
			main.log.WithError(err).Errorf("Error deleting record: %s ID %s", record.Name, record.ID)
			errorCount++
		}
	}
	if errorCount == 0 {
		fmt.Printf("Successfully deleted all %d DNS records\n", len(records))
	} else {
		fmt.Printf("Error deleting %d DNS records.\nPlease review errors and reach out if you believe to be an error with the program", errorCount)
	}
	return nil
}
