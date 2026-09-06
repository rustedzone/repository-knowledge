# Repository knowledge operations

## Bootstrap and install

Run `repo-knowledge install --target PATH` for the default Codex adapter. Use `--agent claude-code` for Claude Code, `--agent antigravity-ide` for Antigravity IDE, or `--agent cursor` for Cursor; repeat `--agent` to install multiple adapters. The aliases `claude` and `antigravity` are also accepted.

Use `repo-knowledge install --target PATH --all-agents` to install every supported adapter. `--all-agents` is mutually exclusive with `--agent`. On update, omitting both preferences retains the manifest's current selection; `update --all-agents` deliberately expands it to the complete canonical list.

Each selected adapter also receives a native lifecycle preflight registration in its documented project hook container. The toolkit manages only its exact nested command and preserves unrelated consumer hook configuration. The hook needs `repo-knowledge` on the host PATH and may require host review or trust before it runs. After installation, run `doctor`, start a fresh agent session, and confirm the first progress update reports `Repository knowledge preflight: loaded` with the selected routes. If native context is unavailable, follow the installed rule's manual preflight; do not silently bootstrap a missing binary.

Installation preserves existing documentation, repository-owned configuration, and unrelated agent rules. It creates a routing index only when one does not exist. Codex uses a managed block in `AGENTS.md`; Claude Code uses `.claude/rules/repository-knowledge.md`; Antigravity IDE uses `.agents/rules/repository-knowledge.md`; Cursor uses an always-applied `.cursor/rules/repository-knowledge.mdc` rule. Each adapter receives the skill in its native project skill directory.

## Scan

Run `repo-knowledge scan`. Scan builds structural discovery data from tracked paths, high-signal manifests, documentation paths, module fingerprints, and capability indicators. It writes `.repo-knowledge/scan-state.json`; it does not claim semantic truth or generate human-readable repository documentation.

Inspect source more deeply only for changed, unknown, or relevant modules. A detected capability is an inspection lead, not a verified architectural claim.

## Audit

Run `repo-knowledge audit`. Audit checks installation health, configured documentation routes, index links, scan freshness, and (when a diff range is provided) documentation impact. Report contradictions with their evidence and claim state. Do not auto-resolve semantic conflicts.

## Rebuild

Run `repo-knowledge rebuild` to produce `.repo-knowledge/rebuild-proposal.md`. The proposal is a mechanically derived structural appendix with explicit uncertainty. It is discovery input for an agent, not a repository overview, architecture guide, or completed documentation set.

Run `repo-knowledge rebuild --apply` only to create or replace the optional generated inventory at the configured path. If that path contains non-generated content, stop and preserve it. Applying the inventory never completes a request to generate repository documentation; follow [documentation-generation.md](documentation-generation.md) to inspect evidence, write semantic guides, and populate knowledge routes.

When a user asks the agent to “rebuild the repository documentation,” treat that as a full semantic documentation request and follow the comprehensive workflow plus repository-type profiles. Invoke the CLI `rebuild` operation only for the optional structural appendix; the shared word “rebuild” must not reduce the user's semantic request to an inventory operation.

## Documentation reconciliation

Use deterministic generation for catalogs that are reliably derivable from source. Maintain semantic explanations for domains, behavior, business rules, cross-component flows, operational expectations, and architectural tradeoffs. Avoid duplicating a source-derived catalog as hand-maintained prose.

When stronger evidence contradicts documentation:

- use `verified_stale` for an objective mismatch that can be safely corrected;
- use `conflict` when authoritative-looking sources disagree and intent matters;
- keep `requirement_only` when the document explicitly describes intended but unimplemented behavior;
- keep `historical` decision records, marking supersession rather than rewriting history.
