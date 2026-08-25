# Repository type profiles and documentation coverage

Read this reference for full documentation generation or rebuilds. Select every profile supported by repository evidence. Profiles are additive: a monorepo or product repository commonly matches several.

The suggested paths are defaults, not a required taxonomy. Reuse accurate existing locations and the repository's own language. A coverage topic may be a focused section rather than a separate file when it is small and tightly coupled; split it when it has independent behavior, contracts, evidence, or change triggers.

## Universal baseline

| Guide | Required coverage |
| --- | --- |
| `docs/repository-overview.md` | Purpose, consumers, repository shapes, major capabilities, technology choices, prerequisites, current limitations, and newcomer reading path. |
| `docs/architecture.md` | System boundary, components, responsibilities, dependency direction, runtime topology, main control/data flows, cross-cutting concerns, and architectural constraints. |
| `docs/index.md` | Recommended reading order, concept routes, “read when” triggers, evidence anchors, and known gaps. |

Add cross-cutting guides when material:

- `docs/development.md` for setup, configuration, common commands, testing strategy, fixtures, debugging, and contribution workflow;
- `docs/security.md` for identity, authorization, secrets, sensitive data, trust boundaries, and security-relevant failure modes;
- `docs/operations/` for deployment, observability, scaling, backup/restore, incident diagnosis, rollback, and scheduled work;
- `docs/decisions/` only for evidenced accepted decisions or explicitly requested proposals.

## Backend service or server application

Detect from server entry points, route/controller definitions, service layers, workers, persistence, deployable service manifests, or server tests.

| Topic | Typical destination | Explain |
| --- | --- | --- |
| Business domains/capabilities | `docs/domains/<domain>.md` | Vocabulary, actors, commands/events, implemented rules, invariants, state transitions, authorization, side effects, ownership boundaries, failure/recovery cases, and change impact. Trace these through actual handlers/services/state/tests; prefer one guide per material domain. |
| HTTP/RPC/GraphQL APIs | `docs/api/<surface>.md` | Consumers, base path/protocol, authentication, authorization, request/response semantics, errors, pagination, idempotency, compatibility/versioning, and handler-to-domain flow. Group cohesive surfaces; do not produce only an endpoint list. |
| Events, queues, and jobs | `docs/events/` or `docs/jobs/` | Producers, consumers, payload meaning, ordering, delivery semantics, retries, dead letters, idempotency, scheduling, and operational recovery. |
| Persistence | `docs/data/<model-or-context>.md` | Entities and relationships, ownership, constraints, transactions, state lifecycle, migrations, consistency, retention, indexes with semantic importance, and access paths. |
| External systems | `docs/integrations/<system>.md` | Purpose, direction, protocol, authentication, mapping, timeouts, retries, circuit breaking, failure effects, reconciliation, sandbox/testing, and operational ownership. Prefer one guide per material external system. |
| Runtime/operations | `docs/operations/` | Processes, configuration, dependencies, health checks, observability, deployment, scaling, migrations, rollback, and troubleshooting. |

Trace synchronous requests and asynchronous flows independently. Document middleware, authorization, validation, transaction boundaries, error mapping, and external failure behavior when present.

For layered or Clean Architecture backends, trace concrete flows through the real composition root, transport middleware, controller or handler, use case or domain service, repository or integration adapter, state store, response mapper, and tests. Document dependency direction and trust boundaries rather than merely naming layers. Include source-derived extension examples for dependency-injection wiring, response and error conventions, repositories or integrations, and tests when those are recurring patterns.

## Frontend or web client

Detect from browser entry points, routes/pages, components, client state, API clients, bundler configuration, or UI tests.

| Topic | Typical destination | Explain |
| --- | --- | --- |
| User capabilities | `docs/features/<feature>.md` or `docs/domains/` | User goal, entry routes, exact screen/component flow, permission/role behavior, loading/empty/success/error states, validation, mutations, and backing services. Follow the implementation rather than summarizing route names. |
| Navigation and composition | `docs/frontend/navigation.md` | Actual route tree, layouts, route groups, redirects, guards, middleware/providers, session or cookie checks, code splitting, deep links, and composition boundaries. Explain the concrete authentication path and cite its implementation. |
| State and data flow | `docs/frontend/state-and-data.md` | Inventory what owns server, global, local, form, and URL state; identify whether Redux/Zustand/Context or no global store is used; explain query keys, caching, invalidation, mutations, optimistic behavior, synchronization, and errors with source-derived examples. |
| UI system | `docs/frontend/ui-system.md` | Component layers, design-system packages, styling/tokens, page/layout composition, accessibility, tables/modals/feedback, forms and validation, internationalization, and extension conventions. Include a current standard page/component and form/schema example when those patterns exist. |
| Backend integrations | `docs/integrations/` | API clients, authentication/session lifecycle, request mapping, error mapping, generated clients, and environment selection. |
| Build and delivery | `docs/operations/frontend-delivery.md` | Environment variables, build modes, assets, observability, deployment, CDN/caching, feature flags, and rollback. |

Trace representative user journeys from route entry to rendered, loading, empty, unauthorized, error, and successful states. For onboarding, demonstrate the repository's actual patterns for adding a route/page, fetching and mutating data, applying a guard, composing design-system components, and building a validated form wherever those mechanisms exist.

## Mobile or desktop application

Detect from platform projects, application lifecycle entry points, screens/views, navigation, local storage, native capabilities, signing, or packaging.

Cover user features, navigation, presentation/state architecture, networking, offline and synchronization behavior, local persistence, platform permissions, deep links, background work, notifications, analytics/crash reporting, build variants, signing, distribution, migrations, and release procedures. Use `docs/features/`, `docs/platform/`, `docs/data/`, `docs/integrations/`, and `docs/release/` as useful—not mandatory—boundaries.

Document platform differences rather than assuming one platform's implementation applies to all targets.

## Library, framework, SDK, or reusable package

Detect from exported modules, package publication metadata, public headers/types, examples, compatibility tests, or release automation.

| Topic | Typical destination | Explain |
| --- | --- | --- |
| Concepts and architecture | `docs/concepts/` and `docs/architecture.md` | Mental model, package/module boundaries, extension model, lifecycle, dependency expectations, and internal versus public surfaces. |
| Public API | `docs/api/` or generated reference plus prose | Supported entry points, contracts, behavior, errors, concurrency, resource ownership, compatibility, and deprecation. |
| Usage | `docs/guides/` and `docs/examples/` | Installation, minimal example, common workflows, integration patterns, and troubleshooting grounded in executable examples/tests. |
| Maintenance | `docs/development.md`, `docs/releasing.md` | Test matrix, supported platforms/versions, build, publication, versioning, changelog, and release validation. |

Do not substitute generated symbol reference for conceptual and behavioral documentation.

## CLI or developer tool

Detect from command registration, argument parsers, executable entry points, shell completions, or command-focused tests.

Cover command hierarchy, common workflows, inputs/outputs, configuration precedence, environment variables, exit statuses, side effects, destructive behavior, interactive versus non-interactive modes, integrations, extension points, compatibility, installation, and release. Split substantial commands or command families into `docs/commands/<name>.md`; keep concise command reference separate from architecture and task-oriented guides.

## Infrastructure, platform, deployment, or GitOps repository

Detect from Terraform/Pulumi/CloudFormation, Kubernetes/Helm, configuration-management code, environment overlays, deployment pipelines, or operational scripts.

Cover environment topology, accounts/projects/clusters, modules and ownership, dependency graph, state and secret management, networking and trust boundaries, change/promotion flow, plan/apply controls, drift handling, policy enforcement, observability, backup/restore, disaster recovery, rollback, and incident procedures. Use `docs/environments/`, `docs/modules/`, `docs/operations/`, and `docs/security.md` where helpful.

Never expose secret values. Distinguish declared infrastructure from runtime facts that were not observed.

## Data, analytics, ETL, or machine-learning repository

Detect from workflow definitions, notebooks, transformations, schemas, datasets, feature/model code, evaluation, orchestration, or serving components.

Cover data sources and ownership, lineage, schemas and quality constraints, transformations, partitioning, orchestration, schedules, backfills, failure recovery, privacy/retention, feature definitions, training, evaluation metrics, experiment tracking, reproducibility, artifacts, model/version promotion, serving, drift/monitoring, and cost or capacity concerns. Useful destinations include `docs/pipelines/`, `docs/data/`, `docs/models/`, and `docs/operations/`.

Separate observed metrics and current model behavior from targets or historical experiment claims.

## Monorepo, multi-service, or plugin ecosystem

Detect from workspace manifests, multiple deployable entry points, packages/apps/services directories, dependency graphs, or independent release definitions.

Start with a repository-wide map covering units, ownership, dependency direction, shared tooling, and build/test/release orchestration. Then apply the appropriate profile to each material deployable or package. Use paths such as `docs/apps/<name>/`, `docs/services/<name>/`, or `docs/packages/<name>/` when they improve navigation.

Document cross-unit flows and shared contracts separately. Do not treat a package directory listing as package documentation, and do not let one well-documented application hide uncovered siblings.

## Firmware, embedded, systems, or hardware-adjacent repository

Detect from board/target definitions, low-level entry points, drivers, cross-compilation, real-time tasks, hardware interfaces, or target-specific tests.

Cover supported targets, boot/lifecycle, modules and scheduling/concurrency, memory/resource constraints, hardware interfaces and protocols, configuration, safety invariants, error/recovery behavior, build/flash/debug workflows, simulation, test hardware, artifact/release process, and target differences. Keep verified software behavior distinct from external hardware assumptions.

## Documentation, specification, configuration, or content repository

Detect when documents, schemas, policies, templates, localization, or structured content are the primary product rather than support files.

Cover audience and publication purpose, information architecture, authoritative sources, content/schema lifecycle, generation and validation, cross-references, versioning, localization, review/approval, publication/deployment, and contributor workflows. Explain semantic ownership and precedence; do not create software-domain pages that have no evidence.

## Unknown or unusual repository

Do not force a familiar software profile. Derive the product boundary, artifact lifecycle, consumers, entry points, transformations, validation, delivery, and maintenance model from evidence. Use the universal baseline, then name additional guides after the repository's real concepts. Explicitly record what could not be classified and why.

## Splitting and depth rules

Create a dedicated page when a topic has one or more of these:

- an independently meaningful domain or user capability;
- a public or cross-boundary contract;
- a distinct data lifecycle or state machine;
- an external owner, service, protocol, or failure mode;
- dedicated configuration, deployment, operational, security, or testing concerns;
- enough relationships and change consequences that a shared page becomes hard to navigate.

Combine topics when they share the same audience, flow, evidence, and change trigger and can still be explained fully. “Any repository” does not mean “every folder”: document semantic units, not directory taxonomy.
