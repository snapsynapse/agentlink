# Project Context: Agentlink

## What it is

Agentlink is a small, dependency-free Go CLI that keeps AI coding-tool
instruction files in sync via plain filesystem symlinks: one real source
file (e.g. `AGENTS.md` or `CLAUDE.md`), symlinked out to every tool-specific
path a machine or repo needs (`CLAUDE.md`, `GEMINI.md`,
`.github/copilot-instructions.md`, `.cursorrules`, etc.). No codegen, no
templating, no transforms: "zero magic."

It began as a fork of [martinmose/agentlink](https://github.com/martinmose/agentlink)
by Martin Mose Facondini (MIT licensed) and is now maintained as a standalone
Snap Synapse project until its retirement on 2026-11-21. Upstream attribution
is preserved in `NOTICE`.

## Audience

- Individuals and teams running multiple AI coding assistants (Claude Code,
  Codex CLI, Gemini CLI, Cursor, Aider, etc.) who are tired of maintaining
  near-duplicate instruction files by hand
- Go developers comfortable installing via Homebrew, a pre-built binary, or
  `go install github.com/snapsynapse/agentlink/cmd/agentlink@latest`
- Readers of the project record; the tool registry is frozen at v0.7.0

## Style / tone (from README and docs)

- Direct, terse, engineer-to-engineer. Short declarative sentences, minimal
  hedging.
- Confident about scope boundaries ("Scope: instruction files only. No MCP
  `.mcp.json` or chain configs. Simple on purpose.")
- Uses concrete code blocks over abstract explanation; every claim is backed
  by a runnable example.
- Not marketing copy; no hype language. Explains *why* a design choice was
  made (e.g. why no codegen, why zero runtime dependencies).
- FAQ section answers scope-boundary questions directly and briefly.

## Key URLs

- Repo: https://github.com/snapsynapse/agentlink
- Site: https://agentlink.run/ (served from `docs/`, GitHub Pages)
- Assistant guide (GuideCheck): https://agentlink.run/.well-known/assistant-guide.txt
- Assistant guide manifest: https://agentlink.run/.well-known/assistant-guide-manifest.txt
- Homebrew tap: https://github.com/snapsynapse/homebrew-tap
- Upstream origin: https://github.com/martinmose/agentlink
- Maintainer: https://snapsynapse.com/

## Current status

- Retired 2026-11-21.
- v0.7.0 (final, released 2026-11-21): security release fixing
  GHSA-jjf5-5jgx-hg9g (High, affects 0.6.0 and earlier). Project configs confined to their
  repository, hooks install retired, hooks remove fully undoes earlier
  installs. Assistant guide 1.2.5 targets the v0.7.0 immutable release anchor.
- v0.6.0: review mitigations, refreshed registry with vendor evidence,
  generated homepage support listing, and explicit --global selection.
- v0.5.0: layered-instruction workflows, deterministic scan contracts, and
  current Go compatibility coverage
- CI (GitHub Actions, `.github/workflows/ci.yml`) verifies the Go 1.23.12
  minimum contract, current Go 1.26.8 and 1.27.1 releases, and Go 1.27.1 on
  macOS. It enforces build, unit tests, integration tests, the race detector,
  module consistency, `gofmt`, `go vet`, Staticcheck 2026.2.1, spelling,
  workflow semantics, reachable-vulnerability checks, and release-script
  syntax on every push/PR
- Distribution: Homebrew tap, pre-built binaries (darwin/linux, amd64/arm64),
  direct `go install`. The AUR package never shipped. The Homebrew formula is
  deprecated at 0.7.0
- Tool registry covers 23 AI coding tools and is frozen at v0.7.0
  (`internal/registry/tools.go`)
- Publishes a "GuideCheck" Human-Verifiable Assistant Guide targeting Level 4
  conformance for bounded, approval-gated local setup/verification work by
  AI assistants

## Standards context

Read [INTENT.md](INTENT.md) for purpose, boundaries, and exceptions, then
[repo-standards.yaml](repo-standards.yaml) for declared applicability, source
and delivery maps, and native verification expectations. The declaration
records adoption claims, not measured conformance. Reconcile it with current
source and dated evidence; review commands before execution under the active
task authority.

## Retirement

Recorded by the `retired-page` skill on 2026-11-21. The site at `docs/` is now the retired record page.

- status_word: Retired
- retired_date: 2026-11-21
- retirement_reason: from the approved legacy video script, "Why it's stopping" chapter
- youtube_id: T_oAHjqxSos ("Agentlink, for the record (2026)", 2:30; replaced YbWshl7B0PA, the 2:13 cut, on 2026-11-21)
- poster: docs/imgs/legacy-poster.jpg
- final_release: v0.7.0, 2026-11-21 (security release; supersedes v0.6.0, 2026-09-05)
- secondary pages: three guides and the supported-tools reference are kept, each with a status banner
- domain_plan: not yet stated by Sam
- mode: retire
- publish_root: docs/
- pre_retirement_commit: fbaacb344ed5018acd94daad60393221c6ff296d (restore with `git checkout fbaacb3 -- docs docs_contract_test.go`)
- Updated for the v0.7.0 final release (2026-11-21): assistant guide 1.2.5 and manifest re-anchored to v0.7.0, README, and "Current status" above. The generated support listing is unchanged; the registry did not change.
