package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PruneDeployments_Branch(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "prune-deployments", "--project", "cloudflare-utils-pages-project", "--branch", "main"})
	assert.NoError(t, err, "Expected no error when running the app with dry-run flag")
}

func Test_PruneDeployments_TimeBefore(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "prune-deployments", "--project", "cloudflare-utils-pages-project", "--before", "2006-01-02T15:04:05"})
	assert.NoError(t, err, "Expected no error when running the app with dry-run flag")
}

func Test_PruneDeployments_TimeAfter(t *testing.T) {
	err := withApp(t, []string{"cloudflare-utils", "prune-deployments", "--project", "cloudflare-utils-pages-project", "--after", "2006-01-02T15:04:05"})
	assert.NoError(t, err, "Expected no error when running the app with dry-run flag")
}
