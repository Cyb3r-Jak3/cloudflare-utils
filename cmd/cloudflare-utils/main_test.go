package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	setupTestHTTPServer(t)
	defer teardownTestHTTPServer()
	t.Setenv("CLOUDFLARE_API_KEY", "exampleKey")
	t.Setenv("CLOUDFLARE_API_EMAIL", "exampleEmail")
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "tunnel-versions"})
	assert.NoError(t, err, "Expected no error when running the app with tunnel-versions command and global auth")
}

func Test_BadAPIPermission(t *testing.T) {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "1")
	t.Setenv("CLOUDFLARE_API_TOKEN", "exampletoken")
	t.Setenv("CLOUDFLARE_BASE_URL", server.URL)
	verifyHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "ed17574386854bf78a67040be0a770b0",
    "status": "active",
    "not_before": "2018-07-01T05:20:00Z",
    "expires_on": "2020-01-01T00:00:00Z"
  }
}`)
	}
	tokenBadPermissionsHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "ed17574386854bf78a67040be0a770b0",
    "name": "readonly token",
    "status": "active",
    "issued_on": "2018-07-01T05:20:00Z",
    "modified_on": "2018-07-02T05:20:00Z",
    "not_before": "2018-07-01T05:20:00Z",
    "expires_on": "2020-01-01T00:00:00Z",
    "policies": [
      {
        "id": "f267e341f3dd4697bd3b9f71dd96247f",
        "effect": "allow",
        "resources": {
          "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4": "*",
          "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43": "*"
        },
        "permission_groups": []
      }
    ],
    "condition": {}
  }
}`)
	}
	mux.HandleFunc("/user/tokens/verify", verifyHandler)
	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0", tokenBadPermissionsHandler)
	app := buildApp()
	err := app.Run(t.Context(), []string{"cloudflare-utils", "tunnel-versions"})
	assert.Error(t, err, "Expected an error when running the app with insufficient permissions")
	assert.Contains(t, err.Error(), "API Token does not have permission [TunnelRead]")
}

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setupTestHTTPServer sets up a test HTTP server with mock API responses.
// It is used to simulate the Cloudflare API for testing purposes.
// All test responses are pulled from cloudflare-go.
func setupTestHTTPServer(t *testing.T) {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// disable rate limits and retries in testing - prepended so any provided value overrides this
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "1")
	t.Setenv("CLOUDFLARE_API_TOKEN", "exampletoken")
	t.Setenv("CLOUDFLARE_BASE_URL", server.URL)
	t.Setenv("LOG_LEVEL_TRACE", "true")
	t.Setenv("CLOUDFLARE_ZONE_ID", "2")
	verifyHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "ed17574386854bf78a67040be0a770b0",
    "status": "active",
    "not_before": "2018-07-01T05:20:00Z",
    "expires_on": "2020-01-01T00:00:00Z"
  }
}`)
	}
	tokenPermissionsHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "ed17574386854bf78a67040be0a770b0",
    "name": "readonly token",
    "status": "active",
    "issued_on": "2018-07-01T05:20:00Z",
    "modified_on": "2018-07-02T05:20:00Z",
    "not_before": "2018-07-01T05:20:00Z",
    "expires_on": "2020-01-01T00:00:00Z",
    "policies": [
      {
        "id": "f267e341f3dd4697bd3b9f71dd96247f",
        "effect": "allow",
        "resources": {
          "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4": "*",
          "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43": "*"
        },
        "permission_groups": [
          {
            "id": "8d28297797f24fb8a0c332fe0866ec89",
            "name": "Pages Write"
          },
          {
            "id": "4755a26eedb94da69e1066d98aa820be",
            "name": "DNS Write"
          },
          {
            "id": "efea2ab8357b47888938f101ae5e053f",
            "name": "Tunnel Read"
          },
          {
            "id": "c07321b023e944ff818fec44d8203567",
            "name": "Tunnel Write"
          }
        ]
      }
    ],
    "condition": {}
  }
}`)
	}
	pagesDeploymentPage1Handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
				"id": "0012e50b-fa5d-44db-8cb5-1f372785dcbe",
				"short_id": "0012e50b",
				"project_id": "80776025-b1bd-4181-993f-8238c27d226f",
				"project_name": "test",
				"environment": "production",
				"url": "https://0012e50b.test.pages.dev",
				"created_on": "2021-01-01T00:00:00Z",
				"modified_on": "2021-01-01T00:00:00Z",
				"latest_stage": {
					"name": "deploy",
					"started_on": "2021-01-01T00:00:00Z",
					"ended_on": "2021-01-01T00:00:00Z",
					"status": "success"
				},
				"deployment_trigger": {
					"type": "ad_hoc",
					"metadata": {
						"branch": "main",
						"commit_hash": "20fb65fa9d7fd2a11f7fa3ebdc44137b263ee835",
						"commit_message": "Test commit"
					}
				},
				"stages": [
					{
						"name": "queued",
						"started_on": "2021-01-01T00:00:00Z",
						"ended_on": "2021-01-01T00:00:00Z",
						"status": "success"
					},
					{
						"name": "initialize",
						"started_on": "2021-01-01T00:00:00Z",
						"ended_on": "2021-01-01T00:00:00Z",
						"status": "success"
					},
					{
						"name": "clone_repo",
						"started_on": "2021-01-01T00:00:00Z",
						"ended_on": "2021-01-01T00:00:00Z",
						"status": "success"
					},
					{
						"name": "build",
						"started_on": "2021-01-01T00:00:00Z",
						"ended_on": "2021-01-01T00:00:00Z",
						"status": "success"
					},
					{
						"name": "deploy",
						"started_on": "2021-01-01T00:00:00Z",
						"ended_on": "2021-01-01T00:00:00Z",
						"status": "success"
					}
				],
				"build_config": {
					"build_command": "bash test.sh",
					"destination_dir": "",
					"root_dir": "",
					"web_analytics_tag": null,
					"web_analytics_token": null
				},
				"source": {
					"type": "github",
					"config": {
						"owner": "coudflare",
						"repo_name": "Test",
						"production_branch": "main",
						"pr_comments_enabled": false
					}
				},
				"env_vars": {
					"NODE_VERSION": {
						"value": "16"
					}
				},
				"aliases": null
			}
					],
			"result_info": {
				"page": 1,
				"per_page": 100,
				"count": 1,
				"total_pages": 1
			  }
				}`)
	}
	dnsRecordsPage1Handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "372e67954025e0ba6aaa6d586b9e0b59",
					"type": "A",
					"name": "example.com",
					"content": "198.51.100.4",
					"proxiable": true,
					"proxied": false,
					"ttl": 120,
					"created_on": "2014-01-01T05:20:00Z",
					"modified_on": "2014-01-01T05:20:00Z",
					"data": {},
					"meta": {
						"auto_added": true,
						"source": "primary"
					}
				}
			],
			"result_info": {
				"count": 1,
				"page": 1,
				"per_page": 20,
				"total_count": 1
			}
		}`)
	}
	zoneLookupHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if r.URL.Query().Get("name") == "example.com" {
			fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
				"id": "2",
				"name": "example.com",
				"status": "active",
				"type": "full",
				"development_mode": 0,
				"created_on": "2014-01-01T05:20:00Z",
				"modified_on": "2014-01-01T05:20:00Z",
				"activated_on": "2014-01-01T05:20:00Z",
				"meta": {
					"step": 1,
					"sld": "example",
					"tld": "com"
				}
			}]
		}`)
		}
		if r.URL.Query().Get("name") == "nonexistent" {
			fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		}`)
		}
	}
	dnsRecordDeleteHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected a DELETE request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "372e67954025e0ba6aaa6d586b9e0b59"
			}
		}`)
	}
	tunnelListHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": [
			{
			  "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
			  "name": "blog",
			  "created_at": "2009-11-10T23:00:00Z",
			  "deleted_at": "2009-11-11T23:00:00Z",
			  "status": "healthy",
			  "connections": [
				{
				  "colo_name": "DFW",
				  "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				  "is_pending_reconnect": false,
				  "client_id": "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
				  "client_version": "2022.2.0",
				  "opened_at": "2021-01-25T18:22:34.317854Z",
				  "origin_ip": "198.51.100.1"
				}
			  ]
			},
			{
			  "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8416",
			  "name": "blog-backup",
			  "created_at": "2009-11-10T23:00:00Z",
			  "deleted_at": "2009-11-12T23:00:00Z",
			  "status": "unhealthy",
			  "connections": [
				{
				  "colo_name": "IAD",
				  "id": "f174e90a-fafe-4643-bbbc-4a0ed4fc8416",
				  "is_pending_reconnect": false,
				  "client_id": "dc6472cc-f1ae-44a0-b795-6b8a0ce29f90",
				  "client_version": "2022.2.0",
				  "opened_at": "2021-01-25T18:22:34.317854Z",
				  "origin_ip": "198.51.100.2"
				}
			  ]
			}
		  ],
		  "result_info": {
			"count": 2,
			"page": 1,
			"per_page": 20,
			"total_count": 2
		  }
		}`)
	}
	deletePagesDeploymentHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected a DELETE request")
		assert.Equal(t, "true", r.URL.Query().Get("force"))
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": null
		}`)
	}
	deletePagesProjectHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected a DELETE request")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": null
		}`)
	}
	mux.HandleFunc("/test-ips.txt", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request for test-ips.txt")
		w.Header().Set("content-type", "text/plain")
		fmt.Fprint(w, "1.2.3.4\n2001:db8::1\n8.8.8.8\n")
	})
	mux.HandleFunc("/user/tokens/verify", verifyHandler)
	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0", tokenPermissionsHandler)
	mux.HandleFunc("/accounts/1/pages/projects/cloudflare-utils-pages-project/deployments", pagesDeploymentPage1Handler)
	mux.HandleFunc("/zones/2/dns_records", dnsRecordsPage1Handler)
	mux.HandleFunc("/zones/", zoneLookupHandler)
	mux.HandleFunc("/zones/2/dns_records/372e67954025e0ba6aaa6d586b9e0b59", dnsRecordDeleteHandler)
	mux.HandleFunc("/accounts/1/cfd_tunnel", tunnelListHandler)
	mux.HandleFunc("/accounts/1/pages/projects/cloudflare-utils-pages-project/deployments/0012e50b-fa5d-44db-8cb5-1f372785dcbe", deletePagesDeploymentHandler)
	mux.HandleFunc("/accounts/1/pages/projects/cloudflare-utils-pages-project", deletePagesProjectHandler)
	mux.HandleFunc("/ips?china_colo=1", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected a GET request for /ips")
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"ipv4_cidrs": [
					"1.1.1.1"
				],
				"ipv6_cidrs": [
					"2606:4700:4700::1111"
				],
				"china_ipv4_cidrs": [
					"2.2.2.2"
				],
				"china_ipv6_cidrs": [
					"2606:4700:4700::2222"
				]
			}
		}`)
	})
	mux.HandleFunc("/accounts/1/rules/lists", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
				{
					"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
					"name": "test-list",
					"description": "This is a note.",
					"kind": "ip",
					"num_items": 10,
					"num_referencing_filters": 2,
					"created_on": "2020-01-01T08:00:00Z",
					"modified_on": "2020-01-10T14:00:00Z"
				}
			],
			"success": true,
			"errors": [],
			"messages": []
		}`)
	})
	mux.HandleFunc("/accounts/1/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	})
	mux.HandleFunc("/accounts/1/rules/lists/bulk_operations/4da8780eeb215e6cb7f48dd981c4ea02", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "4da8780eeb215e6cb7f48dd981c4ea02",
				"status": "completed",
				"error": "",
				"completed": "2020-01-01T08:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	})
}

func teardownTestHTTPServer() {
	server.Close()
}
