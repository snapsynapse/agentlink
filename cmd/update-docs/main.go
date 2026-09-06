// Command update-docs refreshes the generated support listing from the registry.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/snapsynapse/agentlink/internal/registry"
)

func main() {
	check := flag.Bool("check", false, "check for drift without writing")
	flag.Parse()
	const path = "docs/index.html"
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	const begin, end = "<!-- BEGIN GENERATED TOOLS -->\n", "<!-- END GENERATED TOOLS -->"
	before, rest, ok := strings.Cut(string(data), begin)
	if !ok {
		fail(fmt.Errorf("missing start marker"))
	}
	_, after, ok := strings.Cut(rest, end)
	if !ok {
		fail(fmt.Errorf("missing end marker"))
	}
	want := before + begin + registry.Documentation() + end + after
	if want == string(data) {
		return
	}
	if *check {
		fail(fmt.Errorf("support listing drift: run go run ./cmd/update-docs"))
	}
	if err := os.WriteFile(path, []byte(want), 0644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
