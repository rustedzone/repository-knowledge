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

## Install into a repository

Codex is the default adapter:

```bash
repo-knowledge install \
  --target /path/to/service-repository \
  --agent codex \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.6.1
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

Then:

```bash
repo-knowledge doctor --target /path/to/service-repository
repo-knowledge scan --target /path/to/service-repository
repo-knowledge rebuild --target /path/to/service-repository
```

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

Installation preserves all existing docs and consumer-owned agent instructions. If `docs/index.md` already exists, it is untouched. If it does not exist, the toolkit adds a routing index alongside existing files. The Codex adapter replaces only its marked block in `AGENTS.md`; unrelated content remains intact. Claude Code, Antigravity IDE, and Cursor use dedicated toolkit-managed rule files, leaving other files in `.claude/rules/`, `.agents/rules/`, and `.cursor/rules/` untouched. Existing documentation does not need to be relocated into a prescribed taxonomy.

## Updating

Download and verify the new release binary, then run:

```bash
repo-knowledge update \
  --target /path/to/service-repository \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.6.1
```

When `--agent` and `--all-agents` are omitted, `update` keeps the adapter selection recorded by the existing installation. Supply one or more `--agent` options to select a subset, or `--all-agents` to switch the installation to every currently supported adapter. Unmodified obsolete toolkit-managed adapter files are removed, modified obsolete files are preserved and reported, and consumer-owned instructions remain untouched.

The new binary contains its own templates, policy, schemas, and skills, so no central source checkout is required. Run `doctor`, `scan`, and `audit` after updating. Change the GitLab include ref and package version in the same merge request.

## GitLab CI

Add this to the consuming `.gitlab-ci.yml`:

```yaml
include:
  - project: engineering/repository-knowledge
    ref: v0.6.1
    file: /adapters/gitlab/documentation-check.yml

variables:
  REPO_KNOWLEDGE_TOOLKIT_PROJECT_ID: "12345"
  REPO_KNOWLEDGE_TOOLKIT_VERSION: v0.6.1
  REPO_KNOWLEDGE_ENFORCEMENT: advisory
```

`REPO_KNOWLEDGE_TOOLKIT_PROJECT_ID` is the numeric GitLab project ID of the central toolkit. The consuming project must be allowed to download its Generic Package Registry artifacts with `CI_JOB_TOKEN`; configure the toolkit project's job-token allowlist when cross-project access is restricted.

The job detects Linux `amd64` or `arm64`, downloads the pinned binary and `SHA256SUMS`, verifies the artifact, analyzes the merge-request or push diff, uploads a JSON report, and never commits changes. It uses full Git history for reliable diff-base access. Repositories with custom stages that omit `test` must override the included job's stage.
