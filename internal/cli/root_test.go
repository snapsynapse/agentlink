package cli

import "testing"

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name          string
		linkerVersion string
		moduleVersion string
		want          string
	}{
		{name: "release linker flag wins", linkerVersion: "0.4.1", moduleVersion: "v0.4.1", want: "0.4.1"},
		{name: "go install module version", linkerVersion: "dev", moduleVersion: "v0.4.1", want: "0.4.1"},
		{name: "development build", linkerVersion: "dev", moduleVersion: "(devel)", want: "dev"},
		{name: "missing metadata", want: "dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := versionFromBuildInfo(test.linkerVersion, test.moduleVersion); got != test.want {
				t.Fatalf("versionFromBuildInfo(%q, %q) = %q, want %q", test.linkerVersion, test.moduleVersion, got, test.want)
			}
		})
	}
}
