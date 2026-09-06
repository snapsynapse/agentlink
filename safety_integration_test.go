//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func safetyWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func safetyRun(t *testing.T, dir string, env []string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(integrationBinaryPath, args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestIntegrationBackupRejectsSourceParentAliasBeforeMutation(t *testing.T) {
	for _, content := range []string{"", "authoritative source\n"} {
		for _, preview := range []bool{false, true} {
			dir := t.TempDir()
			safetyWrite(t, filepath.Join(dir, "AGENTS.md"), content)
			safetyWrite(t, filepath.Join(dir, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [alias/AGENTS.md]\n")
			if err := os.Symlink(".", filepath.Join(dir, "alias")); err != nil {
				t.Fatal(err)
			}
			args := []string{"sync", "--backup"}
			if preview {
				args = append(args, "--dry-run")
			}
			out, err := safetyRun(t, dir, nil, args...)
			if err == nil || !strings.Contains(out, "refusing to replace source") {
				t.Fatalf("expected source rejection: %v\n%s", err, out)
			}
			got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
			if err != nil || string(got) != content {
				t.Fatalf("source changed: %q %v", got, err)
			}
			if _, err := os.Lstat(filepath.Join(dir, "AGENTS.md.bak")); !os.IsNotExist(err) {
				t.Fatal("backup was created")
			}
		}
	}
}

func TestIntegrationPhysicalLinkTargetAndIndependentCheck(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	physical := filepath.Join(dir, "elsewhere", "deep")
	safetyWrite(t, filepath.Join(repo, "AGENTS.md"), "correct source")
	safetyWrite(t, filepath.Join(repo, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [aliases/CLAUDE.md]\n")
	if err := os.MkdirAll(physical, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physical, filepath.Join(repo, "aliases")); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, "aliases", "CLAUDE.md")
	// The old implementation incorrectly approved this unreadable link.
	if err := os.Symlink("../AGENTS.md", link); err != nil {
		t.Fatal(err)
	}
	if out, err := safetyRun(t, repo, nil, "check"); err == nil {
		t.Fatalf("check approved a broken alias: %s", out)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if out, err := safetyRun(t, repo, nil, "sync"); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}
	got, err := os.ReadFile(link)
	if err != nil || string(got) != "correct source" {
		t.Fatalf("OS cannot read expected bytes: %q %v", got, err)
	}
	if out, err := safetyRun(t, repo, nil, "check"); err != nil {
		t.Fatalf("check rejected correct alias: %v\n%s", err, out)
	}
}

func TestIntegrationUnrelatedBrokenLinksArePreserved(t *testing.T) {
	for _, command := range []string{"sync", "backup", "scan"} {
		dir := t.TempDir()
		safetyWrite(t, filepath.Join(dir, "AGENTS.md"), "source")
		if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
		if command != "scan" {
			safetyWrite(t, filepath.Join(dir, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [CLAUDE.md]\n")
		}
		link := filepath.Join(dir, "CLAUDE.md")
		if err := os.Symlink("missing-personal.md", link); err != nil {
			t.Fatal(err)
		}
		args := []string{command}
		if command == "backup" {
			args = []string{"sync", "--backup"}
		}
		if command == "scan" {
			args = append(args, dir)
		}
		if out, err := safetyRun(t, dir, nil, args...); err == nil {
			t.Fatalf("expected ownership error: %s", out)
		}
		if got, err := os.Readlink(link); err != nil || got != "missing-personal.md" {
			t.Fatalf("unmanaged link changed: %q %v", got, err)
		}
	}
}

func TestIntegrationDanglingConfigDoesNotSelectAnotherScope(t *testing.T) {
	dir := t.TempDir()
	safetyWrite(t, filepath.Join(dir, "AGENTS.md"), "source")
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing-config.yaml", filepath.Join(dir, ".agentlink.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"scan", dir}, {"sync"}, {"check"}, {"clean"}} {
		if out, err := safetyRun(t, dir, nil, args...); err == nil {
			t.Fatalf("accepted dangling config: %v\n%s", args, out)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 3 {
		t.Fatalf("unexpected mutations: %v %v", entries, err)
	}
}

func TestIntegrationHookRemovalDryRunPreservesAllState(t *testing.T) {
	dir := t.TempDir()
	gitConfig := filepath.Join(dir, "gitconfig")
	hooks := filepath.Join(dir, "hooks")
	safetyWrite(t, gitConfig, "[core]\n hooksPath = "+hooks+"\n")
	block := "# >>> agentlink >>>\necho fixture\n# <<< agentlink <<<\n"
	paths := []string{filepath.Join(hooks, "post-checkout"), filepath.Join(hooks, "post-merge"), filepath.Join(dir, ".zshrc")}
	for _, path := range paths {
		safetyWrite(t, path, block)
	}
	plist := filepath.Join(dir, "Library", "LaunchAgents", "com.agentlink.sync.plist")
	safetyWrite(t, plist, "fixture plist")
	stub := filepath.Join(dir, "bin", "launchctl")
	safetyWrite(t, stub, "#!/bin/sh\nprintf invoked > \"$HOME/launchctl-invoked\"\n")
	if err := os.Chmod(stub, 0755); err != nil {
		t.Fatal(err)
	}
	env := []string{"HOME=" + dir, "GIT_CONFIG_GLOBAL=" + gitConfig, "GIT_CONFIG_NOSYSTEM=1", "PATH=" + filepath.Dir(stub) + ":" + os.Getenv("PATH")}
	selectors := []string{"--git", "--zsh", "--all"}
	if runtime.GOOS == "darwin" {
		selectors = append(selectors, "--launchd")
	}
	for _, selector := range selectors {
		if out, err := safetyRun(t, dir, env, "hooks", "remove", selector, "--dry-run"); err != nil {
			t.Fatalf("%s: %v\n%s", selector, err, out)
		}
	}
	for _, path := range paths {
		if data, err := os.ReadFile(path); err != nil || string(data) != block {
			t.Fatalf("dry-run changed %s: %q %v", path, data, err)
		}
	}
	if data, err := os.ReadFile(plist); err != nil || string(data) != "fixture plist" {
		t.Fatal("dry-run changed plist")
	}
	if _, err := os.Stat(filepath.Join(dir, "launchctl-invoked")); !os.IsNotExist(err) {
		t.Fatal("dry-run executed launchctl")
	}
}

func TestIntegrationGlobalSelectorLeavesProjectUntouched(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	repo := filepath.Join(dir, "repo")
	safetyWrite(t, filepath.Join(home, "AGENTS.md"), "global")
	safetyWrite(t, filepath.Join(home, ".config", "agentlink", "config.yaml"), "source: ~/AGENTS.md\nlinks: [~/CLAUDE.md]\n")
	safetyWrite(t, filepath.Join(repo, "AGENTS.md"), "project")
	safetyWrite(t, filepath.Join(repo, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [CLAUDE.md]\n")
	if out, err := safetyRun(t, repo, []string{"HOME=" + home}, "sync", "--global"); err != nil {
		t.Fatalf("global sync: %v\n%s", err, out)
	}
	if got, err := os.ReadFile(filepath.Join(home, "CLAUDE.md")); err != nil || string(got) != "global" {
		t.Fatalf("global alias: %q %v", got, err)
	}
	if _, err := os.Lstat(filepath.Join(repo, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatal("global sync modified project")
	}
	if out, err := safetyRun(t, repo, []string{"HOME=" + home}, "check", "--global"); err != nil {
		t.Fatalf("global check selected the missing project alias: %v\n%s", err, out)
	}
	if out, err := safetyRun(t, repo, []string{"HOME=" + home}, "clean", "--global"); err != nil {
		t.Fatalf("global clean: %v\n%s", err, out)
	}
	if _, err := os.Lstat(filepath.Join(home, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatal("global clean did not remove global alias")
	}
	if got, err := os.ReadFile(filepath.Join(repo, "AGENTS.md")); err != nil || string(got) != "project" {
		t.Fatalf("global clean changed project source: %q %v", got, err)
	}
}

func TestIntegrationForcedSourceSymlinkCannotReplaceBackingFile(t *testing.T) {
	for _, args := range [][]string{{"sync", "--force"}, {"sync", "--force", "--backup"}} {
		dir := t.TempDir()
		safetyWrite(t, filepath.Join(dir, "REAL.md"), "must survive")
		safetyWrite(t, filepath.Join(dir, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [REAL.md]\n")
		if err := os.Symlink("REAL.md", filepath.Join(dir, "AGENTS.md")); err != nil {
			t.Fatal(err)
		}
		if out, err := safetyRun(t, dir, nil, args...); err == nil {
			t.Fatalf("expected source protection: %s", out)
		}
		if got, err := os.ReadFile(filepath.Join(dir, "REAL.md")); err != nil || string(got) != "must survive" {
			t.Fatalf("source backing file changed: %q %v", got, err)
		}
		if _, err := os.Lstat(filepath.Join(dir, "REAL.md.bak")); !os.IsNotExist(err) {
			t.Fatal("source backing file was backed up")
		}
	}
}
