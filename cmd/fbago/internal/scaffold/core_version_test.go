package scaffold

import (
	"runtime/debug"
	"testing"
)

func TestResolveCoreVersionUsesReleaseBuildVersion(t *testing.T) {
	withReadBuildInfo(t, "v1.4.0")

	got, err := resolveCoreVersion("", "")
	if err != nil {
		t.Fatalf("resolveCoreVersion() error = %v", err)
	}
	if got != "v1.4.0" {
		t.Fatalf("resolveCoreVersion() = %q, want release build version", got)
	}
}

func TestResolveCoreVersionKeepsPlaceholderWhenCoreIsReplaced(t *testing.T) {
	withReadBuildInfo(t, "v1.4.0")

	got, err := resolveCoreVersion("", "../fba-go")
	if err != nil {
		t.Fatalf("resolveCoreVersion() error = %v", err)
	}
	if got != developmentCoreVersion {
		t.Fatalf("resolveCoreVersion() = %q, want %q with replace", got, developmentCoreVersion)
	}
}

func TestResolveCoreVersionQueriesLatestWhenRequested(t *testing.T) {
	previous := queryLatestCoreVersion
	queryLatestCoreVersion = func() (string, error) {
		return "v1.5.0", nil
	}
	t.Cleanup(func() {
		queryLatestCoreVersion = previous
	})

	got, err := resolveCoreVersion("latest", "")
	if err != nil {
		t.Fatalf("resolveCoreVersion() error = %v", err)
	}
	if got != "v1.5.0" {
		t.Fatalf("resolveCoreVersion() = %q, want latest version", got)
	}
}

func TestResolveTemplateDependencyUsesEmbeddedVersionByDefault(t *testing.T) {
	t.Setenv("FBAGO_TEMPLATE_REPLACE", "")

	version, replacement := resolveTemplateDependency(templateBundle{
		TemplateModule:  "github.com/yuWorm/fba-go-admin",
		TemplateVersion: "v0.3.0",
		TemplateSource:  "embedded",
	}, "")
	if version != "v0.3.0" {
		t.Fatalf("resolveTemplateDependency() version = %q, want v0.3.0", version)
	}
	if replacement != "" {
		t.Fatalf("resolveTemplateDependency() replacement = %q, want empty replacement", replacement)
	}
}

func withReadBuildInfo(t *testing.T, version string) {
	t.Helper()
	previous := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: version}}, true
	}
	t.Cleanup(func() {
		readBuildInfo = previous
	})
}
