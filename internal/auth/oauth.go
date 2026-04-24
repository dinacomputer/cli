package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	legacyIssuer           = "https://dina.sh"
	legacyAuthorizationURL = legacyIssuer + "/oauth/authorize"
	legacyTokenURL         = legacyIssuer + "/oauth/token"
	legacyClientID         = "dina-cli"
	newClientID            = "d35bfdfa-f03c-481b-a3be-aaf96570611e"
)

// tokenResponse is the JSON body returned by the token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Login discovers the auth server and selects the appropriate flow.
// If the server supports RFC 8628 device authorization, it uses that.
// Otherwise it falls back to PKCE authorization code flow.
func Login() (*Credentials, error) {
	apiBase := ResolveAPIBase()

	issuer, err := DiscoverIssuer(apiBase)
	if err != nil {
		issuer = legacyIssuer
	}

	var creds *Credentials

	oidcCfg, err := DiscoverOIDC(issuer)
	if err != nil {
		creds, err = PKCELogin(legacyClientID)
	} else if oidcCfg != nil && oidcCfg.DeviceAuthorizationEndpoint != "" {
		creds, err = DeviceLogin(oidcCfg, newClientID)
	} else {
		creds, err = PKCELogin(legacyClientID)
	}
	if err != nil {
		return nil, err
	}

	// Store the API base URL so future commands know which API to use.
	creds.APIBaseURL = apiBase
	if err := SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("saving credentials: %w", err)
	}
	return creds, nil
}

// PKCELogin performs the OAuth 2.0 authorization code flow with PKCE.
// Used as a fallback for servers that don't support the device code grant.
func PKCELogin(clientID string) (*Credentials, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	pkce, err := NewPKCE()
	if err != nil {
		listener.Close()
		return nil, err
	}
	stateBuf := make([]byte, 16)
	rand.Read(stateBuf)
	state := base64.RawURLEncoding.EncodeToString(stateBuf)

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s&code_challenge=%s&code_challenge_method=S256&scope=%s",
		legacyAuthorizationURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
		url.QueryEscape(pkce.Challenge),
		url.QueryEscape("openid profile offline_access"),
	)

	fmt.Println("Opening browser to authenticate...")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authURL)
	openBrowser(authURL)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("authorization error: %s — %s", errParam, desc)
			fmt.Fprintf(w, "<html><body><h1>Authentication failed</h1><p>%s</p></body></html>", desc)
			return
		}
		c := r.URL.Query().Get("code")
		if c == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, "<html><body><h1>Authenticated!</h1><p>You can close this tab.</p></body></html>")
		codeCh <- c
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(2 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("timed out waiting for authentication")
	}
	server.Shutdown(context.Background())

	fmt.Println("Exchanging authorization code for tokens...")
	tok, err := exchangeCode(clientID, code, redirectURI, pkce.Verifier)
	if err != nil {
		return nil, err
	}

	result := &Credentials{
		ClientID:     clientID,
		Issuer:       legacyIssuer,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
	}
	if err := SaveCredentials(result); err != nil {
		return nil, fmt.Errorf("saving credentials: %w", err)
	}
	return result, nil
}

// RefreshAccessToken uses a refresh token to obtain a new access token.
func RefreshAccessToken(creds *Credentials) (*Credentials, error) {
	if creds.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available, please run: dina auth login")
	}

	tokenURL := resolveTokenURL(creds.Issuer)

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {creds.ClientID},
		"refresh_token": {creds.RefreshToken},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("refresh failed (status %d), please run: dina auth login", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}

	creds.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		creds.RefreshToken = tok.RefreshToken
	}
	creds.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)

	if err := SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("saving refreshed credentials: %w", err)
	}
	return creds, nil
}

// resolveTokenURL determines the token endpoint for a given issuer.
func resolveTokenURL(issuer string) string {
	if issuer != "" {
		cfg, err := DiscoverOIDC(issuer)
		if err == nil && cfg != nil && cfg.TokenEndpoint != "" {
			return cfg.TokenEndpoint
		}
	}
	return legacyTokenURL
}

func exchangeCode(clientID, code, redirectURI, codeVerifier string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}

	resp, err := http.PostForm(legacyTokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		cmd = "cmd"
		args = []string{"/c", "start", strings.ReplaceAll(url, "&", "^&")}
	}
	if len(args) == 0 {
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}

