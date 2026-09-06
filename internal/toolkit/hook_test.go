package toolkit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRegistersNativeHooksForAllAgents(t *testing.T) {
	root := newGitRepository(t)
	result, err := Install(InstallOptions{Target: root, AllAgentAdapters: true})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	wantPaths := []string{
		".agents/hooks.json",
		".claude/settings.json",
		".codex/hooks.json",
		".cursor/hooks.json",
	}
	if strings.Join(result.AgentHookRegistrations, ",") != strings.Join(wantPaths, ",") {
		t.Fatalf("hook registrations = %v, want %v", result.AgentHookRegistrations, wantPaths)
	}
	for agent, relative := range hookConfigPaths {
		assertFile(t, filepath.Join(root, filepath.FromSlash(relative)))
		ok, detail := hasAgentHookRegistration(root, agent)
		if !ok {
			t.Fatalf("%s hook missing: %s", agent, detail)
		}
	}
}

func TestInstallAndAdapterSwitchPreserveConsumerHookConfiguration(t *testing.T) {
	root := newGitRepository(t)
	consumerConfigs := map[string]map[string]any{
		agentCodex: {
			"consumer": true,
			"hooks": map[string]any{
				"SessionStart": []any{map[string]any{"matcher": "startup", "hooks": []any{map[string]any{"type": "command", "command": "consumer-codex"}}}},
			},
		},
		agentClaudeCode: {
			"consumer": true,
			"hooks": map[string]any{
				"SessionStart": []any{map[string]any{"matcher": "startup", "hooks": []any{map[string]any{"type": "command", "command": "consumer-claude"}}}},
			},
		},
		agentAntigravityIDE: {
			"consumer": map[string]any{"enabled": true, "PreToolUse": []any{map[string]any{"type": "command", "command": "consumer-antigravity"}}},
		},
		agentCursor: {
			"version":  float64(1),
			"consumer": true,
			"hooks":    map[string]any{"sessionStart": []any{map[string]any{"command": "consumer-cursor"}}},
		},
	}
	consumerCommands := map[string]string{
		agentCodex: "consumer-codex", agentClaudeCode: "consumer-claude",
		agentAntigravityIDE: "consumer-antigravity", agentCursor: "consumer-cursor",
	}
	for agent, config := range consumerConfigs {
		path := filepath.Join(root, filepath.FromSlash(hookConfigPaths[agent]))
		mustWriteJSON(t, path, config)
	}

	if _, err := Install(InstallOptions{Target: root, AllAgentAdapters: true}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	for agent, command := range consumerCommands {
		config, _, err := readHookConfig(filepath.Join(root, filepath.FromSlash(hookConfigPaths[agent])))
		if err != nil {
			t.Fatal(err)
		}
		if !containsHookCommand(config, command) || !containsHookCommand(config, managedHookCommand(agent)) {
			t.Fatalf("%s merged config does not contain consumer and managed hooks: %+v", agent, config)
		}
	}

	if _, err := Install(InstallOptions{Target: root, Update: true, AgentAdapters: []string{agentCodex}}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	for agent, command := range consumerCommands {
		config, _, err := readHookConfig(filepath.Join(root, filepath.FromSlash(hookConfigPaths[agent])))
		if err != nil {
			t.Fatal(err)
		}
		if !containsHookCommand(config, command) {
			t.Fatalf("%s consumer hook was removed: %+v", agent, config)
		}
		wantManaged := agent == agentCodex
		if containsHookCommand(config, managedHookCommand(agent)) != wantManaged {
			t.Fatalf("%s managed hook presence does not match selected adapter: %+v", agent, config)
		}
	}
}

func TestInstallValidatesAllHookContainersBeforeWritingAnyHook(t *testing.T) {
	root := newGitRepository(t)
	malformed := filepath.Join(root, ".cursor", "hooks.json")
	mustWrite(t, malformed, "{not-json\n")
	_, err := Install(InstallOptions{Target: root, AllAgentAdapters: true})
	if err == nil || !strings.Contains(err.Error(), "parse JSON") {
		t.Fatalf("Install() error = %v, want malformed hook config error", err)
	}
	for _, agent := range []string{agentCodex, agentClaudeCode, agentAntigravityIDE} {
		if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(hookConfigPaths[agent]))); !os.IsNotExist(statErr) {
			t.Fatalf("install wrote %s hook config before validating every container; stat error = %v", agent, statErr)
		}
	}
	if content := mustRead(t, malformed); content != "{not-json\n" {
		t.Fatalf("install overwrote malformed consumer config: %q", content)
	}
	assertNoFile(t, filepath.Join(root, "AGENTS.md"))
	assertNoFile(t, filepath.Join(root, ".agents", "skills", "repository-knowledge", "SKILL.md"))
}

func TestHookContextUsesNestedTargetAndAgentProtocols(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AllAgentAdapters: true}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	indexPath := filepath.Join(root, "docs", "index.md")
	mustWrite(t, indexPath, mustRead(t, indexPath)+"\nUNIQUE-KNOWLEDGE-ROUTE\n")
	nested := filepath.Join(root, "src", "feature")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, agent := range []string{agentCodex, agentClaudeCode} {
		result, err := HookContext(nested, agent, strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("HookContext(%s) error = %v", agent, err)
		}
		if result.Root != root || !strings.Contains(result.Output, "status: success") || !strings.Contains(result.Output, "UNIQUE-KNOWLEDGE-ROUTE") {
			t.Fatalf("HookContext(%s) output = %q", agent, result.Output)
		}
	}

	cursor, err := HookContext(nested, agentCursor, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	var cursorOutput struct {
		AdditionalContext string `json:"additional_context"`
	}
	if err := json.Unmarshal([]byte(cursor.Output), &cursorOutput); err != nil {
		t.Fatalf("parse Cursor output: %v", err)
	}
	if !strings.Contains(cursorOutput.AdditionalContext, "UNIQUE-KNOWLEDGE-ROUTE") {
		t.Fatalf("Cursor context = %q", cursorOutput.AdditionalContext)
	}

	first, err := HookContext(nested, agentAntigravityIDE, strings.NewReader(`{"invocationNum":0}`))
	if err != nil {
		t.Fatal(err)
	}
	next, err := HookContext(nested, agentAntigravityIDE, strings.NewReader(`{"invocationNum":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(first.Output, "UNIQUE-KNOWLEDGE-ROUTE") || strings.Contains(next.Output, "UNIQUE-KNOWLEDGE-ROUTE") || !strings.Contains(next.Output, "preflight remains active") {
		t.Fatalf("Antigravity full/reminder outputs = %q / %q", first.Output, next.Output)
	}
}

func TestDoctorReportsMissingNativeHook(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentCursor}}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".cursor", "hooks.json")
	config, _, err := readHookConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := removeAgentHookFromConfig(config, agentCursor); err != nil {
		t.Fatal(err)
	}
	mustWriteJSON(t, path, config)
	report, err := Doctor(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "fail" {
		t.Fatalf("Doctor() status = %q, checks = %+v", report.Status, report.Checks)
	}
	found := false
	for _, check := range report.Checks {
		if check.ID == "cursor-preflight-hook" {
			found = true
			if check.OK {
				t.Fatal("missing Cursor hook passed doctor")
			}
		}
	}
	if !found {
		t.Fatal("Doctor() omitted Cursor preflight check")
	}
}

func TestScanIncludesSharedHookContainerWhenItHasConsumerConfiguration(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentCodex}}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".codex", "hooks.json")
	config, _, err := readHookConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	config["consumer"] = true
	mustWriteJSON(t, path, config)
	state, err := Scan(root, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, module := range state.Modules {
		if module.Path == ".codex" {
			found = true
		}
	}
	if !found {
		t.Fatal("Scan() ignored consumer-owned data in shared .codex/hooks.json")
	}
}
