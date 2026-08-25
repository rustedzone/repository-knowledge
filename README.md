# Repository Knowledge

Repository Knowledge is a portable, versioned contract that helps coding agents reuse repository understanding and helps CI detect documentation drift from manually written changes.

It implements two funnels over one policy:

```text
prompt -> native agent rule -> knowledge skill -> verified assessment -> implementation -> reconciliation
diff   -> Go validator -> documentation-impact validation -> report/pass/fail
```

Documentation is an active routing library, not a source of truth that overrides runtime behavior, tests, schemas, migrations, effective configuration, or implementation.

## Agent compatibility

Repository Knowledge provides proactive, prompt-triggered integration for Codex, Claude Code, Antigravity IDE, and Cursor. The Go executable installs each selected agent's native project rule and skill files.

| `--agent` value | Proactive integration |
| --- | --- |
| `codex` | Managed block in `AGENTS.md` and skill in `.agents/skills/repository-knowledge/`. |
| `claude-code` | Rule in `.claude/rules/repository-knowledge.md` and skill in `.claude/skills/repository-knowledge/`. |
| `antigravity-ide` | Rule in `.agents/rules/repository-knowledge.md` and skill in `.agents/skills/repository-knowledge/`. |
| `cursor` | Always-applied rule in `.cursor/rules/repository-knowledge.mdc` and skill in `.cursor/skills/repository-knowledge/`. |

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

The binary does not run as a prompt-interception daemon. Proactive behavior comes from the installed native rule directing the selected agent to load the repository-knowledge skill when a repository task begins.

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
2. **GitLab CI validation** — an optional fallback for changes written without an agent.

Start with proactive agent support. Add GitLab CI later if you need it.

### Step 1: Get the binary

Download the released binary for your operating system plus `LICENSE`, verify both with `SHA256SUMS`, and retain the license with any redistributed binary. Put the executable somewhere on `PATH` as `repo-knowledge`.

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
.agents/skills/repository-knowledge/           Codex or Antigravity IDE skill
.claude/rules/repository-knowledge.md          Claude Code routing rule
.claude/skills/repository-knowledge/           Claude Code skill
.agents/rules/repository-knowledge.md          Antigravity IDE routing rule
.cursor/rules/repository-knowledge.mdc         Cursor always-applied routing rule
.cursor/skills/repository-knowledge/           Cursor skill
```

It does **not**:

- install GitLab CI;
- commit the `repo-knowledge` binary into the repository;
- replace an existing `docs/index.md`;
- overwrite repository-owned configuration or documentation;
- run a scan or rebuild automatically.

The skill cannot be installed usefully as only `SKILL.md`: it depends on the shared contract, repository configuration, and documentation index. The command above installs that complete prompt-side bundle without enabling CI.

### Step 3: Verify the installation

```bash
repo-knowledge doctor --target .
```

After this passes, open the selected agent in the repository and use it normally. For repository-related prompts, its installed rule routes work through the repository-knowledge skill automatically.

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

### Step 5: Optionally add GitLab CI

The proactive agent installation is complete without CI. To validate changes made manually or by other tools, follow [GitLab CI fallback](#gitlab-ci-fallback) and the detailed [installation guide](docs/installation.md#gitlab-ci).

### Record an explicit production source and version

The short command records the running binary version automatically. For centrally managed installations, you can also record the canonical source and pinned ref:

```bash
repo-knowledge install \
  --target /path/to/your-repository \
  --agent codex \
  --source https://github.com/rustedzone/repository-knowledge \
  --ref v0.6.1
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
├── adapters/gitlab/         pinned binary CI adapter
├── evals/                   isolated prompt-driven agent evaluations
├── examples/                adoption examples
└── docs/                    architecture and operating guidance
```

The contract is the product. The Codex, Claude Code, Antigravity IDE, Cursor, and GitLab adapters consume it. The Go executable embeds every toolkit-managed asset, so a released binary can install or update a repository without a toolkit checkout, Go toolchain, Python runtime, or package download. The build has no third-party Go dependencies and release artifacts use `CGO_ENABLED=0`.

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

## Proactive agent behavior

For each selected agent, the binary installs a native project rule plus the repository-knowledge skill. The rule instructs the agent to start from `docs/index.md`, select the smallest relevant knowledge set for ordinary work, verify important claims against repository evidence, and reconcile documentation impact after implementation. Full documentation-generation requests instead require repository-wide classification, evidence-proportional coverage, detailed profile-specific guides, a populated index, and verified capability routes. The binary supports this lifecycle but does not run as a prompt-interception daemon.

## Continuous integration and releases

GitHub pull requests and pushes to `main` run tests, race-enabled tests, vet, formatting checks, and builds for both CLIs using the patched Go release selected by the workflow. Version tags build the cross-platform release set, generate GitHub artifact attestations, and publish the binaries, `LICENSE`, and `SHA256SUMS` to a GitHub release.

The GitLab CI fallback is optional and is not installed by `repo-knowledge install`.

The reusable include downloads the pinned Linux binary from the toolkit project's GitLab Generic Package Registry, verifies `SHA256SUMS`, and runs the same impact validator over the Git diff. It uploads a JSON report and never writes or commits documentation. See [GitLab CI installation](docs/installation.md#gitlab-ci) for the include configuration. Start in `advisory`, move to `acknowledgment` after teams reliably record decisions, and use `enforced` only for deterministic mappings configured by the consuming repository.

## Intentionally deferred from V1

- Semantic LLM analysis inside CI and suggested patch generation.
- Additional agent adapters beyond Codex, Claude Code, Antigravity IDE, and Cursor.
- Framework-specific API, route, schema, permission, and configuration extractors.
- Content-aware incremental deep scans.
- A central knowledge registry, embeddings, or vector retrieval.
- Automated rewriting of semantic documentation during rebuild.

Versions follow Semantic Versioning; consumers pin tags such as `v0.6.1`. See [SECURITY.md](SECURITY.md) for private vulnerability reporting instructions.

## License

Repository Knowledge is licensed under the [Apache License 2.0](LICENSE).
