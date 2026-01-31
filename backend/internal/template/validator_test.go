package template

import (
	"strings"
	"testing"
)

func TestValidate_ValidTemplate(t *testing.T) {
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
			},
			Network: NetworkSpec{
				BindAddress: "0.0.0.0",
				Ports: []PortSpec{
					{Name: "https", Port: 443, Protocol: "tcp"},
					{Name: "http", Port: 8080, Protocol: "tcp"},
				},
			},
			Secrets: SecretsSpec{
				Provider: "builtin",
				Entries: []SecretEntry{
					{Name: "api-key", Type: "api_key", Ref: "builtin://api-key"},
				},
			},
			Observability: ObservabilitySpec{
				Logging: LoggingSpec{Level: "info", Format: "json"},
			},
			Policy: PolicySpec{
				UpdateChannel: "stable",
			},
		},
	}
	if err := Validate(tmpl); err != nil {
		t.Fatalf("expected valid template, got error: %v", err)
	}
}

func TestValidate_MissingRequiredFields(t *testing.T) {
	tmpl := &Template{}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected validation error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 errors (apiVersion, kind, name), got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidate_InvalidRole(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Identity.Role = "not-a-role"
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if !strings.Contains(err.Error(), "not-a-role") {
		t.Errorf("error should mention invalid role: %v", err)
	}
}

func TestValidate_InvalidTarget(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Runtime.DeploymentTarget = "baremetal"
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
	if !strings.Contains(err.Error(), "baremetal") {
		t.Errorf("error should mention invalid target: %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Network.Ports = []PortSpec{
		{Name: "http", Port: 0}, // invalid
	}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestValidate_DuplicatePortNames(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Network.Ports = []PortSpec{
		{Name: "http", Port: 8080},
		{Name: "http", Port: 8081}, // duplicate name
	}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for duplicate port name")
	}
	if !strings.Contains(err.Error(), "duplicates") {
		t.Errorf("error should mention duplicate: %v", err)
	}
}

func TestValidate_DuplicatePortNumbers(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Network.Ports = []PortSpec{
		{Name: "http", Port: 8080},
		{Name: "metrics", Port: 8080}, // duplicate port number
	}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for duplicate port number")
	}
	if !strings.Contains(err.Error(), "conflicts") {
		t.Errorf("error should mention conflict: %v", err)
	}
}

func TestValidate_InvalidBindAddress(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Network.BindAddress = "not-an-ip"
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for invalid bind address")
	}
}

func TestValidate_InvalidSecretType(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Secrets.Entries = []SecretEntry{
		{Name: "foo", Type: "unknown_type", Ref: "builtin://foo"},
	}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for unknown secret type")
	}
}

func TestValidate_SecretMissingRef(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Secrets.Entries = []SecretEntry{
		{Name: "foo", Type: "api_key", Ref: ""}, // missing ref
	}
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for missing secret ref")
	}
	if !strings.Contains(err.Error(), "ref is required") {
		t.Errorf("error should mention ref: %v", err)
	}
}

func TestValidate_GatewayNoPorts(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Identity.Role = RoleGateway
	tmpl.Spec.Network.Ports = nil // gateway requires ports
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error: gateway requires ports")
	}
	if !strings.Contains(err.Error(), "gateway role requires") {
		t.Errorf("error should mention gateway role: %v", err)
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Observability.Logging.Level = "trace"
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestValidate_InvalidUpdateChannel(t *testing.T) {
	tmpl := validBase()
	tmpl.Spec.Policy.UpdateChannel = "canary"
	err := Validate(tmpl)
	if err == nil {
		t.Fatal("expected error for invalid update channel")
	}
}

func validBase() *Template {
	return &Template{
		APIVersion: APIVersion,
		Kind:       KindEncodingTemplate,
		Metadata:   Metadata{Name: "test"},
		Spec: Spec{
			Identity: IdentitySpec{
				InstanceName: "test-1",
				Role:         RoleWorker,
			},
			Runtime: RuntimeSpec{
				DeploymentTarget: TargetPodman,
			},
			Network: NetworkSpec{
				Ports: []PortSpec{{Name: "http", Port: 8080}},
			},
			Secrets: SecretsSpec{Provider: "builtin"},
			Observability: ObservabilitySpec{
				Logging: LoggingSpec{Level: "info", Format: "json"},
			},
			Policy: PolicySpec{UpdateChannel: "stable"},
		},
	}
}
