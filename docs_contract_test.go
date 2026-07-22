package main

import (
	"os"
	"strings"
	"testing"

	"github.com/snapsynapse/agentlink/internal/config"
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

func TestReleaseSurfacesUseCurrentVersion(t *testing.T) {
	const version = "v0.4.1"
	wants := map[string][]string{
		"CHANGELOG.md": {"## [0.4.1] - 2026-07-21"},
		"SECURITY.md":  {"| 0.4.x   | Yes"},
		"docs/index.html": {
			"Agentlink v0.4.1",
			"/releases/tag/v0.4.1",
			"/releases/download/v0.4.1/agentlink-darwin-arm64",
		},
		"docs/llms.txt": {"Current release: v0.4.1."},
		"docs/.well-known/assistant-guide.txt": {
			"guide-version: 1.2.1",
			"go install github.com/snapsynapse/agentlink/cmd/agentlink@v0.4.1",
			"/releases/download/v0.4.1/agentlink-darwin-arm64",
		},
		"docs/.well-known/assistant-guide-manifest.txt": {
			"immutable-release-url: https://github.com/snapsynapse/agentlink/blob/v0.4.1/",
		},
		"RELEASE_NOTES-0.4.1.md": {"# Agentlink v0.4.1"},
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
