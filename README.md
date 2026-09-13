# Repository Knowledge

Repository Knowledge helps coding agents avoid stale architectural assumptions and makes documentation-impact decisions visible.

It installs a shared evidence contract, repository-specific routing, and native lifecycle hooks for Codex, Claude Code, Antigravity IDE, and Cursor. Agents are directed to load the smallest relevant knowledge set, verify important claims against current source evidence, and reconcile documentation after implementation. Optional GitHub Actions and GitLab CI adapters make missing documentation-impact decisions visible when changes bypass an agent.

> **Evidence status:** no causal outcome-benchmark trial has been published yet. The repository contains a reproducible control/treatment harness, but it does not currently claim a measured improvement in agent success, unsupported claims, duration, or token usage. See the [benchmark case](evals/benchmarks/frontend-onboarding/eval.json), [evaluation protocol](evals/README.md), and [raw-results directory](evals/results/README.md).

**Try it:** install into a repository, run `doctor`, then start a new agent session. **Boundary:** Repository Knowledge makes evidence-first behavior cheaper and observable; it cannot guarantee that an agent complies with instructions or understands a repository correctly.

## 60-second hostile demonstration

The committed frontend fixture presents an agent with several plausible but contradictory signals:

| Signal | What the repository actually establishes |
| --- | --- |
| `CLAUDE.md` says React 18 | `package.json` declares React `^19.1.1`; `package-lock.json` resolves React `19.1.1`. |
| `CLAUDE.md` says intelligence metrics come from a live Pega API | `IntelligencePage` imports `src/data/intelligence.json` directly. |
| A protected route looks like ordinary page composition | `ProtectedLayout` calls `requireSession`, which redirects requests without `dashboard_session`. |
| A page is inside the protected layout | `IntelligencePage` still checks the `intelligence:read` permission through `useCheckPermission`. |

The neutral task asks an agent to improve onboarding documentation without naming Repository Knowledge. It must reject the stale prose, trace the session and permission boundaries, explain the form and BFF flow, preserve application files, and cite current evidence.

| Condition | Preparation | Published result |
| --- | --- | --- |
| Control | The fixture alone | Not run or published. |
| Treatment | The same fixture plus the selected Repository Knowledge adapter | Not run or published. |

The case definition, prompt, objective checks, blind rubric, and expected trace are committed under [`evals/benchmarks/frontend-onboarding/`](evals/benchmarks/frontend-onboarding/). This is a demonstration of a falsifiable test, not evidence that the treatment wins.

## Measured results

No outcome results are currently available. Consequently, no causal improvement is claimed.

| Primary measure | Control | Treatment | Observed difference |
| --- | ---: | ---: | ---: |
| First-attempt success (`pass@1`) | Not measured | Not measured | No claim |
| Unsupported architectural claims | Not measured | Not measured | No claim |
| Elapsed time | Not measured | Not measured | No claim |
| Token usage | Not measured | Not measured | No claim |

When trials exist, every successful and failed run must be stored with its source and case revisions, agent/model configuration, duration, token usage when available, deterministic result, blind semantic review, and preserved output artifact. Results use immutable paths under [`evals/results/`](evals/results/); the repository will link the corresponding raw records from this table instead of replacing them with a marketing summary.

Conformance results answer a different question—whether an agent follows the Repository Knowledge contract—and are not presented as causal product evidence.

## Three-command quickstart

First, download the pinned [v0.11.1 release](https://github.com/rustedzone/repository-knowledge/releases/tag/v0.11.1), verify its checksum and attestation, and put the executable on `PATH` as `repo-knowledge`. The [installation guide](docs/installation.md#release-artifacts) provides platform and verification details.

From the repository you want Codex to understand:

```bash
repo-knowledge install --target . --agent codex --source https://github.com/rustedzone/repository-knowledge --ref v0.11.1
repo-knowledge doctor --target .
codex
```

The third command starts a fresh Codex session so its project hook can load the repository preflight. Review or trust the hook if Codex asks. The first repository-related progress update should say `Repository knowledge preflight: loaded` and name the selected documentation routes.

Use `--agent claude-code`, `--agent antigravity-ide`, or `--agent cursor` for another host. Repeat `--agent` for a subset, or use `--all-agents` to install every supported adapter; then start a new session in the selected host.

## Honest boundary

Repository Knowledge makes correct repository behavior cheaper and observable; it does not guarantee agent compliance or understanding.

- Documentation is routing and intent evidence, not unquestionable truth. Current runtime behavior, tests, schemas, migrations, effective configuration, and implementation take precedence for their respective claims.
- Native hooks are activation guardrails. A host may require trust, may disable hooks, or may surface a failed hook without blocking the session. Any supported agent can opt into `--agent-preflight AGENT=strict` to deny repository tool calls until routed knowledge is activated; that still does not prove the model understood the material or prevent a pure-text response.
- Diff-bound evidence receipts reject stale checks and incomplete documentation decisions when invoked, but they are not yet universal host-level stop hooks and cannot prevent an agent from making an unsupported pure-text completion claim.
- `scan` and `rebuild` produce structural discovery data, not semantic documentation. An agent still has to inspect implementation and tests to explain behavior accurately.
- CI validates whether a material diff has documentation or an explicit impact decision. It does not determine semantic correctness, rewrite documentation, or commit changes.
- The current outcome benchmark has not been executed and does not support a performance claim.

## How it works

Repository Knowledge applies one versioned contract through two independent funnels:

```text
prompt -> native hook -> routed knowledge -> verified change -> documentation reconciliation
diff   -> CI adapter  -> documentation-impact validation -> report/pass/fail
```

The Go executable installs the shared policy, schemas, repository defaults, and only the selected agent adapters. It does not run as a daemon or inspect prompt transcripts. Toolkit updates replace toolkit-managed assets while preserving consumer-owned documentation, local rules, repository metadata, ADRs, scan state, and impact acknowledgments.

For documentation generation, ask the installed agent to inspect the complete repository and produce evidence-backed, implementation-ready guides. A scan or generated file inventory is only a discovery input. The [documentation-generation workflow](skills/repository-knowledge/references/documentation-generation.md) requires concrete behavioral traces, rules and failure paths, source-derived examples, a curated `docs/index.md`, and verified capability routes.

## Detailed documentation

- [Installation and agent selection](docs/installation.md) — binary verification, individual and all-agent installation, native hooks, updates, missing-binary recovery, and GitHub/GitLab CI.
- [Architecture](docs/architecture.md) — policy boundaries, ownership, lifecycle behavior, release trust, and CI enforcement.
- [Configuration and contract](docs/configuration.md) — repository metadata, local invariants, impact rules, and evidence precedence.
- [Testing](docs/testing.md) — Go compatibility, workflow-security checks, conformance evaluations, and outcome benchmarks.
- [Evaluation protocol](evals/README.md) — control/treatment preparation, blind semantic grading, immutable raw results, and causal limitations.
- [Extension guide](docs/extension-guide.md) — adding deterministic detectors, adapters, or policy extensions without crossing ownership boundaries.
- [Security policy](SECURITY.md) — supported versions and private vulnerability reporting.

## Command reference

| Command | Outcome |
| --- | --- |
| `install` | Install the shared bundle and selected prompt adapters; `--all-agents` selects every supported adapter. |
| `update` | Refresh toolkit-owned files while preserving consumer-owned knowledge. |
| `doctor` | Validate configuration, binary compatibility, managed files, and native hooks. |
| `scan` | Generate structural discovery state and capability leads. |
| `audit` | Report missing routes, broken index links, stale scans, and optional diff gaps. |
| `rebuild` | Propose or apply the optional generated structural inventory. |
| `impact` | Classify a diff and calculate its material-path fingerprint. |
| `acknowledge` | Record a documentation-impact decision bound to that fingerprint. |
| `validate-doc-impact` | Apply advisory, acknowledgment, or explicitly mapped enforcement. |
| `hook-context` | Emit full or compact bounded preflight context; `--metrics` emits content-free size and generation measurements. |
| `preflight-activate` | Activate a strict token with selected routes and the `standard` or `scoped` workflow. |
| `evidence-run` | Run an explicit validation command and bind its result digest to the current worktree content. |
| `evidence-report` | Require current successful checks, source evidence, and a documentation-impact decision before completion. |

See the [installation guide](docs/installation.md) for complete usage and the [architecture guide](docs/architecture.md) for ownership and safety boundaries.

## License

Repository Knowledge is licensed under the [Apache License 2.0](LICENSE).
