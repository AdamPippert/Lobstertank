package template

import (
	"fmt"
	"net"
	"strings"
)

// ValidationError collects all problems found in a template.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("template validation failed (%d errors):\n  - %s",
		len(e.Errors), strings.Join(e.Errors, "\n  - "))
}

// Validate checks a resolved template for structural and cross-field correctness.
func Validate(t *Template) error {
	var errs []string
	add := func(msg string, args ...any) {
		errs = append(errs, fmt.Sprintf(msg, args...))
	}

	// --- Metadata ---
	if t.APIVersion == "" {
		add("apiVersion is required")
	}
	if t.Kind == "" {
		add("kind is required")
	}
	if t.Metadata.Name == "" {
		add("metadata.name is required")
	}

	// --- Identity ---
	id := &t.Spec.Identity
	if id.Role != "" && !isValidRole(id.Role) {
		add("identity.role %q is not a recognized role (valid: %s)", id.Role, validRolesStr())
	}

	// --- Runtime ---
	rt := &t.Spec.Runtime
	if rt.DeploymentTarget != "" && !isValidTarget(rt.DeploymentTarget) {
		add("runtime.deployment_target %q is not recognized (valid: %s)", rt.DeploymentTarget, validTargetsStr())
	}

	// --- Network ---
	nw := &t.Spec.Network
	if nw.BindAddress != "" {
		if ip := net.ParseIP(nw.BindAddress); ip == nil {
			add("network.bind_address %q is not a valid IP address", nw.BindAddress)
		}
	}
	portNames := make(map[string]int)
	portNumbers := make(map[int]string)
	for i, p := range nw.Ports {
		if p.Name == "" {
			add("network.ports[%d].name is required", i)
		}
		if p.Port <= 0 || p.Port > 65535 {
			add("network.ports[%d].port %d is out of range (1-65535)", i, p.Port)
		}
		if p.HostPort != 0 && (p.HostPort <= 0 || p.HostPort > 65535) {
			add("network.ports[%d].host_port %d is out of range (1-65535)", i, p.HostPort)
		}
		if prev, ok := portNames[p.Name]; ok {
			add("network.ports[%d].name %q duplicates ports[%d]", i, p.Name, prev)
		}
		portNames[p.Name] = i
		if existing, ok := portNumbers[p.Port]; ok {
			add("network.ports[%d].port %d conflicts with port named %q", i, p.Port, existing)
		}
		portNumbers[p.Port] = p.Name
	}
	if nw.PortsMergeStrategy != "" && !isValidMergeStrategy(nw.PortsMergeStrategy) {
		add("network.ports_merge_strategy %q is invalid (valid: replace, append, union)", nw.PortsMergeStrategy)
	}

	// --- Secrets ---
	sec := &t.Spec.Secrets
	if sec.Provider != "" && !isValidSecretsProvider(sec.Provider) {
		add("secrets.provider %q is not recognized (valid: builtin, vault, k8s_secrets)", sec.Provider)
	}
	for i, e := range sec.Entries {
		if e.Name == "" {
			add("secrets.entries[%d].name is required", i)
		}
		if e.Type == "" {
			add("secrets.entries[%d].type is required", i)
		} else if !isValidSecretType(e.Type) {
			add("secrets.entries[%d].type %q is not recognized", i, e.Type)
		}
		if e.Ref == "" {
			add("secrets.entries[%d].ref is required (must be a URI reference, never a plaintext value)", i)
		}
	}
	if sec.EntriesMergeStrategy != "" && !isValidMergeStrategy(sec.EntriesMergeStrategy) {
		add("secrets.entries_merge_strategy %q is invalid", sec.EntriesMergeStrategy)
	}

	// --- Observability ---
	obs := &t.Spec.Observability
	if obs.Logging.Level != "" && !isValidLogLevel(obs.Logging.Level) {
		add("observability.logging.level %q is invalid (valid: debug, info, warn, error)", obs.Logging.Level)
	}
	if obs.Logging.Format != "" && obs.Logging.Format != "json" && obs.Logging.Format != "text" {
		add("observability.logging.format %q is invalid (valid: json, text)", obs.Logging.Format)
	}

	// --- Policy ---
	pol := &t.Spec.Policy
	if pol.UpdateChannel != "" && !isValidUpdateChannel(pol.UpdateChannel) {
		add("policy.update_channel %q is invalid (valid: stable, beta, nightly, pinned)", pol.UpdateChannel)
	}

	// --- Cross-field checks ---
	if id.Role == RoleGateway && len(nw.Ports) == 0 {
		add("gateway role requires at least one network port")
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// --- lookup tables ---

var validRoles = map[string]bool{
	RoleGateway:      true,
	RoleControlPlane: true,
	RoleWorker:       true,
	RoleObservability: true,
	RoleDBAdjunct:    true,
}

var validTargets = map[string]bool{
	TargetPodman:     true,
	TargetKubernetes: true,
	TargetOpenShift:  true,
	TargetDroplet:    true,
	TargetSandbox:    true,
}

var validSecretsProviders = map[string]bool{
	"builtin":     true,
	"vault":       true,
	"k8s_secrets": true,
}

var validSecretTypes = map[string]bool{
	"api_key":    true,
	"node_key":   true,
	"oauth_cred": true,
	"db_cred":    true,
	"cert":       true,
}

var validMergeStrategies = map[string]bool{
	MergeStrategyReplace: true,
	MergeStrategyAppend:  true,
	MergeStrategyUnion:   true,
}

var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

var validUpdateChannels = map[string]bool{
	"stable":  true,
	"beta":    true,
	"nightly": true,
	"pinned":  true,
}

func isValidRole(s string) bool            { return validRoles[s] }
func isValidTarget(s string) bool           { return validTargets[s] }
func isValidSecretsProvider(s string) bool   { return validSecretsProviders[s] }
func isValidSecretType(s string) bool        { return validSecretTypes[s] }
func isValidMergeStrategy(s string) bool     { return validMergeStrategies[s] }
func isValidLogLevel(s string) bool          { return validLogLevels[s] }
func isValidUpdateChannel(s string) bool     { return validUpdateChannels[s] }

func validRolesStr() string {
	return strings.Join([]string{RoleGateway, RoleControlPlane, RoleWorker, RoleObservability, RoleDBAdjunct}, ", ")
}

func validTargetsStr() string {
	return strings.Join([]string{TargetPodman, TargetKubernetes, TargetOpenShift, TargetDroplet, TargetSandbox}, ", ")
}
