package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Cyb3r-Jak3/cloudflare-utils/cmd"
	"github.com/sirupsen/logrus"
)

var (
	logger  = logrus.New()
	version = "DEV"
	date    = "unknown"
)

func main() {
	startTime := time.Now()
	app := cmd.BuildApp(
		cmd.BuildArgs{
			Version:   version,
			Date:      date,
			Logger:    logger,
			StartTime: startTime,
		})
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), cmd.VersionContextKey, version))
	defer cancel()
	err := app.Run(ctx, os.Args)
	logger.Debugf("Running took: %v", time.Since(startTime))
	if err != nil {
		fmt.Printf("Error running app: %s\n", err)
		os.Exit(1)
	}
}
