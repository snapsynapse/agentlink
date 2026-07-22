package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `source: test.md
links:
  - link1.md
  - link2.md
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check values
	expectedSource := filepath.Join(tmpDir, "test.md")
	if cfg.Source != expectedSource {
		t.Errorf("Expected source %s, got %s", expectedSource, cfg.Source)
	}

	if len(cfg.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(cfg.Links))
	}

	expectedLink1 := filepath.Join(tmpDir, "link1.md")
	if cfg.Links[0] != expectedLink1 {
		t.Errorf("Expected first link %s, got %s", expectedLink1, cfg.Links[0])
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Source: "test.md",
				Links:  []string{"link1.md", "link2.md"},
			},
			wantErr: false,
		},
		{
			name: "empty source",
			config: Config{
				Source: "",
				Links:  []string{"link1.md"},
			},
			wantErr: true,
		},
		{
			name: "empty links",
			config: Config{
				Source: "test.md",
				Links:  []string{},
			},
			wantErr: true,
		},
		{
			name: "link is source",
			config: Config{
				Source: "AGENTS.md",
				Links:  []string{"AGENTS.md"},
			},
			wantErr: true,
		},
		{
			name: "link is syntactic alias of source",
			config: Config{
				Source: "docs/../AGENTS.md",
				Links:  []string{"./AGENTS.md"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigRejectsSourceLinkAliases(t *testing.T) {
	tmpDir := t.TempDir()
	outsideName := filepath.Base(tmpDir)

	tests := []struct {
		name   string
		source string
		link   string
	}{
		{name: "identical relative paths", source: "AGENTS.md", link: "AGENTS.md"},
		{name: "dot segment", source: "AGENTS.md", link: "./AGENTS.md"},
		{name: "parent segment", source: "AGENTS.md", link: "nested/../AGENTS.md"},
		{
			name:   "relative and absolute paths",
			source: "AGENTS.md",
			link:   filepath.Join(tmpDir, "AGENTS.md"),
		},
		{
			name:   "config directory parent alias",
			source: "AGENTS.md",
			link:   filepath.Join("..", outsideName, "AGENTS.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "-")+".yaml")
			content := "source: " + tt.source + "\nlinks:\n  - " + tt.link + "\n"
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadConfig(configPath)
			if err == nil {
				t.Fatal("LoadConfig() accepted a link resolving to the source path")
			}
			if !strings.Contains(err.Error(), "cannot be the same as source path") {
				t.Fatalf("LoadConfig() error = %q, want source/link identity error", err)
			}
		})
	}
}

func TestExpandPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tmpDir := t.TempDir()

	config := Config{
		Source: "~/test.md",
		Links:  []string{"relative.md", "~/absolute.md"},
	}

	err = config.ExpandPaths(tmpDir)
	if err != nil {
		t.Fatalf("Failed to expand paths: %v", err)
	}

	expectedSource := filepath.Join(homeDir, "test.md")
	if config.Source != expectedSource {
		t.Errorf("Expected source %s, got %s", expectedSource, config.Source)
	}

	expectedRelative := filepath.Join(tmpDir, "relative.md")
	if config.Links[0] != expectedRelative {
		t.Errorf("Expected relative link %s, got %s", expectedRelative, config.Links[0])
	}

	expectedAbsolute := filepath.Join(homeDir, "absolute.md")
	if config.Links[1] != expectedAbsolute {
		t.Errorf("Expected absolute link %s, got %s", expectedAbsolute, config.Links[1])
	}
}

func TestExpandPathsMakesPathsAbsoluteWithRelativeConfigDir(t *testing.T) {
	config := Config{
		Source: "AGENTS.md",
		Links:  []string{"./nested/../CLAUDE.md"},
	}

	if err := config.ExpandPaths("."); err != nil {
		t.Fatalf("ExpandPaths() failed: %v", err)
	}
	if !filepath.IsAbs(config.Source) {
		t.Errorf("source is not absolute: %s", config.Source)
	}
	if !filepath.IsAbs(config.Links[0]) {
		t.Errorf("link is not absolute: %s", config.Links[0])
	}
}

func TestFindConfigPath(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Test in directory without project config
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	path, isProject := FindConfigPath()
	if isProject {
		t.Error("Expected global config, got project config")
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".config", "agentlink", "config.yaml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// Test in directory with project config
	projectConfig := ".agentlink.yaml"
	os.WriteFile(projectConfig, []byte("source: test.md\nlinks: [test.md]"), 0644)

	path, isProject = FindConfigPath()
	if !isProject {
		t.Error("Expected project config, got global config")
	}

	// Resolve symlinks on both sides: on macOS, t.TempDir() returns a path
	// under /var/folders/... while os.Getwd() resolves it to /private/var/...
	expectedProjectPath := filepath.Join(tmpDir, ".agentlink.yaml")
	resolvedExpected, _ := filepath.EvalSymlinks(expectedProjectPath)
	resolvedGot, _ := filepath.EvalSymlinks(path)
	if resolvedExpected != resolvedGot {
		t.Errorf("Expected project path %s, got %s", expectedProjectPath, path)
	}
}

func TestGlobalConfigPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got := GlobalConfigPath()
	want := filepath.Join(homeDir, ".config", "agentlink", "config.yaml")
	if got != want {
		t.Errorf("GlobalConfigPath() = %s, want %s", got, want)
	}
}

func TestCreateDefaultGlobalConfigUsesAgentsSource(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "config.yaml")
	if err := CreateDefaultGlobalConfig(configPath); err != nil {
		t.Fatalf("CreateDefaultGlobalConfig() failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "source: ~/AGENTS.md") {
		t.Fatalf("global config does not use ~/AGENTS.md as source:\n%s", text)
	}
	for _, want := range []string{"~/.claude/CLAUDE.md", "~/.gemini/GEMINI.md", "~/.qwen/QWEN.md"} {
		if !strings.Contains(text, want) {
			t.Errorf("global config missing %s", want)
		}
	}
}

func TestCreateProjectConfigUsesAgentsSource(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".agentlink.yaml")
	if err := CreateProjectConfig(configPath); err != nil {
		t.Fatalf("CreateProjectConfig() failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "source: AGENTS.md") {
		t.Fatalf("project config does not use AGENTS.md as source:\n%s", text)
	}
	if strings.Contains(text, "OPENCODE.md") {
		t.Fatalf("project config contains obsolete OPENCODE.md link:\n%s", text)
	}
	if _, err := LoadConfig(configPath); err != nil {
		t.Fatalf("generated project config is not loadable: %v", err)
	}
}
