# Repository knowledge index

This is the routing map for the repository-knowledge toolkit itself.

## Architecture and contract

Read when:
- changing ownership boundaries, evidence precedence, policy, adapters, or lifecycle behavior;
- deciding whether logic belongs in the shared contract, an adapter, or a consuming repository.

Knowledge:
- [Architecture](architecture.md)
- [Configuration and contract](configuration.md)

Evidence anchors:
- `policy/`
- `schemas/`
- `assets.go`
- `internal/toolkit/`

## Installation and distribution

Read when:
- changing installed files, update behavior, version pinning, or GitHub/GitLab CI integration;
- testing the toolkit against another repository.

Knowledge:
- [Installation](installation.md)
- [Testing](testing.md)

Evidence anchors:
- `internal/toolkit/install.go`
- `internal/toolkit/hook_config.go`
- `internal/toolkit/hook_context.go`
- `cmd/repo-knowledge/main.go`
- `Makefile`
- `templates/`
- `.github/workflows/documentation-check.yml`
- `adapters/github/`
- `adapters/gitlab/`

## Agent behavior

Read when:
- changing how Codex, Claude Code, Antigravity IDE, or Cursor loads, verifies, assesses, scans, audits, rebuilds, or reconciles knowledge.

Knowledge:
- [Skill entrypoint](../skills/repository-knowledge/SKILL.md)
- [Human-readable documentation generation](../skills/repository-knowledge/references/documentation-generation.md)
- [Documentation quality and implementation readiness](../skills/repository-knowledge/references/documentation-quality.md)
- [Repository type profiles and coverage](../skills/repository-knowledge/references/repository-type-profiles.md)
- [Claim-specific sources of truth](../skills/repository-knowledge/references/source-of-truth.md)
- [Operation guidance](../skills/repository-knowledge/references/operations.md)
- [Missing-binary bootstrap](../skills/repository-knowledge/references/binary-bootstrap.md)

Evidence anchors:
- `skills/repository-knowledge/`
- `policy/contract.json`
- `templates/AGENTS.md`
- `templates/repository-knowledge-rule.md`
- `templates/repository-knowledge-cursor-rule.mdc`

## Agent evaluation

Read when:
- changing generation instructions or assessing whether an agent produces source-grounded, implementation-ready documentation;
- reproducing frontend or layered-backend documentation regressions across Codex, Claude Code, Antigravity IDE, or Cursor.
- comparing neutral control and treatment outcomes without conflating conformance with causal evidence.

Knowledge:
- [Agent evaluation workflow](../evals/README.md)
- [Toolkit testing](testing.md)

Evidence anchors:
- `evals/cases/`
- `evals/benchmarks/`
- `evals/results/`
- `internal/evalharness/`
- `cmd/repo-knowledge-eval/`

## Extension points

Read when:
- adding a new SCM adapter, deterministic detector, invariant, or agent integration.

Knowledge:
- [Extension guide](extension-guide.md)
