package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFilePreservesSymlinkAndMode(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "dotfiles", "zshrc")
	if err := os.MkdirAll(filepath.Dir(real), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, ".zshrc")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWriteFile(link, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("symlink was replaced: %v", err)
	}
	if got, _ := os.ReadFile(real); string(got) != "new" {
		t.Fatalf("target = %q, want new", got)
	}
	if info, _ := os.Stat(real); info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %v, want 0600", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(real))
	if len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v", entries)
	}
}

func TestWriteNewFileNeverFollowsSymlinks(t *testing.T) {
	dir := t.TempDir()
	victim := filepath.Join(dir, "victim")
	for _, overwrite := range []bool{false, true} {
		path := filepath.Join(dir, "dangling")
		os.Remove(path)
		if err := os.Symlink(victim, path); err != nil {
			t.Fatal(err)
		}
		if err := WriteNewFile(path, []byte("x"), 0644, overwrite); err == nil {
			t.Fatalf("overwrite=%v: wrote through a dangling symlink", overwrite)
		}
		if _, err := os.Lstat(victim); !os.IsNotExist(err) {
			t.Fatalf("overwrite=%v: symlink target was created", overwrite)
		}
	}

	existing := filepath.Join(dir, "existing")
	if err := os.WriteFile(existing, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := WriteNewFile(existing, []byte("x"), 0644, false); !errors.Is(err, os.ErrExist) {
		t.Fatalf("error = %v, want ErrExist", err)
	}
	if err := WriteNewFile(existing, []byte("replaced"), 0644, true); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(existing); string(got) != "replaced" {
		t.Fatalf("content = %q", got)
	}

	fresh := filepath.Join(dir, "fresh")
	if err := WriteNewFile(fresh, []byte("new"), 0644, false); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(fresh); string(got) != "new" {
		t.Fatalf("content = %q", got)
	}
}
