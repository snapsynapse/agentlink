<!-- Upstream template: portfolio-search-indexing-audit bundle v5; repository contract v4 -->
---
title: "Search indexing"
purpose: "Property-specific index policy, validation commands, deployment gate, and console follow-up."
status: active
updated: 2026-08-20
owner: "Sam Rogers"
open_tasks: []
---
# Search indexing

Canonical origin: `https://agentlink.run/`

Console property ID: `UNSET: verify before console work`

Property mode: `website`

Generated output: `docs`

If deployment assembles a separate staging directory, this path must name that exact deployable artifact, not its source directory.

## Index policy

| Surface | Policy | Reason |
|---|---|---|
| `/` | Index and include in sitemap | The single canonical reader destination |
| `/404.html` and unknown routes | `noindex` and omit from sitemap | Error surfaces are not content destinations |
| `/robots.txt`, `/sitemap.xml`, `/llms.txt`, favicon, and `/.well-known/*` | Crawlable machine surfaces, omit from the HTML sitemap | Discovery or machine consumption, not canonical HTML |
| GitHub, Homebrew, and other external copies | Omit from sitemap | Distribution surfaces are not site canonical pages |

## Validation lanes

- Offline: `node scripts/check-search.mjs`
- Production after deployment: `node scripts/check-production-search.mjs`
- Machine-readable output: add `--json`
- Local HTTP test: add `--base=http://127.0.0.1:8765/` after starting the static server on port 8765

Exit code `0` is pass, `1` is a site defect, and `2` is configuration or infrastructure failure.

For a creator-profile or external-platform property, replace the website validation lanes with the reports and controls the property actually exposes. Do not invent repository, production, sitemap, or indexing work.

## Deployment and console sequence

1. Run the normal build and offline search contract.
2. If deployment copies or transforms output, stage the exact deployable artifact with the same builder used by release automation.
3. Ensure repository-wide checks include newly scaffolded files, including checks based on `git ls-files`.
4. Deploy through the repository's normal release path.
5. Wait for the deployment to complete.
6. Run the production search contract.
7. Confirm the deployed sitemap URL set matches the repository sitemap.
8. Refresh a materially changed stale sitemap at most once, using its full canonical URL for a domain property.
9. Inspect or request indexing for canonical HTML pages.
10. Start issue-group validation only when matching production behavior is live.
11. Record console state under `ops/search/<provider>/YYYY-MM-DD/`.

## Expected noise

- HTTP and `www` variants should redirect to the bare HTTPS origin.
- `/404.html` is intentionally `noindex` and absent from the sitemap.
- Machine-readable discovery files are crawlable but are not HTML index targets.

## Current baseline

- 2026-08-20 repository contract: one sitemap HTML page, zero offline defects.
- 2026-08-20 production contract before this publication tranche: one sitemap HTML page, zero production defects, exact repository-to-production parity.
- Console property identity and dated provider evidence remain unverified. Do not infer zero activity or perform console actions from this absence.

## Console action ledger

Read this table before opening the console. Add only observed actions and confirmations. An accepted request remains pending until a later report proves completion.

| Provider and property | Action and target | Accepted at | Confirmation | Result class | Repeat policy | Next review |
|---|---|---|---|---|---|---|

Keep rejected attempts and unknown outcomes distinct from accepted actions. Do not repeat an accepted action merely because the provider report remains stale.
