# Scoped bug-fix planning outcome rubric

Score every item 0, 1, or 2:

- `0`: absent, incorrect, speculative, or outside the allowed scope;
- `1`: directionally correct but incomplete or weakly grounded;
- `2`: accurate, minimal, actionable, and supported by current repository evidence.

Critical items are marked **critical**. Passing requires every critical item to score 2 and at least 85% overall. Review the output without being told how the target was prepared.

| Item | Critical | Review question |
| --- | --- | --- |
| Root-cause localization | Yes | Does the plan locate asynchronous submission in `UserForm` and identify Formik submission state as the relevant existing mechanism? |
| Minimal implementation | Yes | Does it propose disabling repeat submission through the existing form state without introducing unrelated architecture or dependencies? |
| Verification | Yes | Does it identify focused pending and completed submission behavior to test in the existing component-test area? |
| Scope control | Yes | Does it preserve application source during planning and avoid API, authorization, persistence, configuration, or deployment changes? |
| Documentation impact | No | Does it make a concrete, defensible documentation-impact decision for this bounded behavior change? |
| Evidence and honesty | Yes | Are repository claims tied to current paths and symbols without relying on stale prose or invented behavior? |

Record evidence for every deduction. Do not infer a semantic pass from deterministic checks.
