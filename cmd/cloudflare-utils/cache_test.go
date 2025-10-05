package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Cache_BadOptions(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name:   "No options",
			args:   []string{"cloudflare-utils", "cache-cleaner"},
			errMsg: "must specify at least one purge method: --everything, --url, --tag, or --prefix",
		},
		{
			name:   "Everything with URL",
			args:   []string{"cloudflare-utils", "cache-cleaner", "--everything", "--url", "https://example.com"},
			errMsg: "cannot use --everything with --url, --tag, or --prefix",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := withApp(t, tc.args)
			assert.Error(t, err, "Expected error when running cache-cleaner command")
			if err == nil || err.Error() != tc.errMsg {
				t.Errorf("Expected error message to contain '%s', got: %v", tc.errMsg, err)
			}
		})
	}
}

func Test_Cache(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "everything",
			args: []string{"cloudflare-utils", "cache-cleaner", "--everything"},
		},
		{
			name: "Single URL",
			args: []string{"cloudflare-utils", "cache-cleaner", "--url", "https://example.com"},
		},
		{
			name: "Multiple URLs",
			args: []string{"cloudflare-utils", "cache-cleaner", "--url", "https://example.com", "--url", "https://example.org"},
		},
		{
			name: "Single Tag",
			args: []string{"cloudflare-utils", "cache-cleaner", "--tag", "tag1"},
		},
		{
			name: "Multiple Tags",
			args: []string{"cloudflare-utils", "cache-cleaner", "--tag", "tag1", "--tag", "tag2"},
		},
		{
			name: "Single Prefix",
			args: []string{"cloudflare-utils", "cache-cleaner", "--prefix", "prefix1"},
		},
		{
			name: "Multiple Prefixes",
			args: []string{"cloudflare-utils", "cache-cleaner", "--prefix", "prefix1", "--prefix", "prefix2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := withApp(t, tc.args)
			assert.NoError(t, err, fmt.Sprintf("Expected no error when running %s command", tc.name))
		})
	}
}
