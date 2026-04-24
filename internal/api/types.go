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
