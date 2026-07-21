package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	oauthCallbackPort = 8976
	oauthCallbackPath = "/oauth/callback"
	oauthLoginTimeout = 5 * time.Minute
	oauthClientID     = "0d52ae70cc5a9d93db44cb8874bb17b5"
)

// These are overridden in tests to point at a local httptest server instead
// of the real Cloudflare OAuth endpoints.
var (
	oauthAuthURL   = "https://dash.cloudflare.com/oauth2/auth"
	oauthTokenURL  = "https://dash.cloudflare.com/oauth2/token" //nolint: gosec
	oauthRevokeURL = "https://dash.cloudflare.com/oauth2/revoke"
)

// generateOauthToken runs the OAuth2 authorization code flow with PKCE.
//
// Cloudflare's authorization server redirects the user's browser to a
// pre-registered redirect_uri with the code in the query string, so a local
// HTTP listener is required to receive it. When creating the OAuth client at
// https://developers.cloudflare.com/fundamentals/oauth/create-an-oauth-client/,
// register the redirect URI printed below (http://localhost:<oauthCallbackPort><oauthCallbackPath>).
func generateOauthToken(ctx context.Context) (*oauth2.Token, error) {
	redirectURI := fmt.Sprintf("http://localhost:%d%s", oauthCallbackPort, oauthCallbackPath)

	conf := &oauth2.Config{
		ClientID:    oauthClientID,
		RedirectURL: redirectURI,
		Scopes:      []string{"page.write", "dns.write", "zone.read", "account-rule-lists.write", "teams-connectors.read", "cache.purge"},
		Endpoint: oauth2.Endpoint{ //nolint:gosec // URLs are Cloudflare's public OAuth endpoints, not credentials
			AuthURL:  oauthAuthURL,
			TokenURL: oauthTokenURL,
		},
	}

	state, err := randomString(32)
	if err != nil {
		return nil, fmt.Errorf("error generating state: %w", err)
	}

	// PKCE protects the exchange itself; state protects against CSRF on the callback.
	// https://www.ietf.org/archive/id/draft-ietf-oauth-security-topics-22.html#name-countermeasures-6
	verifier := oauth2.GenerateVerifier()

	type callbackResult struct {
		code string
		err  error
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(oauthCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if errParam := query.Get("error"); errParam != "" {
			writeCallbackPage(w, false)
			resultCh <- callbackResult{err: fmt.Errorf("authorization server returned error: %s: %s", errParam, query.Get("error_description"))}
			return
		}

		if query.Get("state") != state {
			writeCallbackPage(w, false)
			resultCh <- callbackResult{err: errors.New("state mismatch in oauth callback")}
			return
		}

		code := query.Get("code")
		if code == "" {
			writeCallbackPage(w, false)
			resultCh <- callbackResult{err: errors.New("no code in oauth callback")}
			return
		}

		writeCallbackPage(w, true)
		resultCh <- callbackResult{code: code}
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", oauthCallbackPort))
	if err != nil {
		return nil, fmt.Errorf("error starting local oauth callback listener on port %d: %w", oauthCallbackPort, err)
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	fmt.Printf("Open the following URL in a browser to authorize this application:\n%s\n", authURL)

	loginCtx, cancel := context.WithTimeout(ctx, oauthLoginTimeout)
	defer cancel()

	select {
	case <-loginCtx.Done():
		return nil, fmt.Errorf("timed out waiting for oauth callback: %w", loginCtx.Err())
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}
		tok, err := conf.Exchange(ctx, result.code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("error getting token: %w", err)
		}
		return tok, nil
	}
}

// revokeOauthToken revokes an access or refresh token per RFC 7009.
// https://developers.cloudflare.com/fundamentals/oauth/integrate-with-cloudflare/
func revokeOauthToken(ctx context.Context, token string) error {
	form := url.Values{
		"token":           {token},
		"token_type_hint": {"access_token"},
		"client_id":       {oauthClientID},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthRevokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error building revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending revoke request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func randomString(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeCallbackPage(w http.ResponseWriter, success bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if success {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "<html><body><h1>Authorization complete</h1><p>You can close this tab and return to the terminal.</p></body></html>")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "<html><body><h1>Authorization failed</h1><p>Check the terminal for details.</p></body></html>")
	}
}
