package template

import (
	"testing"
)

func TestResolve_NilBase(t *testing.T) {
	_, err := Resolve(nil)
	if err == nil {
		t.Fatal("expected error for nil base template")
	}
}

func TestResolve_BaseOnly(t *testing.T) {
	base := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test-base", Version: "1.0.0"},
		Spec: Spec{
			Identity: IdentitySpec{
				Role:   RoleWorker,
				Region: "us-east-1",
			},
			Runtime: RuntimeSpec{
				DeploymentTarget: TargetPodman,
				Resources: ResourceSpec{
					CPU:    "500m",
					Memory: "256Mi",
				},
			},
			Network: NetworkSpec{
				BindAddress: "0.0.0.0",
				Ports: []PortSpec{
					{Name: "http", Port: 8080, Protocol: "tcp"},
				},
			},
		},
	}

	result, err := Resolve(base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Spec.Identity.Role != RoleWorker {
		t.Errorf("expected role %q, got %q", RoleWorker, result.Spec.Identity.Role)
	}
	if result.Spec.Runtime.DeploymentTarget != TargetPodman {
		t.Errorf("expected target %q, got %q", TargetPodman, result.Spec.Runtime.DeploymentTarget)
	}
	if len(result.Spec.Network.Ports) != 1 || result.Spec.Network.Ports[0].Port != 8080 {
		t.Errorf("expected port 8080, got %+v", result.Spec.Network.Ports)
	}
}

func TestResolve_BaseWithRoleOverlay(t *testing.T) {
	base := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "base"},
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleWorker, Region: "us-east-1"},
			Runtime:  RuntimeSpec{DeploymentTarget: TargetPodman, Resources: ResourceSpec{CPU: "500m", Memory: "256Mi"}},
			Network: NetworkSpec{
				Ports: []PortSpec{{Name: "http", Port: 8080}},
			},
			Observability: ObservabilitySpec{
				Logging: LoggingSpec{Level: "info", Format: "json"},
			},
		},
	}

	role := &Template{
		APIVersion: APIVersion,
		Kind:       KindRoleOverlay,
		Metadata:   Metadata{Name: "gateway"},
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleGateway},
			Runtime:  RuntimeSpec{Resources: ResourceSpec{CPU: "1000m", Memory: "512Mi"}},
			Network: NetworkSpec{
				Ports: []PortSpec{
					{Name: "https", Port: 443},
					{Name: "http", Port: 8080},
				},
			},
		},
	}

	result, err := Resolve(base, role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Spec.Identity.Role != RoleGateway {
		t.Errorf("role should be overridden: expected %q, got %q", RoleGateway, result.Spec.Identity.Role)
	}
	if result.Spec.Identity.Region != "us-east-1" {
		t.Errorf("region should be preserved: expected %q, got %q", "us-east-1", result.Spec.Identity.Region)
	}
	if result.Spec.Runtime.Resources.CPU != "1000m" {
		t.Errorf("CPU should be overridden: expected %q, got %q", "1000m", result.Spec.Runtime.Resources.CPU)
	}
	// Ports are replaced (default strategy).
	if len(result.Spec.Network.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(result.Spec.Network.Ports))
	}
	if result.Spec.Network.Ports[0].Name != "https" {
		t.Errorf("first port should be https, got %q", result.Spec.Network.Ports[0].Name)
	}
	// Logging should be preserved from base.
	if result.Spec.Observability.Logging.Level != "info" {
		t.Errorf("logging level should be preserved, got %q", result.Spec.Observability.Logging.Level)
	}
}

func TestResolve_ThreeLayerStack(t *testing.T) {
	base := &Template{
		Metadata: Metadata{Name: "base"},
		Spec: Spec{
			Identity:      IdentitySpec{Role: RoleWorker, Region: "us-east-1"},
			Runtime:       RuntimeSpec{DeploymentTarget: TargetPodman},
			Observability: ObservabilitySpec{Logging: LoggingSpec{Level: "info"}},
		},
	}
	role := &Template{
		Metadata: Metadata{Name: "gateway"},
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleGateway},
			Runtime:  RuntimeSpec{Resources: ResourceSpec{CPU: "1000m"}},
		},
	}
	env := &Template{
		Metadata: Metadata{Name: "k8s"},
		Spec: Spec{
			Runtime: RuntimeSpec{DeploymentTarget: TargetKubernetes},
			Secrets: SecretsSpec{Provider: "k8s_secrets"},
		},
	}

	result, err := Resolve(base, role, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Spec.Identity.Role != RoleGateway {
		t.Errorf("expected role %q, got %q", RoleGateway, result.Spec.Identity.Role)
	}
	if result.Spec.Runtime.DeploymentTarget != TargetKubernetes {
		t.Errorf("expected target %q, got %q", TargetKubernetes, result.Spec.Runtime.DeploymentTarget)
	}
	if result.Spec.Secrets.Provider != "k8s_secrets" {
		t.Errorf("expected secrets provider %q, got %q", "k8s_secrets", result.Spec.Secrets.Provider)
	}
	if result.Spec.Runtime.Resources.CPU != "1000m" {
		t.Errorf("CPU from role should survive env overlay: expected %q, got %q", "1000m", result.Spec.Runtime.Resources.CPU)
	}
}

func TestResolve_PortsMergeStrategy_Append(t *testing.T) {
	base := &Template{
		Metadata: Metadata{Name: "base"},
		Spec: Spec{
			Network: NetworkSpec{
				Ports: []PortSpec{{Name: "http", Port: 8080}},
			},
		},
	}
	overlay := &Template{
		Spec: Spec{
			Network: NetworkSpec{
				Ports:              []PortSpec{{Name: "https", Port: 443}},
				PortsMergeStrategy: MergeStrategyAppend,
			},
		},
	}

	result, err := Resolve(base, overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Spec.Network.Ports) != 2 {
		t.Fatalf("expected 2 ports with append, got %d", len(result.Spec.Network.Ports))
	}
}

func TestResolve_PortsMergeStrategy_Union(t *testing.T) {
	base := &Template{
		Metadata: Metadata{Name: "base"},
		Spec: Spec{
			Network: NetworkSpec{
				Ports: []PortSpec{
					{Name: "http", Port: 8080},
					{Name: "metrics", Port: 9090},
				},
			},
		},
	}
	overlay := &Template{
		Spec: Spec{
			Network: NetworkSpec{
				Ports: []PortSpec{
					{Name: "http", Port: 80},   // exists → skip
					{Name: "https", Port: 443}, // new → add
				},
				PortsMergeStrategy: MergeStrategyUnion,
			},
		},
	}

	result, err := Resolve(base, overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Spec.Network.Ports) != 3 {
		t.Fatalf("expected 3 ports with union, got %d: %+v", len(result.Spec.Network.Ports), result.Spec.Network.Ports)
	}
	// http stays as 8080 (from base), not 80 (union skips duplicates by name).
	if result.Spec.Network.Ports[0].Port != 8080 {
		t.Errorf("http port should remain 8080, got %d", result.Spec.Network.Ports[0].Port)
	}
}

func TestResolve_SecretsMergeStrategy_Union(t *testing.T) {
	base := &Template{
		Metadata: Metadata{Name: "base"},
		Spec: Spec{
			Secrets: SecretsSpec{
				Provider:             "builtin",
				Entries:              []SecretEntry{{Name: "api-key", Type: "api_key", Ref: "builtin://api-key"}},
				EntriesMergeStrategy: MergeStrategyUnion,
			},
		},
	}
	overlay := &Template{
		Spec: Spec{
			Secrets: SecretsSpec{
				Entries: []SecretEntry{
					{Name: "api-key", Type: "api_key", Ref: "builtin://new"},
					{Name: "tls-cert", Type: "cert", Ref: "builtin://tls"},
				},
				EntriesMergeStrategy: MergeStrategyUnion,
			},
		},
	}

	result, err := Resolve(base, overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Spec.Secrets.Entries) != 2 {
		t.Fatalf("expected 2 secret entries with union, got %d", len(result.Spec.Secrets.Entries))
	}
}

func TestResolve_LabelsMerge(t *testing.T) {
	base := &Template{
		Metadata: Metadata{
			Name:   "base",
			Labels: map[string]string{"owner": "adam", "env": "prod"},
		},
		Spec: Spec{
			Identity: IdentitySpec{Labels: map[string]string{"tier": "production"}},
		},
	}
	overlay := &Template{
		Metadata: Metadata{
			Labels: map[string]string{"env": "staging", "region": "eu"},
		},
		Spec: Spec{
			Identity: IdentitySpec{Labels: map[string]string{"tier": "edge", "role": "gateway"}},
		},
	}

	result, err := Resolve(base, overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Metadata labels: owner preserved, env overridden, region added.
	if result.Metadata.Labels["owner"] != "adam" {
		t.Errorf("owner should be preserved")
	}
	if result.Metadata.Labels["env"] != "staging" {
		t.Errorf("env should be overridden to staging")
	}
	if result.Metadata.Labels["region"] != "eu" {
		t.Errorf("region should be added")
	}
	// Identity labels: tier overridden, role added.
	if result.Spec.Identity.Labels["tier"] != "edge" {
		t.Errorf("tier should be overridden to edge")
	}
	if result.Spec.Identity.Labels["role"] != "gateway" {
		t.Errorf("role should be added")
	}
}

func TestResolve_DoesNotMutateBase(t *testing.T) {
	base := &Template{
		Metadata: Metadata{Name: "base"},
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleWorker},
		},
	}
	overlay := &Template{
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleGateway},
		},
	}

	_, err := Resolve(base, overlay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base.Spec.Identity.Role != RoleWorker {
		t.Error("base template was mutated")
	}
}

func TestHash_Deterministic(t *testing.T) {
	tmpl := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test"},
		Spec: Spec{
			Identity: IdentitySpec{Role: RoleWorker},
		},
	}

	h1, err := Hash(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := Hash(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("hash not deterministic: %s != %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("expected SHA-256 hex digest (64 chars), got %d chars", len(h1))
	}
}
