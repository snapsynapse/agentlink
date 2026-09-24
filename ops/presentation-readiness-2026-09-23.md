# Presentation readiness record

Repository scope: Agentlink. Assessment: presentation-readiness 0.2.0 review for Moonshots 2026, first run 2026-09-22 at `065b3156d8c63c9039dd456d88abb37c2fa20369` and repeated 2026-09-24 at the same commit. Sam deferred the fixes on 2026-09-22 and authorized them on 2026-09-23. They shipped in PR #9, squash-merged as `f867936`. No release was cut; v0.6.0 remains the current release.

## Findings and disposition

| Finding | Mitigation | Evidence |
|---|---|---|
| P1: the landing page was 848 px wide at 320, 390 and 768 px viewports because `.code-grid` tracks kept the min-content width of long `pre` lines | Code grids use `minmax(0, 1fr)` tracks and `.code-block` sets `min-width: 0`; all `pre` blocks and the link map are keyboard-focusable scroll regions with a visible focus outline | Live page at `f867936`: layout width equals the viewport at 320 and 390 px in a real browser; local render also at 768 and 1280 px; axe 0 violations and 0 incomplete before and after; `go run ./cmd/update-docs -check` passes |
| P1: `README.md` said a symlinked `source` only warns | README states that `sync` refuses a symlinked source by default, exits with an error and changes nothing, and that `--force` links to it anyway while `check` still reports it | v0.6.0 release binary fixtures for default rejection and `--force`; GitHub README at `f867936` |
| P2: hero and feature copy said every tool sees edits instantly | Copy says every linked path holds the same bytes and each tool decides when it rereads them | Live landing at `f867936`; README and the integration reference already state that detection does not prove instruction loading |

## Related work, unchanged by this record

- `handoffs/HANDOFF-2026-09-01-hybrid-instruction-layering.md` and `handoffs/HANDOFF-2026-09-04-skills-instruction-topology.md` stay open as after-event design input. The topology handoff's source-symlink wording item is complete through `f867936`; its report-only diagnostics still need a scope decision and fixtures.
- The AUR package remains planned in `README.md`.
- Local branch `codex/agentlink-homebrew-audit` (`321c931`, remote removed) is patch-equivalent to `fc8abbb` (#7) on `main` (`git cherry` reports it as applied). No branch was deleted during this record.
- Public-claim rule applied here: Agentlink controls file identity, not when or whether a tool loads a file. Copy that implies verified loading across tools needs evidence from that tool.
