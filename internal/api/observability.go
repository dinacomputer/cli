package api

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ---------- Types ----------

type ObservabilityStatus struct {
	Enabled bool   `json:"enabled"`
	Kind    string `json:"kind,omitempty"`
	UIURL   string `json:"ui_url,omitempty"`
}

// Window is a time range. Since takes an RFC 3339 time or a duration such as
// 30m, 24h or 7d; Until takes an RFC 3339 time. Empty means the API default
// (the last hour, ending now).
type Window struct {
	Since string
	Until string
}

func (w Window) apply(q url.Values) {
	if w.Since != "" {
		q.Set("from", w.Since)
	}
	if w.Until != "" {
		q.Set("to", w.Until)
	}
}

type ServiceSummary struct {
	Name              string  `json:"name"`
	Requests          uint64  `json:"requests"`
	Errors            uint64  `json:"errors"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	ErrorRate         float64 `json:"error_rate"`
	AvgMs             float64 `json:"avg_ms"`
	P99Ms             float64 `json:"p99_ms"`
}

type Operation struct {
	Name     string  `json:"name"`
	Requests uint64  `json:"requests"`
	Errors   uint64  `json:"errors"`
	P50Ms    float64 `json:"p50_ms"`
	P95Ms    float64 `json:"p95_ms"`
	P99Ms    float64 `json:"p99_ms"`
}

type ObservedField struct {
	Name     string `json:"name"`
	Context  string `json:"context,omitempty"`
	DataType string `json:"data_type,omitempty"`
}

type FieldsResult struct {
	Fields   []ObservedField `json:"fields,omitempty"`
	Complete bool            `json:"complete"`
	Values   []string        `json:"values,omitempty"`
}

type LogEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	Severity   string            `json:"severity,omitempty"`
	Body       string            `json:"body"`
	Service    string            `json:"service,omitempty"`
	TraceID    string            `json:"trace_id,omitempty"`
	SpanID     string            `json:"span_id,omitempty"`
	Attributes map[string]any    `json:"attributes,omitempty"`
	Resource   map[string]string `json:"resource,omitempty"`
}

type LogPage struct {
	Logs       []LogEntry `json:"logs"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type LogQuery struct {
	Window
	Filter   string
	Service  string
	Severity []string
	Limit    int
	Cursor   string
}

type SpanSummary struct {
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"trace_id"`
	SpanID     string    `json:"span_id"`
	Service    string    `json:"service"`
	Name       string    `json:"name"`
	DurationMs float64   `json:"duration_ms"`
	Error      bool      `json:"error"`
	HTTPStatus string    `json:"http_status,omitempty"`
}

type SpanPage struct {
	Spans      []SpanSummary `json:"spans"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

type SpanQuery struct {
	Window
	Filter        string
	Service       string
	ErrorsOnly    bool
	MinDurationMs int64
	Limit         int
	Cursor        string
}

type Trace struct {
	TraceID        string    `json:"trace_id"`
	RootService    string    `json:"root_service"`
	RootOperation  string    `json:"root_operation"`
	Start          time.Time `json:"start"`
	DurationMs     float64   `json:"duration_ms"`
	SpanCount      uint64    `json:"span_count"`
	ErrorSpanCount uint64    `json:"error_span_count"`
	Truncated      bool      `json:"truncated"`
	Spans          []Span    `json:"spans"`
}

type Span struct {
	SpanID        string            `json:"span_id"`
	ParentSpanID  string            `json:"parent_span_id,omitempty"`
	Name          string            `json:"name"`
	Service       string            `json:"service"`
	Kind          string            `json:"kind,omitempty"`
	Start         time.Time         `json:"start"`
	DurationMs    float64           `json:"duration_ms"`
	Error         bool              `json:"error"`
	StatusMessage string            `json:"status_message,omitempty"`
	Attributes    map[string]any    `json:"attributes,omitempty"`
	Resource      map[string]string `json:"resource,omitempty"`
	Events        []SpanEvent       `json:"events,omitempty"`
}

type SpanEvent struct {
	Name       string         `json:"name"`
	Time       time.Time      `json:"time"`
	Error      bool           `json:"error"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type ExceptionGroup struct {
	GroupID   string    `json:"group_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	Count     uint64    `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type Exception struct {
	GroupID    string    `json:"group_id"`
	ErrorID    string    `json:"error_id"`
	Type       string    `json:"type"`
	Message    string    `json:"message"`
	Stacktrace string    `json:"stacktrace"`
	Escaped    bool      `json:"escaped"`
	Service    string    `json:"service"`
	TraceID    string    `json:"trace_id"`
	SpanID     string    `json:"span_id"`
	Timestamp  time.Time `json:"timestamp"`
}

type ExceptionQuery struct {
	Window
	Service string
	Limit   int
	Offset  int
}

type MetricQuery struct {
	Metric           string   `json:"metric"`
	TimeAggregation  string   `json:"time_aggregation,omitempty"`
	SpaceAggregation string   `json:"space_aggregation,omitempty"`
	GroupBy          []string `json:"group_by,omitempty"`
	Filter           string   `json:"filter,omitempty"`
	From             string   `json:"from,omitempty"`
	To               string   `json:"to,omitempty"`
	StepSeconds      int      `json:"step_seconds,omitempty"`
}

type MetricSeries struct {
	Labels map[string]string `json:"labels"`
	Points []MetricPoint     `json:"points"`
}

type MetricPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

type InfraRecord struct {
	Name           string            `json:"name"`
	Status         string            `json:"status,omitempty"`
	CPU            float64           `json:"cpu"`
	CPUOfRequest   *float64          `json:"cpu_of_request,omitempty"`
	CPUOfLimit     *float64          `json:"cpu_of_limit,omitempty"`
	CPUAllocatable float64           `json:"cpu_allocatable,omitempty"`
	Memory         float64           `json:"memory"`
	MemoryOfReq    *float64          `json:"memory_of_request,omitempty"`
	MemoryOfLimit  *float64          `json:"memory_of_limit,omitempty"`
	MemoryAlloc    float64           `json:"memory_allocatable,omitempty"`
	DiskUsage      float64           `json:"disk_usage,omitempty"`
	Load15         float64           `json:"load15,omitempty"`
	Restarts       int64             `json:"restarts,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

type InfraPage struct {
	Records []InfraRecord `json:"records"`
	Total   int           `json:"total"`
}

type InfraQuery struct {
	Window
	Filter string
	Limit  int
	Offset int
}

type Alert struct {
	Fingerprint      string            `json:"fingerprint"`
	Name             string            `json:"name"`
	State            string            `json:"state"`
	Severity         string            `json:"severity,omitempty"`
	RuleID           string            `json:"rule_id,omitempty"`
	StartsAt         time.Time         `json:"starts_at"`
	Summary          string            `json:"summary,omitempty"`
	Description      string            `json:"description,omitempty"`
	Value            string            `json:"value,omitempty"`
	Threshold        string            `json:"threshold,omitempty"`
	CompareOp        string            `json:"compare_op,omitempty"`
	Labels           map[string]string `json:"labels"`
	UIURL            string            `json:"ui_url,omitempty"`
	RelatedLogsURL   string            `json:"related_logs_url,omitempty"`
	RelatedTracesURL string            `json:"related_traces_url,omitempty"`
}

type AlertRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	State       string `json:"state"`
	Type        string `json:"type"`
	Disabled    bool   `json:"disabled"`
}

type AlertHistory struct {
	Triggers         uint64 `json:"triggers"`
	PreviousTriggers uint64 `json:"previous_triggers"`
}

type AlertContext struct {
	Alert      Alert            `json:"alert"`
	Rule       *AlertRule       `json:"rule,omitempty"`
	History    *AlertHistory    `json:"history,omitempty"`
	Service    string           `json:"service,omitempty"`
	Exceptions []ExceptionGroup `json:"exceptions"`
	ErrorLogs  []LogEntry       `json:"error_logs"`
	Warnings   []string         `json:"warnings,omitempty"`
}

// ---------- Calls ----------

func obsPath(orgID string, parts ...string) string {
	p := "/organizations/" + orgID + "/observability"
	for _, part := range parts {
		p += "/" + url.PathEscape(part)
	}
	return p
}

func withQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

func (c *Client) get(path string, out any) error {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) ObservabilityStatus(orgID string) (*ObservabilityStatus, error) {
	var out ObservabilityStatus
	return &out, c.get(obsPath(orgID), &out)
}

func (c *Client) ListObservedServices(orgID string, w Window) ([]ServiceSummary, error) {
	q := url.Values{}
	w.apply(q)
	var out struct {
		Services []ServiceSummary `json:"services"`
	}
	if err := c.get(withQuery(obsPath(orgID, "services"), q), &out); err != nil {
		return nil, err
	}
	return out.Services, nil
}

func (c *Client) ListServiceOperations(orgID, service string, w Window, limit int) ([]Operation, error) {
	q := url.Values{}
	w.apply(q)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	var out struct {
		Operations []Operation `json:"operations"`
	}
	if err := c.get(withQuery(obsPath(orgID, "services", service, "operations"), q), &out); err != nil {
		return nil, err
	}
	return out.Operations, nil
}

func (c *Client) ListObservedFields(orgID, signal, key string) (*FieldsResult, error) {
	q := url.Values{"signal": {signal}}
	if key != "" {
		q.Set("key", key)
	}
	var out FieldsResult
	return &out, c.get(withQuery(obsPath(orgID, "fields"), q), &out)
}

func (c *Client) SearchLogs(orgID string, in LogQuery) (*LogPage, error) {
	q := url.Values{}
	in.apply(q)
	setIf(q, "filter", in.Filter)
	setIf(q, "service", in.Service)
	if len(in.Severity) > 0 {
		q.Set("severity", strings.Join(in.Severity, ","))
	}
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}
	setIf(q, "cursor", in.Cursor)
	var out LogPage
	return &out, c.get(withQuery(obsPath(orgID, "logs"), q), &out)
}

func (c *Client) SearchSpans(orgID string, in SpanQuery) (*SpanPage, error) {
	q := url.Values{}
	in.apply(q)
	setIf(q, "filter", in.Filter)
	setIf(q, "service", in.Service)
	if in.ErrorsOnly {
		q.Set("errors_only", "true")
	}
	if in.MinDurationMs > 0 {
		q.Set("min_duration_ms", strconv.FormatInt(in.MinDurationMs, 10))
	}
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}
	setIf(q, "cursor", in.Cursor)
	var out SpanPage
	return &out, c.get(withQuery(obsPath(orgID, "traces"), q), &out)
}

func (c *Client) GetTrace(orgID, traceID string) (*Trace, error) {
	var out Trace
	return &out, c.get(obsPath(orgID, "traces", traceID), &out)
}

func (c *Client) ListExceptions(orgID string, in ExceptionQuery) ([]ExceptionGroup, error) {
	q := url.Values{}
	in.apply(q)
	setIf(q, "service", in.Service)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}
	if in.Offset > 0 {
		q.Set("offset", strconv.Itoa(in.Offset))
	}
	var out struct {
		Exceptions []ExceptionGroup `json:"exceptions"`
	}
	if err := c.get(withQuery(obsPath(orgID, "exceptions"), q), &out); err != nil {
		return nil, err
	}
	return out.Exceptions, nil
}

// GetException fetches the occurrence of a group at exactly at, which must
// be the group's last_seen as the API returned it.
func (c *Client) GetException(orgID, groupID string, at time.Time) (*Exception, error) {
	q := url.Values{"at": {at.Format(time.RFC3339Nano)}}
	var out Exception
	return &out, c.get(withQuery(obsPath(orgID, "exceptions", groupID), q), &out)
}

func (c *Client) QueryMetric(orgID string, in MetricQuery) ([]MetricSeries, error) {
	req, err := c.newRequest("POST", obsPath(orgID, "metrics", "query"), in)
	if err != nil {
		return nil, err
	}
	var out struct {
		Series []MetricSeries `json:"series"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Series, nil
}

func (c *Client) ListInfrastructure(orgID, kind string, in InfraQuery) (*InfraPage, error) {
	q := url.Values{}
	in.apply(q)
	setIf(q, "filter", in.Filter)
	if in.Limit > 0 {
		q.Set("limit", strconv.Itoa(in.Limit))
	}
	if in.Offset > 0 {
		q.Set("offset", strconv.Itoa(in.Offset))
	}
	var out InfraPage
	return &out, c.get(withQuery(obsPath(orgID, "infrastructure", kind), q), &out)
}

func (c *Client) ListAlerts(orgID string) ([]Alert, error) {
	var out struct {
		Alerts []Alert `json:"alerts"`
	}
	if err := c.get(obsPath(orgID, "alerts"), &out); err != nil {
		return nil, err
	}
	return out.Alerts, nil
}

func (c *Client) GetAlert(orgID, fingerprint string) (*AlertContext, error) {
	var out AlertContext
	return &out, c.get(obsPath(orgID, "alerts", fingerprint), &out)
}

func setIf(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}
