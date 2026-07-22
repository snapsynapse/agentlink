package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDirectoryAccessDoesNotCreateMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	err := checkDirectoryAccess(path)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("checkDirectoryAccess() error = %v, want missing-directory error", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("checkDirectoryAccess() created or changed missing path: %v", statErr)
	}
}

func TestCheckDirectoryAccessCleansUpProbe(t *testing.T) {
	dir := t.TempDir()
	if err := checkDirectoryAccess(dir); err != nil {
		t.Fatalf("checkDirectoryAccess() failed: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("checkDirectoryAccess() left temporary files: %v", entries)
	}
}

func TestCheckSymlinkSupportUsesIsolatedTemporaryPaths(t *testing.T) {
	if err := checkSymlinkSupport(); err != nil {
		t.Fatalf("checkSymlinkSupport() failed: %v", err)
	}
}
