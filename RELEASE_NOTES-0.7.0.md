# Agentlink v0.7.0

Final release. Agentlink is retired, and no further releases are planned. This release fixes a security issue in how project configs were applied, retires hook installation, and makes hook removal undo earlier installs completely.

## Security

In v0.6.0 and earlier, a project `.agentlink.yaml` could point its source and links anywhere: `~/` and absolute paths, `..` segments, and directory symlinks committed to the repository all escaped it. That config arrives with whatever repository is cloned, and the git and zsh hooks applied it automatically, so cloning or pulling a hostile repository could create symlinks such as a shell startup file pointing at the repository's own content, or link a private file into the repository as an instruction file that AI tools then load.

In v0.7.0, a project config's source and every link must resolve inside the directory that holds the config, after following symlinks. The check runs when the config loads and again immediately before each link is created. `scan` applies it to every repository it visits. The global config at `~/.config/agentlink/config.yaml` is unchanged.

## Changes

- `agentlink hooks install` is retired and only explains how to remove triggers. `hooks status` and `doctor` report triggers that remain.
- `agentlink hooks remove --git` deletes hook files that held only agentlink's block. When the global `core.hooksPath` is agentlink's default `~/.config/git/hooks` and nothing else is left in it, the directory is removed and `core.hooksPath` is unset, so each repository's own `.git/hooks` run again. Otherwise it reports that `core.hooksPath` is still set.
- `agentlink hooks remove --launchd` also deletes the legacy `/tmp/agentlink-sync.log` and `.err` when they are regular files.
- Edits to `~/.zshrc` and Git hooks are atomic. `init`, `detect --generate` and default global-config creation never write through a symlink.
- `go.mod` retracts v0.6.0 and earlier, so `go install ...@latest` resolves to this release.
- The Homebrew formula is marked deprecated and shows removal steps after install.
- Assistant guide 1.2.5 installs this release and uses it as its immutable anchor.

## Upgrade notes

- A project config that points outside its repository now fails as a whole. Move those links to the global config.
- Remove automatic triggers installed by earlier releases: `agentlink hooks remove --all`.
- If you can't upgrade, remove the triggers with your installed version and then run `git config --global --unset core.hooksPath`, which v0.6.0's removal leaves set.
- Check for links an earlier release may have created from a cloned repository's config. The README's Uninstall section has the command.

Binaries support macOS and Linux on ARM64 and AMD64. Native Windows is not a release target. The tool registry is unchanged from v0.6.0.

## Verification

Release preparation runs unit, race, integration, vet, deterministic cross-build, checksum, and host-version checks. Publication follows exact-commit CI and accessibility checks. Download verification compares all binary hashes and checks the host binary version.
