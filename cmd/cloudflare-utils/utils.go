package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
	"strings"
)

// FileExists is a function to check if the file exists at the path
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
	}
	return true
}

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
