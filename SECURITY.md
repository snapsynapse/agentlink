# Security Policy

## Supported versions

Only the latest tagged release on `main` receives security fixes. This is
a small fork maintained by Snap Synapse — there is no LTS branch.

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes       |
| < 0.1   | No        |

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

- Acknowledgement within 72 hours
- Initial assessment within 7 days
- Fix or mitigation plan within 30 days for confirmed vulnerabilities

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
