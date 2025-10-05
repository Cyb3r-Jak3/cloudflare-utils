package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v3"
)

func buildCacheCleanerCommand() *cli.Command {
	return &cli.Command{
		Name:   "cache-cleaner",
		Usage:  "Cleans the cache for a given zone\nAPI Token Requirements: Zone Cache Purge",
		Action: CacheCleaner,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "everything",
				Usage: "Purge everything from the cache. Will throw an error if --url, --tag, or --prefix are also specified.",
				Value: false,
			},
			&cli.StringSliceFlag{
				Name:  "url",
				Usage: "URL to purge from the cache. Can specify multiple times.",
				Value: nil,
			},
			&cli.StringSliceFlag{
				Name:  "tag",
				Usage: "Tag to purge from the cache. Can specify multiple times.",
				Value: nil,
			},
			&cli.StringSliceFlag{
				Name:  "prefix",
				Usage: "Prefix to purge from the cache. Can specify multiple times.",
				Value: nil,
			},
			&cli.StringSliceFlag{
				Name:  "host",
				Usage: "Host to purge from the cache. Can specify multiple times",
				Value: nil,
			},
		},
	}
}

func CacheCleaner(ctx context.Context, c *cli.Command) error {
	logger.Info("Starting cache cleaner")
	everything := c.Bool("everything")
	urls := c.StringSlice("url")
	tags := c.StringSlice("tag")
	prefixes := c.StringSlice("prefix")
	if !everything && len(urls) == 0 && len(tags) == 0 && len(prefixes) == 0 {
		return fmt.Errorf("must specify at least one purge method: --everything, --url, --tag, or --prefix")
	}
	if (everything && len(urls) > 0) || (everything && len(tags) > 0) || (everything && len(prefixes) > 0) {
		return fmt.Errorf("cannot use --everything with --url, --tag, or --prefix")
	}
	if err := CheckAPITokenPermission(ctx, DNSWrite); err != nil {
		return err
	}

	err := GetZoneID(ctx, c)
	if err != nil {
		return err
	}
	if everything {
		logger.Info("Purging everything from cache")
		_, cleanErr := APIClient.PurgeEverything(ctx, zoneRC.Identifier)
		if cleanErr != nil {
			logger.WithError(cleanErr).Error("Error purging everything")
			return cleanErr
		}
		logger.Info("Successfully purged everything")
		return nil
	}
	if len(urls) > 0 {
		logger.Infof("Purging %d URLs from cache", len(urls))
		_, cleanErr := APIClient.PurgeCache(ctx, zoneRC.Identifier, cloudflare.PurgeCacheRequest{
			Files: urls,
		})
		if cleanErr != nil {
			logger.WithError(cleanErr).Error("Error purging URLs")
			return cleanErr
		}
		logger.Info("Successfully purged URLs")
		return nil
	}
	if len(tags) > 0 {
		logger.Infof("Purging %d tags from cache", len(tags))
		_, cleanErr := APIClient.PurgeCache(ctx, zoneRC.Identifier, cloudflare.PurgeCacheRequest{
			Tags: tags,
		})
		if cleanErr != nil {
			logger.WithError(cleanErr).Error("Error purging tags")
			return cleanErr
		}
		logger.Info("Successfully purged tags")
		return nil
	}
	if len(prefixes) > 0 {
		logger.Infof("Purging %d prefixes from cache", len(prefixes))
		_, cleanErr := APIClient.PurgeCache(ctx, zoneRC.Identifier, cloudflare.PurgeCacheRequest{
			Prefixes: prefixes,
		})
		if cleanErr != nil {
			logger.WithError(cleanErr).Error("Error purging prefixes")
			return cleanErr
		}
		logger.Info("Successfully purged prefixes")
		return nil
	}
	return nil
}
