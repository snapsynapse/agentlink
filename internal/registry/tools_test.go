package registry

import (
	"strings"
	"testing"
)

func TestAllReturnsNonEmpty(t *testing.T) {
	tools := All()
	if len(tools) == 0 {
		t.Fatal("All() returned empty slice")
	}
	if len(tools) < 15 {
		t.Errorf("expected at least 15 tools, got %d", len(tools))
	}
}

func TestAllToolsHaveName(t *testing.T) {
	for i, tool := range All() {
		if tool.Name == "" {
			t.Errorf("tool at index %d has empty Name", i)
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty Description", tool.Name)
		}
	}
}

func TestAllToolsHaveUniqueName(t *testing.T) {
	seen := make(map[string]bool)
	for _, tool := range All() {
		if seen[tool.Name] {
			t.Errorf("duplicate tool name: %s", tool.Name)
		}
		seen[tool.Name] = true
	}
}

func TestGlobalConfigPathsUseTildePrefix(t *testing.T) {
	for _, tool := range All() {
		if tool.GlobalConfigPath == "" {
			continue
		}
		if !strings.HasPrefix(tool.GlobalConfigPath, "~/") {
			t.Errorf("tool %q GlobalConfigPath %q should start with ~/", tool.Name, tool.GlobalConfigPath)
		}
	}
}

func TestDetectAllDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DetectAll() panicked: %v", r)
		}
	}()
	_ = DetectAll()
}

func TestToolsWithGlobalConfigFilters(t *testing.T) {
	filtered := ToolsWithGlobalConfig()
	if len(filtered) == 0 {
		t.Fatal("ToolsWithGlobalConfig() returned empty slice")
	}
	for _, tool := range filtered {
		if tool.GlobalConfigPath == "" {
			t.Errorf("tool %q has empty GlobalConfigPath but was returned by ToolsWithGlobalConfig()", tool.Name)
		}
	}
	if len(filtered) >= len(All()) {
		t.Error("ToolsWithGlobalConfig() did not filter anything")
	}
}

func TestToolsReadingAgentsMDFilters(t *testing.T) {
	filtered := ToolsReadingAgentsMD()
	if len(filtered) == 0 {
		t.Fatal("ToolsReadingAgentsMD() returned empty slice")
	}
	for _, tool := range filtered {
		if !tool.ReadsAgentsMD {
			t.Errorf("tool %q has ReadsAgentsMD=false but was returned by ToolsReadingAgentsMD()", tool.Name)
		}
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		input   string
		homeDir string
		want    string
	}{
		{"~/foo", "/home/user", "/home/user/foo"},
		{"~/.claude/CLAUDE.md", "/Users/test", "/Users/test/.claude/CLAUDE.md"},
		{"/absolute/path", "/home/user", "/absolute/path"},
		{"relative/path", "/home/user", "relative/path"},
		{"~", "/home/user", "~"}, // only ~/ prefix is expanded
	}
	for _, tc := range tests {
		got := expandHome(tc.input, tc.homeDir)
		if got != tc.want {
			t.Errorf("expandHome(%q, %q) = %q, want %q", tc.input, tc.homeDir, got, tc.want)
		}
	}
}

func TestDetectToolWithKnownCommand(t *testing.T) {
	// sh is guaranteed to exist on any Unix system
	tool := Tool{
		Name:           "TestTool",
		DetectCommands: []string{"sh"},
	}
	got := detectTool(tool)
	if got == nil {
		t.Fatal("expected detection for 'sh' command, got nil")
	}
	if got.Method != "command" {
		t.Errorf("expected Method='command', got %q", got.Method)
	}
}

func TestDetectToolWithMissingCommand(t *testing.T) {
	tool := Tool{
		Name:           "TestTool",
		DetectCommands: []string{"definitely-not-a-real-binary-xyz123"},
	}
	if got := detectTool(tool); got != nil {
		t.Errorf("expected nil for missing command, got %+v", got)
	}
}
