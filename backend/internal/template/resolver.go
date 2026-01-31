package template

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Resolve merges a base template with optional overlays and instance variables.
//
// Merge order (later wins for scalar fields):
//
//	base → role → environment → instance vars
//
// Array merge behaviour is controlled per-field via *_merge_strategy
// annotations. The default is "replace" — the overlay's array replaces the
// base entirely. Supported strategies: replace, append, union.
func Resolve(base *Template, layers ...*Template) (*Template, error) {
	if base == nil {
		return nil, fmt.Errorf("base template is required")
	}

	// Deep-copy the base so callers keep an unmodified original.
	result, err := deepCopy(base)
	if err != nil {
		return nil, fmt.Errorf("copy base template: %w", err)
	}

	// Force identity on the result.
	result.APIVersion = APIVersion
	result.Kind = KindEncodingTemplate

	for _, layer := range layers {
		if layer == nil {
			continue
		}
		mergeSpec(&result.Spec, &layer.Spec)
		mergeMetadataLabels(result, layer)
	}

	return result, nil
}

// Hash returns a deterministic SHA-256 hex digest over the resolved template.
func Hash(t *Template) (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("marshal for hash: %w", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// --- internal merge helpers ---

func mergeSpec(dst, src *Spec) {
	mergeIdentity(&dst.Identity, &src.Identity)
	mergeRuntime(&dst.Runtime, &src.Runtime)
	mergeNetwork(&dst.Network, &src.Network)
	mergeSecrets(&dst.Secrets, &src.Secrets)
	mergeObservability(&dst.Observability, &src.Observability)
	mergePolicy(&dst.Policy, &src.Policy)
}

func mergeMetadataLabels(dst, src *Template) {
	if len(src.Metadata.Labels) == 0 {
		return
	}
	if dst.Metadata.Labels == nil {
		dst.Metadata.Labels = make(map[string]string)
	}
	for k, v := range src.Metadata.Labels {
		dst.Metadata.Labels[k] = v
	}
}

func mergeIdentity(dst, src *IdentitySpec) {
	setIfNonEmpty(&dst.InstanceName, src.InstanceName)
	setIfNonEmpty(&dst.Role, src.Role)
	setIfNonEmpty(&dst.Region, src.Region)
	mergeMaps(&dst.Labels, src.Labels)
	if len(src.TailnetTags) > 0 {
		dst.TailnetTags = src.TailnetTags
	}
}

func mergeRuntime(dst, src *RuntimeSpec) {
	setIfNonEmpty(&dst.DeploymentTarget, src.DeploymentTarget)
	mergeResource(&dst.Resources, &src.Resources)
	mergeImage(&dst.Image, &src.Image)
}

func mergeResource(dst, src *ResourceSpec) {
	setIfNonEmpty(&dst.CPU, src.CPU)
	setIfNonEmpty(&dst.Memory, src.Memory)
	if len(src.StoragePaths) > 0 {
		dst.StoragePaths = src.StoragePaths
	}
}

func mergeImage(dst, src *ImageSpec) {
	setIfNonEmpty(&dst.Repository, src.Repository)
	setIfNonEmpty(&dst.Tag, src.Tag)
	setIfNonEmpty(&dst.PullPolicy, src.PullPolicy)
}

func mergeNetwork(dst, src *NetworkSpec) {
	setIfNonEmpty(&dst.BindAddress, src.BindAddress)

	strategy := coalesce(src.PortsMergeStrategy, dst.PortsMergeStrategy, MergeStrategyReplace)
	if len(src.Ports) > 0 {
		dst.Ports = mergePortSlice(dst.Ports, src.Ports, strategy)
	}
	// Propagate merge strategy annotation.
	if src.PortsMergeStrategy != "" {
		dst.PortsMergeStrategy = src.PortsMergeStrategy
	}

	if src.ReverseProxy != nil {
		cp := *src.ReverseProxy
		dst.ReverseProxy = &cp
	}
	if src.Tailscale != nil {
		merged := mergeTailscale(dst.Tailscale, src.Tailscale)
		dst.Tailscale = &merged
	}
	if src.MultiGateway != nil {
		cp := *src.MultiGateway
		dst.MultiGateway = &cp
	}
}

func mergeTailscale(dst, src *TailscaleSpec) TailscaleSpec {
	if dst == nil {
		return *src
	}
	out := *dst
	out.Enabled = src.Enabled
	setIfNonEmpty(&out.Provider, src.Provider)
	if len(src.Tags) > 0 {
		out.Tags = src.Tags
	}
	mergeMaps(&out.ACLMapping, src.ACLMapping)
	return out
}

func mergeSecrets(dst, src *SecretsSpec) {
	setIfNonEmpty(&dst.Provider, src.Provider)

	strategy := coalesce(src.EntriesMergeStrategy, dst.EntriesMergeStrategy, MergeStrategyReplace)
	if len(src.Entries) > 0 {
		dst.Entries = mergeSecretEntries(dst.Entries, src.Entries, strategy)
	}
	if src.EntriesMergeStrategy != "" {
		dst.EntriesMergeStrategy = src.EntriesMergeStrategy
	}
}

func mergeObservability(dst, src *ObservabilitySpec) {
	mergeLogging(&dst.Logging, &src.Logging)
	mergeMetrics(&dst.Metrics, &src.Metrics)
	mergeTraces(&dst.Traces, &src.Traces)
	mergeHealthCheck(&dst.HealthCheck, &src.HealthCheck)
}

func mergeLogging(dst, src *LoggingSpec) {
	setIfNonEmpty(&dst.Level, src.Level)
	setIfNonEmpty(&dst.Format, src.Format)
	if len(src.Destinations) > 0 {
		dst.Destinations = src.Destinations
	}
	if len(src.RedactionRules) > 0 {
		dst.RedactionRules = src.RedactionRules
	}
}

func mergeMetrics(dst, src *MetricsSpec) {
	if src.Enabled {
		dst.Enabled = true
	}
	setIfNonEmpty(&dst.Endpoint, src.Endpoint)
	setIfNonEmpty(&dst.Format, src.Format)
}

func mergeTraces(dst, src *TracesSpec) {
	if src.Enabled {
		dst.Enabled = true
	}
	setIfNonEmpty(&dst.Endpoint, src.Endpoint)
	setIfNonEmpty(&dst.Format, src.Format)
}

func mergeHealthCheck(dst, src *HealthCheckSpec) {
	setIfNonEmpty(&dst.Path, src.Path)
	setIfNonEmpty(&dst.Interval, src.Interval)
	setIfNonEmpty(&dst.Timeout, src.Timeout)
	if src.ReadinessGate {
		dst.ReadinessGate = true
	}
}

func mergePolicy(dst, src *PolicySpec) {
	setIfNonEmpty(&dst.UpdateChannel, src.UpdateChannel)
	setIfNonEmpty(&dst.PinnedVersion, src.PinnedVersion)
	if len(src.ApprovedPlugins) > 0 {
		dst.ApprovedPlugins = src.ApprovedPlugins
	}
	if len(src.ApprovedProviders) > 0 {
		dst.ApprovedProviders = src.ApprovedProviders
	}
	if len(src.FilesystemAllowlist) > 0 {
		dst.FilesystemAllowlist = src.FilesystemAllowlist
	}
	if len(src.CommandAllowlist) > 0 {
		dst.CommandAllowlist = src.CommandAllowlist
	}
}

// --- slice merge helpers ---

func mergePortSlice(dst, src []PortSpec, strategy string) []PortSpec {
	switch strategy {
	case MergeStrategyAppend:
		return append(dst, src...)
	case MergeStrategyUnion:
		seen := make(map[string]struct{}, len(dst))
		for _, p := range dst {
			seen[p.Name] = struct{}{}
		}
		out := make([]PortSpec, len(dst))
		copy(out, dst)
		for _, p := range src {
			if _, ok := seen[p.Name]; !ok {
				out = append(out, p)
			}
		}
		return out
	default: // replace
		return src
	}
}

func mergeSecretEntries(dst, src []SecretEntry, strategy string) []SecretEntry {
	switch strategy {
	case MergeStrategyAppend:
		return append(dst, src...)
	case MergeStrategyUnion:
		seen := make(map[string]struct{}, len(dst))
		for _, e := range dst {
			seen[e.Name] = struct{}{}
		}
		out := make([]SecretEntry, len(dst))
		copy(out, dst)
		for _, e := range src {
			if _, ok := seen[e.Name]; !ok {
				out = append(out, e)
			}
		}
		return out
	default: // replace
		return src
	}
}

// --- primitives ---

func setIfNonEmpty(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

func mergeMaps(dst *map[string]string, src map[string]string) {
	if len(src) == 0 {
		return
	}
	if *dst == nil {
		*dst = make(map[string]string)
	}
	for k, v := range src {
		(*dst)[k] = v
	}
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func deepCopy(t *Template) (*Template, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	var out Template
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
