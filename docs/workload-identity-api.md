# Workload identity federation

Lets CI (GitHub Actions, GitLab CI, Forgejo/Gitea Actions, …) authenticate to
Dina **keyless** — no stored `DINA_API_TOKEN`. The CI provider mints a
short-lived OIDC token describing the job; Dina validates it against a
**federation** (a trusted issuer) and its **mappings** (claim → scope rules),
then exchanges it for a short-lived Dina access token. The resulting token flows
through the same proxy/impersonation path as a human's, just as a machine
principal.

Both halves are implemented:

- **CLI login** — `dina auth login --federated` obtains the OIDC token and runs
  the token exchange.
- **CLI management** — `dina identity federation …` and `dina identity mapping …`
  administer the trust policy over the Dina API.

## Runtime: RFC 8693 token exchange

Happens on the auth server (`id.sokkel.io`, the same one device/PKCE use — the
CLI discovers `token_endpoint` via OIDC discovery):

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:token-exchange
subject_token=<CI OIDC JWT>
subject_token_type=urn:ietf:params:oauth:token-type:jwt
audience=<the API resource — https://dina.sh by default>
client_id=d35bfdfa-f03c-481b-a3be-aaf96570611e
```

The `audience` (and the `aud` the CLI requests for the CI token) defaults to the
**API resource** from RFC 9728 metadata (`https://dina.sh`), *not* the auth
issuer. A federation created without explicit `audiences` requires exactly this
server origin, so the default must match it or the exchange is rejected.

Response is the standard token body (`access_token`, `token_type`,
`expires_in`; **no** `refresh_token` — CI re-federates each run):

```json
{ "access_token": "skkl_at_...", "token_type": "Bearer", "expires_in": 300 }
```

`400`/`401` if validation fails; `403` if the identity is valid but no mapping
grants it access. The CLI surfaces the RFC 6749 `error_description` from the body.

## Validation (the security-critical part)

For the `subject_token`:

1. **Signature** — verify against the issuer's JWKS (`iss` →
   `/.well-known/openid-configuration` → `jwks_uri`). Cache keys, handle rotation.
2. **`aud`** — MUST be in the federation's `audiences` (defaults to the server
   origin). Without audience pinning any token the provider minted for another
   service can be replayed here (confused deputy).
3. **`iss`** — MUST match a registered federation.
4. **`exp` / `nbf` / `iat`** — standard time checks.
5. **Claim binding** — the token's claims must satisfy a mapping's `match_claims`.

## Data model (Dina API)

Federations and mappings are org-scoped under `/organizations/{orgID}`.

**Federation** — a trusted OIDC issuer.

| Field | Notes |
|---|---|
| `name` | human-readable label |
| `issuer` | external OIDC issuer URL |
| `audiences[]` | accepted `aud` values; empty ⇒ requires the server origin |
| `subject_claim` | claim identifying the principal (default `sub`) |
| `client_id` | assigned by Dina; the CLI sends it in the exchange |
| `disabled` | toggles the federation without deleting it |

Endpoints: `GET|POST /organizations/{orgID}/federations`,
`GET|PATCH|DELETE /organizations/{orgID}/federations/{id}`.

**Mapping** — a claim → scope rule within a federation.

| Field | Notes |
|---|---|
| `match_claims` | `claim → glob`; all must match |
| `scopes[]` | scopes granted when matched (within the org's ceiling) |
| `name` | optional label |

Endpoints: `GET|POST /organizations/{orgID}/federations/{id}/mappings`,
`DELETE /organizations/{orgID}/federations/{id}/mappings/{mappingID}`.

Per-provider claims to bind on (the `sub` differs, so match the structured
claims):

| Provider | `iss` | Key claims |
|---|---|---|
| GitHub Actions | `https://token.actions.githubusercontent.com` | `repository`, `repository_owner`, `ref`, `environment`, `job_workflow_ref` |
| GitLab CI | `https://gitlab.com` or the self-managed URL | `project_path`, `namespace_path`, `ref`, `ref_type`, `environment` |
| Forgejo/Gitea Actions | the instance URL (per customer) | `repository`, `ref` (GitHub-compatible layout) |

Self-managed GitLab/Forgejo means per-customer issuer registration and JWKS
reachability to those instances — aligned with the per-customer cluster model.

## Example: trust GitHub Actions on `main`

```sh
# 1. register the issuer
dina identity federation create \
  --name github-actions \
  --issuer https://token.actions.githubusercontent.com

# 2. grant deploy scope to this repo's main branch
dina identity mapping create --federation <fed-id> \
  --name deploy-main \
  --match repository=dinacomputer/cli \
  --match ref=refs/heads/main \
  --scope deploy
```

In the workflow (needs `permissions: id-token: write`):

```yaml
- run: dina auth login --federated
```
