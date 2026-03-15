package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/dinacomputer/cli/internal/auth"
)

const (
	defaultBaseURL = "https://dina.sh/api/v1"
	envBaseURL     = "DINA_API_URL"
	envToken       = "DINA_API_TOKEN"
)

// Client is an HTTP client for the Dina API.
type Client struct {
	BaseURL string
	creds   *auth.Credentials
	token   string // static token from env var
	http    *http.Client
}

// NewClient creates a client. It first checks for stored OAuth credentials,
// then falls back to the DINA_API_TOKEN environment variable.
func NewClient() (*Client, error) {
	base := os.Getenv(envBaseURL)
	if base == "" {
		base = defaultBaseURL
	}
	c := &Client{
		BaseURL: base,
		http:    &http.Client{Timeout: 120 * time.Second},
	}

	// Try stored OAuth credentials first.
	creds, err := auth.LoadCredentials()
	if err == nil && creds != nil && creds.AccessToken != "" {
		c.creds = creds
		return c, nil
	}

	// Fall back to env var.
	tok := os.Getenv(envToken)
	if tok != "" {
		c.token = tok
		return c, nil
	}

	return nil, fmt.Errorf("not authenticated — run: dina auth login")
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr ErrorModel
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Detail != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Detail)
		}
		return fmt.Errorf("API error %d", resp.StatusCode)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
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
func (c *Client) DeploySource(appName string, zipData []byte) (*Deployment, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", "source.zip")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(zipData); err != nil {
		return nil, err
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
