package cli

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellQuoteWrapsPathsSafely(t *testing.T) {
	got := shellQuote("/tmp/Agent Link/bin/agent'link")
	want := "'/tmp/Agent Link/bin/agent'\\''link'"
	if got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}

func TestGitHookContentQuotesBinaryPath(t *testing.T) {
	content := gitHookContent("/tmp/Agent Link/bin/agentlink")
	if !strings.Contains(content, "'/tmp/Agent Link/bin/agentlink' sync --quiet") {
		t.Fatalf("gitHookContent() did not quote binary path:\n%s", content)
	}
}

func TestZshHookContentQuotesBinaryPath(t *testing.T) {
	content := zshHookContent("/tmp/Agent Link/bin/agentlink")
	if !strings.Contains(content, "'/tmp/Agent Link/bin/agentlink' sync --quiet") {
		t.Fatalf("zshHookContent() did not quote binary path:\n%s", content)
	}
}

func TestZshHookContentSupportsSubdirectoriesAndWorktrees(t *testing.T) {
	content := zshHookContent("/tmp/agentlink")
	for _, want := range []string{
		"git rev-parse --show-toplevel",
		`[ -f "$agentlink_root/.agentlink.yaml" ]`,
		`cd "$agentlink_root"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("zshHookContent() missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "[ -d .git ]") {
		t.Fatalf("zshHookContent() still requires a .git directory:\n%s", content)
	}
}

func TestLaunchdPlistContentEscapesBinaryPath(t *testing.T) {
	binaryPath := "/tmp/Agent Link/bin/agent'link & <test>"
	content := launchdPlistContent(binaryPath)

	if !strings.Contains(content, "/tmp/Agent Link/bin/agent&#39;link &amp; &lt;test&gt;") {
		t.Fatalf("launchdPlistContent() did not XML-escape binary path:\n%s", content)
	}

	args, err := parseLaunchdProgramArguments(content)
	if err != nil {
		t.Fatalf("launchdPlistContent() produced invalid XML: %v\n%s", err, content)
	}

	want := []string{binaryPath, "sync", "--quiet"}
	if len(args) != len(want) {
		t.Fatalf("ProgramArguments = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("ProgramArguments[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestRemoveMarkedSectionPreservesExecutableMode(t *testing.T) {
	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "post-merge")
	content := "#!/bin/sh\n" +
		"echo before\n" +
		gitHookMarkerStart + "\n" +
		"agentlink sync --quiet\n" +
		gitHookMarkerEnd + "\n" +
		"echo after\n"
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	removed, err := removeMarkedSection(hookPath)
	if err != nil {
		t.Fatalf("removeMarkedSection() failed: %v", err)
	}
	if !removed {
		t.Fatal("removeMarkedSection() removed = false, want true")
	}

	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, gitHookMarkerStart) || strings.Contains(got, "agentlink sync") {
		t.Fatalf("marked section was not removed:\n%s", got)
	}
	if !strings.Contains(got, "echo before") || !strings.Contains(got, "echo after") {
		t.Fatalf("unmanaged hook content was not preserved:\n%s", got)
	}
}

func TestRemoveMarkedSectionMissingFileIsNoop(t *testing.T) {
	removed, err := removeMarkedSection(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("removeMarkedSection() returned unexpected error: %v", err)
	}
	if removed {
		t.Fatal("removeMarkedSection() removed = true, want false")
	}
}

func TestInstallGitHooksRejectsRelativeHooksPath(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	if err := installGitHooksWithDir("relative/hooks", "/tmp/agentlink"); err == nil {
		t.Fatal("installGitHooksWithDir() succeeded for relative hooks path, want error")
	}
}

func TestAppendOrCreateHookMakesExistingHookExecutable(t *testing.T) {
	hookPath := filepath.Join(t.TempDir(), "post-merge")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := appendOrCreateHook(hookPath, gitHookContent("/tmp/agentlink")); err != nil {
		t.Fatalf("appendOrCreateHook() failed: %v", err)
	}
	assertUserExecutable(t, hookPath)
}

func TestAppendOrCreateHookRepairsModeWhenAlreadyInstalled(t *testing.T) {
	hookPath := filepath.Join(t.TempDir(), "post-merge")
	oldContent := gitHookContent("/tmp/old-agentlink")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho before\n"+oldContent+"echo after\n"), 0644); err != nil {
		t.Fatal(err)
	}
	newContent := gitHookContent("/tmp/new-agentlink")
	if err := appendOrCreateHook(hookPath, newContent); err != nil {
		t.Fatalf("appendOrCreateHook() failed: %v", err)
	}
	assertUserExecutable(t, hookPath)
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(got, "/tmp/old-agentlink") || !strings.Contains(got, "/tmp/new-agentlink") {
		t.Fatalf("managed hook section was not updated:\n%s", got)
	}
	if !strings.Contains(got, "echo before") || !strings.Contains(got, "echo after") {
		t.Fatalf("unmanaged hook content was not preserved:\n%s", got)
	}
}

func TestRewriteMarkedSectionRejectsIncompleteBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hook")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+gitHookMarkerStart+"\nold\n"), 0755); err != nil {
		t.Fatal(err)
	}
	updated, err := rewriteMarkedSection(path, gitHookContent("/tmp/agentlink"))
	if err == nil {
		t.Fatal("rewriteMarkedSection() error = nil, want incomplete marker error")
	}
	if updated {
		t.Fatal("rewriteMarkedSection() updated incomplete marker block")
	}
}

func assertUserExecutable(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Fatalf("mode = %v, want user-executable", info.Mode().Perm())
	}
}

func parseLaunchdProgramArguments(content string) ([]string, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	var lastKey string
	var inProgramArguments bool
	var args []string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "key":
				var key string
				if err := decoder.DecodeElement(&key, &t); err != nil {
					return nil, err
				}
				lastKey = key
			case "array":
				inProgramArguments = lastKey == "ProgramArguments"
			case "string":
				if inProgramArguments {
					var value string
					if err := decoder.DecodeElement(&value, &t); err != nil {
						return nil, err
					}
					args = append(args, value)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "array" && inProgramArguments {
				inProgramArguments = false
			}
		}
	}
	return args, nil
}
