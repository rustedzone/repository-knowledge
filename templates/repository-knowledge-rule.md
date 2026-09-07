# Repository knowledge lifecycle

For repository-related work, use the `repository-knowledge` skill before assessing or planning the change. Start at `docs/index.md`, load only relevant knowledge, and verify material claims against repository evidence using the precedence in `.repo-knowledge/policy/contract.json`.

Do not begin normal source discovery until this preflight is complete. In the first progress update, report `Repository knowledge preflight: loaded` and name the documentation routes selected for the request. If native hook context was not available, perform the same preflight from these files manually. A missing binary does not authorize a silent installation; follow the skill's permission-gated bootstrap only when the task needs a CLI operation.

After implementation and validation, assess documentation impact. End every material change with either updated documentation or a justified `documentation impact: not required` decision. Ask for human validation when evidence conflicts, requirements are materially ambiguous, business semantics cannot be inferred safely, or a destructive documentation change is proposed.

Use the standard workflow unless a fix is clearly bounded and low risk. Scoped work must escalate to standard when the actual diff affects API, authorization, persistence, integration, configuration, or deployment surfaces. Do not claim implementation completion without the actual changed files, checks run after the final edit, source/config/test evidence, and the documentation-impact decision. With an active strict-preflight token, use `repo-knowledge evidence-run` and `repo-knowledge evidence-report`; any later content change invalidates prior verification evidence.

Do not treat generated or stale documentation as stronger than executable behavior, schemas, migrations, effective configuration, or source implementation. Preserve requirement-only and historical material unless there is an explicit decision to change it.
