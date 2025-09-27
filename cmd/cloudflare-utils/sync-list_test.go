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

func Test_SyncList_PresetCloudflare(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://cloudflare?china=true", "--comment", "test comment"})
	assert.NoError(t, err, "Expected no error when syncing list from cloudflare-china source")
}

func Test_SyncList_PresetGitHub(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://github", "--no-wait"})
	assert.NoError(t, err, "Expected no error when syncing list from preset source")
}

func Test_SyncList_PresetGitHubExclude(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://github?exclude=actions", "--no-wait"})
	assert.NoError(t, err, "Expected no error when syncing list from preset source")
}

func Test_SyncList_PresetUptime(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://uptime-robot", "--no-comment"})
	assert.NoError(t, err, "Expected no error when syncing list from preset source")
}

func Test_SyncList_InvalidSource(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "invalid://source"})
	assert.Error(t, err, "Expected error for invalid source scheme")
	if err == nil || err.Error() != "invalid source scheme: invalid" {
		t.Errorf("Expected error message to contain 'invalid source scheme', got: %v", err)
	}

	err = withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "ftp://example.com/ips.txt"})
	if err == nil || err.Error() != "invalid source scheme: ftp" {
		t.Errorf("Expected error message to contain 'invalid source scheme', got: %v", err)
	}
}

func Test_SyncList_InvalidIPVersion(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "https://example.com/ips.txt", "--ip-version", "invalid"})

	assert.Error(t, err, "Expected error for invalid IP version")
	if err == nil || err.Error() != "invalid ip-version: invalid. Valid versions are: both, ipv4, ipv6" {
		t.Errorf("Expected error message to contain 'invalid ip version', got: %v", err)
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
