package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// deviceAuthResponse is the JSON body returned by the device authorization endpoint.
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// deviceTokenError is the error response during device code polling.
type deviceTokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// DeviceLogin performs the OAuth 2.0 Device Authorization Grant (RFC 8628).
func DeviceLogin(cfg *OIDCConfig, clientID string) (*Credentials, error) {
	data := url.Values{
		"client_id": {clientID},
		"scope":     {"openid profile email offline_access"},
	}

	resp, err := http.PostForm(cfg.DeviceAuthorizationEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("device authorization failed with status %d", resp.StatusCode)
	}

	var dar deviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&dar); err != nil {
		return nil, fmt.Errorf("decoding device authorization response: %w", err)
	}

	fmt.Printf("\nTo authenticate, visit:\n\n  %s\n\n", dar.VerificationURI)
	fmt.Printf("and enter code: %s\n\n", dar.UserCode)

	if dar.VerificationURIComplete != "" {
		fmt.Printf("Or open this URL directly:\n  %s\n\n", dar.VerificationURIComplete)
		openBrowser(dar.VerificationURIComplete)
	}

	fmt.Println("Waiting for authorization...")

	interval := dar.Interval
	if interval <= 0 {
		interval = 5
	}
	deadline := time.Now().Add(time.Duration(dar.ExpiresIn) * time.Second)

	for {
		time.Sleep(time.Duration(interval) * time.Second)

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device code expired — please try again")
		}

		tok, retry, err := pollDeviceToken(cfg.TokenEndpoint, clientID, dar.DeviceCode)
		if err != nil {
			return nil, err
		}
		if retry == "slow_down" {
			interval += 5
			continue
		}
		if retry == "authorization_pending" {
			continue
		}

		result := &Credentials{
			ClientID:     clientID,
			Issuer:       cfg.Issuer,
			AccessToken:  tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
		}
		if err := SaveCredentials(result); err != nil {
			return nil, fmt.Errorf("saving credentials: %w", err)
		}
		return result, nil
	}
}

// pollDeviceToken makes a single poll request to the token endpoint.
// Returns (token, "", nil) on success, (nil, retryReason, nil) if polling should continue,
// or (nil, "", error) on fatal errors.
func pollDeviceToken(tokenEndpoint, clientID, deviceCode string) (*tokenResponse, string, error) {
	data := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_id":   {clientID},
		"device_code": {deviceCode},
	}

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, "", fmt.Errorf("token poll failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var tok tokenResponse
		if err := json.Unmarshal(body, &tok); err != nil {
			return nil, "", fmt.Errorf("decoding token response: %w", err)
		}
		return &tok, "", nil
	}

	var dte deviceTokenError
	if err := json.Unmarshal(body, &dte); err == nil {
		switch dte.Error {
		case "authorization_pending":
			return nil, "authorization_pending", nil
		case "slow_down":
			return nil, "slow_down", nil
		case "expired_token":
			return nil, "", fmt.Errorf("device code expired — please try again")
		case "access_denied":
			return nil, "", fmt.Errorf("authorization denied by user")
		}
	}

	return nil, "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
}
