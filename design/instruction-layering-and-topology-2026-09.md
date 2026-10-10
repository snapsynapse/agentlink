# Instruction layering and topology diagnostics (2026-09)

Status: deferred design input, no build decision (ops/presentation-readiness-2026-09-23.md). Decision needed: build report-only diagnostics or decline.

Folded from two processed handoffs dated 2026-09-01 (hybrid instruction layering) and 2026-09-04 (skills pilot topology). The source-symlink wording correction already shipped in f867936 (PR #9). Nothing here is implemented.

## Concept

A future report-only capability, preferably `agentlink check --layering`, that recommends the least invasive way to expose shared `AGENTS.md` instructions to each harness while preserving tool-specific layers. Agentlink would never generate, parse or rewrite instruction content.

The working principle is:

1. Prefer native `AGENTS.md` support.
2. Use a symlink when two filenames should expose identical bytes.
3. Use a tool-native rules mechanism for tool-only additions when one exists and its loading behavior is documented.
4. Use a real importing wrapper at an exception point when shared plus local tool-specific content must occupy the tool's expected filename.
5. Fail closed when nested discovery or precedence is not documented.

Classifications: native, identical-alias, configured-native, real-wrapper-present, tool-rules-available, missing-alias, unknown-nested-support, conflict. Regular files are reported as "semantics unknown". Every nested or native-rule recommendation carries a public evidence URL. Keep three evidence levels separate: link identity, documented harness discovery, runtime-observed effective context (unverified unless actually observed).

## Boundaries
Agentlink continues to manage instruction-file aliases only. Skill-directory installation, manifests, tool adapters, prompt transformation, secret handling, wrapper generation, and cross-harness precedence normalization remain outside this work.
Do not infer that a regular file is a correct importing wrapper from its filename. Do not inspect private instruction contents or claim semantic loading from filesystem topology alone.

## Open design questions

1. Should the report live under `check --layering`, `scan --recommend`, or a new `recommend` command?
2. Should registry metadata describe native layer surfaces such as `.claude/rules`, or should the first version cover only native, configurable, import, and symlink choices already represented?
3. Is recommending `.claude/rules` within the instruction-file scope if Agentlink never writes those files?
4. Should a regular expected filename always be classified as an intentional wrapper, or as "regular file, semantics unknown"?
5. Should optional recommendations affect exit status, or should only conflicts fail?
6. Does the report need machine-readable JSON from its first release?

Recommended defaults: `check --layering`; metadata-backed native-layer recommendations; read-only recommendations are in scope; call regular files "semantics unknown"; optional recommendations exit zero; defer JSON until a consumer is identified.

## Deterministic acceptance checks for a future implementation

- Native `AGENTS.md` integrations never recommend a self-link.
- Identical missing aliases are recommended only when the registry supports that repository level.
- Existing correct symlinks are reported as satisfied.
- Existing real files are preserved and reported without content claims.
- Explicit configuration suppresses undeclared root and nested recommendations.
- Unknown nested behavior produces `unknown-nested-support`, never an alias recommendation.
- Every nested or native-rule recommendation includes a valid public evidence URL.
- Dry-run and report-only modes leave a full-tree manifest byte-identical.
- Human and machine output, if both exist, are stable under repeat runs.
- README, landing page, `llms.txt`, help text, registry metadata, and tests remain aligned.

## Topology acceptance cases (skills pilot)

1. A real source and correct relative alias report satisfied without mutation.
2. A symlink source reports invalid-source with an actionable explanation; force is not silently selected.
3. A preserved real wrapper reports semantics unknown, with unchanged bytes.
4. Explicit configuration suppresses undeclared automatic alias recommendations.
5. A global source and separate repo-local rules are shown as separate scopes; no combined effective context is claimed.
6. Unknown nested support remains unknown and produces no invented alias recommendation.
7. A full-tree manifest is identical before and after reporting, including dry-run with force/backup flags.
8. Text, JSON if admitted, help, README, and contract tests agree on statuses and source-symlink behavior.
