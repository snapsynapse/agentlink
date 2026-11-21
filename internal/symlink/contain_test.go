package symlink

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckContained(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{filepath.Join(root, "inner"), outside} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	mustSymlink := func(target, link string) {
		t.Helper()
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
	}
	mustSymlink("../outside", filepath.Join(root, "escape"))
	mustSymlink("inner", filepath.Join(root, "alias"))
	mustSymlink("../missing", filepath.Join(root, "dangling"))
	mustSymlink("../outside/secret", filepath.Join(root, "SECRET.md"))
	mustSymlink(filepath.Join(outside, "secret"), filepath.Join(root, "inner", "LEAF.md"))

	tests := []struct {
		name       string
		path       string
		followLeaf bool
		wantErr    bool
	}{
		{"plain file", filepath.Join(root, "CLAUDE.md"), false, false},
		{"missing nested dirs", filepath.Join(root, "a", "b", "CLAUDE.md"), false, false},
		{"directory symlink inside root", filepath.Join(root, "alias", "CLAUDE.md"), false, false},
		{"existing source", filepath.Join(root, "AGENTS.md"), true, false},
		{"missing source", filepath.Join(root, "MISSING.md"), true, false},
		{"leaf symlink not followed for links", filepath.Join(root, "inner", "LEAF.md"), false, false},
		{"root itself", root, false, true},
		{"parent of root", base, false, true},
		{"sibling directory", filepath.Join(outside, "x.md"), false, true},
		{"prefix-sharing sibling", root + "-evil/CLAUDE.md", false, true},
		{"directory symlink escape", filepath.Join(root, "escape", "x.md"), false, true},
		{"directory symlink escape missing child", filepath.Join(root, "escape", "new", "x.md"), false, true},
		{"dangling directory symlink", filepath.Join(root, "dangling", "x.md"), false, true},
		{"source symlink escape", filepath.Join(root, "SECRET.md"), true, true},
		{"source leaf symlink escape", filepath.Join(root, "inner", "LEAF.md"), true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckContained(root, tt.path, tt.followLeaf)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckContained(%s) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
	if err := CheckContained(root, filepath.Join(outside, "x.md"), false); !errors.Is(err, ErrOutsideRoot) {
		t.Fatalf("escape error = %v, want ErrOutsideRoot", err)
	}
}

func TestRootedManagerRefusesEscapes(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	outside := filepath.Join(base, "outside")
	for _, dir := range []string{root, outside} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(outside, "secret")
	if err := os.WriteFile(secret, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}

	for _, force := range []bool{false, true} {
		manager := NewManager(false, force).WithRoot(root)
		for _, link := range []string{
			filepath.Join(outside, "CLAUDE.md"),
			filepath.Join(root, "escape", "CLAUDE.md"),
		} {
			if _, err := manager.FixLink(link, source); err == nil {
				t.Fatalf("force=%v: FixLink(%s) escaped the root", force, link)
			}
			if _, err := os.Lstat(filepath.Join(outside, "CLAUDE.md")); !os.IsNotExist(err) {
				t.Fatalf("force=%v: link created outside the root", force)
			}
		}
		if _, err := manager.FixLink(filepath.Join(root, "CLAUDE.md"), secret); err == nil {
			t.Fatalf("force=%v: FixLink accepted a source outside the root", force)
		}
		if err := manager.ValidateSource(secret); err == nil {
			t.Fatalf("force=%v: ValidateSource accepted a source outside the root", force)
		}
	}

	// WithForce keeps the root.
	if _, err := NewManager(false, false).WithRoot(root).WithForce(true).FixLink(filepath.Join(outside, "CLAUDE.md"), source); err == nil {
		t.Fatal("WithForce dropped the root")
	}
	// An unrooted manager (global scope) keeps its existing behavior.
	if _, err := NewManager(false, false).FixLink(filepath.Join(outside, "CLAUDE.md"), source); err != nil {
		t.Fatalf("unrooted manager: %v", err)
	}
}
