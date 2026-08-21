# Release checklist

Use this checklist for every Agentlink release. Preparation and publication are separate authority boundaries.

## Prepare

- [ ] Start from a clean branch whose changes have been reviewed against INTENT.md.
- [ ] Select a semantic version and add a dated CHANGELOG.md entry.
- [ ] Add `RELEASE_NOTES-X.Y.Z.md`.
- [ ] Propagate the version and date through README, PROJECT_CONTEXT.md, the landing page, sitemap, llms.txt, assistant guide, manifest, and release-contract tests.
- [ ] Keep root and served assistant-guide and manifest copies byte-identical.
- [ ] Run `sh scripts/check-release-contract.sh`.
- [ ] Run `sh scripts/prepare-release.sh X.Y.Z` to execute unit, race, vet, integration, deterministic cross-build, checksum, and host-binary version checks.
- [ ] Run `node scripts/check-search.mjs`.
- [ ] Run the accessibility workflow against the tracked `docs/` artifact with zero automated violations.
- [ ] Run `git diff --check` and confirm the working tree contains only the intended tranche.

## Review and merge

- [ ] Commit and push the approved branch.
- [ ] Open a pull request and wait for every required check on its exact HEAD.
- [ ] Review source, generated or duplicated trust surfaces, release scripts, and distribution metadata separately.
- [ ] Merge without bypassing a failed or unknown check.
- [ ] Fetch `main` and verify local HEAD equals `origin/main` before tagging.

## Publish

- [ ] Confirm the tag and GitHub Release do not already exist.
- [ ] Run `sh scripts/release.sh X.Y.Z` from clean, synchronized `main`.
- [ ] Record the annotated tag, GitHub Release URL, asset digests, and Homebrew tap commit.
- [ ] Do not rerun a completed external action. Resume from the first unverified step.

## Verify

- [ ] Run `sh scripts/verify-release.sh vX.Y.Z` against downloaded release assets.
- [ ] Install the tagged Go module into a clean temporary `GOBIN` and verify `agentlink --version`.
- [ ] Verify the Homebrew formula version, URLs, checksums, install, and version test.
- [ ] Wait for GitHub Pages to deploy the released commit.
- [ ] Run `node scripts/check-production-search.mjs`.
- [ ] Run the production accessibility scan and record axe-core and browser versions.
- [ ] Verify the canonical page, robots, sitemap, llms.txt, assistant guide, manifest, and a synthetic unknown route over live HTTP.
- [ ] Run the independent closing repository-operations scan.

Google Search Console and Bing console mutations are not release steps. They require separate property-bound authority after production passes.
