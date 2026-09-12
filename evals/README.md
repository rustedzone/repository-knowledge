# Repository Knowledge evaluations

The evaluation harness has two deliberately separate families. Neither family runs an agent automatically; agent authentication, model selection, permissions, and invocation remain outside the harness.

## Evaluation families

### Conformance evaluations

Conformance evaluations answer: **Does an agent follow Repository Knowledge correctly?**

The existing cases under `evals/cases/` explicitly ask the agent to use Repository Knowledge. Preparation installs the selected adapter, and the cases exercise source precedence, semantic depth, behavior tracing, examples, routing, and protected-source rules. They are regression tests for the product contract, not evidence of causal improvement over an unassisted agent.

| Case | Primary regressions |
| --- | --- |
| `frontend-nextjs` | Stale React version, static data mislabeled as live integration, route guards, state ownership, form validation, BFF behavior, and source-derived examples. |
| `backend-clean-architecture` | Package/controller cataloging, missing end-to-end trace, approval transitions, database versus application cleanup, trust boundaries, DI wiring, response conventions, and external synchronization. |

Strict preflight has a separate [manual regression protocol](strict-preflight.md). It uses a neutral Google SSO planning request and verifies the host hook/tool trace, which cannot be established by the fixture-based documentation grader alone. A strict-preflight release requires three passing fresh-task traces for each enabled host.

The existing workflow remains valid:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --case frontend-nextjs \
  --output /tmp/repository-knowledge-conformance-frontend \
  --agent codex

go run ./cmd/repo-knowledge-eval grade \
  --case frontend-nextjs \
  --target /tmp/repository-knowledge-conformance-frontend
```

### Outcome benchmarks

Outcome benchmarks answer: **Does Repository Knowledge improve the same agent's performance on the same neutral task?**

Cases under `evals/benchmarks/` use prompts that do not name Repository Knowledge. Every case records:

- a fixture and the full source commit containing it;
- a neutral task prompt;
- objective checks;
- a condition-blind semantic rubric;
- allowed-change boundaries;
- an expected behavioral trace for reviewers;
- a case revision.

The case loader rejects benchmark prompts that name Repository Knowledge and rejects rubrics that expose the control or treatment condition. Increment the case revision whenever the prompt, fixture, checks, rubric, allowed changes, or expected trace changes, and update the pinned source commit whenever the fixture changes.

List neutral benchmarks:

```bash
go run ./cmd/repo-knowledge-eval list --family benchmark
```

Prepare a control/treatment pair with the same agent, model, reasoning configuration, Repository Knowledge revision, and trial number:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition control \
  --output /tmp/frontend-onboarding-control-1 \
  --agent codex \
  --agent-version codex-desktop-2026.09 \
  --model-version gpt-5.6-sol \
  --reasoning high \
  --repository-knowledge-revision v0.11.1 \
  --trial 1

go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition treatment \
  --output /tmp/frontend-onboarding-treatment-1 \
  --agent codex \
  --agent-version codex-desktop-2026.09 \
  --model-version gpt-5.6-sol \
  --reasoning high \
  --repository-knowledge-revision v0.11.1 \
  --preflight-context full \
  --trial 1
```

Use the released version being evaluated, or replace `v0.11.1` with the exact candidate commit for an unreleased build. Keep that value identical across the pair.

`control` copies only the fixture into the target; it does not install policy, rules, skills, hooks, a toolkit manifest, or an in-target experiment marker. For both conditions, the harness writes baseline metadata to the adjacent `<target>.repo-knowledge-eval-baseline.json` sidecar. `treatment` copies the same fixture and installs only the requested adapter plus the shared toolkit assets that adapter requires. Treatment accepts `--preflight-context full` or `--preflight-context compact` and defaults to `full` for backward compatibility. Control and conformance preparation reject the option because Repository Knowledge is not the experimental input in those conditions.

The sidecar records the condition, case/source revisions, agent and host version, model version, reasoning configuration, Repository Knowledge revision, trial number, selected treatment profile, and content-free hook payload measurements. Hook payload measurements are bytes, characters, generation milliseconds, artifact count, and route count; they are not token estimates. Missing or invalid conditions and profiles fail before the output directory or sidecar is created.

Run the exact printed neutral prompt in each target. Do not show the benchmark rubric or expected trace to the agent. Randomize or alternate condition order when conducting multiple trials to reduce ordering effects.

## Grading and immutable results

The deterministic grader checks objective facts and protected files in both conditions. It cannot establish that prose is coherent, correct, or useful. A deterministic pass remains `pending_semantic_review`; it becomes a semantic pass only when a reviewer explicitly supplies a status, score, and identity.

After the run, preserve the changed output or patch as a regular file. For example:

```bash
tar -czf /tmp/frontend-onboarding-control-1.tar.gz \
  -C /tmp/frontend-onboarding-control-1 docs
```

Then grade and record the trial:

```bash
go run ./cmd/repo-knowledge-eval grade \
  --family benchmark \
  --case frontend-onboarding \
  --target /tmp/frontend-onboarding-control-1 \
  --duration 12m30s \
  --token-source codex \
  --input-tokens 15000 \
  --output-tokens 3420 \
  --cached-tokens 6000 \
  --total-tokens 18420 \
  --semantic-status pass \
  --semantic-score 16 \
  --semantic-available 18 \
  --reviewer reviewer@example.com \
  --results evals/results \
  --run-date 2026-09-06 \
  --artifact /tmp/frontend-onboarding-control-1.tar.gz
```

Omit the semantic flags until blind review is complete. Use only provider-reported token counts and name their source; do not convert characters, bytes, or elapsed time into estimated tokens. Cached tokens are recorded separately and remain part of provider input accounting, so `total_tokens` equals input plus output rather than input plus output plus cached. Use `--token-source unavailable` when the host exposes no usage, with no numeric token flags. The legacy `--tokens` total remains accepted for old automation. `--results` requires the date, duration, and preserved artifact. Control, full treatment, and legacy treatment records use:

```text
evals/results/<benchmark>/<agent>/<date>-<condition>-<trial>.json
evals/results/<benchmark>/<agent>/<date>-<condition>-<trial>-artifact.<ext>
```

Compact treatment results include the selected profile in the immutable filename:

```text
evals/results/<benchmark>/<agent>/<date>-treatment-compact-<trial>.json
evals/results/<benchmark>/<agent>/<date>-treatment-compact-<trial>-artifact.<ext>
```

This prevents matched full and compact results from colliding while preserving existing full-profile result paths.

Both files use exclusive-create semantics. Existing results cannot be overwritten through the harness. Commit every attempted trial, including deterministic or semantic failures; corrections use a new trial number rather than rewriting history.

Result metadata records condition, agent and host version, model version, reasoning configuration, Repository Knowledge version or commit, source commit, case revision, trial number, selected treatment profile, hook payload measurements, duration, source-attributed token usage when available, preserved artifact, deterministic checks, semantic score/status, and reviewer. Absolute disposable-target paths are removed from committed records.

Use [the evaluation report template](report-template.md) to compare a control/treatment pair. Compare paired outcomes rather than treating conformance success as product-effect evidence.

## Reliability and causal metrics

- For conformance, track `pass@1`, target `pass@3 >= 0.90`, and require deterministic `pass^3 = 1.00` for a regression-specific fix before release.
- For outcome benchmarks, compare control and treatment deterministic success, blind semantic score, duration, and token usage for the same case revision and run configuration.
- Do not pool trials across different source commits, case revisions, models, reasoning settings, or allowed-change boundaries without reporting the strata.
- Do not convert deterministic success into semantic success automatically.

## Full-versus-compact baseline protocol

Use this protocol to measure prompt-context changes separately from the control/treatment product experiment. Start with Codex and the released `v0.11.1` binary before refactoring prompt assets.

1. Run `frontend-onboarding` three times with treatment `full` and three times with treatment `compact`.
2. Repeat with `scoped-bugfix-plan`, which exercises ordinary scoped repository planning rather than documentation generation.
3. For each matched full/compact trial, pin the same case revision, fixture source commit, neutral prompt, agent and host version, model, reasoning configuration, permissions, Repository Knowledge revision, and trial number.
4. Use a fresh session for every run and alternate or randomize profile order. Do not let an earlier run's conversation or generated output enter a later target.
5. Record the hook payload metrics from preparation and the provider-reported input, output, cached, and total token counts from the end-to-end run. Record duration, deterministic outcome, blind semantic score, and every failed trial too.
6. Compare medians only within matching benchmark/run strata. A smaller hook payload demonstrates less injected text; only lower provider-reported input tokens demonstrate lower end-to-end input-token usage.

For example, create matched treatment targets with:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition treatment \
  --preflight-context full \
  --output /tmp/frontend-onboarding-full-1 \
  --agent codex \
  --agent-version <exact-host-version> \
  --model-version <exact-model-version> \
  --reasoning <exact-reasoning-setting> \
  --repository-knowledge-revision v0.11.1 \
  --trial 1

go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition treatment \
  --preflight-context compact \
  --output /tmp/frontend-onboarding-compact-1 \
  --agent codex \
  --agent-version <exact-host-version> \
  --model-version <exact-model-version> \
  --reasoning <exact-reasoning-setting> \
  --repository-knowledge-revision v0.11.1 \
  --trial 1
```

The harness prepares, measures, grades, and records trials but deliberately does not launch an agent or invent provider usage. Capture provider counts from the host's authoritative run report; when they are unavailable, record `--token-source unavailable` without numeric values.

The proposed optimization gate is at least 20% lower median provider-reported input-token usage than the v0.11.1 `full` baseline, with no deterministic regression, no decrease in semantic pass rate, no increase in unsupported or stale implementation claims, and no more than 10% median end-to-end latency regression. These are acceptance targets for later measurements, not current results or product claims.
