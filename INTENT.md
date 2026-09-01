# Agentlink intent

## What this product is

Agentlink is a small Go CLI that keeps AI coding-tool instruction files synchronized by making one user-chosen file authoritative and linking each tool-specific alias to it with ordinary filesystem symlinks.

## Why it exists

AI coding tools read different filenames and global paths. Copying or regenerating the same instructions creates drift, hides ownership, and makes edits expensive. Agentlink keeps the mechanism inspectable: one real file, explicit configured links, no template engine, and no content transformation.

## Design invariants

1. One real source file owns instruction content.
2. Agentlink links files; it does not generate or transform their contents.
3. Dry-run must faithfully preview every filesystem mutation.
4. Regular files are replaced only through an explicit backup or force choice.
5. Directories and special files are never recursively replaced.
6. Broken or unrelated links are not removed when ownership cannot be established.
7. Errors preserve enough cause to distinguish invalid configuration, filesystem failure, unsupported layout, and external infrastructure.
8. Project and global configuration remain separate and explicit.
9. Release preparation is deterministic and mutation-free outside ignored build output; publication is a separate operation.
10. Upstream attribution remains visible in NOTICE, README, repository history, and public metadata.

## Scope boundaries

In scope: instruction-file discovery, explicit symlink configuration, synchronization, checking, cleanup of verified managed links, tool detection, repository scanning, and opt-in local hooks.

Out of scope: MCP configuration, prompt templating, content generation, content translation, secret management, remote execution, and ownership of the AI tools whose files Agentlink links.

## Conformance philosophy

N/A because Agentlink is a utility, not an open specification. Public claims are still treated as testable contracts and must have deterministic evidence or an explicit unautomated gap.

## Admission criteria for changes

N/A as an open-spec requirement. Product changes are admitted only when they preserve the design invariants, add discriminating tests for repeatable failure modes, keep documentation and release surfaces aligned, and pass the release checklist.

## Relationships to other PAICE standards

Agentlink publishes a GuideCheck assistant guide for bounded installation and repository-verification actions. It uses the portfolio accessibility and search-contract methods for its hosted site. These integrations do not expand Agentlink's runtime scope.

## Exceptions to Repo Standards

- `docs/llms-full.txt` is omitted because `docs/llms.txt` is a comprehensive standalone description of the single-page product, its concepts, files, lineage, install paths, and related projects. Review this exception if the site gains additional substantive pages.
- Agentlink is a standalone maintained continuation rather than a GitHub-native fork. Fork-tier lineage rules still apply through the upstream remote, NOTICE, README credit, and canonical module path.
- GitHub custom properties and repository-level action SHA-pinning enforcement are unavailable for this user-owned repository. Workflow source pins third-party actions instead.

## Changelog

- 2026-09-01: Clarified that Agentlink manages identical filesystem aliases while explicit real wrappers may layer tool-specific guidance; added integration metadata, config-aware scanning, and opt-in nested alias discovery without adding content generation.
- 2026-08-20: Added the first repo-scoped intent contract during the repository lifecycle pilot.
