package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// FederatedOptions configures workload identity federation (OIDC token
// exchange). All fields are optional; an empty set auto-detects the CI
// provider.
type FederatedOptions struct {
	// Audience is the OIDC audience to request for the subject token. When
	// empty it defaults to the Dina issuer, so the minted token's `aud` claim
	// is bound to Dina and can't be replayed against another service.
	Audience string
	// Token is an explicit OIDC subject token, bypassing provider detection.
	Token string
	// TokenFile is a path to read the OIDC subject token from.
	TokenFile string
}

// FederatedLogin exchanges a CI-provided OIDC token for Dina credentials via
// RFC 8693 token exchange, storing them like an interactive login. The subject
// token is sourced (in order) from an explicit value, a file, the
// DINA_OIDC_TOKEN env var, or the CI provider's token service (GitHub / Forgejo
// Actions).
func FederatedLogin(opts FederatedOptions) (*Credentials, error) {
	apiBase := ResolveAPIBase()

	issuer, err := DiscoverIssuer(apiBase)
	if err != nil || issuer == "" {
		issuer = legacyIssuer
	}

	audience := opts.Audience
	if audience == "" {
		audience = issuer
	}

	subjectToken, err := obtainOIDCToken(opts, audience)
	if err != nil {
		return nil, err
	}

	tok, err := exchangeFederatedToken(issuer, subjectToken, audience)
	if err != nil {
		return nil, err
	}

	creds := &Credentials{
		ClientID:     newClientID,
		Issuer:       issuer,
		APIBaseURL:   apiBase,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken, // usually empty — CI re-federates each run
		ExpiresAt:    time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
	}
	if err := SaveCredentials(creds); err != nil {
		return nil, fmt.Errorf("saving credentials: %w", err)
	}
	return creds, nil
}

// obtainOIDCToken resolves the OIDC subject token to exchange.
func obtainOIDCToken(opts FederatedOptions, audience string) (string, error) {
	if opts.Token != "" {
		return opts.Token, nil
	}
	if opts.TokenFile != "" {
		b, err := os.ReadFile(opts.TokenFile)
		if err != nil {
			return "", fmt.Errorf("reading OIDC token file: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	if v := os.Getenv("DINA_OIDC_TOKEN"); v != "" {
		return strings.TrimSpace(v), nil
	}
	// GitHub Actions (and Forgejo/Gitea Actions, which mirror the same env)
	// expose a token service rather than the token itself.
	if reqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); reqURL != "" {
		return githubActionsOIDCToken(reqURL, os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"), audience)
	}
	return "", fmt.Errorf("no OIDC token found — pass --oidc-token or --oidc-token-file, set DINA_OIDC_TOKEN " +
		"(e.g. a GitLab id_token), or run inside a CI job with id-token permissions")
}

// githubActionsOIDCToken requests an ID token from the GitHub/Forgejo Actions
// token service. Requires the job to declare `permissions: id-token: write`.
func githubActionsOIDCToken(requestURL, requestToken, audience string) (string, error) {
	if requestToken == "" {
		return "", fmt.Errorf("ACTIONS_ID_TOKEN_REQUEST_TOKEN not set — add `permissions: id-token: write` to the job")
	}
	u, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("parsing token request URL: %w", err)
	}
	if audience != "" {
		q := u.Query()
		q.Set("audience", audience)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := discoveryClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting OIDC token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("OIDC token request failed with status %d", resp.StatusCode)
	}

	var out struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decoding OIDC token response: %w", err)
	}
	if out.Value == "" {
		return "", fmt.Errorf("empty OIDC token from provider")
	}
	return out.Value, nil
}

// exchangeFederatedToken performs the RFC 8693 token exchange against the
// issuer's token endpoint.
func exchangeFederatedToken(issuer, subjectToken, audience string) (*tokenResponse, error) {
	tokenURL := resolveTokenURL(issuer)

	data := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":      {subjectToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:jwt"},
		"client_id":          {newClientID},
	}
	if audience != "" {
		data.Set("audience", audience)
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token exchange rejected with status %d — the CI identity may not be granted access", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}
