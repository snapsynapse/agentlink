# Agentlink v0.4.2

Agentlink v0.4.2 completes the repository lifecycle and publication path without changing the CLI's runtime behavior.

## What changed

- Release preparation is now deterministic, locally verifiable, and separate from publication.
- One release-target inventory drives builds, GitHub assets, verification, and Homebrew checksums.
- The landing page has a zero-violation accessibility gate and fixes inline links that previously relied on color alone.
- Offline and production search contracts verify the canonical page, sitemap, robots policy, discovery files, redirects, and 404 behavior.
- GuideCheck review and served trust anchors are tracked as byte-identical root and `docs/.well-known/` pairs.
- Repository-owned intent, release, GitHub desired-state, Dependabot, and lifecycle evidence make future changes resumable and inspectable.
- Third-party GitHub Actions are pinned to immutable commits.

## Runtime compatibility

There are no intentional CLI, configuration, or symlink-behavior changes in this patch release. Go 1.23 or newer remains required for source installation.

## Verification

The release gate runs unit, race, vet, integration, formatting, static analysis, spelling, vulnerability, action-semantic, search, accessibility, deterministic cross-build, checksum, host-binary version, clean consumer, and production checks. Published binaries target darwin and linux on amd64 and arm64.
