package cli

import (
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
