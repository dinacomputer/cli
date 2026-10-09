package skills

// SkillMD returns the contents of the SKILL.md file that teaches AI agents
// how to use the Dina CLI.
func SkillMD() string {
	return `---
name: dina-cli
description: Deploy applications, manage apps, view logs, set env vars, configure hostnames, manage S3-compatible object storage, and inspect observability data (logs, traces, exceptions, metrics, firing alerts) on the Dina platform. Use when the user wants to deploy code, check app status, view logs, manage environment variables, configure custom domains, create storage buckets or access keys, investigate errors, slow requests or firing alerts, or perform any Dina platform operation.
allowed-tools: Bash(dina *), Bash(dina)
---

# Dina CLI

## Introduction

Dina is a platform-as-a-service (PaaS) for deploying and managing containerized applications. With the Dina CLI you can:

- **Deploy** applications from source code (current directory) or pre-built container images
- **Manage apps**: create, update, delete, and inspect applications
- **View logs**: stream runtime logs and build logs for deployments
- **Configure environment variables** for your applications
- **Manage custom hostnames** for your apps
- **Manage object storage**: S3-compatible buckets and the access keys that reach them
- **Investigate production**: firing alerts, exceptions, logs, traces, metrics and service health from the organization's SigNoz instance
- **Manage users** (admin operations)
- **Report bugs and send feedback** to the Sokkel team

## Session start

At the beginning of every session, run these two commands in order:

` + "```bash" + `
dina doctor --fix
dina help
` + "```" + `

**` + "`dina doctor --fix`" + `** verifies authentication is valid, brings any outdated installed skill files up to date with the canonical content in this CLI version, and reports whether a newer CLI version is available. Fixable issues (skills out of date) are repaired automatically. Issues that need user action (not logged in, CLI update available) are reported for the user to handle.

**` + "`dina help`" + `** prints the current list of available commands and their descriptions. Always run this so you know exactly which commands and flags the installed version supports — do not assume commands from this doc are present; verify via help. For sub-command details, run ` + "`dina [command] --help`" + `.

## Quick start

` + "```bash" + `
# authenticate
dina auth login

# create and deploy an app
dina apps create my-app
dina deploy -a my-app

# check status
dina apps info -a my-app
dina apps logs -a my-app
` + "```" + `

## Tips

- Apps must listen on port 8080. Always set PORT=8080 or configure the app to use port 8080.

## Commands

### Authentication

` + "```bash" + `
dina auth login
dina auth logout
dina auth status
` + "```" + `

### Deploy

Deploy from source (zips and uploads current directory):

` + "```bash" + `
dina deploy -a my-app
` + "```" + `

Deploy a pre-built image:

` + "```bash" + `
dina deploy -a my-app --tag registry.example.com/my-app:v1.2
` + "```" + `

Deploy with replica count:

` + "```bash" + `
dina deploy -a my-app --replicas 3
dina deploy -a my-app --tag nginx:latest --replicas 2
` + "```" + `

Deploy and wait for completion:

` + "```bash" + `
dina deploy -a my-app --wait
dina deploy -a my-app --tag nginx:latest -w
` + "```" + `

The ` + "`--wait`" + ` flag polls the deployment status every 2 seconds until it reaches a terminal state. Status progression: ` + "`pending`" + ` → ` + "`building`" + ` → ` + "`deploying`" + ` → ` + "`running`" + `. Failure states: ` + "`failed`" + `, ` + "`build_failed`" + `, ` + "`deploy_failed`" + `. Times out after 10 minutes. The command exits with a non-zero code on failure or timeout. Always use ` + "`--wait`" + ` when you need to confirm a deployment succeeded before proceeding.

### Apps

` + "```bash" + `
# list all apps
dina apps list

# create a new app
dina apps create my-app

# show app details (URL, hostnames, latest deployment)
dina apps info -a my-app

# rename an app
dina apps update -a my-app --name new-name

# delete an app (prompts for the app name to confirm — pass --force to skip)
dina apps delete -a my-app --force
` + "```" + `

### Logs

` + "```bash" + `
# runtime logs (default 100 lines)
dina apps logs -a my-app
dina apps logs -a my-app -n 50

# build logs for a specific deployment
dina apps deployments -a my-app
dina apps deployments logs -a my-app --id <deployment-id>
` + "```" + `

#### Log format

**Runtime logs** (` + "`dina apps logs`" + `): plain-text container output printed directly to stdout, one line per log entry exactly as written by your application. There is no structured JSON wrapper.

**Build logs** (` + "`dina apps deployments logs`" + `): full build output from the container image build process (e.g., Dockerfile steps, package installs, compile output), printed as plain text.

### Environment variables

` + "```bash" + `
dina apps env set -a my-app DATABASE_URL=postgres://localhost/mydb
dina apps env set -a my-app KEY1=val1 KEY2=val2
` + "```" + `

### Custom hostnames

` + "```bash" + `
dina apps hostnames add -a my-app example.com
# remove prompts y/N — pass --force to skip (required in scripts)
dina apps hostnames remove -a my-app example.com --force
` + "```" + `

### Object storage

S3-compatible buckets plus the access keys applications authenticate with. Both
are scoped to an organization; pass ` + "`--org`" + ` only when the account has more than
one (the CLI errors and lists them if it can't pick).

` + "```bash" + `
# buckets
dina storage buckets list
dina storage buckets create uploads
dina storage buckets create assets --public --quota 10GB
dina storage buckets get uploads
dina storage buckets update uploads --quota 50GB
dina storage buckets update assets --public=false
dina storage buckets delete uploads --force

# access keys
dina storage keys list
dina storage keys create my-app
dina storage keys get my-app
dina storage keys rotate my-app --force
dina storage keys delete my-app --force

# grants (a key reaches no bucket until granted)
dina storage keys grant my-app uploads --read --write
dina storage keys grant my-app assets --read
dina storage keys revoke my-app assets --force
` + "```" + `

**Wiring an app to a bucket** — the full sequence:

` + "```bash" + `
dina storage buckets create uploads --quota 10GB
dina storage keys create my-app                      # prints the secret ONCE
dina storage keys grant my-app uploads --read --write
dina storage buckets get uploads                     # endpoint, region, real bucket name
dina apps env set -a my-app \
  AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... \
  S3_ENDPOINT=https://storage.<region>.dina.sh S3_BUCKET=<physical name>
` + "```" + `

Things that will bite you if you skip them:

- **The secret is shown once.** ` + "`keys create`" + ` and ` + "`keys rotate`" + ` print
  ` + "`AWS_SECRET_ACCESS_KEY`" + ` on stdout and it is never retrievable again. Capture it
  in the same step you create the key — with ` + "`-o json`" + ` if you need to parse it — and
  write it straight into the app's env. Never re-run ` + "`create`" + ` hoping to re-read a
  secret; that mints a different key.
- **S3 clients must address the ` + "`physical_name`" + `, not the name you created.** The
  backend namespaces buckets per account, so ` + "`uploads`" + ` really lives at something like
  ` + "`your-org-uploads`" + `. Read it from ` + "`dina storage buckets get`" + ` (or ` + "`.physical_name`" + `
  in JSON). Using the short name gives ` + "`NoSuchBucket`" + `.
- **Path-style addressing is required.** Set ` + "`--endpoint-url`" + ` plus the client's
  force-path-style option (` + "`s3ForcePathStyle: true`" + `, ` + "`AWS_S3_FORCE_PATH_STYLE=true`" + `,
  boto's ` + "`s3={\"addressing_style\": \"path\"}`" + `). Virtual-host style will not resolve.
- **` + "`--public`" + ` is per bucket, not per object.** It serves *everything* in the bucket
  anonymously. Never put user data or anything secret in a public bucket.
- **` + "`rotate`" + ` invalidates the old secret immediately.** Roll the new one out before
  rotating, not after. Grants survive rotation.
- **Quotas** accept size suffixes: ` + "`--quota 500MB`" + `, ` + "`10GB`" + `, ` + "`2TB`" + `, or a plain byte
  count. ` + "`--quota-objects`" + ` caps the object count. Zero means unlimited. Usage figures
  from ` + "`buckets get`" + ` are sampled periodically, so they lag recent writes.

` + "`buckets delete`" + ` destroys every object in the bucket and prompts for the bucket name;
` + "`keys delete`" + ` prompts likewise. Both take ` + "`--force`" + ` for scripts. ` + "`list`" + `, ` + "`get`" + `,
` + "`create`" + `, ` + "`update`" + ` and ` + "`grant`" + ` all support ` + "`-o json`" + `.

### Observability

Read-only access to the organization's SigNoz instance, for orgs where a platform
operator has enabled it. ` + "`dina obs`" + ` is short for ` + "`dina observability`" + `. Pass
` + "`--org`" + ` only when the account has more than one organization.

` + "```bash" + `
dina obs status                                   # enabled? where is the UI?

# what is wrong right now
dina obs alerts                                   # firing alerts, most severe first
dina obs alerts show 3f9a                         # rule, history, exceptions, error logs
dina obs services                                 # rate, error rate, p99 per service

# dig in
dina obs exceptions --service api --since 24h
dina obs exceptions show <group-id> --since 24h   # stacktrace + trace id
dina obs logs --service api --severity error,fatal
dina obs logs --filter "body CONTAINS 'timeout'" --since 6h
dina obs traces --service api --errors
dina obs traces --min-duration 500ms
dina obs trace <trace-id>                         # span tree
dina obs operations api                           # slowest endpoints of a service

# resources
dina obs infra pods --filter "k8s.namespace.name = 'api'"
dina obs metrics http.server.request.duration --space-agg p99 --group-by service.name

# field names for filters
dina obs fields --signal logs
dina obs fields --signal traces --key service.name
` + "```" + `

**Investigating an alert** — start with ` + "`dina obs alerts`" + `, then
` + "`dina obs alerts show <fingerprint>`" + ` (a unique prefix is enough). It prints what
breached and by how much, how often the rule fired in the last 24h against the 24h
before, and the affected service's exceptions and error logs since shortly before it
started. Follow a trace id from there with ` + "`dina obs trace`" + `.

Things to know:

- **Filters are a SQL WHERE clause without the WHERE**:
  ` + "`service.name = 'api' AND http.response.status_code >= 500`" + `. Supported: comparisons,
  AND/OR/NOT, parentheses, IN, LIKE, ILIKE, CONTAINS, EXISTS, IS [NOT] NULL. Quote
  strings with single quotes. To search log text use ` + "`body CONTAINS 'x'`" + `; a bare
  string is rejected.
- **Field names follow OpenTelemetry**, not other log tools: ` + "`severity_text`" + ` not
  ` + "`level`" + `, ` + "`body`" + ` not ` + "`message`" + `, ` + "`service.name`" + ` not ` + "`service`" + `. An unknown field is
  rejected with a suggestion; when unsure, run ` + "`dina obs fields`" + ` first. Prefer the
  ` + "`--service`" + `, ` + "`--severity`" + `, ` + "`--errors`" + ` and ` + "`--min-duration`" + ` shorthands over writing filters.
- **Time ranges default to the last hour.** Widen with ` + "`--since 24h`" + ` or ` + "`7d`" + ` before
  concluding there is nothing. ` + "`--until`" + ` takes an RFC 3339 time.
- **Logs and traces page**: when more exist, stderr says ` + "`--cursor <c>`" + `; pass it back for
  older entries.
- **Read-only.** Alert rules, notification channels and dashboards are managed in the
  SigNoz UI (` + "`dina obs status`" + ` prints its address). ` + "`dina apps logs`" + ` still shows live
  pod logs for Dina apps; ` + "`dina obs logs`" + ` searches what was sent to SigNoz.
- "observability is not enabled for this organization" means a platform operator has
  to enable it; there is nothing to configure from the CLI.

Every ` + "`dina obs`" + ` command supports ` + "`-o json`" + `.

### Users (admin)

` + "```bash" + `
dina users list
dina users activate <user-id>
` + "```" + `

### Feedback

Submit bug reports, feature requests, or general feedback to the Sokkel Signals API. Bug reports work anonymously; feature requests and general feedback require ` + "`dina auth login`" + `.

` + "```bash" + `
# bug report (interactive form if flags are omitted)
dina feedback bug --title "..." --description "..." --severity high --context-file ./build.log

# read bug context from stdin (useful in pipelines)
some-command 2>&1 | dina feedback bug --title "..." --description "..." --context-file -

# feature request
dina feedback feature --title "..." --description "..."

# general feedback
dina feedback --message "..." --rating 5
` + "```" + `

Bug report flags: ` + "`--title`" + `, ` + "`--description`" + `, ` + "`--severity low|medium|high`" + `, ` + "`--context`" + ` (inline), ` + "`--context-file`" + ` (path, or ` + "`-`" + ` for stdin). OS and CLI version are attached automatically.

For non-interactive use, supply all required flags and pass ` + "`--no-input`" + `. The server returns a submission ID on stdout — capture it if the user wants a reference.

If a submission can't reach the server (offline, server down, not authenticated for ` + "`feature`" + `/general feedback), the body is **stored locally** under ` + "`~/.config/dina/feedback-queue/`" + ` so it isn't lost. ` + "`dina doctor`" + ` reports any pending items, and ` + "`dina doctor --fix`" + ` retries them and removes those that succeed. Don't tell the user "submission failed, please retry" — instead point them at ` + "`dina doctor --fix`" + ` once the underlying issue (e.g. login, network) is resolved.

### Other

` + "```bash" + `
dina version
dina update                 # install the latest CLI version
dina update --check         # check for updates without installing
dina install --skills       # install this agent skill for supported AI tools
dina doctor                 # run diagnostic checks
dina doctor --fix           # auto-repair any fixable issues
` + "```" + `

## Common flag patterns

The ` + "`-a`" + ` / ` + "`--app`" + ` flag specifies the app name. It is required for most commands:

` + "```bash" + `
dina deploy -a my-app
dina apps info -a my-app
dina apps logs -a my-app
dina apps env set -a my-app KEY=value
dina apps hostnames add -a my-app example.com
dina apps deployments -a my-app
dina apps delete -a my-app
` + "```" + `

## Global flags

These flags work on every command:

- ` + "`-o`" + ` / ` + "`--output`" + ` — output format, ` + "`text`" + ` (default) or ` + "`json`" + `. Use ` + "`-o json`" + ` on ` + "`apps list`" + `, ` + "`apps info`" + `, ` + "`apps deployments`" + `, ` + "`users list`" + `, ` + "`auth status`" + `, and the ` + "`storage`" + ` commands when you need to parse the output. JSON goes to stdout; progress messages stay on stderr so the JSON body is pipeable straight into ` + "`jq`" + `.
- ` + "`-q`" + ` / ` + "`--quiet`" + ` — suppress informational stderr lines (` + "`Fetching...`" + `, ` + "`Packaging...`" + `, post-action confirmations).
- ` + "`--no-input`" + ` — disable interactive prompts; required fields must be supplied as flags. Use this (together with ` + "`--force`" + ` where relevant) when running in scripts.
- ` + "`--no-color`" + ` — disable ANSI color. Also honors ` + "`NO_COLOR`" + ` and ` + "`TERM=dumb`" + `.
- ` + "`--debug`" + ` — log each outgoing HTTP request's method/URL and response status. Also honors ` + "`DEBUG=1`" + `.

## Error output

Errors go to stderr as a single line prefixed ` + "`error:`" + `. The CLI rewrites API errors into actionable text — for example a 401 reads ` + "`not authenticated or token expired — run: dina auth login`" + `, and a 404 on an app path reads ` + "`no app named 'X' — list apps with: dina apps list`" + `. Transport errors include the host that failed and mention ` + "`DINA_API_URL`" + ` for overriding it.

## Reporting bugs and soliciting feedback

Use the feedback commands proactively — they're the primary channel for the Sokkel team to hear about what works and what doesn't.

**File a bug with ` + "`dina feedback bug`" + ` when:**

- A command crashes, hangs, or returns an error that looks like a CLI bug rather than a user mistake
- A command's behavior doesn't match its ` + "`--help`" + ` text or this document
- You hit a confusing error message that doesn't explain what to do next
- The user explicitly asks you to file a bug

Always include the exact failing command and the relevant output via ` + "`--context`" + ` (short inline snippet) or ` + "`--context-file`" + ` (for logs). Use ` + "`--severity high`" + ` only for crashes, data loss, or a fully blocked workflow; otherwise leave it unset or use ` + "`medium`" + `/` + "`low`" + `. Don't file duplicate bugs within the same session.

**Submit a feature request with ` + "`dina feedback feature`" + ` when:**

- The user describes a workflow the CLI can't currently support
- You find yourself working around a missing capability to get the user's task done
- The user explicitly asks for a feature

**Send general feedback with ` + "`dina feedback`" + ` when:**

- The user wants to share a reaction to the CLI or platform
- A docs section or command description was unclear — cite what was confusing so the team can fix it
- The user is happy with something and wants to say so (include ` + "`--rating 5`" + ` when appropriate)

Before filing anything, run ` + "`dina <command> --help`" + ` to confirm the behavior isn't already documented, and skim this skill doc. Never file bug reports for user typos, expected 401s when not authenticated, or behavior that's explicitly documented.

## Example: Deploy a project from scratch

` + "```bash" + `
dina auth login
dina apps create my-api
dina apps env set -a my-api PORT=8080 DATABASE_URL=postgres://localhost/mydb
dina deploy -a my-api
dina apps logs -a my-api
dina apps hostnames add -a my-api api.example.com
` + "```" + `

## Example: Deploy a pre-built image

` + "```bash" + `
dina apps create my-service
dina deploy -a my-service --tag ghcr.io/org/my-service:v1.0 --replicas 2
dina apps info -a my-service
` + "```" + `

## Example: Debug a failing deployment

` + "```bash" + `
dina apps deployments -a my-app
dina apps deployments logs -a my-app --id <deployment-id>
dina apps logs -a my-app -n 200
` + "```" + `
`
}
