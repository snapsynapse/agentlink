package main

import (
	"os"
	"strings"
	"testing"

	"github.com/snapsynapse/agentlink/internal/config"
	"github.com/snapsynapse/agentlink/internal/registry"
)

func TestPublishedHookExamplesIncludeRequiredSelector(t *testing.T) {
	for _, path := range []string{"README.md", "docs/index.html"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, want := range []string{"agentlink hooks install --all", "agentlink hooks remove --all"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s missing executable example %q", path, want)
			}
		}
	}
}

func TestPublishedGoInstallUsesCanonicalModule(t *testing.T) {
	const command = "go install github.com/snapsynapse/agentlink/cmd/agentlink@latest"
	for _, path := range []string{"README.md", "docs/index.html", "docs/llms.txt"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), command) {
			t.Errorf("%s missing canonical install command %q", path, command)
		}
	}
}

func TestPublishedLayeringAndNestedScanContract(t *testing.T) {
	const jonathanGuide = "https://limitededitionjonathan.substack.com/p/i-wrote-this-agentsmd-guide-for-the"
	wants := map[string][]string{
		"README.md": {
			"Identical aliases vs. layered wrappers",
			"agentlink detect --generate --prefer-native",
			"agentlink scan --nested",
			jonathanGuide,
		},
		"docs/index.html": {
			"Layered Claude wrapper",
			"agentlink detect --generate --prefer-native",
			"agentlink scan ~/Git --nested",
			jonathanGuide,
		},
		"docs/llms.txt": {
			"layered wrapper:",
			"detect --generate --prefer-native",
			"opt-in --nested mode",
			jonathanGuide,
		},
	}
	for path, required := range wants {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range required {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s missing layering contract %q", path, want)
			}
		}
	}
}

func TestReadmeSupportedToolsTableMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "### Supported tools")
	end := strings.Index(text, "## Repo Scanning")
	if start == -1 || end == -1 || end <= start {
		t.Fatal("README supported-tools table boundaries not found")
	}

	type documentedTool struct {
		global      string
		repo        string
		integration string
	}
	documented := make(map[string]documentedTool)
	for _, line := range strings.Split(text[start:end], "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) != 4 {
			continue
		}
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if cells[0] == "Tool" || strings.HasPrefix(cells[0], "---") {
			continue
		}
		documented[cells[0]] = documentedTool{
			global:      cells[1],
			repo:        cells[2],
			integration: cells[3],
		}
	}

	tools := registry.All()
	if len(documented) != len(tools) {
		t.Errorf("README documents %d tools, registry contains %d", len(documented), len(tools))
	}
	for _, tool := range tools {
		got, ok := documented[tool.Name]
		if !ok {
			t.Errorf("README supported-tools table is missing registry tool %q", tool.Name)
			continue
		}
		wantGlobal := tool.GlobalConfigPath
		if wantGlobal == "" {
			wantGlobal = "--"
		}
		wantRepo := tool.RepoFileName
		if wantRepo == "" {
			wantRepo = "--"
		}
		wantIntegration := map[registry.AgentsMDIntegration]string{
			registry.IntegrationNative:       "Native",
			registry.IntegrationConfigurable: "Configurable",
			registry.IntegrationImport:       "Import from real " + tool.RepoFileName,
			registry.IntegrationSymlink:      "Symlink",
			registry.IntegrationUnsupported:  "Unsupported",
		}[tool.AgentsMDIntegration()]
		if tool.AgentsMDIntegration() == registry.IntegrationNative && tool.RepoFileName == "" && tool.GlobalConfigPath != "" {
			wantIntegration = "Native (global)"
		}
		if got.global != wantGlobal || got.repo != wantRepo || got.integration != wantIntegration {
			t.Errorf("README row for %q = {%q, %q, %q}, want {%q, %q, %q}", tool.Name, got.global, got.repo, got.integration, wantGlobal, wantRepo, wantIntegration)
		}
		delete(documented, tool.Name)
	}
	for name := range documented {
		t.Errorf("README supported-tools table contains unknown tool %q", name)
	}
}

func TestReleaseSurfacesUseCurrentVersion(t *testing.T) {
	const version = "v0.4.2"
	wants := map[string][]string{
		"CHANGELOG.md": {"## [0.4.2] - 2026-08-20"},
		"SECURITY.md":  {"| 0.4.x   | Yes"},
		"docs/index.html": {
			"Agentlink v0.4.2",
			"/releases/tag/v0.4.2",
			"/releases/download/v0.4.2/agentlink-darwin-arm64",
		},
		"docs/llms.txt": {"Current release: v0.4.2."},
		"docs/.well-known/assistant-guide.txt": {
			"guide-version: 1.2.2",
			"go install github.com/snapsynapse/agentlink/cmd/agentlink@v0.4.2",
			"/releases/download/v0.4.2/agentlink-darwin-arm64",
		},
		"docs/.well-known/assistant-guide-manifest.txt": {
			"immutable-release-url: https://github.com/snapsynapse/agentlink/blob/v0.4.2/",
		},
		"assistant-guide.txt": {
			"guide-version: 1.2.2",
			"applies-to: agentlink >=0.4.2",
		},
		"assistant-guide-manifest.txt": {
			"immutable-release-url: https://github.com/snapsynapse/agentlink/blob/v0.4.2/",
		},
		"RELEASE_NOTES-0.4.2.md": {"# Agentlink v0.4.2"},
	}

	for path, required := range wants {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range required {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s missing %s release marker %q", path, version, want)
			}
		}
	}
}

func TestLLMSKeyFileReferencesExist(t *testing.T) {
	for _, path := range []string{
		"examples/project.agentlink.yaml",
		"examples/global.agentlink.yaml",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("docs/llms.txt references missing file %s: %v", path, err)
		}
		if _, err := config.LoadConfig(path); err != nil {
			t.Errorf("example config %s is not loadable: %v", path, err)
		}
	}
}

func TestStandaloneIssueTemplateDoesNotClaimForkStatus(t *testing.T) {
	data, err := os.ReadFile(".github/ISSUE_TEMPLATE/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "This is a fork") {
		t.Fatal("issue template still describes the standalone repository as a fork")
	}
}
