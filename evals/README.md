# Repository Knowledge agent evaluations

These evaluations measure the prompt-driven semantic documentation workflow, not the deterministic `repo-knowledge scan` or structural-inventory output.

## Evaluation model

Each case contains:

- an isolated fixture repository with deliberately stale or misleading discovery signals;
- the same repository-neutral generation prompt a consumer can use;
- deterministic checks for objective facts and artifact integrity;
- a semantic rubric for behavior, depth, actionability, and honesty.

The evaluator separates code-based results from human/model review. Keyword checks can establish that a required source path or known fact appears, but they cannot prove that an architectural explanation is coherent.

## Cases

| Case | Primary regressions |
| --- | --- |
| `frontend-nextjs` | Stale React version, static data mislabeled as live integration, route guards, state ownership, form validation, BFF behavior, and source-derived examples. |
| `backend-clean-architecture` | Package/controller cataloging, missing end-to-end trace, approval transitions, database versus application cleanup, trust boundaries, DI wiring, response conventions, and external synchronization. |

## Workflow

Prepare a disposable repository:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --case frontend-nextjs \
  --output /tmp/repository-knowledge-eval-frontend \
  --agent codex
```

The command installs the current skill, records protected fixture hashes, and prints the case prompt. Run the selected agent in that disposable repository, then grade the result:

```bash
go run ./cmd/repo-knowledge-eval grade \
  --case frontend-nextjs \
  --target /tmp/repository-knowledge-eval-frontend
```

Use `--json` for machine-readable output. The deterministic grade fails when a required check fails. Complete the case's `rubric.md` separately; semantic acceptance requires every critical item to score 2 and at least 85% of all available points.

Record every run with [the trial report template](report-template.md). Keep the prepared target or an immutable patch/artifact with the report so another reviewer can reproduce the score. Never promote `pending_semantic_review` to pass from the deterministic output alone.

## Reliability metrics

- Track `pass@1` as the first-attempt success rate across cases and agents.
- Target capability reliability: `pass@3 >= 0.90`.
- For a skill change fixing a known regression, require deterministic `pass^3 = 1.00` on the relevant case before release.
- Record model, agent adapter, skill version, case revision, duration, deterministic result, rubric score, and reviewer.

The harness does not invoke an agent automatically. Agent CLIs and authentication differ across Codex, Claude Code, Antigravity IDE, and Cursor; keeping preparation and grading separate makes the same fixture usable across all four without executing arbitrary configured commands.
