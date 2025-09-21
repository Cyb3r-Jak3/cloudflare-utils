package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SyncList_HTTPSource(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	serverBase := os.Getenv("CLOUDFLARE_BASE_URL")
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", fmt.Sprintf("%s/test-ips.txt", serverBase)}
	err := app.Run(t.Context(), args)
	assert.NoError(t, err, "Expected no error when syncing list from HTTP source")
}

func Test_SyncList_HTTPSourceDryRun(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	serverBase := os.Getenv("CLOUDFLARE_BASE_URL")
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--dry-run", "--source", fmt.Sprintf("%s/test-ips.txt", serverBase)}
	err := app.Run(t.Context(), args)
	assert.NoError(t, err, "Expected no error when syncing list from HTTP source with dry-run")
}

func Test_SyncList_FileSource(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	fileName := "test-ips.txt"
	err := os.WriteFile(fileName, []byte("1.2.3.4\n2001:db8::1\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(fileName)
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--no-wait", "--source", "file://" + fileName}
	err = app.Run(t.Context(), args)
	assert.NoError(t, err, "Expected no error when syncing list from file source")
}

func Test_SyncList_PresetSource(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "preset://cloudflare"}
	err := app.Run(t.Context(), args)
	assert.NoError(t, err, "Expected no error when syncing list from preset source")
}

func Test_SyncList_InvalidSource(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "invalid://source"}
	err := app.Run(t.Context(), args)
	assert.Error(t, err, "Expected error for invalid source scheme")
}

func Test_SyncList_EmptyIPs(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	// Point to an empty file
	fileName := "empty-ips.txt"
	err := os.WriteFile(fileName, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(fileName)
	args := []string{"cloudflare-utils", "sync-list", "--list-name", "test-list", "--source", "file://" + fileName}
	err = app.Run(t.Context(), args)
	assert.Error(t, err, "Expected error for empty IPs list")
}
