package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rustedzone/repository-knowledge/internal/toolkit"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(--version) code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "repo-knowledge ") {
		t.Fatalf("run(--version) output = %q", stdout.String())
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(help) code = %d", code)
	}
	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("run(help) output = %q", stdout.String())
	}
}

func TestRunInstallWithMultipleAgentAdapters(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"install",
		"--target", root,
		"--agent", "claude-code",
		"--agent", "antigravity-ide",
		"--agent", "cursor",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(install) code = %d, stderr = %q", code, stderr.String())
	}
	for _, relative := range []string{
		".claude/skills/repository-knowledge/SKILL.md",
		".agents/skills/repository-knowledge/SKILL.md",
		".claude/rules/repository-knowledge.md",
		".agents/rules/repository-knowledge.md",
		".cursor/skills/repository-knowledge/SKILL.md",
		".cursor/rules/repository-knowledge.mdc",
	} {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("installed adapter file %s: info = %v, error = %v", relative, info, err)
		}
	}
}

func TestRunInstallAllAgents(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"install", "--target", root, "--all-agents"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(install --all-agents) code = %d, stderr = %q", code, stderr.String())
	}
	for _, relative := range []string{
		"AGENTS.md",
		".codex/hooks.json",
		".agents/rules/repository-knowledge.md",
		".agents/hooks.json",
		".claude/rules/repository-knowledge.md",
		".claude/settings.json",
		".cursor/rules/repository-knowledge.mdc",
		".cursor/hooks.json",
	} {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("installed adapter file %s: info = %v, error = %v", relative, info, err)
		}
	}
}

func TestRunInstallStrictAntigravityPreflightAndDoctorLiveHooks(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"install", "--target", root, "--agent", "antigravity-ide", "--antigravity-preflight", "strict",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(strict install) code = %d, stderr = %q", code, stderr.String())
	}
	data, err := os.ReadFile(filepath.Join(root, ".agents", "hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "preflight-gate --agent antigravity-ide") {
		t.Fatalf("strict install did not register preflight gate: %s", data)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"doctor", "--target", root, "--live-hooks"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "preflight-gate-self-test") {
		t.Fatalf("run(doctor --live-hooks) code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestRunHookContextUsesAgentProtocol(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"install", "--target", root, "--all-agents"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(install) code = %d, stderr = %q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code := runWithInput(
		[]string{"hook-context", "--target", root, "--agent", "cursor"},
		strings.NewReader("{}"), &stdout, &stderr,
	)
	if code != 0 {
		t.Fatalf("run(hook-context) code = %d, stderr = %q", code, stderr.String())
	}
	var output struct {
		AdditionalContext string `json:"additional_context"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("parse hook output %q: %v", stdout.String(), err)
	}
	if !strings.Contains(output.AdditionalContext, "Repository Knowledge Preflight") {
		t.Fatalf("additional context = %q", output.AdditionalContext)
	}
}

func TestRunInstallRecordsCanonicalReleaseSourceByDefault(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"install", "--target", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(install) code = %d, stderr = %q", code, stderr.String())
	}
	data, err := os.ReadFile(filepath.Join(root, ".repo-knowledge", "toolkit.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Source string `json:"source"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Source != "https://github.com/rustedzone/repository-knowledge" {
		t.Fatalf("manifest source = %q", manifest.Source)
	}
}

func TestRunRejectsAllAgentsWithExplicitAgent(t *testing.T) {
	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"install", "--target", root, "--all-agents", "--agent", "cursor"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "--all-agents cannot be combined with --agent") {
		t.Fatalf("run(conflicting flags) code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".repo-knowledge", "toolkit.json")); !os.IsNotExist(err) {
		t.Fatalf("conflicting flags wrote toolkit manifest; stat error = %v", err)
	}
}

func TestInstallHelpDocumentsAllAgents(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"install", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(install --help) code = %d", code)
	}
	if !strings.Contains(stderr.String(), "all-agents") {
		t.Fatalf("install help does not document --all-agents: %q", stderr.String())
	}
}

func TestUpdateHelpDocumentsRetainedAgentSelection(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := run([]string{"update", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(update --help) code = %d", code)
	}
	if !strings.Contains(stderr.String(), "all-agents") || !strings.Contains(stderr.String(), "retain installed selection") {
		t.Fatalf("update help does not explain adapter preferences: %q", stderr.String())
	}
}

func TestRebuildSummaryDoesNotClaimSemanticDocumentation(t *testing.T) {
	result := toolkit.RebuildResult{
		Proposal:                      "/tmp/proposal.md",
		Applied:                       true,
		GeneratedInventory:            "/tmp/repository-inventory.md",
		ArtifactKind:                  "structural_inventory",
		SemanticDocumentationComplete: false,
		NextStep:                      "inspect evidence and write repository guides",
	}
	output := summarize("rebuild", result)
	if !strings.Contains(output, "structural inventory") || !strings.Contains(output, "does not generate semantic documentation") {
		t.Fatalf("summarize(rebuild) output = %q", output)
	}
}
