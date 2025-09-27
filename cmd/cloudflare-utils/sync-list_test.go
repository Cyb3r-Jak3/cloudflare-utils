package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SyncList_HTTPSource(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "https://www.cloudflare.com/ips-v4"})
	assert.NoError(t, err, "Expected no error when syncing list from HTTP source")
}

func Test_SyncList_SourceAsArg(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "https://www.cloudflare.com/ips-v4"})
	assert.NoError(t, err, "Expected no error when syncing list from HTTP source as argument")
}

func Test_SyncList_Errors(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name:   "Missing source",
			args:   []string{"cloudflare-utils", "sync-list", "--list-name", "test-list"},
			errMsg: "source must be provided as an argument or with --source",
		},
		{
			name:   "Missing list ID and name",
			args:   []string{"cloudflare-utils", "sync-list", "https://www.cloudflare.com/ips-v4"},
			errMsg: "either --list-id or --list-name must be provided",
		},
		{
			name:   "invalid ip version",
			args:   []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "https://www.cloudflare.com/ips-v4", "--ip-version", "invalid"},
			errMsg: "invalid ip-version: invalid. Valid versions are: both, ipv4, ipv6",
		},
		{
			name:   "unsupported source scheme",
			args:   []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "ftp://example.com/ips.txt"},
			errMsg: "invalid source scheme: ftp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := withApp(t, tc.args)
			assert.Error(t, err, "Expected error when running sync-list command")
			if err == nil || err.Error() != tc.errMsg {
				t.Errorf("Expected error message to contain '%s', got: %v", tc.errMsg, err)
			}
		})
	}
}

func Test_SyncList_TestID(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-id", "2c0fc9fa937b11eaa1b71c4d701ab86e", "https://www.cloudflare.com/ips-v4"})
	assert.NoError(t, err, "Expected no error when syncing list from HTTP source when using list ID")
}

func Test_SyncList_HTTPSourceDryRun(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--dry-run", "--source", "https://www.cloudflare.com/ips-v4"})
	assert.NoError(t, err, "Expected no error when dry-running sync list from HTTP source")
}

func Test_SyncList_FileSource(t *testing.T) {
	fileName := "test-ips.txt"
	err := os.WriteFile(fileName, []byte("1.2.3.4\n2001:db8::1\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(fileName)
	err = withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--no-wait", "--source", "file://" + fileName})
	assert.NoError(t, err, "Expected no error when syncing list from file source")
}

func Test_SyncList_Presets(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "Uptime Robot Preset",
			args: []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://uptime-robot", "--no-comment"},
		},
		{
			name: "GitHub Preset with Exclude",
			args: []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://github?exclude=actions", "--no-wait"},
		},
		{
			name: "GitHub Preset with Exclude",
			args: []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://github"},
		},
		{
			name: "Cloudflare Preset without China",
			args: []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://cloudflare?china=false", "--ip-version", "both"},
		},
		{
			name: "Cloudflare Preset with China",
			args: []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://cloudflare?china=true", "--ip-version", "ipv6"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, withApp(t, tc.args), "Expected error when running sync-list command")
		})
	}
}

func Test_SyncList_EmptyIPs(t *testing.T) {
	fileName := "empty-ips.txt"
	err := os.WriteFile(fileName, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(fileName)
	err = withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "file://" + fileName})
	assert.Error(t, err, "Expected error for empty IPs list")
}
