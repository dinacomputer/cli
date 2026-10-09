package cli

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dinacomputer/cli/internal/api"
)

func TestMatchFingerprint(t *testing.T) {
	alerts := []api.Alert{{Fingerprint: "3f9a11aa22bb33cc"}, {Fingerprint: "3f8b44dd55ee66ff"}, {Fingerprint: "a1b2c3d4e5f60718"}}

	if got, err := matchFingerprint(alerts, "3f9"); err != nil || got != "3f9a11aa22bb33cc" {
		t.Errorf("unique prefix: got %q, %v", got, err)
	}
	if _, err := matchFingerprint(alerts, "3f"); err == nil || !strings.Contains(err.Error(), "matches 2 alerts") {
		t.Errorf("ambiguous prefix: got %v", err)
	}
	if _, err := matchFingerprint(alerts, "ff"); err == nil || !strings.Contains(err.Error(), "may have resolved") {
		t.Errorf("no match: got %v", err)
	}
}

func TestBreach(t *testing.T) {
	cases := map[string]api.Alert{
		"-":          {},
		"12":         {Value: "12"},
		"12 above 5": {Value: "12", Threshold: "5", CompareOp: "above"},
		"12 vs 5":    {Value: "12", Threshold: "5"},
	}
	for want, a := range cases {
		if got := breach(a); got != want {
			t.Errorf("breach(%+v) = %q, want %q", a, got, want)
		}
	}
}

func TestFormatMs(t *testing.T) {
	cases := map[float64]string{0.42: "0.4ms", 12.4: "12ms", 250: "250ms", 1500: "1.50s", 90_000: "1.5m"}
	for in, want := range cases {
		if got := formatMs(in); got != want {
			t.Errorf("formatMs(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatAgo(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	cases := map[time.Duration]string{30 * time.Second: "30s", 12 * time.Minute: "12m", 5 * time.Hour: "5h", 72 * time.Hour: "3d"}
	for d, want := range cases {
		if got := formatAgo(now.Add(-d), now); got != want {
			t.Errorf("formatAgo(-%v) = %q, want %q", d, got, want)
		}
	}
}

func TestPrintSpanTreeKeepsOrphans(t *testing.T) {
	spans := []api.Span{
		{SpanID: "root", Name: "GET /users", Service: "api", DurationMs: 80},
		{SpanID: "db", ParentSpanID: "root", Name: "SELECT", Service: "api", DurationMs: 70, Error: true, StatusMessage: "timeout"},
		{SpanID: "lost", ParentSpanID: "missing", Name: "cache.get", Service: "api", DurationMs: 1},
	}
	out := captureStdout(t, func() { printSpanTree(spans) })
	want := "GET /users  api  80ms\n  SELECT  api  70ms  ERROR: timeout\ncache.get  api  1.0ms\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	fn()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}
