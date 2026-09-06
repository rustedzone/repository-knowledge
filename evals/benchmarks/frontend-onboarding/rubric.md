# Frontend onboarding outcome rubric

Score every item 0, 1, or 2:

- `0`: absent, false, or only a structural catalog;
- `1`: partially correct but missing important mechanisms, failures, or evidence;
- `2`: accurate, connected, actionable, and supported by current repository evidence.

Critical items are marked **critical**. Passing requires every critical item to score 2 and at least 85% overall. Review the output without being told how the target was prepared.

| Item | Critical | Review question |
| --- | --- | --- |
| Factual accuracy | Yes | Does the documentation establish React 19 from current manifest and lockfile evidence, reject stale React 18 prose, and identify intelligence as static JSON? |
| Protected navigation | Yes | Does it connect the protected layout, session lookup, redirect behavior, and permission checks with concrete symbols and failure outcomes? |
| User-creation flow | Yes | Does it trace page composition, form state, schema validation, submission, server errors, successful navigation, and the API boundary? |
| State ownership | No | Does it distinguish server, shared, local, form, and URL state using mechanisms actually present in the fixture? |
| BFF boundary | Yes | Does it explain catch-all routing, cookie/header propagation, upstream request construction, response mapping, and errors? |
| Practical examples | Yes | Are at least three examples faithful, source-attributed, safe, and useful for distinct recurring changes? |
| Onboarding readiness | Yes | Could a newcomer locate the files and tests for a protected form-backed page without re-deriving the architecture? |
| Honesty and gaps | No | Are unknowns explicit and are static fixtures, declarations, and live integrations correctly distinguished? |
| Source preservation | Yes | Are application and configuration files unchanged? |

Record evidence for every deduction. Do not infer a semantic pass from deterministic checks.
