# Architecture

## Product boundary

The product is the versioned repository knowledge contract. The Codex, Claude Code, Antigravity IDE, Cursor, GitHub Actions, and GitLab adapters consume the same contract:

```text
shared contract + schemas + defaults
              |
       +------+------+
       |             |
  agent adapter    SCM/CI adapter
       |             |
 hook + rule + skill   Go binary over Git diff
```

The prompt funnel loads routed knowledge before assessment, verifies material claims, and reconciles documentation after implementation. The SCM funnel sees changes that bypass agents and validates whether the diff contains a documentation change or an explicit, diff-bound impact decision.

## Binary architecture

`assets.go` embeds `VERSION`, policy, schemas, templates, and the shared agent skill—including its missing-binary bootstrap helpers—using `go:embed`. `cmd/repo-knowledge` owns argument parsing and human or JSON output. `internal/toolkit` owns deterministic repository operations. It uses only the Go standard library and the Git executable.

A release produces static binaries for Linux, macOS, and Windows, the executable GitHub adapter, `LICENSE`, and `SHA256SUMS`. The executable installs embedded assets but does not copy itself into the consumer repository. Developers and CI runners obtain the appropriate released binary independently and retain the license when redistributing it.

The installed skill closes the cloned-repository gap without making installation implicit. When a required command is absent, agent instructions resolve a pinned source and ref from the managed manifest, run the embedded platform helper only after a user-visible plan and explicit permission, verify the checksum and available GitHub attestation, install into a user-local PATH location, and confirm the resulting version. Unix helpers never edit shell startup files or elevate privileges; the Windows helper changes the user PATH only when that mutation was included in the approval.

## Component responsibilities

| Component | Responsibility |
| --- | --- |
| `policy/` | Runtime-neutral evidence, confidence, safety, ownership, invariant, and impact defaults. |
| `schemas/` | Stable machine-readable interfaces for consumer configuration and acknowledgments. |
| `skills/` | Shared prompt routing, semantic decision behavior, and permission-gated missing-binary recovery helpers for supported agents. |
| `assets.go` | Compile toolkit-managed resources into the release binary. |
| `cmd/` | CLI interface and output contracts. |
| `internal/toolkit/` | Installation, native-hook context and merging, inventory, validation, impact, audit, and rebuild mechanics. |
| `templates/` | Initial consumer-owned files, native agent rules, and the managed Codex `AGENTS.md` block. |
| `adapters/` | Tested GitHub and GitLab SCM/runtime glue that invokes the released binary. |
| `.github/workflows/documentation-check.yml` | Public reusable workflow that supplies GitHub checkout, release verification, permissions, and report upload. |
| `evals/` | Disposable agent-behavior fixtures, prompts, deterministic expectations, and semantic rubrics. |
| `examples/` | Concrete adoption configurations, not alternative policy sources. |

## Ownership and upgrades

Installation records every fully toolkit-owned file and digest in `.repo-knowledge/toolkit.json`. `update` replaces installed skills, adapter rules, policy snapshots, schemas, and the managed Codex `AGENTS.md` block. Native hook containers are shared files: the toolkit merges or removes only its exact `repo-knowledge` hook entries and preserves unrelated keys and hooks. It does not replace `.repo-knowledge/repository.json`, consumer rules, documentation, scan state, rebuild proposals, or impact decisions. Without `--agent` or `--all-agents`, update retains the adapters recorded in the existing manifest. Explicit `--agent` values select a subset; `--all-agents` resolves to the complete canonical adapter list before the same install/update reconciliation runs.

Consumers pin a release tag. They update by running the new binary and changing the GitHub workflow tag and toolkit input or GitLab include and package version in the same merge request. GitHub tag builds publish checksummed, attested binaries and the GitHub adapter; the GitLab pipeline remains available for registry-based distribution. The binary follows Semantic Versioning precedence and refuses a version downgrade, including a release-to-prerelease downgrade, unless `--allow-downgrade` is explicit.

## CI and release trust boundary

Every external GitHub Action used by the toolkit workflows is pinned to a full commit SHA with its exact release version recorded in a comment. A repository test scans every workflow and rejects mutable tags, branches, short SHAs, and mutable Docker tags. This turns action immutability into a maintained invariant rather than a review convention.

Ordinary tests, vet, formatting, and both CLI builds run on Go 1.22.0—the minimum declared by `go.mod`—and Go 1.25.14, the release toolchain. Race-enabled tests run on Go 1.25.14 only.

Tag releases cross a two-job privilege boundary. The build job has read-only contents access, checks the tag, runs the source checks, builds the release set, and uploads a short-lived immutable workflow artifact. The dependent publish job downloads that exact set and is the only job granted `contents: write`, `id-token: write`, and `attestations: write`; it attests and publishes without checking out or rebuilding source. The transfer actions are SHA-pinned under the same workflow invariant.

## CI adapter lifecycle

The GitHub and GitLab integrations resolve platform-specific CI metadata but converge before enforcement:

```text
pull request or push
    -> full-history checkout
    -> explicit base and head Git object IDs
    -> pinned release download and SHA-256 verification
    -> GitHub: artifact-attestation verification
    -> repo-knowledge validate-doc-impact
    -> JSON report artifact and advisory/pass/fail job status
```

The GitHub reusable workflow has read-only contents and attestation permissions, does not persist checkout credentials, and requires no secrets. Its shell adapter is shipped as a release asset so the exact executed logic is covered by `SHA256SUMS` and release provenance. GitLab retains its package-registry and job-token transport. Neither adapter writes or commits repository documentation.

## Prompt-driven lifecycle

The Go process does not run as a daemon or parse prompt transcripts. Installation creates a native lifecycle preflight plus a portable rule fallback:

```text
agent session or invocation
    -> selected host runs repo-knowledge hook-context
       Codex: SessionStart in .codex/hooks.json
       Claude Code: SessionStart in .claude/settings.json
       Antigravity IDE: PreInvocation in .agents/hooks.json
       Cursor: sessionStart in .cursor/hooks.json
    -> hook resolves the repository root from .repo-knowledge/toolkit.json
    -> hook injects the configured full or compact routing context
    -> native project rule routes repository work to the installed skill and supplies a manual fallback
    -> agent reports Repository knowledge preflight: loaded and the selected routes
    -> skill loads contract, repository metadata, and docs/index.md
    -> agent selects standard or scoped work, verifies evidence, and either:
       - performs repository work and reconciles documentation impact, or
       - classifies repository shapes and writes detailed profile-specific guides and verified routes for a documentation request
```

Codex and Claude Code receive text context, Cursor receives `additional_context`, and Antigravity receives an `injectSteps` payload. The default `full` profile includes bounded contract, configuration, and index content; `compact` includes the routing/safety summary and bounded index while requiring disk reads for selected evidence. Antigravity gets the configured preflight on the first invocation and a bounded reminder later. The injected consumer-owned configuration and documentation are explicitly labeled as evidence rather than executable instructions. Content-free metrics expose payload bytes, characters, generation milliseconds, artifact count, and route count; they do not persist the payload or estimate model tokens.

Every supported adapter can additionally use opt-in `strict` preflight mode. A 30-minute pending session is bound to the repository, adapter, host conversation, and digests of the installed contract, repository configuration, and documentation index. The injected activation command records selected documentation routes. Managed gates use each host's native `PreToolUse` protocol: Codex and Claude Code return a hook permission decision, Cursor returns a hook permission value, and Antigravity returns its gate decision. While pending, the gate allows only knowledge reads, supported external research tools, and the exact activation command; it denies repository discovery, commands, writes, and subagents until activation. The session state lives outside the consumer repository in an owner-only temporary runtime directory and contains hashes and operational metadata, not prompt or documentation content. A changed knowledge input or expired session returns the task to pending.

Activation also records `standard` or `scoped` workflow intent. Scoped work is an advisory fast path, not reduced correctness: completion escalates deterministically when the actual diff is classified as API, authorization, persistence, integration, configuration, or deployment work. In strict sessions, `evidence-run` directly executes an explicit argument vector and stores only its label, arguments, timestamps, exit code, output digest, and content-sensitive diff fingerprint. `evidence-report` requires at least one successful verification matching the final fingerprint, existing source/config/test references, and either changed documentation or a concrete no-impact reason. This receipt is stronger than the path-only documentation-impact fingerprint, but it still proves execution and attribution—not semantic correctness.

This keeps behavior semantics in the agent adapter while deterministic context loading remains reusable and testable. Strict mode is an enforcement boundary for the host tool calls that reach its local hook; it is not proof that the model read or understood the material, and an agent can still produce a pure-text answer or use hosted tools outside a local hook path. Project hook trust and enablement remain under the host, a missing `repo-knowledge` executable causes a visible hook failure, and hosts may fail open. The native rule still directs a manual preflight, `doctor` detects missing registrations, and the first-update receipt makes activation observable. Binary recovery remains permission-gated and is attempted only when a task needs a CLI operation.

The evidence receipt is an agent/user-invoked completion check in this release, not a universal native stop hook. This keeps one deterministic contract across hosts without pretending that every host exposes the same completion-blocking lifecycle. A future stop integration should be added per adapter only when the host can reliably pass the active conversation identity and honor a blocking response; until then, pure-text completion remains outside the runtime enforcement boundary.

Evidence selection is claim-specific as well as precedence-based. Existing prose and agent instruction files route discovery and preserve intent, but they do not establish current technical state. The skill directs agents to manifests and lockfiles for dependency declarations and resolution, effective configuration for runtime selection, tests and implementation for behavior, schemas and migrations for data, and deployment definitions for declared operations. Objective prose mismatches become `verified_stale`; disagreement between authoritative current sources remains a `conflict` requiring the contract's normal handling.

Semantic completion is quality-gated in the skill rather than the Go runtime. Generated guides must connect implementation mechanisms across triggers, orchestration, authorization/validation, state, boundaries, errors, and tests. The skill rejects unsupported speculative filler and requires concise source-derived examples for recurring change patterns. Unknowns remain explicit gaps, while the implementation-readiness review checks whether a newcomer can locate and execute a representative change without rediscovering basic architecture from source.

Full generation maintains a working coverage ledger and trace dossiers for material capabilities. The final semantic pass compares those dossiers with every guide and loops back into source inspection and revision when coverage fails. This remediation loop applies only to mutating generation, rebuild, improvement, completion, or fix requests; review and assessment requests preserve read-only behavior.

## Agent evaluation boundary

`cmd/repo-knowledge-eval` and `internal/evalharness` are developer-only test surfaces with two evaluation families. Conformance cases retain explicit Repository Knowledge prompts and always install the selected adapter; they answer whether the product contract is followed. Outcome benchmarks use neutral prompts and require an explicit `control` or `treatment` condition; they answer whether the same agent performs better with the product than without it. Conformance outcomes are not treated as causal evidence.

For an outcome pair, both targets copy the fixture pinned by case revision and source commit. Control leaves the target as the fixture with no toolkit or experiment paths. Treatment installs only the selected adapter and its required shared assets. The harness writes each condition's baseline to an adjacent sidecar, outside the agent's target, recording condition, agent host version, model version, reasoning configuration, Repository Knowledge revision, and trial number. Missing or invalid conditions fail before target or sidecar creation, and the blind rubric and expected trace are not printed with the neutral task.

After a trial, the runner checks objective facts such as required artifacts, known source-of-truth claims, evidence anchors, attributed examples, valid links, and preservation of source. Those checks are necessary but not sufficient. Deterministic success remains pending until a reviewer explicitly supplies semantic status, score, and identity. Immutable result recording preserves both failed and successful trials plus a patch or output artifact under `evals/results/`; exclusive creation prevents accidental replacement. The runner does not invoke an agent CLI and is not embedded in or shipped as part of the consumer `repo-knowledge` release binary.

## Safe rebuild and enforcement

`rebuild` always writes a structural-inventory proposal with `semantic_documentation_complete: false`. `rebuild --apply` writes only the configured generated appendix and refuses to replace a file without generated markers. It never claims to generate semantic documentation. Human-readable guides remain agent-authored, evidence-backed, and consumer-owned. The skill treats repository shape as multi-label and combines relevant coverage profiles so backend, frontend, mobile, library, CLI, infrastructure, data/ML, monorepo, embedded, documentation/configuration, and unusual repositories receive appropriate depth without a universal empty taxonomy. Consumer-controlled paths and managed manifest paths are constrained to remain within the target repository.

- `advisory`: report missing impact decisions without failing.
- `acknowledgment`: fail when a material diff has neither documentation changes nor a matching acknowledgment.
- `enforced`: additionally apply consumer-configured classification-to-document mappings.

Path classification remains a routing heuristic and is not a semantic source of truth.
