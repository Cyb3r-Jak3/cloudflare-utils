package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DNSPurge(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "dns-purge", "--zone-name", "2", "--confirm"})
	assert.NoError(t, err, "Expected no error when running the app with dns-purge command")
}
