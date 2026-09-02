<!-- Upstream template: portfolio-search-indexing-audit bundle v5; repository contract v4 -->
---
title: "Search indexing"
purpose: "Property-specific index policy, validation commands, deployment gate, and console follow-up."
status: active
updated: 2026-09-01
owner: "Sam Rogers"
open_tasks: []
---
# Search indexing

Canonical origin: `https://agentlink.run/`

Console property ID: `sc-domain:agentlink.run` (verified in the visible property selector and GSC `resource_id` on 2026-09-01)

Property mode: `website`

Generated output: `docs`

If deployment assembles a separate staging directory, this path must name that exact deployable artifact, not its source directory.

## Index policy

| Surface | Policy | Reason |
|---|---|---|
| `/` | Index and include in sitemap | Canonical product landing page |
| `/guides/agents-md-symlinks/` | Index and include in sitemap | Practical one-source, many-alias guide |
| `/guides/claude-md-and-agents-md/` | Index and include in sitemap | Tool-specific wrapper decision guide |
| `/guides/monorepo-instruction-files/` | Index and include in sitemap | Root and nested scope guide |
| `/reference/supported-tools/` | Index and include in sitemap | Versioned integration reference |
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
9. Inspect or request indexing for canonical HTML pages only with separate approval.
10. Start issue-group validation only when matching production behavior is live.
11. Record console state under `ops/search/<provider>/YYYY-MM-DD/`.

## Expected noise

- HTTP and `www` variants should redirect to the bare HTTPS origin.
- `/404.html` is intentionally `noindex` and absent from the sitemap.
- Machine-readable discovery files are crawlable but are not HTML index targets.

## Current baseline

- 2026-09-01 pre-change repository contract: one sitemap HTML page, zero offline defects.
- 2026-09-01 pre-change production contract: one sitemap HTML page, zero production defects, exact repository-to-production parity.
- 2026-09-01 GSC Page indexing report, last updated 2026-08-27: one indexed canonical page and four excluded URLs. Three are expected HTTP or `www` redirects. One is the crawlable `/.well-known/assistant-guide.txt` machine surface and is intentionally omitted from the HTML sitemap.
- 2026-09-01 GSC sitemap observation: `https://agentlink.run/sitemap.xml` was successful, submitted 2026-04-20, last read 2026-05-26, and reported one discovered page before this publication tranche.
- 2026-09-01 publication commit `a7008e305e3fefcfbcbebdd6299e13c13b95dbba`: five canonical HTML pages; offline and approved local-server contracts both passed with zero defects and zero infrastructure findings.
- GitHub Actions checks, accessibility regression checks, and the legacy Pages build all completed successfully for the publication commit. Pages remained configured for `main:/docs`, the `agentlink.run` custom domain, and enforced HTTPS.
- The deployed production contract passed with five pages, zero defects, and zero infrastructure findings.
- GSC accepted one refresh of the verified production sitemap at 2026-09-01 20:37 MDT. The row then reported `Success`, submitted 2026-09-01, last read 2026-09-01, and five discovered pages. Discovery is confirmed; inclusion of the four new pages in search results remains pending recrawl.

## Console action ledger

Read this table before opening the console. Add only observed actions and confirmations. An accepted request remains pending until a later report proves completion.

| Provider and property | Action and target | Accepted at | Confirmation | Result class | Repeat policy | Next review |
|---|---|---|---|---|---|---|
| Google Search Console `sc-domain:agentlink.run` | Existing sitemap submission for `https://agentlink.run/sitemap.xml` | 2026-04-20 (observed 2026-09-01) | Status Success; last read 2026-05-26; one discovered page | completed with stale evidence | Superseded by the one authorized refresh below; never remove and re-add as routine recovery. | Closed by the 2026-09-01 refresh |
| Google Search Console `sc-domain:agentlink.run` | Refresh `https://agentlink.run/sitemap.xml` after verified five-page deployment | 2026-09-01 20:37 MDT | `Sitemap submitted successfully`; then Status Success, submitted 2026-09-01, last read 2026-09-01, five discovered pages | completed sitemap refresh; new-page indexing pending recrawl | Do not repeat before a material sitemap change or a named GSC failure. Do not request indexing or validate exclusions under this action. | Next Page indexing report update after 2026-09-01; checkpoint 2026-09-15 |

Keep rejected attempts and unknown outcomes distinct from accepted actions. Do not repeat an accepted action merely because the provider report remains stale.
