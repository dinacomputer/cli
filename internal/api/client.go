package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/auth"
)

const (
	defaultBaseURL        = "https://dina.sh/api/v1"
	defaultSignalsBaseURL = "https://signals.sokkel.io"
	envBaseURL            = "DINA_API_URL"
	envSignalsBaseURL     = "DINA_SIGNALS_URL"
	envToken              = "DINA_API_TOKEN"
)

// UserAgent is the value sent in the User-Agent header on outbound requests.
// The CLI sets this at startup with the build version and runtime info.
var UserAgent = "dina/unknown"

// Debug, when true, logs each outgoing request's method, URL, and response
// status to stderr. The CLI wires this from --debug / DEBUG=1.
var Debug bool

// Client is an HTTP client for the Dina API.
type Client struct {
	BaseURL        string
	SignalsBaseURL string
	creds          *auth.Credentials
	token          string // static token from env var
	http           *http.Client
}

// newClient builds a client, loading stored credentials or the DINA_API_TOKEN
// env var if present. The returned client may be unauthenticated.
func newClient() *Client {
	base := os.Getenv(envBaseURL)

	c := &Client{
		http: &http.Client{Timeout: 120 * time.Second},
	}

	creds, err := auth.LoadCredentials()
	if err == nil && creds != nil && creds.AccessToken != "" {
		c.creds = creds
		if base == "" && creds.APIBaseURL != "" {
			base = creds.APIBaseURL
		}
	} else if tok := os.Getenv(envToken); tok != "" {
		c.token = tok
	}

	if base == "" {
		base = defaultBaseURL
	}
	c.BaseURL = base

	signalsBase := os.Getenv(envSignalsBaseURL)
	if signalsBase == "" {
		signalsBase = defaultSignalsBaseURL
	}
	c.SignalsBaseURL = signalsBase

	return c
}

// NewClient returns a client for authenticated Dina API calls. Fails if no
// credentials are available.
func NewClient() (*Client, error) {
	c := newClient()
	if !c.hasAuth() {
		return nil, fmt.Errorf("not authenticated — run: dina auth login")
	}
	return c, nil
}

// NewAnonymousClient returns a client whether or not credentials are available.
// Suitable for endpoints that treat auth as optional.
func NewAnonymousClient() *Client {
	return newClient()
}

func (c *Client) hasAuth() bool {
	if c.token != "" {
		return true
	}
	return c.creds != nil && c.creds.AccessToken != ""
}

// bearerToken returns a valid access token, refreshing if needed.
func (c *Client) bearerToken() (string, error) {
	if c.token != "" {
		return c.token, nil
	}
	if c.creds == nil {
		return "", fmt.Errorf("not authenticated — run: dina auth login")
	}
	if c.creds.Expired() {
		refreshed, err := auth.RefreshAccessToken(c.creds)
		if err != nil {
			return "", err
		}
		c.creds = refreshed
	}
	return c.creds.AccessToken, nil
}

// ---------- request helpers ----------

func (c *Client) newRequest(method, path string, body any) (*http.Request, error) {
	url := c.BaseURL + path

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	tok, err := c.bearerToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("User-Agent", UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	if Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", req.Method, req.URL)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach %s — check your connection or set %s: %w", req.URL.Host, envBaseURL, err)
	}
	defer resp.Body.Close()
	if Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: <- %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode >= 400 {
		var apiErr ErrorModel
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return wrapHTTPError(req, resp, apiErr)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// wrapHTTPError turns a non-2xx API response into a message that tells the
// user what to do next.
func wrapHTTPError(req *http.Request, resp *http.Response, apiErr ErrorModel) error {
	detail := apiErr.Detail
	path := req.URL.Path

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("not authenticated or token expired — run: dina auth login")
	case http.StatusForbidden:
		if detail != "" {
			return fmt.Errorf("forbidden: %s", detail)
		}
		return fmt.Errorf("forbidden — your account may not have permission for this operation")
	case http.StatusNotFound:
		if name, ok := appNameFromPath(path); ok {
			return fmt.Errorf("no app named %q — list apps with: dina apps list", name)
		}
		if detail != "" {
			return fmt.Errorf("not found: %s", detail)
		}
		return fmt.Errorf("not found (%s)", path)
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited by %s — retry in a few moments", req.URL.Host)
	}

	if resp.StatusCode >= 500 {
		if detail != "" {
			return fmt.Errorf("server error %d from %s: %s — retry shortly", resp.StatusCode, req.URL.Host, detail)
		}
		return fmt.Errorf("server error %d from %s — retry shortly", resp.StatusCode, req.URL.Host)
	}

	if detail != "" {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, detail)
	}
	return fmt.Errorf("API error %d", resp.StatusCode)
}

// appNameFromPath pulls the <name> out of any /.../apps/<name>/... path so
// 404s can tell the user which app was missing. Handles both /apps/foo and
// /api/v1/apps/foo.
func appNameFromPath(p string) (string, bool) {
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if seg == "apps" && i+1 < len(parts) && parts[i+1] != "" {
			return parts[i+1], true
		}
	}
	return "", false
}

// ---------- Apps ----------

func (c *Client) ListApps() ([]GetAppOutput, error) {
	req, err := c.newRequest("GET", "/apps", nil)
	if err != nil {
		return nil, err
	}
	var out ListAppsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Apps, nil
}

func (c *Client) CreateApp(name string) (*App, error) {
	req, err := c.newRequest("POST", "/apps", CreateAppInput{Name: name})
	if err != nil {
		return nil, err
	}
	var app App
	return &app, c.do(req, &app)
}

func (c *Client) GetApp(name string) (*GetAppOutput, error) {
	req, err := c.newRequest("GET", "/apps/"+name, nil)
	if err != nil {
		return nil, err
	}
	var out GetAppOutput
	return &out, c.do(req, &out)
}

func (c *Client) UpdateApp(name string, input UpdateAppInput) (*App, error) {
	req, err := c.newRequest("PATCH", "/apps/"+name, input)
	if err != nil {
		return nil, err
	}
	var app App
	return &app, c.do(req, &app)
}

func (c *Client) DeleteApp(name string) error {
	req, err := c.newRequest("DELETE", "/apps/"+name, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// ---------- Deploy ----------

func (c *Client) Deploy(appName string, input DeployInput) (*Deployment, error) {
	req, err := c.newRequest("POST", "/apps/"+appName+"/deploy", input)
	if err != nil {
		return nil, err
	}
	var dep Deployment
	return &dep, c.do(req, &dep)
}

// DeploySource uploads a zip archive via the deploy/upload endpoint.
func (c *Client) DeploySource(appName string, zipData []byte, buildArgs map[string]string) (*Deployment, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", "source.zip")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(zipData); err != nil {
		return nil, err
	}

	if len(buildArgs) > 0 {
		argsJSON, err := json.Marshal(buildArgs)
		if err != nil {
			return nil, err
		}
		if err := w.WriteField("build_args", string(argsJSON)); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	url := c.BaseURL + "/apps/" + appName + "/deploy/upload"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	tok, err := c.bearerToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())

	var dep Deployment
	return &dep, c.do(req, &dep)
}

// ---------- Deployments ----------

func (c *Client) ListDeployments(appName string) ([]Deployment, error) {
	req, err := c.newRequest("GET", "/apps/"+appName+"/deployments", nil)
	if err != nil {
		return nil, err
	}
	var out ListDeploymentsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Deployments, nil
}

func (c *Client) GetDeployment(appName, deploymentID string) (*Deployment, error) {
	req, err := c.newRequest("GET", "/apps/"+appName+"/deployments/"+deploymentID, nil)
	if err != nil {
		return nil, err
	}
	var dep Deployment
	return &dep, c.do(req, &dep)
}

func (c *Client) GetDeploymentLogs(appName, deploymentID string) (string, error) {
	req, err := c.newRequest("GET", "/apps/"+appName+"/deployments/"+deploymentID+"/logs", nil)
	if err != nil {
		return "", err
	}
	var out DeploymentLogsOutput
	return out.BuildLog, c.do(req, &out)
}

// ---------- Logs ----------

func (c *Client) GetAppLogs(appName string, lines int) (string, error) {
	path := fmt.Sprintf("/apps/%s/logs?lines=%d", appName, lines)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return "", err
	}
	var out PodLogsOutput
	return out.Logs, c.do(req, &out)
}

// ---------- Env ----------

func (c *Client) SetEnvVars(appName string, vars map[string]string) ([]EnvVar, error) {
	req, err := c.newRequest("PUT", "/apps/"+appName+"/env", PutEnvVarsInput{Vars: vars})
	if err != nil {
		return nil, err
	}
	var out PutEnvVarsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Vars, nil
}

// ---------- Hostnames ----------

func (c *Client) AddHostname(appName, hostname string) (*Hostname, error) {
	req, err := c.newRequest("POST", "/apps/"+appName+"/hostnames", AddHostnameInput{Hostname: hostname})
	if err != nil {
		return nil, err
	}
	var h Hostname
	return &h, c.do(req, &h)
}

func (c *Client) RemoveHostname(appName, hostname string) error {
	req, err := c.newRequest("DELETE", "/apps/"+appName+"/hostnames/"+hostname, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// ---------- Users ----------

func (c *Client) ListUsers() ([]User, error) {
	req, err := c.newRequest("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	var out ListUsersOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Users, nil
}

func (c *Client) ActivateUser(id string) (*User, error) {
	req, err := c.newRequest("PATCH", "/users/"+id+"/activate", nil)
	if err != nil {
		return nil, err
	}
	var user User
	return &user, c.do(req, &user)
}
