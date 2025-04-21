package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_DNSPurge(t *testing.T) {
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
	err := app.Run(t.Context(), []string{"cloudflare-utils", "dns-purge", "--zone-name", zone, "--confirm"})
	assert.NoError(t, err, "Expected no error when running the app with dns-purge command")
}
