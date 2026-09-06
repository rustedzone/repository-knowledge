package toolkit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallBootstrapsBlankRepositoryAndPassesDoctor(t *testing.T) {
	root := newGitRepository(t)
	result, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Action != "installed" {
		t.Fatalf("Install() action = %q, want installed", result.Action)
	}
	assertFile(t, filepath.Join(root, "docs", "index.md"))
	assertFile(t, filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md"))
	bootstrap := filepath.Join(root, ".agents", "skills", "repository-knowledge", "scripts", "install-binary.sh")
	assertFile(t, bootstrap)
	info, err := os.Stat(bootstrap)
	if err != nil {
		t.Fatalf("stat bootstrap: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("bootstrap mode = %v, want 0755", info.Mode().Perm())
	}
	assertFile(t, filepath.Join(root, ".repo-knowledge", "schemas", "repository.schema.json"))
	if _, err := os.Stat(filepath.Join(root, ".repo-knowledge", "runtime")); !os.IsNotExist(err) {
		t.Fatalf("install created a vendored runtime; stat error = %v", err)
	}
	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	if report.Status != "pass" {
		t.Fatalf("Doctor() status = %q, checks = %+v", report.Status, report.Checks)
	}
}

func TestInstallSupportsAgentAdapters(t *testing.T) {
	tests := []struct {
		name            string
		adapter         string
		canonical       string
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name:      "codex",
			adapter:   "codex",
			canonical: "codex",
			expectedFiles: []string{
				"AGENTS.md",
				".agents/skills/repository-knowledge/SKILL.md",
				".codex/hooks.json",
			},
			unexpectedFiles: []string{
				".claude/skills/repository-knowledge/SKILL.md",
				".agents/rules/repository-knowledge.md",
			},
		},
		{
			name:      "claude-code canonical name",
			adapter:   "claude-code",
			canonical: "claude-code",
			expectedFiles: []string{
				".claude/rules/repository-knowledge.md",
				".claude/skills/repository-knowledge/SKILL.md",
				".claude/settings.json",
			},
			unexpectedFiles: []string{
				"AGENTS.md",
				".agents/skills/repository-knowledge/SKILL.md",
			},
		},
		{
			name:      "claude alias",
			adapter:   "claude",
			canonical: "claude-code",
			expectedFiles: []string{
				".claude/rules/repository-knowledge.md",
				".claude/skills/repository-knowledge/SKILL.md",
				".claude/settings.json",
			},
		},
		{
			name:      "antigravity-ide canonical name",
			adapter:   "antigravity-ide",
			canonical: "antigravity-ide",
			expectedFiles: []string{
				".agents/rules/repository-knowledge.md",
				".agents/skills/repository-knowledge/SKILL.md",
				".agents/hooks.json",
			},
			unexpectedFiles: []string{
				"AGENTS.md",
				".claude/skills/repository-knowledge/SKILL.md",
			},
		},
		{
			name:      "antigravity alias",
			adapter:   "antigravity",
			canonical: "antigravity-ide",
			expectedFiles: []string{
				".agents/rules/repository-knowledge.md",
				".agents/skills/repository-knowledge/SKILL.md",
				".agents/hooks.json",
			},
		},
		{
			name:      "cursor",
			adapter:   "cursor",
			canonical: "cursor",
			expectedFiles: []string{
				".cursor/rules/repository-knowledge.mdc",
				".cursor/skills/repository-knowledge/SKILL.md",
				".cursor/hooks.json",
			},
			unexpectedFiles: []string{
				"AGENTS.md",
				".agents/skills/repository-knowledge/SKILL.md",
				".claude/skills/repository-knowledge/SKILL.md",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newGitRepository(t)
			if _, err := Install(InstallOptions{
				Target: root, Source: "test", Ref: "v0.1.0", AgentAdapters: []string{test.adapter},
			}); err != nil {
				t.Fatalf("Install() error = %v", err)
			}
			for _, relative := range test.expectedFiles {
				assertFile(t, filepath.Join(root, filepath.FromSlash(relative)))
			}
			for _, relative := range test.unexpectedFiles {
				assertNoFile(t, filepath.Join(root, filepath.FromSlash(relative)))
			}

			manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
			if err != nil {
				t.Fatalf("read manifest: %v", err)
			}
			if len(manifest.AgentAdapters) != 1 || manifest.AgentAdapters[0] != test.canonical {
				t.Fatalf("agent adapters = %v, want [%s]", manifest.AgentAdapters, test.canonical)
			}

			report, err := Doctor(root)
			if err != nil {
				t.Fatalf("Doctor() error = %v", err)
			}
			if report.Status != "pass" {
				t.Fatalf("Doctor() status = %q, checks = %+v", report.Status, report.Checks)
			}
		})
	}
}

func TestInstallSupportsMultipleAgentAdapters(t *testing.T) {
	root := newGitRepository(t)
	result, err := Install(InstallOptions{
		Target: root,
		Source: "test",
		Ref:    "v0.1.0",
		AgentAdapters: []string{
			"codex", "claude-code", "antigravity-ide", "cursor", "claude", "antigravity",
		},
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	for _, relative := range []string{
		"AGENTS.md",
		".codex/hooks.json",
		".agents/rules/repository-knowledge.md",
		".agents/hooks.json",
		".agents/skills/repository-knowledge/SKILL.md",
		".claude/rules/repository-knowledge.md",
		".claude/settings.json",
		".claude/skills/repository-knowledge/SKILL.md",
		".cursor/rules/repository-knowledge.mdc",
		".cursor/hooks.json",
		".cursor/skills/repository-knowledge/SKILL.md",
	} {
		assertFile(t, filepath.Join(root, filepath.FromSlash(relative)))
	}
	if result.ManagedFileCount == 0 {
		t.Fatal("Install() managed no files")
	}
	manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	want := []string{"codex", "claude-code", "antigravity-ide", "cursor"}
	if strings.Join(manifest.AgentAdapters, ",") != strings.Join(want, ",") {
		t.Fatalf("agent adapters = %v, want %v", manifest.AgentAdapters, want)
	}
}

func TestInstallSupportsAllAgentAdaptersPreference(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AllAgentAdapters: true}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	want := []string{"codex", "claude-code", "antigravity-ide", "cursor"}
	if strings.Join(manifest.AgentAdapters, ",") != strings.Join(want, ",") {
		t.Fatalf("agent adapters = %v, want %v", manifest.AgentAdapters, want)
	}
	for _, relative := range []string{
		"AGENTS.md",
		".agents/rules/repository-knowledge.md",
		".claude/rules/repository-knowledge.md",
		".cursor/rules/repository-knowledge.mdc",
	} {
		assertFile(t, filepath.Join(root, filepath.FromSlash(relative)))
	}
}

func TestInstallRejectsAllAgentAdaptersWithExplicitAgent(t *testing.T) {
	root := newGitRepository(t)
	_, err := Install(InstallOptions{
		Target: root, AllAgentAdapters: true, AgentAdapters: []string{"cursor"},
	})
	if err == nil || !strings.Contains(err.Error(), "--all-agents cannot be combined with --agent") {
		t.Fatalf("Install() error = %v, want conflicting preference error", err)
	}
	assertNoFile(t, filepath.Join(root, ".repo-knowledge", "toolkit.json"))
}

func TestUpdateAllAgentAdaptersReplacesExistingSelection(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{"claude-code"}}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true, AllAgentAdapters: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	want := []string{"codex", "claude-code", "antigravity-ide", "cursor"}
	if strings.Join(manifest.AgentAdapters, ",") != strings.Join(want, ",") {
		t.Fatalf("agent adapters = %v, want %v", manifest.AgentAdapters, want)
	}
	assertFile(t, filepath.Join(root, "AGENTS.md"))
	assertFile(t, filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc"))
}

func TestUpdatePreservesSelectedAdaptersAndConsumerRules(t *testing.T) {
	root := newGitRepository(t)
	localRule := filepath.Join(root, ".claude", "rules", "local-team.md")
	mustWrite(t, localRule, "# Local Claude rule\n")
	if _, err := Install(InstallOptions{
		Target: root, Source: "test", Ref: "v0.1.0", AgentAdapters: []string{"claude-code"},
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if content := mustRead(t, localRule); content != "# Local Claude rule\n" {
		t.Fatalf("update changed consumer rule: %q", content)
	}
	manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if len(manifest.AgentAdapters) != 1 || manifest.AgentAdapters[0] != "claude-code" {
		t.Fatalf("agent adapters = %v, want [claude-code]", manifest.AgentAdapters)
	}
	assertFile(t, filepath.Join(root, ".claude", "skills", "repository-knowledge", "SKILL.md"))
	assertNoFile(t, filepath.Join(root, "AGENTS.md"))
}

func TestUpdateCanSwitchAgentAdaptersWithoutRemovingConsumerInstructions(t *testing.T) {
	root := newGitRepository(t)
	agentsPath := filepath.Join(root, "AGENTS.md")
	mustWrite(t, agentsPath, "# Consumer instructions\n")
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{
		Target: root, Update: true, AgentAdapters: []string{"claude-code"},
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	content := mustRead(t, agentsPath)
	if !strings.Contains(content, "# Consumer instructions") {
		t.Fatalf("update removed consumer AGENTS.md content: %q", content)
	}
	if strings.Contains(content, ManagedBegin) || strings.Contains(content, ManagedEnd) {
		t.Fatalf("update left obsolete managed Codex block: %q", content)
	}
	assertNoFile(t, filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md"))
	assertFile(t, filepath.Join(root, ".claude", "skills", "repository-knowledge", "SKILL.md"))
}

func TestInstallRejectsUnknownAgentAdapter(t *testing.T) {
	root := newGitRepository(t)
	_, err := Install(InstallOptions{Target: root, AgentAdapters: []string{"unknown"}})
	if err == nil || !strings.Contains(err.Error(), "codex, claude-code, antigravity-ide, cursor") {
		t.Fatalf("Install() error = %v, want supported adapter list", err)
	}
}

func TestCursorUpdatePreservesConsumerRules(t *testing.T) {
	root := newGitRepository(t)
	localRule := filepath.Join(root, ".cursor", "rules", "local-team.mdc")
	mustWrite(t, localRule, "---\ndescription: Local team rule\nalwaysApply: true\n---\n\n# Local rule\n")
	if _, err := Install(InstallOptions{
		Target: root, Source: "test", Ref: "v0.5.0", AgentAdapters: []string{"cursor"},
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if content := mustRead(t, localRule); !strings.Contains(content, "# Local rule") {
		t.Fatalf("update changed consumer Cursor rule: %q", content)
	}
	assertFile(t, filepath.Join(root, ".cursor", "skills", "repository-knowledge", "SKILL.md"))
	assertFile(t, filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc"))
}

func TestCursorRuleIsAlwaysApplied(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{"cursor"}}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	rule := mustRead(t, filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc"))
	if !strings.HasPrefix(rule, "---\n") || !strings.Contains(rule, "\nalwaysApply: true\n---\n") {
		t.Fatalf("Cursor rule does not use always-applied MDC frontmatter:\n%s", rule)
	}
}

func TestUpdateCanSwitchFromCursorWithoutRemovingConsumerRules(t *testing.T) {
	root := newGitRepository(t)
	localRule := filepath.Join(root, ".cursor", "rules", "local-team.mdc")
	mustWrite(t, localRule, "---\ndescription: Local team rule\nalwaysApply: true\n---\n\n# Local rule\n")
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{"cursor"}}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true, AgentAdapters: []string{"codex"}}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	assertNoFile(t, filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc"))
	assertNoFile(t, filepath.Join(root, ".cursor", "skills", "repository-knowledge", "SKILL.md"))
	if content := mustRead(t, localRule); !strings.Contains(content, "# Local rule") {
		t.Fatalf("adapter switch changed consumer Cursor rule: %q", content)
	}
	assertFile(t, filepath.Join(root, "AGENTS.md"))
}

func TestInstallPreservesConsumerFilesOnUpdate(t *testing.T) {
	root := newGitRepository(t)
	mustWrite(t, filepath.Join(root, "docs", "index.md"), "# Existing index\n")
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "# Local rules\n")
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config := readConfigForTest(t, configPath)
	config.Repository.Owners = []string{"local-team"}
	mustWriteJSON(t, configPath, config)

	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0", Update: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if content := mustRead(t, filepath.Join(root, "docs", "index.md")); content != "# Existing index\n" {
		t.Fatalf("update changed existing docs/index.md: %q", content)
	}
	if content := mustRead(t, filepath.Join(root, "AGENTS.md")); !strings.Contains(content, "# Local rules") {
		t.Fatalf("update removed local AGENTS.md content: %q", content)
	}
	updated := readConfigForTest(t, configPath)
	if len(updated.Repository.Owners) != 1 || updated.Repository.Owners[0] != "local-team" {
		t.Fatalf("update changed consumer-owned configuration: %+v", updated.Repository.Owners)
	}
}

func TestUpdateRejectsDowngrade(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	manifestPath := filepath.Join(root, ".repo-knowledge", "toolkit.json")
	manifest, err := readJSON[ToolkitManifest](manifestPath, true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest.ToolkitVersion = "99.0.0"
	mustWriteJSON(t, manifestPath, manifest)
	if _, err := Install(InstallOptions{Target: root, Update: true}); err == nil || !strings.Contains(err.Error(), "refusing to downgrade") {
		t.Fatalf("Update() error = %v, want downgrade refusal", err)
	}
}

func TestUpdateRemovesOnlyUnmodifiedObsoleteManagedFiles(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	manifestPath := filepath.Join(root, ".repo-knowledge", "toolkit.json")
	manifest, err := readJSON[ToolkitManifest](manifestPath, true)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	obsolete := filepath.Join(root, ".repo-knowledge", "runtime", "obsolete.txt")
	modified := filepath.Join(root, ".repo-knowledge", "runtime", "modified.txt")
	mustWrite(t, obsolete, "old runtime")
	mustWrite(t, modified, "old runtime")
	obsoleteDigest, _ := sha256File(obsolete)
	modifiedDigest, _ := sha256File(modified)
	manifest.ManagedFiles = append(manifest.ManagedFiles,
		ManagedFile{Path: ".repo-knowledge/runtime/obsolete.txt", SHA256: obsoleteDigest},
		ManagedFile{Path: ".repo-knowledge/runtime/modified.txt", SHA256: modifiedDigest},
	)
	mustWriteJSON(t, manifestPath, manifest)
	mustWrite(t, modified, "consumer modified this file")

	result, err := Install(InstallOptions{Target: root, Update: true})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if _, err := os.Stat(obsolete); !os.IsNotExist(err) {
		t.Fatalf("obsolete managed file still exists; stat error = %v", err)
	}
	if !regularFile(modified) {
		t.Fatal("modified obsolete managed file was removed")
	}
	if len(result.ObsoleteManagedFilesRemoved) != 1 || result.ObsoleteManagedFilesRemoved[0] != ".repo-knowledge/runtime/obsolete.txt" {
		t.Fatalf("removed files = %v", result.ObsoleteManagedFilesRemoved)
	}
	if len(result.ModifiedObsoleteFilesPreserved) != 1 || result.ModifiedObsoleteFilesPreserved[0] != ".repo-knowledge/runtime/modified.txt" {
		t.Fatalf("preserved files = %v", result.ModifiedObsoleteFilesPreserved)
	}
}

func TestScanRecordsDiscoverySignals(t *testing.T) {
	root := newGitRepository(t)
	mustWrite(t, filepath.Join(root, "src", "api", "handler.txt"), "content")
	mustWrite(t, filepath.Join(root, "go.mod"), "module sample\n")
	state, err := Scan(root, true)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	found := false
	for _, signal := range state.CapabilitySignals {
		if signal.ID == "api" {
			found = true
			if signal.Confidence != "discovery_signal_only" {
				t.Fatalf("api confidence = %q", signal.Confidence)
			}
		}
	}
	if !found {
		t.Fatal("Scan() did not detect api structural signal")
	}
	if !contains(state.Manifests, "go.mod") {
		t.Fatalf("Scan() manifests = %v, want go.mod", state.Manifests)
	}
}

func TestScanExcludesToolkitManagedAgentAdapterFiles(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root,
		AgentAdapters: []string{
			"codex", "claude-code", "antigravity-ide", "cursor",
		},
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	state, err := Scan(root, false)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	for _, module := range state.Modules {
		if module.Path == ".agents" || module.Path == ".claude" || module.Path == ".cursor" {
			t.Fatalf("Scan() included toolkit-managed adapter module: %+v", module)
		}
	}
}

func TestRebuildProtectsHandMaintainedDocumentation(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	inventory := filepath.Join(root, "docs", "repository-inventory.md")
	mustWrite(t, inventory, "# Human knowledge\n")
	result, err := Rebuild(root, false)
	if err != nil {
		t.Fatalf("Rebuild(proposal) error = %v", err)
	}
	if result.Applied {
		t.Fatalf("Rebuild(proposal) applied = %v", result.Applied)
	}
	if _, err := Rebuild(root, true); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("Rebuild(apply) error = %v, want overwrite refusal", err)
	}
	if content := mustRead(t, inventory); content != "# Human knowledge\n" {
		t.Fatalf("Rebuild changed hand-maintained inventory: %q", content)
	}
}

func TestRebuildReportsStructuralInventoryIsNotSemanticDocumentation(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.2.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	result, err := Rebuild(root, true)
	if err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}
	if result.ArtifactKind != "structural_inventory" {
		t.Fatalf("Rebuild() artifact kind = %q", result.ArtifactKind)
	}
	if result.SemanticDocumentationComplete {
		t.Fatal("Rebuild() incorrectly reported semantic documentation complete")
	}
	if strings.TrimSpace(result.NextStep) == "" {
		t.Fatal("Rebuild() returned no semantic documentation next step")
	}
}

func TestScanGroupsSourceContainersAndSeparatesRootFiles(t *testing.T) {
	root := newGitRepository(t)
	mustWrite(t, filepath.Join(root, "package.json"), "{}\n")
	mustWrite(t, filepath.Join(root, "src", "app", "page.tsx"), "export default function Page() {}\n")
	mustWrite(t, filepath.Join(root, "src", "components", "button.tsx"), "export function Button() {}\n")
	state, err := Scan(root, false)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if !contains(state.RootFiles, "package.json") {
		t.Fatalf("Scan() root files = %v, want package.json", state.RootFiles)
	}
	modulePaths := make([]string, 0, len(state.Modules))
	for _, module := range state.Modules {
		modulePaths = append(modulePaths, module.Path)
	}
	if !contains(modulePaths, "src/app") || !contains(modulePaths, "src/components") {
		t.Fatalf("Scan() modules = %v, want expanded src modules", modulePaths)
	}
	if contains(modulePaths, "package.json") {
		t.Fatalf("Scan() treated root file as a module: %v", modulePaths)
	}
}

func TestRenderedInventoryIsReadableDiscoveryInput(t *testing.T) {
	state := ScanState{
		GeneratedAt:   "2026-08-24T00:00:00Z",
		FileCount:     3,
		RootFiles:     []string{"package.json"},
		Documentation: []string{"README.md"},
		Modules: []ModuleInventory{
			{Path: "src/app", FileCount: 2, Fingerprint: strings.Repeat("a", 64)},
		},
	}
	inventory := renderInventory(state)
	for _, wanted := range []string{"Structural discovery inventory", "Root files", "Source structure", "Existing documentation", "src/app", "README.md"} {
		if !strings.Contains(inventory, wanted) {
			t.Fatalf("renderInventory() missing %q:\n%s", wanted, inventory)
		}
	}
	if strings.Contains(strings.ToLower(inventory), "fingerprint") {
		t.Fatalf("renderInventory() exposed machine fingerprint in human-facing output:\n%s", inventory)
	}
}

func TestAuditWarnsWhenRepositoryHasNoKnowledgeRoutes(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.2.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Scan(root, true); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	report, err := Audit(root, "", "")
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	found := false
	for _, finding := range report.Findings {
		if finding.ID == "no-knowledge-routes" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Audit() findings = %+v, want no-knowledge-routes warning", report.Findings)
	}

	overview := filepath.Join(root, "docs", "overview.md")
	mustWrite(t, overview, "# Repository overview\n\nVerified guide.\n")
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config := readConfigForTest(t, configPath)
	config.Capabilities = []Capability{{
		ID:            "repository-overview",
		Documentation: []string{"docs/overview.md"},
		Evidence:      []string{"README.md"},
	}}
	mustWriteJSON(t, configPath, config)
	reconciled, err := Audit(root, "", "")
	if err != nil {
		t.Fatalf("Audit(reconciled) error = %v", err)
	}
	for _, finding := range reconciled.Findings {
		if finding.ID == "no-knowledge-routes" {
			t.Fatalf("Audit(reconciled) retained no-knowledge-routes: %+v", reconciled.Findings)
		}
	}
}

func TestRebuildRejectsPathOutsideRepository(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config := readConfigForTest(t, configPath)
	config.Documentation.GeneratedInventory = "../outside.md"
	mustWriteJSON(t, configPath, config)
	if _, err := Rebuild(root, true); err == nil || !strings.Contains(err.Error(), "outside the repository") {
		t.Fatalf("Rebuild() error = %v, want path boundary error", err)
	}
}

func TestAcknowledgmentSurvivesCommitAndInvalidatesOnNewMaterialPath(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	base := commitAll(t, root, "baseline")
	mustWrite(t, filepath.Join(root, "src", "logic.txt"), "change")

	missing, err := ValidateImpact(root, base, "", "acknowledgment")
	if err != nil {
		t.Fatalf("ValidateImpact(missing) error = %v", err)
	}
	if missing.Status != "fail" {
		t.Fatalf("ValidateImpact(missing) status = %q", missing.Status)
	}
	if _, err := Acknowledge(root, "not-required", "Internal fixture change with no semantic behavior.", base, ""); err != nil {
		t.Fatalf("Acknowledge() error = %v", err)
	}
	matching, err := ValidateImpact(root, base, "", "acknowledgment")
	if err != nil || matching.Status != "pass" {
		t.Fatalf("ValidateImpact(matching) status = %q, error = %v", matching.Status, err)
	}
	head := commitAll(t, root, "acknowledged change")
	committed, err := ValidateImpact(root, base, head, "acknowledgment")
	if err != nil || committed.Status != "pass" {
		t.Fatalf("ValidateImpact(committed) status = %q, error = %v", committed.Status, err)
	}
	mustWrite(t, filepath.Join(root, "src", "second.txt"), "another")
	stale, err := ValidateImpact(root, base, "", "acknowledgment")
	if err != nil {
		t.Fatalf("ValidateImpact(stale) error = %v", err)
	}
	if stale.Status != "fail" {
		t.Fatalf("ValidateImpact(stale) status = %q", stale.Status)
	}
}

func TestDocumentationChangeSatisfiesAcknowledgmentMode(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	base := commitAll(t, root, "baseline")
	mustWrite(t, filepath.Join(root, "source.txt"), "behavior")
	index := filepath.Join(root, "docs", "index.md")
	mustWrite(t, index, mustRead(t, index)+"\nUpdated behavior knowledge.\n")
	report, err := ValidateImpact(root, base, "", "acknowledgment")
	if err != nil {
		t.Fatalf("ValidateImpact() error = %v", err)
	}
	if report.Status != "pass" {
		t.Fatalf("ValidateImpact() status = %q, findings = %+v", report.Status, report.Findings)
	}
}

func TestImpactIgnoresToolkitManagedAgentAdapterFiles(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root,
		AgentAdapters: []string{
			"codex", "claude-code", "antigravity-ide", "cursor",
		},
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	base := commitAll(t, root, "baseline")
	claudeRule := filepath.Join(root, ".claude", "rules", "repository-knowledge.md")
	mustWrite(t, claudeRule, mustRead(t, claudeRule)+"\nToolkit update.\n")
	antigravitySkill := filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md")
	mustWrite(t, antigravitySkill, mustRead(t, antigravitySkill)+"\nToolkit update.\n")
	cursorRule := filepath.Join(root, ".cursor", "rules", "repository-knowledge.mdc")
	mustWrite(t, cursorRule, mustRead(t, cursorRule)+"\nToolkit update.\n")

	report, err := Impact(root, base, "")
	if err != nil {
		t.Fatalf("Impact() error = %v", err)
	}
	if len(report.MaterialChanges) != 0 {
		t.Fatalf("Impact() material changes = %+v, want none", report.MaterialChanges)
	}
	if len(report.IgnoredChanges) != 3 {
		t.Fatalf("Impact() ignored changes = %+v, want three adapter files", report.IgnoredChanges)
	}
}

func TestGlobMatchSupportsDoubleStar(t *testing.T) {
	tests := []struct {
		pattern string
		value   string
		want    bool
	}{
		{"docs/**", "docs/api/public.md", true},
		{"**/README.md", "packages/tool/README.md", true},
		{"docs/*.md", "docs/api/public.md", false},
		{"README.md", "README.md", true},
	}
	for _, test := range tests {
		t.Run(test.pattern+"/"+test.value, func(t *testing.T) {
			if got := globMatch(test.pattern, test.value); got != test.want {
				t.Fatalf("globMatch(%q, %q) = %v, want %v", test.pattern, test.value, got, test.want)
			}
		})
	}
}

func newGitRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runTestGit(t, root, "init", "--quiet")
	runTestGit(t, root, "config", "user.email", "test@example.invalid")
	runTestGit(t, root, "config", "user.name", "Repository Knowledge Test")
	return root
}

func commitAll(t *testing.T, root, message string) string {
	t.Helper()
	runTestGit(t, root, "add", ".")
	runTestGit(t, root, "commit", "--quiet", "-m", message)
	return runTestGit(t, root, "rev-parse", "HEAD")
}

func runTestGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	args := append([]string{"-C", root}, arguments...)
	command := exec.Command("git", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(value)
}

func mustWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	mustWrite(t, path, string(data))
}

func readConfigForTest(t *testing.T, path string) RepositoryConfig {
	t.Helper()
	value, err := readJSON[RepositoryConfig](path, true)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	return value
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	if !regularFile(path) {
		t.Fatalf("expected regular file: %s", path)
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file not to exist: %s (stat error = %v)", path, err)
	}
}
