# Google Search Console audit: agentlink.run

Property: `sc-domain:agentlink.run`

Observation date: 2026-09-01

Authority: audit plus one post-deployment sitemap refresh. No URL Inspection indexing requests and no issue-group validation.

## Property identity

- The existing shared, signed-in Google Search Console tab showed `agentlink.run` in the visible property selector.
- The same tab URL contained `resource_id=sc-domain%3Aagentlink.run`.
- No property was inferred, created, or changed.

## Dated report evidence

| Report | Report date or window | Observation | Classification |
|---|---|---|---|
| Page indexing | Last updated 2026-08-27 | One indexed URL: `https://agentlink.run/` | policy decision; matches the pre-change index policy |
| Page indexing | Last updated 2026-08-27 | Three excluded HTTP or `www` variants redirect to the bare HTTPS origin | expected noise |
| Page indexing | Last updated 2026-08-27 | `https://agentlink.run/.well-known/assistant-guide.txt` was crawled but not indexed | policy decision; crawlable machine surface, not an HTML sitemap target |
| Sitemaps | Observed 2026-09-01 | `https://agentlink.run/sitemap.xml` was successful; submitted 2026-04-20; last read 2026-05-26; one discovered page | pending recrawl after material sitemap revision |
| Sitemaps | Refreshed 2026-09-01 at 20:37 MDT | Submission succeeded; submitted and last-read dates changed to 2026-09-01; five pages discovered; Status Success | pending recrawl for page indexing; sitemap refresh completed |
| HTTPS | Last updated 2026-08-29 | One HTTPS URL, zero non-HTTPS URLs, no issues detected in the report window | expected noise; no actionable issue |
| Core Web Vitals | Last updated 2026-08-30 | Insufficient usage data for mobile and desktop | external limitation |
| Manual actions | Observed 2026-09-01 | No issues | expected noise; no actionable issue |
| Security issues | Observed 2026-09-01 | No issues | expected noise; no actionable issue |
| Performance | 2026-05-31 through 2026-08-30 | 2 clicks, 300 impressions, 0.7% CTR, average position 46.8 | unknown; insufficient demand evidence and query text not archived |

## Pre-change validation

- Repository contract: one canonical HTML page, zero defects, zero infrastructure findings.
- Production contract: one canonical HTML page, zero defects, zero infrastructure findings.
- Repository and production bytes matched for the HTML page, sitemap, robots file, LLM summary, GuideCheck files, AI posture, and 404 page.

## Post-change validation

- Publication commit: `a7008e305e3fefcfbcbebdd6299e13c13b95dbba`.
- Repository contract: five canonical HTML pages, zero defects, zero infrastructure findings.
- Approved local-server production contract: five canonical HTML pages, zero defects, zero infrastructure findings.
- GitHub Actions checks, accessibility regression gate, and legacy Pages build: completed successfully for the publication commit.
- GitHub Pages source remained `main:/docs`; custom domain remained `agentlink.run`; HTTPS remained enforced.
- Deployed production contract: five canonical HTML pages, zero defects, zero infrastructure findings.

## Finding ledger

| Finding | Classification | Handling |
|---|---|---|
| The only pre-change canonical HTML page was indexed. | policy decision | Preserve while adding useful canonical pages. |
| Three protocol or host variants were excluded as redirects. | expected noise | Do not validate a fix or request indexing. |
| The served assistant guide was crawled but not indexed. | policy decision | Preserve as a crawlable machine surface; do not add it to the sitemap. |
| The successful sitemap had not been read since 2026-05-26. | pending recrawl | Refresh once only after the revised production sitemap is verified live. |
| Core Web Vitals had no field-data sample. | external limitation | Do not describe missing data as a pass or failure. |
| GSC discovered all five URLs after the verified production sitemap refresh. | pending recrawl | Do not repeat the refresh; wait for a later Page indexing report to adjudicate the four new pages. |

## Action ledger

| Provider and property | Action and target | Accepted at | Confirmation | Result class | Repeat policy | Next review |
|---|---|---|---|---|---|---|
| Google Search Console `sc-domain:agentlink.run` | Refresh `https://agentlink.run/sitemap.xml` after the five-page production deployment | 2026-09-01 20:37 MDT | `Sitemap submitted successfully`; Status Success; submitted and last read 2026-09-01; five discovered pages | completed sitemap refresh; new-page indexing pending recrawl | Do not repeat before a material sitemap change or a named GSC failure. Do not request indexing or validate exclusions under this action. | Next Page indexing report update after 2026-09-01; checkpoint 2026-09-15 |
