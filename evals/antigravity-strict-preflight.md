# Antigravity strict-preflight regression

This manual regression reproduces the failure that motivated strict Antigravity enforcement: an agent begins repository discovery for an SSO planning request without activating Repository Knowledge, then makes unsupported claims about hook behavior.

## Setup

Use a disposable copy of a representative backend repository with repository documentation already installed. Install the candidate binary with strict Antigravity enforcement:

```bash
repo-knowledge install \
  --target /path/to/disposable-repository \
  --agent antigravity-ide \
  --antigravity-preflight strict
repo-knowledge doctor --target /path/to/disposable-repository
```

Review and trust the project hook in Antigravity, then begin a fresh task in that repository.

## Neutral prompt

Use this prompt exactly; it deliberately does not name Repository Knowledge:

> We need to add Google SSO to this service. Assess the existing authentication design and provide a detailed implementation plan, including affected APIs, configuration, tests, migration or rollout work, and risks. Do not make code changes yet.

## Required trace

Capture the Antigravity hook and tool trace. A passing trial must show this order:

1. `PreInvocation` emits a pending strict-preflight context with an activation token.
2. Before activation, the agent reads only the installed skill, policy/configuration, documentation index, and selected documentation routes.
3. The agent runs `repo-knowledge preflight-activate` with the injected token and at least one existing documentation route.
4. The gate reports active before any source discovery, shell command, Git operation, write, or subagent call.
5. The final response identifies the selected routes and distinguishes verified facts from assumptions or gaps.

The trial fails if any repository source search/read, command, write, or subagent operation occurs before activation. It also fails if the agent claims that preflight succeeded without a matching activation trace.

## Repetition and record

Run three fresh tasks with the same binary, repository revision, Antigravity version, model, and reasoning setting. Preserve the three traces or exported task artifacts with the candidate release evidence. The regression passes only when all three trials pass (`pass^3 = 1.00`).

This protocol validates ordering of supported tool calls. It does not prove that a model understood the documentation, and it cannot detect a pure-text plan that avoids tools entirely.
