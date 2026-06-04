package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSecurityPolicyMatchesLatestLocalTagFamily(t *testing.T) {
	out, err := exec.Command("git", "tag", "--sort=-v:refname").Output()
	if err != nil {
		t.Skipf("cannot inspect local git tags: %v", err)
	}
	tags := strings.Fields(string(out))
	if len(tags) == 0 {
		t.Skip("no local git tags available")
	}

	latest := strings.TrimPrefix(tags[0], "v")
	parts := strings.Split(latest, ".")
	if len(parts) < 2 {
		t.Fatalf("latest tag %q is not semver-like", tags[0])
	}
	supportedFamily := parts[0] + "." + parts[1] + ".x"
	unsupportedFloor := "< " + parts[0] + "." + parts[1]

	data, err := os.ReadFile("SECURITY.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "| "+supportedFamily+"   | Yes") {
		t.Fatalf("SECURITY.md does not list latest tag family %s as supported", supportedFamily)
	}
	if !strings.Contains(text, "| "+unsupportedFloor+"   | No") {
		t.Fatalf("SECURITY.md does not list %s as unsupported", unsupportedFloor)
	}
}
