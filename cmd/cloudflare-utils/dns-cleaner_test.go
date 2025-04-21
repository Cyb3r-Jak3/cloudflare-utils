package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DNSCleanerRootDownload(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	zone := os.Getenv("CLOUDFLARE_ZONE_NAME")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	if zone == "" {
		t.Skip("CLOUDFLARE_ZONE_NAME environment variable not set")
	}
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--trace", "dns-cleaner", "--zone-name", zone, "--dns-file", "test.yaml"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err := os.Stat("test.yaml"); os.IsNotExist(err) {
		t.Errorf("File test.yaml was not created")
	} else {
		// Remove the file after the test
		err := os.Remove("test.yaml")
		if err != nil {
			t.Errorf("Error removing test.yaml: %v", err)
		}
	}
}

func Test_DNSCleanerDownload(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	zone := os.Getenv("CLOUDFLARE_ZONE_NAME")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	if zone == "" {
		t.Skip("CLOUDFLARE_ZONE_NAME environment variable not set")
	}
	//t.Setenv("CLOUDFLARE_API_TOKEN", token)
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--trace", "dns-cleaner", "download", "--zone-name", zone, "--dns-file", "test.yaml"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err := os.Stat("test.yaml"); os.IsNotExist(err) {
		t.Errorf("File test.yaml was not created")
	} else {
		// Remove the file after the test
		err := os.Remove("test.yaml")
		if err != nil {
			t.Errorf("Error removing test.yaml: %v", err)
		}
	}
}

func Test_Failed_DNSCleanerDownload(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	t.Setenv("CLOUDFLARE_ZONE_NAME", "example.com")
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--trace", "dns-cleaner", "download", "--dns-file", "test.yaml"})
	assert.EqualError(t, err, "zone could not be found", "Expected error when running the app with dns-cleaner download command with missing zone name")
}

func Test_DNSCleaner(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	zone := os.Getenv("CLOUDFLARE_ZONE_NAME")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	if zone == "" {
		t.Skip("CLOUDFLARE_ZONE_NAME environment variable not set")
	}
	makeTestRecords(t)
	app := buildApp()
	outputFileName := "complete.yaml"
	err := app.Run(t.Context(), []string{"cloudflare-utils", "dns-cleaner", "--zone-name", zone, "--dns-file", outputFileName, "--no-keep"})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner download command")
	// Check if the file was created
	if _, err := os.Stat(outputFileName); os.IsNotExist(err) {
		t.Errorf("File %s was not created", outputFileName)
	}
	app = buildApp()
	err = app.Run(t.Context(), []string{"cloudflare-utils", "--trace", "dns-cleaner", "--zone-name", zone, "--dns-file", outputFileName})
	assert.NoError(t, err, "Expected no error when running the app with dns-cleaner upload command")

	removeErr := os.Remove(outputFileName)
	assert.NoError(t, removeErr, "Expected no error when removing the output file")
}
