# Repository Knowledge

**Give coding agents a reliable path from documentation to code.**

Repository Knowledge gives Codex, Claude Code, Antigravity IDE, and Cursor a shared way to find relevant documentation, check important claims against current evidence, and account for documentation when code changes. A small Go CLI installs the contract and native agent hooks. Optional GitHub Actions and GitLab CI adapters check documentation-impact decisions on material diffs.

[Get started](#get-started) · [See the example](#a-repository-that-tries-to-mislead-the-agent) · [How it works](#how-it-works) · [Read the docs](#documentation)

> **Evidence status:** No outcome results have been published. The [benchmark](evals/benchmarks/frontend-onboarding/eval.json) is ready for control/treatment trials, but it does not establish a measured improvement.

## The problem it solves

An agent can read a convincing `CLAUDE.md`, then confidently document the wrong React version, invent a live API integration, or miss a permission check. More context alone does not settle which source is current.

Repository Knowledge gives the agent a route to the right files and an evidence order: use documentation to navigate, then verify behavior in code, tests, schemas, and effective configuration. It also makes the documentation decision visible at the end of a change.

| You need | Repository Knowledge provides |
| --- | --- |
| Relevant context without loading the whole repository | A documentation index, repository metadata, and native preflight hooks for the selected agent. |
| Claims you can trace | A shared evidence contract and a workflow for checking claims against their authoritative sources. |
| Documentation that follows code changes | Diff-bound impact decisions, with optional CI validation. |
| Safer updates | Toolkit-managed files can be refreshed while consumer-owned docs, metadata, local rules, ADRs, scan state, and impact acknowledgments are preserved. |

## Get started

Download the matching executable from the pinned [v0.11.1 release](https://github.com/rustedzone/repository-knowledge/releases/tag/v0.11.1), verify its checksum and GitHub attestation, and put it on `PATH` as `repo-knowledge`. The [installation guide](docs/installation.md#release-artifacts) has platform-specific instructions. You also need the selected agent installed.

From the repository you want Codex to work in:

```bash
repo-knowledge install --target . --agent codex --source https://github.com/rustedzone/repository-knowledge --ref v0.11.1
repo-knowledge doctor --target .
codex
```

Start a **new** agent session so the project hook can load its preflight. Review or trust the hook if the host asks. The agent's first repository-related update should say `Repository knowledge preflight: loaded` and name the documentation routes it selected.

For another host, set `--agent` to `claude-code`, `antigravity-ide`, or `cursor`, then open a new session in that host. Repeat `--agent` to select several hosts, or use `--all-agents` to install all supported adapters. See [agent selection and hook behavior](docs/installation.md#install-into-a-repository).

To ask for human-readable documentation after installation, try:

> Use repository-knowledge to document this repository. Trace each material capability through its entry point, rules, data or external boundaries, failure paths, and tests. Ground current-state claims in source evidence and route the resulting guides from `docs/index.md`.

`scan` and `rebuild` can help discover structure, but they do not write the semantic guides described in that request.

## A repository that tries to mislead the agent

The [frontend onboarding benchmark](evals/benchmarks/frontend-onboarding/eval.json) includes plausible instructions that conflict with the fixture's implementation:

| Tempting claim | What current evidence shows |
| --- | --- |
| [`CLAUDE.md`](evals/cases/frontend-nextjs/fixture/CLAUDE.md) says React 18. | [`package.json`](evals/cases/frontend-nextjs/fixture/package.json) declares React `^19.1.1`; the lockfile resolves `19.1.1`. |
| The same file says intelligence metrics come from a live Pega API. | [`IntelligencePage`](<evals/cases/frontend-nextjs/fixture/src/app/(protected)/intelligence/page.tsx>) imports static JSON. |
| A protected route appears to be ordinary page composition. | [`ProtectedLayout`](<evals/cases/frontend-nextjs/fixture/src/app/(protected)/layout.tsx>) calls `requireSession` before rendering. |
| The protected layout seems sufficient for access control. | The intelligence page also checks `intelligence:read` through `useCheckPermission`. |

The neutral task asks an agent to improve onboarding docs, trace the session and permission boundaries, explain the form and BFF flow, and preserve application files. The [prompt](evals/benchmarks/frontend-onboarding/prompt.md), [objective checks](evals/benchmarks/frontend-onboarding/eval.json), and [blind rubric](evals/benchmarks/frontend-onboarding/rubric.md) make the comparison reproducible.

There is no measured claim for first-attempt success, unsupported claims, elapsed time, or token use. The [evaluation protocol](evals/README.md) distinguishes contract conformance from causal outcomes; published trials will be linked from the [raw-results directory](evals/results/README.md).

## How it works

```text
agent task -> native hook -> routed knowledge -> source verification -> change -> documentation reconciliation
code diff   -> CI adapter  -> documentation-impact validation -> report
```

The versioned [contract](policy/contract.json) defines evidence precedence, claim states, ownership, and documentation-impact outcomes. The installed [skill](skills/repository-knowledge/SKILL.md) tells the agent how to apply that contract. The Go CLI handles deterministic installation, inspection, and validation; it does not run as a daemon or read prompt transcripts.

Native hooks load bounded repository context at session or invocation time. By default they are observational: host trust settings or disabled hooks can prevent them from running. An opt-in strict preflight can gate repository tool calls until the agent selects routes and activates a workflow. Diff-bound evidence receipts can verify that checks and documentation decisions match the current worktree when invoked. See the [architecture guide](docs/architecture.md#prompt-driven-lifecycle) for the exact host and enforcement boundaries.

The toolkit can help agents verify claims; it cannot guarantee that an agent reads, understands, or follows the evidence. CI checks whether a material diff has documentation or an explicit impact decision. It does not judge the semantic accuracy of the docs or commit changes for you.

## Documentation

| Guide | Start here for |
| --- | --- |
| [Installation](docs/installation.md) | Release verification, agent selection, native hooks, updates, and CI setup. |
| [Architecture](docs/architecture.md) | Contract boundaries, lifecycle behavior, ownership, and release trust. |
| [Configuration and contract](docs/configuration.md) | Repository metadata, local rules, and evidence precedence. |
| [Testing](docs/testing.md) | Go checks, security checks, conformance evaluation, and outcome benchmarks. |
| [Evaluation protocol](evals/README.md) | Control/treatment setup, blind review, and raw results. |
| [Extension guide](docs/extension-guide.md) | Adding detectors, adapters, or policy extensions. |
| [Security policy](SECURITY.md) | Supported versions and vulnerability reporting. |

## CLI at a glance

| Command | Purpose |
| --- | --- |
| `install`, `update`, `doctor` | Set up or refresh selected adapters, then check the installation. |
| `scan`, `audit`, `rebuild` | Discover repository structure, find knowledge gaps, and propose a generated inventory. |
| `impact`, `acknowledge`, `validate-doc-impact` | Classify changes and record or validate documentation-impact decisions. |
| `hook-context`, `preflight-activate` | Supply agent context and activate an optional strict session. |
| `evidence-run`, `evidence-report` | Bind checks, source references, and documentation decisions to current worktree content. |

The [installation guide](docs/installation.md) covers flags and workflows in detail.

## License

Apache License 2.0. See [LICENSE](LICENSE).
