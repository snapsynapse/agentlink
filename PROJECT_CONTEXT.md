# Project Context: Agentlink

## What it is

Agentlink is a small, dependency-free Go CLI that keeps AI coding-tool
instruction files in sync via plain filesystem symlinks — one real source
file (e.g. `AGENTS.md` or `CLAUDE.md`), symlinked out to every tool-specific
path a machine or repo needs (`CLAUDE.md`, `GEMINI.md`,
`.github/copilot-instructions.md`, `.cursorrules`, etc.). No codegen, no
templating, no transforms — "zero magic."

It began as a fork of [martinmose/agentlink](https://github.com/martinmose/agentlink)
by Martin Mose Facondini (MIT licensed) and is now maintained as a standalone
Snap Synapse project. Upstream attribution is preserved in `NOTICE`.

## Audience

- Individuals and teams running multiple AI coding assistants (Claude Code,
  Codex CLI, Gemini CLI, Cursor, Aider, etc.) who are tired of maintaining
  near-duplicate instruction files by hand
- Go developers comfortable installing via Homebrew, a pre-built binary, or
  `go install github.com/snapsynapse/agentlink/cmd/agentlink@latest`
- Contributors extending the tool registry (`internal/registry/tools.go`)
  when new AI coding tools appear

## Style / tone (from README and docs)

- Direct, terse, engineer-to-engineer. Short declarative sentences, minimal
  hedging.
- Confident about scope boundaries ("Scope: instruction files only. No MCP
  `.mcp.json` or chain configs. Simple on purpose.")
- Uses concrete code blocks over abstract explanation; every claim is backed
  by a runnable example.
- Not marketing copy — no hype language. Explains *why* a design choice was
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

- Actively maintained; v0.4.1 packages the 2026-07-21 release-readiness,
  filesystem-safety, hook-reliability, adoption, and CI work
- CI (GitHub Actions, `.github/workflows/ci.yml`) enforces build, unit tests,
  integration tests, the race detector, module consistency, `gofmt`, `go vet`,
  `staticcheck`, spelling, workflow semantics, reachable-vulnerability checks,
  and release-script syntax on every push/PR
- Distribution: Homebrew tap, pre-built binaries (darwin/linux, amd64/arm64),
  direct `go install`; AUR package listed as planned/not yet shipped
- Tool registry covers ~23 AI coding tools; new tools are added via PR to
  `internal/registry/tools.go`
- Publishes a "GuideCheck" Human-Verifiable Assistant Guide targeting Level 4
  conformance for bounded, approval-gated local setup/verification work by
  AI assistants
