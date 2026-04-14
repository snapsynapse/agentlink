# Contributing

Thanks for considering a contribution. This is a Snap Synapse fork of
[martinmose/agentlink](https://github.com/martinmose/agentlink) — upstream
credit and MIT license are preserved. Contributions that improve both the
fork and upstream are welcome and will be offered back where applicable.

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

## Where to ask questions

Open a GitHub issue. Discussions are off for now — if traffic picks up,
they'll be enabled.
