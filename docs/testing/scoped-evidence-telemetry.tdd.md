# Scoped evidence and telemetry TDD evidence

This record covers the implementation of compact context measurement, scoped-work escalation, diff-bound verification receipts, and source-attributed evaluation usage. It preserves the test-first checkpoints required to distinguish intended behavior from tests written after implementation.

## RED: core behavior

Commit: `ddc3fcf` (`test: define scoped evidence and telemetry behavior`)

Command:

```bash
go test ./internal/toolkit ./internal/evalharness
```

The build failed because `ContextMetrics`, the context-profile settings, workflow activation options, evidence commands/receipts, and `TokenUsageDetails` did not exist. The failures were compile-time contract failures in the new tests, not unrelated environment failures.

## GREEN: core behavior

Commit: `a458baf` (`feat: add scoped evidence and telemetry core`)

Command and result:

```text
go test ./internal/toolkit ./internal/evalharness
ok github.com/rustedzone/repository-knowledge/internal/toolkit
ok github.com/rustedzone/repository-knowledge/internal/evalharness
```

The passing behavior includes retained full/compact profiles, content-free measurements, standard/scoped activation, deterministic high-risk escalation, output-digest-only verification records, content-sensitive invalidation, evidence-reference existence checks, documentation decisions, and backward-compatible detailed token usage.

## RED: CLI contract

Commit: `8a75385` (`test: define CLI evidence and usage commands`)

Command:

```bash
go test ./cmd/repo-knowledge ./cmd/repo-knowledge-eval
```

Both packages failed on missing flags: `repo-knowledge` rejected `--preflight-context`, and `repo-knowledge-eval` rejected `--token-source`.

## GREEN: CLI contract

Commit: `2f45266` (`feat: expose evidence and telemetry commands`)

Command and result:

```text
go test ./cmd/repo-knowledge ./cmd/repo-knowledge-eval ./internal/toolkit ./internal/evalharness
ok github.com/rustedzone/repository-knowledge/cmd/repo-knowledge
ok github.com/rustedzone/repository-knowledge/cmd/repo-knowledge-eval
ok github.com/rustedzone/repository-knowledge/internal/toolkit
ok github.com/rustedzone/repository-knowledge/internal/evalharness
```

## Coverage and final verification

Focused coverage after boundary-case additions:

```text
internal/toolkit/evidence.go: 139/161 statements = 86.3%
internal/evalharness/validateTokenUsage: 96.9%
```

The whole packages remain below those values because they include unrelated legacy surfaces. The feature-specific implementation exceeds the 80% affected-scope threshold.

Final repository check:

```text
make check
go test ./...: pass
go vet ./...: pass
go test -race ./...: pass
```

## Security and evidence boundaries

- Verification commands are an explicit argument vector after `--`; the runtime does not add a shell.
- Session state stores command metadata and an output SHA-256, never command output, prompt content, source content, or documentation content.
- The worktree fingerprint includes current file content, so a post-test edit invalidates the check.
- Provider token counts are labeled by source; bytes and characters are never promoted to measured tokens.
- Evidence references establish path existence and attribution only. They do not prove that an optional anchor is semantically correct.
- The completion receipt is invoked by the agent or user. It is not yet a universal host-level stop hook, so pure-text completion remains outside this deterministic boundary.
