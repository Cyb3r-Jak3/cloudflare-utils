package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
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

// BuildDNSCleanerCommand builds the `dns-cleaner` command for the application.
func BuildDNSCleanerCommand() *cli.Command {
	return &cli.Command{
		Name:   "dns-cleaner",
		Usage:  "Clean dns records.\nAPI Token Requirements: DNS:Edit",
		Action: DNSCleaner,
		Subcommands: []*cli.Command{
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
				EnvVars: []string{"DNS_RECORD_FILE"},
				Value:   "./dns-records.yml",
			},
			&cli.BoolFlag{
				Name:    noKeepFlag,
				Usage:   "Mark records for removal by default",
				Aliases: []string{"k"},
				EnvVars: []string{"NO_KEEP"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    quickCleanFlag,
				Usage:   "Auto marks dns records that are numeric to be removed",
				Aliases: []string{"q"},
				EnvVars: []string{"QUICK_CLEAN"},
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
// It checks if a DNS file exists. If there isn't a DNS file then it downloads records, if there is a file there then it uploads records.
func DNSCleaner(c *cli.Context) error {
	logger.Info("Running DNS Cleaner")

	fileExists := common.FileExists(c.String(dnsFileFlag))
	logger.Debugf("Existing DNS file: %t", fileExists)
	if !fileExists {
		logger.Info("Downloading DNS Records")
		if err := DownloadDNS(c); err != nil {
			return err
		}
	} else {
		logger.Info("Uploading DNS Records")
		if err := UploadDNS(c); err != nil {
			return err
		}
	}
	return nil
}

// quickClean checks to see if a DNS record is numeric.
func quickClean(zoneName, record string) bool {
	r := strings.Split(record, fmt.Sprintf(".%s", zoneName))[0]
	logger.Debugf("Stripped record: %s", r)
	_, err := strconv.Atoi(r)
	logger.Debugf("Error converting: %t", err != nil)
	return err != nil
}

// DownloadDNS downloads current DNS records from Cloudflare.
func DownloadDNS(c *cli.Context) error {
	// Make sure that we don't overwrite if told not to
	if common.FileExists(c.String(dnsFileFlag)) && c.Bool(noOverwriteFlag) {
		return errors.New("existing DNS file found and no overwrite flag is set")
	}

	zoneID, err := GetZoneID(c)
	if err != nil {
		return err
	}
	zoneName := c.String(zoneNameFlag)
	if zoneName == "" {
		zoneName = zoneID
	}
	// Get all dns records
	records, _, err := APIClient.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		logger.WithError(err).Error("Error getting zone info with ID")
		return err
	}
	recordFile := &RecordFile{
		ZoneID:   zoneID,
		ZoneName: zoneName,
	}
	// Default keep action
	toKeep := !c.Bool(noKeepFlag)

	// If using quick clean to filter out numeric records
	useQuickClean := c.Bool(quickCleanFlag)

	if useQuickClean && !toKeep {
		return errors.New("using `--quick-clean` is not supported with `--no-keep`")
	}

	// Create Records for the RecordFile
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
		logger.WithError(err).Error("Error marshalling yaml data")
		return err
	}
	// Write out the RecordFile Data
	if err := os.WriteFile(c.String(dnsFileFlag), data, 0600); err != nil {
		logger.WithError(err).Error("Error writing DNS file")
		return err
	}
	return nil
}

// UploadDNS makes the changes to DNS records based on the dns file.
func UploadDNS(c *cli.Context) error {
	// Make sure the dns File exists
	dnsFilePath := c.String(dnsFileFlag)
	if !common.FileExists(dnsFilePath) {
		return fmt.Errorf("no DNS file found at '%s'", dnsFilePath)
	}

	// Read the dns file and parse the data
	file, err := os.ReadFile(dnsFilePath)
	if err != nil {
		logger.WithError(err).Error("Error reading DNS file")
		return err
	}

	recordFile := &RecordFile{}
	if err := yaml.Unmarshal(file, recordFile); err != nil {
		logger.WithError(err).Error("Error unmarshalling yaml")
		return err
	}

	// Remove the records
	zoneID := recordFile.ZoneID
	recordCount, errorCount, toRemove := len(recordFile.Records), 0, 0
	for _, record := range recordFile.Records {
		if !record.Keep {
			toRemove++
			if c.Bool(dryRunFlag) {
				fmt.Printf("Dry Run: Would have removed %s,", record.Name)
				continue
			}
			if err := APIClient.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), record.ID); err != nil {
				logger.WithError(err).WithField("record", record.ID).Error("error deleting DNS record")
				errorCount++
			}
		}
	}
	if c.Bool(dryRunFlag) {
		return nil
	}
	logger.Infof("%d total records. %d to removed. %d errors removing records", recordCount, errorCount, toRemove)

	// Remove dns record file
	if c.Bool(removeDNSFileFlag) {
		if err := os.Remove(dnsFilePath); err != nil {
			logger.WithError(err).Warn("Error deleting old DNS file")
		}
	}
	return nil
}
