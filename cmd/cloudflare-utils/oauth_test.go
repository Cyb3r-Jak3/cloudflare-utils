package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callbackURLFromStdout runs fn while capturing os.Stdout and returns the
// authorization URL printed by generateOauthToken.
func callbackURLFromStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	done := make(chan struct{})
	var output []byte
	go func() {
		output, _ = io.ReadAll(r)
		close(done)
	}()

	fn()

	_ = w.Close()
	<-done
	os.Stdout = orig

	re := regexp.MustCompile(`https?://\S+`)
	match := re.Find(output)
	require.NotNil(t, match, "expected an authorization URL to be printed, got: %s", output)
	return string(match)
}

func stateFromAuthURL(t *testing.T, authURL string) string {
	t.Helper()
	parsed, err := url.Parse(authURL)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state, "expected state query param in auth URL")
	return state
}

func Test_RandomString(t *testing.T) {
	s1, err := randomString(32)
	require.NoError(t, err)
	assert.Len(t, s1, 64) // hex encoding doubles byte length

	s2, err := randomString(32)
	require.NoError(t, err)
	assert.NotEqual(t, s1, s2, "expected two random strings to differ")
}

func Test_WriteCallbackPage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeCallbackPage(rec, true)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Authorization complete")
	})

	t.Run("failure", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeCallbackPage(rec, false)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Authorization failed")
	})
}

// setupOAuthTestServer points the OAuth auth/token endpoints at a local
// httptest server and restores the originals on test cleanup.
func setupOAuthTestServer(t *testing.T, tokenHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", tokenHandler)
	srv := httptest.NewServer(mux)

	origAuth, origToken := oauthAuthURL, oauthTokenURL
	oauthAuthURL = srv.URL + "/oauth2/auth"
	oauthTokenURL = srv.URL + "/oauth2/token"

	t.Cleanup(func() {
		srv.Close()
		oauthAuthURL = origAuth
		oauthTokenURL = origToken
	})
	return srv
}

func Test_GenerateOauthToken_Success(t *testing.T) {
	setupOAuthTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"test-access-token","token_type":"bearer","refresh_token":"test-refresh-token"}`)
	})

	type result struct {
		token string
		err   error
	}
	resultCh := make(chan result, 1)

	authURL := callbackURLFromStdout(t, func() {
		go func() {
			tok, err := generateOauthToken(t.Context())
			if err != nil {
				resultCh <- result{err: err}
				return
			}
			resultCh <- result{token: tok.AccessToken}
		}()
		// Give the listener a moment to start before hitting it.
		time.Sleep(100 * time.Millisecond)
	})

	state := stateFromAuthURL(t, authURL)
	callbackURL := fmt.Sprintf("http://localhost:%d%s?state=%s&code=test-code", oauthCallbackPort, oauthCallbackPath, state)
	resp, err := http.Get(callbackURL) //nolint:gosec // localhost URL built from a fixed const port/path in test
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	select {
	case res := <-resultCh:
		require.NoError(t, res.err)
		assert.Equal(t, "test-access-token", res.token)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for generateOauthToken to return")
	}
}

func Test_GenerateOauthToken_CallbackErrors(t *testing.T) {
	setupOAuthTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("token endpoint should not be called")
	})

	testCases := []struct {
		name      string
		query     func(state string) string
		errSubstr string
	}{
		{
			name:      "authorization server error",
			query:     func(state string) string { return "error=access_denied&error_description=user+said+no" },
			errSubstr: "authorization server returned error: access_denied: user said no",
		},
		{
			name:      "state mismatch",
			query:     func(state string) string { return "state=wrong-state&code=abc" },
			errSubstr: "state mismatch in oauth callback",
		},
		{
			name:      "missing code",
			query:     func(state string) string { return "state=" + state },
			errSubstr: "no code in oauth callback",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			type result struct {
				err error
			}
			resultCh := make(chan result, 1)

			authURL := callbackURLFromStdout(t, func() {
				go func() {
					_, err := generateOauthToken(t.Context())
					resultCh <- result{err: err}
				}()
				time.Sleep(100 * time.Millisecond)
			})

			state := stateFromAuthURL(t, authURL)
			callbackURL := fmt.Sprintf("http://localhost:%d%s?%s", oauthCallbackPort, oauthCallbackPath, tc.query(state))
			resp, err := http.Get(callbackURL) //nolint:gosec // localhost URL built from a fixed const port/path in test
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			select {
			case res := <-resultCh:
				require.Error(t, res.err)
				assert.Contains(t, res.err.Error(), tc.errSubstr)
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for generateOauthToken to return")
			}
		})
	}
}

func Test_GenerateOauthToken_Timeout(t *testing.T) {
	setupOAuthTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("token endpoint should not be called")
	})

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	type result struct {
		err error
	}
	resultCh := make(chan result, 1)

	_ = callbackURLFromStdout(t, func() {
		tok, err := generateOauthToken(ctx)
		_ = tok
		resultCh <- result{err: err}
	})

	res := <-resultCh
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "timed out waiting for oauth callback")
}

func Test_RevokeOauthToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/oauth2/revoke", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "test-token", r.FormValue("token"))
			assert.Equal(t, "access_token", r.FormValue("token_type_hint"))
			assert.Equal(t, oauthClientID, r.FormValue("client_id"))
			w.WriteHeader(http.StatusOK)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		orig := oauthRevokeURL
		oauthRevokeURL = srv.URL + "/oauth2/revoke"
		defer func() { oauthRevokeURL = orig }()

		err := revokeOauthToken(t.Context(), "test-token")
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/oauth2/revoke", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "invalid_token")
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		orig := oauthRevokeURL
		oauthRevokeURL = srv.URL + "/oauth2/revoke"
		defer func() { oauthRevokeURL = orig }()

		err := revokeOauthToken(t.Context(), "test-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "revoke request failed with status 400")
		assert.Contains(t, err.Error(), "invalid_token")
	})
}
