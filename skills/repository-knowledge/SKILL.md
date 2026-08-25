---
name: repository-knowledge
description: Generate and maintain human-readable, evidence-backed repository documentation; route repository assessment and implementation through installed knowledge; and reconcile documentation impact. Use for repository documentation, explanation, planning, change, scan, audit, bootstrap, or rebuild requests; do not use for general questions unrelated to the current repository.
---

# Repository Knowledge

Use the repository's knowledge layer to reduce rediscovery without allowing stale documentation to override stronger evidence.

## Load the contract

From the repository root, read these before making repository-level claims:

1. `.repo-knowledge/policy/contract.json`, or `policy/contract.json` when working on the toolkit itself.
2. `.repo-knowledge/repository.json` when installed.
3. `docs/index.md` when present.

Read local invariants or impact rules only when they affect the current request. Do not load every document. Use the index and prompt concepts to select the smallest relevant knowledge set, then verify important claims against corresponding repository evidence.

Existing prose—including `README.md`, `CLAUDE.md`, `AGENTS.md`, prior generated guides, and code comments—is a navigation and intent source, not automatic proof of current technical state. For dependency versions, runtime behavior, interfaces, configuration, data, deployment, and other current-state claims, verify against the claim-specific authoritative evidence described in [references/source-of-truth.md](references/source-of-truth.md). Agent instruction files describe how an agent should work; they are not authoritative for application framework versions or implementation behavior.

## Human-readable documentation requests

When the user asks to generate, rebuild, improve, or explain repository documentation, read [references/documentation-generation.md](references/documentation-generation.md) and follow its semantic documentation workflow. That workflow requires repository-type classification, claim-specific source-of-truth verification, implementation-level depth, source-derived examples, and evidence-proportional coverage; a brief overview is not sufficient for a repository with multiple material domains, interfaces, data models, integrations, or operational surfaces.

Preserve the request's action intent. `generate`, `rebuild`, `improve`, `complete`, or `fix` authorizes writing the requested repository documentation and requires continuing through the workflow's remediation loop. Do not stop after assessing that existing docs are insufficient or ask whether to begin the already-requested rebuild. `review`, `audit`, or `assess` alone is read-only: report gaps without modifying documentation. Pause only for a shared-contract human decision gate or missing authority outside the requested scope.

`scan` and `rebuild` produce discovery evidence and a structural inventory. They do not explain repository purpose, architecture, behavior, data flow, or operating model. Never claim a documentation request is complete merely because the generated inventory was applied. Completion requires factual, implementation-ready guides grounded in inspected code and tests, a verified end-to-end trace for each material executable capability, coverage of every material repository surface, practical examples for patterns engineers need to copy, populated routes in `docs/index.md`, and corresponding verified capability entries in `.repo-knowledge/repository.json`.

## Normal repository work

Before proposing a plan:

- identify the requested outcome and likely affected capabilities;
- load routed documentation for those capabilities;
- inspect high-signal implementation, schema, configuration, test, and decision evidence;
- distinguish verified facts, inference, requirements, history, uncertainty, and conflicts using the contract's claim states;
- surface a human decision gate only for the conditions named by the contract.

The skill supplies knowledge and impact assessment. It does not replace the coding, testing, security, deployment, or domain-specific capability needed to implement the change.

After implementation:

- validate the actual diff and behavior;
- reassess which claims and routes changed;
- update only meaningfully affected documentation;
- preserve requirement-only and historical content unless the user explicitly changes it;
- report either `documentation impact: required` with the reconciled files, or `documentation impact: not required` with a concrete reason.

For a committed no-impact decision, run:

```text
repo-knowledge acknowledge --impact not-required --reason "Concrete reason tied to this diff"
```

## Explicit operations

For scan, audit, rebuild, bootstrap, and conflict handling, read [references/operations.md](references/operations.md). Prefer the installed `repo-knowledge` binary for deterministic mechanics.

Before the first CLI-dependent operation, resolve `repo-knowledge` by command name. If the installed skill exists but the command is missing, read [references/binary-bootstrap.md](references/binary-bootstrap.md). Resolve the exact pinned release from repository evidence, preview the user-local destination, and ask for explicit permission before any network request, binary replacement, user-directory write, or PATH change. Never silently install a binary, select an unpinned `latest` release, use `sudo`, or edit a Unix shell profile. After approval, use the bundled verified bootstrap script and confirm `repo-knowledge --version` matches the pinned ref.

## Safety boundaries

- Treat repository documentation as routing knowledge, not unquestionable truth.
- Never silently promote `inferred`, `unverified`, or `unknown` claims to `verified`.
- Do not delete business requirements, external operational knowledge, or ADR history merely because code cannot prove them.
- Rebuild generated inventory with `--apply` only after reviewing the proposal; do not overwrite hand-maintained documentation.
- CI analyzes and validates. It does not silently commit generated documentation.
