package template

import "fmt"

// Renderer produces an InstallBundle from a resolved template.
type Renderer interface {
	// Render generates deployment artifacts for the given template.
	Render(t *Template) (*InstallBundle, error)
}

// NewRenderer returns the appropriate renderer for the template's deployment target.
func NewRenderer(target string) (Renderer, error) {
	switch target {
	case TargetPodman:
		return &PodmanRenderer{}, nil
	case TargetKubernetes, TargetOpenShift:
		return &KubernetesRenderer{}, nil
	case TargetSandbox:
		return &SandboxRenderer{}, nil
	default:
		return nil, fmt.Errorf("unsupported deployment target: %s", target)
	}
}
