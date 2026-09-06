# Extension guide

Add an extension at the narrowest layer that owns its concern:

- New evidence or safety semantics belong in the shared contract with a contract-version review.
- Deterministic detectors and validators belong in `internal/toolkit` with table-driven or integration tests.
- CLI presentation belongs in `cmd/repo-knowledge`.
- Embedded managed assets must be included by `assets.go`; avoid creating a second copy solely for embedding.
- Agent-only routing belongs in the appropriate skill and should reference, not duplicate, shared policy.
- GitHub or another SCM integration belongs under `adapters/` and should invoke the same released binary.
- Repository-specific business rules belong in the consuming repository's local files, not central defaults.

Default invariants must be portable and objective. Capability-specific defaults may activate only when a capability is declared or reliably detected. A classifier token is not sufficient evidence for a blocking semantic invariant.

When changing installed managed assets, keep `install` and `update` behavior symmetric and add an ownership-preservation test. When changing a release filename, update the Makefile, release workflow, GitHub reusable workflow and adapter, GitLab adapter, checksum verification, tests, and installation documentation together.

## Adding an agent adapter

An agent adapter consists of the smallest native always-on rule that routes repository work to the shared skill, a native project skill destination, and—when the host supports it—a lifecycle hook that injects the preflight before normal discovery. Keep semantic behavior in `skills/repository-knowledge/`; adapter rules should route to that skill rather than duplicate it, while hook protocol encoding belongs in `internal/toolkit`.

When adding an adapter:

1. Add its canonical `--agent` name and any deliberate aliases to `internal/toolkit/install.go`.
   Add the canonical name to `supportedAgentAdapters` so `--all-agents` includes it in stable manifest order.
2. Map embedded rules and the shared skill to the agent's documented project locations.
3. If the host has lifecycle hooks, register the earliest context-injection event, encode the host's output protocol in `hook-context`, preserve every unrelated field in shared hook files, and keep host trust prompts visible. If the host can deny tool use, keep the gate adapter-specific and require an explicit opt-in mode, bounded state, deterministic recovery, and a narrow knowledge-read allowlist.
4. Add adapter-aware `doctor` checks. Exclude fully toolkit-managed adapter files from scan and impact analysis; exclude shared hook containers only when they contain no consumer-owned configuration.
5. Verify single-adapter, multi-adapter, update-preservation, explicit adapter-switch, malformed-hook, consumer-hook preservation, context-output, gate activation, expiry, cross-repository isolation, and missing-registration behavior.
6. Document the native paths, trust/reload behavior, binary-on-PATH requirement, and installation command.
