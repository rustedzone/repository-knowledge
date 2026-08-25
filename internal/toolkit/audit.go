package toolkit

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
var markdownCommentPattern = regexp.MustCompile(`(?s)<!--.*?-->`)

func Doctor(root string) (DoctorReport, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return DoctorReport{}, fmt.Errorf("resolve repository root: %w", err)
	}
	checks := make([]Check, 0)
	add := func(id string, ok bool, detail string) {
		checks = append(checks, Check{ID: id, OK: ok, Detail: detail})
	}

	manifestPath := filepath.Join(root, ".repo-knowledge", "toolkit.json")
	manifest, exists, err := readOptionalManifest(manifestPath)
	if err != nil {
		return DoctorReport{}, err
	}
	add("toolkit-manifest", exists, manifestPath)
	add("binary-version", exists && manifest.ToolkitVersion == repositoryknowledge.Version(), fmt.Sprintf("installed=%s binary=%s", manifest.ToolkitVersion, repositoryknowledge.Version()))
	adapters, adapterErr := normalizeAgentAdapters(manifest.AgentAdapters)
	adapterDetail := strings.Join(adapters, ", ")
	if adapterErr != nil {
		adapterDetail = adapterErr.Error()
	}
	add("agent-adapters", exists && adapterErr == nil && len(adapters) > 0, adapterDetail)
	if contains(adapters, agentCodex) {
		skillPath := filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md")
		add("codex-skill", regularFile(skillPath), ".agents/skills/repository-knowledge/SKILL.md")
		agentsPath := filepath.Join(root, "AGENTS.md")
		agentsData, _ := os.ReadFile(agentsPath)
		add("codex-entry-contract", strings.Contains(string(agentsData), ManagedBegin), "managed AGENTS.md block")
	}
	if contains(adapters, agentClaudeCode) {
		add("claude-code-skill", regularFile(filepath.Join(root, ".claude", "skills", "repository-knowledge", "SKILL.md")), ".claude/skills/repository-knowledge/SKILL.md")
		add("claude-code-entry-contract", regularFile(filepath.Join(root, ".claude", "rules", "repository-knowledge.md")), ".claude/rules/repository-knowledge.md")
	}
	if contains(adapters, agentAntigravityIDE) {
		add("antigravity-ide-skill", regularFile(filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md")), ".agents/skills/repository-knowledge/SKILL.md")
		add("antigravity-ide-entry-contract", regularFile(filepath.Join(root, ".agents", "rules", "repository-knowledge.md")), ".agents/rules/repository-knowledge.md")
	}
	if contains(adapters, agentCursor) {
		add("cursor-skill", regularFile(filepath.Join(root, ".cursor", "skills", "repository-knowledge", "SKILL.md")), ".cursor/skills/repository-knowledge/SKILL.md")
		add("cursor-entry-contract", regularFile(filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc")), ".cursor/rules/repository-knowledge.mdc")
	}

	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config, configErr := readJSON[RepositoryConfig](configPath, false)
	configErrors := validateRepositoryConfig(config)
	if configErr != nil {
		configErrors = append(configErrors, configErr.Error())
	}
	add("repository-config", len(configErrors) == 0, detailOrPath(configErrors, configPath))
	indexPath, pathErr := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if pathErr != nil {
		add("documentation-index", false, pathErr.Error())
	} else {
		add("documentation-index", regularFile(indexPath), config.Documentation.Index)
	}
	add("policy-contract", regularFile(filepath.Join(root, ".repo-knowledge", "policy", "contract.json")), "installed policy snapshot")

	if exists {
		mismatches := make([]string, 0)
		for _, item := range manifest.ManagedFiles {
			path, err := repositoryPath(root, item.Path, "managed file path")
			if err != nil {
				mismatches = append(mismatches, item.Path)
				continue
			}
			digest, err := sha256File(path)
			if err != nil || digest != item.SHA256 {
				mismatches = append(mismatches, item.Path)
			}
		}
		sort.Strings(mismatches)
		detail := "all managed files match"
		if len(mismatches) > 0 {
			detail = strings.Join(mismatches, ", ")
		}
		add("managed-file-integrity", len(mismatches) == 0, detail)
	}
	status := "pass"
	for _, check := range checks {
		if !check.OK {
			status = "fail"
			break
		}
	}
	return DoctorReport{Status: status, Checks: checks}, nil
}

func Audit(root, base, head string) (AuditReport, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return AuditReport{}, fmt.Errorf("resolve repository root: %w", err)
	}
	health, err := Doctor(root)
	if err != nil {
		return AuditReport{}, err
	}
	findings := make([]AuditFinding, 0)
	for _, check := range health.Checks {
		if !check.OK {
			findings = append(findings, AuditFinding{Severity: "error", ID: check.ID, Message: check.Detail})
		}
	}
	config, err := readJSON[RepositoryConfig](filepath.Join(root, ".repo-knowledge", "repository.json"), false)
	if err != nil {
		return AuditReport{}, err
	}
	if len(config.Capabilities) == 0 {
		findings = append(findings, AuditFinding{
			Severity: "warning",
			ID:       "no-knowledge-routes",
			Message:  "no verified capability routes are configured; structural inventory alone is not human-readable repository documentation",
		})
	}
	for _, capability := range config.Capabilities {
		if len(capability.Documentation) == 0 {
			findings = append(findings, AuditFinding{Severity: "warning", ID: "unrouted-capability", Message: capability.ID})
		}
		for _, document := range capability.Documentation {
			path, err := repositoryPath(root, document, "capability documentation")
			if err != nil || !regularFile(path) {
				findings = append(findings, AuditFinding{Severity: "error", ID: "missing-capability-document", Message: document})
			}
		}
	}
	index, err := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if err == nil && regularFile(index) {
		links, readErr := markdownLinks(index)
		if readErr != nil {
			return AuditReport{}, readErr
		}
		for _, link := range links {
			withoutAnchor := strings.SplitN(link, "#", 2)[0]
			if withoutAnchor == "" {
				continue
			}
			target := filepath.Join(filepath.Dir(index), filepath.FromSlash(withoutAnchor))
			if _, err := os.Stat(target); err != nil {
				findings = append(findings, AuditFinding{Severity: "warning", ID: "broken-index-link", Message: link})
			}
		}
	}
	scanPath := filepath.Join(root, ".repo-knowledge", "scan-state.json")
	if !regularFile(scanPath) {
		findings = append(findings, AuditFinding{Severity: "warning", ID: "never-scanned", Message: "run repo-knowledge scan"})
	} else {
		state, err := readJSON[ScanState](scanPath, true)
		if err != nil {
			return AuditReport{}, err
		}
		if state.ScannedCommit != currentCommit(root) {
			findings = append(findings, AuditFinding{Severity: "warning", ID: "scan-stale", Message: "scan state does not match HEAD"})
		}
	}
	var impact *ImpactReport
	if base != "" || head != "" {
		value, err := ValidateImpact(root, base, head, "")
		if err != nil {
			return AuditReport{}, err
		}
		impact = &value
		for _, finding := range value.Findings {
			findings = append(findings, AuditFinding{Severity: "warning", ID: finding.ID, Message: finding.Message})
		}
	}
	status := "pass"
	for _, finding := range findings {
		if finding.Severity == "error" {
			status = "fail"
			break
		}
		status = "warning"
	}
	return AuditReport{Status: status, Findings: findings, Impact: impact}, nil
}

func Rebuild(root string, apply bool) (RebuildResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return RebuildResult{}, fmt.Errorf("resolve repository root: %w", err)
	}
	state, err := Scan(root, true)
	if err != nil {
		return RebuildResult{}, err
	}
	inventory := renderInventory(state)
	proposal, err := repositoryPath(root, ".repo-knowledge/rebuild-proposal.md", "rebuild proposal")
	if err != nil {
		return RebuildResult{}, err
	}
	if err := writeFileAtomic(proposal, []byte(inventory), 0o644); err != nil {
		return RebuildResult{}, err
	}
	result := RebuildResult{
		Proposal:                      proposal,
		ArtifactKind:                  "structural_inventory",
		SemanticDocumentationComplete: false,
		NextStep:                      "inspect repository evidence, write human-readable guides, and populate docs/index.md plus .repo-knowledge/repository.json routes",
	}
	if !apply {
		return result, nil
	}
	config, err := readJSON[RepositoryConfig](filepath.Join(root, ".repo-knowledge", "repository.json"), true)
	if err != nil {
		return RebuildResult{}, err
	}
	destination, err := repositoryPath(root, config.Documentation.GeneratedInventory, "documentation.generated_inventory")
	if err != nil {
		return RebuildResult{}, err
	}
	if info, err := os.Stat(destination); err == nil {
		if info.IsDir() {
			return RebuildResult{}, fmt.Errorf("generated inventory path is a directory: %s", destination)
		}
		existing, readErr := os.ReadFile(destination)
		if readErr != nil {
			return RebuildResult{}, fmt.Errorf("read generated inventory %s: %w", destination, readErr)
		}
		if !strings.Contains(string(existing), GeneratedBegin) || !strings.Contains(string(existing), GeneratedEnd) {
			return RebuildResult{}, fmt.Errorf("refusing to overwrite non-generated documentation: %s", destination)
		}
	} else if !os.IsNotExist(err) {
		return RebuildResult{}, fmt.Errorf("inspect generated inventory %s: %w", destination, err)
	}
	if err := writeFileAtomic(destination, []byte(inventory), 0o644); err != nil {
		return RebuildResult{}, err
	}
	result.Applied = true
	result.GeneratedInventory = destination
	return result, nil
}

func renderInventory(state ScanState) string {
	var output strings.Builder
	fmt.Fprintln(&output, GeneratedBegin)
	fmt.Fprintln(&output, "# Structural discovery inventory")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "> This generated appendix is discovery input, not human-readable repository documentation. It does not explain purpose, architecture, behavior, or business meaning.")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "Generated: %s\n", state.GeneratedAt)
	commit := state.ScannedCommit
	if commit == "" {
		commit = "uncommitted/non-git"
	}
	fmt.Fprintf(&output, "Scanned commit: `%s`\n", commit)
	fmt.Fprintf(&output, "Files inventoried: %d\n\n", state.FileCount)
	fmt.Fprintln(&output, "This file is mechanically derived. Capability signals are inspection leads, not verified semantic claims.")
	fmt.Fprint(&output, "\n## High-signal manifests\n\n")
	if len(state.Manifests) == 0 {
		fmt.Fprintln(&output, "- None detected.")
	} else {
		for _, manifest := range state.Manifests {
			fmt.Fprintf(&output, "- `%s`\n", manifest)
		}
	}
	fmt.Fprint(&output, "\n## Root files\n\n")
	if len(state.RootFiles) == 0 {
		fmt.Fprintln(&output, "- None detected.")
	} else {
		for _, rootFile := range state.RootFiles {
			fmt.Fprintf(&output, "- `%s`\n", rootFile)
		}
	}
	fmt.Fprint(&output, "\n## Source structure\n\n")
	if len(state.Modules) == 0 {
		fmt.Fprintln(&output, "- None detected.")
	} else {
		for _, module := range state.Modules {
			fmt.Fprintf(&output, "- `%s` — %d files\n", module.Path, module.FileCount)
		}
	}
	fmt.Fprint(&output, "\n## Existing documentation\n\n")
	if len(state.Documentation) == 0 {
		fmt.Fprintln(&output, "- None detected.")
	} else {
		for _, document := range state.Documentation {
			fmt.Fprintf(&output, "- `%s`\n", document)
		}
	}
	fmt.Fprint(&output, "\n## Capability leads\n\n")
	if len(state.CapabilitySignals) == 0 {
		fmt.Fprintln(&output, "- None detected from structural signals.")
	} else {
		for _, signal := range state.CapabilitySignals {
			anchors := signal.DetectedBy
			if len(anchors) > 5 {
				anchors = anchors[:5]
			}
			quoted := make([]string, 0, len(anchors))
			for _, anchor := range anchors {
				quoted = append(quoted, "`"+anchor+"`")
			}
			label := strings.ReplaceAll(signal.ID, "_", " ")
			fmt.Fprintf(&output, "- **%s** (unverified discovery signal) — %s\n", label, strings.Join(quoted, ", "))
		}
	}
	fmt.Fprint(&output, "\n## Interpretation limits\n\n")
	fmt.Fprintln(&output, "- Verify behavior with targeted tests, configuration, schemas, and source inspection.")
	fmt.Fprintln(&output, "- Preserve requirements, operational knowledge, and historical decisions that source cannot derive.")
	fmt.Fprintln(&output, "- Use the repository-knowledge skill to write semantic guides and populate `docs/index.md`; applying this inventory does not complete a documentation request.")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, GeneratedEnd)
	return output.String()
}

func validateRepositoryConfig(config RepositoryConfig) []string {
	errors := make([]string, 0)
	if config.SchemaVersion != "1.0" {
		errors = append(errors, "schema_version must be 1.0")
	}
	if strings.TrimSpace(config.Repository.Name) == "" {
		errors = append(errors, "repository.name is required")
	}
	if strings.TrimSpace(config.Documentation.Index) == "" {
		errors = append(errors, "documentation.index is required")
	}
	if config.CI.Enforcement != "advisory" && config.CI.Enforcement != "acknowledgment" && config.CI.Enforcement != "enforced" {
		errors = append(errors, "ci.enforcement must be advisory, acknowledgment, or enforced")
	}
	identifiers := make(map[string]struct{})
	for index, capability := range config.Capabilities {
		if strings.TrimSpace(capability.ID) == "" {
			errors = append(errors, fmt.Sprintf("capabilities[%d].id is required", index))
			continue
		}
		if _, exists := identifiers[capability.ID]; exists {
			errors = append(errors, "capability ids must be unique")
		}
		identifiers[capability.ID] = struct{}{}
	}
	return errors
}

func markdownLinks(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Markdown links from %s: %w", path, err)
	}
	content := markdownCommentPattern.ReplaceAllString(string(data), "")
	matches := markdownLinkPattern.FindAllStringSubmatch(content, -1)
	links := make([]string, 0, len(matches))
	for _, match := range matches {
		value := match[1]
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "mailto:") || strings.HasPrefix(value, "#") {
			continue
		}
		links = append(links, value)
	}
	return links, nil
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func detailOrPath(values []string, path string) string {
	if len(values) == 0 {
		return path
	}
	return strings.Join(values, "; ")
}
