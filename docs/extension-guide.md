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

An agent adapter consists of the smallest native always-on rule that routes repository work to the shared skill, plus a native project skill destination. Keep the lifecycle semantics in `skills/repository-knowledge/`; adapter rules should route to that skill rather than duplicate it.

When adding an adapter:

1. Add its canonical `--agent` name and any deliberate aliases to `internal/toolkit/install.go`.
   Add the canonical name to `supportedAgentAdapters` so `--all-agents` includes it in stable manifest order.
2. Map embedded rules and the shared skill to the agent's documented project locations.
3. Add adapter-aware `doctor` checks and exclude toolkit-managed adapter files from scan inventory and documentation-impact classification.
4. Verify single-adapter, multi-adapter, update-preservation, explicit adapter-switch, and consumer-rule preservation behavior.
5. Document the native paths and installation command.
