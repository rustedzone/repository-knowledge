# Strict Antigravity preflight TDD evidence

## Source and user journeys

This change was derived from the reviewed Antigravity Google SSO planning failure and its approved implementation plan; no separate plan file was supplied.

1. As a repository user, I want Antigravity to block repository discovery until Repository Knowledge is activated, so that planning starts from routed evidence.
2. As a repository user, I want activation to be scoped to one repository and conversation and to expire when knowledge inputs change, so that stale or cross-repository state cannot unlock tools.
3. As a toolkit maintainer, I want installation and `doctor` to expose strict-mode wiring and self-test it deterministically, so that broken registrations are visible before a live agent run.

## RED and GREEN evidence

The initial focused RED run was:

```text
go test ./internal/toolkit -run 'TestStrictAntigravity' -count=1
```

It failed to compile because `AntigravityPreflightStrict`, `PreflightGate`, and `PreflightActivate` did not exist. That is the intended missing-feature signal.

After implementation, focused behavior was GREEN:

```text
go test ./internal/toolkit -run 'Test(Strict|Observe|Preflight)Antigravity' -count=1
```

The package passed. The full repository verification was also GREEN:

```text
make check
```

It completed `go test ./...` and `go vet ./...` successfully.

## Test specification

| Guarantee | Test target | Result |
| --- | --- | --- |
| Pending strict preflight denies repository discovery and permits selected knowledge reads | `TestStrictAntigravityPreflightBlocksDiscoveryUntilActivation` | PASS |
| Successful activation allows normal repository discovery and later invocations remain active | `TestStrictAntigravityPreflightBlocksDiscoveryUntilActivation` | PASS |
| Tokens cannot activate another repository and expired tokens fail | `TestStrictAntigravityPreflightRejectsCrossRepositoryActivation`; `TestStrictAntigravityPreflightExpiresBeforeActivation` | PASS |
| Only documentation routes activate a session; duplicate routes are normalized | `TestStrictAntigravityPreflightRejectsSourceRoutesAndDeduplicatesDocumentationRoutes` | PASS |
| Pending state permits external research and knowledge reads but denies source reads | `TestStrictAntigravityGateAllowsOnlyKnowledgeReadsAndExternalResearchWhilePending` | PASS |
| Shell-chained activation commands are rejected | `TestStrictAntigravityGateOnlyAllowsSafeActivationCommand` | PASS |
| Strict mode registers the gate, `doctor` detects it, and `doctor --live-hooks` executes pending → active behavior | `TestStrictAntigravityInstallRegistersGateAndDoctorRequiresIt`; `TestStrictAntigravityDoctorLiveHooksExercisesPendingAndActiveGate`; `TestRunInstallStrictAntigravityPreflightAndDoctorLiveHooks` | PASS |
| Observe mode remains non-blocking and updates retain recorded strict mode | `TestObserveAntigravityPreflightDoesNotGateToolsAndUpdatesRetainStrictMode` | PASS |

## Coverage and known limits

Focused coverage was measured with:

```text
go test -coverprofile=/tmp/repository-knowledge-preflight.cover ./internal/toolkit
```

`internal/toolkit/preflight.go` reached 80.0% statement coverage (188/235). The existing whole-package coverage is 76.2%, and the CLI package is 46.4%; this repository does not currently enforce a global coverage threshold, so those pre-existing package-level baselines remain a follow-up rather than being represented as satisfied by this feature.

The manual [Google SSO Antigravity regression](../../evals/antigravity-strict-preflight.md) still requires three fresh host traces before release. Automated tests validate the binary and hook protocol, not whether a live host trusted its hooks or whether the model understood the loaded documentation.
