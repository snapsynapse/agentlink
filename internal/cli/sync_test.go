package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinmose/agentlink/internal/symlink"
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

	manager := symlink.NewManager(true, true, false)
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
