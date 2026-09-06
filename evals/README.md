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

Prepare a paired trial with the same agent, model, reasoning configuration, Repository Knowledge revision, and trial number:

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
  --repository-knowledge-revision v0.10.0 \
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
  --repository-knowledge-revision v0.10.0 \
  --trial 1
```

Use the released version being evaluated, or replace `v0.10.0` with the exact candidate commit for an unreleased build. Keep that value identical across the pair.

`control` copies only the fixture into the target; it does not install policy, rules, skills, hooks, a toolkit manifest, or an in-target experiment marker. For both conditions, the harness writes baseline metadata to the adjacent `<target>.repo-knowledge-eval-baseline.json` sidecar. `treatment` copies the same fixture and installs only the requested adapter plus the shared toolkit assets that adapter requires. The sidecar records the condition, case/source revisions, agent and host version, model version, reasoning configuration, Repository Knowledge revision, and trial number. Missing or invalid conditions fail before the output directory or sidecar is created.

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
  --tokens 18420 \
  --semantic-status pass \
  --semantic-score 16 \
  --semantic-available 18 \
  --reviewer reviewer@example.com \
  --results evals/results \
  --run-date 2026-09-06 \
  --artifact /tmp/frontend-onboarding-control-1.tar.gz
```

Omit the semantic flags until blind review is complete. `--results` requires the date, duration, and preserved artifact. It writes:

```text
evals/results/<benchmark>/<agent>/<date>-<condition>-<trial>.json
evals/results/<benchmark>/<agent>/<date>-<condition>-<trial>-artifact.<ext>
```

Both files use exclusive-create semantics. Existing results cannot be overwritten through the harness. Commit every attempted trial, including deterministic or semantic failures; corrections use a new trial number rather than rewriting history.

Result metadata records condition, agent and host version, model version, reasoning configuration, Repository Knowledge version or commit, source commit, case revision, trial number, duration, optional token usage, preserved artifact, deterministic checks, semantic score/status, and reviewer. Absolute disposable-target paths are removed from committed records.

Use [the evaluation report template](report-template.md) to compare a control/treatment pair. Compare paired outcomes rather than treating conformance success as product-effect evidence.

## Reliability and causal metrics

- For conformance, track `pass@1`, target `pass@3 >= 0.90`, and require deterministic `pass^3 = 1.00` for a regression-specific fix before release.
- For outcome benchmarks, compare control and treatment deterministic success, blind semantic score, duration, and token usage for the same case revision and run configuration.
- Do not pool trials across different source commits, case revisions, models, reasoning settings, or allowed-change boundaries without reporting the strata.
- Do not convert deterministic success into semantic success automatically.
