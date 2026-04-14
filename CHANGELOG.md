# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-04-14

First tagged release of the Snap Synapse fork of
[martinmose/agentlink](https://github.com/martinmose/agentlink). Preserves
upstream core (init, sync, check, clean, doctor) and adds detection,
scanning, automatic triggers, safe replacement, and global-config support.

### Added

- `agentlink detect` — scan the local system for installed AI coding tools
  via a registry of ~20 tools (Claude Code, Codex, Gemini CLI, Cursor,
  Aider, Amp, Factory, Goose, OpenCode, Windsurf, Zed, and more). `--generate`
  emits a `.agentlink.yaml` scaffold from what was found.
- `agentlink scan [dir]` — walk a directory tree, find git repos containing
  AGENTS.md, and wire tool-specific symlinks (CLAUDE.md, GEMINI.md, etc.)
  in each. Default scan root: `~/Git`. Supports `--dry-run`.
- `agentlink hooks install|remove|status` — install automatic sync triggers:
  git hooks (post-checkout, post-merge via `core.hooksPath`), zsh `chpwd`
  hook, and macOS launchd 60-minute heartbeat. All injected content wrapped
  in `# >>> agentlink >>>` / `# <<< agentlink <<<` markers for clean removal.
- `sync --backup` — back up existing target files to `<name>.bak` (or
  `<name>.<timestamp>.bak` if `.bak` exists) before replacing with a symlink.
  Empty files (0 bytes) are skipped with a warning — no `.bak` created.
- Global config at `~/.config/agentlink/config.yaml` — sync AI instruction
  files across the home directory, not just within a single repo.
- Tool registry (`internal/registry/`) listing known AI coding tools with
  their global config path, repo file path, and `reads_agents_md` flag.
  Unit-tested.
- Integration tests covering detect, scan, sync backup behavior, empty-file
  skip, quiet flag, hooks install/status, and the existing force-flag path.
- `--quiet` / `-q` flag on `sync` to suppress non-error output.

### Changed

- README rewritten to lead with the problem ("sync one AGENTS.md to every
  AI coding tool") rather than the tool category. Added prior-art table
  of ten alternative projects.
- CI workflow badge points at the snapsynapse fork while upstream PR #2
  is in review.

### Fixed

- `TestFindConfigPath` now passes on macOS. The test previously failed
  because `t.TempDir()` returns paths under `/var/folders/...` while
  `os.Getwd()` resolves them to `/private/var/...` via the OS-level
  symlink. Fix: resolve symlinks on both sides before comparing.
- `staticcheck` issues in upstream code cleaned up as a pre-fork hygiene
  pass.

### Attribution

Fork of [martinmose/agentlink](https://github.com/martinmose/agentlink)
by Martin Mose Facondini (MIT). Fork additions offered back upstream in
[martinmose/agentlink#2](https://github.com/martinmose/agentlink/pull/2).
See [NOTICE](NOTICE) for the full fork provenance.

[Unreleased]: https://github.com/snapsynapse/agentlink/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/snapsynapse/agentlink/releases/tag/v0.1.0
