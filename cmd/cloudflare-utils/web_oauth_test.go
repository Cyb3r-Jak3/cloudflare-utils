package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWebOAuthTestServer points webOAuthBase at a local httptest server and
// restores the original value, along with fast polling timings, on cleanup.
func setupWebOAuthTestServer(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(mux)

	origBase := webOAuthBase
	origDelay, origInterval := pollForTokenInitialDelay, pollForTokenInterval
	webOAuthBase = srv.URL + "/"
	pollForTokenInitialDelay = 1 * time.Millisecond
	pollForTokenInterval = 5 * time.Millisecond

	t.Cleanup(func() {
		srv.Close()
		webOAuthBase = origBase
		pollForTokenInitialDelay = origDelay
		pollForTokenInterval = origInterval
	})
	return srv
}

func Test_GetWebOauthToken_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"registration_id":"reg-123","url":"https://example.com/authorize","expires_in":30}`)
	})
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json+oauthv1")
		fmt.Fprintf(w, `{"status":"success","access_token":"test-token","expires_at":%d}`, time.Now().Add(time.Hour).Unix())
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := GetWebOauthToken(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "test-token", tok.AccessToken)
	assert.Equal(t, "Bearer", tok.TokenType)
}

func Test_GetWebOauthToken_RegisterError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `not-json`)
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := GetWebOauthToken(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting web oauth token")
	assert.Empty(t, tok.AccessToken)
}

func Test_PollForToken_PendingThenSuccess(t *testing.T) {
	var callCount int
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"status":"pending"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json+oauthv1")
		fmt.Fprintf(w, `{"status":"success","access_token":"final-token","expires_at":%d}`, time.Now().Add(time.Hour).Unix())
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := PollForToken(t.Context(), "reg-123")
	require.NoError(t, err)
	assert.Equal(t, "final-token", tok.AccessToken)
	assert.GreaterOrEqual(t, callCount, 3)
}

func Test_PollForToken_NonPendingStatusOn202(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"status":"success"}`)
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := PollForToken(t.Context(), "reg-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "got unknown status for pending message")
	assert.Empty(t, tok.AccessToken)
}

func Test_PollForToken_Non200Status(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := PollForToken(t.Context(), "reg-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-200 status code from polling web oauth token")
	assert.Empty(t, tok.AccessToken)
}

func Test_PollForToken_UnexpectedContentType(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "unexpected")
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := PollForToken(t.Context(), "reg-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected content type from polling web oauth token")
	assert.Empty(t, tok.AccessToken)
}

func Test_PollForToken_MalformedPendingBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `not-json`)
	})
	setupWebOAuthTestServer(t, mux)

	tok, err := PollForToken(t.Context(), "reg-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error polling web oauth token")
	assert.Empty(t, tok.AccessToken)
}

func Test_PollForToken_ContextCanceled(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token/reg-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"status":"pending"}`)
	})
	setupWebOAuthTestServer(t, mux)

	// Interval is longer than the context deadline so the ticker never fires
	// before ctx.Done() wins the select.
	pollForTokenInterval = 200 * time.Millisecond

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	tok, err := PollForToken(ctx, "reg-123")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Empty(t, tok.AccessToken)
}
