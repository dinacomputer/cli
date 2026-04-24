package skills

// SkillMD returns the contents of the SKILL.md file that teaches AI agents
// how to use the Dina CLI.
func SkillMD() string {
	return `---
name: dina-cli
description: Deploy applications, manage apps, view logs, set env vars, and configure hostnames on the Dina platform. Use when the user wants to deploy code, check app status, view logs, manage environment variables, configure custom domains, or perform any Dina platform operation.
---

# Dina CLI

## Introduction

Dina is a platform-as-a-service (PaaS) for deploying and managing containerized applications. With the Dina CLI you can:

- **Deploy** applications from source code (current directory) or pre-built container images
- **Manage apps**: create, update, delete, and inspect applications
- **View logs**: stream runtime logs and build logs for deployments
- **Configure environment variables** for your applications
- **Manage custom hostnames** for your apps
- **Manage users** (admin operations)
- **Report bugs and send feedback** to the Sokkel team

## Session start

At the beginning of every session, run:

` + "```bash" + `
dina help
` + "```" + `

This prints the current list of available commands and their descriptions. Always run this before issuing any other Dina command so you know exactly which commands and flags are available in the installed version. For sub-command details, run ` + "`dina [command] --help`" + `.

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

# delete an app
dina apps delete -a my-app
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
dina apps hostnames remove -a my-app example.com
` + "```" + `

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

# feature request
dina feedback feature --title "..." --description "..."

# general feedback
dina feedback --message "..." --rating 5
` + "```" + `

Bug report flags: ` + "`--title`" + `, ` + "`--description`" + `, ` + "`--severity low|medium|high`" + `, ` + "`--context`" + ` (inline), ` + "`--context-file`" + ` (path). OS and CLI version are attached automatically.

For non-interactive use, supply all required flags and pass ` + "`--no-input`" + `. The server returns a submission ID on stdout — capture it if the user wants a reference.

### Other

` + "```bash" + `
dina version
dina install --skills
dina doctor               # run diagnostic checks
dina doctor --fix         # auto-repair any fixable issues
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
