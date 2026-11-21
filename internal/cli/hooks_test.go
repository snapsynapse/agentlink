package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacyGitHookBlock is the block earlier releases wrote into Git hooks.
func legacyGitHookBlock(binaryPath string) string {
	return gitHookMarkerStart + "\n" +
		"# Sync agent instruction symlinks after git operations.\n" +
		"# Installed by: agentlink hooks install --git\n" +
		"'" + binaryPath + "' sync --quiet 2>/dev/null || true\n" +
		gitHookMarkerEnd + "\n"
}

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func setDryRun(t *testing.T, value bool) {
	t.Helper()
	old := dryRun
	dryRun = value
	t.Cleanup(func() { dryRun = old })
}

func TestHooksInstallIsRetired(t *testing.T) {
	for _, args := range [][]string{{}, {"--all"}, {"--git", "--zsh"}} {
		err := hooksInstallCmd.RunE(hooksInstallCmd, args)
		if !errors.Is(err, errHooksInstallRetired) {
			t.Fatalf("install %v error = %v, want retirement error", args, err)
		}
	}
	if !strings.Contains(errHooksInstallRetired.Error(), "agentlink hooks remove --all") {
		t.Fatal("retirement error does not explain removal")
	}
}

func TestRemoveMarkedSectionPreservesExecutableMode(t *testing.T) {
	hookPath := filepath.Join(t.TempDir(), "post-merge")
	content := "#!/bin/sh\n" +
		"echo before\n" +
		legacyGitHookBlock("/tmp/agentlink") +
		"echo after\n"
	writeFile(t, hookPath, content, 0755)

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
	got, _ := os.ReadFile(hookPath)
	if string(got) != "#!/bin/sh\necho before\necho after\n" {
		t.Fatalf("unexpected content after removal:\n%s", got)
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

func TestRemoveMarkedSectionRejectsIncompleteBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "post-checkout")
	original := "#!/bin/sh\n" + gitHookMarkerStart + "\nagentlink sync --quiet\n"
	writeFile(t, path, original, 0755)
	removed, err := removeMarkedSection(path)
	if err == nil || removed {
		t.Fatalf("removeMarkedSection() = %v, %v; want incomplete marker error", removed, err)
	}
	if got, _ := os.ReadFile(path); string(got) != original {
		t.Fatalf("incomplete block was modified:\n%s", got)
	}
}

func TestRemoveMarkedSectionRestoresHookWithEarlyExit(t *testing.T) {
	// Earlier releases inserted the block right after the shebang.
	path := filepath.Join(t.TempDir(), "post-checkout")
	original := "#!/bin/sh\nexit 0\n"
	writeFile(t, path, "#!/bin/sh\n"+legacyGitHookBlock("/tmp/agentlink")+"exit 0\n", 0755)
	if _, err := removeMarkedSection(path); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != original {
		t.Fatalf("original hook not restored: %q", got)
	}
}

func TestRemoveGitHooksUndoesDefaultInstall(t *testing.T) {
	hooksDir := filepath.Join(t.TempDir(), ".config", "git", "hooks")
	for _, name := range gitHookNames {
		writeFile(t, filepath.Join(hooksDir, name), legacyHookStub+legacyGitHookBlock("/tmp/agentlink"), 0755)
	}
	unset := 0
	if err := removeGitHooksFromDir(hooksDir, hooksDir, func() error { unset++; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(hooksDir); !os.IsNotExist(err) {
		t.Fatalf("hooks directory not removed: %v", err)
	}
	if unset != 1 {
		t.Fatalf("unset core.hooksPath called %d times, want 1", unset)
	}
}

func TestRemoveGitHooksKeepsHooksPathForOtherHooks(t *testing.T) {
	hooksDir := filepath.Join(t.TempDir(), "hooks")
	writeFile(t, filepath.Join(hooksDir, "post-checkout"), legacyHookStub+legacyGitHookBlock("/tmp/agentlink"), 0755)
	userHook := "#!/bin/sh\n" + legacyGitHookBlock("/tmp/agentlink") + "echo mine\n"
	writeFile(t, filepath.Join(hooksDir, "post-merge"), userHook, 0755)
	writeFile(t, filepath.Join(hooksDir, "pre-commit"), "#!/bin/sh\ngitleaks protect\n", 0755)

	for _, defaultDir := range []string{hooksDir, filepath.Join(t.TempDir(), "elsewhere")} {
		unset := 0
		if err := removeGitHooksFromDir(hooksDir, defaultDir, func() error { unset++; return nil }); err != nil {
			t.Fatal(err)
		}
		if unset != 0 {
			t.Fatalf("defaultDir=%s: unset core.hooksPath while other hooks remain", defaultDir)
		}
	}
	if _, err := os.Lstat(filepath.Join(hooksDir, "post-checkout")); !os.IsNotExist(err) {
		t.Fatal("stub-only hook was not deleted")
	}
	if got, _ := os.ReadFile(filepath.Join(hooksDir, "post-merge")); string(got) != "#!/bin/sh\necho mine\n" {
		t.Fatalf("user hook content not preserved: %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(hooksDir, "pre-commit")); !strings.Contains(string(got), "gitleaks") {
		t.Fatal("unrelated hook changed")
	}
}

func TestRemoveGitHooksDryRunChangesNothing(t *testing.T) {
	setDryRun(t, true)
	hooksDir := filepath.Join(t.TempDir(), "hooks")
	content := legacyHookStub + legacyGitHookBlock("/tmp/agentlink")
	for _, name := range gitHookNames {
		writeFile(t, filepath.Join(hooksDir, name), content, 0755)
	}
	if err := removeGitHooksFromDir(hooksDir, hooksDir, func() error {
		t.Fatal("dry run unset core.hooksPath")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range gitHookNames {
		if got, err := os.ReadFile(filepath.Join(hooksDir, name)); err != nil || string(got) != content {
			t.Fatalf("dry run modified %s: %q %v", name, got, err)
		}
	}
}

func TestRemoveLegacyLaunchdLogsSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "agentlink-sync.log")
	writeFile(t, logPath, "old output", 0644)
	victim := filepath.Join(dir, "victim")
	writeFile(t, victim, "keep", 0644)
	linkPath := filepath.Join(dir, "agentlink-sync.err")
	if err := os.Symlink(victim, linkPath); err != nil {
		t.Fatal(err)
	}

	removeLegacyLaunchdLogs([]string{logPath, linkPath, filepath.Join(dir, "missing")})

	if _, err := os.Lstat(logPath); !os.IsNotExist(err) {
		t.Fatal("regular legacy log was not removed")
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatal("symlinked log path was removed")
	}
	if got, _ := os.ReadFile(victim); string(got) != "keep" {
		t.Fatal("symlink target was modified")
	}
}
