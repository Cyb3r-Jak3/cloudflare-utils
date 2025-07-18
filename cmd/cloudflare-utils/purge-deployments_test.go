package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PurgeDeployments_DryRun(t *testing.T) {
	setupTestHTTPServer(t)

	defer teardownTestHTTPServer()

	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "purge-deployments", "--project", "cloudflare-utils-pages-project", "--dry-run"})
	assert.NoError(t, err, "Expected no error when running the app with dry-run flag")
}

func Test_PurgeDeployments(t *testing.T) {
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()

	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "purge-deployments", "--project", "cloudflare-utils-pages-project", "--delete-project", "--lots-of-deployments"})
	assert.NoError(t, err, "Expected no error when running the app with delete-project flag")
}
