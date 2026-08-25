# Human-readable repository documentation

Use this workflow when the user wants documentation that helps a person learn, navigate, operate, or change the repository. A generated file list is discovery input, not the deliverable.

For every generation or full rebuild request, also read:

- [repository-type-profiles.md](repository-type-profiles.md) for the baseline library, repository-shape profiles, topic splitting rules, and expected guide depth;
- [source-of-truth.md](source-of-truth.md) for claim-specific evidence authority, version semantics, stale-document handling, and conflict rules;
- [documentation-quality.md](documentation-quality.md) for the anti-speculation gate, mechanism-level depth, source-derived example rules, and implementation-readiness review.

## Preserve action intent

Choose the mode from the user's request:

- **Generation mode:** `generate`, `rebuild`, `improve`, `complete`, or `fix` means inspect and write or revise the documentation now. Existing shallow docs are inputs to repair, not a reason to stop. If the quality review finds a gap, continue into source inspection and rewrite it in the same task. Do not end with “Would you like me to begin rebuilding?” because permission was already given.
- **Assessment mode:** `review`, `audit`, or `assess` without a request to change files is read-only. Apply the same standards, report evidence-backed gaps, and do not rewrite documents.
- **Explanation mode:** answer the scoped question from routed knowledge and verified evidence; do not expand it into a full documentation rebuild unless requested.

Only pause generation for a human validation condition in the shared contract, an unavailable required repository, or a change outside the user's authorized scope.

## Outcome

Produce a comprehensive, evidence-backed knowledge library proportional to the repository. A new engineer should be able to answer:

- What does this repository build, why does it exist, and who or what uses it?
- How do its major components, entry points, dependencies, and runtime boundaries fit together?
- What are its material domains or capabilities, and what behavior and business rules does each own?
- What are the important request, event, data, authentication, authorization, failure, and recovery flows?
- What interfaces, data models, external systems, configuration, and operational procedures matter?
- Where should an engineer look, what tests protect the behavior, and what consequences should they expect when changing it?
- Which important facts are verified, inferred, requirement-only, historical, unknown, or conflicting?

Comprehensive means covering every material surface, not maximizing word count. Prefer several focused, navigable guides over one shallow overview or one enormous document. Do not create empty pages, repeat the same content across files, or impose folders that repository evidence does not justify.

## 1. Classify the repository before choosing documents

Treat repository kind as multi-label. A monorepo may contain a backend service, web client, shared library, infrastructure, and data pipeline. Identify every material shape using manifests, entry points, build configuration, deployment files, source boundaries, and tests—not names alone.

Use the matching profiles in [repository-type-profiles.md](repository-type-profiles.md). Combine their coverage rather than forcing the repository into one template. Record briefly in the overview which shapes were found and which expected surfaces were not present or could not be verified.

## 2. Build a coverage plan from evidence

Before writing, form a documentation map and a private coverage ledger. Use working notes rather than publishing process scaffolding unless it helps maintainers. Give every material capability or artifact lifecycle a row containing:

- target guide and audience;
- named normal flow or lifecycle;
- exact entry point and successive symbols or paths;
- authorization, validation, and business-rule evidence;
- state or data transitions and side effects;
- persistence, event, API, or external boundary;
- alternate, failure, retry, and recovery evidence;
- tests, fixtures, configuration, and observability;
- source-derived extension example;
- status: traced, known gap, conflict, or not applicable.

Build that ledger as follows:

1. Inventory existing prose such as README files, `CLAUDE.md`, `AGENTS.md`, prior guides, and comments as discovery leads. Do not copy their current-state claims into new documentation yet.
2. Inspect manifests, lockfiles, toolchain declarations, effective configuration, entry points, tests, schemas, migrations, deployment files, API or event contracts, and decision records. Use the source-of-truth matrix to choose authority per claim.
3. Run `repo-knowledge scan` when a current structural map would help. Use capability signals to choose inspection targets, never as semantic conclusions.
4. Enumerate material units: applications, services, packages, domains, public interfaces, persistent models, integrations, jobs, operational entry points, and cross-cutting concerns.
5. Map each unit to a planned guide or an explicit section in a broader guide. Split a topic when it has its own behavior, contract, data lifecycle, integration boundary, owners, or meaningful change triggers.
6. Create a trace dossier for at least one representative path through each material executable capability. Read and record the actual trigger, every architectural hop, authorization or validation, state or boundary interaction, response or side effects, failure handling, and relevant tests. For a non-executable repository, trace its primary artifact lifecycle from source or input through validation or transformation to publication or consumption. Inspect all distinct entry-point families and boundary types; do not generalize from only the first route, controller, screen, command, or test found.
7. Compare relevant existing prose claims with authoritative current evidence. Keep a working list of confirmed, stale, conflicting, intent-only, and still-unverified claims.

The coverage map can be working notes, but the final `docs/index.md` must make the resulting coverage visible. Do not draft a capability guide from names and declarations before its dossier exists. For very large repositories, work in coherent slices and continue until all material surfaces are covered. If the requested scope cannot be completed, label the documentation partial and list exact uncovered areas; never present a partial pass as a complete rebuild.

## 3. Trace behavior, not just structure

Follow representative execution from trigger to outcome. Depending on repository shape, this can include:

- HTTP route -> middleware -> authorization -> handler -> domain/service -> persistence or external call -> response and error mapping;
- event producer -> transport or topic -> consumer -> idempotency/retry -> state change -> downstream effects;
- UI route -> screen -> state/query layer -> API client -> loading/error/permission states;
- CLI command -> parsing -> configuration -> core operation -> output, exit status, and side effects;
- pipeline input -> validation/transformation -> storage/model -> quality checks -> publication or serving;
- infrastructure change -> module -> environment composition -> plan/apply -> rollout, rollback, and observability.

Name each documented trace by its concrete outcome and identify every hop with a source path and symbol. Explain what enters and leaves the hop, which rule it enforces, how state changes, what side effects occur, and how errors propagate. A directory or controller catalog is not a trace.

When a model has status or state fields, or code writes state-like values, locate every material write and transition guard. Document the verified transition table, triggering operation, authorization, persistence boundary, side effects, and invalid-transition behavior. When deletion or retention matters, inspect schema or migration constraints plus application deletion logic and integration cleanup; do not infer cascading behavior from table relationships alone.

Use tests to learn edge cases and invariants, schemas or migrations to learn data constraints, manifests and lockfiles to learn declared versus resolved dependency state, and effective configuration to learn runtime behavior. Do not infer business meaning from filenames alone.

Classify material claims using the shared states: `verified`, `verified_stale`, `inferred`, `requirement_only`, `historical`, `unverified`, `conflict`, or `unknown`. Prose can remain readable—apply labels where uncertainty or provenance affects decisions rather than tagging every sentence. A stale claim in `CLAUDE.md`, README, or an older guide must never override a current manifest, lockfile, effective configuration, schema, test, or implementation. Preserve business intent that source cannot prove, but label it rather than presenting it as implemented behavior. Do not use `likely`, `typically`, `expected`, `probably`, `appears`, or similar hedging as a substitute for inspection; follow the anti-speculation rules in the quality reference.

## 4. Write a detailed knowledge library

Preserve accurate existing documentation and improve it in place when practical. Unless accurate equivalents already exist, every repository gets:

- `docs/repository-overview.md` for purpose, users, repository shapes, primary capabilities, technology, and a newcomer path;
- `docs/architecture.md` for boundaries, components, dependency direction, runtime topology, and important end-to-end flows;
- `docs/index.md` as the curated reading order and concept-to-evidence router.

Then add the applicable profile documents. For example, a backend with verified transaction and approval domains, an HTTP surface, a transaction data model, and an IIAM integration could reasonably produce:

```text
docs/
├── index.md
├── repository-overview.md
├── architecture.md
├── domains/
│   ├── transaction.md
│   └── approval.md
├── api/
│   └── transaction-api.md
├── data/
│   └── transaction-model.md
├── integrations/
│   └── iiam.md
├── operations/
│   └── running-and-deployment.md
└── decisions/
    └── adr-003-approval-workflow.md
```

This is an example, not a mandatory tree. Use repository language for names. Do not collapse independent domains or integrations into generic files merely to keep the document count small. Conversely, keep closely related material together when separate pages would be repetitive.

Create or revise an ADR only when repository evidence establishes an actual accepted decision and its context. Do not invent rationale, dates, alternatives, or acceptance status. Document an observable but unexplained choice in architecture as inferred or unknown; create a proposed decision only when the user asks for one.

Use `repo-knowledge rebuild --apply` only when the structural appendix is useful. Keep it clearly labeled as generated inventory and do not substitute it for semantic guides.

## 5. Meet the page depth standard

A guide must teach, not catalogue. Include the applicable parts of this structure:

- purpose, scope, users, and non-goals;
- key concepts and terminology;
- components and their responsibilities, relationships, and dependency direction;
- step-by-step normal flows plus meaningful alternate, authorization, validation, failure, retry, and recovery paths;
- public contracts: APIs, commands, events, schemas, configuration, extension points, compatibility, or versioning;
- data ownership, lifecycle, state transitions, consistency, transactions, retention, and sensitive-data handling;
- external dependencies, trust boundaries, assumptions, timeouts, retries, idempotency, and failure behavior;
- local development, configuration, testing, observability, deployment, rollback, or troubleshooting where applicable;
- a practical change guide naming implementation anchors, relevant tests, and evidence-supported downstream effects;
- evidence anchors and clearly bounded uncertainties.

Not every page needs every heading. Select what explains that topic fully. Use tables for contracts and comparisons, and Mermaid diagrams when they materially clarify component, sequence, state, or data relationships. Every diagram must agree with the prose and evidence.

### Newcomer orientation and visual explanation

After factual and semantic depth are established, make the library approachable without turning it into marketing copy. The overview should open in repository language and give a newcomer a quick mental model: what the repository produces, its principal entry point or runtime, the few components or capabilities that matter first, and where to go next. Prefer a short verified “first 15 minutes” or “first change” path when repository evidence establishes the prerequisites, commands, and tests. Explain local terms on first use and write directly to an engineer who has not yet learned the folder structure.

Use a small number of diagrams only when they reduce the effort required to understand a real relationship:

- a component or flowchart for several runtime components or trust boundaries;
- a sequence diagram for a material request, event, or artifact lifecycle with multiple hops or alternate outcomes;
- a state diagram for a verified lifecycle whose permitted transitions matter;
- a data-flow diagram when ownership or movement across stores and integrations is otherwise difficult to follow.

Keep the surrounding prose authoritative. A diagram supplements the named behavioral trace; it does not replace step responsibilities, errors, side effects, tests, or evidence anchors. Use concrete repository terminology, include only relationships verified by current evidence, and omit a diagram when a short paragraph or table is clearer. Check that Mermaid syntax renders in the repository's documentation host when that can be verified.

For technology, dependency, runtime, command, interface, schema, and deployment assertions, include an evidence anchor near the relevant section so reviewers can verify freshness quickly. When a manifest declares a version range and a lockfile resolves an exact version, describe them accurately instead of turning the declared range into an exact installed version.

Include short source-derived examples for the recurring patterns a newcomer is expected to copy. Examples must name their source path and symbol, remain faithful to the current implementation, and follow the validation and secret-safety rules in the quality reference. A list of library names such as SWR, Formik, Yup, or Zod is not a state-management or form guide; explain and demonstrate the repository's concrete composition.

Avoid dumping source snippets or endpoint/file inventories into prose. Generated catalogs may supplement a guide, but the guide must explain semantics, relationships, invariants, and change consequences.

For each distinct architecture extension pattern a newcomer will use, provide a verified example. Examples include dependency injection or composition-root wiring, controller or handler response and error conventions, repository implementations, middleware or guards, event registration, schema or migration changes, pipeline stages, infrastructure modules, and tests. Choose the repository's real patterns rather than only the easiest snippet to find.

## 6. Reconcile routing and repository metadata

Update `docs/index.md` with:

- a recommended newcomer reading order;
- routes grouped by repository concepts rather than arbitrary folders;
- a short “read when” trigger for every guide;
- relevant evidence anchors;
- a coverage or known-gaps section when anything material remains undocumented or uncertain.

Update `.repo-knowledge/repository.json` with verified capability entries that reference the written guides and their evidence. Use one capability per meaningful change surface; multiple guides may support one capability, and one cross-cutting guide may support multiple capabilities. Do not promote raw scan signals into verified capabilities without inspection.

## 7. Validate completion

Before reporting success:

1. Check every local link and evidence path.
2. Cross-check every current-state technical claim against its claim-specific authoritative evidence. Re-read manifests and lockfiles directly for framework, runtime, and dependency versions; do not validate those claims against existing prose.
3. Compare the final library against every applicable row in the repository-type profiles. Cover it, mark it not applicable with evidence, or list it as a known gap.
4. Confirm every material unit from the coverage plan appears in `docs/index.md` and `.repo-knowledge/repository.json` where the claim is verified.
5. Run `repo-knowledge audit` and address missing routes, broken links, and stale scan findings relevant to the request.
6. Run a stale-prose review: compare relevant README, `CLAUDE.md`, `AGENTS.md`, prior generated docs, and comments with the new guides. Correct stale files inside the user-authorized documentation scope; report stale files outside that scope without copying their claims or silently rewriting them.
7. Search generated prose for unsupported hedge language and review every occurrence using the anti-speculation gate.
8. Compare every coverage-ledger row with its guide. Confirm each traced flow, transition, boundary, failure path, test, and extension example appears where useful; a path listed only under “Evidence” does not satisfy the row.
9. Run the implementation-readiness review in the quality reference. A newcomer must be able to locate and follow at least one real change pattern for each material repository profile.
10. Read every generated or revised semantic guide, not only `docs/index.md` and the first lines. The library must explain the system without requiring source inspection for the basic mental model.
11. Review the newcomer path and any diagrams as reader-experience polish. Confirm the opening uses repository language, the reading order leads to a practical next action, and every visual is readable, evidence-backed, and consistent with the prose. Do not fail an otherwise clear guide merely because a diagram would be decorative rather than useful.

### Remediation loop

If any completion or quality check fails in generation mode:

1. name the failing ledger row and missing evidence;
2. inspect the relevant implementation, tests, schemas, configuration, or contracts more deeply;
3. revise the affected guide and routing metadata;
4. rerun the failed check and then the final quality gate;
5. repeat until the library passes or a genuine contract-defined human gate remains.

Do not replace remediation with an assessment report or a request to begin. In assessment mode, report the same findings without applying this mutation loop.

The request is not complete when any of these are the only result:

- `docs/repository-inventory.md`;
- a placeholder or thin `docs/index.md`;
- one generic overview for a repository with several material surfaces;
- `.repo-knowledge/scan-state.json` or a rebuild proposal;
- prose that merely repeats directory names, file counts, fingerprints, endpoint names, or capability signal labels;
- a document tree whose pages contain only outlines, TODOs, or generic framework descriptions.
- technology or version claims sourced only from README, `CLAUDE.md`, `AGENTS.md`, prior generated prose, or comments when manifests, lockfiles, configuration, or implementation are available.
- domain pages that are only glossaries or endpoint lists and do not establish implemented rules, invariants, state transitions, permissions, side effects, and failures where those exist;
- frontend pages that name libraries but do not explain actual route guards, state ownership, data flow, forms/validation, or UI composition;
- onboarding guides with no source-derived example for a recurring implementation pattern when such a pattern exists in the repository.
- capability guides with no named end-to-end behavioral trace or artifact lifecycle;
- stateful domains that list tables or statuses without locating transition writes, guards, side effects, and invalid-transition behavior;
- generation tasks that identify these gaps but stop before revising the authorized documentation.

Report the guides and routes created, repository shapes covered, end-to-end flows verified, authoritative evidence checked, stale/conflicting prose found, and exact remaining uncertainty. Never say repository documentation was generated when only mechanical artifacts or shallow summaries were produced.
