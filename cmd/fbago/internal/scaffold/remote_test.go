package scaffold

import "testing"

func TestParseRemoteGitTemplateHostedSubdirRef(t *testing.T) {
	spec, ok, err := parseRemoteGitTemplate("github.com/acme/fba-go-template/admin@v0.1.0")
	if err != nil {
		t.Fatalf("parseRemoteGitTemplate() error = %v", err)
	}
	if !ok {
		t.Fatal("parseRemoteGitTemplate() ok = false, want true")
	}
	if spec.CloneURL != "https://github.com/acme/fba-go-template.git" {
		t.Fatalf("CloneURL = %q", spec.CloneURL)
	}
	if spec.Subdir != "admin" {
		t.Fatalf("Subdir = %q", spec.Subdir)
	}
	if spec.Ref != "v0.1.0" {
		t.Fatalf("Ref = %q", spec.Ref)
	}
}

func TestParseRemoteGitTemplateURLSubdirRef(t *testing.T) {
	spec, ok, err := parseRemoteGitTemplate("https://github.com/acme/fba-go-template.git//admin@v0.1.0")
	if err != nil {
		t.Fatalf("parseRemoteGitTemplate() error = %v", err)
	}
	if !ok {
		t.Fatal("parseRemoteGitTemplate() ok = false, want true")
	}
	if spec.CloneURL != "https://github.com/acme/fba-go-template.git" {
		t.Fatalf("CloneURL = %q", spec.CloneURL)
	}
	if spec.Subdir != "admin" {
		t.Fatalf("Subdir = %q", spec.Subdir)
	}
	if spec.Ref != "v0.1.0" {
		t.Fatalf("Ref = %q", spec.Ref)
	}
}
