package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/martinmose/agentlink/internal/registry"
)

func TestGenerateConfigNativeToolsIsSuccessfulNoOp(t *testing.T) {
	withDetectTestDir(t)
	setDetectTestFlags(t)

	detected := []registry.Detected{{Tool: registry.Tool{Name: "Native", RepoFileName: "AGENTS.md", ReadsAgentsMD: true}}}
	stdout, _ := captureOutput(t, func() {
		if err := generateConfig(detected); err != nil {
			t.Fatalf("generateConfig() error = %v, want nil", err)
		}
	})
	if !strings.Contains(stdout, "Detected tools read AGENTS.md directly; no config is needed") {
		t.Fatalf("generateConfig() output = %q, want native-tool no-op message", stdout)
	}
	if _, statErr := os.Stat(".agentlink.yaml"); !os.IsNotExist(statErr) {
		t.Fatalf(".agentlink.yaml exists after no-op generation; stat error = %v", statErr)
	}
}

func TestGenerateConfigNoDetectedToolsIsSuccessfulNoOp(t *testing.T) {
	withDetectTestDir(t)
	setDetectTestFlags(t)

	stdout, _ := captureOutput(t, func() {
		if err := generateConfig(nil); err != nil {
			t.Fatalf("generateConfig(nil) error = %v, want nil", err)
		}
	})
	if !strings.Contains(stdout, "No supported tools detected; no config generated") {
		t.Fatalf("generateConfig(nil) output = %q, want no-tools no-op message", stdout)
	}
	if _, statErr := os.Stat(".agentlink.yaml"); !os.IsNotExist(statErr) {
		t.Fatalf(".agentlink.yaml exists after no-op generation; stat error = %v", statErr)
	}
}

func TestGenerateConfigEmptyLinkSetDoesNotOverwriteExistingConfigWithForce(t *testing.T) {
	withDetectTestDir(t)
	setDetectTestFlags(t)
	force = true
	const existing = "source: README.md\nlinks:\n  - AGENTS.md\n"
	if err := os.WriteFile(".agentlink.yaml", []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateConfig(nil); err != nil {
		t.Fatalf("generateConfig(nil) error = %v, want nil", err)
	}
	got, err := os.ReadFile(".agentlink.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != existing {
		t.Fatalf("existing config was modified: got %q, want %q", got, existing)
	}
}

func TestGenerateConfigProducesLoadableConfig(t *testing.T) {
	dir := withDetectTestDir(t)
	setDetectTestFlags(t)

	detected := []registry.Detected{
		{Tool: registry.Tool{Name: "Native", RepoFileName: "AGENTS.md", ReadsAgentsMD: true}},
		{Tool: registry.Tool{Name: "Claude", RepoFileName: "CLAUDE.md"}},
		{Tool: registry.Tool{Name: "Claude duplicate", RepoFileName: "CLAUDE.md"}},
	}
	if err := generateConfig(detected); err != nil {
		t.Fatalf("generateConfig() error = %v", err)
	}

	cfg, err := config.LoadConfig(filepath.Join(dir, ".agentlink.yaml"))
	if err != nil {
		t.Fatalf("generated config is not loadable: %v", err)
	}
	if len(cfg.Links) != 1 || filepath.Base(cfg.Links[0]) != "CLAUDE.md" {
		t.Fatalf("generated links = %v, want one CLAUDE.md link", cfg.Links)
	}
}

func withDetectTestDir(t *testing.T) string {
	t.Helper()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })
	return dir
}

func setDetectTestFlags(t *testing.T) {
	t.Helper()
	oldDryRun, oldForce := dryRun, force
	dryRun, force = false, false
	t.Cleanup(func() { dryRun, force = oldDryRun, oldForce })
}
