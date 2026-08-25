# Claim-specific sources of truth

Use this reference whenever repository prose contributes to an assessment or documentation generation. The shared evidence precedence remains the safety contract, but authority is also claim-specific: a test can verify behavior while a package manifest or lockfile is the appropriate source for a dependency version.

## Core rule

Existing documentation is a discovery, intent, and historical source. It is not self-verifying current state.

This applies to README files, `CLAUDE.md`, `AGENTS.md`, agent rules, wikis copied into the repository, prior generated documentation, ADR summaries, comments, examples, and the generated structural inventory. Use them to find concepts and possible evidence, then verify their technical claims against the current checkout.

`CLAUDE.md`, `AGENTS.md`, and similar files are authoritative only for their scoped agent or contributor instructions. They are not authoritative for application framework versions, dependency state, routes, runtime topology, data schemas, or implemented behavior.

## Authority matrix

| Claim | Prefer current evidence from | Important interpretation |
| --- | --- | --- |
| Language, runtime, framework, and dependency declarations | Package/build manifests, workspace configuration, toolchain/version files | A declared constraint or version range is not necessarily the exact resolved or deployed version. |
| Exact resolved dependency versions | Lockfiles, dependency resolution output, vendored module metadata | Account for workspace overrides, replacements, patches, and platform-specific resolution. Do not regenerate a lockfile merely to document it. |
| Runtime version used in CI/build/deployment | CI image/toolchain setup, container base image, deployment/build configuration | Separate local declared support from the version actually selected by each environment. |
| Build, test, lint, development, and release commands | Manifest scripts, Make/Task files, build configuration, CI jobs, executable scripts | README commands may be stale. Explain differences between local and CI workflows when they exist. |
| Component boundaries and dependency direction | Entry points, module/workspace declarations, imports, composition/wiring, build graph | Directory names and architecture prose are leads; verify actual connections. |
| HTTP/RPC/GraphQL interfaces | Authoritative API schema when enforced, route/controller registration, request/response types, contract tests | If schema and implementation disagree, report a conflict or stale artifact rather than choosing silently. |
| Events, queues, and scheduled jobs | Producer/consumer registration, payload schemas/types, broker/job configuration, behavioral tests | Topic names alone do not prove delivery, retry, ordering, or idempotency semantics. |
| Implemented behavior and edge cases | Executable tests and source implementation; observed runtime behavior when available | Requirements may describe intended behavior and remain `requirement_only` if implementation does not prove it. |
| Data model and persistence | Current schemas, migrations, ORM/entity definitions, constraints, repository/query code, persistence tests | Distinguish desired model definitions, migration history, and observed deployed state. |
| Configuration and environment variables | Configuration loaders and schemas, defaults, deployment definitions, validated examples, configuration tests | `.env.example` documents an input surface but does not prove an effective production value. Never expose secrets. |
| Authentication and authorization | Middleware/guards/policies, identity-provider configuration, permission definitions, security tests | UI visibility alone does not prove server-side authorization. |
| External integrations | Client construction and calls, formal contracts, mapping code, authentication/configuration, integration tests | A dependency package or environment variable alone does not prove an active integration or its failure behavior. |
| Deployment and operations | IaC, manifests, container definitions, CI/CD workflows, health checks, telemetry configuration, operational scripts | Declared deployment configuration is not observed production state unless runtime evidence is available. |
| Public library or CLI contract | Export definitions, public types, command registration/help, compatibility tests, packaging metadata | Examples illustrate usage but may not cover the supported contract. |
| Product purpose, business rules, and ownership | Accepted requirements, current product/domain documentation, tests and implementation for realized behavior, ownership configuration | Source may prove mechanics without proving why they exist. Preserve explicit intent with the appropriate claim state. |
| Architectural rationale and history | Accepted ADRs, decision records, changelog, version history, traceable discussions stored in the repository | Current code proves the chosen shape, not the original alternatives or rationale. Never invent them. |

When a repository uses different authoritative artifacts, follow the same principle and name them in the generated guide.

## Version claims

Treat versions precisely:

- A manifest entry such as `"react": "^19.1.0"` is a declared compatible range, not proof that exactly 19.1.0 is installed.
- A lockfile can establish the exact resolved package in that lock state. Check workspace scopes, peer variants, overrides, replacements, and multiple resolved versions.
- A runtime or build environment can select another tool version. Check toolchain files, CI, containers, and deployment configuration before saying what production uses.
- Source imports and APIs can corroborate actual use, but an installed or declared dependency alone does not prove that the repository actively uses it.

For an onboarding technology summary, state only the useful level of precision and attach the authoritative path. For example: “The web application declares React 19 (`package.json`)” is better than copying “React 18” from `CLAUDE.md`; add an exact resolved version only when that distinction helps the reader and the lockfile supports it.

## Verification procedure

For every material current-state claim:

1. Identify the claim type before selecting evidence.
2. Open the authoritative artifact in the current checkout; do not rely on a summary of it in another document.
3. Check corroborating implementation or tests when declaration does not prove active behavior.
4. Compare any existing prose claim with the authoritative evidence.
5. Record the outcome:
   - `verified` when current evidence directly supports it;
   - `verified_stale` when current authoritative evidence objectively contradicts existing prose;
   - `conflict` when authoritative sources disagree and intent or effective truth cannot be resolved safely;
   - another contract state when the claim is inferred, intent-only, historical, unverified, or unknown.
6. Write the generated guide from the verified evidence, not by lightly editing the stale sentence.

Do not use the number of agreeing prose files as a confidence signal. Several copied documents can repeat the same stale claim.

## Stale prose and scope

During full documentation generation, perform a staleness pass over relevant existing prose:

- Correct stale files inside the documentation scope the user authorized, while preserving requirement-only and historical material.
- Do not silently rewrite agent instructions, ADR history, or files outside the authorized documentation scope. Report objective mismatches with the authoritative evidence path so their owners can reconcile them.
- If an outdated file is also an agent instruction file, generated repository guides must still reflect current technical evidence. Mention the stale instruction file in the handoff rather than allowing it to contaminate the docs.
- When two authoritative-looking technical sources disagree, describe the conflict and request human validation only when required by the shared contract.

The documentation task is not complete if a material current-state claim remains sourced only from existing prose while stronger repository evidence is available.
