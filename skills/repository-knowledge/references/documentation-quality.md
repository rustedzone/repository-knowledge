# Documentation quality and implementation readiness

Use this quality gate for every full documentation generation or rebuild. Coverage is not complete merely because a topic has a file. The content must be factual, mechanism-level, traceable, and useful for making a real change.

## Anti-speculation gate

Write a technical statement only when one of these is true:

- current authoritative repository evidence verifies it;
- an accepted source states it as a requirement or historical fact and the guide labels it accordingly;
- a bounded inference is necessary and the guide names both the evidence and the unresolved uncertainty.

Do not fill evidence gaps with generic ecosystem knowledge. Words such as `likely`, `typically`, `expected`, `probably`, `appears`, `seems`, `may`, `might`, and `could` are review signals. They are not categorically forbidden, but every occurrence must express a sourced contract, a clearly labeled requirement, or a bounded inference with named evidence. Remove sentences like “this likely triggers downstream workflows” or “integrations typically represent asynchronous flows” when the repository does not establish them.

If implementation evidence does not answer a material question, write a concise known gap:

```text
Not established by repository evidence: session behavior after a role change.
Inspected: <auth/session implementation paths and relevant tests>.
```

Do not invent an answer, and do not expand the gap into speculative prose. A known gap does not count as verified coverage; expose it in `docs/index.md` and the final handoff.

Before completion, search all generated or revised guides for hedge terms and inspect every match. Also look for generic claims that could describe any project using the same framework. Replace them with repository-specific mechanisms and evidence or remove them.

## Mechanism-level depth

For each material capability, inspect and connect the applicable implementation chain:

1. trigger or entry point;
2. routing, composition, or orchestration;
3. authentication, authorization, validation, and business-rule enforcement;
4. state ownership and transitions;
5. persistence, API, event, or external-system boundary;
6. response, rendering, side effects, errors, retry, and recovery;
7. tests, fixtures, configuration, and observability that establish the behavior.

Do not assume every repository has every layer. State an important absence only after inspection. The resulting guide should explain how the layers interact, not present separate inventories of filenames and endpoints.

For each material executable capability, the guide must contain at least one named behavioral trace backed by a completed trace dossier. Present it as a numbered flow or a sequence diagram plus explanatory text. Every step must identify a concrete path and symbol and explain its input, responsibility, output or state effect, and error behavior. A file list connected by arrows without those semantics does not pass.

### Domain and feature guides

A domain page is not complete as a glossary plus endpoint list. Establish from evidence, where applicable:

- actors and permissions;
- commands, queries, events, and entry points;
- business rules and where each rule is enforced;
- state model and permitted transitions;
- invariants, uniqueness or naming constraints, and validation;
- side effects and downstream dependencies;
- transaction, consistency, concurrency, retry, and idempotency behavior;
- failure/recovery paths and user-visible consequences;
- relevant tests and the safest extension points.

If questions such as “what happens to a session after a role change?” cannot be answered, state the exact gap and inspected evidence. Do not manufacture domain policy from UI labels, type names, or endpoints.

When state-like fields exist, search for all material assignments, updates, transition methods, guards, and persistence calls. Include a transition table with source state, trigger, guard or actor, destination state, side effects, and failure behavior when evidence establishes them. For deletion, document application cleanup and external synchronization as well as database cascade, restrict, or null behavior.

### Frontend guides

For a frontend repository, replace library-name summaries with the actual application mechanisms:

- route tree, layouts, route groups, redirects, and deep links;
- guard execution: middleware, server layout, provider, hook, cookie/session check, permission utility, and redirect/error outcome;
- state ownership by category: server cache, shared client state, local component state, form state, and URL state—including confirmation that a global store is absent when verified;
- query keys, fetching, mutation, cache invalidation, optimistic updates, and loading/empty/error behavior;
- design-system source, component composition, styling/tokens, accessibility, feedback, tables/modals, and page conventions;
- form composition, schema location, validation lifecycle, submission, server errors, and success behavior;
- BFF/API-client request construction, session/header propagation, response mapping, and failures;
- tests that demonstrate these patterns.

The guide must answer “where do I add a page, fetch data, protect it, build its form, and test it?” using this repository's code.

### Integrations and operations

For an integration, follow actual construction and call sites through authentication, request/event mapping, response handling, timeouts, retries, errors, reconciliation, and tests. An environment variable or dependency name alone does not establish an active workflow.

For development and deployment, derive exact commands, prerequisites, configuration precedence, build outputs, CI stages, deployment targets, health checks, observability, migrations, rollback, and troubleshooting from executable configuration. Do not turn familiar framework defaults into repository facts.

### Architecture and trust boundaries

Architecture documentation must explain responsibilities and dependency consequences across layers, processes, stores, and external systems. For every important boundary, identify what crosses it, which identity or credential is trusted, where validation and authorization occur, how errors are translated, and which side owns retry or recovery. Explain historical “why” only when an accepted decision or other authoritative intent source supports it; otherwise document the verified responsibility and mark rationale unknown.

## Source-derived examples

Include concise examples for recurring patterns a newcomer is expected to copy or adapt. Select examples based on repository shape and actual evidence. Useful examples may include:

- a standard page or feature composition;
- a route guard or permission check;
- a server-state query and mutation with cache invalidation;
- a form paired with its actual validation schema;
- a BFF proxy or API-client call and error mapping;
- a backend handler-to-domain-to-persistence flow;
- an event producer/consumer contract;
- a CLI command definition;
- a configuration/module/pipeline extension;
- the test pattern protecting any of the above.

Cover the distinct extension patterns surfaced by the coverage ledger. For a layered backend this commonly means composition-root or dependency-injection wiring, handler or controller conventions, use-case or domain orchestration, repository or integration adapters, and their tests. For other profiles choose equivalent real extension seams. One incidental snippet does not satisfy unrelated patterns.

For every snippet:

- name the source path and symbol immediately before or after it;
- copy a small, coherent excerpt or label it `adapted from` when irrelevant detail is removed;
- preserve real names, types, control flow, and error behavior needed to understand the pattern;
- compare copied excerpts with current source before finishing;
- if an adapted example is meant to compile, run the relevant formatter/typecheck/test when practical;
- never include secrets, credentials, tokens, personal data, generated bulk output, or a large source dump.

Do not fabricate an idealized example that the repository does not use. Do not show generic framework tutorial code. If no stable recurring pattern exists, explain the variation with concrete source anchors instead of forcing a snippet.

## Reader-experience polish

Apply this after the factuality, depth, and coverage work—not as a substitute for it. Help a newcomer form a useful mental model before presenting exhaustive detail:

- open the overview with the repository's purpose, primary runtime or artifact, and the few concepts needed to understand the rest;
- give `docs/index.md` a recommended reading path and a concrete next action;
- when verified commands and tests exist, include a concise first-run or first-change path;
- define repository-specific terminology on first use and prefer direct engineering language over policy narration;
- add Mermaid diagrams selectively for multi-component boundaries, multi-hop flows, verified state machines, or data movement that prose alone makes difficult to retain.

Diagrams are optional. Every visual must be evidence-backed, consistent with the prose, and accompanied by the semantics it illustrates. Do not penalize a clear short explanation for omitting a decorative diagram, and do not let a polished introduction hide an unsupported claim or missing behavioral trace.

## Implementation-readiness review

Read the final library as a new engineer and test whether it supports a first real change. For every material repository profile, choose at least one representative change and confirm the docs alone identify:

- the starting file or entry point;
- the participating components and control/data flow;
- the rule, state, contract, or integration constraints that must remain true;
- the existing pattern or source-derived example to follow;
- the relevant tests and commands;
- impact areas supported by the traced dependencies.

Examples include adding a protected frontend page with a validated form, changing a backend state transition, adding an API field through persistence and mapping, extending a CLI command, or modifying an infrastructure module safely.

The full documentation request remains incomplete if a representative change still requires rediscovering basic architecture or established coding patterns directly from source.

## Final quality gate

Every applicable row must pass:

| Dimension | Pass condition |
| --- | --- |
| Factuality | Material claims have claim-appropriate evidence; unsupported speculation is absent. |
| Depth | Guides connect mechanisms, rules, state, boundaries, errors, and tests rather than listing concepts. |
| Coverage | Every material capability and repository profile is documented or listed as an exact known gap. |
| Actionability | A newcomer can follow representative change paths and commands. |
| Examples | Recurring implementation patterns have concise source-derived examples or a justified concrete alternative. |
| Traceability | Evidence paths and symbols make important claims and snippets reviewable. |
| Honesty | Unknowns, stale claims, and conflicts are explicit and never disguised as likely behavior. |

In generation mode, a failed row triggers more inspection and revision. Do not stop at a critique, offer to start the rebuild, or declare the docs insufficient as the final outcome when the user already requested generation or rebuild. In assessment mode, preserve read-only scope and report the failed rows.

If any applicable dimension fails, continue inspecting and revising or report the documentation as partial. Do not claim comprehensive completion.
