package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunScanPreservesUnmanagedRegularFile(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("instructions"), 0644); err != nil {
		t.Fatal(err)
	}

	targets := repoLinkTargets()
	if len(targets) == 0 {
		t.Fatal("registry has no repo link targets to exercise")
	}
	if err := os.WriteFile(filepath.Join(repo, targets[0]), []byte("conflict"), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested := dryRun, force, verbose, scanDir, scanNested
	t.Cleanup(func() {
		dryRun, force, verbose, scanDir, scanNested = oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested
	})
	dryRun, force, verbose, scanDir, scanNested = false, false, false, "", false

	err := runScan(scanCmd, []string{root})
	if err != nil {
		t.Fatalf("runScan() error = %v, want preserved unmanaged file", err)
	}
	got, err := os.ReadFile(filepath.Join(repo, targets[0]))
	if err != nil || string(got) != "conflict" {
		t.Fatalf("existing file changed: content=%q err=%v", got, err)
	}
}

func TestRunScanRejectsInvalidAgentsSource(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, "AGENTS.md"), 0755); err != nil {
		t.Fatal(err)
	}

	oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested := dryRun, force, verbose, scanDir, scanNested
	t.Cleanup(func() {
		dryRun, force, verbose, scanDir, scanNested = oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested
	})
	dryRun, force, verbose, scanDir, scanNested = false, false, false, "", false

	err := runScan(scanCmd, []string{root})
	if err == nil {
		t.Fatal("runScan() error = nil, want invalid source error")
	}
	if got := err.Error(); got != "scan completed with 1 link error(s)" {
		t.Fatalf("runScan() error = %q, want aggregate error count", got)
	}
	for _, target := range repoLinkTargets() {
		if _, err := os.Lstat(filepath.Join(repo, target)); !os.IsNotExist(err) {
			t.Fatalf("scan created %s for invalid source", target)
		}
	}
}

func TestRunScanRespectsExplicitProjectConfig(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		"AGENTS.md":       "shared",
		"CLAUDE.md":       "@AGENTS.md\n\n# Claude-specific\n",
		".agentlink.yaml": "source: AGENTS.md\nlinks:\n  - GEMINI.md\n",
	} {
		if err := os.WriteFile(filepath.Join(repo, path), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	setScanTestFlags(t, false)
	if err := runScan(scanCmd, []string{root}); err != nil {
		t.Fatalf("runScan() error = %v", err)
	}
	assertScanSymlinkTarget(t, filepath.Join(repo, "GEMINI.md"), "AGENTS.md")
	got, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil || string(got) != "@AGENTS.md\n\n# Claude-specific\n" {
		t.Fatalf("layered CLAUDE.md changed: content=%q err=%v", got, err)
	}
}

func TestRunScanKeepsDeclaredLinkStrict(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{
		"AGENTS.md":       "shared",
		"CLAUDE.md":       "@AGENTS.md\n",
		".agentlink.yaml": "source: AGENTS.md\nlinks:\n  - CLAUDE.md\n",
	} {
		if err := os.WriteFile(filepath.Join(repo, path), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	setScanTestFlags(t, false)
	err := runScan(scanCmd, []string{root})
	if err == nil || err.Error() != "scan completed with 1 link error(s)" {
		t.Fatalf("runScan() error = %v, want one declared-link error", err)
	}
	got, readErr := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if readErr != nil || string(got) != "@AGENTS.md\n" {
		t.Fatalf("declared conflicting file changed: content=%q err=%v", got, readErr)
	}
}

func TestRunScanExplicitConfigOverridesNestedDiscoveryAndDefaultSource(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	packageDir := filepath.Join(repo, "packages", "api")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(repo, "GUIDE.md"):        "configured source",
		filepath.Join(repo, "AGENTS.md"):       "unmanaged root source",
		filepath.Join(repo, "CLAUDE.md"):       "@GUIDE.md\n\n# Claude-only\n",
		filepath.Join(packageDir, "AGENTS.md"): "unmanaged nested source",
		filepath.Join(repo, ".agentlink.yaml"): "source: GUIDE.md\nlinks:\n  - GEMINI.md\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	setScanTestFlags(t, true)
	if err := runScan(scanCmd, []string{root}); err != nil {
		t.Fatalf("runScan() error = %v", err)
	}
	assertScanSymlinkTarget(t, filepath.Join(repo, "GEMINI.md"), "GUIDE.md")
	if got, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md")); err != nil || string(got) != "@GUIDE.md\n\n# Claude-only\n" {
		t.Fatalf("configured wrapper changed: content=%q err=%v", got, err)
	}
	for _, path := range []string{
		filepath.Join(repo, "QWEN.md"),
		filepath.Join(packageDir, "CLAUDE.md"),
		filepath.Join(packageDir, "GEMINI.md"),
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Errorf("explicit config allowed undeclared scan output %s: %v", path, err)
		}
	}
}

func TestRunScanNestedIsOptIn(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	packageDir := filepath.Join(repo, "packages", "api")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(repo, "AGENTS.md"), filepath.Join(packageDir, "AGENTS.md")} {
		if err := os.WriteFile(path, []byte("instructions"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	setScanTestFlags(t, false)
	if err := runScan(scanCmd, []string{root}); err != nil {
		t.Fatalf("root-only runScan() error = %v", err)
	}
	if _, err := os.Lstat(filepath.Join(packageDir, nestedRepoLinkTargets()[0])); !os.IsNotExist(err) {
		t.Fatalf("nested link exists without --nested: %v", err)
	}

	setScanTestFlags(t, true)
	if err := runScan(scanCmd, []string{root}); err != nil {
		t.Fatalf("nested runScan() error = %v", err)
	}
	assertScanSymlinkTarget(t, filepath.Join(packageDir, nestedRepoLinkTargets()[0]), "AGENTS.md")
	if _, err := os.Lstat(filepath.Join(packageDir, "QWEN.md")); !os.IsNotExist(err) {
		t.Fatalf("nested scan created an alias without verified nested support: %v", err)
	}
}

func TestNestedRepoLinkTargetsFailClosed(t *testing.T) {
	got := nestedRepoLinkTargets()
	want := map[string]bool{"CLAUDE.md": true, "GEMINI.md": true}
	if len(got) != len(want) {
		t.Fatalf("nestedRepoLinkTargets() = %v, want only documented targets", got)
	}
	for _, target := range got {
		if !want[target] {
			t.Errorf("nestedRepoLinkTargets() includes undocumented target %q", target)
		}
	}
}

func TestRepoLinkTargetsNeverSelfLinkAgentsSource(t *testing.T) {
	for _, target := range repoLinkTargets() {
		if filepath.Clean(target) == "AGENTS.md" {
			t.Fatalf("repoLinkTargets() includes source self-link %q", target)
		}
	}
}

func TestFindNestedAgentsFilesSkipsGeneratedAndNestedRepos(t *testing.T) {
	repo := t.TempDir()
	wanted := filepath.Join(repo, "packages", "api", "AGENTS.md")
	skipped := []string{
		filepath.Join(repo, "node_modules", "pkg", "AGENTS.md"),
		filepath.Join(repo, ".hidden", "AGENTS.md"),
		filepath.Join(repo, "nested-repo", "AGENTS.md"),
	}
	for _, path := range append([]string{wanted}, skipped...) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("instructions"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(repo, "nested-repo", ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	got := findNestedAgentsFiles(repo)
	if len(got) != 1 || got[0] != wanted {
		t.Fatalf("findNestedAgentsFiles() = %v, want [%s]", got, wanted)
	}
}

func setScanTestFlags(t *testing.T, nested bool) {
	t.Helper()
	oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested := dryRun, force, verbose, scanDir, scanNested
	dryRun, force, verbose, scanDir, scanNested = false, false, false, "", nested
	t.Cleanup(func() {
		dryRun, force, verbose, scanDir, scanNested = oldDryRun, oldForce, oldVerbose, oldScanDir, oldScanNested
	})
}

func assertScanSymlinkTarget(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%s): %v", path, err)
	}
	if got != want {
		t.Fatalf("Readlink(%s) = %q, want %q", path, got, want)
	}
}

func TestFindGitReposDetectsStandardRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	repos := findGitRepos(root)
	if len(repos) != 1 || repos[0] != repo {
		t.Fatalf("findGitRepos() = %v, want [%s]", repos, repo)
	}
}

func TestFindGitReposDetectsWorktreeStyleRepo(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "worktree-repo")
	if err := os.MkdirAll(repo, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".git"), []byte("gitdir: /tmp/example"), 0644); err != nil {
		t.Fatal(err)
	}

	repos := findGitRepos(root)
	if len(repos) != 1 || repos[0] != repo {
		t.Fatalf("findGitRepos() = %v, want [%s]", repos, repo)
	}
}
