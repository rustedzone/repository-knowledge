# Repository knowledge index

This index routes engineers and agents to human-readable, evidence-backed repository knowledge.

> This is a bootstrap placeholder, not completed repository documentation. A documentation-generation request must replace the placeholder below with useful guides and verified routes after inspecting repository evidence.

## Start here

- Start with `repository-overview.md` and `architecture.md` after they have been created, or link accurate equivalents here.
- Use `docs/repository-inventory.md` only as an optional structural appendix; it does not explain repository behavior or architecture.
- For historical architectural intent, route to the repository's ADR or decision-record location if one exists.
- When a topic is absent, inspect implementation, tests, schemas, configuration, and decisions before writing evidence-backed documentation.

## Knowledge routes

<!-- Add entries in this form:
### Capability or domain name

Read when:
- a concrete trigger or concept changes

Knowledge:
- [Document title](relative/path.md)

Evidence anchors:
- `path/or/glob`
-->

No capability-specific routes have been recorded yet. The repository knowledge layer is incomplete until an agent classifies every repository shape, inspects the repository, writes detailed guides for every material capability and applicable surface, updates this index, and records verified capabilities in `.repo-knowledge/repository.json`.

## Documentation conventions

- Deterministic catalogs should be generated or mechanically reconciled where practical.
- Semantic documents should explain purpose, relationships, flows, behavior, constraints, business meaning, and tradeoffs—not merely list files.
- Full generation should cover every material domain, interface, data model, integration, and operational surface supported by evidence; known gaps must be explicit.
- Each material executable capability should route to a named end-to-end behavioral trace with implementation symbols, state or side effects, failures, tests, and recurring extension patterns—not only a glossary or catalog.
- Requirements, operational knowledge, and historical decisions must not be removed merely because they are not derivable from source.
