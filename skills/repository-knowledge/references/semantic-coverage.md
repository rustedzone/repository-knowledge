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
