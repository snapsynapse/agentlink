# Security Policy

## Status

Agentlink is retired as of November 21, 2026. v0.7.0 is the final release.
No further releases are planned, and fixes are not guaranteed. Earlier
releases are retracted in `go.mod`, so `go install ...@latest` resolves to
v0.7.0.

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.7.x   | Yes, 0.7.0 is the final release. |
| < 0.7   | No. 0.6.0 and earlier have a vulnerability fixed in 0.7.0; upgrade. |

If you installed automatic triggers with an earlier release, remove them:
`agentlink hooks remove --all`. See the README's Uninstall section.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security reports.

Use one of:

1. **GitHub Security Advisories** — preferred. Open a private advisory at
   https://github.com/snapsynapse/agentlink/security/advisories/new
2. **Email** — security@snapsynapse.com with subject
   `[agentlink] <short description>`.

Include:

- Affected version (`agentlink --version` output)
- OS and shell
- Reproduction steps or proof-of-concept
- What you believe the impact is

## What to expect

Reports are still read. Because the project is retired, a confirmed
vulnerability gets a published advisory with mitigation guidance; a new
release is made only at the maintainer's discretion. Expect an
acknowledgement within 7 days.

Credit in the release notes if you'd like it. Anonymous reports are fine.

## Scope

In scope:

- Path traversal or symlink-attack vectors in `sync`, `scan`, or `hooks`
- Command injection via config fields, scan paths, or launchd plist
  generation
- Privilege escalation via installed hooks
- Data loss from `--backup` or `--force` behavior that contradicts the
  documented semantics

Out of scope:

- Anything affecting only forks that have diverged from upstream
  `martinmose/agentlink` + Snap Synapse additions
- Vulnerabilities in AI tools agentlink links to — report those to the
  respective projects

## Defensive notes

Agentlink manages symlinks and writes to paths from user-controlled
config. It does not make network calls, does not execute remote code, and
does not embed a runtime. The main attack surface is the filesystem
operations in `internal/symlink/` and the hook installers in
`internal/cli/hooks.go`.

Project configs and scanned repositories are untrusted input. Their source and
links must resolve inside the repository (`symlink.CheckContained`), checked at
load time and again immediately before each link is created. Only the global
config at `~/.config/agentlink/config.yaml` may point elsewhere.
