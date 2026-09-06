# Testing the toolkit

## Automated checks

```bash
make check
```

This runs all Go tests, `go vet`, and `gofmt` verification. To build every release target and checksums:

```bash
make release
```

GitHub CI treats the declared minimum and the release toolchain as separate compatibility targets. Go 1.22.0 and Go 1.25.14 each run `make check` and build both `repo-knowledge` and `repo-knowledge-eval`; race-enabled tests run once on Go 1.25.14. To reproduce the minimum-version checks with Go's toolchain selection:

```bash
GOTOOLCHAIN=go1.22.0 make check
GOTOOLCHAIN=go1.22.0 make build
GOTOOLCHAIN=go1.22.0 go build -trimpath -o /tmp/repo-knowledge-eval ./cmd/repo-knowledge-eval
```

`workflow_security_test.go` examines every YAML workflow in `.github/workflows/`. External actions must use a full 40-character commit SHA, Docker actions must use a SHA-256 digest, and local actions may use a repository-relative path. The focused regression includes a deliberately mutable `actions/checkout@v6` fixture and must reject it:

```bash
go test -run TestImmutableUsesValidationRejectsMutableReferences ./...
```

## Agent evaluation

### Conformance evaluations

List the isolated evaluation cases:

```bash
make eval-list
```

Prepare a disposable target for one supported agent:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --case backend-clean-architecture \
  --output /tmp/repository-knowledge-eval-backend \
  --agent claude-code
```

Run that agent in the printed target with the exact printed prompt. The runner does not launch the agent or modify its authentication and model configuration. When the agent finishes, run:

```bash
go run ./cmd/repo-knowledge-eval grade \
  --case backend-clean-architecture \
  --target /tmp/repository-knowledge-eval-backend
```

The grade covers only objective artifact and evidence checks. Complete the case's `rubric.md` and record the trial with [`evals/report-template.md`](../evals/report-template.md). Acceptance requires all critical rubric items to score 2 and at least 85% of available semantic points. A deterministic pass without that review is pending, not a successful evaluation.

Use fresh output directories for repeated trials. Track first-attempt `pass@1`; target `pass@3 >= 0.90` across representative cases and agents. For a change intended to fix a known regression, require all three repeated relevant trials to pass (`pass^3 = 1.00`) before release.

### Outcome benchmarks

Outcome benchmarks use a neutral prompt and explicit conditions. `control` copies only the fixture and leaves no experiment marker inside the target; `treatment` installs only the requested adapter. Both conditions record their baseline in an adjacent sidecar. Model and reasoning metadata are required, and both conditions must use the same values and trial number:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition control \
  --output /tmp/frontend-control \
  --agent codex \
  --agent-version codex-desktop-2026.09 \
  --model-version gpt-5.6-sol \
  --reasoning high \
  --repository-knowledge-revision v0.10.0 \
  --trial 1

go run ./cmd/repo-knowledge-eval prepare \
  --family benchmark \
  --case frontend-onboarding \
  --condition treatment \
  --output /tmp/frontend-treatment \
  --agent codex \
  --agent-version codex-desktop-2026.09 \
  --model-version gpt-5.6-sol \
  --reasoning high \
  --repository-knowledge-revision v0.10.0 \
  --trial 1
```

Run the printed prompt independently in each target, preserve each patch or output archive, and grade with `--family benchmark`. Do not expose the rubric or expected trace before the reviewer scores the output. Recording with `--results evals/results --run-date <YYYY-MM-DD> --duration <duration> --artifact <file>` uses exclusive-create semantics and accepts failed as well as successful trials. See [`evals/README.md`](../evals/README.md) for the complete paired workflow and result schema.

Harness tests verify that invalid or missing conditions write nothing, control has no toolkit assets, treatment installs only the requested adapter, fixture application files are identical, metadata survives grading, protected-source checks run under both conditions, failed trials can be recorded, and existing result paths cannot be overwritten. A deterministic pass remains pending until semantic status, score, and reviewer are supplied explicitly.

## Disposable target smoke test

```bash
make build
TARGET_REPOSITORY="$(mktemp -d)"
git -C "${TARGET_REPOSITORY}" init
./repo-knowledge install --target "${TARGET_REPOSITORY}" --all-agents --source local --ref v0.10.0
./repo-knowledge doctor --target "${TARGET_REPOSITORY}"
./repo-knowledge scan --target "${TARGET_REPOSITORY}"
./repo-knowledge rebuild --target "${TARGET_REPOSITORY}"
./repo-knowledge rebuild --target "${TARGET_REPOSITORY}" --apply
./repo-knowledge audit --target "${TARGET_REPOSITORY}"
```

## First real-repository scenarios

1. Install into a repository with no `docs/`. Confirm the prompt adapter and routing index are created, scan records structure, and rebuild proposes before applying.
2. Install into a repository with existing `AGENTS.md`, `.claude/rules/`, `.agents/rules/`, `.cursor/rules/`, `docs/index.md`, and semantic docs. Confirm consumer-owned content remains unchanged outside toolkit-managed files and the managed Codex block.
3. Install each of `codex`, `claude-code`, `antigravity-ide`, and `cursor` separately, then use `--all-agents`. Confirm the manifest records all four in canonical order and `doctor` requires every native rule, skill, and hook registration. Confirm combining `--all-agents` with `--agent` fails without writing an installation.
4. Seed each native hook container with unrelated consumer settings and commands. Install, update, and switch adapter subsets; confirm those values remain semantically unchanged while only the toolkit command is added, canonicalized, or removed. Confirm malformed JSON stops before any hook container is written.
5. Run `hook-context` for each host protocol from a nested repository path. Confirm Codex and Claude Code receive plain preflight context, Cursor receives valid `additional_context` JSON, and Antigravity receives a full first `injectSteps` payload followed by a bounded reminder. Confirm the context names the contract, repository configuration, documentation index, selected-route receipt, and evidence-only boundary.
6. Review/trust the project hook where required, open a fresh session in each selected agent, and request a repository change without mentioning documentation. Confirm the first progress update says `Repository knowledge preflight: loaded`, names selected routes, loads the skill, and routes through `docs/index.md`. Remove the hook command and confirm `doctor` fails. Temporarily remove the binary from the host PATH and confirm the host surfaces a hook failure while the rule directs manual preflight without silently installing anything.
7. Ask each selected agent: “Use repository-knowledge to generate docs for this repository.” Test at least a backend, frontend, reusable library, infrastructure/data repository, and one mixed monorepo. Confirm the agent classifies every applicable shape; does not stop after `scan`, `rebuild`, or a shallow overview; traces each material capability; and writes the baseline plus relevant profile-specific guides. For a backend with several domains, APIs, models, and integrations, confirm it creates focused pages instead of collapsing them into one summary. Confirm `docs/index.md` exposes coverage and `.repo-knowledge/repository.json` records verified capability routes.
6. Inspect generated pages for semantic depth: purpose, concepts, component relationships, normal and failure flows, contracts, data lifecycle, operational behavior, change guidance, tests, evidence, and bounded uncertainty where applicable. Reject pages that are only headings, endpoint/file lists, framework descriptions, or TODOs. Confirm ADRs are not invented when accepted decisions lack evidence.
7. Create a frontend fixture where `CLAUDE.md` says React 18, `package.json` declares React 19, the lockfile resolves React 19, and source imports React. Generate docs and confirm they say React 19 with manifest/lockfile evidence, classify `CLAUDE.md` as stale, and do not rewrite it unless that file was explicitly in scope. Repeat with a manifest range and confirm the agent does not misreport it as an exact resolved version.
8. Create claim-specific conflicts: stale README commands versus manifest scripts, prose routes versus registered routes, and old model prose versus current migrations. Confirm generated docs follow the authoritative current artifact, report objective stale prose, and request validation only when authoritative sources—not merely prose and source—materially conflict.
9. Seed domain and integration names without implementation evidence. Confirm the agent does not fill pages with “likely,” “typically,” or “expected” workflows. It must either trace concrete calls/rules/tests or record an exact known gap with inspected paths. Search generated prose for hedge terms and review every occurrence.
10. In a frontend fixture, confirm the generated guides explain the actual authentication guard mechanism, route/layout composition, server/global/local/form/URL state ownership, query and mutation lifecycle, form/schema validation, design-system composition, BFF/API behavior, and relevant tests. Library-name lists do not pass.
11. Confirm the frontend docs include small source-derived examples for recurring patterns such as a protected page, query/mutation, validated form, BFF request, or component composition. Every snippet must name a current source path and symbol, avoid secrets, and remain faithful to source. Then use the docs to locate the files and tests needed for one representative first change.
12. In a layered Go backend fixture, require a named request trace through the actual gateway or identity boundary, transport middleware, controller, use case, Postgres/Redis or integration adapter, external authorization synchronization, response wrapper, and tests. Confirm each hop explains input, responsibility, state/side effect, and error propagation rather than merely listing directories.
13. For the same backend, confirm source-derived examples cover the real composition-root dependency-injection pattern, controller response/error convention, repository or integration seam, and protecting test pattern. One generic Go snippet or controller catalog does not pass.
14. Add a stateful approval model and a deletion path. Confirm the docs locate transition writes and guards, produce an evidence-backed transition table, and distinguish database cascade/restrict/null rules from application cleanup and external synchronization. Missing evidence must be an exact known gap.
15. Ask the agent to generate or rebuild docs from intentionally shallow existing guides. Confirm its quality assessment triggers deeper inspection and revision in the same task; it must not end by asking whether to begin rebuilding. Then ask only to review the same docs and confirm it reports gaps without modifying them.
16. Run `rebuild --json` and confirm `artifact_kind` is `structural_inventory`, `semantic_documentation_complete` is false, and a semantic next step is present. Confirm the Markdown omits fingerprints and expands common source containers such as `src/app`.
17. Run `audit` before adding routes and confirm it reports `no-knowledge-routes`; populate verified routes and confirm the warning clears.
18. Make a code-only commit. Confirm advisory CI reports a missing decision but passes; switch to acknowledgment and confirm it fails.
19. Run `acknowledge --impact not-required`, commit it, and confirm the same diff passes. Add another material path and confirm the stale fingerprint fails.
20. Change code and relevant documentation together. Confirm acknowledgment mode passes without a no-impact file.
21. Configure an enforced mapping such as `persistence -> docs/data/**`; verify an unrelated doc fails while a matching data doc passes.
22. Put hand-maintained text at the configured generated-inventory path. Confirm `rebuild --apply` refuses to overwrite it.
23. Run `update` with a newer binary and verify the selected adapters and consumer-owned files remain unchanged. Run `update --all-agents` from a subset installation and confirm it expands to every adapter. Switch back to an explicit subset and confirm only obsolete unmodified toolkit-managed adapter files are removed. Confirm an older binary is rejected without `--allow-downgrade`.
24. Install an agent adapter into a disposable repository and confirm both bootstrap helpers are managed files, the Unix helper has mode `0755`, and `doctor` accepts their recorded digests.
25. Run the Unix helper against local fake release assets: verify it refuses an unconfirmed non-interactive install, accepts only a pinned semantic tag, checks `SHA256SUMS`, writes only to a directory already on PATH, and installs mode `0755`. The GitHub CI Windows job must also pass its PowerShell dry-run; before release, exercise the full Windows checksum, confirmation, user-local destination, and explicit user-PATH cases.
26. Run the GitHub adapter integration test over committed base/head SHAs. Confirm advisory mode reports without blocking, acknowledgment mode propagates a failing exit code, invalid modes and object IDs fail closed, and the JSON report remains available.
27. Validate the reusable GitHub workflow contract: read-only contents and attestation permissions, full checkout without persisted credentials, matching pinned release inputs, binary and adapter checksum/attestation verification, ordinary pull-request context, and unconditional JSON artifact upload.
28. Publish a tag and verify every binary plus `repo-knowledge-github-adapter.sh` against `SHA256SUMS`. Call the tagged reusable workflow from a separate GitHub repository, exercise missing-binary recovery from a separate clone, and exercise the GitLab include from a separate project.

Before tagging, ensure `VERSION`, the adapter package version, documentation examples, and changelog agree.

For a GitHub tag, also confirm the read-only build job used Go 1.25.14 and transferred the complete release set, the publishing job alone received contents, identity-token, and attestation write permissions, every file listed in `SHA256SUMS` including `repo-knowledge-github-adapter.sh` and `LICENSE` was attached, and the artifacts have verifiable attestations. Enable private vulnerability reporting, secret scanning, and push protection in the public repository settings before accepting contributions.
