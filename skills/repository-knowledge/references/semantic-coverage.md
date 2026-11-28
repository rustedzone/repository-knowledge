# Semantic coverage report

Use this reference for an explicit semantic rescan, a full documentation generation, or a change to a capability already listed in `.repo-knowledge/semantic-coverage.json`. The report is consumer-owned and versioned with the repository. It is a compact map to inspected guides, source evidence, and known gaps; the guides carry the actual behavior explanation.

`repo-knowledge scan` remains structural discovery. A semantic rescan inspects source, tests, schema, configuration, and existing guides across every material repository surface. Enumerate material units independently of path-name signals and `.repo-knowledge/repository.json`; both can omit real capabilities. Follow the trace dossier and quality gates in [documentation-generation.md](documentation-generation.md) and [documentation-quality.md](documentation-quality.md). Do not mark a unit `traced` from names, declarations, or an uninspected guide.

Write `.repo-knowledge/semantic-coverage.json` with schema version `1.0`:

```json
{
  "schema_version": "1.0",
  "scope": "partial",
  "units": [
    {
      "id": "example-capability",
      "status": "partial",
      "guide": "docs/example.md#request-flow",
      "trace": "Complete an example request",
      "evidence": [
        {"path": "internal/example/handler.go", "symbol": "HandleRequest"}
      ],
      "gaps": [
        {"question": "Retry behavior after the external call fails is not established", "inspected": [{"path": "internal/example/handler.go"}]}
      ]
    }
  ]
}
```

Use `full` only after accounting for every material capability and artifact lifecycle. A `full` report may contain `partial` rows with explicit gaps, but no `unassessed` rows. For a bounded change or incomplete rescan, use `partial`; do not silently carry old rows forward as newly reviewed. A `traced` row needs a named trace, a guide, and at least one exact source or test file plus symbol or locator. A `partial` row needs a specific gap. A `unassessed` row is allowed only in a partial report.

After inspecting the cited files, run `repo-knowledge coverage-stamp` to fill missing SHA-256 digests, then `repo-knowledge coverage-check`. Stamping refuses to replace a stale digest. Reinspect changed evidence and explicitly clear its old `sha256` before stamping the revised row. Stamping is a mechanical snapshot, not proof that the agent understood the files. The checker validates format, safe and existing paths, declared capability presence in a full report, and evidence-content freshness. It does not validate the named symbol, guide anchor, flow semantics, or completeness of the agent's capability inventory. Review those manually against the guides and source. A missing report means `not_assessed`, not failure.

For ordinary source changes, inspect and update only affected report rows when their trace, evidence, or gap changed. Keep untouched rows as prior assessments. An existing `full` report can retain that scope only when no material unit was added and the untouched evidence remains current; a targeted pass is not itself a full rescan. If no report exists, do not create one merely to finish a routine code change. The report is ignored by documentation-impact classification: a report-only edit does not replace a guide update or a supported no-impact acknowledgment. Preflight does not inject the report; read it on demand.

## When to request a full rescan

Request one when the user explicitly wants a repository-wide semantic baseline or complete documentation rebuild; when existing coverage is missing or unreliable **and** a repository-wide baseline is the task; or when a material change adds or reshapes capability families so widely that affected units cannot be bounded from the diff and current inventory. Examples include reorganizing service boundaries or replacing an integration architecture when their affected flows cannot be enumerated. A large diff alone is not sufficient. If the changed capabilities and downstream effects can be identified, update their guides, routes, and coverage rows instead. If no report exists during routine work, leave it absent unless the user requests a baseline.

## Copy-paste prompts for an installed agent

Use the first prompt after a material change when the required rescan scope is uncertain. It authorizes the agent to complete the appropriate documentation work; it does not require a second approval to begin.

```text
Use the installed repository-knowledge skill, shared contract, and docs/index.md to reconcile documentation after the current change. Inspect the actual diff, affected behavior, and existing .repo-knowledge/semantic-coverage.json if present. Decide whether the affected capabilities and artifact lifecycles can be bounded from current evidence. If they can, verify the changed flows and downstream effects, update only affected guides, routes, repository metadata, and coverage rows, and preserve prior assessments for untouched units. Do not create a coverage report merely because a routine change lacks one.

If the material inventory or existing coverage cannot be trusted enough to bound the changed capabilities, or the change introduces or reshapes capability families whose effects cannot be bounded, perform a full semantic rescan now. Classify all repository shapes; independently enumerate every material capability and artifact lifecycle; use repo-knowledge scan only as structural discovery; trace entry points through rules, state/data and external boundaries, failures, and tests. Reconcile source-grounded guides, docs/index.md, .repo-knowledge/repository.json, and the versioned semantic coverage report. Mark scope full only after every material unit is accounted for; otherwise mark it partial and name the exact gaps. Run repo-knowledge coverage-stamp only after inspecting the cited evidence, then repo-knowledge coverage-check, repo-knowledge audit, and relevant validation. Finish with the rescan decision and reason, files changed, checks run, documentation-impact outcome, and remaining gaps. Preserve consumer-owned rules, ADRs, and historical or requirement-only content.
```

Use this prompt when a full baseline or rebuild is explicitly required, including an existing repository whose installed toolkit and skill are already current:

```text
Use the installed repository-knowledge skill and shared contract to perform a full semantic rescan of this repository now. Follow its documentation-generation and semantic-coverage references. Classify every repository shape and independently inventory every material capability and artifact lifecycle; treat repo-knowledge scan as a structural hint, not semantic proof. For each material executable capability, trace a real flow from entry point through validation and authorization, rules, state or external boundaries, side effects, failures, and tests. For non-executable units, trace the artifact lifecycle. Verify current-state claims against source, tests, schemas, and effective configuration.

Create or repair implementation-ready guides, route them from docs/index.md, update verified capabilities in .repo-knowledge/repository.json, and write .repo-knowledge/semantic-coverage.json with exact evidence and gaps. Use full scope only if every material unit is accounted for; otherwise use partial scope and report uncovered units. After inspecting cited files, run repo-knowledge coverage-stamp and coverage-check, then audit and relevant checks. Report the guides and routes changed, verified flows, validation results, documentation-impact outcome, and unresolved questions. Preserve consumer-owned rules, ADRs, requirements, and historical context; do not modify application code unless the task separately requires it.
```
