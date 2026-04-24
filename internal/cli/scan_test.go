package cli

import (
	"os"
	"path/filepath"
	"testing"
)

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
