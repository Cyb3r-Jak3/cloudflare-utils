package main

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func buildTestClient(t *testing.T) *cloudflare.API {
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		t.Errorf("Failed to create Cloudflare API client: %v", err)
	}
	return api
}

func makeTestRecords(t *testing.T) {
	api := buildTestClient(t)
	zoneID, err := api.ZoneIDByName(os.Getenv("CLOUDFLARE_ZONE_NAME"))
	zoneRC := cloudflare.ResourceIdentifier(zoneID)
	if err != nil {
		t.Errorf("Failed to get zone ID: %v", err)
	}
	dnsRecords := []cloudflare.DNSRecord{
		{
			Type:    "A",
			Name:    "test1",
			Content: "127.0.0.1",
		},
	}
	for _, record := range dnsRecords {
		_, createErr := api.CreateDNSRecord(context.Background(), zoneRC, cloudflare.CreateDNSRecordParams{
			Type:    record.Type,
			Name:    record.Name,
			Content: record.Content,
		})
		if createErr != nil {
			if strings.Contains(createErr.Error(), "An identical record already exists") {
				t.Logf("DNS record already exists: %s", createErr)
				continue
			} else {
				t.Errorf("Failed to create DNS record: %v", createErr)
			}
		}
	}
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
	err := app.Run(t.Context(), []string{"missing"})
	assert.NoError(t, err, "Expected no error when running the app with missing command")
}

func Test_Basic_Flags(t *testing.T) {
	// Test the basic functionality of the app
	app := buildApp()
	err := app.Run(t.Context(), []string{"--debug", "--rate-limit", "5"})
	assert.NoError(t, err, "Expected no error when running the app with missing command")
}

func Test_GenDocs(t *testing.T) {
	// Test the documentation generation
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "generate-doc"})
	assert.NoError(t, err, "Expected no error when running the app with generate-doc command")
}

func Test_GlobalAuth(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_EMAIL", "example@example.com")
	t.Setenv("CLOUDFLARE_API_KEY", "examplekey")
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "tunnel-versions"})
	assert.EqualError(t, err, "Unable to authenticate request (10001)", "Expected error when running the app with tunnel-versions command")
}
