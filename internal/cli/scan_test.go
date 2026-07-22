package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunScanReturnsErrorWhenLinkCannotBeCreated(t *testing.T) {
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

	oldDryRun, oldForce, oldVerbose, oldScanDir := dryRun, force, verbose, scanDir
	t.Cleanup(func() {
		dryRun, force, verbose, scanDir = oldDryRun, oldForce, oldVerbose, oldScanDir
	})
	dryRun, force, verbose, scanDir = false, false, false, ""

	err := runScan(scanCmd, []string{root})
	if err == nil {
		t.Fatal("runScan() error = nil, want aggregate link error")
	}
	if got := err.Error(); got != "scan completed with 1 link error(s)" {
		t.Fatalf("runScan() error = %q, want aggregate error count", got)
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

	oldDryRun, oldForce, oldVerbose, oldScanDir := dryRun, force, verbose, scanDir
	t.Cleanup(func() {
		dryRun, force, verbose, scanDir = oldDryRun, oldForce, oldVerbose, oldScanDir
	})
	dryRun, force, verbose, scanDir = false, false, false, ""

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
