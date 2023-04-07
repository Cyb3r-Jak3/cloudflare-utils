package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

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
