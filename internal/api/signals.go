package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Client) newSignalsRequest(method, path string, body any) (*http.Request, error) {
	url := c.SignalsBaseURL + path

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
	if c.hasAuth() {
		tok, err := c.bearerToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) SubmitBug(body BugBody) (*SubmitResponseBody, error) {
	req, err := c.newSignalsRequest("POST", "/v1/bugs", body)
	if err != nil {
		return nil, err
	}
	var out SubmitResponseBody
	return &out, c.do(req, &out)
}

func (c *Client) SubmitFeatureRequest(body FeatureRequestBody) (*SubmitResponseBody, error) {
	req, err := c.newSignalsRequest("POST", "/v1/feature-requests", body)
	if err != nil {
		return nil, err
	}
	var out SubmitResponseBody
	return &out, c.do(req, &out)
}

func (c *Client) SubmitFeedback(body FeedbackBody) (*SubmitResponseBody, error) {
	req, err := c.newSignalsRequest("POST", "/v1/feedback", body)
	if err != nil {
		return nil, err
	}
	var out SubmitResponseBody
	return &out, c.do(req, &out)
}
