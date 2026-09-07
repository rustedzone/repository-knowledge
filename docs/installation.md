# Installation

## Runtime requirements

- A released `repo-knowledge` executable for the current OS and architecture.
- Git for diff analysis and efficient tracked-file fingerprints.

Consumers do not need Go, Python, or third-party packages. Building the toolkit from source requires the Go version declared by `go.mod` or newer.

## Release artifacts

Tagged GitHub workflows publish these files to GitHub Releases. The existing GitLab pipeline can publish the same set to the Generic Package Registry:

```text
repo-knowledge-linux-amd64
repo-knowledge-linux-arm64
repo-knowledge-darwin-amd64
repo-knowledge-darwin-arm64
repo-knowledge-windows-amd64.exe
repo-knowledge-github-adapter.sh
LICENSE
SHA256SUMS
```

Download the matching executable, license, and checksum file from the pinned release. Verify the checksum before placing the executable on `PATH`. Retain the license with redistributed binaries.

GitHub release artifacts also receive build-provenance attestations. After downloading an artifact, verify both its checksum and attestation:

```bash
shasum -a 256 -c SHA256SUMS --ignore-missing
gh attestation verify repo-knowledge-darwin-arm64 --repo rustedzone/repository-knowledge
```

The attestation is bound to the canonical `rustedzone/repository-knowledge` source repository.

## Missing-binary recovery

A cloned repository can contain `.repo-knowledge/` and an installed agent skill while the new machine has no `repo-knowledge` command. The skill handles this state only when the current task needs a deterministic CLI operation:

1. It checks whether `repo-knowledge` resolves by command name.
2. It reads the exact source and semantic release ref from `.repo-knowledge/toolkit.json`. Legacy manifests whose source is `release-binary` resolve to the canonical GitHub repository. It never chooses `latest`.
3. It runs the installed bootstrap helper in dry-run mode to resolve the platform artifact and destination without using the network or writing files.
4. It tells the user the source, version, artifact, destination, verification steps, replacement behavior, and any Windows user-PATH change, then asks for permission.
5. Only after approval, it downloads the binary and `SHA256SUMS` over HTTPS, checks the binary digest, verifies the downloaded binary reports the pinned version, verifies the GitHub artifact attestation when a compatible `gh` command is available, and writes the executable.
6. It resolves `repo-knowledge` by name and verifies the version again. A newly persisted Windows user PATH may require an open terminal or agent host to restart; the helper has already verified the exact installed artifact before that point.

On macOS and Linux, `skills/repository-knowledge/scripts/install-binary.sh` chooses `~/.local/bin` or `~/bin` only when that directory is already on the current `PATH`. It does not call `sudo` or edit a shell startup file. If neither directory is on `PATH`, the agent must ask whether to use another user-owned PATH directory or let the user configure PATH.

On Windows, `skills/repository-knowledge/scripts/install-binary.ps1` defaults to `%LOCALAPPDATA%\Programs\repo-knowledge\bin`. Because that directory is not normally on PATH initially, the agent's permission request must explicitly include adding it to the current user's PATH; already-open terminals may need to restart.

If the manifest is absent, its ref is not an exact semantic tag, or it names a noncanonical source, the skill does not guess. It asks the user to confirm an exact trusted release. Declining installation does not prevent source inspection or semantic documentation work, but CLI-dependent operations remain unavailable and must be reported as such.

## Install into a repository

Codex is the default adapter:

```bash
repo-knowledge install \
  --target /path/to/service-repository \
  --agent codex \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.11.0
```

Select Claude Code, Antigravity IDE, or Cursor with their canonical names:

```bash
repo-knowledge install --target /path/to/service-repository --agent claude-code
repo-knowledge install --target /path/to/service-repository --agent antigravity-ide
repo-knowledge install --target /path/to/service-repository --agent cursor
```

To support several agents in one repository, repeat the option:

```bash
repo-knowledge install \
  --target /path/to/service-repository \
  --agent codex \
  --agent claude-code \
  --agent antigravity-ide \
  --agent cursor
```

To install every supported adapter, use the single preference flag:

```bash
repo-knowledge install \
  --target /path/to/service-repository \
  --all-agents
```

`--all-agents` and `--agent` are mutually exclusive. The all-agents selection is recorded as the canonical adapter list in `.repo-knowledge/toolkit.json`.

The aliases `claude` and `antigravity` are accepted and recorded in `.repo-knowledge/toolkit.json` as `claude-code` and `antigravity-ide`.

Installation also registers a native repository preflight hook for each selected adapter:

| Agent | Shared hook container | Event |
| --- | --- | --- |
| Codex | `.codex/hooks.json` | `SessionStart` (and `PreToolUse` in strict mode) |
| Claude Code | `.claude/settings.json` | `SessionStart` (and `PreToolUse` in strict mode) |
| Antigravity IDE | `.agents/hooks.json` | `PreInvocation` (and `PreToolUse` in strict mode) |
| Cursor | `.cursor/hooks.json` | `sessionStart` (and `preToolUse` in strict mode) |

The installer merges only its exact `repo-knowledge hook-context` entry. Existing setting values and consumer hooks are preserved, and switching adapters removes only the obsolete toolkit entry. Invalid existing JSON stops installation rather than overwriting the file.

The default `full` hook profile maximizes self-contained context. Use `--preflight-context compact` to inject a smaller routing payload and bounded documentation index; the agent must still read the selected contract and routes from disk. Updates retain the selected profile. Compare actual payload sizes and generation time without exposing content:

```bash
repo-knowledge hook-context --target . --agent codex --metrics
```

Every adapter uses observational preflight by default. Enable tool-level repository preflight enforcement explicitly for each selected agent:

```bash
repo-knowledge install \
  --target /path/to/service-repository \
  --agent codex \
  --agent claude-code \
  --agent antigravity-ide \
  --agent cursor \
  --agent-preflight codex=strict \
  --agent-preflight claude-code=strict \
  --agent-preflight antigravity-ide=strict \
  --agent-preflight cursor=strict
```

Strict mode adds each host's managed tool gate alongside its lifecycle hook. The first hook invocation injects an opaque activation token. Before activation, the gate permits only Repository Knowledge and routed-documentation reads, supported external research tools, and the exact activation command; it denies repository discovery, commands, writes, and subagents. The agent activates after selecting documentation routes, for example:

```bash
repo-knowledge preflight-activate \
  --token <injected-token> \
  --route docs/index.md \
  --route docs/architecture.md \
  --workflow standard
```

`standard` is the default. Use `--workflow scoped` only for a bounded low-risk fix with known validation. A scoped completion report automatically rejects an actual diff classified as API, authorization, persistence, integration, configuration, or deployment work; reactivate the same token with `--workflow standard` when that happens.

For a strict session, bind validation and completion claims to the current content rather than a changed-path estimate:

```bash
repo-knowledge evidence-run --token <injected-token> --label unit-tests -- go test ./...
repo-knowledge evidence-report --token <injected-token> \
  --evidence internal/example.go#Symbol \
  --documentation-impact not-required \
  --reason "Internal refactor; documented behavior is unchanged."
```

`evidence-run` executes the argument vector after `--` directly and hashes its combined output without persisting that output in session state. Any later content change changes the worktree fingerprint and makes the verification stale. Use `--documentation-impact required --documentation-file docs/example.md` when the current diff includes reconciled documentation.

`update` keeps existing per-agent modes when the flag is omitted. Use `--agent-preflight cursor=observe`, for example, to remove only that toolkit-managed gate and return that adapter to injection-only behavior. `--antigravity-preflight` remains a compatibility alias for Antigravity only.

The hook command requires `repo-knowledge` to resolve on the host process's `PATH`. Review and trust the new project hook when the agent asks; do not bypass the host's trust boundary. Start a new session after installation. Cursor normally reloads a saved hook file, but restarting the host is the reliable fallback. If a hook cannot run, the native project rule still requires a manual preflight and the agent should not silently install a missing binary. Permission-gated binary recovery applies only when the current task actually needs a CLI operation.

Then:

```bash
repo-knowledge doctor --target /path/to/service-repository
repo-knowledge scan --target /path/to/service-repository
repo-knowledge rebuild --target /path/to/service-repository
```

`doctor` checks the rule, skill, managed-file digests, lifecycle hook, and strict gate registration for every selected adapter. `repo-knowledge doctor --live-hooks` additionally performs an in-process pending → activation → active gate self-test for every strict adapter; it verifies toolkit behavior but cannot prove that an external host trusted or invoked its configured hook. In a new repository task, the agent's first progress update should include `Repository knowledge preflight: loaded` and the selected documentation routes. That receipt is the user-visible confirmation that activation happened; native hooks remain a guardrail rather than proof of model understanding.

`scan` and `rebuild` produce structural discovery artifacts, not human-readable repository documentation. To generate documentation, ask the installed agent to classify every repository shape, inspect evidence across all material surfaces, write detailed profile-specific guides, populate `docs/index.md`, and record verified capability routes. For example:

```text
Use repository-knowledge to generate comprehensive, human-readable documentation for this entire repository. Cover every applicable repository profile and material domain, feature, interface, data model, integration, and operational surface. Do not stop after scan, rebuild, or a shallow overview.
```

Review `.repo-knowledge/rebuild-proposal.md`. If the optional structural appendix is useful, apply it with:

```bash
repo-knowledge rebuild --target /path/to/service-repository --apply
```

Edit `.repo-knowledge/repository.json` to declare verified capabilities and route them to existing documentation. Replace the placeholder section in `docs/index.md` with capability-driven routes.

## Existing documentation

Installation preserves all existing docs and consumer-owned agent instructions. If `docs/index.md` already exists, it is untouched. If it does not exist, the toolkit adds a routing index alongside existing files. The Codex adapter replaces only its marked block in `AGENTS.md`; unrelated content remains intact. Claude Code, Antigravity IDE, and Cursor use dedicated toolkit-managed rule files, leaving other files in `.claude/rules/`, `.agents/rules/`, and `.cursor/rules/` untouched. Native hook JSON files are shared containers; only the toolkit's nested registration is managed. Existing documentation does not need to be relocated into a prescribed taxonomy.

## Updating

Download and verify the new release binary, then run:

```bash
repo-knowledge update \
  --target /path/to/service-repository \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.11.0
```

When `--agent` and `--all-agents` are omitted, `update` keeps the adapter selection recorded by the existing installation. Supply one or more `--agent` options to select a subset, or `--all-agents` to switch the installation to every currently supported adapter. Unmodified obsolete toolkit-managed adapter files are removed, modified obsolete files are preserved and reported, and consumer-owned instructions remain untouched.

The new binary contains its own templates, policy, schemas, and skills, so no central source checkout is required. Run `doctor`, `scan`, and `audit` after updating. Change any GitHub reusable-workflow tag, `toolkit-version`, GitLab include ref, and GitLab package version in the same merge request.

## GitHub Actions CI

Copy [the consumer example](../examples/github/repository-knowledge.yml) to `.github/workflows/repository-knowledge.yml` in the consuming repository:

```yaml
name: Repository knowledge

on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read
  attestations: read

jobs:
  documentation-impact:
    uses: rustedzone/repository-knowledge/.github/workflows/documentation-check.yml@v0.11.0
    with:
      toolkit-version: v0.11.0
      enforcement: advisory
```

Update `uses` and `toolkit-version` together. The workflow ref selects the CI orchestration; `toolkit-version` selects the checksummed and attested binary plus `repo-knowledge-github-adapter.sh` from GitHub Releases. The example uses the convenient release tag. Organizations requiring immutable workflow references should replace `@v0.11.0` with the full commit SHA for that release while retaining `toolkit-version: v0.11.0`.

The reusable workflow:

- checks out the caller with `fetch-depth: 0` and without persisted credentials;
- uses the pull-request base/head SHAs or push before/current SHAs, with empty-tree handling for an initial push;
- supports Linux `amd64` and `arm64` runners internally;
- verifies `SHA256SUMS`, the binary's reported version, and GitHub artifact attestations;
- pins its GitHub-owned checkout and artifact-upload dependencies to full commit SHAs;
- uploads `repository-knowledge-impact.json` for 14 days, including when acknowledgment or enforced validation blocks the job; and
- never commits, opens a pull request, or writes documentation.

The caller grants only `contents: read` and `attestations: read`. Reusable workflows cannot elevate permissions granted by their caller, so both are declared explicitly. No secrets are required. Use the ordinary `pull_request` event; do not change the example to `pull_request_target`, which has a more privileged security model and is unnecessary for this read-only validator.

For unusual events, optional `base-sha` and `head-sha` inputs accept explicit 40- or 64-character Git object IDs. Those objects must already exist in the full checkout: the workflow deliberately does not persist credentials or fetch arbitrary caller-supplied objects. `retention-days` changes report retention. Start with `advisory`, then move to `acknowledgment` or `enforced` using the same rollout criteria as GitLab.

## GitLab CI

Add this to the consuming `.gitlab-ci.yml`:

```yaml
include:
  - project: engineering/repository-knowledge
    ref: v0.11.0
    file: /adapters/gitlab/documentation-check.yml

variables:
  REPO_KNOWLEDGE_TOOLKIT_PROJECT_ID: "12345"
  REPO_KNOWLEDGE_TOOLKIT_VERSION: v0.11.0
  REPO_KNOWLEDGE_ENFORCEMENT: advisory
```

`REPO_KNOWLEDGE_TOOLKIT_PROJECT_ID` is the numeric GitLab project ID of the central toolkit. The consuming project must be allowed to download its Generic Package Registry artifacts with `CI_JOB_TOKEN`; configure the toolkit project's job-token allowlist when cross-project access is restricted.

The job detects Linux `amd64` or `arm64`, downloads the pinned binary and `SHA256SUMS`, verifies the artifact, analyzes the merge-request or push diff, uploads a JSON report, and never commits changes. It uses full Git history for reliable diff-base access. Repositories with custom stages that omit `test` must override the included job's stage.
