# Changelog

All notable changes follow Keep a Changelog. Versions follow Semantic Versioning.

## [Unreleased]

### Fixed

- Strict preflight resolves requested knowledge paths and their allowed roots before containment checks, denying symlinked reads that leave the permitted knowledge directory.

## [0.11.0] - 2026-09-07

### Added

- Full and compact preflight context profiles with content-free payload size, character count, generation time, artifact count, and route count metrics.
- Standard and scoped workflow activation, with deterministic escalation for API, authorization, persistence, integration, configuration, and deployment changes.
- Direct `evidence-run` verification and typed `evidence-report` receipts bound to a content-sensitive worktree fingerprint, plus an installed JSON Schema.
- Source-attributed input, output, cached, and total token fields in the outcome evaluation harness while retaining the legacy total-token field.

### Changed

- Agent completion guidance now requires actual diff, post-edit validation, source/config/test attribution, and a documentation-impact decision; strict sessions can enforce freshness mechanically.

## [0.10.1] - 2026-09-07

### Changed

- Generalized opt-in strict Repository Knowledge preflight to Codex, Claude Code, Cursor, and Antigravity IDE through repeatable `--agent-preflight AGENT=strict` configuration. Each adapter now registers its native managed tool gate, emits its host-specific denial response, binds sessions to adapter and conversation, and receives strict gate checks in `doctor` and `doctor --live-hooks`.
- Added cross-host strict-preflight TDD evidence and a three-trial-per-host manual regression protocol; `--antigravity-preflight` remains a compatibility alias.

### Added

- Opt-in Antigravity `--antigravity-preflight strict` mode with an opaque, expiring activation session and a managed `PreToolUse` gate that blocks repository discovery, commands, writes, and subagents until routed documentation is activated.
- `preflight-activate` and `preflight-gate` commands, strict gate registration checks in `doctor`, and focused regressions for pending, active, expired, cross-repository, and shell-chained activation behavior.
- A repeatable neutral Google SSO Antigravity strict-preflight regression protocol requiring three successful traces before release.

## [0.10.0] - 2026-09-06

### Added

- Workflow-security regression coverage that rejects mutable action tags, branches, short commit references, and mutable Docker tags across every GitHub Actions workflow.
- Go 1.22.0 and Go 1.25.14 compatibility jobs that run ordinary checks and build both CLIs, with race tests isolated to Go 1.25.14.
- A separate neutral outcome-benchmark family with pinned fixture provenance, objective checks, blind semantic rubrics, allowed-change boundaries, expected behavioral traces, and an independently reproducible frontend control/treatment case.
- Immutable result recording for successful and failed trials, including run configuration, duration, optional token usage, explicit semantic review metadata, and a preserved patch or output artifact.
- README regression coverage for outcome-first section order, explicit product limits, and consistency between the published evidence statement and committed raw results.

### Changed

- GitHub Actions dependencies are pinned to verified full commit SHAs with exact release comments.
- Tag releases now verify and build with read-only repository access, transfer the release set through pinned artifact actions, and grant write, identity-token, and attestation permissions only to the dependent publishing job.
- `repo-knowledge-eval prepare` now requires explicit `control` or `treatment` conditions for outcome benchmarks while retaining the existing conformance workflow by default.
- Deterministic success remains semantically pending unless a reviewer explicitly supplies semantic status, score, and identity.
- The README now leads with the user outcome, a hostile stale-evidence example, an explicit no-results-yet measurement table, a three-command quickstart, and the product's non-guarantees; detailed mechanics route to maintained guides.

## [0.9.0] - 2026-09-06

### Added

- Native lifecycle preflight hooks for Codex, Claude Code, Antigravity IDE, and Cursor, installed together with each selected rule and skill adapter.
- Cross-host `hook-context` command that injects the installed contract, repository routing configuration, and documentation index using each host's native output protocol.
- Doctor checks and regression coverage for hook registration, shared-container preservation, adapter switching, malformed JSON, nested repository paths, and host-specific context output.

### Changed

- Agent rules now require an observable first-update preflight receipt and retain a manual loading fallback when native hook context is unavailable.
- Hook configuration is partially managed: install and update merge or remove only the toolkit command while preserving unrelated consumer settings and hooks.
- Structural scanning ignores a shared hook container only when it contains no consumer-owned configuration.

## [0.8.0] - 2026-08-26

### Added

- Reusable GitHub Actions documentation-impact workflow for pull requests and pushes in consuming repositories.
- Tested GitHub adapter release asset with committed-range resolution, initial-push handling, enforcement exit propagation, and persistent JSON reporting.
- Read-only consumer example with checksum, binary-version, and GitHub artifact-attestation verification.

### Changed

- Release packages now include `repo-knowledge-github-adapter.sh` in `SHA256SUMS` and provenance attestations.
- Installation, architecture, testing, extension, and top-level guidance now document GitHub Actions and GitLab CI as equivalent transports over the shared validator.

## [0.7.0] - 2026-08-26

### Added

- Permission-gated missing-binary recovery for installed skills on macOS, Linux, and Windows.
- Embedded bootstrap helpers that select only a pinned platform artifact, verify `SHA256SUMS`, verify GitHub attestations when supported, and install into a user-local PATH location.
- Offline regression coverage for confirmation, checksum verification, executable mode, embedded packaging, and canonical source metadata, plus a Windows runner dry-run for the PowerShell bootstrap plan.

### Changed

- Fresh installations now record the canonical GitHub repository URL instead of the ambiguous `release-binary` source label.
- Toolkit-managed bootstrap scripts are installed with executable permissions while all other managed assets retain their existing modes.
- Installation and architecture guidance now distinguishes repository asset installation from the separately approved user-level binary recovery flow.

## [0.6.1] - 2026-08-26

### Added

- GitHub CI for tests, race detection, vet, formatting, and both CLI builds.
- Attested GitHub release builds and a private vulnerability reporting policy.
- Apache License 2.0 terms for public use, modification, contribution, and distribution.

### Changed

- The Go module and public installation examples now use the canonical `github.com/rustedzone/repository-knowledge` source identity.

### Fixed

- Root-anchored binary ignores no longer exclude `cmd/repo-knowledge/main.go`, and Finder metadata is ignored.
- Version ordering now follows Semantic Versioning prerelease precedence, preventing a stable release from being treated as equal to its release candidate.

## [0.6.0] - 2026-08-25

### Added

- `install --all-agents` and `update --all-agents` select every canonical prompt adapter with one flag.
- Conflict validation rejects combining `--all-agents` with any explicit `--agent`, and tests cover fresh installation, update expansion, stable manifest ordering, help output, and the no-write error path.

### Changed

- Supported adapter names now have one ordered source used by validation and the all-agents preference, so future adapters can extend both consistently.
- Installation, operations, architecture, extension, and smoke-test guidance documents all-agents and retained-update semantics.

## [0.5.0] - 2026-08-25

### Added

- First-class Cursor support through an always-applied `.cursor/rules/repository-knowledge.mdc` project rule and `.cursor/skills/repository-knowledge/` Agent Skill.
- Cursor-aware installation, update retention and switching, `doctor` validation, scan exclusion, documentation-impact exclusion, evaluation preparation, and consumer-rule preservation tests.

### Changed

- Compatibility, installation, architecture, testing, and operations guidance now covers all four supported prompt adapters: Codex, Claude Code, Antigravity IDE, and Cursor.

## [0.4.4] - 2026-08-25

### Added

- Isolated frontend and layered-backend agent evaluation fixtures with deliberately stale prose, shallow documentation, and source-of-truth conflicts.
- A developer-only Go evaluation runner that prepares fixtures for Codex, Claude Code, or Antigravity IDE and grades objective documentation properties without executing an agent command.
- Semantic rubrics and trial-report guidance that keep implementation readiness, behavioral accuracy, and anti-hallucination review separate from deterministic keyword and artifact checks.

### Changed

- Generated overview and index guidance now includes a repository-specific newcomer mental model, reading path, and practical first action when supported by evidence.
- Mermaid guidance now recommends only a small number of evidence-backed component, sequence, state, or data-flow diagrams when they materially improve understanding; diagrams remain non-scored polish and never replace semantic traces.

## [0.4.3] - 2026-08-25

### Added

- Private coverage-ledger and trace-dossier workflow for each material capability or artifact lifecycle.
- Named behavioral-flow requirements with concrete symbols, state effects, side effects, error propagation, tests, and extension patterns.
- Stateful-model checks for transition writes and guards plus deletion analysis across database, application, and external synchronization behavior.
- Generation-mode remediation loop and explicit read-only assessment mode.

### Fixed

- Generation and rebuild requests no longer stop after declaring existing documentation insufficient or ask permission to begin the already-requested work.
- Layered backend documentation must explain concrete flow, dependency direction, trust boundaries, response/error handling, and recurring extension seams instead of cataloging packages and controllers.
- A single incidental code example no longer qualifies as coverage for unrelated architecture patterns.

## [0.4.2] - 2026-08-25

### Added

- Documentation quality gate covering unsupported speculation, mechanism-level capability depth, source-derived examples, and implementation readiness.
- Frontend-specific requirements for concrete navigation/guard behavior, state ownership, query/mutation lifecycle, UI composition, forms/validation, BFF behavior, and test patterns.
- Regression scenarios for speculative domain/integration prose, shallow frontend summaries, and onboarding examples.

### Fixed

- Domain glossaries and endpoint lists no longer qualify as completed domain documentation without verified rules, invariants, state transitions, permissions, side effects, failures, and tests where applicable.
- Unknown behavior must now be recorded as an exact evidence gap instead of being padded with “likely,” “typically,” or “expected” explanations.
- Full generation now requires concise current-source examples for recurring patterns that engineers need to copy or adapt.

## [0.4.1] - 2026-08-24

### Added

- Claim-specific source-of-truth guidance for dependencies, runtime/toolchain versions, commands, architecture, interfaces, behavior, data, configuration, security, integrations, deployment, product intent, and historical rationale.
- Regression scenarios for stale `CLAUDE.md`, README, route, command, and data-model claims.

### Fixed

- Documentation generation no longer treats agent instructions or existing prose as proof of current technical state when stronger manifest, lockfile, configuration, schema, test, or implementation evidence is available.
- Version documentation now distinguishes declared ranges, exact lockfile resolution, runtime selection, and corroborating source usage.
- Full generation reports stale prose outside the authorized documentation scope instead of copying it or rewriting it silently.

## [0.4.0] - 2026-08-24

### Added

- Multi-label repository classification and documentation coverage profiles for backend, frontend, mobile/desktop, libraries/SDKs, CLIs, infrastructure/GitOps, data/ML, monorepos, embedded systems, documentation/configuration repositories, and unusual repository shapes.
- A detailed page standard covering behavior, relationships, normal and failure flows, contracts, data lifecycle, operations, change guidance, tests, evidence, and uncertainty.
- Coverage planning and completion gates for every material domain, interface, data model, integration, job, package, and operational surface.

### Changed

- Full documentation generation now requires a comprehensive, evidence-proportional knowledge library rather than the smallest useful guide set.
- Repository shape is additive, allowing mixed repositories and monorepos to combine relevant documentation profiles.
- Bootstrap and installation guidance now make focused domain/API/data/integration pages and explicit known-gap reporting part of the expected result.

## [0.3.0] - 2026-08-24

### Added

- Evidence-backed semantic documentation workflow for requests to generate, improve, rebuild, or explain repository documentation.
- Audit warning when a repository still has no verified capability routes.
- Readable structural grouping and existing-document discovery in generated inventories.

### Changed

- `rebuild` now reports an explicit structural-inventory contract with `semantic_documentation_complete: false` and a semantic documentation next step.
- Generated inventory Markdown is labeled as a discovery appendix, expands common source containers, and omits machine fingerprints from human-facing output.
- The bootstrap index clearly identifies itself as incomplete until useful guides and routes are created.

## [0.2.0] - 2026-08-23

### Added

- Proactive Claude Code support through a project rule and `.claude/skills/repository-knowledge/`.
- Proactive Antigravity IDE support through a workspace rule and `.agents/skills/repository-knowledge/`.
- Repeatable multi-agent installation plus `claude` and `antigravity` aliases.

### Changed

- `update` now retains the installed adapter selection when `--agent` is omitted and safely reconciles explicitly changed adapter selections.
- `doctor` validates native rule and skill files for every selected adapter.

## [0.1.0] - 2026-08-23

### Added

- Initial repository knowledge contract and Codex adapter.
- Statically compiled Go CLI with embedded assets for install, update, scan, audit, rebuild, doctor, and documentation-impact validation.
- Cross-platform release builds, checksums, and GitLab Generic Package Registry publishing.
- GitLab CI include that downloads and verifies a pinned release binary for advisory, acknowledgment, and enforced modes.
- Bootstrap templates and example configurations.
