# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Four tools in the detection registry, bringing it to 23: Crush
  (AGENTS.md), Kiro (AGENTS.md), Qwen Code (QWEN.md, `~/.qwen/QWEN.md`),
  and OpenClaw (global `~/.openclaw/workspace/AGENTS.md`, no repo file).

### Changed

- The repository is now a standalone GitHub repository instead of a GitHub
  fork relationship, so it appears in GitHub search. Attribution to the
  upstream original is unchanged (NOTICE, README, commit history); NOTICE
  records the historical PR #2 to #3 lineage after #2 was auto-closed by
  the conversion.
- CONTRIBUTING, README, NOTICE, landing page, and llms.txt now frame
  Agentlink as a Snap Synapse project with upstream lineage, with this
  repository as the canonical home for current direction and releases.
- README, landing page, llms.txt, NOTICE, and sitemap refreshed: standalone
  continuation wording, the 23-tool table, Homebrew install, and a
  2026-06-10 revision date.

## [0.3.0] - 2026-06-10

### Added

- Homebrew tap install: `brew install snapsynapse/tap/agentlink`.
- Pre-built `linux-arm64` release binary alongside the existing darwin and
  linux amd64/arm64 assets.
- The assistant guide now covers workstation install: Homebrew tap as the
  preferred path and a pinned, checksum-gated binary download as fallback,
  each as approval-gated GuideCheck action blocks (guide-version 1.1.0).

### Changed

- The assistant guide is restructured to pass the GuideCheck reference
  verifier 0.5.0 with zero findings: title format, "Stop and ask" heading,
  negation-safe phrasing, and the 8192-byte size cap. The "Public
  information safety" section was folded into the safety rules to stay
  under the cap.

- GuideCheck adoption at the highest guide-file level: a Level 4 target
  `assistant-guide.txt`, sidecar manifest, repository hash anchor, and
  discovery links from README, `llms.txt`, and the landing page.
- User-facing GuideCheck copy now tells operators to verify the guide, read it
  in full, and approve the reported level before assistant action.

### Fixed

- `agentlink clean` now skips broken symlinks instead of deleting them,
  because their original target ownership cannot be verified once the target
  is missing.
- Generated launchd plists now XML-escape interpolated string values, so
  binary paths containing XML-sensitive characters still produce valid plist
  files.
- CI now uses a failing gofmt check instead of rewriting files, removes the
  stale `golint` gate, and pins external check tools instead of installing
  moving `latest` versions.
- `SECURITY.md` now lists the current `0.2.x` release line as supported,
  matching the policy that only the latest tagged release on `main` receives
  security fixes.
- `agentlink sync --dry-run` is now a true preview mode when combined with
  `--backup` or `--force`; it no longer removes conflicting files, rewrites
  symlinks, or creates backup files.
- `agentlink sync --force` now refuses to replace directories and special
  files, avoiding recursive deletion from a misconfigured link path.
- `agentlink hooks remove` now preserves existing hook file permissions and
  reports write failures instead of silently succeeding.
- `agentlink hooks install|remove|status --git` now refuses relative global
  `core.hooksPath` values instead of resolving them from an arbitrary current
  directory.
- `agentlink scan` now recognizes repositories where `.git` is a file, so
  worktrees and many submodule-style checkouts are scanned correctly.
- `agentlink doctor` now checks the real global config path at
  `~/.config/agentlink/config.yaml` instead of accidentally re-checking the
  project config when `.agentlink.yaml` exists.
- Generated git and zsh hook scripts now shell-quote the resolved
  `agentlink` binary path, so installs under paths with spaces continue to
  run correctly.

## [0.2.0] — 2026-04-14

First tagged release of the Snap Synapse fork of
[martinmose/agentlink](https://github.com/martinmose/agentlink). Upstream
never tagged a release, so this fork treats the upstream baseline as an
implicit 0.1.x and starts its own release line at 0.2.0 to avoid collision
and make the fork's additions readable against a clean version boundary.
Preserves the upstream core (init, sync, check, clean, doctor) and adds
detection, scanning, automatic triggers, safe replacement, and
global-config support. Also launches the canonical landing page at
[agentlink.run](https://agentlink.run/).

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

[Unreleased]: https://github.com/snapsynapse/agentlink/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/snapsynapse/agentlink/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/snapsynapse/agentlink/releases/tag/v0.2.0
