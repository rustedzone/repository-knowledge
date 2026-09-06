# Configuration and contract

## File map in a consuming repository

| Path | Owner | Purpose |
| --- | --- | --- |
| `.repo-knowledge/policy/*` | Toolkit | Installed shared policy and defaults. |
| `.repo-knowledge/schemas/*` | Toolkit | Installed JSON Schema interfaces. |
| `.agents/skills/repository-knowledge/*` | Toolkit | Shared Codex and Antigravity IDE skill. |
| `.claude/skills/repository-knowledge/*` | Toolkit | Claude Code skill. |
| `.cursor/skills/repository-knowledge/*` | Toolkit | Cursor skill. |
| `.agents/rules/repository-knowledge.md` | Toolkit | Antigravity IDE prompt-routing rule. |
| `.claude/rules/repository-knowledge.md` | Toolkit | Claude Code prompt-routing rule. |
| `.cursor/rules/repository-knowledge.mdc` | Toolkit | Cursor always-applied prompt-routing rule. |
| `.codex/hooks.json` nested `SessionStart` entry | Shared | Toolkit owns only its exact Codex preflight command; all unrelated hook configuration is consumer-owned. |
| `.claude/settings.json` nested `SessionStart` entry | Shared | Toolkit owns only its exact Claude Code preflight command; all unrelated settings and hooks are consumer-owned. |
| `.agents/hooks.json` `repository-knowledge-preflight` entry | Shared | Toolkit owns only its named Antigravity preflight entry; all other hook configuration is consumer-owned. |
| `.cursor/hooks.json` nested `sessionStart` entry | Shared | Toolkit owns only its exact Cursor preflight command; all unrelated hook configuration is consumer-owned. |
| `.repo-knowledge/toolkit.json` | Toolkit-generated | Version, source, agent adapters, ownership manifest, and managed-file digests. |
| `AGENTS.md` managed block | Toolkit | Codex prompt routing; surrounding instructions remain consumer-owned. |
| `.repo-knowledge/repository.json` | Consumer | Repository identity, knowledge routes, and CI mode. |
| `.repo-knowledge/local-invariants.json` | Consumer | Repository-specific objective or semantic contracts. |
| `.repo-knowledge/local-impact-rules.json` | Consumer | Additional globs, classifiers, and enforced mappings. |
| `.repo-knowledge/scan-state.json` | Consumer-generated | Last structural inventory and module fingerprints. |
| `.repo-knowledge/doc-impact.json` | Consumer-generated | Decision bound to the current material changed-path fingerprint. |
| `docs/**` | Consumer | Generated catalogs and maintained semantic knowledge. |

The release binary is installed on developer machines and CI runners, not committed into the consuming repository. The managed skill includes platform bootstrap helpers so a clone with a missing command can recover its pinned binary after explicit user permission; those helpers remain toolkit-owned and their digests are recorded in the manifest.

Toolkit-managed adapter rules and skills are excluded from structural scan inventory and ignored by default documentation-impact classification. A shared hook container is excluded from scan only when removing the toolkit entry leaves it empty; a container with consumer settings remains normal repository evidence. Hook containers are not blanket-ignored by documentation-impact classification because that would conceal consumer hook changes.

The generated inventory is an optional structural appendix. Semantic guides, `docs/index.md`, and capability routes are consumer-owned knowledge maintained by agents and humans after evidence inspection. An empty `capabilities` list causes `audit` to report `no-knowledge-routes`.

## Capabilities

A verified capability entry routes agents and audits to real documentation and evidence:

```json
{
  "id": "public-api",
  "documentation": ["docs/api/public-api.md"],
  "evidence": ["openapi/openapi.json", "src/http/"]
}
```

Do not copy scan signals directly into this list without verification. Capability absence means “not declared,” not “does not exist.”

## Local impact mappings

In enforced mode, a consumer may require a classified change to touch a particular documentation route:

```json
{
  "required_doc_mappings": {
    "persistence": ["docs/data/**"],
    "authorization": ["docs/security/authorization.md"]
  }
}
```

Only configure mappings deterministic enough to block a merge. Keep fuzzy semantic checks advisory.

## Impact acknowledgments

Run the binary instead of editing the fingerprint by hand:

```bash
repo-knowledge acknowledge \
  --impact not-required \
  --reason "Internal refactor; public behavior and operating model are unchanged."
```

The fingerprint covers material changed paths and statuses while excluding documentation and the acknowledgment itself. Any later material path change invalidates the decision.
