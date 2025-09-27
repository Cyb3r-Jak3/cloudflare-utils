package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Cyb3r-Jak3/common/v5"
	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-github/v75/github"
	"github.com/urfave/cli/v3"
)

const (
	ipv4Flag   = "ipv4"
	ipv6Flag   = "ipv6"
	ipBothFlag = "both"
)

var (
	validPresets = []string{"cloudflare", "uptime-robot", "github"}
)

func buildListSyncCommand() *cli.Command {
	return &cli.Command{
		Name:   "sync-list",
		Usage:  "Syncs a list of IPs with a Cloudflare List. This currently replaces all items in a list\nAPI Token Requirements: Account Filter Lists:Edit",
		Action: SyncList,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "list-name",
				Usage: "Name of the list to sync with. If the list does not exist, it will be created.",
			},
			&cli.StringFlag{
				Name:  "list-id",
				Usage: "ID of the list to sync with. If both list-name and list-id are provided, list-id will be used.",
			},
			&cli.StringFlag{
				Name: "source",
				Usage: "Source of the IPs to sync. Can be a URL, file path, or preset. URL and file path must start with http(s):// or file:// respectively.\n" +
					"Presets starts with preset://. Currently, support presets are: \n" +
					"  - cloudflare. You can also do ?include=china to include China DC IP addresses\n" +
					"  - uptime-robot\n" +
					"  - github\n" +
					"For more information on formats, see: https://cloudflare-utils.cyberjake.xyz/lists/sync-list/",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					sourceURL, err := url.Parse(s)
					if err != nil {
						return fmt.Errorf("precheck error parsing source URL: %w", err)
					}
					switch sourceURL.Scheme {
					case "http", "https", "file":
						return nil
					case "preset":
						if !common.StringSearch(sourceURL.Host, validPresets) {
							return fmt.Errorf("invalid preset: %s. Valid presets are: %s", sourceURL.Host, strings.Join(validPresets, ", "))
						}
					default:
						return fmt.Errorf("invalid source scheme: %s", sourceURL.Scheme)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "ip-version",
				Usage: fmt.Sprintf("IP version to sync. Can be either %s, %s, or %s. Default is %s.", ipv4Flag, ipv6Flag, ipBothFlag, ipBothFlag),
				Value: "both",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					validVersions := []string{ipv4Flag, ipv6Flag, ipBothFlag}
					if !common.StringSearch(s, validVersions) {
						return fmt.Errorf("invalid ip-version: %s. Valid versions are: %s", s, strings.Join(validVersions, ", "))
					}

					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "no-wait",
				Usage: "If set, the command will not wait for the list sync operation to complete. This means that the command will exit immediately after starting the operation. You can check the status of the operation later using the operation ID.",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  dryRunFlag,
				Usage: "Don't actually sync anything. Just print what would be synced.",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "no-comment",
				Usage: "If set, the command will not add a comment to each list item indicating when it was added. This is useful if you want to keep the list items clean.",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "comment",
				Usage: "Custom comment to add to each list item. If not set, a default comment will be used. Ignored if --no-comment is set.",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					if len(s) > 64 {
						return fmt.Errorf("comment cannot be longer than 64 characters")
					}
					return nil
				},
			},
		}, githubTokenFlag),
	}
}

func SyncList(ctx context.Context, c *cli.Command) error {
	listName := c.String("list-name")
	listID := c.String("list-id")
	if listName == "" && listID == "" {
		return fmt.Errorf("either --list-id or --list-name must be provided")
	}
	listSource := c.String("source")
	if listSource == "" {
		listSource = c.Args().First()
		if listSource == "" {
			return fmt.Errorf("source must be provided as an argument or with --source")
		}
	}
	sourceURL, err := url.Parse(listSource)
	if err != nil {
		return fmt.Errorf("error parsing source URL: %w", err)
	}
	var ips []string
	switch sourceURL.Scheme {
	case "preset":
		switch sourceURL.Host {
		case "cloudflare":
			ips, err = getCloudflareIPs(c, sourceURL.Query())
			if err != nil {
				return fmt.Errorf("error getting Cloudflare IPs: %w", err)
			}
		case "uptime-robot":
			ips, err = getUptimeRobotIPs(ctx)
			if err != nil {
				return fmt.Errorf("error getting Uptime Robot IPs: %w", err)
			}
		case "github":
			ips, err = getGitHubIPs(ctx, c, sourceURL.Query())
			if err != nil {
				return fmt.Errorf("error getting GitHub IPs: %w", err)
			}
		default:
			return fmt.Errorf("invalid preset: %s", sourceURL.Host)
		}
	case "http", "https":
		ips, err = getIPsFromURL(ctx, listSource)
		if err != nil {
			return fmt.Errorf("error getting IPs from URL: %w", err)
		}

	case "file":
		filePath := sourceURL.Host
		if !common.FileExists(filePath) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				ips = append(ips, line)
			}
		}
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IPs found to sync")
	}

	if listID == "" {
		listID, err = getCloudflareList(ctx, c)
		if err != nil {
			return fmt.Errorf("error getting Cloudflare list: %w", err)
		}
	}
	if listID == "" {
		return fmt.Errorf("list ID is empty after attempting to fetch or create the list")
	}

	listItems := getFilteredIPs(ips, c)

	logger.Infof("Syncing %d IPs to list ID %s", len(listItems), listID)
	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry Run: Would sync %d IPs to list ID %s\n", len(listItems), listID)
		return nil
	}
	syncStart := time.Now()
	opID, err := APIClient.ReplaceListItemsAsync(ctx, accountRC, cloudflare.ListReplaceItemsParams{ID: listID, Items: listItems})
	if err != nil {
		return fmt.Errorf("error replacing list items: %w", err)
	}
	if c.Bool("no-wait") {
		fmt.Printf("Started async operation to replace list items. Operation ID: %s\n", opID.Result.OperationID)
		return nil
	}
	logger.Infof("Started async operation to replace list items. Operation ID: %s", opID.Result.OperationID)
	err = PollListBulkOperation(ctx, accountRC, opID.Result.OperationID)
	if err != nil {
		return fmt.Errorf("error polling list bulk operation: %w", err)
	}
	logger.Debugf("List sync operation completed in %s", time.Since(syncStart).String())
	fmt.Printf("Successfully synced %d IPs to list ID %s\n", len(listItems), listID)
	return nil
}

func getFilteredIPs(ips []string, c *cli.Command) []cloudflare.ListItemCreateRequest {
	listItems := make([]cloudflare.ListItemCreateRequest, 0, len(ips))
	ipVersion := c.String("ip-version")

	comment := "Added by cloudflare-utils sync-list on " + startTime.Format(time.RFC822Z)
	customComment := c.String("comment")
	if customComment != "" {
		comment = customComment
	}
	// Override comment if no-comment is set
	if c.Bool("no-comment") {
		comment = ""
	}
	for _, ip := range ips {
		if ip == "" {
			logger.Warn("Skipping empty IP")
			continue
		}
		if ipVersion == ipv4Flag && strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
			listItems = append(listItems, cloudflare.ListItemCreateRequest{
				IP:      cloudflare.StringPtr(ip),
				Comment: comment,
			})
		} else if ipVersion == ipv6Flag && strings.Contains(ip, ":") && !strings.Contains(ip, ".") {
			listItems = append(listItems, cloudflare.ListItemCreateRequest{
				IP:      cloudflare.StringPtr(ip),
				Comment: comment,
			})
		} else if ipVersion == ipBothFlag {
			listItems = append(listItems, cloudflare.ListItemCreateRequest{
				IP:      cloudflare.StringPtr(ip),
				Comment: comment,
			})
		}
	}
	return listItems
}

func getCloudflareList(ctx context.Context, c *cli.Command) (string, error) {
	listName := c.String("list-name")
	// Fetch list by Name
	logger.Infof("Fetching list by name: %s", listName)
	if accountRC == nil {
		return "", fmt.Errorf("accountRC is nil")
	}
	lists, err := APIClient.ListLists(ctx, accountRC, cloudflare.ListListsParams{})
	if err != nil {
		return "", fmt.Errorf("error fetching lists: %w", err)
	}
	for _, list := range lists {
		if list.Name == listName {
			logger.Infof("Found list with name %s and ID %s", list.Name, list.ID)
			return list.ID, nil
		}
	}
	// Create list if not found
	if c.Bool(dryRunFlag) {
		fmt.Printf("Dry Run: Would have created list with name %s\n", listName)
		return "dry-run-list-id", nil
	}
	if listName == "" {
		return "", fmt.Errorf("could not find list and list name is empty, cannot create list")
	}
	logger.Infof("List with name %s not found, creating it", listName)
	newList, err := APIClient.CreateList(ctx, accountRC, cloudflare.ListCreateParams{
		Name:        listName,
		Description: "Created by cloudflare-utils",
		Kind:        "ip",
	})
	if err != nil {
		return "", fmt.Errorf("error creating list: %w", err)
	}
	logger.Infof("Created list with name %s and ID %s", newList.Name, newList.ID)
	return newList.ID, nil
}

func getCloudflareIPs(c *cli.Command, query url.Values) ([]string, error) {
	var ips []string
	ranges, err := cloudflare.IPs()
	if err != nil {
		return nil, fmt.Errorf("error fetching Cloudflare IPs: %w", err)
	}
	includeChina := queryToList(query)["china"]
	ipVersion := c.String("ip-version")
	if ipVersion == ipv4Flag || ipVersion == ipBothFlag {
		ips = append(ips, ranges.IPv4CIDRs...)
	}
	if ipVersion == ipv6Flag || ipVersion == ipBothFlag {
		ips = append(ips, ranges.IPv6CIDRs...)
	}
	if includeChina {
		if ipVersion == ipv4Flag || ipVersion == ipBothFlag {
			ips = append(ips, ranges.ChinaIPv4CIDRs...)
		}
		if ipVersion == ipv6Flag || ipVersion == ipBothFlag {
			ips = append(ips, ranges.ChinaIPv6CIDRs...)
		}
	}

	return ips, nil
}

func getUptimeRobotIPs(ctx context.Context) ([]string, error) {
	return getIPsFromURL(ctx, "https://cdn.uptimerobot.com/api/IPv4andIPv6.txt")
}

func getGitHubIPs(ctx context.Context, c *cli.Command, query url.Values) ([]string, error) {
	githubToken := c.String(githubTokenFlagName)
	gClient := github.NewClient(nil)
	if githubToken != "" {
		gClient = github.NewClient(nil).WithAuthToken(githubToken)
	}
	results, _, err := gClient.Meta.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching GitHub IPs: %w", err)
	}
	exclude := queryToList(query)
	var ips []string
	if !exclude["git"] {
		ips = append(ips, results.Git...)
	}
	if !exclude["hooks"] {
		ips = append(ips, results.Hooks...)
	}
	if !exclude["pages"] {
		ips = append(ips, results.Pages...)
	}
	if !exclude["importer"] {
		ips = append(ips, results.Importer...)
	}
	if !exclude["actions"] {
		ips = append(ips, results.Actions...)
	}
	if !exclude["dependabot"] {
		ips = append(ips, results.Dependabot...)
	}
	if !exclude["actions-macos"] {
		ips = append(ips, results.ActionsMacos...)
	}
	if !exclude["api"] {
		ips = append(ips, results.API...)
	}
	if !exclude["packages"] {
		ips = append(ips, results.Packages...)
	}
	if !exclude["web"] {
		ips = append(ips, results.Web...)
	}

	// Remove duplicates
	unique := make(map[string]struct{})
	var deduped []string
	for _, ip := range ips {
		if _, exists := unique[ip]; !exists {
			unique[ip] = struct{}{}
			deduped = append(deduped, ip)
		}
	}

	return deduped, nil
}

func getIPsFromURL(ctx context.Context, url string) ([]string, error) {
	var ips []string
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching IPs from URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching IPs from URL: received status code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading IPs from URL: %w", err)
	}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			ips = append(ips, line)
		}
	}
	return ips, nil
}

func queryToList(query url.Values) map[string]bool {
	// Parse include query parameter
	// Example: ?include=china,foo,bar
	// Result: map[string]bool{"china": true, "foo": true, "bar": true}

	includes := make(map[string]bool)
	include := query.Get("include")
	if include != "" {
		for _, inc := range strings.Split(include, ",") {
			includes[strings.TrimSpace(strings.ToLower(inc))] = true
		}
	}
	return includes
}
