# Frontend Next.js semantic rubric

Score every item 0, 1, or 2:

- `0`: absent, false, or structural catalog only;
- `1`: partially correct but missing important mechanisms, failures, or evidence;
- `2`: accurate, connected, actionable, and supported by current repository evidence.

Critical items are marked **critical**. Passing requires every critical item to score 2 and at least 85% overall.

| Item | Critical | Review question |
| --- | --- | --- |
| Factuality and traps | Yes | Does the documentation use React 19 evidence, identify intelligence as static JSON, reject the stale live-Pega claim, and avoid unsupported speculation? |
| Protected navigation | Yes | Does it trace `ProtectedLayout` through `requireSession`, cookie lookup, redirect behavior, and permission checks with concrete symbols and failure outcomes? |
| State ownership | Yes | Does it distinguish server, shared, local, form, and URL state and explain the actual mechanisms rather than listing libraries? |
| Forms and validation | No | Does it explain `UserForm`, `userSchema`, submission, validation errors, server errors, and successful navigation with a faithful example? |
| BFF boundary | Yes | Does it trace catch-all API handling, cookie/header propagation, upstream request construction, response mapping, and errors? |
| UI composition | No | Can a newcomer identify the page/layout/component pattern and the design-system seam used by this fixture? |
| Source-derived examples | Yes | Are examples faithful, attributed to current paths/symbols, safe, and useful for distinct recurring patterns? |
| First-change readiness | Yes | Could a newcomer add a protected form-backed page and its test without rediscovering the basic architecture from source? |
| Honesty and gaps | No | Are unknowns explicit and are unused declarations distinguished from active integrations? |

## Optional polish observations

These do not affect the score. Record whether the overview gives a friendly, repository-specific first reading path and whether a small component or request-flow diagram would materially reduce onboarding effort. Do not reward decorative diagrams or warmth that obscures evidence.
