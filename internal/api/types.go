package api

import "time"

// ---------- Error ----------

type ErrorModel struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title,omitempty"`
	Status int    `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// ---------- App ----------

type CreateAppInput struct {
	Name string `json:"name"`
}

type App struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	OwnerID   string     `json:"owner_id"`
	URL       string     `json:"url"`
	Namespace string     `json:"namespace"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Hostnames []Hostname `json:"hostnames,omitempty"`
}

type ListAppsOutput struct {
	Apps []GetAppOutput `json:"apps"`
}

type GetAppOutput struct {
	App              App         `json:"app"`
	LatestDeployment *Deployment `json:"latest_deployment,omitempty"`
}

type UpdateAppInput struct {
	Name string `json:"name,omitempty"`
}

// ---------- Deploy ----------

type DeployInput struct {
	Image    string            `json:"image,omitempty"`
	Port     *int              `json:"port,omitempty"`
	Replicas *int              `json:"replicas,omitempty"`
	BuildArgs map[string]string `json:"build_args,omitempty"`
}

type Deployment struct {
	ID         string    `json:"id"`
	AppID      string    `json:"app_id"`
	SourceType string    `json:"source_type"`
	SourceRef  string    `json:"source_ref"`
	Status     string    `json:"status"`
	Port       int       `json:"port"`
	Replicas   int       `json:"replicas"`
	Image      string    `json:"image,omitempty"`
	BuildLog   string    `json:"build_log,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListDeploymentsOutput struct {
	Deployments []Deployment `json:"deployments"`
}

type DeploymentLogsOutput struct {
	BuildLog string `json:"build_log"`
}

// ---------- Logs ----------

type PodLogsOutput struct {
	Logs string `json:"logs"`
}

// ---------- Env ----------

type PutEnvVarsInput struct {
	Vars map[string]string `json:"vars"`
}

type PutEnvVarsOutput struct {
	Vars []EnvVar `json:"vars"`
}

type EnvVar struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------- Hostnames ----------

type AddHostnameInput struct {
	Hostname string `json:"hostname"`
}

type Hostname struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Hostname  string    `json:"hostname"`
	CreatedAt time.Time `json:"created_at"`
}

// ---------- Users ----------

type ListUsersOutput struct {
	Users []User `json:"users"`
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------- Clusters ----------

type Cluster struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	DisplayName       string    `json:"display_name,omitempty"`
	Region            string    `json:"region,omitempty"`
	Status            string    `json:"status"`
	KubernetesVersion string    `json:"kubernetes_version,omitempty"`
	// Endpoint and CACertificate are only populated by GetCluster — the list
	// endpoint omits them. CACertificate is base64-encoded PEM, ready to drop
	// into kubeconfig's certificate-authority-data.
	Endpoint      string    `json:"endpoint,omitempty"`
	CACertificate string    `json:"ca_certificate,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type ListClustersOutput struct {
	Clusters []Cluster `json:"clusters"`
}

// ---------- Organizations ----------

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IdpOrgID  string    `json:"idp_org_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListOrganizationsOutput struct {
	Organizations []Organization `json:"organizations"`
}

// ---------- Federations (workload identity) ----------

// Federation registers an external OIDC issuer (a CI provider) whose tokens can
// be exchanged for short-lived Dina credentials via workload identity federation.
type Federation struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ClientID       string    `json:"client_id"`
	Name           string    `json:"name"`
	Issuer         string    `json:"issuer"`
	Audiences      []string  `json:"audiences"`
	SubjectClaim   string    `json:"subject_claim"`
	Disabled       bool      `json:"disabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListFederationsOutput struct {
	Federations []Federation `json:"federations"`
}

type CreateFederationInput struct {
	Name         string   `json:"name"`
	Issuer       string   `json:"issuer"`
	Audiences    []string `json:"audiences,omitempty"`
	SubjectClaim string   `json:"subject_claim,omitempty"`
}

type UpdateFederationInput struct {
	Name         string   `json:"name,omitempty"`
	Audiences    []string `json:"audiences,omitempty"`
	SubjectClaim string   `json:"subject_claim,omitempty"`
	Disabled     *bool    `json:"disabled,omitempty"`
}

// Mapping binds a federation's token claims to a set of granted scopes. All
// MatchClaims must match (glob-aware) for the mapping's scopes to apply.
type Mapping struct {
	ID           string            `json:"id"`
	FederationID string            `json:"federation_id"`
	Name         string            `json:"name"`
	MatchClaims  map[string]string `json:"match_claims"`
	Scopes       []string          `json:"scopes"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ListMappingsOutput struct {
	Mappings []Mapping `json:"mappings"`
}

type CreateMappingInput struct {
	Name        string            `json:"name,omitempty"`
	MatchClaims map[string]string `json:"match_claims,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
}

// SessionToken is the caller's current Dina access token and its expiry. The
// clusters exec credential plugin hands this to kubectl, which authenticates to
// Dina's cluster proxy with the same token the CLI uses for /api/v1.
type SessionToken struct {
	Token  string
	Expiry time.Time
}

// ---------- Signals ----------

type BugBody struct {
	Product     string `json:"product"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity,omitempty"`
	OS          string `json:"os,omitempty"`
	Version     string `json:"version,omitempty"`
	Context     string `json:"context,omitempty"`
}

type FeatureRequestBody struct {
	Product     string `json:"product"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type FeedbackBody struct {
	Product string `json:"product"`
	Message string `json:"message"`
	Rating  int    `json:"rating,omitempty"`
}

type SubmitResponseBody struct {
	ID string `json:"id"`
}

// ---------- Object storage ----------

// StorageEndpoint is the S3-compatible endpoint a bucket is served from.
type StorageEndpoint struct {
	URL       string `json:"url"`
	Region    string `json:"region"`
	PathStyle bool   `json:"path_style"`
}

// StorageUsage is the last observed consumption for a bucket. The backend
// samples it periodically, so ObservedAt can lag recent writes.
type StorageUsage struct {
	Bytes      int64     `json:"bytes"`
	Objects    int64     `json:"objects"`
	ObservedAt time.Time `json:"observed_at"`
}

type Bucket struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	BackendID string `json:"backend_id"`
	Name      string `json:"name"`
	// PhysicalName is the bucket name on the backend, which is namespaced and
	// therefore differs from Name. S3 clients must address this one.
	PhysicalName string          `json:"physical_name"`
	Status       string          `json:"status"`
	Public       bool            `json:"public"`
	PublicURL    string          `json:"public_url,omitempty"`
	QuotaBytes   int64           `json:"quota_bytes"`
	QuotaObjects int64           `json:"quota_objects"`
	Endpoint     StorageEndpoint `json:"endpoint"`
	Usage        *StorageUsage   `json:"usage,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ListBucketsOutput struct {
	Buckets []Bucket `json:"buckets"`
}

type CreateBucketInput struct {
	Name         string `json:"name"`
	Backend      string `json:"backend,omitempty"`
	Public       bool   `json:"public,omitempty"`
	QuotaBytes   int64  `json:"quota_bytes,omitempty"`
	QuotaObjects int64  `json:"quota_objects,omitempty"`
}

// UpdateBucketInput uses pointers throughout so an unset field is omitted
// rather than sent as a zero value the server would apply.
type UpdateBucketInput struct {
	Public       *bool  `json:"public,omitempty"`
	QuotaBytes   *int64 `json:"quota_bytes,omitempty"`
	QuotaObjects *int64 `json:"quota_objects,omitempty"`
}

// StorageGrant is one key's permissions on one bucket.
type StorageGrant struct {
	KeyID      string    `json:"key_id"`
	KeyName    string    `json:"key_name,omitempty"`
	BucketID   string    `json:"bucket_id"`
	BucketName string    `json:"bucket_name,omitempty"`
	Read       bool      `json:"read"`
	Write      bool      `json:"write"`
	Owner      bool      `json:"owner"`
	CreatedAt  time.Time `json:"created_at"`
}

// StorageKey is an S3 access key. The secret is returned only by
// CreateStorageKey and RotateStorageKey — it is never readable afterwards.
type StorageKey struct {
	ID          string         `json:"id"`
	OrgID       string         `json:"org_id"`
	BackendID   string         `json:"backend_id"`
	Name        string         `json:"name"`
	AccessKeyID string         `json:"access_key_id"`
	Grants      []StorageGrant `json:"grants,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ListStorageKeysOutput struct {
	Keys []StorageKey `json:"keys"`
}

type CreateStorageKeyInput struct {
	Name    string `json:"name"`
	Backend string `json:"backend,omitempty"`
}

// IssuedStorageKey carries the one-time secret alongside the key record.
type IssuedStorageKey struct {
	Key             StorageKey `json:"key"`
	SecretAccessKey string     `json:"secret_access_key"`
}

// SetGrantInput always sends read and write; the API requires both.
type SetGrantInput struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Owner bool `json:"owner,omitempty"`
}
