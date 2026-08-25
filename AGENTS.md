# Repository knowledge toolkit instructions

This repository is the central toolkit. For repository-level work, use [skills/repository-knowledge/SKILL.md](skills/repository-knowledge/SKILL.md) and the shared contract at [policy/contract.json](policy/contract.json). Start with [docs/index.md](docs/index.md) when routing context.

Keep the shared policy runtime-neutral. Agent behavior belongs in `skills/`, deterministic mechanics in `internal/toolkit/`, CLI presentation in `cmd/`, distribution-specific behavior in `adapters/`, and consuming-repository defaults in `templates/`.

Do not duplicate the full policy into the skill or CI adapter. Preserve the ownership boundary: toolkit updates may replace toolkit-managed installed files but must not overwrite consumer metadata, local rules, semantic documentation, ADRs, scan state, or impact acknowledgments.

Every material toolkit change ends with either reconciled repository documentation or a documented no-impact reason.
