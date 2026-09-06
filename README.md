# Repository Knowledge

Repository Knowledge is a portable, versioned contract that helps coding agents reuse repository understanding and helps CI detect documentation drift from manually written changes.

It implements two funnels over one policy:

```text
prompt -> native agent rule -> knowledge skill -> verified assessment -> implementation -> reconciliation
diff   -> Go validator -> documentation-impact validation -> report/pass/fail
```

Documentation is an active routing library, not a source of truth that overrides runtime behavior, tests, schemas, migrations, effective configuration, or implementation.

## Agent compatibility

Repository Knowledge provides proactive integration for Codex, Claude Code, Antigravity IDE, and Cursor. The Go executable installs each selected agent's native project rule, skill files, and lifecycle preflight hook.

| `--agent` value | Proactive integration |
| --- | --- |
| `codex` | Managed block in `AGENTS.md`, skill in `.agents/skills/repository-knowledge/`, and [`SessionStart`](https://developers.openai.com/codex/hooks) entry in `.codex/hooks.json`. |
| `claude-code` | Rule and skill under `.claude/`, plus a [`SessionStart`](https://code.claude.com/docs/en/hooks) entry merged into `.claude/settings.json`. |
| `antigravity-ide` | Rule and skill under `.agents/`, plus a [`PreInvocation`](https://antigravity.google/docs/ide/hooks/) entry merged into `.agents/hooks.json`. |
| `cursor` | Always-applied rule and skill under `.cursor/`, plus a [`sessionStart`](https://prod.cursor.com/docs/hooks) entry merged into `.cursor/hooks.json`. |

The Cursor paths follow its official [Project Rules](https://docs.cursor.com/context/rules) and [Agent Skills](https://cursor.com/docs/skills) conventions.

`claude` is accepted as an alias for `claude-code`, and `antigravity` is accepted as an alias for `antigravity-ide`. The manifest records canonical names. Install every supported adapter with one preference flag:

```bash
repo-knowledge install --target . --all-agents
```

For a subset, repeat `--agent`:

```bash
repo-knowledge install --target . \
  --agent codex \
  --agent claude-code \
  --agent antigravity-ide \
  --agent cursor
```

The binary does not run as a daemon and does not inspect prompt transcripts. Each native lifecycle hook runs `repo-knowledge hook-context` to inject the installed contract, repository routing configuration, and `docs/index.md` before ordinary repository discovery. The rule remains the portable fallback, and the agent must emit an observable `Repository knowledge preflight: loaded` progress receipt naming the selected routes.

This is a strong activation guardrail, not an absolute enforcement boundary. The binary must resolve on the agent host's `PATH`, project hooks must be enabled and trusted where the host requires review, and the host may surface hook failures without blocking a session. `repo-knowledge doctor --target .` validates the configured entries; the agent's progress receipt makes a skipped preflight visible. Existing hook configuration is preserved because the installer owns only its exact nested command entry, not the surrounding JSON file.

## Generate documentation people can learn from

Ask a supported agent for human-readable repository documentation, for example:

```text
Use repository-knowledge to inspect this entire repository and generate comprehensive, human-readable documentation. Classify every repository shape present, then explain its purpose, architecture, major flows, domains or features, interfaces, data, integrations, security, development, and operations wherever applicable. Create focused guides for each material capability, populate docs/index.md and verified capability routes, and do not stop after scan, rebuild, or a shallow overview.
```

The agent first classifies the repository using every applicable profile: backend, frontend, mobile or desktop, library or SDK, CLI, infrastructure, data or ML, monorepo, embedded, documentation/configuration, or an evidence-derived unusual shape. It then inspects implementation, tests, manifests, lockfiles, configuration, schemas, contracts, deployment evidence, and existing documentation before writing a detailed guide set proportional to the repository.

Existing prose—including README files, `CLAUDE.md`, `AGENTS.md`, comments, and older generated guides—is treated as a discovery or intent source, not automatic proof of current technical state. Current claims are verified against the authoritative artifact for that claim type. For example, framework and dependency declarations come from package/build manifests, exact resolutions come from lockfiles, implemented behavior comes from tests and source, and data shape comes from schemas and migrations. If `CLAUDE.md` says React 18 while the current manifest and lockfile establish React 19, the generated docs use React 19 and report `CLAUDE.md` as stale.

The documentation must also pass an implementation-readiness gate. Unsupported phrases such as “likely,” “typically,” or “expected” cannot substitute for source inspection. Domain guides must connect rules, invariants, state transitions, permissions, side effects, failures, and tests. Frontend guides must explain the actual route guards, state ownership, queries and mutations, forms and validation, design-system composition, BFF/API behavior, and test patterns—not merely name libraries. Recurring patterns include concise examples derived from current source with paths and symbols so a newcomer can make a first change safely.

Generation is self-remediating. Before drafting, the agent builds a private coverage ledger and a trace dossier for every material capability. Each executable capability receives a named end-to-end flow with concrete paths and symbols, state effects, failures, tests, and extension patterns. If the final quality gate finds shallow coverage, the agent returns to the implementation and revises the guides in the same task. A request to generate or rebuild must not stop at an insufficiency assessment or ask permission to begin work that was already requested; a standalone review or assessment remains read-only.

Once that evidence-backed depth is established, the agent also improves the reader experience: the overview provides a repository-specific mental model and practical newcomer path, and a small number of Mermaid component, sequence, state, or data-flow diagrams may be added when they explain relationships more clearly than prose. Diagrams are optional and never replace behavioral details, errors, tests, or evidence.

Every repository receives an overview, architecture guide, and curated index unless accurate equivalents already exist. The agent then adds focused pages for each material domain, feature, API, data model, integration, operational surface, or other profile-specific concept. A backend result might look like:

```text
docs/
├── index.md
├── repository-overview.md
├── architecture.md
├── domains/
│   ├── transaction.md
│   └── approval.md
├── api/
│   └── transaction-api.md
├── data/
│   └── transaction-model.md
├── integrations/
│   └── iiam.md
├── operations/
│   └── running-and-deployment.md
└── decisions/
    └── adr-003-approval-workflow.md
```

The tree adapts to repository evidence; it is not a mandatory template. ADRs are created only for decisions that are actually evidenced. Each guide explains behavior, flows, contracts, edge cases, change consequences, tests, and evidence—not just files or endpoints. The agent updates `docs/index.md` and `.repo-knowledge/repository.json` so readers can navigate from concepts to guides and implementation.

`repo-knowledge scan` and `repo-knowledge rebuild` do not perform semantic analysis. They produce discovery data and an optional structural appendix for the agent. A file list, fingerprint catalog, empty index, or applied inventory is not considered completed repository documentation.

## Evaluate agent behavior

The repository includes isolated prompt-driven evaluations for the failure modes reported in frontend and layered Go backend repositories. They test stale prose against stronger current evidence, end-to-end behavior tracing, state and deletion semantics, source-derived examples, verified routes, and preservation of repository source.

Prepare a disposable case, run the selected agent in the printed target using the printed prompt, then grade the result:

```bash
go run ./cmd/repo-knowledge-eval prepare \
  --case frontend-nextjs \
  --output /tmp/repository-knowledge-eval-frontend \
  --agent codex

go run ./cmd/repo-knowledge-eval grade \
  --case frontend-nextjs \
  --target /tmp/repository-knowledge-eval-frontend
```

The objective grader cannot establish that prose is correct or useful. A deterministic pass remains `pending_semantic_review` until the case rubric confirms behavioral accuracy, implementation readiness, and absence of unsupported claims. See [agent evaluations](evals/README.md) for the cases, rubric threshold, repeat-trial metrics, and report format.

## Installation

There are two independent parts:

1. **Proactive agent support** — used when a repository change starts from a Codex, Claude Code, Antigravity IDE, or Cursor prompt.
2. **CI validation** — an optional GitHub Actions or GitLab fallback for changes written without an agent.

Start with proactive agent support. Add CI later if you need it.

### Step 1: Get the binary

Download the released binary for your operating system plus `LICENSE`, verify both with `SHA256SUMS`, and retain the license with any redistributed binary. Put the executable somewhere on `PATH` as `repo-knowledge`.

If a repository already contains the complete installed skill but the command is missing—for example, after cloning the repository onto a new machine—the skill can recover it. When a CLI operation is needed, the agent reads the exact release ref and source from `.repo-knowledge/toolkit.json`, previews the platform artifact and user-local destination, and asks permission before downloading or writing anything. After approval, the bundled bootstrap helper verifies `SHA256SUMS`, verifies the GitHub artifact attestation when a compatible `gh` command is available, and installs the executable into a user-owned directory already on `PATH`. It never selects `latest`, uses `sudo`, or silently edits a Unix shell profile. On Windows, adding the default user-local directory to the user PATH is included explicitly in the permission request.

This is a recovery path, not a way to bootstrap an untrusted loose `SKILL.md`. It requires a complete toolkit installation with a pinned manifest, or the toolkit source tree with its `VERSION` file. See [Missing-binary recovery](docs/installation.md#missing-binary-recovery).

If you are testing from this source repository instead, build it locally:

```bash
make build
./repo-knowledge --version
```

When using a local build, replace `repo-knowledge` in the examples below with the full path to `./repo-knowledge`.

### Step 2: Install proactive agent support — no CI

From the repository you want the agent to understand, choose one or more adapters:

```bash
cd /path/to/your-repository
repo-knowledge install --target .
```

Codex is the default, so the short command above is equivalent to `--agent codex`. Select any other adapter explicitly:

```bash
repo-knowledge install --target . --agent claude-code
repo-knowledge install --target . --agent antigravity-ide
repo-knowledge install --target . --agent cursor
```

Repeat `--agent` to support multiple agents in the same repository. The installer adds the shared runtime-neutral bundle plus only the selected prompt adapters:

To select all supported agents instead, use:

```bash
repo-knowledge install --target . --all-agents
```

`--all-agents` cannot be combined with `--agent`.

```text
.repo-knowledge/policy/                        shared contract and defaults
.repo-knowledge/schemas/                       configuration schemas
.repo-knowledge/repository.json                repository-owned configuration
.repo-knowledge/local-invariants.json          repository-owned rules
.repo-knowledge/local-impact-rules.json        repository-owned impact rules
docs/index.md                                  created only when absent

AGENTS.md                                      Codex only: managed routing block
.codex/hooks.json                              Codex shared hook container; toolkit manages one SessionStart entry
.agents/skills/repository-knowledge/           Codex or Antigravity IDE skill
.claude/rules/repository-knowledge.md          Claude Code routing rule
.claude/skills/repository-knowledge/           Claude Code skill
.claude/settings.json                          Claude Code shared settings; toolkit manages one SessionStart entry
.agents/rules/repository-knowledge.md          Antigravity IDE routing rule
.agents/hooks.json                             Antigravity shared hook container; toolkit manages one named entry
.cursor/rules/repository-knowledge.mdc         Cursor always-applied routing rule
.cursor/skills/repository-knowledge/           Cursor skill
.cursor/hooks.json                             Cursor shared hook container; toolkit manages one sessionStart entry
```

It does **not**:

- install GitHub Actions or GitLab CI;
- commit the `repo-knowledge` binary into the repository;
- replace an existing `docs/index.md`;
- overwrite repository-owned configuration or documentation;
- run a scan or rebuild automatically.

The repository installer does not copy its running executable into the consuming repository. The installed skill does include small bootstrap helpers that may later install the same pinned release into the user's PATH, but only after a separate, explicit permission prompt.

The skill cannot be installed usefully as only `SKILL.md`: it depends on the shared contract, repository configuration, and documentation index. The command above installs that complete prompt-side bundle without enabling CI.

### Step 3: Verify the installation

```bash
repo-knowledge doctor --target .
```

After this passes, review/trust the project hook if the host asks, then open a new agent session in the repository. For repository-related prompts, confirm the first progress update reports `Repository knowledge preflight: loaded` and names the selected routes.

### Step 4: Optionally create a structural appendix

This step is optional and separate from generating human-readable documentation:

```bash
repo-knowledge scan --target .
repo-knowledge rebuild --target .
```

Review `.repo-knowledge/rebuild-proposal.md`. To accept the generated structural appendix:

```bash
repo-knowledge rebuild --target . --apply
```

The inventory is discovery input only. Ask the installed agent to inspect the evidence, write semantic guides, and add verified knowledge routes to `.repo-knowledge/repository.json` and `docs/index.md`.

### Step 5: Optionally add CI

The proactive agent installation is complete without CI. GitHub repositories can call the pinned reusable workflow:

```yaml
jobs:
  documentation-impact:
    permissions:
      contents: read
      attestations: read
    uses: rustedzone/repository-knowledge/.github/workflows/documentation-check.yml@v0.9.0
    with:
      toolkit-version: v0.9.0
      enforcement: advisory
```

The caller normally triggers this job for `pull_request` and pushes to `main`. GitLab repositories use the existing pinned include. See [GitHub Actions CI](docs/installation.md#github-actions-ci) and [GitLab CI](docs/installation.md#gitlab-ci) for complete examples and security boundaries.

### Record an explicit production source and version

The short command records the running binary version automatically. For centrally managed installations, you can also record the canonical source and pinned ref:

```bash
repo-knowledge install \
  --target /path/to/your-repository \
  --agent codex \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.9.0
```

## V1 contents

```text
repository-knowledge/
├── assets.go                embedded managed assets
├── cmd/repo-knowledge/      binary entrypoint
├── internal/toolkit/        deterministic operations
├── policy/                  shared runtime-neutral contract and defaults
├── schemas/                 consumer configuration interfaces
├── skills/                  shared agent skill
├── templates/               bootstrap and consumer-owned defaults
├── adapters/                GitHub and GitLab CI adapters
├── evals/                   isolated prompt-driven agent evaluations
├── examples/                adoption examples
└── docs/                    architecture and operating guidance
```

The contract is the product. The Codex, Claude Code, Antigravity IDE, Cursor, GitHub Actions, and GitLab adapters consume it. The Go executable embeds every toolkit-managed asset, so a released binary can install or update a repository without a toolkit checkout, Go toolchain, Python runtime, or package download. The build has no third-party Go dependencies and release artifacts use `CGO_ENABLED=0`.

See [Installation](docs/installation.md), [Testing](docs/testing.md), and [Architecture](docs/architecture.md).

## Commands

| Command | Outcome |
| --- | --- |
| `install` | Install embedded policy, schemas, selected prompt adapters, and missing consumer defaults; use `--all-agents` for every supported adapter. |
| `update` | Refresh only toolkit-owned files; omit adapter flags to retain the selection or use `--all-agents` to switch to every adapter. |
| `doctor` | Validate configuration, binary compatibility, and managed-file integrity. |
| `scan` | Write machine-readable structural scan state with manifests, module fingerprints, existing docs, and capability leads. |
| `audit` | Report missing routes, broken index links, stale scan state, and optional diff gaps. |
| `rebuild` | Write an optional structural-inventory proposal; `--apply` updates that generated appendix only, not semantic documentation. |
| `impact` | Classify a diff and calculate its stable material-path fingerprint. |
| `acknowledge` | Record a `required` or `not_required` decision bound to that fingerprint. |
| `validate-doc-impact` | Apply advisory, acknowledgment, or explicitly mapped enforced checks. |
| `hook-context` | Internal host-hook command that emits the bounded repository preflight in the selected agent's native protocol. |

## Proactive agent behavior

For each selected agent, the binary installs a native project rule, the repository-knowledge skill, and a lifecycle hook. The hook injects the routing preflight; the rule supplies the manual fallback and instructs the agent to start from `docs/index.md`, select the smallest relevant knowledge set for ordinary work, verify important claims against repository evidence, and reconcile documentation impact after implementation. The first progress update must confirm the preflight and selected routes. Full documentation-generation requests instead require repository-wide classification, evidence-proportional coverage, detailed profile-specific guides, a populated index, and verified capability routes. The binary supports this lifecycle without running as a daemon or parsing prompt transcripts.

## Continuous integration and releases

GitHub pull requests and pushes to `main` run ordinary tests, vet, formatting checks, and builds for both CLIs on Go 1.22.0 and Go 1.25.14. Race-enabled tests run once on Go 1.25.14. Every external workflow action is pinned to a full commit SHA, and a repository test rejects mutable tags, branches, short SHAs, and mutable Docker tags.

Version tags use a read-only job to verify source and build the cross-platform release set. A SHA-pinned artifact transfer carries that set to a separate publishing job, which alone receives contents, identity-token, and attestation write permissions. The publishing job generates GitHub artifact attestations and publishes the binaries, GitHub adapter, `LICENSE`, and `SHA256SUMS` without rebuilding them.

The GitHub Actions and GitLab CI integrations are optional and are not installed by `repo-knowledge install`.

The GitHub reusable workflow checks out the caller with full history and without persisted credentials, downloads the pinned Linux binary and GitHub adapter from the matching release, verifies their checksums and attestations, runs the validator over the explicit pull-request or push range, and uploads `repository-knowledge-impact.json` even when blocking validation fails. It pins its GitHub-owned action dependencies to full commit SHAs, needs only read access to contents and attestations, and does not use `pull_request_target`, secrets, or repository write permissions.

The GitLab reusable include downloads the pinned Linux binary from the toolkit project's Generic Package Registry, verifies `SHA256SUMS`, and runs the same impact validator over the Git diff. Both CI adapters upload a JSON report and never write or commit documentation. Start in `advisory`, move to `acknowledgment` after teams reliably record decisions, and use `enforced` only for deterministic mappings configured by the consuming repository.

## Intentionally deferred from V1

- Semantic LLM analysis inside CI and suggested patch generation.
- Additional agent adapters beyond Codex, Claude Code, Antigravity IDE, and Cursor.
- Framework-specific API, route, schema, permission, and configuration extractors.
- Content-aware incremental deep scans.
- A central knowledge registry, embeddings, or vector retrieval.
- Automated rewriting of semantic documentation during rebuild.

Versions follow Semantic Versioning; consumers pin tags such as `v0.9.0`. See [SECURITY.md](SECURITY.md) for private vulnerability reporting instructions.

## License

Repository Knowledge is licensed under the [Apache License 2.0](LICENSE).
