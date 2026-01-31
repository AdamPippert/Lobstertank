// Package cli implements the Lobstertank command-line interface.
//
// Subcommands:
//
//	serve              — start the HTTP API server (default)
//	template init      — create a new template registry
//	template list      — list available templates, roles, environments
//	template show      — show a resolved template
//	template add-role  — scaffold a new role overlay
//	template add-env   — scaffold a new environment overlay
//	render             — render a template into an install bundle
//	plan               — preview what render would produce
//	apply              — execute an install bundle
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	tmpl "github.com/AdamPippert/Lobstertank/internal/template"
	"gopkg.in/yaml.v3"
)

const defaultRegistryDir = "templates"

// Run is the top-level CLI dispatcher. It returns an exit code.
func Run(args []string) int {
	if len(args) < 2 {
		return runServe()
	}

	switch args[1] {
	case "serve":
		return runServe()
	case "template":
		return runTemplate(args[2:])
	case "render":
		return runRender(args[2:])
	case "plan":
		return runPlan(args[2:])
	case "apply":
		return runApply(args[2:])
	case "help", "--help", "-h":
		printUsage()
		return 0
	case "version", "--version":
		fmt.Println("lobstertank v0.2.0")
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[1])
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: lobstertank <command> [options]

Commands:
  serve                Start the HTTP API server (default)
  template init        Create a new template registry
  template list        List available templates, roles, environments
  template show        Show a resolved template (base + overlays)
  template add-role    Scaffold a new role overlay
  template add-env     Scaffold a new environment overlay
  render               Render a template into an install bundle
  plan                 Preview what render would produce
  apply                Execute an install bundle

Run "lobstertank <command> --help" for details.
`)
}

// runServe delegates to the existing server bootstrap in main.go.
// This is a sentinel — the actual logic stays in main() to avoid
// circular imports. The caller checks for this.
func runServe() int {
	// Return -1 as sentinel: "fall through to existing server boot"
	return -1
}

// --- template subcommand ---

func runTemplate(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: lobstertank template <init|list|show|add-role|add-env>")
		return 1
	}
	switch args[0] {
	case "init":
		return cmdTemplateInit(args[1:])
	case "list":
		return cmdTemplateList(args[1:])
	case "show":
		return cmdTemplateShow(args[1:])
	case "add-role":
		return cmdTemplateAddRole(args[1:])
	case "add-env":
		return cmdTemplateAddEnv(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown template subcommand: %s\n", args[0])
		return 1
	}
}

func cmdTemplateInit(args []string) int {
	fs := flag.NewFlagSet("template init", flag.ExitOnError)
	base := fs.String("base", "default", "name of the initial base template")
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	_ = fs.Parse(args)

	reg := tmpl.NewRegistry(*dir)
	if err := reg.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	// Write a starter base template.
	starter := &tmpl.Template{
		APIVersion: tmpl.APIVersion,
		Kind:       tmpl.KindEncodingTemplate,
		Metadata: tmpl.Metadata{
			Name:        *base,
			Version:     "0.1.0",
			Description: "Base encoding template — edit to match your standards",
		},
		Spec: tmpl.Spec{
			Identity: tmpl.IdentitySpec{
				Role:   tmpl.RoleWorker,
				Region: "us-east-1",
			},
			Runtime: tmpl.RuntimeSpec{
				DeploymentTarget: tmpl.TargetPodman,
				Resources: tmpl.ResourceSpec{
					CPU:    "500m",
					Memory: "256Mi",
				},
			},
			Network: tmpl.NetworkSpec{
				BindAddress: "0.0.0.0",
				Ports: []tmpl.PortSpec{
					{Name: "http", Port: 8080, Protocol: "tcp"},
				},
			},
			Secrets: tmpl.SecretsSpec{
				Provider: "builtin",
			},
			Observability: tmpl.ObservabilitySpec{
				Logging: tmpl.LoggingSpec{
					Level:        "info",
					Format:       "json",
					Destinations: []string{"stdout"},
				},
				HealthCheck: tmpl.HealthCheckSpec{
					Path:     "/healthz",
					Interval: "30s",
					Timeout:  "5s",
				},
			},
			Policy: tmpl.PolicySpec{
				UpdateChannel: "stable",
			},
		},
	}

	if err := reg.SaveBase(starter); err != nil {
		fmt.Fprintf(os.Stderr, "error writing base template: %v\n", err)
		return 1
	}

	fmt.Printf("Template registry initialized at %s/\n", *dir)
	fmt.Printf("Base template created: %s/base/%s.yaml\n", *dir, *base)
	return 0
}

func cmdTemplateList(args []string) int {
	fs := flag.NewFlagSet("template list", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	_ = fs.Parse(args)

	reg := tmpl.NewRegistry(*dir)

	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tNAME")
	fmt.Fprintln(tw, "----\t----")

	bases, _ := reg.ListBase()
	for _, b := range bases {
		fmt.Fprintf(tw, "base\t%s\n", b)
	}

	roles, _ := reg.ListRoles()
	for _, r := range roles {
		fmt.Fprintf(tw, "role\t%s\n", r)
	}

	envs, _ := reg.ListEnvironments()
	for _, e := range envs {
		fmt.Fprintf(tw, "environment\t%s\n", e)
	}

	tw.Flush()
	return 0
}

func cmdTemplateShow(args []string) int {
	fs := flag.NewFlagSet("template show", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	base := fs.String("base", "", "base template name (required)")
	role := fs.String("role", "", "role overlay name")
	env := fs.String("env", "", "environment overlay name")
	vars := fs.String("vars", "", "instance variables file path")
	format := fs.String("format", "yaml", "output format (yaml|json)")
	_ = fs.Parse(args)

	if *base == "" {
		fmt.Fprintln(os.Stderr, "error: --base is required")
		return 1
	}

	resolved, exitCode := resolveFromFlags(*dir, *base, *role, *env, *vars)
	if exitCode != 0 {
		return exitCode
	}

	switch *format {
	case "json":
		data, _ := json.MarshalIndent(resolved, "", "  ")
		fmt.Println(string(data))
	default:
		data, _ := yaml.Marshal(resolved)
		fmt.Print(string(data))
	}

	return 0
}

func cmdTemplateAddRole(args []string) int {
	fs := flag.NewFlagSet("template add-role", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	name := fs.String("name", "", "role name (required)")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		return 1
	}

	reg := tmpl.NewRegistry(*dir)
	role := &tmpl.Template{
		APIVersion: tmpl.APIVersion,
		Kind:       tmpl.KindRoleOverlay,
		Metadata: tmpl.Metadata{
			Name:        *name,
			Description: fmt.Sprintf("Role overlay for %s — edit to customize", *name),
		},
		Spec: tmpl.Spec{
			Identity: tmpl.IdentitySpec{
				Role: *name,
			},
		},
	}
	if err := reg.SaveRole(role); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("Role overlay created: %s/roles/%s.yaml\n", *dir, *name)
	return 0
}

func cmdTemplateAddEnv(args []string) int {
	fs := flag.NewFlagSet("template add-env", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	name := fs.String("name", "", "environment name (required)")
	target := fs.String("target", "", "deployment target (podman|kubernetes|openshift|droplet|sandbox)")
	_ = fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required")
		return 1
	}

	reg := tmpl.NewRegistry(*dir)
	envOverlay := &tmpl.Template{
		APIVersion: tmpl.APIVersion,
		Kind:       tmpl.KindEnvironmentOverlay,
		Metadata: tmpl.Metadata{
			Name:        *name,
			Description: fmt.Sprintf("Environment overlay for %s — edit to customize", *name),
		},
		Spec: tmpl.Spec{
			Runtime: tmpl.RuntimeSpec{
				DeploymentTarget: *target,
			},
		},
	}
	if err := reg.SaveEnvironment(envOverlay); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("Environment overlay created: %s/environments/%s.yaml\n", *dir, *name)
	return 0
}

// --- render ---

func runRender(args []string) int {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	base := fs.String("template", "", "base template name (required)")
	role := fs.String("role", "", "role overlay name")
	env := fs.String("env", "", "environment overlay name")
	vars := fs.String("vars", "", "instance variables file path")
	out := fs.String("out", "bundle", "output directory")
	_ = fs.Parse(args)

	if *base == "" {
		fmt.Fprintln(os.Stderr, "error: --template is required")
		return 1
	}

	resolved, exitCode := resolveFromFlags(*dir, *base, *role, *env, *vars)
	if exitCode != 0 {
		return exitCode
	}

	if err := tmpl.Validate(resolved); err != nil {
		fmt.Fprintf(os.Stderr, "validation error:\n%v\n", err)
		return 1
	}

	target := resolved.Spec.Runtime.DeploymentTarget
	if target == "" {
		target = tmpl.TargetPodman
	}

	renderer, err := tmpl.NewRenderer(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	bundle, err := renderer.Render(resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		return 1
	}

	outDir := filepath.Join(*out, resolved.Spec.Identity.InstanceName)
	if resolved.Spec.Identity.InstanceName == "" {
		outDir = *out
	}

	if err := tmpl.WriteBundle(bundle, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		return 1
	}

	fmt.Printf("Bundle rendered to %s/\n", outDir)
	fmt.Printf("  Instance:    %s\n", resolved.Spec.Identity.InstanceName)
	fmt.Printf("  Role:        %s\n", resolved.Spec.Identity.Role)
	fmt.Printf("  Target:      %s\n", target)
	fmt.Printf("  Source hash: %s\n", bundle.SourceHash)
	fmt.Printf("  Files:       %d\n", len(bundle.Files))
	return 0
}

// --- plan ---

func runPlan(args []string) int {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	dir := fs.String("dir", defaultRegistryDir, "template registry directory")
	base := fs.String("template", "", "base template name (required)")
	role := fs.String("role", "", "role overlay name")
	env := fs.String("env", "", "environment overlay name")
	vars := fs.String("vars", "", "instance variables file path")
	_ = fs.Parse(args)

	if *base == "" {
		fmt.Fprintln(os.Stderr, "error: --template is required")
		return 1
	}

	resolved, exitCode := resolveFromFlags(*dir, *base, *role, *env, *vars)
	if exitCode != 0 {
		return exitCode
	}

	if err := tmpl.Validate(resolved); err != nil {
		fmt.Fprintf(os.Stderr, "validation error:\n%v\n", err)
		return 1
	}

	target := resolved.Spec.Runtime.DeploymentTarget
	if target == "" {
		target = tmpl.TargetPodman
	}

	renderer, err := tmpl.NewRenderer(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	bundle, err := renderer.Render(resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render error: %v\n", err)
		return 1
	}

	plan := tmpl.Plan(bundle)

	fmt.Println("Plan — the following bundle would be generated:")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Instance:    %s\n", plan.Instance)
	fmt.Printf("  Role:        %s\n", plan.Role)
	fmt.Printf("  Target:      %s\n", plan.Target)
	fmt.Printf("  Source hash: %s\n", plan.SourceHash)
	fmt.Println()
	fmt.Println("  Files:")
	for _, f := range plan.Files {
		fmt.Printf("    + %s\n", f)
	}
	fmt.Println()
	fmt.Println("Run with 'lobstertank render' to write these files.")

	return 0
}

// --- apply ---

func runApply(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	bundleDir := fs.String("bundle", "bundle", "path to a rendered bundle directory")
	dryRun := fs.Bool("dry-run", false, "print commands without executing")
	_ = fs.Parse(args)

	// If there's a positional arg, use that as bundle dir.
	if fs.NArg() > 0 {
		*bundleDir = fs.Arg(0)
	}

	installScript := filepath.Join(*bundleDir, "install.sh")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: no install.sh found in %s\n", *bundleDir)
		fmt.Fprintln(os.Stderr, "Run 'lobstertank render' first to generate a bundle.")
		return 1
	}

	if *dryRun {
		fmt.Printf("Would execute: bash %s\n", installScript)
		return 0
	}

	fmt.Printf("==> Applying bundle from %s\n", *bundleDir)
	fmt.Printf("==> Running: bash %s\n", installScript)
	fmt.Println()

	// Use os/exec to run the install script.
	// Importing os/exec here to keep the package lightweight.
	cmd := execCommand("bash", installScript)
	cmd.Dir = *bundleDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: install script failed: %v\n", err)
		return 1
	}

	fmt.Println("\n==> Apply complete.")
	return 0
}

// --- helpers ---

func resolveFromFlags(dir, baseName, roleName, envName, varsPath string) (*tmpl.Template, int) {
	reg := tmpl.NewRegistry(dir)

	base, err := reg.LoadBase(baseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading base template %q: %v\n", baseName, err)
		return nil, 1
	}

	var layers []*tmpl.Template

	if roleName != "" {
		role, err := reg.LoadRole(roleName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading role overlay %q: %v\n", roleName, err)
			return nil, 1
		}
		layers = append(layers, role)
	}

	if envName != "" {
		envOverlay, err := reg.LoadEnvironment(envName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading environment overlay %q: %v\n", envName, err)
			return nil, 1
		}
		layers = append(layers, envOverlay)
	}

	if varsPath != "" {
		iv, err := reg.LoadInstanceVars(varsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading instance variables from %q: %v\n", varsPath, err)
			return nil, 1
		}
		layers = append(layers, iv)
	}

	resolved, err := tmpl.Resolve(base, layers...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving template: %v\n", err)
		return nil, 1
	}

	return resolved, 0
}
