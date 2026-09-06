# Review mitigation record

Repository scope: Agentlink. Baseline reviewed: `0c8f88c56d773d0fecc02490662fad0867e6c396`. Local changes prepared September 5, 2026, following authorization to mitigate the code, documentation, and website findings and consolidate support information. No release, commit, push, or deployment was performed.

## Findings and disposition

| Finding | Mitigation | Evidence |
|---|---|---|
| R1: hook removal ignored dry-run | Managed-section removal and LaunchAgent unload/removal respect preview; all trigger selectors tested with isolated HOME and a launchctl stub | `safety_integration_test.go`, `internal/cli/hooks.go` |
| R2: backup could remove a source through a parent alias | Shared preflight runs before backups or empty-file removal; source backing files are protected even with forced source symlinks | Source-alias and source-backing-file integration fixtures |
| R3: lexical path comparisons could approve unreadable links | Check resolves the actual filesystem link; creation computes relative targets from physical parent paths | Physical-parent fixture reads the expected bytes and checks the resulting alias |
| R4: broken unowned links overwritten implicitly | Replacement requires force; backup authorizes only regular files | Sync, sync-backup, and scan preserve unrelated broken links |
| R5: dangling project config caused scope fallback | Lstat preserves project configuration authority; load failures remain errors | Scan/sync/check/clean fixtures fail without creating default aliases |
| R6: hook composition broke other interpreters or appended after exit | Preflight rejects non-shell/non-regular hooks; shell insertion precedes existing body and removes cleanly | Python hook unchanged; shell trigger executes before early exit and original body is restored |
| R7: checksum action only fetched hashes | Separate download and executable comparison actions; asset filename consistent through install | Valid, tampered, and missing-entry checksum tests; guide/manifests match hashes and profile limits |
| R8: YAML examples contained shell commands | File contents and terminal commands are separate | Strict YAML decoding of five website examples |
| R9: stale tool classifications and paths | Revalidated registry includes dated vendor references and client caveats; Cline, Aider, Continue, Junie, Copilot, Qwen, Goose, and global paths corrected | Registry contracts; generated homepage matches runtime metadata |
| R10: homepage implied sync updated both scopes | Explain selection; add explicit --global to sync/check/clean | Isolated global-selector fixture verifies all three commands |
| R11: accessibility gate covered only homepage | Pinned action now discovers and rewrites the full sitemap for local scanning | Workflow validator passes; hosted accessibility execution remains pending |

## Consolidation and supporting fixes

The homepage is the complete support listing: macOS/Linux on ARM64/AMD64 plus all 23 AI-tool integrations. It is generated from `internal/registry/tools.go`, with project/global paths, integration modes, vendor references, and caveats. README and the integration reference link to this listing. The latter retains guidance about modes and nested discovery. Tests reject generated-listing drift; `go run ./cmd/update-docs` refreshes it.

Mobile navigation remains visible and can wrap. The support table is a labeled, keyboard-focusable scrolling region with row and column headers. This is source-level remediation; no rendered browser or assistive-technology verification was performed after the user dropped Comet.

Contributor field names now match the registry. The multi-page llms.txt exception was reconsidered in INTENT.md. The monorepo guide explains separate package configs for independently scoped sources and the alternative unconfigured nested-scan workflow. Release guidance distinguishes recovery of a missing release, missing asset, failed tap update, and remaining verification without replaying completed publication steps.

## Validation

- Passed `go test -race -tags=integration ./...` on Go 1.27.0, darwin/arm64, including the new safety and documentation fixtures.
- Passed `go vet ./...`, `go mod tidy -diff`, pinned Staticcheck 2026.2.1, and actionlint v1.7.7.
- Passed generated support-listing check, release contract, release shell syntax, spelling checks of changed source/docs, and whitespace checks.
- Search contract: five sitemap pages, zero defects, zero infrastructure failures.
- Cross-builds passed for darwin/arm64, darwin/amd64, linux/amd64, and linux/arm64. Go emitted non-fatal sandbox cache-write warnings while recording VCS version metadata; all four builds exited successfully. Cross-builds do not prove runtime behavior on the other three targets.
- The old installed Staticcheck could not parse the current Go export format. The repository-pinned version was fetched into temporary directories and passed; no global tool installation was changed.

## Release and migration boundaries at mitigation closeout

These changes are local and unreleased. Published binaries remain v0.5.0; the website labels the refreshed listing and --global as source changes. Existing user configurations are not automatically rewritten. Review old Qwen/Goose aliases for duplicate instruction loading, and remove Continue settings-file paths from Agentlink's links list before replacing them with an instruction-rule alias.

The revised assistant guide is version 1.2.4, status draft. Its manifest deliberately records `pending-publication` rather than claiming that the old v0.5.0 anchor contains new bytes. Release preparation rejects the draft. Preparing the next release requires activating the reviewed guide and assigning the prospective release-tag anchor; independent remote verification follows publication. Matching local copies alone do not establish GuideCheck Level 4.

Still pending: exact-commit hosted CI and accessibility scan, rendered/keyboard review, publication and deployed-byte verification, and tool-specific fresh-session loading checks. Vendor documentation supports the listed integrations; it does not substitute for smoke tests of every installed client/version.

## Authorized release preparation

The user subsequently authorized staging, commit, push, tag, and release. Version 0.6.0 carries these changes. The guide is activated and its manifest names the prospective v0.6.0 anchor. Publication completion must be established from the remote tag, GitHub Release, exact-commit workflows, and deployed bytes; the earlier draft and local-only statements above describe the pre-release checkpoint.
