package template

import (
	"strings"
	"testing"
)

func TestPodmanRenderer_Render(t *testing.T) {
	tmpl := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test"},
		Spec: Spec{
			Identity: IdentitySpec{
				InstanceName: "gw-1",
				Role:         RoleGateway,
				Region:       "us-east-1",
			},
			Runtime: RuntimeSpec{
				DeploymentTarget: TargetPodman,
				Resources: ResourceSpec{
					CPU:    "1000m",
					Memory: "512Mi",
				},
				Image: ImageSpec{
					Repository: "ghcr.io/openclaw/openclaw",
					Tag:        "stable",
				},
			},
			Network: NetworkSpec{
				BindAddress: "0.0.0.0",
				Ports: []PortSpec{
					{Name: "https", Port: 443, HostPort: 443},
					{Name: "http", Port: 8080, HostPort: 8080},
				},
			},
			Secrets: SecretsSpec{
				Provider: "builtin",
				Entries: []SecretEntry{
					{Name: "api-key", Type: "api_key", Ref: "builtin://api-key"},
				},
			},
			Observability: ObservabilitySpec{
				Logging:     LoggingSpec{Level: "info", Format: "json"},
				HealthCheck: HealthCheckSpec{Path: "/healthz", Interval: "15s", Timeout: "3s"},
			},
			Policy: PolicySpec{UpdateChannel: "stable"},
		},
	}

	renderer := &PodmanRenderer{}
	bundle, err := renderer.Render(tmpl)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// Check expected files exist.
	expectedFiles := []string{"openclaw.yaml", "podman-compose.yml", "install.sh", "upgrade.sh", "verify.sh"}
	for _, f := range expectedFiles {
		if _, ok := bundle.Files[f]; !ok {
			t.Errorf("missing file: %s", f)
		}
	}

	// Check hash is set.
	if bundle.SourceHash == "" {
		t.Error("source hash should be set")
	}

	// Check openclaw.yaml content.
	oc := string(bundle.Files["openclaw.yaml"])
	if !strings.Contains(oc, "instance_name: gw-1") {
		t.Error("openclaw.yaml should contain instance_name")
	}
	if !strings.Contains(oc, "role: gateway") {
		t.Error("openclaw.yaml should contain role")
	}

	// Check compose content.
	compose := string(bundle.Files["podman-compose.yml"])
	if !strings.Contains(compose, "openclaw-gw-1") {
		t.Error("compose should contain container name")
	}
	if !strings.Contains(compose, "443:443") {
		t.Error("compose should contain port mapping")
	}

	// Check install script.
	install := string(bundle.Files["install.sh"])
	if !strings.Contains(install, "#!/usr/bin/env bash") {
		t.Error("install.sh should have shebang")
	}
	if !strings.Contains(install, "podman") {
		t.Error("install.sh should reference podman")
	}
}

func TestKubernetesRenderer_Render(t *testing.T) {
	tmpl := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test"},
		Spec: Spec{
			Identity: IdentitySpec{
				InstanceName: "cp-1",
				Role:         RoleControlPlane,
				Region:       "eu-west-1",
			},
			Runtime: RuntimeSpec{
				DeploymentTarget: TargetKubernetes,
				Resources: ResourceSpec{
					CPU:    "1000m",
					Memory: "1Gi",
				},
				Image: ImageSpec{
					Repository: "ghcr.io/openclaw/openclaw",
					Tag:        "v1.2.0",
				},
			},
			Network: NetworkSpec{
				Ports: []PortSpec{
					{Name: "api", Port: 8080},
				},
			},
			Observability: ObservabilitySpec{
				HealthCheck: HealthCheckSpec{Path: "/healthz"},
			},
		},
	}

	renderer := &KubernetesRenderer{}
	bundle, err := renderer.Render(tmpl)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	expectedFiles := []string{
		"base/openclaw.yaml",
		"base/deployment.yaml",
		"base/service.yaml",
		"base/configmap.yaml",
		"base/kustomization.yaml",
		"install.sh",
		"verify.sh",
	}
	for _, f := range expectedFiles {
		if _, ok := bundle.Files[f]; !ok {
			t.Errorf("missing file: %s", f)
		}
	}

	// Check deployment content.
	deploy := string(bundle.Files["base/deployment.yaml"])
	if !strings.Contains(deploy, "openclaw-cp-1") {
		t.Error("deployment should contain instance name")
	}
	if !strings.Contains(deploy, "control-plane") {
		t.Error("deployment should contain role label")
	}

	// Check kustomization.
	kust := string(bundle.Files["base/kustomization.yaml"])
	if !strings.Contains(kust, "deployment.yaml") {
		t.Error("kustomization should reference deployment")
	}
}

func TestSandboxRenderer_Render(t *testing.T) {
	tmpl := &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test"},
		Spec: Spec{
			Identity: IdentitySpec{
				InstanceName: "dev-1",
				Role:         RoleWorker,
			},
			Runtime: RuntimeSpec{
				DeploymentTarget: TargetSandbox,
			},
			Network: NetworkSpec{
				Ports: []PortSpec{{Name: "http", Port: 9090}},
			},
			Observability: ObservabilitySpec{
				HealthCheck: HealthCheckSpec{Path: "/healthz"},
			},
		},
	}

	renderer := &SandboxRenderer{}
	bundle, err := renderer.Render(tmpl)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if _, ok := bundle.Files["openclaw.yaml"]; !ok {
		t.Error("missing openclaw.yaml")
	}
	if _, ok := bundle.Files["run.sh"]; !ok {
		t.Error("missing run.sh")
	}

	runSh := string(bundle.Files["run.sh"])
	if !strings.Contains(runSh, "9090") {
		t.Error("run.sh should use port from template")
	}
}

func TestNewRenderer_InvalidTarget(t *testing.T) {
	_, err := NewRenderer("invalid-target")
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestNewRenderer_Podman(t *testing.T) {
	r, err := NewRenderer(TargetPodman)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*PodmanRenderer); !ok {
		t.Errorf("expected *PodmanRenderer, got %T", r)
	}
}

func TestNewRenderer_Kubernetes(t *testing.T) {
	r, err := NewRenderer(TargetKubernetes)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(*KubernetesRenderer); !ok {
		t.Errorf("expected *KubernetesRenderer, got %T", r)
	}
}
