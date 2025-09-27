package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DNSCleanerRootDownload(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "--trace", "dns-cleaner", "--zone-id", "2", "--dns-file", "test.yaml"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err = os.Stat("test.yaml"); os.IsNotExist(err) {
		t.Errorf("File test.yaml was not created")
	} else {
		// Remove the file after the test
		err = os.Remove("test.yaml")
		if err != nil {
			t.Fatalf("Error removing test.yaml: %v", err)
		}
	}
}

func Test_DNSCleanerDownload(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "--trace", "dns-cleaner", "download", "--zone-id", "2", "--dns-file", "test.yaml", "--quick-clean"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err = os.Stat("test.yaml"); os.IsNotExist(err) {
		t.Errorf("File test.yaml was not created")
	} else {
		// Remove the file after the test
		err = os.Remove("test.yaml")
		if err != nil {
			t.Errorf("Error removing test.yaml: %v", err)
		}
	}
}

func Test_Failed_DNSCleanerDownload(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	t.Setenv("CLOUDFLARE_ZONE_NAME", "nonexistent")
	t.Setenv("CLOUDFLARE_ZONE_ID", "")
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--trace", "dns-cleaner", "download", "--dns-file", "test.yaml"})
	assert.EqualError(t, err, "zone could not be found", "Expected error when running the app with dns-cleaner download command with missing zone name")
}

func Test_DNSCleaner(t *testing.T) {
	outputFileName := "complete.yaml"
	err := withApp(t, []string{"cloudflare-utils", "dns-cleaner", "--zone-name", "2", "--dns-file", outputFileName, "--no-keep"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err = os.Stat(outputFileName); os.IsNotExist(err) {
		t.Errorf("File %s was not created", outputFileName)
	}
	err = withApp(t, []string{"cloudflare-utils", "dns-cleaner", "--zone-name", "2", "--dns-file", outputFileName})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner upload command")

	removeErr := os.Remove(outputFileName)
	assert.NoError(t, removeErr, "Expected no error when removing the output file")
}
