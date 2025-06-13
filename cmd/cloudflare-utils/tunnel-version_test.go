package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TunnelVersion(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "tunnel-versions"})
	assert.NoError(t, err, "Expected no error when running the app with tunnel-versions command")
}

func Test_TunnelVersionActive(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "--verbose", "tunnel-versions", "--healthy-only"})
	assert.NoError(t, err, "Expected no error when running the app with tunnel-versions command")
}
