package template

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Registry manages encoding templates stored on the filesystem.
//
// Expected directory layout:
//
//	<root>/
//	  base/          — base templates (Kind: EncodingTemplate)
//	  roles/         — role overlays  (Kind: RoleOverlay)
//	  environments/  — env overlays   (Kind: EnvironmentOverlay)
type Registry struct {
	root string
}

// NewRegistry creates a registry backed by the given directory.
func NewRegistry(root string) *Registry {
	return &Registry{root: root}
}

// Root returns the registry root path.
func (r *Registry) Root() string { return r.root }

// Init creates the directory skeleton for a new template registry.
func (r *Registry) Init() error {
	dirs := []string{
		filepath.Join(r.root, "base"),
		filepath.Join(r.root, "roles"),
		filepath.Join(r.root, "environments"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", d, err)
		}
	}
	return nil
}

// LoadBase loads a base template by name.
func (r *Registry) LoadBase(name string) (*Template, error) {
	return r.loadTemplate(filepath.Join(r.root, "base", name+".yaml"))
}

// LoadRole loads a role overlay by name.
func (r *Registry) LoadRole(name string) (*Template, error) {
	return r.loadTemplate(filepath.Join(r.root, "roles", name+".yaml"))
}

// LoadEnvironment loads an environment overlay by name.
func (r *Registry) LoadEnvironment(name string) (*Template, error) {
	return r.loadTemplate(filepath.Join(r.root, "environments", name+".yaml"))
}

// LoadInstanceVars loads instance variables from an arbitrary path.
func (r *Registry) LoadInstanceVars(path string) (*Template, error) {
	return r.loadTemplate(path)
}

// SaveBase writes a base template to the registry.
func (r *Registry) SaveBase(tmpl *Template) error {
	return r.saveTemplate(filepath.Join(r.root, "base", tmpl.Metadata.Name+".yaml"), tmpl)
}

// SaveRole writes a role overlay to the registry.
func (r *Registry) SaveRole(tmpl *Template) error {
	return r.saveTemplate(filepath.Join(r.root, "roles", tmpl.Metadata.Name+".yaml"), tmpl)
}

// SaveEnvironment writes an environment overlay to the registry.
func (r *Registry) SaveEnvironment(tmpl *Template) error {
	return r.saveTemplate(filepath.Join(r.root, "environments", tmpl.Metadata.Name+".yaml"), tmpl)
}

// ListBase returns names of all base templates.
func (r *Registry) ListBase() ([]string, error) {
	return r.listDir(filepath.Join(r.root, "base"))
}

// ListRoles returns names of all role overlays.
func (r *Registry) ListRoles() ([]string, error) {
	return r.listDir(filepath.Join(r.root, "roles"))
}

// ListEnvironments returns names of all environment overlays.
func (r *Registry) ListEnvironments() ([]string, error) {
	return r.listDir(filepath.Join(r.root, "environments"))
}

func (r *Registry) loadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &tmpl, nil
}

func (r *Registry) saveTemplate(path string, tmpl *Template) error {
	data, err := yaml.Marshal(tmpl)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (r *Registry) listDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			names = append(names, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
		}
	}
	sort.Strings(names)
	return names, nil
}
