package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellQuoteWrapsPathsSafely(t *testing.T) {
	got := shellQuote("/tmp/Agent Link/bin/agent'link")
	want := "'/tmp/Agent Link/bin/agent'\\''link'"
	if got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}

func TestGitHookContentQuotesBinaryPath(t *testing.T) {
	content := gitHookContent("/tmp/Agent Link/bin/agentlink")
	if !strings.Contains(content, "'/tmp/Agent Link/bin/agentlink' sync --quiet") {
		t.Fatalf("gitHookContent() did not quote binary path:\n%s", content)
	}
}

func TestZshHookContentQuotesBinaryPath(t *testing.T) {
	content := zshHookContent("/tmp/Agent Link/bin/agentlink")
	if !strings.Contains(content, "'/tmp/Agent Link/bin/agentlink' sync --quiet") {
		t.Fatalf("zshHookContent() did not quote binary path:\n%s", content)
	}
}

func TestRemoveMarkedSectionPreservesExecutableMode(t *testing.T) {
	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "post-merge")
	content := "#!/bin/sh\n" +
		"echo before\n" +
		gitHookMarkerStart + "\n" +
		"agentlink sync --quiet\n" +
		gitHookMarkerEnd + "\n" +
		"echo after\n"
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	removed, err := removeMarkedSection(hookPath)
	if err != nil {
		t.Fatalf("removeMarkedSection() failed: %v", err)
	}
	if !removed {
		t.Fatal("removeMarkedSection() removed = false, want true")
	}

	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, gitHookMarkerStart) || strings.Contains(got, "agentlink sync") {
		t.Fatalf("marked section was not removed:\n%s", got)
	}
	if !strings.Contains(got, "echo before") || !strings.Contains(got, "echo after") {
		t.Fatalf("unmanaged hook content was not preserved:\n%s", got)
	}
}

func TestRemoveMarkedSectionMissingFileIsNoop(t *testing.T) {
	removed, err := removeMarkedSection(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("removeMarkedSection() returned unexpected error: %v", err)
	}
	if removed {
		t.Fatal("removeMarkedSection() removed = true, want false")
	}
}

func TestInstallGitHooksRejectsRelativeHooksPath(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	if err := installGitHooksWithDir("relative/hooks", "/tmp/agentlink"); err == nil {
		t.Fatal("installGitHooksWithDir() succeeded for relative hooks path, want error")
	}
}
