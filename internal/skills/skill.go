package skills

// SkillMD returns the contents of the SKILL.md file that teaches AI agents
// how to use the Dina CLI.
func SkillMD() string {
	return `---
name: dina-cli
description: >
  Use the Dina CLI to deploy applications, view logs, manage env vars, hostnames,
  and deployments on the Dina platform. Activate this skill when the user mentions
  deploying, checking app status, viewing logs, setting environment variables,
  managing hostnames, or any Dina platform operation.
metadata:
  author: dinacomputer
  version: "0.1"
---

# Dina CLI

The Dina CLI (` + "`dina`" + `) is the command-line tool for the Dina platform.

## Authentication

Set the ` + "`DINA_API_TOKEN`" + ` environment variable before using the CLI.
Optionally set ` + "`DINA_API_URL`" + ` to override the default API endpoint (` + "`https://dina.sh/api/v1`" + `).

## Commands

### dina deploy

Deploy an application. If ` + "`--tag`" + ` is provided, deploys a pre-built image.
Otherwise, the current directory is zipped and uploaded as source code.

` + "```bash" + `
# Deploy from source (zips current directory)
dina deploy -a my-app

# Deploy a pre-built image
dina deploy -a my-app --tag registry.example.com/my-app:v1.2

# Deploy with a specific replica count
dina deploy -a my-app --replicas 3
` + "```" + `

| Flag       | Short | Required | Description                                    |
|------------|-------|----------|------------------------------------------------|
| --app      | -a    | Yes      | Application name                               |
| --tag      | -t    | No       | Pre-built image tag (omit to upload source)    |
| --replicas |       | No       | Number of replicas (default: 1)                |

### dina apps create

Create a new application.

` + "```bash" + `
dina apps create my-app
` + "```" + `

### dina apps info

Show application details including hostnames and latest deployment.

` + "```bash" + `
dina apps info -a my-app
` + "```" + `

### dina apps delete

Delete an application.

` + "```bash" + `
dina apps delete -a my-app
` + "```" + `

### dina apps logs

View runtime logs for an application.

` + "```bash" + `
dina apps logs -a my-app
dina apps logs -a my-app -n 50
` + "```" + `

| Flag    | Short | Required | Description                        |
|---------|-------|----------|------------------------------------|
| --app   | -a    | Yes      | Application name                   |
| --lines | -n    | No       | Number of log lines (default: 100) |

### dina apps deployments

List all deployments for an application.

` + "```bash" + `
dina apps deployments -a my-app
` + "```" + `

### dina apps deployments logs

View build logs for a specific deployment.

` + "```bash" + `
dina apps deployments logs -a my-app --id <deployment-id>
` + "```" + `

### dina apps env set

Set environment variables (KEY=VALUE pairs).

` + "```bash" + `
dina apps env set -a my-app DATABASE_URL=postgres://... SECRET_KEY=abc123
` + "```" + `

### dina apps hostnames add

Add a custom hostname to an application.

` + "```bash" + `
dina apps hostnames add -a my-app example.com
` + "```" + `

### dina apps hostnames remove

Remove a custom hostname.

` + "```bash" + `
dina apps hostnames remove -a my-app example.com
` + "```" + `

## Typical workflow

1. Create the app: ` + "`dina apps create my-app`" + `
2. Set env vars: ` + "`dina apps env set -a my-app PORT=8080`" + `
3. Deploy: ` + "`dina deploy -a my-app`" + `
4. Check logs: ` + "`dina apps logs -a my-app`" + `
5. Add a custom domain: ` + "`dina apps hostnames add -a my-app mydomain.com`" + `
`
}
