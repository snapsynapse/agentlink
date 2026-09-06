package main

import (
	"html"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/snapsynapse/agentlink/internal/config"
	"github.com/snapsynapse/agentlink/internal/registry"
	"gopkg.in/yaml.v3"
)

func TestOpenGraphImagesFollowPortfolioDimensions(t *testing.T) {
	const wantWidth, wantHeight = 1200, 630
	var first []byte
	for _, path := range []string{"imgs/og.png", "docs/imgs/og.png"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		config, _, err := image.DecodeConfig(strings.NewReader(string(data)))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if config.Width != wantWidth || config.Height != wantHeight {
			t.Errorf("%s dimensions = %dx%d, want %dx%d", path, config.Width, config.Height, wantWidth, wantHeight)
		}
		if first == nil {
			first = data
		} else if string(data) != string(first) {
			t.Errorf("%s does not match imgs/og.png", path)
		}
	}
}

func TestAIPostureTrustAnchorsMatch(t *testing.T) {
	const postureURL = "https://agentlink.run/.well-known/aiposture"
	root, err := os.ReadFile("aiposture")
	if err != nil {
		t.Fatal(err)
	}
	served, err := os.ReadFile("docs/.well-known/aiposture")
	if err != nil {
		t.Fatal(err)
	}
	if string(root) != string(served) {
		t.Fatal("AI Posture trust-anchor copies differ")
	}
	for _, want := range []string{"AI Posture: Agentlink", "Data residency: local filesystem only", postureURL} {
		if !strings.Contains(string(root), want) {
			t.Errorf("aiposture missing %q", want)
		}
	}
	for _, path := range []string{"README.md", "docs/index.html", "docs/llms.txt"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), postureURL) {
			t.Errorf("%s does not link to AI Posture", path)
		}
	}
}

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

func TestPublishedGoCompatibilityContract(t *testing.T) {
	for _, path := range []string{"README.md", "docs/index.html", "docs/llms.txt"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"Go 1.23", "1.26.8", "1.27.1"} {
			if !strings.Contains(string(data), want) {
				t.Errorf("%s missing Go compatibility marker %q", path, want)
			}
		}
	}
}

func TestLandingPageBylineIncludesLastUpdatedDate(t *testing.T) {
	data, err := os.ReadFile("docs/index.html")
	if err != nil {
		t.Fatal(err)
	}
	const want = "Last updated <time datetime=\"2026-09-05\">September 5, 2026</time>"
	if !strings.Contains(string(data), want) {
		t.Errorf("landing page byline missing %q", want)
	}
}

func TestPublishedLayeringAndNestedScanContract(t *testing.T) {
	const jonathanGuide = "https://limitededitionjonathan.substack.com/p/i-wrote-this-agentsmd-guide-for-the"
	const attribution = "Limited Edition Jonathan"
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
		if !strings.Contains(string(data), attribution) {
			t.Errorf("%s missing attribution %q", path, attribution)
		}
		if strings.Contains(string(data), "Jonathan"+" Whitaker") {
			t.Errorf("%s has superseded attribution", path)
		}
	}
}

func TestHomepageSupportListingMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile("docs/index.html")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(data), "<!-- BEGIN GENERATED TOOLS -->\n")
	if !ok {
		t.Fatal("missing generated listing")
	}
	listing, _, ok := strings.Cut(rest, "<!-- END GENERATED TOOLS -->")
	if !ok || listing != registry.Documentation() {
		t.Fatal("homepage registry drift: run go run ./cmd/update-docs")
	}
	for _, path := range []string{"README.md", "docs/reference/supported-tools/index.html"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "#supported-tools") {
			t.Errorf("%s must link to canonical support listing", path)
		}
	}
}

func TestWebsiteConfigurationExamplesAreValidYAML(t *testing.T) {
	paths, err := filepath.Glob("docs/guides/*/index.html")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "docs/index.html")
	blocks := regexp.MustCompile(`(?s)<pre>(?:<code>)?(.*?)(?:</code>)?</pre>`)
	count := 0
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range blocks.FindAllStringSubmatch(string(data), -1) {
			text := html.UnescapeString(match[1])
			if !strings.Contains(text, "\nsource:") {
				continue
			}
			count++
			var example struct {
				Source string   `yaml:"source"`
				Links  []string `yaml:"links"`
			}
			decoder := yaml.NewDecoder(strings.NewReader(text))
			decoder.KnownFields(true)
			if err := decoder.Decode(&example); err != nil {
				t.Errorf("%s has invalid config example: %v", path, err)
			}
			if example.Source == "" || len(example.Links) == 0 {
				t.Errorf("%s has incomplete config example", path)
			}
		}
	}
	if count < 5 {
		t.Fatalf("expected at least five configuration examples, found %d", count)
	}
}

func TestReleaseSurfacesUseCurrentVersion(t *testing.T) {
	const version = "v0.6.0"
	wants := map[string][]string{
		"CHANGELOG.md": {"## [0.6.0] - 2026-09-05"},
		"SECURITY.md":  {"| 0.6.x   | Yes"},
		"docs/index.html": {
			"Agentlink v0.6.0",
			"/releases/tag/v0.6.0",
			"/releases/download/v0.6.0/agentlink-darwin-arm64",
		},
		"docs/llms.txt": {"Current release: v0.6.0."},
		"docs/.well-known/assistant-guide.txt": {
			"guide-version: 1.2.4",
			"go install github.com/snapsynapse/agentlink/cmd/agentlink@v0.6.0",
			"/releases/download/v0.6.0/agentlink-darwin-arm64",
		},
		"docs/.well-known/assistant-guide-manifest.txt": {
			"immutable-release-url: https://github.com/snapsynapse/agentlink/blob/v0.6.0/docs/.well-known/assistant-guide.txt",
		},
		"assistant-guide.txt": {
			"guide-version: 1.2.4",
			"applies-to: agentlink >=0.6.0",
		},
		"assistant-guide-manifest.txt": {
			"immutable-release-url: https://github.com/snapsynapse/agentlink/blob/v0.6.0/docs/.well-known/assistant-guide.txt",
		},
		"RELEASE_NOTES-0.6.0.md": {"# Agentlink v0.6.0"},
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
