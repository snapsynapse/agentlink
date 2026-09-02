# Agentlink v0.5.0

Agentlink v0.5.0 adds explicit support for layered instruction-file workflows
while maintaining its narrow filesystem scope.

## What changed

- `detect --generate --prefer-native` now favors native, configurable, and
  import-based integrations before symlink aliases.
- `scan --nested` can create documented nested-capable aliases beside nested
  `AGENTS.md` files, while unknown integrations stay root-only.
- Deterministic contracts cover compiled-CLI topology, relative targets,
  explicit configuration, wrapper preservation, idempotence, dry runs, and
  registry and documentation parity.
- CI covers the Go 1.23.12 minimum contract, Go 1.26.8, Go 1.27.1, and Go
  1.27.1 on macOS. Its stable quality lane uses Staticcheck 2026.2.1 and
  Govulncheck 1.7.0.
- A byte-identical root and served AI Posture declaration records the CLI's
  local-only data and execution boundaries. OpenGraph images now use the
  portfolio-standard 1200×630 dimensions.

## Runtime compatibility

Go 1.23 or newer remains required for source installation. The CLI remains
dependency-free at runtime. Existing configurations retain their behavior.

## Verification

The release gate runs unit, race, vet, integration, formatting, static
analysis, spelling, vulnerability, action-semantic, search, accessibility,
deterministic cross-build, checksum, host-binary version, clean consumer, and
production checks. Published binaries target darwin and linux on amd64 and
arm64.
