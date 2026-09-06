# Release checklist

Use this checklist for every Agentlink release. Preparation and publication are separate authority boundaries.

## Prepare

- [ ] Start from a clean branch whose changes have been reviewed against INTENT.md.
- [ ] Select a semantic version and add a dated CHANGELOG.md entry.
- [ ] Add `RELEASE_NOTES-X.Y.Z.md`.
- [ ] Propagate the version and date through README, PROJECT_CONTEXT.md, the landing page, sitemap, llms.txt, assistant guide, manifest, and release-contract tests.
- [ ] Keep root and served assistant-guide and manifest copies byte-identical. Set the guide to active and its manifest to the prospective release-tag URL before preparation; independently verify that anchor after publication. A draft guide blocks release preparation.
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

## Recover a partial publication

Inspect the remote tag, GitHub Release asset inventory, and tap formula before writing anything. Record which step actually completed. `release.sh` intentionally refuses an existing tag; do not delete or move a published tag to bypass that guard.

- Tag exists but no release: verify that the tag resolves to the reviewed commit. Use a separate checkout of that exact commit and prepare its assets. With publication authorization, create only the missing GitHub Release using those assets and that tag. Preserve the original release notes and asset filenames.
- Release exists but an asset is missing: download existing assets and compare them with the exact-tag deterministic build. With authorization, upload only the missing asset. If an existing asset differs, stop and prepare a corrective release instead of clobbering it.
- Release and assets are complete but the Homebrew update failed: run the published-asset verifier first. In a clean checkout of the tap, update only `Formula/agentlink.rb`, using the four URLs and SHA-256 values from the published release. The formula template is in `scripts/release.sh`. Review the formula diff, run its install/version test, and commit and push only the tap change with authorization. If the formula already matches, do not make another commit.
- Tap and release are complete: resume the Verify checklist, including exact-commit Pages and live guide-anchor verification. Verification failure is not permission to republish completed steps.

For the verifier command, replace the release tag with the actual partially published tag.
Replace: RELEASE_TAG -> the existing release tag to verify
Customize
```sh
sh scripts/verify-release.sh RELEASE_TAG
```
