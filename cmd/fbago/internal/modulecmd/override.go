package modulecmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UseOptions identifies a versioned module and the local checkout that should replace it.
type UseOptions struct {
	ProjectDir string
	Module     string
	Path       string
}

// Use records a Go module replace directive for an existing local checkout.
// The checkout may be a Git submodule, a fork, or an ordinary development clone.
func Use(opts UseOptions) error {
	projectDir, modulePath, checkout, err := validateUseOptions(opts)
	if err != nil {
		return err
	}
	actualModule, err := localModulePath(checkout)
	if err != nil {
		return err
	}
	if actualModule != modulePath {
		return fmt.Errorf("local checkout %s declares module %s, want %s", checkout, actualModule, modulePath)
	}
	replacement, err := filepath.Rel(projectDir, checkout)
	if err != nil {
		return err
	}
	replacement = filepath.ToSlash(replacement)
	if replacement != "." && !strings.HasPrefix(replacement, ".") {
		replacement = "./" + replacement
	}
	return runGoModEdit(projectDir, "-replace="+modulePath+"="+replacement)
}

// Reset removes a local replacement so the selected version is resolved through Go modules again.
func Reset(projectDir string, modulePath string) error {
	projectDir, err := resolveProjectDir(projectDir)
	if err != nil {
		return err
	}
	modulePath = strings.TrimSpace(modulePath)
	if modulePath == "" || strings.ContainsAny(modulePath, " \t\r\n") {
		return fmt.Errorf("module path is required")
	}
	return runGoModEdit(projectDir, "-dropreplace="+modulePath)
}

func validateUseOptions(opts UseOptions) (projectDir string, modulePath string, checkout string, err error) {
	projectDir, err = resolveProjectDir(opts.ProjectDir)
	if err != nil {
		return "", "", "", err
	}
	modulePath = strings.TrimSpace(opts.Module)
	if modulePath == "" || strings.ContainsAny(modulePath, " \t\r\n") {
		return "", "", "", fmt.Errorf("module path is required")
	}
	if strings.TrimSpace(opts.Path) == "" {
		return "", "", "", fmt.Errorf("local module path is required")
	}
	checkout = opts.Path
	if !filepath.IsAbs(checkout) {
		checkout = filepath.Join(projectDir, checkout)
	}
	checkout, err = filepath.Abs(checkout)
	if err != nil {
		return "", "", "", err
	}
	info, err := os.Stat(checkout)
	if err != nil {
		return "", "", "", err
	}
	if !info.IsDir() {
		return "", "", "", fmt.Errorf("local module path %s is not a directory", checkout)
	}
	return projectDir, modulePath, checkout, nil
}

func resolveProjectDir(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		value = "."
	}
	projectDir, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("project %s has no go.mod", projectDir)
		}
		return "", err
	}
	return projectDir, nil
}

func localModulePath(checkout string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Path}}")
	cmd.Dir = checkout
	cmd.Env = withGoWorkOff(os.Environ())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("inspect local module %s: %w: %s", checkout, err, detail)
		}
		return "", fmt.Errorf("inspect local module %s: %w", checkout, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func runGoModEdit(projectDir string, argument string) error {
	cmd := exec.Command("go", "mod", "edit", argument)
	cmd.Dir = projectDir
	cmd.Env = withGoWorkOff(os.Environ())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return fmt.Errorf("go mod edit: %w: %s", err, detail)
		}
		return fmt.Errorf("go mod edit: %w", err)
	}
	return nil
}

func withGoWorkOff(environment []string) []string {
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if strings.HasPrefix(item, "GOWORK=") {
			continue
		}
		result = append(result, item)
	}
	return append(result, "GOWORK=off")
}
