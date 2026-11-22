# Contributing

Agentlink was retired on November 21, 2026, and v0.7.0 is the final release.
Pull requests, feature requests, and registry additions are not being
accepted. Report security issues privately as described in
[SECURITY.md](SECURITY.md).

Agentlink is MIT licensed, so forks are welcome. It began as
[martinmose/agentlink](https://github.com/martinmose/agentlink); keep the
upstream credit in [NOTICE](NOTICE) intact in any fork. The notes below
record how the project was built and tested, for anyone continuing it.

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
`detect`, and `hooks` against temporary directories. Its scan contracts check
exact root and nested topology, relative targets, explicit-config authority,
repository containment, wrapper preservation, idempotence, and byte-preserving
dry runs.

## Code style

- `gofmt -s` clean. `go vet ./...` clean. `staticcheck ./...` clean.
- The Go directive remains the minimum consumer contract: 1.23. CI tests
  exact Go 1.23.12, 1.26.8, and 1.27.1, with the current stable release also
  exercised on macOS. Staticcheck 2026.2.1 runs on the Go 1.27.1 quality lane
  because it supports Go 1.27 and itself requires Go 1.26 or newer. Do not run
  an older Staticcheck binary against packages compiled by a newer Go release.
- The compiled binary has zero runtime dependencies.
- Project configuration is untrusted input: a project config's source and
  links must resolve inside its repository after symlinks are followed.

## The tool registry

The registry is frozen at v0.7.0. Entries live in `internal/registry/tools.go`.
Each entry has:

- `Name`: canonical display name
- `DetectCommands`: the binary name to look for in `PATH` (or empty)
- `GlobalConfigPath`: the global config path using `~` for home
- `RepoFileName`: the file the tool reads from a repo root
- `ReadsAgentsMD`: true if the tool already reads `AGENTS.md` natively
- `PreferredIntegration`: native, configurable, import, symlink, or unsupported
- `SupportsNestedRepoFile`: true only for documented nested discovery
- `NestedSupportReference`: public HTTPS documentation required when nested discovery is enabled

`IntegrationReference`, `ReviewedOn`, and `IntegrationNotes` record vendor
evidence. Registry invariants and the generated homepage listing are tested;
`go run ./cmd/update-docs` regenerates the listing.

## Releasing

No further releases are planned. `RELEASE_CHECKLIST.md` records the process,
kept for an emergency security release only. Preparation and publication are
separate:

```bash
scripts/prepare-release.sh X.Y.Z
scripts/release.sh X.Y.Z
```

`prepare-release.sh` builds and verifies deterministic assets without external
writes. `release.sh` tags the existing commit, publishes the declared assets,
updates the Homebrew formula, and independently downloads and verifies the
release. It does not edit, commit, or push `main`.

## Questions

The project is retired. Issues are read but may not get a response.
