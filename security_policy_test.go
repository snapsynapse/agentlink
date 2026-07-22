package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

type releaseVersion struct {
	major int
	minor int
	patch int
}

func (v releaseVersion) newerThan(other releaseVersion) bool {
	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	return v.patch > other.patch
}

func parseReleaseVersion(value string) (releaseVersion, error) {
	var version releaseVersion
	_, err := fmt.Sscanf(strings.TrimPrefix(value, "v"), "%d.%d.%d", &version.major, &version.minor, &version.patch)
	return version, err
}

func TestSecurityPolicyMatchesLatestDeclaredReleaseFamily(t *testing.T) {
	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\] - \d{4}-\d{2}-\d{2}$`).FindSubmatch(changelog)
	if match == nil {
		t.Fatal("CHANGELOG.md has no dated release")
	}
	latest, err := parseReleaseVersion(string(match[1]))
	if err != nil {
		t.Fatal(err)
	}

	if out, tagErr := exec.Command("git", "tag", "--sort=-v:refname").Output(); tagErr == nil {
		tags := strings.Fields(string(out))
		if len(tags) > 0 {
			tagVersion, parseErr := parseReleaseVersion(tags[0])
			if parseErr != nil {
				t.Fatalf("latest tag %q is not semver-like: %v", tags[0], parseErr)
			}
			if tagVersion.newerThan(latest) {
				latest = tagVersion
			}
		}
	}

	supportedFamily := fmt.Sprintf("%d.%d.x", latest.major, latest.minor)
	unsupportedFloor := fmt.Sprintf("< %d.%d", latest.major, latest.minor)

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
