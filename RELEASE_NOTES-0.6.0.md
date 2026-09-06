# Agentlink v0.6.0

This release protects instruction sources before replacement, makes hook removal honor dry-run, and refreshes supported-tool integrations.

- Source aliases and backing files cannot be removed by backup or force. Relative links resolve correctly through symlinked parent directories.
- Broken or misdirected links now require explicit `--force`. `--backup` authorizes regular-file replacement only.
- Dangling project configuration fails instead of silently selecting another scope. `sync`, `check`, and `clean` accept `--global` for explicit global selection.
- Git-hook installation refuses unsupported interpreters and preserves existing shell bodies. `hooks --all` includes launchd only on macOS.
- All 23 integrations have dated vendor references and caveats. The homepage consolidates operating-system support and the generated tool listing.
- Website examples separate YAML and shell commands; mobile navigation stays available. Accessibility CI scans every sitemap page.
- Assistant guide 1.2.4 performs executable checksum verification and uses this release as its immutable anchor.

## Upgrade notes

Inspect conflicting links before using force. Existing configuration files are not migrated automatically. Current Qwen and Goose clients can load AGENTS.md directly; old aliases may duplicate instructions. Continue settings YAML must remain real: remove settings paths from Agentlink's links list and use `.continue/rules/AGENTS.md` for instruction aliases.

Binaries support macOS and Linux on ARM64 and AMD64. Native Windows is not a release target. Client-specific instruction loading remains governed by the vendor's documented behavior.

## Verification

Release preparation runs unit, race, integration, vet, deterministic cross-build, checksum, and host-version checks. Publication follows exact-commit CI and accessibility checks. Download verification compares all binary hashes and checks the host binary version.
