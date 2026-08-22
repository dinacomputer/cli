package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultIssuer  = "https://dina.sh"
	defaultAPIBase = "https://dina.sh/api/v1"
	envAuthIssuer  = "DINA_AUTH_ISSUER"
	envAPIURL      = "DINA_API_URL"
)

// OIDCConfig holds endpoints discovered from an OpenID Connect provider.
type OIDCConfig struct {
	Issuer                      string `json:"issuer"`
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitempty"`
}

// ProtectedResource holds the response from RFC 9728 /.well-known/oauth-protected-resource.
type ProtectedResource struct {
	// Resource is the resource identifier clients must use as the audience of
	// tokens minted for this API — https://dina.sh, not the auth issuer.
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
}

var discoveryClient = &http.Client{Timeout: 10 * time.Second}

// DiscoverProtectedResource fetches RFC 9728 /.well-known/oauth-protected-resource
// from the API host.
func DiscoverProtectedResource(apiBaseURL string) (*ProtectedResource, error) {
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing API base URL: %w", err)
	}

	wellKnown := fmt.Sprintf("%s://%s/.well-known/oauth-protected-resource", u.Scheme, u.Host)
	resp, err := discoveryClient.Get(wellKnown)
	if err != nil {
		return nil, fmt.Errorf("fetching protected resource metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("protected resource metadata not found at %s", wellKnown)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("protected resource metadata returned status %d", resp.StatusCode)
	}

	var pr ProtectedResource
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decoding protected resource metadata: %w", err)
	}
	return &pr, nil
}

// DiscoverIssuer fetches /.well-known/oauth-protected-resource from the API host
// and returns the first authorization server (the issuer).
func DiscoverIssuer(apiBaseURL string) (string, error) {
	pr, err := DiscoverProtectedResource(apiBaseURL)
	if err != nil {
		return "", err
	}
	if len(pr.AuthorizationServers) == 0 {
		return "", fmt.Errorf("no authorization servers in protected resource metadata")
	}
	return strings.TrimRight(pr.AuthorizationServers[0], "/"), nil
}

// ResolveResourceAudience returns the resource identifier the API expects as the
// audience of a federated token — the `resource` from RFC 9728 metadata, falling
// back to the API base origin. This is what the federation validates the CI
// token's `aud` against; it is the API resource (https://dina.sh), not the auth
// issuer (https://id.sokkel.io).
func ResolveResourceAudience(apiBaseURL string) string {
	if pr, err := DiscoverProtectedResource(apiBaseURL); err == nil && pr.Resource != "" {
		return strings.TrimRight(pr.Resource, "/")
	}
	if u, err := url.Parse(apiBaseURL); err == nil && u.Host != "" {
		return u.Scheme + "://" + u.Host
	}
	return ""
}

var (
	oidcCache   = map[string]*OIDCConfig{}
	oidcCacheMu sync.Mutex
)

// DiscoverOIDC fetches /.well-known/openid-configuration from the issuer.
// Returns (nil, nil) if the endpoint returns 404 — the caller should fall back to hardcoded endpoints.
func DiscoverOIDC(issuer string) (*OIDCConfig, error) {
	oidcCacheMu.Lock()
	if cfg, ok := oidcCache[issuer]; ok {
		oidcCacheMu.Unlock()
		return cfg, nil
	}
	oidcCacheMu.Unlock()

	wellKnown := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	resp, err := discoveryClient.Get(wellKnown)
	if err != nil {
		return nil, fmt.Errorf("fetching OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("OIDC configuration returned status %d", resp.StatusCode)
	}

	var cfg OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding OIDC configuration: %w", err)
	}

	oidcCacheMu.Lock()
	oidcCache[issuer] = &cfg
	oidcCacheMu.Unlock()

	return &cfg, nil
}

// ResolveAPIBase returns the API base URL from DINA_API_URL or the default.
func ResolveAPIBase() string {
	if v := os.Getenv(envAPIURL); v != "" {
		return v
	}
	return defaultAPIBase
}
