# Strict preflight across agents: TDD evidence

## Scope

Extend opt-in strict Repository Knowledge preflight from Antigravity IDE to Codex, Claude Code, and Cursor without changing the shared policy contract or consumer-owned files.

## RED checkpoint

Commit `0506262` added the reproducer `internal/toolkit/preflight_multi_agent_test.go` before production changes. The focused test failed because there was no per-agent preflight mode, persisted mode map, or host-neutral token extraction/session behavior:

```text
unknown field PreflightModes in InstallOptions
ToolkitManifest.PreflightModes undefined
undefined: preflightTokenFromText
```

## GREEN checkpoint

Commit `0388f50` introduced:

- persisted per-agent `preflight_modes`, retaining the legacy Antigravity field for installed repositories;
- an explicit repeatable `--agent-preflight AGENT=observe|strict` setting and compatible `--antigravity-preflight` alias;
- sessions bound to repository, adapter, host conversation, expiry, and knowledge-input digests;
- native gate output for Codex and Claude Code (`hookSpecificOutput.permissionDecision`), Cursor (`permission`), and Antigravity (`decision`);
- managed strict gate registration, doctor checks, and live pending → active checks for every selected strict adapter.

The original reproducer passed after the change:

```text
go test ./internal/toolkit -run 'Test(StrictPreflightBlocksRepositoryToolsForEverySupportedAgent|PreflightModesRetainStrictSettingsForEverySelectedAdapter)' -count=1
ok github.com/rustedzone/repository-knowledge/internal/toolkit
```

## Verification

```text
go test ./...
ok github.com/rustedzone/repository-knowledge
ok github.com/rustedzone/repository-knowledge/cmd/repo-knowledge
ok github.com/rustedzone/repository-knowledge/internal/evalharness
ok github.com/rustedzone/repository-knowledge/internal/toolkit
```

Manual host-trace acceptance remains required before release: [strict-preflight regression](../../evals/strict-preflight.md).

Focused statement coverage for `internal/toolkit/preflight.go` was 80.2% (210/262); whole-package coverage was 75.8%, consistent with the repository's existing package-level baseline and not treated as a new global threshold.
