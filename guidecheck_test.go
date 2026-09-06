package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		"Stop and ask",
		"Install actions",
		"Threat model",
		"Untrusted content handling",
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
	if manifest["immutable-release-url"] == "pending-publication" && !strings.Contains(string(guide), "\nstatus: draft\n") {
		t.Fatal("a guide without an immutable anchor must remain draft")
	}
}

func TestGuideCheckTrustAnchorsMatch(t *testing.T) {
	pairs := [][2]string{
		{"assistant-guide.txt", "docs/.well-known/assistant-guide.txt"},
		{"assistant-guide-manifest.txt", "docs/.well-known/assistant-guide-manifest.txt"},
	}
	for _, pair := range pairs {
		reviewCopy, err := os.ReadFile(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		servedCopy, err := os.ReadFile(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if string(reviewCopy) != string(servedCopy) {
			t.Fatalf("trust-anchor drift between %s and %s", pair[0], pair[1])
		}
	}
}

func TestGuideChecksumActionRejectsTamperingAndMissingEntries(t *testing.T) {
	guide, err := os.ReadFile("assistant-guide.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, block, ok := strings.Cut(string(guide), "id: verify-checksum\n")
	if !ok {
		t.Fatal("missing verification action")
	}
	block, _, _ = strings.Cut(block, "[/action]")
	_, command, ok := strings.Cut(block, "command: sh -c '")
	if !ok {
		t.Fatal("missing executable checksum comparison")
	}
	command, _, _ = strings.Cut(command, "'\n")
	sum := sha256.Sum256([]byte("original binary"))
	for _, tc := range []struct {
		name, binary, checksums string
		wantSuccess             bool
	}{
		{"valid", "original binary", fmt.Sprintf("%x  agentlink-darwin-arm64\n", sum), true},
		{"tampered", "modified binary", fmt.Sprintf("%x  agentlink-darwin-arm64\n", sum), false},
		{"missing-entry", "original binary", fmt.Sprintf("%x  other-binary\n", sum), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for path, text := range map[string]string{"agentlink-darwin-arm64": tc.binary, "SHA256SUMS.txt": tc.checksums} {
				if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.wantSuccess {
				t.Fatalf("checksum result: %v, output: %s", err, out)
			}
		})
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
