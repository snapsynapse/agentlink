# Agentlink

[![Checks](https://github.com/snapsynapse/agentlink/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/snapsynapse/agentlink/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> **Origin.** Agentlink began as [martinmose/agentlink](https://github.com/martinmose/agentlink) by Martin Mose Facondini (MIT). This repository is the Snap Synapse-maintained standalone continuation, extended with `detect`, `scan`, `hooks`, automatic backup, global-config support, and integration tests. See [NOTICE](NOTICE) for full attribution and lineage.

Sync one AGENTS.md to every AI coding tool on your machine, with **zero magic**, just symlinks.

Different tools want different files at project root: `AGENTS.md` (OpenAI/Codex, OpenCode), `CLAUDE.md` (Claude Code), `GEMINI.md`, etc. There's no standard, and I'm not waiting for one. **Agentlink** solves the basic need: keep your **personal** instruction files (in `~`) and your **project** instruction files in sync **without generators**. Edit one, they all reflect it.

Creating instruction files is easy with `/init` commands, but keeping them up to date is the hard part -- and expensive too. Good instruction files are often crucial and make a huge difference when using agentic tools. Since they're so important, these files are typically generated with expensive models. Why pay repeatedly to regenerate similar content across different tools?

**Future-proof by design:** We don't know what tomorrow brings in the AI tooling space, but agentlink is ready. New tool expects `.newtool/ai-config.md`? Just add it to your config. Complex nested structure like `workspace/ai/tools/newframework/instructions.md`? No problem. Agentlink automatically creates the directories and symlinks without any code changes needed.

> Scope: **instruction files only**. No MCP `.mcp.json` or chain configs. Simple on purpose.

## GuideCheck

Agentlink publishes a GuideCheck Human-Verifiable Assistant Guide for bounded local review, build, and test work:

- Assistant guide: https://agentlink.run/.well-known/assistant-guide.txt
- Manifest: https://agentlink.run/.well-known/assistant-guide-manifest.txt
- Target conformance: GuideCheck Level 4, the highest guide-file level. Level 5 requires a conformant assistant runtime and is not claimed by this repository.

The guide is also stored at `docs/.well-known/assistant-guide.txt` so the public repository copy can serve as an independent Level 4 hash anchor.
Before asking an assistant to perform local Agentlink setup or verification work, verify the guide, read it in full, and approve proceeding under the reported level.

---

## Why Agentlink?

- **One real file, many aliases**: pick a *source* (`CLAUDE.md` or `AGENTS.md` or whatever), symlink the rest.
- **No codegen**: no templates, no transforms, no surprise diffs.
- **Project + global**: works in repos *and* under `~/.config/...`.
- **Auto-detect**: scans your system for installed AI tools and reports what it finds.
- **Repo scanning**: walks a directory tree and wires up symlinks in every git repo that has an AGENTS.md.
- **Automatic triggers**: git hooks, shell hooks, and launchd keep things synced without manual runs.
- **Idempotent**: re-run safely; it fixes broken/misdirected links.
- **Portable**: works on macOS and Linux.
- **Future-ready**: handles any directory structure, automatically creates paths. Tomorrow's AI tool? Just add its path.

---

## How it works

You tell Agentlink which file is the **source**, and which other files should **link** to it. Agentlink creates/fixes symlinks accordingly.

```yaml
# .agentlink.yaml (in project root)
source: AGENTS.md
links:
  - CLAUDE.md                             # Claude Code
  - .github/copilot-instructions.md       # GitHub Copilot
  - .cursorrules                           # Cursor AI
  - GEMINI.md                              # Gemini CLI
```

Result:
```
./AGENTS.md                              # real file you edit
./CLAUDE.md                           -> AGENTS.md  (symlink)
./.github/copilot-instructions.md     -> ../AGENTS.md  (symlink)
./.cursorrules                        -> AGENTS.md  (symlink)
./GEMINI.md                           -> AGENTS.md  (symlink)
```

Global mode (in HOME) is the same idea:

```yaml
# ~/.config/agentlink/config.yaml
source: ~/AGENTS.md
links:
  - ~/.claude/CLAUDE.md
  - ~/.codex/AGENTS.md
  - ~/.gemini/GEMINI.md
  - ~/.config/AGENTS.md
  - ~/.factory/AGENTS.md
  - ~/.config/opencode/AGENTS.md
```

---

## Install

### Homebrew (macOS and Linux)

```bash
brew install snapsynapse/tap/agentlink
```

The formula lives in [snapsynapse/homebrew-tap](https://github.com/snapsynapse/homebrew-tap) and installs the same checksummed binaries as the release page.

### Pre-built binary

Download the binary for your platform from the [latest release](https://github.com/snapsynapse/agentlink/releases/latest), verify it against `SHA256SUMS.txt`, then:

```bash
chmod +x agentlink-darwin-arm64 && mv agentlink-darwin-arm64 /usr/local/bin/agentlink
```

Binaries: `darwin-arm64`, `darwin-amd64`, `linux-amd64`, `linux-arm64`.

### From source (requires Go 1.23+)

```bash
go install github.com/snapsynapse/agentlink/cmd/agentlink@latest
```

This puts the binary in your Go bin directory (usually `~/go/bin/`). Make sure it's in your PATH:

```bash
# Check if it worked
which agentlink

# If "command not found", add Go's bin to your PATH.
# Add this line to your ~/.zshrc or ~/.bashrc:
export PATH="$HOME/go/bin:$PATH"
```

### Planned distribution

- **AUR**: `yay -S agentlink-bin`

---

## Quick Start: Global Setup

If you maintain one set of AI instructions for all your tools, this is the fastest path.

**1. Create your source file** (skip if you already have one):

```bash
# ~/AGENTS.md is the recommended global location.
# Put your instructions, conventions, and preferences here.
vim ~/AGENTS.md
```

**2. Detect your installed tools:**

```bash
agentlink detect -v
```

This shows which tools are installed and their expected config paths.

**3. Create the global config:**

```bash
mkdir -p ~/.config/agentlink
cat > ~/.config/agentlink/config.yaml << 'EOF'
source: ~/AGENTS.md
links:
  - ~/.claude/CLAUDE.md
  - ~/.codex/AGENTS.md
  - ~/.gemini/GEMINI.md
  # Add paths from 'agentlink detect' for tools you use
EOF
```

**4. Sync:**

```bash
agentlink sync
```

If any target path already has a real file, agentlink will refuse to overwrite it and tell you the file size and modification date. Your options:

```bash
cat ~/.codex/AGENTS.md         # inspect the existing file
agentlink sync --backup        # back up existing regular files to .bak, then replace
agentlink sync --force         # replace without backup (destructive)
agentlink sync --dry-run       # preview only, with no filesystem changes
```

**5. Install automatic triggers** (optional but recommended):

```bash
agentlink hooks install --all
```

This installs git hooks, a zsh directory-change hook, and a 60-minute launchd heartbeat so syncs happen automatically. Generated hook scripts safely quote the installed binary path, so installs under directories with spaces still work. If your global Git `core.hooksPath` is relative, agentlink refuses to guess where to install hooks; unset it or change it to an absolute path first.

**6. Scan your repos** (optional):

```bash
agentlink scan ~/Git
```

Finds git repos with AGENTS.md and creates tool-specific symlinks (CLAUDE.md, GEMINI.md, etc.) in each, including git worktrees and submodule-style checkouts where `.git` is a file rather than a directory.

---

## Quick Start: Per-Project Setup

```bash
# Initialize in your project
agentlink init

# Edit the created .agentlink.yaml to match your needs
# Create your source file (e.g., AGENTS.md)

# Sync to create symlinks
agentlink sync
```

### Commands

```bash
agentlink init               # create .agentlink.yaml in current directory
agentlink sync               # create/fix symlinks based on config
agentlink check              # print status and problems
agentlink clean              # remove managed symlinks (non-destructive)
agentlink doctor             # environment + permissions sanity checks for project and global config
agentlink detect             # auto-detect installed AI coding tools
agentlink scan [dir]         # scan git repos and manage symlinks
agentlink hooks install --all # install all automatic sync triggers
agentlink hooks remove --all  # remove all automatic sync triggers
agentlink hooks status       # show installed trigger status
```

### Helpful flags

```bash
agentlink sync --dry-run     # show what would change without filesystem changes
agentlink sync --backup      # back up existing regular files to .bak before replacing
agentlink sync --force       # replace existing regular files without backup (or -f)
agentlink sync --quiet       # suppress non-error output (or -q)
agentlink --verbose          # detailed output for any command (or -v)
```

### Handling existing files

When a target path already contains a real file (not a symlink), agentlink stops and reports the conflict with the file size and last-modified date. It never silently overwrites your files. Options:

- `--backup` backs up the existing file to `<name>.bak` (or `<name>.<timestamp>.bak`, with a numeric suffix for further collisions), then creates the symlink. Backups use an exclusive hard link so an existing backup is never overwritten. On filesystems without hard-link support, sync stops and leaves the original unchanged.
- `--force` replaces a regular file without backup. Use when you've already inspected or don't care about the existing content.
- `--dry-run` is a hard preview mode. It does not create symlinks, remove files, fix broken links, or write backups, even when combined with `--backup` or `--force`.
- Neither flag: agentlink reports the conflict and skips the file.

Agentlink refuses to recursively replace directories or special files. If a configured link path is a directory, move or rename it yourself before running `sync`.

---

## Tool Detection

Agentlink maintains a registry of known AI coding tools and their configuration paths. Run `detect` to see what's installed:

```bash
agentlink detect             # list installed tools
agentlink detect --generate  # generate .agentlink.yaml from detected tools
agentlink detect -v          # show global config paths and AGENTS.md support
```

### Supported tools

| Tool | Global Config | Repo File | Reads AGENTS.md |
|------|--------------|-----------|-----------------|
| Aider | -- | AGENTS.md | Yes |
| Amp | ~/.config/AGENTS.md | AGENTS.md | Yes |
| Antigravity | -- | AGENTS.md | Yes |
| Autohand | -- | AGENTS.md | Yes |
| Claude Code | ~/.claude/CLAUDE.md | CLAUDE.md | No (CLAUDE.md) |
| Cline | -- | -- | No |
| Continue | ~/.continue/config.yaml | -- | No |
| Crush | -- | AGENTS.md | Yes |
| Cursor | -- | AGENTS.md | Yes |
| Factory (Droid) | ~/.factory/AGENTS.md | AGENTS.md | Yes |
| Gemini CLI | ~/.gemini/GEMINI.md | GEMINI.md | No (GEMINI.md) |
| GitHub Copilot | -- | .github/copilot-instructions.md | No |
| Goose | ~/.config/goose/.goosehints | .goosehints | No |
| Junie | -- | .junie/AGENTS.md | Yes |
| Kilo Code | -- | AGENTS.md | Yes |
| Kiro | -- | AGENTS.md | Yes |
| Codex CLI | ~/.codex/AGENTS.md | AGENTS.md | Yes |
| OpenClaw | ~/.openclaw/workspace/AGENTS.md | -- | Yes |
| OpenCode | ~/.config/opencode/AGENTS.md | AGENTS.md | Yes |
| Qwen Code | ~/.qwen/QWEN.md | QWEN.md | No (QWEN.md) |
| RooCode | -- | .roo/rules/rules.md | No |
| Windsurf | -- | AGENTS.md | Yes |
| Zed | -- | AGENTS.md | Yes |

To add a new tool, edit `internal/registry/tools.go` and add an entry to the `All()` function.

---

## Repo Scanning

Scan a directory tree to find git repos and wire up symlinks:

```bash
agentlink scan                    # scan ~/Git (default)
agentlink scan ~/Projects         # scan a different directory
agentlink scan --dir ~/Work       # alternative syntax
agentlink scan --dry-run          # preview without changes
```

The scanner finds repos containing `AGENTS.md` and creates symlinks for tool-specific filenames (`CLAUDE.md`, `GEMINI.md`, etc.). It recognizes standard repos, git worktrees, and submodule-style checkouts where `.git` is a file. It does **not** inject `AGENTS.md` into repos that lack one.

The default scan directory is `~/Git`. Override it per-invocation with the `--dir` flag or positional argument. To change the compiled default, build with:

```bash
go build -ldflags "-X github.com/snapsynapse/agentlink/internal/cli.DefaultScanDir=/your/path" ./cmd/agentlink/
```

---

## Automatic Triggers

Keep symlinks current without manual runs:

```bash
agentlink hooks install --all      # install all triggers
agentlink hooks install --git      # git post-checkout + post-merge hooks
agentlink hooks install --zsh      # zsh chpwd hook (sync on cd)
agentlink hooks install --launchd  # macOS LaunchAgent (60-minute heartbeat)
agentlink hooks status             # check what's installed
agentlink hooks remove --all       # clean up all triggers
```

**Git hooks** use `core.hooksPath` for global hooks. After any checkout or merge, agentlink syncs the current repo's symlinks. Agentlink only installs into absolute global hook paths. If `git config --global core.hooksPath` returns a relative path, either unset it so agentlink can create `~/.config/git/hooks`, or set it to an absolute directory before running `agentlink hooks install --git`.

**Zsh hook** fires on every `cd` into a directory that contains both a git checkout and `.agentlink.yaml`. Runs in the background so it never slows your shell.

**LaunchAgent** runs `agentlink sync` every 60 minutes and at login. Logs to `/tmp/agentlink-sync.log`.

All injected content is wrapped in markers (`# >>> agentlink >>>` / `# <<< agentlink <<<`) for clean removal. Removing those sections preserves the hook file's existing permissions. Generated hook commands shell-quote the binary path so installs under paths with spaces remain valid.

---

## Config

### Project config (recommended)

See [`examples/project.agentlink.yaml`](examples/project.agentlink.yaml) for a
copyable project configuration.

Place a single file at repo root:

`.agentlink.yaml`
```yaml
source: AGENTS.md
links:
  - CLAUDE.md
  - GEMINI.md
```

Notes:
- **`source` must be a real file**, not a symlink (Agentlink warns if it is).
- Paths in `links` are relative to the project root.

### Global config

See [`examples/global.agentlink.yaml`](examples/global.agentlink.yaml) for a
copyable global configuration.

`~/.config/agentlink/config.yaml`
```yaml
source: ~/AGENTS.md
links:
  - ~/.claude/CLAUDE.md
  - ~/.codex/AGENTS.md
  - ~/.gemini/GEMINI.md
```

### Doctor behavior

`agentlink doctor` checks project and global configuration separately. In particular, the "Global Configuration" section always inspects `~/.config/agentlink/config.yaml` directly, even when `.agentlink.yaml` exists in the current repo.

---

## Platform notes

- **macOS + Linux**: standard POSIX symlinks (`ln -s`), same behavior on both.
- **Git**: symlinks are stored as links (not file copies). That's fine; teams who dislike that can add them to `.gitignore`.

### Gitignore patterns

Since agentlink creates multiple instruction files but only one is the real source, you can gitignore all AI instruction files except your chosen source:

```gitignore
# Ignore all AI instruction files
AGENTS.md
CLAUDE.md
GEMINI.md
OPENCODE.md
.cursorrules
.github/copilot-instructions.md

# But track your chosen source file (example: tracking AGENTS.md)
!AGENTS.md
```

This keeps your repository clean while ensuring your source file is version controlled.

- **Editors/IDEs**: most follow symlinks transparently.

---

## Prior Art

Agentlink is not the first tool to tackle this problem. The ecosystem is young and fragmented; many people have built solutions independently. Here's what exists as of April 2026:

| Tool | Language | Approach | Scope |
|------|----------|----------|-------|
| [agentsync](https://github.com/dallay/agentsync) | Rust | Symlinks + MCP config generation | Per-project .agents/ directory |
| [agents](https://github.com/amtiYo/agents) | TypeScript | Config-file generation + watch mode | MCP servers + skills + instructions |
| [AI Rules Sync](https://github.com/lbb00/ai-rules-sync) | Node | Git-repo-sourced rules + adapters | Multi-repo rule federation |
| [Rulesync](https://github.com/dyoshikawa/rulesync) | Node | Bidirectional format conversion | Rules + commands + MCP + skills |
| [Vibe Rules](https://github.com/sky1core/viberules) | Node | Hard links/symlinks + VS Code ext | 15+ tools, skill management |
| [agent-sync](https://github.com/ZacheryGlass/agent-sync) | Python | Format conversion, hub-and-spoke | Claude + Copilot + Gemini |
| [agent-sync](https://github.com/GowayLee/agent-sync) | OCaml | Symlinks, AGENT_GUIDE.md canonical | Early stage |
| [claude-agents-sync](https://github.com/alexandrbasis/claude-agents-sync) | Python | PostToolUse hook, bidirectional | Claude Code only |
| [AgentLoom](https://github.com/Alpha-Coders/agent-loom) | Rust/Svelte | Desktop GUI for skill symlinks | Skills only (12+ tools) |
| [DevKit](https://github.com/ngxtm/devkit) | Node | Skill marketplace + auto-detect | Skill discovery + install |
| [AGR](https://pypi.org/project/agr/) | Python | Package manager for agent resources | Team skill distribution |

Agentlink differentiates by staying minimal (symlinks only, no codegen), supporting both global and project-level configs, and adding automation (detect, scan, hooks) without requiring Node/Python runtimes. The compiled Go binary has zero runtime dependencies.

For the emerging AGENTS.md standard, see [agents.md](https://agents.md/) (now under the Agentic AI Foundation / Linux Foundation).

---

## FAQ

**Why not templates or generators?**
Because 90% of the time the files **should be identical**. When they're not, this tool isn't the right fit (or add a second source and stop linking that one).

**What if my source differs per project?**
Perfect -- put a `.agentlink.yaml` in each repo and choose the source you actually edit there.

**Can the source be `AGENTS.md` instead of `CLAUDE.md`?**
Yes. The source is *whatever you want to edit*. The others link to it.

**What happens when a new AI tool comes out?**
Just add its expected path to your config. If "SuperCoder AI" expects `.supercoder/prompts/main.md`, add that path and run `agentlink sync`. Directories are created automatically, symlink points to your source file. Zero code changes, zero updates needed. Or submit a PR to add it to the tool registry.

**MCP / `.mcp.json`?**
Out of scope. Formats differ between tools; symlinking a single JSON to multiple consumers usually doesn't make sense.

**What about local models (Ollama, LM Studio, etc.)?**
Local model runners don't read AGENTS.md or any instruction file convention. The model itself has no filesystem protocol -- it depends on the harness. If a tool built on top of local models adds AGENTS.md support, we'll add it to the registry.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). To add a tool to the registry, open an issue using the "New tool" template.

## About

Maintained as a [Snap Synapse](https://snapsynapse.com/) project. Agentlink began as [martinmose/agentlink](https://github.com/martinmose/agentlink) by Martin Mose Facondini (MIT); upstream credit remains intact in [NOTICE](NOTICE). This repository is the canonical home for current project direction, issues, releases, and documentation.
