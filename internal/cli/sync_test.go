package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/snapsynapse/agentlink/internal/symlink"
)

func TestProcessLinkDryRunBackupDoesNotMutateExistingFile(t *testing.T) {
	oldDryRun := dryRun
	oldSyncBackup := syncBackup
	dryRun = true
	syncBackup = true
	defer func() {
		dryRun = oldDryRun
		syncBackup = oldSyncBackup
	}()

	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "source.md")
	link := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	manager := symlink.NewManager(true, true)
	if err := processLink(manager, link, source); err != nil {
		t.Fatalf("processLink() failed: %v", err)
	}

	data, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Fatalf("dry-run backup changed existing file to %q", data)
	}
	if _, err := os.Stat(link + ".bak"); err == nil {
		t.Fatal("dry-run backup created a backup file")
	}
}

func TestBackupFileNeverOverwritesExistingBackups(t *testing.T) {
	oldBackupNow := backupNow
	backupNow = func() time.Time { return time.Date(2026, 7, 10, 12, 34, 56, 0, time.UTC) }
	defer func() { backupNow = oldBackupNow }()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(path, []byte("current"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".bak", []byte("original backup"), 0644); err != nil {
		t.Fatal(err)
	}

	// Occupy the timestamped name. The implementation must choose a suffix
	// instead of replacing it.
	timestamped := path + ".20260710-123456.bak"
	if err := os.WriteFile(timestamped, []byte("timestamped backup"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := backupFile(path); err != nil {
		t.Fatalf("backupFile() failed: %v", err)
	}
	data, err := os.ReadFile(path + ".bak")
	if err != nil || string(data) != "original backup" {
		t.Fatalf("original backup changed: data=%q err=%v", data, err)
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	foundCurrent := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "AGENTS.md.") {
			data, readErr := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
			if readErr == nil && string(data) == "current" {
				foundCurrent = true
			}
		}
	}
	if !foundCurrent {
		t.Fatal("new backup content was not preserved under a unique name")
	}
}
