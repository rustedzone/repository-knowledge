# Strict-preflight regression

This manual regression verifies the host boundary that deterministic unit tests cannot prove: the selected host invokes and honors its configured strict gate.

Use a disposable copy of a documented backend repository. Install the candidate binary for one host at a time, for example:

```bash
repo-knowledge install \
  --target /path/to/disposable-repository \
  --agent codex \
  --agent-preflight codex=strict
```

Repeat with `claude-code`, `antigravity-ide`, and `cursor`. Start a fresh trusted project session, then use the neutral request: “Plan Google SSO for this repository. Do not make changes.”

A passing trace shows this order:

1. The lifecycle hook injects a pending strict-preflight context and token.
2. A repository source read, search, shell command, write, or subagent attempt is denied by that host's native gate.
3. The agent selects documentation routes and runs the exact injected `repo-knowledge preflight-activate` command.
4. Repository discovery is permitted only after activation.
5. The first progress update reports `Repository knowledge preflight: loaded` and names the routes.

The trial fails if repository discovery or mutation occurs before activation, the host does not invoke the configured gate, or the agent claims activation without a matching command trace. Run three fresh tasks for each host with the same repository revision, host version, model, and reasoning setting. Retain the traces with release evidence; each host requires three passing trials.

This regression proves host invocation only. It does not prove model comprehension, block pure-text answers, or extend a host's local gate to hosted tools that bypass that hook path.
