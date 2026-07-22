# Agentlink v0.4.1

Agentlink v0.4.1 is a focused adoption-path patch following v0.4.0.

## Fixed

- `go install github.com/snapsynapse/agentlink/cmd/agentlink@v0.4.1` now produces a binary whose `agentlink --version` output is `0.4.1` instead of `dev`.
- Packaged release binaries continue to use the explicit release linker version.
- Website, `llms.txt`, assistant guide 1.2.1, manifest, and release contracts now point to v0.4.1.

## Verification

- Go 1.23.12 unit, integration, race, vet, build, static analysis, workflow, spelling, and vulnerability checks passed.
- A clean public-module install was executed from the tagged module and its version output verified.
- Hosted GuideCheck verification and release-asset checksum verification are performed after publication.

## Residuals

- GitHub Pages does not emit `X-Content-Type-Options: nosniff`, which GuideCheck reports as a non-blocking hosted warning.
- `govulncheck` reports no reachable vulnerabilities; unreachable advisories remain under the Go 1.23 floor.

See [CHANGELOG.md](CHANGELOG.md) for the complete change history.
