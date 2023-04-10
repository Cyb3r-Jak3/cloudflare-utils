package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/Cyb3r-Jak3/cloudflare-utils/internal/consts"
	"github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SetLogLevel sets the log level based on the CLI flags.
func SetLogLevel(c *cli.Context, logger *logrus.Logger) {
	if c.Bool("debug") {
		logger.SetLevel(logrus.DebugLevel)
	} else if c.Bool("verbose") {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
		case "trace":
			logger.SetLevel(logrus.TraceLevel)
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		default:
			logger.SetLevel(logrus.WarnLevel)
		}
	}
	logger.Debugf("Log Level set to %v", logger.Level)
}

// GetZoneID gets the zone ID from the CLI flags either by name or ID.
func GetZoneID(c *cli.Context, apiClient *cloudflare.API, logger *logrus.Logger) (string, error) {
	zoneName := c.String(consts.ZoneNameFlag)
	zoneID := c.String(consts.ZoneIDFlag)
	if zoneName == "" && zoneID == "" {
		return "", fmt.Errorf("need `%s` or `%s` set", consts.ZoneNameFlag, consts.ZoneIDFlag)
	}

	if zoneID == "" {
		id, err := apiClient.ZoneIDByName(zoneName)
		if err != nil {
			logger.WithError(err).Debug("Error getting zone id from name")
			return "", err
		}
		zoneID = id
	}
	return zoneID, nil
}
