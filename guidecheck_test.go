package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGuideCheckArtifactProfile(t *testing.T) {
	data, err := os.ReadFile("docs/.well-known/assistant-guide.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 8192 {
		t.Fatalf("assistant-guide.txt is %d bytes, want <= 8192", len(data))
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 400 {
		t.Fatalf("assistant-guide.txt has %d lines, want <= 400", len(lines))
	}
	for i, b := range data {
		if b != '\n' && (b < 0x20 || b > 0x7e) {
			t.Fatalf("assistant-guide.txt has disallowed byte 0x%02x at offset %d", b, i)
		}
	}
	for i, line := range lines {
		if len(line) > 120 {
			t.Fatalf("assistant-guide.txt line %d is %d bytes, want <= 120", i+1, len(line))
		}
	}

	required := []string{
		"Before acting:",
		"[assistant-guide-metadata]",
		"profile: human-verifiable-assistant-guide",
		"canonical-url: https://agentlink.run/.well-known/assistant-guide.txt",
		"repository-url: https://github.com/snapsynapse/agentlink",
		"manifest-url: https://agentlink.run/.well-known/assistant-guide-manifest.txt",
		"Assistant invocation prompt",
		"Stop-and-ask conditions",
		"Threat model",
		"Untrusted content handling",
		"Public information safety",
		"Acceptance checklist",
	}
	text := string(data)
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("assistant-guide.txt missing required content %q", want)
		}
	}
}

func TestGuideCheckManifestMatchesGuide(t *testing.T) {
	guide, err := os.ReadFile("docs/.well-known/assistant-guide.txt")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := readGuideCheckManifest("docs/.well-known/assistant-guide-manifest.txt")
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(guide)
	gotHash := hex.EncodeToString(sum[:])
	if manifest["guide-sha256"] != gotHash {
		t.Fatalf("manifest guide-sha256 = %s, want %s", manifest["guide-sha256"], gotHash)
	}

	gotBytes, err := strconv.Atoi(manifest["guide-bytes"])
	if err != nil {
		t.Fatalf("manifest guide-bytes is invalid: %v", err)
	}
	if gotBytes != len(guide) {
		t.Fatalf("manifest guide-bytes = %d, want %d", gotBytes, len(guide))
	}

	for _, key := range []string{"guide-path", "guide-version", "immutable-release-url"} {
		if manifest[key] == "" {
			t.Fatalf("manifest missing %s", key)
		}
	}
}

func readGuideCheckManifest(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return nil, fmt.Errorf("malformed manifest line %q", line)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
