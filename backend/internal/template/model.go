// Package template implements OpenClaw encoding templates — layered
// installation profiles that compile into deploy-ready bundles.
package template

// Supported Kind values for template documents.
const (
	KindEncodingTemplate  = "EncodingTemplate"
	KindRoleOverlay       = "RoleOverlay"
	KindEnvironmentOverlay = "EnvironmentOverlay"
	KindInstanceVars      = "InstanceVars"

	APIVersion = "lobstertank.io/v1alpha1"
)

// Supported roles.
const (
	RoleGateway      = "gateway"
	RoleControlPlane = "control-plane"
	RoleWorker       = "worker"
	RoleObservability = "observability"
	RoleDBAdjunct    = "db-adjunct"
)

// Supported deployment targets.
const (
	TargetPodman    = "podman"
	TargetKubernetes = "kubernetes"
	TargetOpenShift = "openshift"
	TargetDroplet   = "droplet"
	TargetSandbox   = "sandbox"
)

// Supported merge strategies for array fields.
const (
	MergeStrategyReplace = "replace"
	MergeStrategyAppend  = "append"
	MergeStrategyUnion   = "union"
)

// Template is the top-level encoding template document.
type Template struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata provides identity and versioning for a template document.
type Metadata struct {
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version,omitempty" json:"version,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// Spec holds the full configuration surface of an OpenClaw install.
type Spec struct {
	Identity      IdentitySpec      `yaml:"identity" json:"identity"`
	Runtime       RuntimeSpec       `yaml:"runtime" json:"runtime"`
	Network       NetworkSpec       `yaml:"network" json:"network"`
	Secrets       SecretsSpec       `yaml:"secrets" json:"secrets"`
	Observability ObservabilitySpec `yaml:"observability" json:"observability"`
	Policy        PolicySpec        `yaml:"policy" json:"policy"`
}

// IdentitySpec defines the instance identity and role.
type IdentitySpec struct {
	InstanceName string            `yaml:"instance_name,omitempty" json:"instance_name,omitempty"`
	Role         string            `yaml:"role,omitempty" json:"role,omitempty"`
	Region       string            `yaml:"region,omitempty" json:"region,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	TailnetTags  []string          `yaml:"tailnet_tags,omitempty" json:"tailnet_tags,omitempty"`
}

// RuntimeSpec defines the deployment target and resource hints.
type RuntimeSpec struct {
	DeploymentTarget string       `yaml:"deployment_target,omitempty" json:"deployment_target,omitempty"`
	Resources        ResourceSpec `yaml:"resources,omitempty" json:"resources,omitempty"`
	Image            ImageSpec    `yaml:"image,omitempty" json:"image,omitempty"`
}

// ResourceSpec defines resource requests/limits.
type ResourceSpec struct {
	CPU          string   `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory       string   `yaml:"memory,omitempty" json:"memory,omitempty"`
	StoragePaths []string `yaml:"storage_paths,omitempty" json:"storage_paths,omitempty"`
}

// ImageSpec defines the container image to use.
type ImageSpec struct {
	Repository string `yaml:"repository,omitempty" json:"repository,omitempty"`
	Tag        string `yaml:"tag,omitempty" json:"tag,omitempty"`
	PullPolicy string `yaml:"pull_policy,omitempty" json:"pull_policy,omitempty"`
}

// NetworkSpec defines networking configuration.
type NetworkSpec struct {
	Ports        []PortSpec        `yaml:"ports,omitempty" json:"ports,omitempty"`
	BindAddress  string            `yaml:"bind_address,omitempty" json:"bind_address,omitempty"`
	ReverseProxy *ReverseProxySpec `yaml:"reverse_proxy,omitempty" json:"reverse_proxy,omitempty"`
	Tailscale    *TailscaleSpec    `yaml:"tailscale,omitempty" json:"tailscale,omitempty"`
	MultiGateway *MultiGatewaySpec `yaml:"multi_gateway,omitempty" json:"multi_gateway,omitempty"`

	// MergeStrategy controls how the ports array is combined during overlay.
	// Allowed values: replace (default), append, union.
	PortsMergeStrategy string `yaml:"ports_merge_strategy,omitempty" json:"ports_merge_strategy,omitempty"`
}

// PortSpec defines a network port binding.
type PortSpec struct {
	Name     string `yaml:"name" json:"name"`
	Port     int    `yaml:"port" json:"port"`
	HostPort int    `yaml:"host_port,omitempty" json:"host_port,omitempty"`
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
}

// ReverseProxySpec defines reverse proxy expectations.
type ReverseProxySpec struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"` // nginx, caddy, traefik
	TLS      bool   `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// TailscaleSpec defines Tailscale/Headscale networking.
type TailscaleSpec struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Provider   string            `yaml:"provider,omitempty" json:"provider,omitempty"` // tailscale, headscale
	Tags       []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	ACLMapping map[string]string `yaml:"acl_mapping,omitempty" json:"acl_mapping,omitempty"`
}

// MultiGatewaySpec defines multi-gateway topology.
type MultiGatewaySpec struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Mode     string `yaml:"mode,omitempty" json:"mode,omitempty"` // active-standby, priority
	Priority int    `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// SecretsSpec defines secret management configuration.
type SecretsSpec struct {
	Provider string       `yaml:"provider,omitempty" json:"provider,omitempty"` // builtin, vault, k8s_secrets
	Entries  []SecretEntry `yaml:"entries,omitempty" json:"entries,omitempty"`

	// EntriesMergeStrategy controls how entries are combined during overlay.
	EntriesMergeStrategy string `yaml:"entries_merge_strategy,omitempty" json:"entries_merge_strategy,omitempty"`
}

// SecretEntry defines a single secret reference (never stores the value).
type SecretEntry struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"` // api_key, node_key, oauth_cred, db_cred, cert
	Ref  string `yaml:"ref" json:"ref"`   // URI: builtin://..., vault://..., k8s://ns/secret/key
}

// ObservabilitySpec defines logging, metrics, traces, and health checks.
type ObservabilitySpec struct {
	Logging     LoggingSpec     `yaml:"logging,omitempty" json:"logging,omitempty"`
	Metrics     MetricsSpec     `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Traces      TracesSpec      `yaml:"traces,omitempty" json:"traces,omitempty"`
	HealthCheck HealthCheckSpec `yaml:"health_check,omitempty" json:"health_check,omitempty"`
}

// LoggingSpec defines log output configuration.
type LoggingSpec struct {
	Destinations   []string `yaml:"destinations,omitempty" json:"destinations,omitempty"`
	RedactionRules []string `yaml:"redaction_rules,omitempty" json:"redaction_rules,omitempty"`
	Level          string   `yaml:"level,omitempty" json:"level,omitempty"`
	Format         string   `yaml:"format,omitempty" json:"format,omitempty"` // json, text
}

// MetricsSpec defines metrics export configuration.
type MetricsSpec struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Format   string `yaml:"format,omitempty" json:"format,omitempty"` // prometheus, otlp
}

// TracesSpec defines distributed tracing configuration.
type TracesSpec struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Format   string `yaml:"format,omitempty" json:"format,omitempty"` // otlp, zipkin
}

// HealthCheckSpec defines readiness and liveness probes.
type HealthCheckSpec struct {
	Path          string `yaml:"path,omitempty" json:"path,omitempty"`
	Interval      string `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout       string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	ReadinessGate bool   `yaml:"readiness_gate,omitempty" json:"readiness_gate,omitempty"`
}

// PolicySpec defines operational guardrails.
type PolicySpec struct {
	UpdateChannel       string   `yaml:"update_channel,omitempty" json:"update_channel,omitempty"`
	PinnedVersion       string   `yaml:"pinned_version,omitempty" json:"pinned_version,omitempty"`
	ApprovedPlugins     []string `yaml:"approved_plugins,omitempty" json:"approved_plugins,omitempty"`
	ApprovedProviders   []string `yaml:"approved_providers,omitempty" json:"approved_providers,omitempty"`
	FilesystemAllowlist []string `yaml:"filesystem_allowlist,omitempty" json:"filesystem_allowlist,omitempty"`
	CommandAllowlist    []string `yaml:"command_allowlist,omitempty" json:"command_allowlist,omitempty"`
}

// InstallBundle is the rendered output of a resolved template.
type InstallBundle struct {
	// SourceHash is the SHA-256 of the deterministic JSON encoding
	// of the resolved template. Same inputs always produce the same hash.
	SourceHash string `json:"source_hash"`

	// ResolvedTemplate is the fully merged template.
	ResolvedTemplate *Template `json:"resolved_template"`

	// Files maps relative path → content for every artifact in the bundle.
	Files map[string][]byte `json:"-"`
}
