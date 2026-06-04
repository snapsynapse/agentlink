package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRemovesOnlyVerifiedManagedSymlinks(t *testing.T) {
	tmpDir := setupCleanTestRepo(t)
	withWorkingDir(t, tmpDir)
	withDryRun(t, false)

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("runClean() failed: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(tmpDir, "managed.md")); !os.IsNotExist(err) {
		t.Fatalf("managed symlink still exists or stat failed with non-missing error: %v", err)
	}
	assertSymlinkTarget(t, filepath.Join(tmpDir, "broken.md"), "unrelated-missing.md")
}

func TestCleanDryRunReportsBrokenSymlinkSkipWithoutMutation(t *testing.T) {
	tmpDir := setupCleanTestRepo(t)
	withWorkingDir(t, tmpDir)
	withDryRun(t, true)

	stdout, stderr := captureOutput(t, func() {
		if err := runClean(nil, nil); err != nil {
			t.Fatalf("runClean() failed: %v", err)
		}
	})

	assertSymlinkTarget(t, filepath.Join(tmpDir, "managed.md"), "source.md")
	assertSymlinkTarget(t, filepath.Join(tmpDir, "broken.md"), "unrelated-missing.md")
	if !strings.Contains(stdout, "would remove 1 symlinks, skip 1 items") {
		t.Fatalf("dry-run summary did not report expected counts:\nstdout=%s\nstderr=%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "broken symlink; target ownership cannot be verified") {
		t.Fatalf("dry-run did not warn about skipped broken symlink:\nstdout=%s\nstderr=%s", stdout, stderr)
	}
}

func setupCleanTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, ".agentlink.yaml"), []byte("source: source.md\nlinks:\n  - managed.md\n  - broken.md\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "source.md"), []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("source.md", filepath.Join(tmpDir, "managed.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("unrelated-missing.md", filepath.Join(tmpDir, "broken.md")); err != nil {
		t.Fatal(err)
	}
	return tmpDir
}

func assertSymlinkTarget(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%s) failed: %v", path, err)
	}
	if got != want {
		t.Fatalf("Readlink(%s) = %q, want %q", path, got, want)
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

func withDryRun(t *testing.T, value bool) {
	t.Helper()

	oldDryRun := dryRun
	dryRun = value
	t.Cleanup(func() { dryRun = oldDryRun })
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	fn()

	if err := stdoutWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := stderrWriter.Close(); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := io.Copy(&stdout, stdoutReader); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(&stderr, stderrReader); err != nil {
		t.Fatal(err)
	}
	return stdout.String(), stderr.String()
}
