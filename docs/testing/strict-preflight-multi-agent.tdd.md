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

## Resolved-path containment regression

### Scope and user journey

As a user enabling strict preflight for an untrusted repository, I want pending knowledge reads to be checked after symlink resolution so that a path which appears to be documentation cannot reach content outside its permitted knowledge directory.

This fix is shared by every adapter through `knowledgePathReadAllowed`; it does not change hook schemas, session handling, installation, or the shared policy contract.

### RED checkpoint

Commit `9fc6e43` added `TestStrictPreflightRejectsResolvedKnowledgePathEscapes`. Before the production change, the focused test reported that:

- relative and absolute symlinks from `docs/` to a non-knowledge repository path were allowed;
- an absolute symlink from `docs/` to a path outside the repository was allowed; and
- the Codex strict gate returned `allow` for that external target.

The same test confirmed that ordinary absolute documentation paths and symlinks whose targets remained in `docs/` were expected to stay allowed.

### GREEN checkpoint

Commit `14a7a15` routes absolute inputs through the existing repository-containment validator, resolves the requested path and allowed knowledge roots, and compares those resolved paths. The focused regression and existing absolute-path compatibility test passed:

```text
go test ./internal/toolkit -run 'Test(StrictPreflightRejectsResolvedKnowledgePathEscapes|PreflightGateSupportsOtherAgentsAndAllowsAbsoluteKnowledgePaths)' -count=1
ok github.com/rustedzone/repository-knowledge/internal/toolkit
```

### Test specification

| Guarantee | Evidence | Type | Result |
| --- | --- | --- | --- |
| Ordinary absolute documentation paths remain readable while pending | `TestPreflightGateSupportsOtherAgentsAndAllowsAbsoluteKnowledgePaths` | Unit | PASS |
| Relative and absolute symlinks that remain inside the documentation directory remain readable | `TestStrictPreflightRejectsResolvedKnowledgePathEscapes` | Unit | PASS |
| Relative and absolute symlinks that leave the permitted knowledge directory are denied | `TestStrictPreflightRejectsResolvedKnowledgePathEscapes` | Unit | PASS |
| An absolute documentation symlink that resolves outside the repository is denied by the host-facing gate | `TestStrictPreflightRejectsResolvedKnowledgePathEscapes` | Integration | PASS |

### Verification and coverage

`make check` passed all Go tests, formatting verification, and `go vet ./...`. Toolkit package coverage was 76.4%, consistent with the established package baseline; the changed `preflightPath` resolver measured 86.7% statement coverage. A binary-level disposable-repository reproduction returned `allow` for the real absolute documentation index and `deny` for both absolute and relative symlink escapes.

The path check remains a host-facing gate decision rather than an opened file descriptor, so it does not claim to eliminate operating-system-level time-of-check/time-of-use races. Addressing that would require host integration beyond this focused regression.
