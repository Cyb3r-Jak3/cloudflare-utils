package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestAppBuild(t *testing.T) {
	// Test the app build function
	app := buildApp()
	if app.Name != "cloudflare-utils" {
		t.Errorf("Expected app name 'cloudflare-utils', got '%s'", app.Name)
	}
	if len(app.Commands) == 0 {
		t.Error("Expected at least one command in the app")
	}
	if len(app.Flags) == 0 {
		t.Error("Expected at least one flag in the app")
	}
	err := app.Run(context.Background(), []string{"missing"})
	assert.NoError(t, err, "Expected no error when running the app with missing command")
}

func Test_Basic_Flags(t *testing.T) {
	// Test the basic functionality of the app
	app := buildApp()
	err := app.Run(context.Background(), []string{"--debug", "--rate-limit", "5"})
	assert.NoError(t, err, "Expected no error when running the app with missing command")
}

func Test_GenDocs(t *testing.T) {
	// Test the documentation generation
	app := buildApp()
	err := app.Run(context.Background(), []string{"cloudflare-utils", "generate-doc"})
	assert.NoError(t, err, "Expected no error when running the app with generate-doc command")
}

func Test_GlobalAuth(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_EMAIL", "example@example.com")
	t.Setenv("CLOUDFLARE_API_KEY", "examplekey")
	app := buildApp()
	err := app.Run(context.Background(), []string{"cloudflare-utils", "tunnel-versions"})
	assert.EqualError(t, err, "Unable to authenticate request (10001)", "Expected error when running the app with tunnel-versions command")
}
