# Repository Knowledge evaluation report

## Trial identity

- Evaluation family: conformance/outcome benchmark
- Condition: conformance/control/treatment
- Date:
- Reviewer:
- Case and revision:
- Source commit:
- Agent adapter and host version:
- Model version:
- Reasoning configuration:
- Repository Knowledge version or commit:
- Trial number:
- Preflight context (`full`/`compact`, treatment only):
- Hook payload bytes:
- Hook payload characters:
- Hook generation milliseconds:
- Provider token source or unavailable:
- Provider-reported input tokens:
- Provider-reported output tokens:
- Provider-reported cached input tokens:
- Provider-reported total tokens:
- End-to-end duration:
- Preserved target or patch:

Hook payload size is not token usage. Copy payload measurements from the prepared baseline sidecar and token counts only from the named provider's authoritative run report.

## Results

- Deterministic status:
- Deterministic checks passed/failed:
- Semantic score / available points:
- Every critical rubric item scored 2: yes/no
- Semantic status: pass/fail
- Overall status: pass/fail

An overall pass requires both a deterministic pass and semantic acceptance. `pending_semantic_review` is not a pass.

For outcome benchmarks, score without revealing the condition. Record the condition only after the review is complete.

## Semantic findings

For every deducted rubric point, record the affected guide, the unsupported or missing claim, and the source evidence used to judge it.

| Rubric item | Score | Evidence and finding |
| --- | ---: | --- |
| | | |

## Regressions and observations

- Unsupported claims:
- Stale implementation claims:
- Missing or shallow flows:
- Incorrect source precedence:
- Unusable or inaccurate examples:
- Source or toolkit-managed files changed:
- Other observations:

## Decision

- Accepted or rejected:
- Reason:
- Follow-up issue or change:

## Paired outcome comparison

Complete this section for either a control/treatment pair or a treatment full/compact pair with the same case/source revision, agent host version, model version, reasoning configuration, permissions, task prompt, and trial number. Rename the two comparison columns when comparing profiles.

| Measure | Control or full | Treatment or compact | Difference |
| --- | ---: | ---: | ---: |
| Deterministic pass | | | |
| Semantic points | | | |
| Hook payload bytes | | | |
| Hook payload characters | | | |
| Hook generation time | | | |
| Provider input tokens | | | |
| Provider output tokens | | | |
| Provider cached input tokens | | | |
| Provider total tokens | | | |
| End-to-end duration | | | |
| Unsupported claims | | | |
| Stale implementation claims | | | |

- Condition order or randomization:
- Intended preparation differences verified: yes/no
- Same case/source revision, agent host, model, reasoning, permissions, prompt, and trial number verified: yes/no
- Causal interpretation and limitations:
