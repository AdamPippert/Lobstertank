package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// WriteBundle writes all files from a rendered InstallBundle to the given directory.
func WriteBundle(bundle *InstallBundle, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Write each artifact.
	for relPath, content := range bundle.Files {
		absPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("create directory for %s: %w", relPath, err)
		}
		perm := os.FileMode(0o644)
		if isScript(relPath) {
			perm = 0o755
		}
		if err := os.WriteFile(absPath, content, perm); err != nil {
			return fmt.Errorf("write %s: %w", relPath, err)
		}
	}

	// Write a manifest file listing all artifacts + source hash.
	manifest := BundleManifest{
		SourceHash: bundle.SourceHash,
		Instance:   bundle.ResolvedTemplate.Spec.Identity.InstanceName,
		Role:       bundle.ResolvedTemplate.Spec.Identity.Role,
		Target:     bundle.ResolvedTemplate.Spec.Runtime.DeploymentTarget,
	}
	for relPath := range bundle.Files {
		manifest.Files = append(manifest.Files, relPath)
	}
	sort.Strings(manifest.Files)

	mdata, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	mpath := filepath.Join(dir, "bundle-manifest.json")
	if err := os.WriteFile(mpath, mdata, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}

// BundleManifest is written alongside the bundle for provenance tracking.
type BundleManifest struct {
	SourceHash string   `json:"source_hash"`
	Instance   string   `json:"instance"`
	Role       string   `json:"role"`
	Target     string   `json:"target"`
	Files      []string `json:"files"`
}

// PlanSummary describes what a render would produce without writing files.
type PlanSummary struct {
	Instance   string   `json:"instance"`
	Role       string   `json:"role"`
	Target     string   `json:"target"`
	SourceHash string   `json:"source_hash"`
	Files      []string `json:"files"`
}

// Plan generates a summary of what rendering would produce.
func Plan(bundle *InstallBundle) *PlanSummary {
	var files []string
	for relPath := range bundle.Files {
		files = append(files, relPath)
	}
	sort.Strings(files)

	return &PlanSummary{
		Instance:   bundle.ResolvedTemplate.Spec.Identity.InstanceName,
		Role:       bundle.ResolvedTemplate.Spec.Identity.Role,
		Target:     bundle.ResolvedTemplate.Spec.Runtime.DeploymentTarget,
		SourceHash: bundle.SourceHash,
		Files:      files,
	}
}

func isScript(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".sh"
}
