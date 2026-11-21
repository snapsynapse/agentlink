//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hostileEnv isolates HOME and global git config so a regression cannot
// touch the real home directory of whoever runs the suite.
func hostileEnv(t *testing.T, home string) []string {
	t.Helper()
	binDir := filepath.Dir(integrationBinaryPath)
	return append(os.Environ(),
		"HOME="+home,
		"GIT_CONFIG_GLOBAL="+filepath.Join(home, ".gitconfig"),
		"GIT_CONFIG_NOSYSTEM=1",
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func assertAbsent(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("%s was created outside the repository (err=%v)", path, err)
		}
	}
}

func TestIntegrationProjectConfigCannotEscapeRepository(t *testing.T) {
	cases := []struct {
		name   string
		config string
		setup  func(t *testing.T, repo, home string)
	}{
		{name: "home link", config: "source: AGENTS.md\nlinks: [~/.zshenv]\n"},
		{name: "dot-dot link", config: "source: AGENTS.md\nlinks: [../escaped.md]\n"},
		{name: "absolute link", config: "source: AGENTS.md\nlinks: [ABSOLUTE]\n"},
		{
			name:   "directory symlink link",
			config: "source: AGENTS.md\nlinks: [sub/fish/conf.d/x.fish]\n",
			setup: func(t *testing.T, repo, home string) {
				if err := os.Symlink(filepath.Join(home, ".config"), filepath.Join(repo, "sub")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "dangling directory symlink link",
			config: "source: AGENTS.md\nlinks: [sub/x.md]\n",
			setup: func(t *testing.T, repo, home string) {
				if err := os.Symlink(filepath.Join(home, "not-yet"), filepath.Join(repo, "sub")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{name: "home source", config: "source: ~/secret\nlinks: [CLAUDE.md]\n"},
		{name: "dot-dot source", config: "source: ../home/secret\nlinks: [CLAUDE.md]\n"},
	}
	for _, tc := range cases {
		for _, args := range [][]string{{"sync"}, {"sync", "--force"}, {"sync", "--backup"}, {"check"}, {"clean"}} {
			t.Run(tc.name+"/"+strings.Join(args, " "), func(t *testing.T) {
				base := t.TempDir()
				home := filepath.Join(base, "home")
				repo := filepath.Join(base, "repo")
				safetyWrite(t, filepath.Join(home, "secret"), "SECRET")
				safetyWrite(t, filepath.Join(home, ".config", "keep"), "")
				safetyWrite(t, filepath.Join(repo, "AGENTS.md"), "echo pwned\n")
				config := strings.ReplaceAll(tc.config, "ABSOLUTE", filepath.Join(home, ".zprofile"))
				safetyWrite(t, filepath.Join(repo, ".agentlink.yaml"), config)
				if tc.setup != nil {
					tc.setup(t, repo, home)
				}
				out, err := safetyRun(t, repo, hostileEnv(t, home), args...)
				if err == nil {
					t.Fatalf("%v accepted an escaping project config:\n%s", args, out)
				}
				if !strings.Contains(out, "outside the project directory") && !strings.Contains(out, "dangling symlink") {
					t.Fatalf("unexpected error output:\n%s", out)
				}
				assertAbsent(t,
					filepath.Join(home, ".zshenv"),
					filepath.Join(home, ".zprofile"),
					filepath.Join(base, "escaped.md"),
					filepath.Join(home, ".config", "fish"),
					filepath.Join(home, "not-yet"),
					filepath.Join(repo, "CLAUDE.md"),
				)
			})
		}
	}
}

func TestIntegrationHostileCloneWithGitHooksIsContained(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	base := t.TempDir()
	home := filepath.Join(base, "home")
	origin := filepath.Join(base, "origin")
	safetyWrite(t, filepath.Join(home, "secret"), "SECRET")
	safetyWrite(t, filepath.Join(origin, "AGENTS.md"), "echo pwned\n")
	safetyWrite(t, filepath.Join(origin, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [~/.zshenv, ../escaped.md, CLAUDE.md]\n")
	env := hostileEnv(t, home)
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git(origin, "init", "-q")
	git(origin, "-c", "user.email=test@example.invalid", "-c", "user.name=test", "add", "-A")
	git(origin, "-c", "user.email=test@example.invalid", "-c", "user.name=test", "commit", "-qm", "hostile")

	installLegacyGitHooks(t, home, env)
	git(base, "clone", "-q", origin, "victim")

	assertAbsent(t, filepath.Join(home, ".zshenv"), filepath.Join(base, "escaped.md"))
	// The whole config is rejected, including its in-repo link.
	assertAbsent(t, filepath.Join(base, "victim", "CLAUDE.md"))
}

func TestIntegrationScanConfinesEachRepository(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	root := filepath.Join(base, "git")
	safetyWrite(t, filepath.Join(home, "secret"), "SECRET")

	configured := filepath.Join(root, "configured")
	safetyWrite(t, filepath.Join(configured, ".git", "HEAD"), "ref: refs/heads/main\n")
	safetyWrite(t, filepath.Join(configured, "AGENTS.md"), "echo pwned\n")
	safetyWrite(t, filepath.Join(configured, ".agentlink.yaml"), "source: AGENTS.md\nlinks: [~/.zshenv]\n")

	// An unconfigured repository whose AGENTS.md is a committed symlink to a
	// secret must not be linked even under --force.
	unconfigured := filepath.Join(root, "unconfigured")
	safetyWrite(t, filepath.Join(unconfigured, ".git", "HEAD"), "ref: refs/heads/main\n")
	if err := os.Symlink(filepath.Join(home, "secret"), filepath.Join(unconfigured, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{{"scan", root}, {"scan", root, "--force"}} {
		if out, err := safetyRun(t, base, hostileEnv(t, home), args...); err == nil {
			t.Fatalf("%v accepted escaping repositories:\n%s", args, out)
		}
		assertAbsent(t, filepath.Join(home, ".zshenv"), filepath.Join(unconfigured, "CLAUDE.md"))
	}
}

// installLegacyGitHooks recreates what 'hooks install --git' wrote in earlier
// releases: hooks in ~/.config/git/hooks and a global core.hooksPath. Users
// who installed them keep running whatever agentlink binary is on PATH.
func installLegacyGitHooks(t *testing.T, home string, env []string) string {
	t.Helper()
	hooksDir := filepath.Join(home, ".config", "git", "hooks")
	block := "#!/bin/sh\n# >>> agentlink >>>\n# Sync agent instruction symlinks after git operations.\n# Installed by: agentlink hooks install --git\n'" +
		integrationBinaryPath + "' sync --quiet 2>/dev/null || true\n# <<< agentlink <<<\n"
	for _, name := range []string{"post-checkout", "post-merge"} {
		path := filepath.Join(hooksDir, name)
		safetyWrite(t, path, block)
		if err := os.Chmod(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("set core.hooksPath: %v\n%s", err, out)
	}
	return hooksDir
}

func TestIntegrationHooksRemoveUndoesLegacyGitInstall(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	home := t.TempDir()
	env := hostileEnv(t, home)
	hooksDir := installLegacyGitHooks(t, home, env)

	if out, err := safetyRun(t, home, env, "hooks", "remove", "--git", "--dry-run"); err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(hooksDir, "post-checkout")); err != nil {
		t.Fatal("dry run deleted a hook")
	}

	out, err := safetyRun(t, home, env, "hooks", "remove", "--git")
	if err != nil {
		t.Fatalf("hooks remove: %v\n%s", err, out)
	}
	if _, err := os.Lstat(hooksDir); !os.IsNotExist(err) {
		t.Fatalf("hooks directory left behind:\n%s", out)
	}
	cmd := exec.Command("git", "config", "--global", "core.hooksPath")
	cmd.Env = env
	if value, err := cmd.Output(); err == nil {
		t.Fatalf("core.hooksPath still set to %q", value)
	}
	if out, err := safetyRun(t, home, env, "hooks", "status"); err != nil || !strings.Contains(out, "git hooks:    not installed") {
		t.Fatalf("status after removal: %v\n%s", err, out)
	}
}

func TestIntegrationInitRefusesSymlinkedConfig(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	repo := filepath.Join(base, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, ".zshenv"), filepath.Join(repo, ".agentlink.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init"}, {"init", "--force"}} {
		if out, err := safetyRun(t, repo, hostileEnv(t, home), args...); err == nil {
			t.Fatalf("%v wrote through a symlinked .agentlink.yaml:\n%s", args, out)
		}
		assertAbsent(t, filepath.Join(home, ".zshenv"))
	}
}
