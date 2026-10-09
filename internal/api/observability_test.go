package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeAPI answers one route and records the request it got.
func fakeAPI(t *testing.T, status int, body any) (*Client, *http.Request, *[]byte) {
	t.Helper()
	var got http.Request
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(r.Context())
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	c := &Client{BaseURL: srv.URL, token: "tok", http: srv.Client()}
	return c, &got, &gotBody
}

func TestSearchLogsEncodesQuery(t *testing.T) {
	c, got, _ := fakeAPI(t, 200, map[string]any{
		"logs":        []any{map[string]any{"body": "boom", "severity": "ERROR"}},
		"next_cursor": "c2",
	})
	page, err := c.SearchLogs("org1", LogQuery{
		Window:   Window{Since: "24h"},
		Filter:   "body CONTAINS 'x'",
		Service:  "api",
		Severity: []string{"error", "fatal"},
		Limit:    10,
		Cursor:   "c1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.URL.Path != "/organizations/org1/observability/logs" {
		t.Errorf("path = %s", got.URL.Path)
	}
	q := got.URL.Query()
	for k, want := range map[string]string{
		"from": "24h", "filter": "body CONTAINS 'x'", "service": "api",
		"severity": "error,fatal", "limit": "10", "cursor": "c1",
	} {
		if q.Get(k) != want {
			t.Errorf("query %s = %q, want %q", k, q.Get(k), want)
		}
	}
	if _, set := q["to"]; set {
		t.Error("an empty --until must not be sent")
	}
	if len(page.Logs) != 1 || page.Logs[0].Body != "boom" || page.NextCursor != "c2" {
		t.Errorf("page = %+v", page)
	}
}

func TestListReturnsDecodedSlices(t *testing.T) {
	c, _, _ := fakeAPI(t, 200, map[string]any{"alerts": []any{map[string]any{"fingerprint": "abc", "name": "High error rate"}}})
	alerts, err := c.ListAlerts("org1")
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Name != "High error rate" {
		t.Fatalf("alerts = %+v", alerts)
	}
}

func TestGetExceptionSendsExactTimestamp(t *testing.T) {
	at := time.Date(2026, 10, 9, 11, 45, 0, 123456789, time.UTC)
	c, got, _ := fakeAPI(t, 200, map[string]any{"group_id": "g1"})
	if _, err := c.GetException("org1", "g1", at); err != nil {
		t.Fatal(err)
	}
	if got.URL.Query().Get("at") != "2026-10-09T11:45:00.123456789Z" {
		t.Fatalf("at = %q; the API matches the occurrence to the nanosecond", got.URL.Query().Get("at"))
	}
}

func TestQueryMetricPostsBody(t *testing.T) {
	c, got, body := fakeAPI(t, 200, map[string]any{"series": []any{}})
	_, err := c.QueryMetric("org1", MetricQuery{Metric: "m", SpaceAggregation: "p99", GroupBy: []string{"service.name"}, From: "6h"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != "POST" || got.URL.Path != "/organizations/org1/observability/metrics/query" {
		t.Errorf("%s %s", got.Method, got.URL.Path)
	}
	if !strings.Contains(string(*body), `"group_by":["service.name"]`) || strings.Contains(string(*body), "time_aggregation") {
		t.Errorf("body = %s", *body)
	}
}

func TestObservabilityNotFoundUsesAPIMessage(t *testing.T) {
	c, _, _ := fakeAPI(t, 404, map[string]any{"status": 404, "detail": "observability is not enabled for this organization; ask a platform operator to enable it"})
	_, err := c.ListAlerts("org1")
	if err == nil || !strings.HasPrefix(err.Error(), "observability is not enabled") {
		t.Fatalf("err = %v", err)
	}
}
