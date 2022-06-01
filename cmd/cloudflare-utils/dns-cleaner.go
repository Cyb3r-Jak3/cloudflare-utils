package main

import (
	"errors"
	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	dnsFileFlag = "dns-file"
	noKeepFlag  = "no-keep"
)

type DNSRecord struct {
	ID   string `yaml:"id"`
	Keep bool   `yaml:"keep"`
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type RecordFile struct {
	ZoneName string      `yaml:"zone_name"`
	ZoneID   string      `yaml:"zone_id"`
	Records  []DNSRecord `yaml:"records"`
}

func BuildDNSCleanerCommand() *cli.Command {
	return &cli.Command{
		Name:   "dns-cleaner",
		Usage:  "Clean DNS records. API Token Requirements: DNS:Edit",
		Action: DNSCleaner,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    dnsFileFlag,
				Usage:   "path to the dns record file",
				EnvVars: []string{"DNS_RECORD_FILE"},
				Value:   "./dns-records.yml",
			},
			&cli.BoolFlag{
				Name:    noKeepFlag,
				Usage:   "mark records for removal by default",
				EnvVars: []string{"NO_KEEP"},
				Value:   false,
			},
		},
	}
}

func DNSCleaner(c *cli.Context) error {
	if err := setup(c); err != nil {
		return err
	}
	log.Info("Running DNS Cleaner")

	fileExists := true
	_, err := os.Stat(c.String(dnsFileFlag))
	if err != nil {
		fileExists = os.IsExist(err)
	}
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

func DownloadDNS(c *cli.Context) error {
	zoneName := c.String(zoneNameFlag)
	log.Infof("Zone name: %s. DNS file: %s", zoneName, c.String(dnsFileFlag))
	if zoneName == "" {
		return errors.New("need zone-name set when downloading DNS")
	}
	id, err := APIClient.ZoneIDByName(zoneName)
	if err != nil {
		log.WithError(err).Debug("Error getting zone id from name")
		return err
	}
	records, err := APIClient.DNSRecords(ctx, id, cloudflare.DNSRecord{})
	if err != nil {
		log.WithError(err).Error("Error getting zone info with ID")
		return err
	}
	recordFile := &RecordFile{
		ZoneID:   id,
		ZoneName: zoneName,
	}
	for _, record := range records {
		recordFile.Records = append(recordFile.Records, DNSRecord{
			Name: record.Name,
			ID:   record.ID,
			Type: record.Type,
			Keep: c.Bool(noKeepFlag),
		})
	}
	data, err := yaml.Marshal(&recordFile)
	if err != nil {
		log.WithError(err).Error("Error marshing yaml data")
		return err
	}
	if err := os.WriteFile(c.String(dnsFileFlag), data, 0755); err != nil {
		log.WithError(err).Error("Error writing DNS file")
		return err
	}
	return nil
}

func UploadDNS(c *cli.Context) error {
	file, err := os.ReadFile(c.String(dnsFileFlag))
	if err != nil {
		log.WithError(err).Error("Error reading DNS file")
		return err
	}
	recordFile := &RecordFile{}
	if err := yaml.Unmarshal(file, recordFile); err != nil {
		log.WithError(err).Error("Error unmarshalling yaml")
		return err
	}
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
	return nil
}
