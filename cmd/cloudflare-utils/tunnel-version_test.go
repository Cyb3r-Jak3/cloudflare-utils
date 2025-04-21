package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TunnelVersion(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	account := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	if account == "" {
		t.Skip("CLOUDFLARE_ACCOUNT_ID environment variable not set")
	}
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--debug", "tunnel-versions"})
	assert.NoError(t, err, "Expected no error when running the app with tunnel-versions command")
}

func Test_TunnelVersionActive(t *testing.T) {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	account := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if token == "" {
		t.Skip("CLOUDFLARE_API_TOKEN environment variable not set")
	}
	if account == "" {
		t.Skip("CLOUDFLARE_ACCOUNT_ID environment variable not set")
	}
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--verbose", "tunnel-versions", "--healthy-only"})
	assert.NoError(t, err, "Expected no error when running the app with tunnel-versions command")
}
