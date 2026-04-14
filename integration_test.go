//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationBasicWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	
	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	// Create a fake .git directory
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatal(err)
	}
	
	// Use the pre-built binary
	binaryPath := filepath.Join(origDir, "agentlink")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found at %s. Make sure to run 'go build -o agentlink ./cmd/agentlink' first", binaryPath)
	}
	
	// Test 1: Init command
	cmd := exec.Command(binaryPath, "init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "Created") {
		t.Errorf("Init output doesn't contain 'Created': %s", output)
	}
	
	// Verify config file was created
	if _, err := os.Stat(".agentlink.yaml"); err != nil {
		t.Error("Config file was not created")
	}
	
	// Test 2: Check command (should show problems since source doesn't exist)
	cmd = exec.Command(binaryPath, "check")
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("Check should have failed since source doesn't exist")
	}
	
	// Test 3: Create source file and sync
	sourceContent := "# Test Source\nThis is a test instruction file."
	if err := os.WriteFile("CLAUDE.md", []byte(sourceContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	cmd = exec.Command(binaryPath, "sync")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Sync failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "[create]") {
		t.Errorf("Sync output doesn't contain '[create]': %s", output)
	}
	
	// Test 4: Verify symlinks were created
	for _, link := range []string{"AGENTS.md", "OPENCODE.md"} {
		info, err := os.Lstat(link)
		if err != nil {
			t.Errorf("Link %s was not created: %v", link, err)
			continue
		}
		
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("File %s is not a symlink", link)
		}
		
		// Verify content is accessible through symlink
		content, err := os.ReadFile(link)
		if err != nil {
			t.Errorf("Cannot read through symlink %s: %v", link, err)
			continue
		}
		
		if string(content) != sourceContent {
			t.Errorf("Content through symlink %s doesn't match source", link)
		}
	}
	
	// Test 5: Check command (should pass now)
	cmd = exec.Command(binaryPath, "check")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Check failed after sync: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "All links are correctly configured") {
		t.Errorf("Check output doesn't indicate success: %s", output)
	}
	
	// Test 6: Sync again (should be idempotent)
	cmd = exec.Command(binaryPath, "sync")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Second sync failed: %v\nOutput: %s", err, output)
	}
	
	// Should not contain [create] this time
	if strings.Contains(string(output), "[create]") {
		t.Errorf("Second sync should not create new links: %s", output)
	}
	
	// Test 7: Clean command
	cmd = exec.Command(binaryPath, "clean", "--dry-run")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Clean dry-run failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "would remove") {
		t.Errorf("Clean dry-run output doesn't mention removal: %s", output)
	}
}

func TestIntegrationDoctorCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Use the pre-built binary
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(origDir, "agentlink")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found at %s. Make sure to run 'go build -o agentlink ./cmd/agentlink' first", binaryPath)
	}
	
	// Run doctor command
	cmd := exec.Command(binaryPath, "doctor")
	output, _ := cmd.CombinedOutput()
	// Doctor might return non-zero exit code if there are warnings, but that's OK
	
	expectedStrings := []string{
		"Agentlink Doctor",
		"Operating System:",
		"Symlink Support:",
		"Binary Location:",
		"Configuration:",
		"Project Configuration:",
		"Global Configuration:",
	}
	
	outputStr := string(output)
	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Doctor output missing '%s':\n%s", expected, outputStr)
		}
	}
}

func TestIntegrationDetectCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(origDir, "agentlink")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found at %s", binaryPath)
	}

	cmd := exec.Command(binaryPath, "detect")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect failed: %v\n%s", err, output)
	}

	out := string(output)
	for _, want := range []string{"Tool Detection", "Installed tools:", "known tools detected"} {
		if !strings.Contains(out, want) {
			t.Errorf("detect output missing %q:\n%s", want, out)
		}
	}
}

func TestIntegrationDetectGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	binaryPath := filepath.Join(origDir, "agentlink")
	cmd := exec.Command(binaryPath, "detect", "--generate")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect --generate failed: %v\n%s", err, output)
	}

	data, err := os.ReadFile(".agentlink.yaml")
	if err != nil {
		t.Fatalf(".agentlink.yaml not created: %v", err)
	}
	if !strings.Contains(string(data), "source: AGENTS.md") {
		t.Errorf("generated config missing source line:\n%s", data)
	}
}

func TestIntegrationScanCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake git repo with AGENTS.md
	repoDir := filepath.Join(tmpDir, "fake-repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	agentsContent := "# Agents\nshared instructions"
	if err := os.WriteFile(filepath.Join(repoDir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Repo without AGENTS.md — should be skipped
	emptyRepo := filepath.Join(tmpDir, "empty-repo")
	if err := os.MkdirAll(filepath.Join(emptyRepo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	binaryPath := filepath.Join(origDir, "agentlink")
	cmd := exec.Command(binaryPath, "scan", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scan failed: %v\n%s", err, output)
	}

	out := string(output)
	if !strings.Contains(out, "Found 2 git repositories") {
		t.Errorf("scan did not find both repos:\n%s", out)
	}

	// Verify CLAUDE.md was created as symlink in fake-repo
	claudePath := filepath.Join(repoDir, "CLAUDE.md")
	info, err := os.Lstat(claudePath)
	if err != nil {
		t.Fatalf("CLAUDE.md not created in repo with AGENTS.md: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("CLAUDE.md is not a symlink")
	}

	// Verify empty-repo was left alone
	if _, err := os.Stat(filepath.Join(emptyRepo, "CLAUDE.md")); err == nil {
		t.Error("scan created CLAUDE.md in repo without AGENTS.md")
	}
}

func TestIntegrationScanDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	repoDir := filepath.Join(tmpDir, "repo")
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.WriteFile(filepath.Join(repoDir, "AGENTS.md"), []byte("content"), 0644)

	binaryPath := filepath.Join(origDir, "agentlink")
	cmd := exec.Command(binaryPath, "scan", "--dry-run", tmpDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("scan --dry-run failed: %v", err)
	}

	// Nothing should exist after dry run
	if _, err := os.Stat(filepath.Join(repoDir, "CLAUDE.md")); err == nil {
		t.Error("dry-run created files")
	}
}

func TestIntegrationSyncBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	os.Mkdir(".git", 0755)

	binaryPath := filepath.Join(origDir, "agentlink")
	if err := exec.Command(binaryPath, "init").Run(); err != nil {
		t.Fatal(err)
	}

	sourceContent := "# Source"
	os.WriteFile("CLAUDE.md", []byte(sourceContent), 0644)

	// Create real non-empty AGENTS.md that conflicts
	existingContent := "old content to back up"
	os.WriteFile("AGENTS.md", []byte(existingContent), 0644)

	cmd := exec.Command(binaryPath, "sync", "--backup")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync --backup failed: %v\n%s", err, output)
	}

	// AGENTS.md should now be a symlink
	info, err := os.Lstat("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("AGENTS.md was not replaced with symlink")
	}

	// AGENTS.md.bak should exist with old content
	bakData, err := os.ReadFile("AGENTS.md.bak")
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if string(bakData) != existingContent {
		t.Errorf("backup content mismatch: got %q, want %q", bakData, existingContent)
	}
}

func TestIntegrationSyncBackupEmptyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	os.Mkdir(".git", 0755)

	binaryPath := filepath.Join(origDir, "agentlink")
	if err := exec.Command(binaryPath, "init").Run(); err != nil {
		t.Fatal(err)
	}

	os.WriteFile("CLAUDE.md", []byte("source"), 0644)
	// Create 0-byte conflicting file
	os.WriteFile("AGENTS.md", []byte{}, 0644)

	cmd := exec.Command(binaryPath, "sync", "--backup")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync --backup failed on empty file: %v\n%s", err, output)
	}

	// Symlink should be created
	info, err := os.Lstat("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("AGENTS.md was not replaced with symlink")
	}

	// No .bak should exist for empty file
	if _, err := os.Stat("AGENTS.md.bak"); err == nil {
		t.Error("backup created for empty file (should be skipped)")
	}

	// Warning should mention empty
	if !strings.Contains(string(output), "empty") {
		t.Errorf("expected warning about empty file:\n%s", output)
	}
}

func TestIntegrationQuietFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	os.Mkdir(".git", 0755)

	binaryPath := filepath.Join(origDir, "agentlink")
	exec.Command(binaryPath, "init").Run()
	os.WriteFile("CLAUDE.md", []byte("src"), 0644)

	cmd := exec.Command(binaryPath, "sync", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync --quiet failed: %v\n%s", err, output)
	}

	// Stdout should be empty (errors go to stderr)
	if len(strings.TrimSpace(string(output))) > 0 {
		t.Errorf("--quiet produced output: %q", output)
	}
}

func TestIntegrationHooksStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(origDir, "agentlink")

	cmd := exec.Command(binaryPath, "hooks", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hooks status failed: %v\n%s", err, output)
	}

	out := string(output)
	for _, want := range []string{"Trigger Status", "git hooks:", "zsh chpwd:", "launchd:"} {
		if !strings.Contains(out, want) {
			t.Errorf("hooks status missing %q:\n%s", want, out)
		}
	}
}

func TestIntegrationHooksInstallRequiresFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(origDir, "agentlink")

	cmd := exec.Command(binaryPath, "hooks", "install")
	if err := cmd.Run(); err == nil {
		t.Error("hooks install with no flag should fail")
	}
}

func TestIntegrationForceFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	// Use the pre-built binary
	binaryPath := filepath.Join(origDir, "agentlink")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found at %s. Make sure to run 'go build -o agentlink ./cmd/agentlink' first", binaryPath)
	}
	
	// Create .git and initialize project
	os.Mkdir(".git", 0755)
	
	cmd := exec.Command(binaryPath, "init")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	
	// Create source file
	if err := os.WriteFile("CLAUDE.md", []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create conflicting file
	if err := os.WriteFile("AGENTS.md", []byte("conflicting content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Sync without force should fail
	cmd = exec.Command(binaryPath, "sync")
	if err := cmd.Run(); err == nil {
		t.Error("Sync should have failed due to conflicting file")
	}
	
	// Sync with force should succeed
	cmd = exec.Command(binaryPath, "sync", "--force")
	if err := cmd.Run(); err != nil {
		t.Errorf("Sync with --force failed: %v", err)
	}
	
	// Verify the file was replaced with a symlink
	info, err := os.Lstat("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("File was not replaced with symlink")
	}
}