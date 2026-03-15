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
	issuer           = "https://dina.sh"
	authorizationURL = issuer + "/oauth/authorize"
	tokenURL         = issuer + "/oauth/token"
)

// tokenResponse is the JSON body returned by the token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Login performs the full OAuth 2.0 authorization code flow with PKCE.
func Login() (*Credentials, error) {
	// 1. Start local callback server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 2. Use fixed client ID.
	clientID := "dina-cli"

	// 3. Generate PKCE + state.
	pkce, err := NewPKCE()
	if err != nil {
		listener.Close()
		return nil, err
	}
	stateBuf := make([]byte, 16)
	rand.Read(stateBuf)
	state := base64.RawURLEncoding.EncodeToString(stateBuf)

	// 4. Build authorization URL and open browser.
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s&code_challenge=%s&code_challenge_method=S256&scope=%s",
		authorizationURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
		url.QueryEscape(pkce.Challenge),
		url.QueryEscape("openid profile offline_access"),
	)

	fmt.Println("Opening browser to authenticate...")
	fmt.Printf("If the browser doesn't open, visit:\n  %s\n\n", authURL)
	openBrowser(authURL)

	// 5. Wait for callback.
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
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, "<html><body><h1>Authenticated!</h1><p>You can close this tab.</p></body></html>")
		codeCh <- code
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

	// 6. Exchange code for tokens.
	fmt.Println("Exchanging authorization code for tokens...")
	tok, err := exchangeCode(clientID, code, redirectURI, pkce.Verifier)
	if err != nil {
		return nil, err
	}

	result := &Credentials{
		ClientID:     clientID,
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

func exchangeCode(clientID, code, redirectURI, codeVerifier string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}

	resp, err := http.PostForm(tokenURL, data)
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
