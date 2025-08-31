package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

const (
	downloadSubCommand = "download"
	uploadSubCommand   = "upload"
	dnsFileFlag        = "dns-file"
	noKeepFlag         = "no-keep"
	quickCleanFlag     = "quick-clean"
	noOverwriteFlag    = "no-overwrite"
	removeDNSFileFlag  = "remove-file"
)

type DNSRecord struct {
	ID      string `yaml:"id"`
	Keep    bool   `yaml:"keep"`
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Content string `yaml:"content"`
}

// RecordFile is the struct of the YAML DNS file.
type RecordFile struct {
	ZoneName string      `yaml:"zone_name"`
	ZoneID   string      `yaml:"zone_id"`
	Records  []DNSRecord `yaml:"records"`
}

// buildDNSCleanerCommand builds the `dns-cleaner` command for the application.
func buildDNSCleanerCommand() *cli.Command {
	return &cli.Command{
		Name:   "dns-cleaner",
		Usage:  "Clean dns records.\nAPI Token Requirements: DNS:Edit",
		Action: DNSCleaner,
		Commands: []*cli.Command{
			{
				Name:   downloadSubCommand,
				Action: DownloadDNS,
				Usage:  "Download dns records",
			},
			{
				Name:   uploadSubCommand,
				Action: UploadDNS,
				Usage:  "Upload dns records",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    dnsFileFlag,
				Usage:   "Path to the DNS record file",
				Aliases: []string{"f"},
				Sources: cli.EnvVars("DNS_RECORD_FILE"),
				Value:   "./dns-records.yml",
			},
			&cli.BoolFlag{
				Name:    noKeepFlag,
				Usage:   "Mark records for removal by default",
				Aliases: []string{"k"},
				Sources: cli.EnvVars("NO_KEEP"),
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    quickCleanFlag,
				Usage:   "Auto marks dns records that are numeric to be removed",
				Aliases: []string{"q"},
				Sources: cli.EnvVars("QUICK_CLEAN"),
			},
			&cli.BoolFlag{
				Name:    noOverwriteFlag,
				Usage:   "Do not replace existing DNS file",
				Aliases: []string{"n"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  removeDNSFileFlag,
				Usage: "Remove the dns file once the upload completes",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  dryRunFlag,
				Usage: "Do not make any changes. Only applies to upload",
				Value: false,
			},
		},
	}
}

// DNSCleaner is the main action function for the dns-cleaner command.
// It checks if a DNS file exists. If there isn't a DNS file, then it downloads records, if there is a file there, then it uploads records.
func DNSCleaner(ctx context.Context, c *cli.Command) error {
	logger.Infoln("Starting DNS Cleaner")

	if err := CheckAPITokenPermission(ctx, DNSWrite); err != nil {
		return err
	}

	fileExists := common.FileExists(c.String(dnsFileFlag))
	logger.Debugf("Existing DNS file: %t\n", fileExists)
	if !fileExists {
		logger.Infoln("Downloading DNS Records")
		if err := DownloadDNS(ctx, c); err != nil {
			return err
		}
	} else {
		logger.Infoln("Uploading DNS Records")
		if err := UploadDNS(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// quickClean checks to see if a DNS record is numeric.
func quickClean(zoneName, record string) bool {
	r := strings.Split(record, fmt.Sprintf(".%s", zoneName))[0]
	logger.Debugf("Stripped record: %s\n", r)
	_, err := strconv.Atoi(r)
	logger.Debugf("Error converting: %t\n", err != nil)
	return err != nil
}

// DownloadDNS downloads current DNS records from Cloudflare.
func DownloadDNS(ctx context.Context, c *cli.Command) error {
	if common.FileExists(c.String(dnsFileFlag)) && c.Bool(noOverwriteFlag) {
		return errors.New("existing DNS file found and no overwrite flag is set")
	}

	zoneID, err := GetZoneID(ctx, c)
	if err != nil {
		return err
	}
	zoneName := strings.TrimSpace(c.String(zoneNameFlag))
	if zoneName == "" {
		zoneName = zoneID
	}
	records, _, err := APIClient.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		logger.WithError(err).Errorln("Error getting zone info with ID")
		return err
	}
	recordFile := &RecordFile{
		ZoneID:   zoneID,
		ZoneName: zoneName,
	}
	toKeep := !c.Bool(noKeepFlag)

	useQuickClean := c.Bool(quickCleanFlag)

	if useQuickClean && !toKeep {
		return errors.New("using `--quick-clean` is not supported with `--no-keep`")
	}

	for _, record := range records {
		var keepValue bool
		if useQuickClean {
			keepValue = quickClean(zoneName, record.Name)
		} else {
			keepValue = toKeep
		}
		recordFile.Records = append(recordFile.Records, DNSRecord{
			Name:    record.Name,
			ID:      record.ID,
			Type:    record.Type,
			Keep:    keepValue,
			Content: record.Content,
		})
	}
	data, err := yaml.Marshal(&recordFile)
	if err != nil {
		logger.WithError(err).Errorln("Error marshalling yaml data")
		return err
	}
	if err := os.WriteFile(c.String(dnsFileFlag), data, 0600); err != nil {
		logger.WithError(err).Errorln("Error writing DNS file")
		return err
	}
	return nil
}

// UploadDNS makes the changes to DNS records based on the dns file.
func UploadDNS(_ context.Context, c *cli.Command) error {
	dnsFilePath := c.String(dnsFileFlag)
	if !common.FileExists(dnsFilePath) {
		return fmt.Errorf("no DNS file found at '%s'", dnsFilePath)
	}

	file, err := os.ReadFile(dnsFilePath)
	if err != nil {
		logger.WithError(err).Errorln("Error reading DNS file")
		return err
	}

	recordFile := &RecordFile{}
	if err := yaml.Unmarshal(file, recordFile); err != nil {
		logger.WithError(err).Errorln("Error unmarshalling yaml")
		return err
	}

	zoneResource := cloudflare.ZoneIdentifier(recordFile.ZoneID)
	recordCount, errorCount := len(recordFile.Records), 0
	var toRemove []cloudflare.DNSRecord

	for _, record := range recordFile.Records {
		if !record.Keep {
			if c.Bool(dryRunFlag) {
				fmt.Printf("Dry Run: Would have removed %s\n", record.Name)
				continue
			}
			toRemove = append(toRemove, cloudflare.DNSRecord{
				ID: record.ID,
			})
		}
	}

	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry Run: Would have removed %d records\n", len(toRemove))
		return nil
	}
	removeErrors := RapidDNSDelete(zoneResource, toRemove)
	errorCount = len(removeErrors)

	if errorCount == 0 {
		fmt.Printf("Successfully deleted all %d dns records\n", len(toRemove))
	} else {
		fmt.Printf("Error deleting %d dns records.\nPlease review errors and reach out if you believe to be an error with the program\n", errorCount)
		if logger.IsLevelEnabled(logrus.InfoLevel) {
			logger.Infoln("Errors:")
			for deleteID, deleteErr := range removeErrors {
				logger.Debugf("Error deleting record: %s: %s\n", deleteID, deleteErr)
			}
		}
	}

	logger.Infof("%d total records. %d to removed. %d errors removing records", recordCount, len(toRemove), errorCount)
	if c.Bool(removeDNSFileFlag) {
		if err := os.Remove(dnsFilePath); err != nil {
			logger.WithError(err).Warnln("Error deleting old DNS file")
		}
	}
	return nil
}
