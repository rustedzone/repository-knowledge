# Architecture

## Product boundary

The product is the versioned repository knowledge contract. The Codex, Claude Code, Antigravity IDE, Cursor, GitHub Actions, and GitLab adapters consume the same contract:

```text
shared contract + schemas + defaults
              |
       +------+------+
       |             |
  prompt adapter   SCM/CI adapter
       |             |
 native rule + skill   Go binary over Git diff
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
| `internal/toolkit/` | Installation, inventory, validation, impact, audit, and rebuild mechanics. |
| `templates/` | Initial consumer-owned files, native agent rules, and the managed Codex `AGENTS.md` block. |
| `adapters/` | Tested GitHub and GitLab SCM/runtime glue that invokes the released binary. |
| `.github/workflows/documentation-check.yml` | Public reusable workflow that supplies GitHub checkout, release verification, permissions, and report upload. |
| `evals/` | Disposable agent-behavior fixtures, prompts, deterministic expectations, and semantic rubrics. |
| `examples/` | Concrete adoption configurations, not alternative policy sources. |

## Ownership and upgrades

Installation records every toolkit-owned file and digest in `.repo-knowledge/toolkit.json`. `update` replaces installed skills, adapter rules, policy snapshots, schemas, and the managed Codex `AGENTS.md` block. It does not replace `.repo-knowledge/repository.json`, consumer rules, documentation, scan state, rebuild proposals, or impact decisions. Without `--agent` or `--all-agents`, update retains the adapters recorded in the existing manifest. Explicit `--agent` values select a subset; `--all-agents` resolves to the complete canonical adapter list before the same install/update reconciliation runs.

Consumers pin a release tag. They update by running the new binary and changing the GitHub workflow tag and toolkit input or GitLab include and package version in the same merge request. GitHub tag builds publish checksummed, attested binaries and the GitHub adapter; the GitLab pipeline remains available for registry-based distribution. The binary follows Semantic Versioning precedence and refuses a version downgrade, including a release-to-prerelease downgrade, unless `--allow-downgrade` is explicit.

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

The Go process does not intercept prompts. Installation creates the proactive adapter:

```text
user prompt
    -> selected agent loads its native project rule
       Codex: AGENTS.md managed block
       Claude Code: .claude/rules/repository-knowledge.md
       Antigravity IDE: .agents/rules/repository-knowledge.md
       Cursor: .cursor/rules/repository-knowledge.mdc
    -> native rule routes repository work to the installed skill
    -> skill loads contract, repository metadata, and docs/index.md
    -> agent verifies evidence and either:
       - performs repository work and reconciles documentation impact, or
       - classifies repository shapes and writes detailed profile-specific guides and verified routes for a documentation request
```

This keeps prompt semantics in the agent adapter while deterministic mechanics remain reusable by agents, developers, and CI.

Evidence selection is claim-specific as well as precedence-based. Existing prose and agent instruction files route discovery and preserve intent, but they do not establish current technical state. The skill directs agents to manifests and lockfiles for dependency declarations and resolution, effective configuration for runtime selection, tests and implementation for behavior, schemas and migrations for data, and deployment definitions for declared operations. Objective prose mismatches become `verified_stale`; disagreement between authoritative current sources remains a `conflict` requiring the contract's normal handling.

Semantic completion is quality-gated in the skill rather than the Go runtime. Generated guides must connect implementation mechanisms across triggers, orchestration, authorization/validation, state, boundaries, errors, and tests. The skill rejects unsupported speculative filler and requires concise source-derived examples for recurring change patterns. Unknowns remain explicit gaps, while the implementation-readiness review checks whether a newcomer can locate and execute a representative change without rediscovering basic architecture from source.

Full generation maintains a working coverage ledger and trace dossiers for material capabilities. The final semantic pass compares those dossiers with every guide and loops back into source inspection and revision when coverage fails. This remediation loop applies only to mutating generation, rebuild, improvement, completion, or fix requests; review and assessment requests preserve read-only behavior.

## Agent evaluation boundary

`cmd/repo-knowledge-eval` and `internal/evalharness` are developer-only test surfaces. The runner copies an isolated fixture, installs the current embedded skill for the chosen adapter, records protected-file hashes, and prints the prompt. It deliberately does not invoke an agent CLI: authentication, model selection, permissions, and interactive behavior belong to each host.

After a trial, the runner checks objective facts such as required artifacts, known source-of-truth claims, evidence anchors, attributed examples, valid links, verified route metadata, and preservation of source. Those checks are necessary but not sufficient. Behavioral correctness, explanation quality, and implementation readiness are reviewed with the case rubric, and an objective pass remains pending until that semantic review is recorded. The evaluation runner is not embedded in or shipped as part of the consumer `repo-knowledge` release binary.

## Safe rebuild and enforcement

`rebuild` always writes a structural-inventory proposal with `semantic_documentation_complete: false`. `rebuild --apply` writes only the configured generated appendix and refuses to replace a file without generated markers. It never claims to generate semantic documentation. Human-readable guides remain agent-authored, evidence-backed, and consumer-owned. The skill treats repository shape as multi-label and combines relevant coverage profiles so backend, frontend, mobile, library, CLI, infrastructure, data/ML, monorepo, embedded, documentation/configuration, and unusual repositories receive appropriate depth without a universal empty taxonomy. Consumer-controlled paths and managed manifest paths are constrained to remain within the target repository.

- `advisory`: report missing impact decisions without failing.
- `acknowledgment`: fail when a material diff has neither documentation changes nor a matching acknowledgment.
- `enforced`: additionally apply consumer-configured classification-to-document mappings.

Path classification remains a routing heuristic and is not a semantic source of truth.
