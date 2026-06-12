# Contributing

Thanks for considering a contribution. Agentlink is maintained as a Snap
Synapse project with lineage from
[martinmose/agentlink](https://github.com/martinmose/agentlink). Upstream
credit and the MIT license are preserved; current project direction,
issues, and releases live in this repository.

## Build and test locally

Requires Go 1.23+.

```bash
git clone https://github.com/snapsynapse/agentlink.git
cd agentlink
go build ./...
go test ./...
go test -tags=integration ./...
```

The integration suite builds a real binary and exercises `sync`, `scan`,
`detect`, and `hooks` against temporary directories. Run it before sending
a pull request.

## Code style

- `gofmt -s` clean. `go vet ./...` clean. `staticcheck ./...` clean.
- No new runtime dependencies unless discussed in an issue first. The
  compiled binary has zero runtime dependencies today — keeping it that
  way is a feature.
- One logical change per pull request.

## Submitting a pull request

1. Open an issue first for anything beyond a small fix so we can agree
   on scope before you invest time.
2. Branch off `main`. Name the branch after the change
   (`feature/scan-gitignore-respect`, `fix/launchd-path-escape`).
3. Add or update tests. Integration tests live in
   `integration_test.go` behind `//go:build integration`.
4. Update [CHANGELOG.md](CHANGELOG.md) under `[Unreleased]` with one
   bullet describing the user-visible change.
5. Ensure `go build ./...`, `go test ./...`, and the integration suite
   all pass.
6. Open the PR against `main`. Reference the issue in the description.

## Commit messages

Short imperative subject line, no trailing period. Conventional prefix is
fine but not required:

```
feat: add --respect-gitignore flag to scan
fix: handle launchd plist path escaping on paths with spaces
docs: clarify backup behavior for empty files
```

## Adding a new AI tool to the registry

New tools go in `internal/registry/tools.go`. Each entry needs:

- `Name` — canonical display name
- `Command` — the binary name to look for in `PATH` (or empty)
- `GlobalConfig` — the global config path using `~` for home
- `RepoFile` — the file the tool reads from a repo root
- `ReadsAgentsMD` — true if the tool already reads `AGENTS.md` natively

Add a unit-test entry in `tools_test.go` confirming the new tool is
returned and has a unique name. No code changes elsewhere are required.

## Releasing (maintainers)

Distribution has two halves: GitHub release assets on this repo, and the
Homebrew formula in [snapsynapse/homebrew-tap](https://github.com/snapsynapse/homebrew-tap)
(`Formula/agentlink.rb`), which is what `brew install snapsynapse/tap/agentlink`
resolves to. Both are updated by one script:

```bash
scripts/release.sh 0.4.0
```

It runs the test gates, cross-builds the four binaries with version stamps,
writes `SHA256SUMS.txt`, bumps the landing page, tags and publishes the
GitHub release, rewrites the tap formula with the new URLs and checksums,
and runs `scripts/verify-release.sh`. Two manual follow-ups: add the
CHANGELOG section before running the script (it refuses to release a
version the CHANGELOG does not mention), and if the assistant guide's
pinned fallback download URLs should track the new release, update
`docs/.well-known/assistant-guide.txt`, re-verify it, and update its
manifest.

## Where to ask questions

Open a GitHub issue. Discussions are off for now — if traffic picks up,
they'll be enabled.
