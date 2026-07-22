# Agentlink v0.4.0

Agentlink v0.4.0 makes `AGENTS.md` the default source across generated project and global configurations, hardens filesystem and hook behavior, and improves installation and release reliability.

## Highlights

- Canonical Go module identity at `github.com/snapsynapse/agentlink`, enabling `go install github.com/snapsynapse/agentlink/cmd/agentlink@v0.4.0`.
- Generated project and global configurations now default to `AGENTS.md` and current registry paths.
- Safer source and destination alias validation, collision-safe backups, executable managed hooks, aggregate hook-removal errors, and nonzero scan and clean failures.
- Git and zsh hook updates preserve user content and permissions; zsh hooks resolve repository roots for subdirectories and worktrees.
- Copyable configuration examples and contract tests for published commands, files, and canonical install instructions.
- Linux and macOS CI with unit, integration, race, vet, build, module, formatting, static analysis, spelling, workflow, shell, and reachable-vulnerability checks.
- Updated website, `llms.txt`, GuideCheck 0.7.1 assistant guide 1.2.0, security policy, release tooling, and documentation.
- Removed unused internal code and the unused Flox environment.

## Verification

- Go 1.23.12 unit and integration tests passed.
- Race detector, `go vet`, `staticcheck`, formatting, spelling, and module consistency passed.
- `actionlint` v1.7.7 and shell syntax checks passed.
- `govulncheck` v1.1.4 found no reachable vulnerabilities.
- Assistant-guide byte profile, SHA-256 manifest, and repository contract tests passed.

## Residuals

- `govulncheck` reports advisories in imported or required packages under the Go 1.23 floor, but no affected symbols are reachable from Agentlink.
- GuideCheck Level 4 remains a guide-file conformance target, not a safety certification.
- The AUR package remains planned and is not part of this release.

See [CHANGELOG.md](CHANGELOG.md) for the complete change list.
