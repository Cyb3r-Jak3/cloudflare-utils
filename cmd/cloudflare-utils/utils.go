package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

// setLogLevel sets the log level based on the CLI flags
func setLogLevel(c *cli.Context) {
	if c.Bool("debug") {
		log.SetLevel(logrus.DebugLevel)
	} else if c.Bool("verbose") {
		log.SetLevel(logrus.InfoLevel)
	} else {
		switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
		case "trace":
			log.SetLevel(logrus.TraceLevel)
		case "debug":
			log.SetLevel(logrus.DebugLevel)
		default:
			log.SetLevel(logrus.WarnLevel)
		}
	}
	log.Debugf("Log Level set to %v", log.Level)
}

// GetZoneID gets the zone ID from the CLI flags either by name or ID
func GetZoneID(c *cli.Context) (string, error) {
	zoneName := c.String(zoneNameFlag)
	zoneID := c.String(zoneIDFlag)
	if zoneName == "" && zoneID == "" {
		return "", fmt.Errorf("need `%s` or `%s` set", zoneNameFlag, zoneIDFlag)
	}

	if zoneID == "" {
		id, err := APIClient.ZoneIDByName(zoneName)
		if err != nil {
			log.WithError(err).Debug("Error getting zone id from name")
			return "", err
		}
		zoneID = id
	}
	return zoneID, nil
}
