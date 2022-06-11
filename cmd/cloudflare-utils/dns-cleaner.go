package main

import (
	"errors"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
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

// RecordFile is the struct of the YAML DNS file
type RecordFile struct {
	ZoneName string      `yaml:"zone_name"`
	ZoneID   string      `yaml:"zone_id"`
	Records  []DNSRecord `yaml:"records"`
}

// BuildDNSCleanerCommand builds the `dns-cleaner` command for the application
func BuildDNSCleanerCommand() *cli.Command {
	return &cli.Command{
		Name:   "dns-cleaner",
		Usage:  "Clean DNS records. API Token Requirements: DNS:Edit",
		Action: DNSCleaner,
		Subcommands: []*cli.Command{
			{
				Name:   downloadSubCommand,
				Action: DownloadDNS,
				Usage:  "Download DNS records",
			},
			{
				Name:   uploadSubCommand,
				Action: UploadDNS,
				Usage:  "Upload DNS records",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    dnsFileFlag,
				Usage:   "Path to the dns record file",
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
				Usage:   "Auto marks DNS records that are numeric to be removed",
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
				Usage: "Remove the DNS file once the upload completes",
				Value: false,
			},
		},
	}
}

// DNSCleaner is the main action function for the dns-cleaner command.
// It checks if a DNS file exists. If there isn't a DNS file then it downloads records, if there is a file there then it uploads records
func DNSCleaner(c *cli.Context) error {
	if err := setup(c); err != nil {
		return err
	}
	log.Info("Running DNS Cleaner")

	fileExists := FileExists(c.String(dnsFileFlag))
	log.Debugf("Existing DNS file: %t", fileExists)
	if !fileExists {
		log.Info("Downloading DNS Records")
		if err := DownloadDNS(c); err != nil {
			return err
		}
	} else {
		log.Info("Uploading DNS Records")
		if err := UploadDNS(c); err != nil {
			return err
		}
	}
	return nil
}

// quickClean checks to see if a DNS record is numeric
func quickClean(zoneName, record string) bool {
	r := strings.Split(record, fmt.Sprintf(".%s", zoneName))[0]
	log.Debugf("Stripped record: %s", r)
	_, err := strconv.Atoi(r)
	log.Debugf("Error converting: %t", err != nil)
	// if err != nil there is no number, and we should keep
	return err != nil
}

// DownloadDNS downloads current DNS records from Cloudflare
func DownloadDNS(c *cli.Context) error {
	// Always need to set up
	if APIClient == nil {
		if err := setup(c); err != nil {
			return err
		}
	}

	// Make sure that we don't overwrite if told not to
	if FileExists(c.String(dnsFileFlag)) && c.Bool(noOverwriteFlag) {
		return errors.New("existing DNS file found and no overwrite flag is set")
	}

	zoneName := c.String(zoneNameFlag)
	log.Infof("Zone name: %s. DNS file: %s", zoneName, c.String(dnsFileFlag))
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
	recordFile := &RecordFile{
		ZoneID:   id,
		ZoneName: zoneName,
	}
	// Default keep action
	toKeep := !c.Bool(noKeepFlag)
	log.Debugf("ToKeep setting: %t", toKeep)

	// If using quick clean to filter out numeric records
	useQuickClean := c.Bool(quickCleanFlag)
	log.Debugf("Quick Clean setting: %t", useQuickClean)

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
		log.WithError(err).Error("Error marshalling yaml data")
		return err
	}
	// Write out the RecordFile Data
	if err := os.WriteFile(c.String(dnsFileFlag), data, 0755); err != nil {
		log.WithError(err).Error("Error writing DNS file")
		return err
	}
	return nil
}

// UploadDNS makes the changes to DNS records based on the DNS file
func UploadDNS(c *cli.Context) error {

	// Always need to set up
	if APIClient == nil {
		if err := setup(c); err != nil {
			return err
		}
	}

	// Make sure the DNS File exists
	dnsFilePath := c.String(dnsFileFlag)
	if !FileExists(c.String(dnsFilePath)) {
		return fmt.Errorf("no DNS file found at '%s'", dnsFilePath)
	}

	// Read the DNS file and parse the data
	file, err := os.ReadFile(dnsFilePath)
	if err != nil {
		log.WithError(err).Error("Error reading DNS file")
		return err
	}

	recordFile := &RecordFile{}
	if err := yaml.Unmarshal(file, recordFile); err != nil {
		log.WithError(err).Error("Error unmarshalling yaml")
		return err
	}

	// Remove the records
	zoneID := recordFile.ZoneID
	recordCount, errorCount, toRemove := len(recordFile.Records), 0, 0
	for _, record := range recordFile.Records {
		if !record.Keep {
			toRemove++
			if err := APIClient.DeleteDNSRecord(ctx, zoneID, record.ID); err != nil {
				log.WithError(err).Errorf("Error deleting record: %s ID %s", record.Name, record.ID)
				errorCount++
			}
		}
	}
	log.Infof("%d total records. %d to removed. %d errors removing records", recordCount, errorCount, toRemove)

	// Remove DNS record file
	if c.Bool(removeDNSFileFlag) {
		if err := os.Remove(dnsFilePath); err != nil {
			log.WithError(err).Warn("Error deleting old DNS file")
		}
	}
	return nil
}
