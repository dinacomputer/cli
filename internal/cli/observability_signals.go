package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/dinacomputer/cli/internal/api"
	"github.com/spf13/cobra"
)

// --- services ---

var obsServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Request rate, error rate and latency per service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		services, err := client.ListObservedServices(orgID, obsWindow())
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(services)
		}
		if len(services) == 0 {
			fmt.Println("No services reported traces in this window.")
			return nil
		}
		fmt.Printf("%-32s  %9s  %8s  %7s  %9s  %9s\n", "SERVICE", "REQUESTS", "REQ/S", "ERRORS", "P99", "AVG")
		for _, s := range services {
			fmt.Printf("%-32s  %9d  %8.2f  %7s  %9s  %9s\n",
				truncate(s.Name, 32), s.Requests, s.RequestsPerSecond, formatPercent(s.ErrorRate), formatMs(s.P99Ms), formatMs(s.AvgMs))
		}
		return nil
	},
}

var obsOperationsLimit int

var obsOperationsCmd = &cobra.Command{
	Use:     "operations <service>",
	Aliases: []string{"ops"},
	Short:   "The slowest operations of one service",
	Example: `  dina obs operations api --since 24h`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		ops, err := client.ListServiceOperations(orgID, args[0], obsWindow(), obsOperationsLimit)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(ops)
		}
		if len(ops) == 0 {
			fmt.Printf("No operations for %s in this window.\n", args[0])
			return nil
		}
		fmt.Printf("%-48s  %9s  %7s  %9s  %9s  %9s\n", "OPERATION", "REQUESTS", "ERRORS", "P50", "P95", "P99")
		for _, o := range ops {
			fmt.Printf("%-48s  %9d  %7d  %9s  %9s  %9s\n",
				truncate(o.Name, 48), o.Requests, o.Errors, formatMs(o.P50Ms), formatMs(o.P95Ms), formatMs(o.P99Ms))
		}
		return nil
	},
}

// --- fields ---

var (
	obsFieldsSignal string
	obsFieldsKey    string
)

var obsFieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "List the field names you can filter on",
	Long: `List the field names that exist in your telemetry, for writing --filter
expressions. With --key, list that field's most common values instead.`,
	Example: `  dina obs fields --signal logs
  dina obs fields --signal traces --key service.name`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		res, err := client.ListObservedFields(orgID, obsFieldsSignal, obsFieldsKey)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(res)
		}
		if obsFieldsKey != "" {
			for _, v := range res.Values {
				fmt.Println(v)
			}
			return nil
		}
		for _, f := range res.Fields {
			fmt.Printf("%-48s  %-10s  %s\n", f.Name, f.Context, f.DataType)
		}
		if !res.Complete {
			Infoln("The list was capped; more fields exist.")
		}
		return nil
	},
}

// --- logs ---

var (
	obsLogsFilter   string
	obsLogsService  string
	obsLogsSeverity []string
	obsLogsLimit    int
	obsLogsCursor   string
)

var obsLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Search logs, newest first",
	Example: `  dina obs logs --service api --severity error,fatal
  dina obs logs --since 24h --filter "body CONTAINS 'timeout'"
  dina obs logs --filter "k8s.pod.name = 'api-7f9c'" -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		page, err := client.SearchLogs(orgID, api.LogQuery{
			Window:   obsWindow(),
			Filter:   obsLogsFilter,
			Service:  obsLogsService,
			Severity: obsLogsSeverity,
			Limit:    obsLogsLimit,
			Cursor:   obsLogsCursor,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(page)
		}
		if len(page.Logs) == 0 {
			fmt.Println("No logs matched.")
			return nil
		}
		for _, l := range page.Logs {
			sev := l.Severity
			if sev == "" {
				sev = "-"
			}
			fmt.Printf("%s  %-5s  %-20s  %s\n", formatTime(l.Timestamp), truncate(strings.ToUpper(sev), 5), truncate(l.Service, 20), oneLine(l.Body))
		}
		moreHint(page.NextCursor)
		return nil
	},
}

// --- traces ---

var (
	obsTracesFilter      string
	obsTracesService     string
	obsTracesErrors      bool
	obsTracesMinDuration time.Duration
	obsTracesLimit       int
	obsTracesCursor      string
)

var obsTracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "Search spans, newest first",
	Long:  "Search spans, newest first. Show a whole trace with: dina obs trace <trace-id>",
	Example: `  dina obs traces --service api --errors
  dina obs traces --min-duration 500ms --since 6h
  dina obs traces --filter "http.route = '/users/{id}'"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		page, err := client.SearchSpans(orgID, api.SpanQuery{
			Window:        obsWindow(),
			Filter:        obsTracesFilter,
			Service:       obsTracesService,
			ErrorsOnly:    obsTracesErrors,
			MinDurationMs: obsTracesMinDuration.Milliseconds(),
			Limit:         obsTracesLimit,
			Cursor:        obsTracesCursor,
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(page)
		}
		if len(page.Spans) == 0 {
			fmt.Println("No spans matched.")
			return nil
		}
		for _, s := range page.Spans {
			status := s.HTTPStatus
			if s.Error {
				status = strings.TrimSpace("ERR " + status)
			}
			fmt.Printf("%s  %-20s  %-40s  %9s  %-7s  %s\n",
				formatTime(s.Timestamp), truncate(s.Service, 20), truncate(s.Name, 40), formatMs(s.DurationMs), status, s.TraceID)
		}
		moreHint(page.NextCursor)
		return nil
	},
}

var obsTraceCmd = &cobra.Command{
	Use:     "trace <trace-id>",
	Short:   "Show every span of one trace as a tree",
	Example: `  dina obs trace 4bf92f3577b34da6a3ce929d0e0e4736`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		t, err := client.GetTrace(orgID, args[0])
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(t)
		}
		fmt.Printf("Trace %s\n", t.TraceID)
		fmt.Printf("%s %s, %s, %d spans (%d with errors), started %s\n\n",
			t.RootService, t.RootOperation, formatMs(t.DurationMs), t.SpanCount, t.ErrorSpanCount, formatTime(t.Start))
		printSpanTree(t.Spans)
		if t.Truncated {
			Infoln("The trace is large; only part of it was returned.")
		}
		return nil
	},
}

// printSpanTree prints spans indented under their parents. Spans whose parent
// is missing from the response are printed as roots, so nothing is dropped.
func printSpanTree(spans []api.Span) {
	present := make(map[string]bool, len(spans))
	children := map[string][]api.Span{}
	for _, s := range spans {
		present[s.SpanID] = true
	}
	var roots []api.Span
	for _, s := range spans {
		if s.ParentSpanID == "" || !present[s.ParentSpanID] {
			roots = append(roots, s)
			continue
		}
		children[s.ParentSpanID] = append(children[s.ParentSpanID], s)
	}
	var walk func(s api.Span, depth int)
	walk = func(s api.Span, depth int) {
		line := fmt.Sprintf("%s%s  %s  %s", strings.Repeat("  ", depth), s.Name, s.Service, formatMs(s.DurationMs))
		if s.Error {
			line += "  ERROR"
			if s.StatusMessage != "" {
				line += ": " + oneLine(s.StatusMessage)
			}
		}
		fmt.Println(line)
		for _, c := range children[s.SpanID] {
			walk(c, depth+1)
		}
	}
	for _, r := range roots {
		walk(r, 0)
	}
}

// --- exceptions ---

var (
	obsExceptionsService string
	obsExceptionsLimit   int
)

var obsExceptionsCmd = &cobra.Command{
	Use:     "exceptions",
	Aliases: []string{"errors"},
	Short:   "Exceptions grouped by type and message, most recent first",
	Example: `  dina obs exceptions --service api --since 24h
  dina obs exceptions show <group-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		groups, err := client.ListExceptions(orgID, api.ExceptionQuery{Window: obsWindow(), Service: obsExceptionsService, Limit: obsExceptionsLimit})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(groups)
		}
		if len(groups) == 0 {
			fmt.Println("No exceptions in this window.")
			return nil
		}
		now := time.Now()
		fmt.Printf("%-8s  %6s  %-20s  %-60s  %s\n", "LAST", "COUNT", "SERVICE", "EXCEPTION", "GROUP")
		for _, g := range groups {
			fmt.Printf("%-8s  %6d  %-20s  %-60s  %s\n",
				formatAgo(g.LastSeen, now)+" ago", g.Count, truncate(g.Service, 20), truncate(oneLine(g.Type+": "+g.Message), 60), g.GroupID)
		}
		return nil
	},
}

var obsExceptionShowCmd = &cobra.Command{
	Use:   "show <group-id>",
	Short: "Show the latest occurrence of an exception, with its stacktrace",
	Long: `Show the most recent occurrence of an exception group from
` + "`dina obs exceptions`" + `, with its stacktrace and the trace it happened in. The
group must have occurred within --since (default 1h).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		// The API needs the exact time of an occurrence, which only the
		// group list knows.
		groups, err := client.ListExceptions(orgID, api.ExceptionQuery{Window: obsWindow(), Service: obsExceptionsService, Limit: 500})
		if err != nil {
			return err
		}
		var found *api.ExceptionGroup
		for i := range groups {
			if groups[i].GroupID == args[0] {
				found = &groups[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("no exception group %q in this window — widen it with --since 24h, or list groups with: dina obs exceptions", args[0])
		}
		e, err := client.GetException(orgID, found.GroupID, found.LastSeen)
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(e)
		}
		fmt.Printf("%s: %s\n", e.Type, e.Message)
		fmt.Printf("Service: %s\nWhen:    %s (%d occurrences since %s)\n", e.Service, formatTime(e.Timestamp), found.Count, formatTime(found.FirstSeen))
		if e.TraceID != "" {
			fmt.Printf("Trace:   %s  (dina obs trace %s)\n", e.TraceID, e.TraceID)
		}
		if e.Stacktrace != "" {
			fmt.Printf("\n%s\n", strings.TrimRight(e.Stacktrace, "\n"))
		}
		return nil
	},
}

// --- metrics ---

var (
	obsMetricsTimeAgg  string
	obsMetricsSpaceAgg string
	obsMetricsGroupBy  []string
	obsMetricsFilter   string
	obsMetricsStep     time.Duration
)

var obsMetricsCmd = &cobra.Command{
	Use:   "metrics <metric>",
	Short: "Query a metric as time series",
	Long: `Query a metric as time series. --time-agg reduces each series within a
step (avg, sum, rate, latest, ...); --space-agg combines series across
--group-by (sum, avg, max, p99, ...). Both default to avg; use --time-agg rate
or increase for counters. Find metric names with: dina obs fields --signal metrics`,
	Example: `  dina obs metrics http.server.request.duration --space-agg p99 --group-by service.name
  dina obs metrics k8s.pod.memory.working_set --space-agg max --group-by k8s.pod.name --since 6h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		series, err := client.QueryMetric(orgID, api.MetricQuery{
			Metric:           args[0],
			TimeAggregation:  obsMetricsTimeAgg,
			SpaceAggregation: obsMetricsSpaceAgg,
			GroupBy:          obsMetricsGroupBy,
			Filter:           obsMetricsFilter,
			From:             obsSince,
			To:               obsUntil,
			StepSeconds:      int(obsMetricsStep.Seconds()),
		})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(series)
		}
		if len(series) == 0 {
			fmt.Println("No data for this metric in this window.")
			return nil
		}
		for i, s := range series {
			if i > 0 {
				fmt.Println()
			}
			labels := make([]string, 0, len(s.Labels))
			for _, k := range sortedKeys(s.Labels) {
				labels = append(labels, k+"="+s.Labels[k])
			}
			if len(labels) == 0 {
				labels = append(labels, "(all)")
			}
			fmt.Println(strings.Join(labels, " "))
			for _, p := range s.Points {
				fmt.Printf("  %s  %g\n", formatTime(p.Time), p.Value)
			}
		}
		return nil
	},
}

// --- infrastructure ---

var (
	obsInfraFilter string
	obsInfraLimit  int
)

var obsInfraCmd = &cobra.Command{
	Use:       "infra <hosts|pods|nodes>",
	Aliases:   []string{"infrastructure"},
	Short:     "Resource usage per host, pod or node",
	Example:   `  dina obs infra pods --filter "k8s.namespace.name = 'api'"`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"hosts", "pods", "nodes"},
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := args[0]
		if kind != "hosts" && kind != "pods" && kind != "nodes" {
			return fmt.Errorf("unknown kind %q — use hosts, pods or nodes", kind)
		}
		client, orgID, err := obsClient()
		if err != nil {
			return err
		}
		page, err := client.ListInfrastructure(orgID, kind, api.InfraQuery{Window: obsWindow(), Filter: obsInfraFilter, Limit: obsInfraLimit})
		if err != nil {
			return err
		}
		if JSONOutput() {
			return writeJSON(page)
		}
		if len(page.Records) == 0 {
			fmt.Printf("No %s reported metrics in this window.\n", kind)
			return nil
		}
		switch kind {
		case "hosts":
			fmt.Printf("%-32s  %-8s  %6s  %6s  %6s  %7s\n", "HOST", "STATUS", "CPU", "MEM", "DISK", "LOAD15")
			for _, r := range page.Records {
				fmt.Printf("%-32s  %-8s  %6s  %6s  %6s  %7.2f\n", truncate(r.Name, 32), r.Status,
					formatPercent(r.CPU), formatPercent(r.Memory), formatPercent(r.DiskUsage), r.Load15)
			}
		case "pods":
			fmt.Printf("%-44s  %-10s  %6s  %9s  %9s  %9s  %8s\n", "POD", "STATUS", "CPU", "OF LIMIT", "MEMORY", "OF LIMIT", "RESTARTS")
			for _, r := range page.Records {
				fmt.Printf("%-44s  %-10s  %6s  %9s  %9s  %9s  %8d\n", truncate(r.Name, 44), r.Status,
					formatCores(r.CPU), formatRatio(r.CPUOfLimit), formatMem(r.Memory), formatRatio(r.MemoryOfLimit), r.Restarts)
			}
		case "nodes":
			fmt.Printf("%-32s  %-10s  %15s  %21s\n", "NODE", "CONDITION", "CPU / ALLOC", "MEMORY / ALLOC")
			for _, r := range page.Records {
				fmt.Printf("%-32s  %-10s  %15s  %21s\n", truncate(r.Name, 32), r.Status,
					formatCores(r.CPU)+" / "+formatCores(r.CPUAllocatable), formatMem(r.Memory)+" / "+formatMem(r.MemoryAlloc))
			}
		}
		if page.Total > len(page.Records) {
			Infof("Showing %d of %d; raise --limit to see more.\n", len(page.Records), page.Total)
		}
		return nil
	},
}

// formatCores renders CPU cores; zero is "-", since a pod without a limit has
// no number to show.
func formatCores(c float64) string {
	if c == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", c)
}

func init() {
	for _, c := range []*cobra.Command{obsServicesCmd, obsOperationsCmd, obsLogsCmd, obsTracesCmd, obsExceptionsCmd, obsExceptionShowCmd, obsMetricsCmd, obsInfraCmd} {
		addWindowFlags(c)
	}

	obsOperationsCmd.Flags().IntVar(&obsOperationsLimit, "limit", 20, "Maximum operations to show")

	obsFieldsCmd.Flags().StringVar(&obsFieldsSignal, "signal", "logs", "Which telemetry: logs, traces or metrics")
	obsFieldsCmd.Flags().StringVar(&obsFieldsKey, "key", "", "List this field's most common values instead")

	obsLogsCmd.Flags().StringVar(&obsLogsFilter, "filter", "", "Filter expression (see dina obs --help)")
	obsLogsCmd.Flags().StringVar(&obsLogsService, "service", "", "Only this service")
	obsLogsCmd.Flags().StringSliceVar(&obsLogsSeverity, "severity", nil, "Only these severities, e.g. error,fatal")
	obsLogsCmd.Flags().IntVar(&obsLogsLimit, "limit", 50, "Maximum entries (1-1000)")
	obsLogsCmd.Flags().StringVar(&obsLogsCursor, "cursor", "", "Continue from a previous page")

	obsTracesCmd.Flags().StringVar(&obsTracesFilter, "filter", "", "Filter expression (see dina obs --help)")
	obsTracesCmd.Flags().StringVar(&obsTracesService, "service", "", "Only this service")
	obsTracesCmd.Flags().BoolVar(&obsTracesErrors, "errors", false, "Only failed spans")
	obsTracesCmd.Flags().DurationVar(&obsTracesMinDuration, "min-duration", 0, "Only spans at least this slow, e.g. 500ms")
	obsTracesCmd.Flags().IntVar(&obsTracesLimit, "limit", 50, "Maximum spans (1-1000)")
	obsTracesCmd.Flags().StringVar(&obsTracesCursor, "cursor", "", "Continue from a previous page")

	for _, c := range []*cobra.Command{obsExceptionsCmd, obsExceptionShowCmd} {
		c.Flags().StringVar(&obsExceptionsService, "service", "", "Only this service")
	}
	obsExceptionsCmd.Flags().IntVar(&obsExceptionsLimit, "limit", 50, "Maximum groups (1-500)")
	obsExceptionsCmd.AddCommand(obsExceptionShowCmd)

	obsMetricsCmd.Flags().StringVar(&obsMetricsTimeAgg, "time-agg", "", "latest, sum, avg, min, max, count, rate, increase (default avg)")
	obsMetricsCmd.Flags().StringVar(&obsMetricsSpaceAgg, "space-agg", "", "sum, avg, min, max, count, p50, p75, p90, p95, p99 (default avg)")
	obsMetricsCmd.Flags().StringSliceVar(&obsMetricsGroupBy, "group-by", nil, "Labels to split series by, e.g. service.name")
	obsMetricsCmd.Flags().StringVar(&obsMetricsFilter, "filter", "", "Filter on labels (see dina obs --help)")
	obsMetricsCmd.Flags().DurationVar(&obsMetricsStep, "step", 0, "Resolution, e.g. 5m (default: about 120 points)")

	obsInfraCmd.Flags().StringVar(&obsInfraFilter, "filter", "", "Filter on resource attributes, e.g. k8s.namespace.name = 'api'")
	obsInfraCmd.Flags().IntVar(&obsInfraLimit, "limit", 50, "Maximum records (1-500)")

	observabilityCmd.AddCommand(obsServicesCmd, obsOperationsCmd, obsFieldsCmd, obsLogsCmd, obsTracesCmd, obsTraceCmd, obsExceptionsCmd, obsMetricsCmd, obsInfraCmd)
}

// formatRatio renders a usage-of-limit ratio as a percentage, or "-" when the
// pod sets no limit.
func formatRatio(r *float64) string {
	if r == nil {
		return "-"
	}
	return formatPercent(*r)
}

// formatMem renders bytes; zero is "-" rather than formatBytes' "unlimited",
// which only makes sense for quotas.
func formatMem(b float64) string {
	if b == 0 {
		return "-"
	}
	return formatBytes(int64(b))
}
