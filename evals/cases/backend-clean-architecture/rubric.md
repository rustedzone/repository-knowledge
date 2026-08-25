# Backend Clean Architecture semantic rubric

Score every item 0, 1, or 2:

- `0`: absent, false, or structural catalog only;
- `1`: partially correct but missing important mechanisms, failures, or evidence;
- `2`: accurate, connected, actionable, and supported by current repository evidence.

Critical items are marked **critical**. Passing requires every critical item to score 2 and at least 85% overall.

| Item | Critical | Review question |
| --- | --- | --- |
| Approval request trace | Yes | Does a named flow connect Oathkeeper identity, Gin middleware, controller, use case, Postgres, Keto synchronization, response mapping, and tests step by step? |
| State lifecycle | Yes | Are approval writes and guards represented as an accurate transition table with actors, side effects, invalid transitions, and failure behavior? |
| Deletion semantics | Yes | Does the guide distinguish `ON DELETE RESTRICT`, application cleanup, Keto tuple deletion, ordering, and partial-failure consequences? |
| Trust boundaries | Yes | Does architecture explain what crosses the Oathkeeper, HTTP, persistence, Redis, and Keto boundaries, including identity, validation, error translation, and recovery ownership? |
| Dependency direction | No | Does it explain the verified responsibilities and consequences of the layered architecture without inventing historical rationale? |
| Extension examples | Yes | Are DI wiring, controller response/error, use-case/repository or integration, and test examples faithful and separately useful? |
| Data and consistency | No | Does it explain transaction boundaries, Redis use, persistence constraints, side effects, and consistency limitations established by source? |
| First-change readiness | Yes | Could a newcomer add an approval transition or repository dependency and locate the correct tests without rediscovering the architecture? |
| Honesty and gaps | No | Are unimplemented retries, recovery, or business meaning recorded as exact gaps instead of generic assumptions? |

## Optional polish observations

These do not affect the score. Record whether the overview gives a friendly, repository-specific first reading path and whether component, approval-sequence, or state diagrams would materially reduce onboarding effort. Do not reward decorative diagrams or warmth that obscures evidence.
