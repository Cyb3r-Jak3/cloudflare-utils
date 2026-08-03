package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Cyb3r-Jak3/common/v5"
	"golang.org/x/oauth2"
)

var webOAuthBase = "https://cloudflare-utils.cyberjake.xyz/oauth/"

var (
	pollForTokenInitialDelay = 1 * time.Second
	pollForTokenInterval     = 5 * time.Second
)

type OauthStatus string

const (
	OauthPending OauthStatus = "pending"
	OauthSuccess OauthStatus = "success"
)

type OAuthV1RegistrationResponse struct {
	RegistrationID string `json:"registration_id"`
	URL            string `json:"url"`
	ExpiresIN      int    `json:"expires_in"`
}

type OAuthV1Response struct {
	Status      OauthStatus `json:"status"`
	AccessToken string      `json:"access_token,omitempty"`
	ExpiresAt   int64       `json:"expires_at,omitempty"`
}

func GetWebOauthToken(ctx context.Context) (*oauth2.Token, error) {
	var registrationResponse OAuthV1RegistrationResponse
	resp, err := common.DoJSONRequestWithContext(ctx, "POST", webOAuthBase+"register", nil, &registrationResponse)
	if err != nil {
		return &oauth2.Token{}, fmt.Errorf("error getting web oauth token: %v", err)
	}
	err = resp.Body.Close()
	if err != nil {
		logger.WithError(err).Warning("error closing response body")
	}
	timeoutTime := time.Now().Add(time.Duration(registrationResponse.ExpiresIN) * time.Second)
	fmt.Printf("Web oauth token started. Please visit\n%s\nThis request expires at: %s\n", registrationResponse.URL, timeoutTime.Format(time.DateTime))
	deadline, cancel := context.WithDeadline(ctx, timeoutTime)
	defer cancel()
	response, err := PollForToken(deadline, registrationResponse.RegistrationID)
	if err != nil {
		return &oauth2.Token{}, fmt.Errorf("error getting web oauth token: %v", err)
	}
	return response, nil
}

func PollForToken(ctx context.Context, registrationID string) (*oauth2.Token, error) {
	time.Sleep(pollForTokenInitialDelay)
	ticker := time.NewTicker(pollForTokenInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Warning("PollForToken stopped due to context canceled")
			return &oauth2.Token{}, ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%stoken/%s", webOAuthBase, registrationID), nil)
			if err != nil {
				return &oauth2.Token{}, fmt.Errorf("error creating poll request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", fmt.Sprintf("cloudflare-utils/%s", version))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return &oauth2.Token{}, fmt.Errorf("error polling web oauth token: %v", err)
			}
			if resp.StatusCode == http.StatusAccepted {
				var registrationResponse OAuthV1Response
				decErr := json.NewDecoder(resp.Body).Decode(&registrationResponse)
				if decErr != nil {
					logger.Warning("PollForToken stopped due to decoding pending message error:", decErr)
					return &oauth2.Token{}, errors.New("error polling web oauth token")
				}
				if registrationResponse.Status != OauthPending {
					logger.Warning("PollForToken stopped due to polling status:", registrationResponse.Status)
					return &oauth2.Token{}, errors.New("got unknown status for pending message")
				}
				closeErr := resp.Body.Close()
				if closeErr != nil {
					logger.WithError(closeErr).Errorln("error closing response body for registration response")
				}
				continue
			} else if resp.StatusCode != http.StatusOK {
				closeErr := resp.Body.Close()
				if closeErr != nil {
					logger.WithError(closeErr).Errorln("error closing response body for registration response")
				}
				logger.WithField("response status code", resp.StatusCode).Warning("PollForToken stopped due to non-200 status code.")
				return &oauth2.Token{}, errors.New("non-200 status code from polling web oauth token")
			}
			contentType := resp.Header.Get("Content-Type")
			if contentType == "application/json+oauthv1" {
				var registrationResponse OAuthV1Response
				decErr := json.NewDecoder(resp.Body).Decode(&registrationResponse)
				if decErr != nil {
					logger.WithError(decErr).Errorln("error polling web oauth token")
					return &oauth2.Token{}, errors.New("error polling web oauth token")
				}
				if registrationResponse.Status != OauthSuccess {
					logger.WithField("registration status", registrationResponse.Status).Errorf("PollForToken stopped due to polling status")
				}
				return &oauth2.Token{
					AccessToken: registrationResponse.AccessToken,
					TokenType:   "Bearer",
					Expiry:      time.Unix(registrationResponse.ExpiresAt, 0),
				}, nil
			}
			logger.Warning("PollForToken stopped due to unexpected content type")
			return &oauth2.Token{}, errors.New("unexpected content type from polling web oauth token")
		}
	}
}
